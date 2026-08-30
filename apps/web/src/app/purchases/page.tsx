"use client";

import Image from "next/image";
import Link from "next/link";
import { useCallback, useEffect, useMemo, useState } from "react";
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
import {
  fulfillmentLabel,
  itemCount,
  matchesPurchaseFilter,
  matchesPurchaseQuery,
  purchaseFilters,
  purchaseItems,
  type Purchase,
  type PurchaseFilter,
} from "@/lib/purchase-filters";

function PurchasesContent() {
  const { request } = useSession();
  const [items, setItems] = useState<Purchase[] | null>(null);
  const [filter, setFilter] = useState<PurchaseFilter>("all");
  const [query, setQuery] = useState("");
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
    () =>
      (items ?? []).filter(
        (purchase) =>
          matchesPurchaseFilter(purchase, filter) &&
          matchesPurchaseQuery(purchase, query),
      ),
    [filter, items, query],
  );
  const counts = useMemo(
    () =>
      Object.fromEntries(
        purchaseFilters.map((option) => [
          option.value,
          (items ?? []).filter((purchase) =>
            matchesPurchaseFilter(purchase, option.value),
          ).length,
        ]),
      ) as Record<PurchaseFilter, number>,
    [items],
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
                  {option.label} <b>{counts[option.value] ?? 0}</b>
                </button>
              ))}
            </div>
            <p className="ledger-count">
              {filtered.length} of {items.length} purchase
              {items.length === 1 ? "" : "s"}
            </p>
          </div>
          <div className="workspace-filter-band purchase-filter-band">
            <label>
              <span>Search purchases</span>
              <input
                type="search"
                value={query}
                onChange={(event) => setQuery(event.target.value)}
                placeholder="Reference, product, or store"
              />
            </label>
            {query ? (
              <button
                className="button button-secondary filter-clear"
                type="button"
                onClick={() => setQuery("")}
              >
                Clear search
              </button>
            ) : null}
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
                  <div role="cell" className="purchase-item-summary">
                    {purchaseItems(purchase)[0]?.imageUrl ? (
                      <span className="purchase-item-thumb">
                        <Image
                          src={purchaseItems(purchase)[0].imageUrl}
                          alt=""
                          fill
                          sizes="56px"
                        />
                      </span>
                    ) : (
                      <span className="purchase-item-thumb-fallback">
                        No image
                      </span>
                    )}
                    <span>
                      <strong>
                        {purchaseItems(purchase)[0]?.productName ??
                          "No item details"}
                      </strong>
                      <small>
                        {itemCount(purchase)} item
                        {itemCount(purchase) === 1 ? "" : "s"}
                        {purchaseItems(purchase).length > 1
                          ? ` / +${purchaseItems(purchase).length - 1} more`
                          : ""}
                      </small>
                    </span>
                  </div>
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
