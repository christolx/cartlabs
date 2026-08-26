"use client";

import Link from "next/link";
import { useCallback, useEffect, useState } from "react";
import type { components } from "@/lib/api/schema";
import { RequireRole } from "@/components/require-role";
import { useSession } from "@/components/session-provider";
import { ErrorState, LoadingState } from "@/components/async-state";
import { Money, PageHeading, Status } from "@/components/marketplace-ui";
import { BrowserAPIError, errorMessage } from "@/lib/api/browser";

type Store = components["schemas"]["Store"];
type Product = components["schemas"]["ProductDetail"];
type Order = components["schemas"]["SellerOrder"];

function SellerHomeContent() {
  const { request } = useSession();
  const [data, setData] = useState<{
    store: Store | null;
    products: Product[];
    orders: Order[];
  } | null>(null);
  const [error, setError] = useState("");
  const load = useCallback(async () => {
    setError("");
    try {
      const [store, products, orders] = await Promise.all([
        request<Store>("/seller/store").catch((cause) =>
          cause instanceof BrowserAPIError && cause.status === 404
            ? null
            : Promise.reject(cause),
        ),
        request<{ items: Product[] }>("/seller/products"),
        request<{ items: Order[] }>("/seller/orders"),
      ]);
      setData({ store, products: products.items, orders: orders.items });
    } catch (cause) {
      setError(errorMessage(cause, "Seller workspace unavailable."));
    }
  }, [request]);
  useEffect(() => {
    const timer = window.setTimeout(() => void load(), 0);
    return () => window.clearTimeout(timer);
  }, [load]);
  if (error)
    return (
      <ErrorState
        title="Seller workspace unavailable"
        message={error}
        retry={() => void load()}
      />
    );
  if (!data) return <LoadingState label="Loading seller workspace" />;
  const lifecycle = data.products.reduce<Record<string, number>>(
    (totals, product) => ({
      ...totals,
      [product.status]: (totals[product.status] ?? 0) + 1,
    }),
    {},
  );
  const actionable = data.orders.filter((order) =>
    ["paid", "processing", "shipped"].includes(order.status),
  );
  const readiness = [
    {
      label: "Store approved",
      complete: data.store?.status === "approved",
      href: "/seller/store",
      note: data.store
        ? data.store.status.replaceAll("_", " ")
        : "Create store",
    },
    {
      label: "Product identity",
      complete: data.products.length > 0,
      href: "/seller/products",
      note: `${data.products.length} product${data.products.length === 1 ? "" : "s"}`,
    },
    {
      label: "Variant and stock",
      complete: data.products.some((product) =>
        product.variants.some((variant) => variant.active && variant.stock > 0),
      ),
      href: "/seller/products",
      note: "Active inventory",
    },
    {
      label: "Product media",
      complete: data.products.some((product) => product.images.length > 0),
      href: "/seller/products",
      note: "At least one image",
    },
    {
      label: "Published listing",
      complete: data.products.some((product) => product.status === "published"),
      href: "/seller/products",
      note: `${lifecycle.published ?? 0} published`,
    },
  ];
  return (
    <>
      <PageHeading
        title="Seller home"
        description="Store readiness, catalog supply, and owned fulfillment."
        family="SELLER / OPERATIONS"
        index="01"
        meta={
          <span>
            {actionable.length} order{actionable.length === 1 ? "" : "s"}{" "}
            needing action
          </span>
        }
      />
      <div className="seller-home-grid">
        <section className="workspace-panel seller-readiness">
          <div className="heading-status">
            <div>
              <h2>Store readiness</h2>
              <p className="helper-text">
                {data.store?.name ?? "No store created"}
              </p>
            </div>
            {data.store ? <Status value={data.store.status} /> : null}
          </div>
          {data.store?.status === "rejected" ? (
            <p className="moderation-note">
              Verification note:{" "}
              {data.store.moderationNote || "No note provided."}
            </p>
          ) : null}
          <div
            className="readiness-list"
            aria-label="Store publishing readiness"
          >
            {readiness.map((item) => (
              <Link
                className={item.complete ? "is-complete" : ""}
                href={item.href}
                key={item.label}
              >
                <i aria-hidden="true">{item.complete ? "✓" : "!"}</i>
                <strong>{item.label}</strong>
                <small>{item.note}</small>
              </Link>
            ))}
          </div>
          <Link
            className="button button-primary"
            href={data.store ? "/seller/products" : "/seller/store"}
          >
            {data.store ? "Review next step" : "Create store"}
          </Link>
        </section>
        <section className="workspace-panel">
          <h2>Product supply</h2>
          <div className="metric-row">
            <div>
              <strong>{data.products.length}</strong>
              <span>Total</span>
            </div>
            <div>
              <strong>{lifecycle.draft ?? 0}</strong>
              <span>Draft</span>
            </div>
            <div>
              <strong>{lifecycle.published ?? 0}</strong>
              <span>Published</span>
            </div>
            <div>
              <strong>{lifecycle.archived ?? 0}</strong>
              <span>Archived</span>
            </div>
            <div>
              <strong>{lifecycle.suspended ?? 0}</strong>
              <span>Suspended</span>
            </div>
          </div>
          <Link
            className="button button-secondary"
            href={
              data.store?.status === "approved"
                ? "/seller/products/new"
                : "/seller/products"
            }
          >
            {data.store?.status === "approved"
              ? "Create product"
              : "Review products"}
          </Link>
        </section>
        <section className="workspace-panel workspace-wide">
          <div className="panel-heading-row">
            <div>
              <h2>Orders needing action</h2>
              <p>Paid, processing, and shipped orders.</p>
            </div>
            <Link className="text-link" href="/seller/orders">
              All orders
            </Link>
          </div>
          {actionable.length ? (
            <div className="record-list">
              {actionable.slice(0, 5).map((order) => (
                <div className="record-row" key={order.id}>
                  <div>
                    <strong>{order.reference}</strong>
                    <span>
                      {order.items.length} items from purchase{" "}
                      {order.purchaseId}
                    </span>
                  </div>
                  <div className="row-actions">
                    <Status value={order.status} />
                    <strong>
                      <Money value={order.subtotalMinor} />
                    </strong>
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <p className="empty-copy">
              No orders currently need seller action.
            </p>
          )}
        </section>
      </div>
    </>
  );
}

export default function SellerHome() {
  return (
    <main className="workspace-page shell">
      <RequireRole role="seller">
        <SellerHomeContent />
      </RequireRole>
    </main>
  );
}
