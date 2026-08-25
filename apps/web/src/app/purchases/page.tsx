"use client";

import Link from "next/link";
import { useCallback, useEffect, useState } from "react";
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

function PurchasesContent() {
  const { request } = useSession();
  const [items, setItems] = useState<Purchase[] | null>(null);
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
      />
      {!items.length ? (
        <EmptyState
          title="No purchases yet"
          message="Completed checkouts appear here."
          href="/"
          action="Browse catalog"
        />
      ) : (
        <div className="record-grid">
          {items.map((purchase) => (
            <article className="record-card" key={purchase.id}>
              <div className="record-card-heading">
                <div>
                  <h2>{purchase.reference}</h2>
                  <time dateTime={purchase.createdAt}>
                    {formatDate(purchase.createdAt)}
                  </time>
                </div>
                <Status value={purchase.status} />
              </div>
              <div className="status-pair">
                <span>
                  Payment <Status value={purchase.paymentStatus} />
                </span>
                <span>
                  {purchase.sellerOrders.length} seller order
                  {purchase.sellerOrders.length === 1 ? "" : "s"}
                </span>
              </div>
              <ul className="compact-status-list">
                {purchase.sellerOrders.map((order) => (
                  <li key={order.id}>
                    <span>{order.storeName}</span>
                    <Status value={order.status} />
                  </li>
                ))}
              </ul>
              <footer>
                <strong>
                  <Money value={purchase.totalMinor} />
                </strong>
                <Link
                  className="button button-secondary"
                  href={`/purchases/${purchase.id}`}
                >
                  View purchase
                </Link>
              </footer>
            </article>
          ))}
        </div>
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
