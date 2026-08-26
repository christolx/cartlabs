"use client";

import Image from "next/image";
import Link from "next/link";
import { useCallback, useEffect, useState, type FormEvent } from "react";
import { useParams } from "next/navigation";
import type { components } from "@/lib/api/schema";
import { RequireRole } from "@/components/require-role";
import { useSession } from "@/components/session-provider";
import { ErrorState, LoadingState } from "@/components/async-state";
import {
  formatDate,
  Money,
  PageHeading,
  Status,
} from "@/components/marketplace-ui";
import { errorMessage } from "@/lib/api/browser";

type Purchase = components["schemas"]["Purchase"];
type Review = components["schemas"]["Review"];

function PurchaseContent() {
  const { purchaseId } = useParams<{ purchaseId: string }>();
  const { demoEnabled, request } = useSession();
  const [purchase, setPurchase] = useState<Purchase | null>(null);
  const [error, setError] = useState("");
  const [message, setMessage] = useState("");
  const [busy, setBusy] = useState("");
  const [reviewed, setReviewed] = useState<Set<string>>(new Set());
  const load = useCallback(async () => {
    setError("");
    try {
      setPurchase(
        await request<Purchase>(`/purchases/${encodeURIComponent(purchaseId)}`),
      );
    } catch (cause) {
      setError(errorMessage(cause, "Purchase unavailable."));
    }
  }, [purchaseId, request]);
  useEffect(() => {
    const timer = window.setTimeout(() => void load(), 0);
    return () => window.clearTimeout(timer);
  }, [load]);

  async function mutate(label: string, path: string, body: object) {
    setBusy(label);
    setMessage("");
    try {
      setPurchase(
        await request<Purchase>(path, {
          method: "POST",
          body: JSON.stringify(body),
        }),
      );
      setMessage("Purchase updated from server state.");
    } catch (cause) {
      setMessage(errorMessage(cause, "Action failed."));
      await load();
    } finally {
      setBusy("");
    }
  }

  async function review(
    event: FormEvent<HTMLFormElement>,
    purchaseItemId: string,
  ) {
    event.preventDefault();
    const form = event.currentTarget;
    const data = new FormData(form);
    setBusy(purchaseItemId);
    setMessage("");
    try {
      await request<Review>("/reviews", {
        method: "POST",
        body: JSON.stringify({
          purchaseItemId,
          rating: Number(data.get("rating")),
          title: data.get("title"),
          body: data.get("body"),
        }),
      });
      setReviewed((current) => new Set(current).add(purchaseItemId));
      setMessage("Verified review published.");
    } catch (cause) {
      setMessage(errorMessage(cause, "Review failed."));
    } finally {
      setBusy("");
    }
  }

  if (error)
    return (
      <ErrorState
        title="Purchase unavailable"
        message={error}
        retry={() => void load()}
      />
    );
  if (!purchase) return <LoadingState label="Loading purchase" />;
  const paymentEligible = demoEnabled && purchase.status === "pending_payment";
  const cancellable =
    ["pending_payment", "paid"].includes(purchase.status) &&
    purchase.sellerOrders.every((order) =>
      ["pending_payment", "paid"].includes(order.status),
    );
  return (
    <>
      <Link className="purchase-detail-back" href="/purchases">
        ← Back to purchases
      </Link>
      <PageHeading
        title={purchase.reference}
        description={`Created ${formatDate(purchase.createdAt)}. Reservation expires ${formatDate(purchase.reservationExpiresAt)}.`}
        family="COMMERCE / PURCHASE"
        index="03"
        meta={
          <>
            <span>Payment intent {purchase.paymentIntentId}</span>
            <span>
              {purchase.sellerOrders.length} seller order
              {purchase.sellerOrders.length === 1 ? "" : "s"}
            </span>
          </>
        }
        actions={
          <div className="status-pair">
            <Status value={purchase.status} />
            <Status value={purchase.paymentStatus} />
          </div>
        }
      />
      {message ? (
        <p className="action-message" role="status">
          {message}
        </p>
      ) : null}
      <div className="purchase-detail-layout">
        <div className="purchase-orders">
          {purchase.sellerOrders.map((order) => (
            <section className="purchase-order" key={order.id}>
              <header>
                <div>
                  <h2>
                    <span className="section-index">
                      /
                      {String(
                        purchase.sellerOrders.indexOf(order) + 1,
                      ).padStart(2, "0")}
                    </span>{" "}
                    {order.storeName}
                  </h2>
                  <span>{order.reference}</span>
                </div>
                <Status value={order.status} />
              </header>
              <div className="timeline">
                <span>Created {formatDate(order.createdAt)}</span>
                {order.processingAt ? (
                  <span>Processing {formatDate(order.processingAt)}</span>
                ) : null}
                {order.shippedAt ? (
                  <span>Shipped {formatDate(order.shippedAt)}</span>
                ) : null}
                {order.deliveredAt ? (
                  <span>Delivered {formatDate(order.deliveredAt)}</span>
                ) : null}
                {order.cancelledAt ? (
                  <span>
                    Cancelled {formatDate(order.cancelledAt)}:{" "}
                    {order.cancellationReason}
                  </span>
                ) : null}
              </div>
              {order.items.map((item) => (
                <article className="purchase-item" key={item.id}>
                  {item.imageUrl ? (
                    <Image src={item.imageUrl} alt="" width={88} height={66} />
                  ) : (
                    <div className="line-image-fallback">No image</div>
                  )}
                  <div>
                    <h3>{item.productName}</h3>
                    <p>
                      {item.variantName} / {item.sku} / quantity {item.quantity}
                    </p>
                    <strong>
                      <Money value={item.lineTotalMinor} />
                    </strong>
                  </div>
                  {order.status === "delivered" && !reviewed.has(item.id) ? (
                    <details className="review-disclosure">
                      <summary>Write review</summary>
                      <form
                        className="stack-form compact-form"
                        onSubmit={(event) => void review(event, item.id)}
                      >
                        <label>
                          <span>Rating</span>
                          <select name="rating" defaultValue="5" required>
                            <option value="5">5</option>
                            <option value="4">4</option>
                            <option value="3">3</option>
                            <option value="2">2</option>
                            <option value="1">1</option>
                          </select>
                        </label>
                        <label>
                          <span>Title</span>
                          <input
                            name="title"
                            minLength={2}
                            maxLength={120}
                            required
                          />
                        </label>
                        <label>
                          <span>Review</span>
                          <textarea
                            name="body"
                            minLength={2}
                            maxLength={2000}
                            rows={3}
                            required
                          />
                        </label>
                        <button
                          className="button button-primary"
                          disabled={busy === item.id}
                        >
                          Publish review
                        </button>
                      </form>
                    </details>
                  ) : null}
                </article>
              ))}
            </section>
          ))}
        </div>
        <aside className="order-summary">
          <h2>Purchase total</h2>
          <dl>
            <div>
              <dt>Subtotal</dt>
              <dd>
                <Money value={purchase.subtotalMinor} />
              </dd>
            </div>
            <div>
              <dt>Total</dt>
              <dd>
                <Money value={purchase.totalMinor} />
              </dd>
            </div>
          </dl>
          {paymentEligible ? (
            <div className="stack-actions">
              <button
                className="button button-primary"
                disabled={Boolean(busy)}
                onClick={() =>
                  void mutate("pay-success", `/purchases/${purchase.id}/pay`, {
                    outcome: "succeeded",
                  })
                }
              >
                Complete demo payment
              </button>
              <button
                className="button button-secondary"
                disabled={Boolean(busy)}
                onClick={() =>
                  void mutate("pay-fail", `/purchases/${purchase.id}/pay`, {
                    outcome: "failed",
                  })
                }
              >
                Simulate failed payment
              </button>
            </div>
          ) : null}
          {cancellable ? (
            <details className="cancel-disclosure">
              <summary>Cancel purchase</summary>
              <form
                className="stack-form compact-form"
                onSubmit={(event) => {
                  event.preventDefault();
                  const reason = String(
                    new FormData(event.currentTarget).get("reason"),
                  );
                  void mutate("cancel", `/purchases/${purchase.id}/cancel`, {
                    reason,
                  });
                }}
              >
                <label>
                  <span>Reason</span>
                  <textarea
                    name="reason"
                    minLength={2}
                    maxLength={500}
                    required
                  />
                </label>
                <button
                  className="button button-danger"
                  disabled={Boolean(busy)}
                >
                  Confirm cancellation
                </button>
              </form>
            </details>
          ) : null}
        </aside>
      </div>
    </>
  );
}

export default function PurchasePage() {
  return (
    <main className="workspace-page shell">
      <RequireRole role="buyer">
        <PurchaseContent />
      </RequireRole>
    </main>
  );
}
