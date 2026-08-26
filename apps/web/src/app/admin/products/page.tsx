"use client";

import Image from "next/image";
import {
  useCallback,
  useEffect,
  useMemo,
  useState,
  type FormEvent,
} from "react";
import type { components } from "@/lib/api/schema";
import { RequireRole } from "@/components/require-role";
import { useSession } from "@/components/session-provider";
import { EmptyState, ErrorState, LoadingState } from "@/components/async-state";
import {
  Money,
  PageHeading,
  Status,
  formatDate,
} from "@/components/marketplace-ui";
import { errorMessage } from "@/lib/api/browser";

type Product = components["schemas"]["AdminProduct"];
type EnforcementStatus = "suspended" | "published";

function totalStock(product: Product) {
  return product.variants.reduce((total, variant) => total + variant.stock, 0);
}

function priceRange(product: Product) {
  const prices = product.variants.map((variant) => variant.priceMinor);
  return { min: Math.min(...prices), max: Math.max(...prices) };
}

function ProductEnforcementContent() {
  const { request } = useSession();
  const [items, setItems] = useState<Product[] | null>(null);
  const [query, setQuery] = useState("");
  const [statusFilter, setStatusFilter] = useState("");
  const [stockFilter, setStockFilter] = useState("");
  const [error, setError] = useState("");
  const [message, setMessage] = useState("");
  const [busy, setBusy] = useState("");

  const load = useCallback(async () => {
    setError("");
    try {
      setItems((await request<{ items: Product[] }>("/admin/products")).items);
    } catch (cause) {
      setError(errorMessage(cause, "Listing enforcement unavailable."));
    }
  }, [request]);

  useEffect(() => {
    const timer = window.setTimeout(() => void load(), 0);
    return () => window.clearTimeout(timer);
  }, [load]);

  const filtered = useMemo(() => {
    const normalized = query.trim().toLocaleLowerCase();
    return (items ?? []).filter((product) => {
      const stock = totalStock(product);
      return (
        (!normalized ||
          `${product.name} ${product.storeName ?? ""} ${product.slug}`
            .toLocaleLowerCase()
            .includes(normalized)) &&
        (!statusFilter || product.status === statusFilter) &&
        (!stockFilter || (stockFilter === "in" ? stock > 0 : stock === 0))
      );
    });
  }, [items, query, statusFilter, stockFilter]);

  async function updateStatus(
    event: FormEvent<HTMLFormElement>,
    productId: string,
    status: EnforcementStatus,
  ) {
    event.preventDefault();
    const form = event.currentTarget;
    const data = new FormData(form);
    setBusy(productId);
    setMessage("");
    try {
      const next = await request<Product>(
        `/admin/products/${productId}/status`,
        {
          method: "PATCH",
          body: JSON.stringify({ status, reason: data.get("reason") }),
        },
      );
      setItems(
        (current) =>
          current?.map((item) => (item.id === next.id ? next : item)) ?? null,
      );
      setMessage(`${next.name} is ${next.status}.`);
      form.reset();
      form.closest("details")?.removeAttribute("open");
    } catch (cause) {
      setMessage(
        errorMessage(
          cause,
          "Enforcement action failed. Refreshing server state.",
        ),
      );
      await load();
    } finally {
      setBusy("");
    }
  }

  if (error)
    return (
      <ErrorState
        title="Listing enforcement unavailable"
        message={error}
        retry={() => void load()}
      />
    );
  if (!items) return <LoadingState label="Loading listings" />;

  return (
    <>
      <PageHeading
        title="Listing index"
        description="Inspect marketplace supply in compact records. Open one listing for moderation detail and enforcement history."
        family="ADMIN / PRODUCTS"
        index="02"
        meta={
          <span>
            {filtered.length} of {items.length} loaded listings
          </span>
        }
      />
      {message ? (
        <p className="action-message" role="status">
          {message}
        </p>
      ) : null}
      <div className="moderation-filter-band">
        <label>
          <span>Search listings</span>
          <input
            type="search"
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            placeholder="Product, store, slug"
          />
        </label>
        <label>
          <span>Status</span>
          <select
            value={statusFilter}
            onChange={(event) => setStatusFilter(event.target.value)}
          >
            <option value="">All statuses</option>
            <option value="published">Published</option>
            <option value="suspended">Suspended</option>
          </select>
        </label>
        <label>
          <span>Stock</span>
          <select
            value={stockFilter}
            onChange={(event) => setStockFilter(event.target.value)}
          >
            <option value="">All stock</option>
            <option value="in">In stock</option>
            <option value="out">Out of stock</option>
          </select>
        </label>
      </div>
      {!filtered.length ? (
        <EmptyState
          title="No listings match"
          message="Change search or moderation filters."
        />
      ) : (
        <div className="admin-product-records" role="list">
          <div className="admin-product-records-head" aria-hidden="true">
            <span>Listing</span>
            <span>Store / category</span>
            <span>Status</span>
            <span>Supply</span>
            <span>Price</span>
            <span>Inspect</span>
          </div>
          {filtered.map((product) => {
            const stock = totalStock(product);
            const range = priceRange(product);
            const target: EnforcementStatus =
              product.status === "published" ? "suspended" : "published";
            return (
              <details className="admin-product-record" key={product.id}>
                <summary>
                  {product.images[0] ? (
                    <Image
                      src={product.images[0].url}
                      alt=""
                      width={64}
                      height={52}
                    />
                  ) : (
                    <span className="admin-record-image-fallback">--</span>
                  )}
                  <span className="admin-record-title">
                    <strong>{product.name}</strong>
                    <small>{product.slug}</small>
                  </span>
                  <span className="admin-record-store">
                    <strong>{product.storeName || "Seller store"}</strong>
                    <small>{product.category.name}</small>
                  </span>
                  <Status value={product.status} />
                  <span className="admin-record-data">
                    {product.variants.length} variants / {stock} units
                  </span>
                  <span className="admin-record-data admin-record-price">
                    <Money value={range.min} />
                    {range.max !== range.min ? (
                      <>
                        {" "}
                        – <Money value={range.max} />
                      </>
                    ) : null}
                  </span>
                  <span className="admin-record-open">Inspect +</span>
                </summary>
                <div className="admin-product-detail">
                  <div className="admin-product-detail-copy">
                    <p>{product.description}</p>
                    <small>
                      Updated {formatDate(product.updatedAt)}. Created{" "}
                      {formatDate(product.createdAt)}.
                    </small>
                    {product.enforcementReason ? (
                      <p className="moderation-note">
                        Latest enforcement: {product.enforcementReason}
                        {product.enforcedBy ? ` / ${product.enforcedBy}` : ""}
                        {product.enforcedAt
                          ? ` / ${formatDate(product.enforcedAt)}`
                          : ""}
                      </p>
                    ) : null}
                  </div>
                  <div className="moderation-supply">
                    <section>
                      <h3>Variants</h3>
                      {product.variants.map((variant) => (
                        <div key={variant.id}>
                          <div>
                            <strong>{variant.name}</strong>
                            <span>
                              {variant.sku}. {variant.stock} in stock.
                            </span>
                          </div>
                          <Money
                            value={variant.priceMinor}
                            currency={variant.currency}
                          />
                        </div>
                      ))}
                    </section>
                    <section>
                      <h3>Images</h3>
                      <div className="moderation-images">
                        {product.images.length ? (
                          product.images.map((image) => (
                            <figure key={image.id}>
                              <Image
                                src={image.url}
                                alt={image.altText}
                                width={140}
                                height={92}
                              />
                              <figcaption>
                                {image.altText || "No alt text"}
                              </figcaption>
                            </figure>
                          ))
                        ) : (
                          <span className="helper-text">
                            No product images.
                          </span>
                        )}
                      </div>
                    </section>
                  </div>
                  <div className="moderation-actions">
                    <details>
                      <summary>
                        {target === "suspended"
                          ? "Suspend listing"
                          : "Reinstate listing"}
                      </summary>
                      <form
                        className="stack-form compact-form"
                        onSubmit={(event) =>
                          void updateStatus(event, product.id, target)
                        }
                      >
                        <p>
                          {target === "suspended"
                            ? "Listing becomes unavailable to catalog, cart additions, and checkout."
                            : "Store approval and product completeness are revalidated."}
                        </p>
                        <label>
                          <span>Reason</span>
                          <textarea
                            name="reason"
                            minLength={1}
                            maxLength={500}
                            rows={3}
                            required
                          />
                        </label>
                        <button
                          className={
                            target === "suspended"
                              ? "button button-danger"
                              : "button button-primary"
                          }
                          disabled={busy === product.id}
                        >
                          {busy === product.id ? "Saving" : `Confirm ${target}`}
                        </button>
                      </form>
                    </details>
                  </div>
                </div>
              </details>
            );
          })}
        </div>
      )}
    </>
  );
}

export default function AdminProductsPage() {
  return (
    <main className="workspace-page shell">
      <RequireRole role="admin">
        <ProductEnforcementContent />
      </RequireRole>
    </main>
  );
}
