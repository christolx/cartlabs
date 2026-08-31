"use client";

import { useCallback, useEffect, useState, type FormEvent } from "react";
import Link from "next/link";
import { useParams } from "next/navigation";
import type { components } from "@/lib/api/schema";
import { RequireRole } from "@/components/require-role";
import { useSession } from "@/components/session-provider";
import { ErrorState, LoadingState } from "@/components/async-state";
import { formatDate, Money, Status } from "@/components/marketplace-ui";
import { ProductImageManager } from "@/components/product-image-manager";
import { errorMessage } from "@/lib/api/browser";

type Product = components["schemas"]["ProductDetail"];
type Category = components["schemas"]["Category"];

function VariantAttributeFields() {
  const [rows, setRows] = useState([{ id: 0, key: "", value: "" }]);
  return (
    <div className="attribute-fields">
      <span>Attributes</span>
      {rows.map((row, index) => (
        <div className="attribute-row" key={row.id}>
          <input
            name="attributeKey"
            aria-label={`Attribute ${index + 1} name`}
            placeholder="Name"
          />
          <input
            name="attributeValue"
            aria-label={`Attribute ${index + 1} value`}
            placeholder="Value"
          />
          {rows.length > 1 ? (
            <button
              className="text-button"
              type="button"
              aria-label={`Remove attribute ${index + 1}`}
              onClick={() =>
                setRows((current) =>
                  current.filter((item) => item.id !== row.id),
                )
              }
            >
              −
            </button>
          ) : null}
        </div>
      ))}
      <button
        className="text-button attribute-add"
        type="button"
        onClick={() =>
          setRows((current) => [
            ...current,
            { id: Date.now(), key: "", value: "" },
          ])
        }
      >
        + Add attribute
      </button>
    </div>
  );
}

function ProductManagementContent() {
  const { productId } = useParams<{ productId: string }>();
  const { request } = useSession();
  const [product, setProduct] = useState<Product | null>(null);
  const [categories, setCategories] = useState<Category[]>([]);
  const [error, setError] = useState("");
  const [message, setMessage] = useState("");
  const [busy, setBusy] = useState("");
  const load = useCallback(async () => {
    setError("");
    try {
      const [nextProduct, categoryResponse] = await Promise.all([
        request<Product>(`/seller/products/${encodeURIComponent(productId)}`),
        request<{ items: Category[] }>("/categories"),
      ]);
      setProduct(nextProduct);
      setCategories(categoryResponse.items);
    } catch (cause) {
      setError(
        errorMessage(cause, "Product unavailable or not owned by this seller."),
      );
    }
  }, [productId, request]);
  useEffect(() => {
    const timer = window.setTimeout(() => void load(), 0);
    return () => window.clearTimeout(timer);
  }, [load]);

  async function run(
    label: string,
    path: string,
    init: RequestInit,
    successMessage: string,
  ) {
    setBusy(label);
    setMessage("");
    try {
      await request(path, init);
      await load();
      setMessage(successMessage);
    } catch (cause) {
      setMessage(errorMessage(cause, "Action failed."));
    } finally {
      setBusy("");
    }
  }

  async function updateProduct(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const data = new FormData(event.currentTarget);
    await run(
      "product",
      `/seller/products/${productId}`,
      {
        method: "PATCH",
        body: JSON.stringify({
          categoryId: data.get("categoryId"),
          name: data.get("name"),
          slug: data.get("slug"),
          description: data.get("description"),
        }),
      },
      "Product details saved.",
    );
  }
  async function addVariant(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = event.currentTarget;
    const data = new FormData(form);
    const attributeKeys = data.getAll("attributeKey").map(String);
    const attributeValues = data.getAll("attributeValue").map(String);
    const attributes = Object.fromEntries(
      attributeKeys
        .map((key, index) => [
          key.trim(),
          (attributeValues[index] ?? "").trim(),
        ])
        .filter(([key, value]) => Boolean(key && value)),
    );
    await run(
      "variant",
      `/seller/products/${productId}/variants`,
      {
        method: "POST",
        body: JSON.stringify({
          sku: data.get("sku"),
          name: data.get("name"),
          attributes,
          priceMinor: Number(data.get("priceMinor")),
          currency: "IDR",
          stock: Number(data.get("stock")),
        }),
      },
      "Variant added.",
    );
    form.reset();
  }
  async function adjustInventory(
    event: FormEvent<HTMLFormElement>,
    variantId: string,
  ) {
    event.preventDefault();
    const form = event.currentTarget;
    const data = new FormData(form);
    await run(
      variantId,
      `/seller/variants/${variantId}/inventory`,
      {
        method: "PATCH",
        body: JSON.stringify({
          delta: Number(data.get("delta")),
          reason: data.get("reason"),
        }),
      },
      "Inventory adjusted.",
    );
    form.reset();
  }

  if (error)
    return (
      <ErrorState
        title="Product unavailable"
        message={error}
        retry={() => void load()}
      />
    );
  if (!product) return <LoadingState label="Loading product" />;
  const canPublish =
    (product.status === "draft" || product.status === "archived") &&
    product.variants.some((variant) => variant.active && variant.stock > 0) &&
    product.images.length > 0;
  const hasStock = product.variants.some(
    (variant) => variant.active && variant.stock > 0,
  );
  const readiness = [
    {
      href: "#identity",
      label: "Product details",
      detail: "Category, name, URL, description",
      complete: true,
    },
    {
      href: "#variants",
      label: "In-stock variant",
      detail: hasStock
        ? `${product.variants.length} variant${product.variants.length === 1 ? "" : "s"} ready`
        : "Add active stock before publishing",
      complete: hasStock,
    },
    {
      href: "#media",
      label: "Product image",
      detail: product.images.length
        ? `${product.images.length} of 8 images uploaded`
        : "Upload at least one image",
      complete: product.images.length > 0,
    },
  ];
  return (
    <>
      <header className="product-editor-header">
        <div className="product-editor-context">
          <Link href="/seller/products">← Back to products</Link>
          <span>Product editor</span>
        </div>
        <div className="product-editor-heading">
          <div>
            <h1>{product.name}</h1>
            <p>Updated {formatDate(product.updatedAt, "date")}</p>
          </div>
          <Status value={product.status} />
        </div>
      </header>
      {product.status === "suspended" ? (
        <div className="notice notice-danger">
          <strong>Listing suspended</strong>
          <p>
            You may update content and inventory, but only admin can reinstate
            publication.
          </p>
        </div>
      ) : null}
      {message ? (
        <p className="action-message" role="status">
          {message}
        </p>
      ) : null}
      <div className="product-editor-layout">
        <div className="product-editor-main">
          <section className="form-panel product-editor-section" id="identity">
            <h2>Product identity</h2>
            <form className="stack-form" onSubmit={updateProduct}>
              <label>
                <span>Category</span>
                <select
                  name="categoryId"
                  defaultValue={product.category.id}
                  required
                  disabled={Boolean(busy)}
                >
                  {categories.map((category) => (
                    <option key={category.id} value={category.id}>
                      {category.name}
                    </option>
                  ))}
                </select>
              </label>
              <label>
                <span>Name</span>
                <input
                  name="name"
                  defaultValue={product.name}
                  required
                  disabled={Boolean(busy)}
                />
              </label>
              <label>
                <span>Slug</span>
                <input
                  name="slug"
                  pattern="[a-z0-9]+(?:-[a-z0-9]+)*"
                  defaultValue={product.slug}
                  required
                  disabled={Boolean(busy)}
                />
              </label>
              <label>
                <span>Description</span>
                <textarea
                  name="description"
                  rows={6}
                  defaultValue={product.description}
                  required
                  disabled={Boolean(busy)}
                />
              </label>
              <button
                className="button button-primary"
                disabled={Boolean(busy)}
              >
                Save product details
              </button>
            </form>
          </section>
          <section className="form-panel product-editor-section" id="variants">
            <div className="panel-heading-row">
              <div>
                <h2>Variants and inventory</h2>
                <p>Track price and stock for each sellable option.</p>
              </div>
              <span className="count-badge">{product.variants.length}</span>
            </div>
            <p className="editor-section-note">
              Existing variants cannot be edited or deleted.
            </p>
            <div className="variant-management-list">
              {product.variants.map((variant) => (
                <article className="variant-record" key={variant.id}>
                  <div className="variant-record-copy">
                    <div className="variant-record-title">
                      <h3>{variant.name}</h3>
                      <span>{variant.sku}</span>
                    </div>
                    {Object.keys(variant.attributes).length ? (
                      <p>
                        {Object.entries(variant.attributes)
                          .map(([key, value]) => `${key}: ${value}`)
                          .join(", ")}
                      </p>
                    ) : null}
                  </div>
                  <dl className="variant-record-metrics">
                    <div>
                      <dt>Stock</dt>
                      <dd>{variant.stock}</dd>
                    </div>
                    <div>
                      <dt>Price</dt>
                      <dd>
                        <Money value={variant.priceMinor} />
                      </dd>
                    </div>
                  </dl>
                  <details className="variant-inventory-control">
                    <summary>Adjust inventory</summary>
                    <form
                      className="stack-form compact-form"
                      onSubmit={(event) =>
                        void adjustInventory(event, variant.id)
                      }
                    >
                      <label>
                        <span>Change</span>
                        <input
                          name="delta"
                          type="number"
                          required
                          min={-1000000}
                          max={1000000}
                        />
                      </label>
                      <label>
                        <span>Reason</span>
                        <input
                          name="reason"
                          minLength={2}
                          maxLength={500}
                          required
                        />
                      </label>
                      <button
                        className="button button-secondary"
                        disabled={Boolean(busy)}
                      >
                        Apply adjustment
                      </button>
                    </form>
                  </details>
                </article>
              ))}
            </div>
            <details className="management-disclosure">
              <summary>Add variant</summary>
              <form
                className="stack-form two-column-form compact-form"
                onSubmit={addVariant}
              >
                <label>
                  <span>SKU</span>
                  <input name="sku" required />
                </label>
                <label>
                  <span>Variant name</span>
                  <input name="name" required />
                </label>
                <label>
                  <span>Price (IDR)</span>
                  <input name="priceMinor" type="number" min="0" required />
                </label>
                <label>
                  <span>Initial stock</span>
                  <input name="stock" type="number" min="0" required />
                </label>
                <VariantAttributeFields />
                <button
                  className="button button-primary form-wide"
                  disabled={Boolean(busy)}
                >
                  Add variant
                </button>
              </form>
            </details>
          </section>
          <div id="media" className="editor-media-section">
            <ProductImageManager
              productId={productId}
              status={product.status}
              images={product.images}
              onChanged={load}
            />
          </div>
        </div>
        <aside className="editor-sidebar" id="publication">
          <section className="editor-readiness-panel">
            <div className="editor-sidebar-heading">
              <div>
                <span>Publication</span>
                <h2>Listing readiness</h2>
              </div>
              <strong>
                {readiness.filter((item) => item.complete).length} /{" "}
                {readiness.length}
              </strong>
            </div>
            <ol className="editor-readiness-list">
              {readiness.map((item) => (
                <li
                  className={item.complete ? "is-complete" : ""}
                  key={item.href}
                >
                  <a href={item.href}>
                    <i aria-hidden="true">{item.complete ? "✓" : "!"}</i>
                    <span>
                      <strong>{item.label}</strong>
                      <small>{item.detail}</small>
                    </span>
                  </a>
                </li>
              ))}
            </ol>
            <div className="editor-publication-action">
              {product.status === "published" ? (
                <>
                  <p>
                    Archive removes listing from marketplace until you republish
                    it.
                  </p>
                  <button
                    className="button button-secondary"
                    type="button"
                    disabled={Boolean(busy)}
                    onClick={() =>
                      void run(
                        "archive",
                        `/seller/products/${productId}/archive`,
                        {
                          method: "POST",
                        },
                        "Listing archived.",
                      )
                    }
                  >
                    {busy === "archive" ? "Archiving" : "Archive listing"}
                  </button>
                </>
              ) : product.status === "suspended" ? (
                <p>Admin reinstatement required.</p>
              ) : (
                <>
                  <p>
                    Publishing is immediate. Product needs active in-stock
                    variant and image.
                  </p>
                  <button
                    className="button button-primary"
                    type="button"
                    disabled={!canPublish || Boolean(busy)}
                    onClick={() =>
                      void run(
                        "publish",
                        `/seller/products/${productId}/publish`,
                        {
                          method: "POST",
                        },
                        product.status === "archived"
                          ? "Listing republished."
                          : "Listing published.",
                      )
                    }
                  >
                    {busy === "publish"
                      ? "Publishing"
                      : product.status === "archived"
                        ? "Republish"
                        : "Publish"}
                  </button>
                  {!canPublish ? (
                    <p className="helper-text">
                      Complete missing items above to enable publishing.
                    </p>
                  ) : null}
                </>
              )}
            </div>
          </section>
        </aside>
      </div>
    </>
  );
}

export default function SellerProductPage() {
  return (
    <main className="workspace-page shell">
      <RequireRole role="seller">
        <ProductManagementContent />
      </RequireRole>
    </main>
  );
}
