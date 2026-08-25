"use client";

import { useCallback, useEffect, useState, type FormEvent } from "react";
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

type Order = components["schemas"]["SellerOrder"];
type NextStatus = "processing" | "shipped" | "delivered" | "cancelled";

const nextStatus: Partial<Record<Order["status"], NextStatus>> = {
  paid: "processing",
  processing: "shipped",
  shipped: "delivered",
};

function OrdersContent() {
  const { request } = useSession();
  const [items, setItems] = useState<Order[] | null>(null);
  const [error, setError] = useState("");
  const [message, setMessage] = useState("");
  const [busy, setBusy] = useState("");
  const load = useCallback(async () => {
    setError("");
    try {
      setItems((await request<{ items: Order[] }>("/seller/orders")).items);
    } catch (cause) {
      setError(errorMessage(cause, "Orders unavailable."));
    }
  }, [request]);
  useEffect(() => {
    const timer = window.setTimeout(() => void load(), 0);
    return () => window.clearTimeout(timer);
  }, [load]);

  async function update(orderId: string, status: NextStatus, reason = "") {
    setBusy(orderId);
    setMessage("");
    try {
      const next = await request<Order>(`/seller/orders/${orderId}/status`, {
        method: "PATCH",
        body: JSON.stringify({ status, reason }),
      });
      setItems(
        (current) =>
          current?.map((order) => (order.id === next.id ? next : order)) ??
          null,
      );
      setMessage(`Order moved to ${next.status.replaceAll("_", " ")}.`);
    } catch (cause) {
      setMessage(
        errorMessage(cause, "Order changed elsewhere. Server state refreshed."),
      );
      await load();
    } finally {
      setBusy("");
    }
  }

  if (error)
    return (
      <ErrorState
        title="Orders unavailable"
        message={error}
        retry={() => void load()}
      />
    );
  if (!items) return <LoadingState label="Loading seller orders" />;
  return (
    <>
      <PageHeading
        title="Fulfillment"
        description="Only orders belonging to your store appear here. Server validates every transition."
      />
      {message ? (
        <p className="action-message" role="status">
          {message}
        </p>
      ) : null}
      {!items.length ? (
        <EmptyState
          title="No seller orders"
          message="Paid buyer orders for your store appear here."
        />
      ) : (
        <div className="seller-order-cards">
          {items.map((order) => {
            const advance = nextStatus[order.status];
            const cancellable = ["paid", "processing"].includes(order.status);
            return (
              <article key={order.id}>
                <header>
                  <div>
                    <h2>{order.reference}</h2>
                    <span>
                      Purchase {order.purchaseId} /{" "}
                      {formatDate(order.createdAt)}
                    </span>
                  </div>
                  <Status value={order.status} />
                </header>
                <div className="order-items">
                  {order.items.map((item) => (
                    <div key={item.id}>
                      <div>
                        <strong>{item.productName}</strong>
                        <span>
                          {item.variantName} / {item.sku} / quantity{" "}
                          {item.quantity}
                        </span>
                      </div>
                      <Money value={item.lineTotalMinor} />
                    </div>
                  ))}
                </div>
                <footer>
                  <strong>
                    <Money value={order.subtotalMinor} />
                  </strong>
                  <div className="row-actions">
                    {advance ? (
                      <button
                        className="button button-primary"
                        type="button"
                        disabled={busy === order.id}
                        onClick={() => void update(order.id, advance)}
                      >
                        Mark {advance}
                      </button>
                    ) : null}
                    {cancellable ? (
                      <details className="inline-disclosure">
                        <summary>Cancel order</summary>
                        <form
                          className="inline-form"
                          onSubmit={(event: FormEvent<HTMLFormElement>) => {
                            event.preventDefault();
                            void update(
                              order.id,
                              "cancelled",
                              String(
                                new FormData(event.currentTarget).get("reason"),
                              ),
                            );
                          }}
                        >
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
                            className="button button-danger"
                            disabled={busy === order.id}
                          >
                            Confirm
                          </button>
                        </form>
                      </details>
                    ) : null}
                  </div>
                </footer>
                {order.cancellationReason ? (
                  <p className="moderation-note">
                    Cancellation reason: {order.cancellationReason}
                  </p>
                ) : null}
              </article>
            );
          })}
        </div>
      )}
    </>
  );
}

export default function SellerOrdersPage() {
  return (
    <main className="workspace-page shell">
      <RequireRole role="seller">
        <OrdersContent />
      </RequireRole>
    </main>
  );
}
