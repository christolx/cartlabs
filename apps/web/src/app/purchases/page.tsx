"use client";

import Link from "next/link";
import { useCallback, useEffect, useMemo, useState } from "react";
import type { components } from "@/lib/api/schema";
import { RequireRole } from "@/components/require-role";
import { useSession } from "@/components/session-provider";
import { EmptyState, ErrorState, LoadingState } from "@/components/async-state";
import {
  formatDate,
  Money,
  PageHeading,
  Status,
} from "@/components/marketplace-ui";
import { errorMessage } from "@/lib/api/browser";

type Purchase = components["schemas"]["Purchase"];
type PurchaseFilter = "all" | "awaiting" | "active" | "delivered" | "cancelled";

const purchaseFilters: { value: PurchaseFilter; label: string }[] = [
  { value: "all", label: "All" },
  { value: "awaiting", label: "Awaiting payment" },
  { value: "active", label: "Active" },
  { value: "delivered", label: "Delivered" },
  { value: "cancelled", label: "Cancelled" },
];

function itemCount(purchase: Purchase) {
  return purchase.sellerOrders.reduce(
    (total, order) =>
      total + order.items.reduce((count, item) => count + item.quantity, 0),
    0,
  );
}

function fulfillmentLabel(purchase: Purchase) {
  const delivered = purchase.sellerOrders.filter(
    (order) => order.status === "delivered",
  ).length;
  return `${delivered}/${purchase.sellerOrders.length} delivered`;
}

function matchesFilter(purchase: Purchase, filter: PurchaseFilter) {
  if (filter === "all") return true;
  if (filter === "awaiting") return purchase.status === "pending_payment";
  if (filter === "cancelled")
    return ["cancelled", "expired", "payment_failed"].includes(purchase.status);
  if (filter === "delivered")
    return (
      purchase.sellerOrders.length > 0 &&
      purchase.sellerOrders.every((order) => order.status === "delivered")
    );
  return (
    ["paid"].includes(purchase.status) ||
    purchase.sellerOrders.some((order) =>
      ["processing", "shipped"].includes(order.status),
    )
  );
}

function PurchasesContent() {
  const { request } = useSession();
  const [items, setItems] = useState<Purchase[] | null>(null);
  const [filter, setFilter] = useState<PurchaseFilter>("all");
  const [error, setError] = useState("");
  const load = useCallback(async () => {
    setError("");
    try {
      setItems((await request<{ items: Purchase[] }>("/purchases")).items);
    } catch (cause) {
      setError(errorMessage(cause, "Purchase history unavailable."));
    }
  }, [request]);
  useEffect(() => {
    const timer = window.setTimeout(() => void load(), 0);
    return () => window.clearTimeout(timer);
  }, [load]);
  const filtered = useMemo(
    () => (items ?? []).filter((purchase) => matchesFilter(purchase, filter)),
    [filter, items],
  );
  if (error)
    return (
      <ErrorState
        title="Purchases unavailable"
        message={error}
        retry={() => void load()}
      />
    );
  if (!items) return <LoadingState label="Loading purchases" />;
  return (
    <>
      <PageHeading
        title="Purchases"
        description="Payment and each store's fulfillment remain separate."
        family="COMMERCE / LEDGER"
        index="02"
      />
      {!items.length ? (
        <EmptyState
          title="No purchases yet"
          message="Completed checkouts appear here."
          href="/"
          action="Browse catalog"
        />
      ) : (
        <>
          <div className="ledger-toolbar">
            <div
              className="ledger-tabs"
              role="tablist"
              aria-label="Purchase filters"
            >
              {purchaseFilters.map((option) => (
                <button
                  className={`ledger-tab${filter === option.value ? " is-active" : ""}`}
                  key={option.value}
                  type="button"
                  role="tab"
                  aria-selected={filter === option.value}
                  onClick={() => setFilter(option.value)}
                >
                  {option.label}
                </button>
              ))}
            </div>
            <p className="ledger-count">
              {filtered.length} of {items.length} purchase
              {items.length === 1 ? "" : "s"}
            </p>
          </div>
          {!filtered.length ? (
            <EmptyState
              title="No purchases in this queue"
              message="Choose another status filter or complete checkout to create a purchase."
            />
          ) : (
            <div
              className="purchase-ledger"
              role="table"
              aria-label="Purchase ledger"
            >
              <div className="purchase-ledger-header" role="row">
                <span>Reference</span>
                <span>Created</span>
                <span>Payment / purchase</span>
                <span>Items</span>
                <span>Fulfillment</span>
                <span>Total</span>
              </div>
              {filtered.map((purchase) => (
                <div
                  className="purchase-ledger-row"
                  role="row"
                  key={purchase.id}
                >
                  <div role="cell">
                    <strong>{purchase.reference}</strong>
                    <small>
                      {purchase.sellerOrders.length} seller order
                      {purchase.sellerOrders.length === 1 ? "" : "s"}
                    </small>
                  </div>
                  <time role="cell" dateTime={purchase.createdAt}>
                    {formatDate(purchase.createdAt)}
                  </time>
                  <div role="cell" className="status-pair">
                    <Status value={purchase.paymentStatus} />
                    <Status value={purchase.status} />
                  </div>
                  <span role="cell">{itemCount(purchase)}</span>
                  <span role="cell">{fulfillmentLabel(purchase)}</span>
                  <div role="cell">
                    <strong>
                      <Money value={purchase.totalMinor} />
                    </strong>
                    <Link
                      className="button button-secondary"
                      href={`/purchases/${purchase.id}`}
                    >
                      View purchase
                    </Link>
                  </div>
                </div>
              ))}
            </div>
          )}
        </>
      )}
    </>
  );
}

export default function PurchasesPage() {
  return (
    <main className="workspace-page shell">
      <RequireRole role="buyer">
        <PurchasesContent />
      </RequireRole>
    </main>
  );
}
