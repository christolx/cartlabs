"use client";

import Image from "next/image";
import {
  useEffect,
  useRef,
  useState,
  type ChangeEvent,
  type DragEvent,
  type FormEvent,
} from "react";
import type { components } from "@/lib/api/schema";
import { useSession } from "@/components/session-provider";
import { errorMessage } from "@/lib/api/browser";

type ProductImage = components["schemas"]["ProductImage"];

const allowedTypes = new Set([
  "image/jpeg",
  "image/png",
  "image/webp",
  "image/avif",
]);
const maximumBytes = 5 * 1024 * 1024;
const maximumDimension = 4096;

async function validateFile(file: File) {
  if (!allowedTypes.has(file.type) || file.size > maximumBytes) {
    throw new Error("Choose JPEG, PNG, WebP, or AVIF file up to 5 MB.");
  }
  const { width, height } = await imageDimensions(file);
  if (width > maximumDimension || height > maximumDimension) {
    throw new Error("Image dimensions must not exceed 4096 × 4096 pixels.");
  }
}

function imageDimensions(file: File) {
  const objectURL = URL.createObjectURL(file);
  return new Promise<{ width: number; height: number }>((resolve, reject) => {
    const image = document.createElement("img");
    const timeout = window.setTimeout(
      () => finish(new Error("Image could not be decoded.")),
      10_000,
    );
    function finish(error?: Error) {
      window.clearTimeout(timeout);
      URL.revokeObjectURL(objectURL);
      image.onload = null;
      image.onerror = null;
      if (error) reject(error);
      else resolve({ width: image.naturalWidth, height: image.naturalHeight });
    }
    image.onload = () => finish();
    image.onerror = () => finish(new Error("Image could not be decoded."));
    image.src = objectURL;
  });
}

async function uploadToCloudinary(
  file: File,
  onProgress: (value: number) => void,
) {
  const configResponse = await fetch("/api/cloudinary-config", {
    cache: "no-store",
  });
  if (!configResponse.ok)
    throw new Error("Cloudinary upload configuration is unavailable.");
  const config = (await configResponse.json()) as {
    cloudName?: string;
    uploadPreset?: string;
  };
  const { cloudName, uploadPreset } = config;
  if (!cloudName || !uploadPreset)
    throw new Error("Cloudinary upload is not configured.");

  return new Promise<string>((resolve, reject) => {
    const request = new XMLHttpRequest();
    request.open(
      "POST",
      `https://api.cloudinary.com/v1_1/${encodeURIComponent(cloudName)}/image/upload`,
    );
    request.timeout = 30_000;
    request.upload.addEventListener("progress", (event) => {
      if (event.lengthComputable)
        onProgress(Math.round((event.loaded / event.total) * 100));
    });
    request.addEventListener("load", () => {
      try {
        const response = JSON.parse(request.responseText) as {
          secure_url?: string;
          error?: { message?: string };
        };
        if (
          request.status < 200 ||
          request.status >= 300 ||
          !response.secure_url
        ) {
          reject(
            new Error(response.error?.message || "Cloudinary upload failed."),
          );
          return;
        }
        resolve(response.secure_url);
      } catch {
        reject(new Error("Cloudinary returned an invalid response."));
      }
    });
    request.addEventListener("error", () =>
      reject(new Error("Cloudinary upload failed.")),
    );
    request.addEventListener("timeout", () =>
      reject(new Error("Cloudinary upload timed out.")),
    );
    const form = new FormData();
    form.set("file", file);
    form.set("upload_preset", uploadPreset);
    request.send(form);
  });
}

export function ProductImageManager({
  productId,
  status,
  images,
  onChanged,
}: {
  productId: string;
  status: string;
  images: ProductImage[];
  onChanged: () => Promise<void>;
}) {
  const { request } = useSession();
  const [file, setFile] = useState<File | null>(null);
  const [preview, setPreview] = useState("");
  const [altText, setAltText] = useState("");
  const [progress, setProgress] = useState(0);
  const [validating, setValidating] = useState(false);
  const [busy, setBusy] = useState("");
  const [message, setMessage] = useState("");
  const [replacementAlt, setReplacementAlt] = useState<Record<string, string>>(
    {},
  );
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(
    () => () => {
      if (preview) URL.revokeObjectURL(preview);
    },
    [preview],
  );

  async function selectFile(nextFile?: File) {
    setMessage("");
    if (!nextFile) return;
    setValidating(true);
    try {
      await validateFile(nextFile);
      if (preview) URL.revokeObjectURL(preview);
      setFile(nextFile);
      setPreview(URL.createObjectURL(nextFile));
    } catch (cause) {
      setFile(null);
      setPreview("");
      setMessage(errorMessage(cause, "Invalid image."));
    } finally {
      setValidating(false);
    }
  }

  async function upload(
    fileToUpload: File,
    path: string,
    method: "POST" | "PUT",
    nextAltText: string,
  ) {
    const url = await uploadToCloudinary(fileToUpload, setProgress);
    await request(path, {
      method,
      body: JSON.stringify({ url, altText: nextAltText.trim() }),
    });
  }

  async function addImage(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!file) {
      setMessage("Choose image first.");
      return;
    }
    setBusy("add");
    setMessage("");
    setProgress(0);
    try {
      await upload(
        file,
        `/seller/products/${productId}/images`,
        "POST",
        altText,
      );
      setFile(null);
      setAltText("");
      setPreview("");
      if (inputRef.current) inputRef.current.value = "";
      await onChanged();
      setMessage("Image added.");
    } catch (cause) {
      setMessage(errorMessage(cause, "Image upload failed."));
    } finally {
      setBusy("");
      setProgress(0);
    }
  }

  async function replaceImage(
    event: ChangeEvent<HTMLInputElement>,
    image: ProductImage,
  ) {
    const replacement = event.currentTarget.files?.[0];
    event.currentTarget.value = "";
    if (!replacement) return;
    setBusy(image.id);
    setMessage("");
    setProgress(0);
    try {
      await validateFile(replacement);
      await upload(
        replacement,
        `/seller/products/${productId}/images/${image.id}`,
        "PUT",
        replacementAlt[image.id] ?? image.altText,
      );
      await onChanged();
      setMessage("Image replaced.");
    } catch (cause) {
      setMessage(errorMessage(cause, "Image replacement failed."));
    } finally {
      setBusy("");
      setProgress(0);
    }
  }

  async function deleteImage(imageId: string) {
    if (
      !window.confirm(
        "Delete this product image? Remaining images will shift forward.",
      )
    )
      return;
    setBusy(imageId);
    setMessage("");
    try {
      await request(`/seller/products/${productId}/images/${imageId}`, {
        method: "DELETE",
      });
      await onChanged();
      setMessage("Image deleted.");
    } catch (cause) {
      setMessage(errorMessage(cause, "Image deletion failed."));
    } finally {
      setBusy("");
    }
  }

  const atLimit = images.length >= 8;
  const lastPublishedImage = status === "published" && images.length === 1;
  return (
    <section className="form-panel image-manager">
      <div className="panel-heading-row">
        <div>
          <h2>Product images</h2>
          <p>Upload JPEG, PNG, WebP, or AVIF. 5 MB and 4096 px maximum.</p>
        </div>
        <span className="count-badge">{images.length} / 8</span>
      </div>
      {images.length ? (
        <div className="image-management-grid">
          {images.map((image) => (
            <figure key={image.id}>
              <Image
                src={image.url}
                alt={image.altText}
                width={240}
                height={180}
              />
              <figcaption>
                {image.position === 0 ? "Cover · " : ""}
                {image.altText || "No alt text"}
              </figcaption>
              <label>
                <span>Alt text for replacement</span>
                <input
                  value={replacementAlt[image.id] ?? image.altText}
                  maxLength={160}
                  onChange={(event) =>
                    setReplacementAlt((values) => ({
                      ...values,
                      [image.id]: event.currentTarget.value,
                    }))
                  }
                  disabled={Boolean(busy)}
                />
              </label>
              <div className="image-actions">
                <label className="button button-secondary">
                  {busy === image.id ? "Working" : "Replace"}
                  <input
                    type="file"
                    accept="image/jpeg,image/png,image/webp,image/avif"
                    disabled={Boolean(busy)}
                    onChange={(event) => void replaceImage(event, image)}
                  />
                </label>
                <button
                  className="text-button danger-button"
                  type="button"
                  disabled={Boolean(busy) || lastPublishedImage}
                  title={
                    lastPublishedImage
                      ? "Published product must keep one image"
                      : undefined
                  }
                  onClick={() => void deleteImage(image.id)}
                >
                  Delete
                </button>
              </div>
            </figure>
          ))}
        </div>
      ) : (
        <p className="empty-copy">No images uploaded.</p>
      )}
      <form
        className="stack-form compact-form image-upload-form"
        onSubmit={addImage}
      >
        <div
          className="image-drop-zone"
          onDragOver={(event: DragEvent) => event.preventDefault()}
          onDrop={(event: DragEvent) => {
            event.preventDefault();
            void selectFile(event.dataTransfer.files[0]);
          }}
        >
          {preview ? (
            <>
              {/* eslint-disable-next-line @next/next/no-img-element -- Blob URLs bypass Next image optimization. */}
              <img
                src={preview}
                alt="Selected image preview"
                width={240}
                height={180}
              />
            </>
          ) : (
            <p>
              {validating
                ? "Checking image…"
                : "Drop image here or choose file."}
            </p>
          )}
          <label className="sr-only" htmlFor="product-image-file">
            Product image file
          </label>
          <input
            id="product-image-file"
            ref={inputRef}
            type="file"
            accept="image/jpeg,image/png,image/webp,image/avif"
            disabled={atLimit || validating || Boolean(busy)}
            onChange={(event) =>
              void selectFile(event.currentTarget.files?.[0])
            }
          />
        </div>
        <label>
          <span>Alt text</span>
          <input
            value={altText}
            onChange={(event) => setAltText(event.currentTarget.value)}
            maxLength={160}
            required
            disabled={atLimit || validating || Boolean(busy)}
          />
        </label>
        {busy && progress ? (
          <progress value={progress} max="100">
            {progress}%
          </progress>
        ) : null}
        <button
          className="button button-primary"
          disabled={atLimit || !file || validating || Boolean(busy)}
        >
          {busy === "add"
            ? `Uploading ${progress}%`
            : atLimit
              ? "Eight-image limit reached"
              : "Add image"}
        </button>
      </form>
      {message ? (
        <p className="form-message" role="status">
          {message}
        </p>
      ) : null}
    </section>
  );
}
