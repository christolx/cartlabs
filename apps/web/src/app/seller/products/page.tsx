"use client";

import Image from "next/image";
import Link from "next/link";
import { useCallback, useEffect, useMemo, useState } from "react";
import type { components } from "@/lib/api/schema";
import { RequireRole } from "@/components/require-role";
import { useSession } from "@/components/session-provider";
import { EmptyState, ErrorState, LoadingState } from "@/components/async-state";
import { Money, PageHeading, Status } from "@/components/marketplace-ui";
import { errorMessage } from "@/lib/api/browser";

type Product = components["schemas"]["ProductDetail"];

function ProductsContent() {
  const { request } = useSession();
  const [items, setItems] = useState<Product[] | null>(null);
  const [filter, setFilter] = useState("all");
  const [query, setQuery] = useState("");
  const [error, setError] = useState("");
  const load = useCallback(async () => {
    setError("");
    try {
      setItems((await request<{ items: Product[] }>("/seller/products")).items);
    } catch (cause) {
      setError(errorMessage(cause, "Products unavailable."));
    }
  }, [request]);
  useEffect(() => {
    const timer = window.setTimeout(() => void load(), 0);
    return () => window.clearTimeout(timer);
  }, [load]);
  const filtered = useMemo(() => {
    const normalized = query.trim().toLocaleLowerCase();
    return (items ?? []).filter(
      (product) =>
        (filter === "all" || product.status === filter) &&
        (!normalized ||
          `${product.name} ${product.slug}`
            .toLocaleLowerCase()
            .includes(normalized)),
    );
  }, [filter, items, query]);
  const lifecycle = (items ?? []).reduce<Record<string, number>>(
    (totals, product) => ({
      ...totals,
      [product.status]: (totals[product.status] ?? 0) + 1,
    }),
    {},
  );
  if (error)
    return (
      <ErrorState
        title="Products unavailable"
        message={error}
        retry={() => void load()}
      />
    );
  if (!items) return <LoadingState label="Loading seller products" />;
  return (
    <>
      <PageHeading
        title="Products"
        description="Owned product lifecycle, variants, and current stock."
        family="SELLER / CATALOG"
        index="02"
        meta={
          <span>
            {filtered.length} of {items.length} listings
          </span>
        }
        actions={
          <Link className="button button-primary" href="/seller/products/new">
            Create draft
          </Link>
        }
      />
      {!items.length ? (
        <EmptyState
          title="No products"
          message="Verified store owners can create first product draft."
          href="/seller/products/new"
          action="Create draft"
        />
      ) : (
        <>
          <div
            className="product-status-tabs"
            role="tablist"
            aria-label="Product lifecycle"
          >
            {["all", "draft", "published", "archived", "suspended"].map(
              (status) => (
                <button
                  className={filter === status ? "is-active" : ""}
                  key={status}
                  type="button"
                  role="tab"
                  aria-selected={filter === status}
                  onClick={() => setFilter(status)}
                >
                  {status}{" "}
                  <b>
                    {status === "all" ? items.length : (lifecycle[status] ?? 0)}
                  </b>
                </button>
              ),
            )}
          </div>
          <span className="tab-scroll-hint" aria-hidden="true">
            More lifecycle filters available
          </span>
          <div className="moderation-filter-band seller-filter-band">
            <label>
              <span>Search products</span>
              <input
                type="search"
                value={query}
                onChange={(event) => setQuery(event.target.value)}
                placeholder="Name or slug"
              />
            </label>
          </div>
          {!filtered.length ? (
            <EmptyState
              title="No products match"
              message="Change lifecycle or search filters."
            />
          ) : (
            <div className="seller-product-list">
              {filtered.map((product) => {
                const stock = product.variants.reduce(
                  (total, variant) => total + variant.stock,
                  0,
                );
                const lowest = product.variants.reduce<number | null>(
                  (value, variant) =>
                    value === null || variant.priceMinor < value
                      ? variant.priceMinor
                      : value,
                  null,
                );
                return (
                  <article key={product.id}>
                    {product.images[0] ? (
                      <Image
                        src={product.images[0].url}
                        alt={product.images[0].altText}
                        width={120}
                        height={90}
                      />
                    ) : (
                      <div className="line-image-fallback">No image</div>
                    )}
                    <div className="seller-product-copy">
                      <div className="heading-status">
                        <h2>{product.name}</h2>
                        <Status value={product.status} />
                      </div>
                      <p>
                        {product.category.name}. {product.variants.length}{" "}
                        variants. {stock} units in stock
                        {lowest !== null ? (
                          <>
                            . From <Money value={lowest} />
                          </>
                        ) : null}
                        .
                      </p>
                      <Link
                        className="text-link"
                        href={`/seller/products/${product.id}`}
                      >
                        Manage product
                      </Link>
                    </div>
                  </article>
                );
              })}
            </div>
          )}
        </>
      )}
    </>
  );
}

export default function SellerProductsPage() {
  return (
    <main className="workspace-page shell">
      <RequireRole role="seller">
        <ProductsContent />
      </RequireRole>
    </main>
  );
}
