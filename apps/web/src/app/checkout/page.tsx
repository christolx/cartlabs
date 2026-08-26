"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { useRouter } from "next/navigation";
import type { components } from "@/lib/api/schema";
import { RequireRole } from "@/components/require-role";
import { useSession } from "@/components/session-provider";
import { EmptyState, ErrorState, LoadingState } from "@/components/async-state";
import { Money, PageHeading } from "@/components/marketplace-ui";
import { errorMessage } from "@/lib/api/browser";

type Cart = components["schemas"]["Cart"];
type Purchase = components["schemas"]["Purchase"];

function CheckoutContent() {
  const { request } = useSession();
  const router = useRouter();
  const keyRef = useRef("");
  const [cart, setCart] = useState<Cart | null>(null);
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const [message, setMessage] = useState("");
  const load = useCallback(async () => {
    setError("");
    try {
      setCart(await request<Cart>("/cart"));
    } catch (cause) {
      setError(errorMessage(cause, "Checkout unavailable."));
    }
  }, [request]);
  useEffect(() => {
    const timer = window.setTimeout(() => void load(), 0);
    return () => window.clearTimeout(timer);
  }, [load]);

  async function submit() {
    if (busy) return;
    if (!keyRef.current) keyRef.current = crypto.randomUUID();
    setBusy(true);
    setMessage("");
    try {
      const purchase = await request<Purchase>("/checkout", {
        method: "POST",
        headers: { "Idempotency-Key": keyRef.current },
      });
      window.dispatchEvent(
        new CustomEvent("cart:updated", { detail: { quantity: 0 } }),
      );
      router.replace(`/purchases/${purchase.id}`);
    } catch (cause) {
      setMessage(
        errorMessage(cause, "Checkout failed. Retry uses same submission key."),
      );
      await load();
    } finally {
      setBusy(false);
    }
  }

  if (error)
    return (
      <ErrorState
        title="Checkout unavailable"
        message={error}
        retry={() => void load()}
      />
    );
  if (!cart) return <LoadingState label="Loading checkout" />;
  if (!cart.totalQuantity)
    return (
      <>
        <PageHeading
          title="Checkout"
          description="Confirm current cart before purchase creation."
          family="COMMERCE / CHECKOUT"
          index="01"
        />
        <EmptyState
          title="Nothing to check out"
          message="Cart is empty or was already submitted."
          href="/"
          action="Browse catalog"
        />
      </>
    );
  const invalid = cart.stores.some((store) =>
    store.items.some(
      (item) => item.quantity < 1 || item.quantity > item.availableStock,
    ),
  );
  return (
    <>
      <PageHeading
        title="Confirm checkout"
        description="One purchase is created. Each store fulfills its own seller order."
        family="COMMERCE / CHECKOUT"
        index="01"
      />
      <div className="checkout-steps" aria-label="Checkout progress">
        <div className="checkout-step is-active">
          <strong>01</strong>
          <span>Review bag</span>
        </div>
        <div className="checkout-step">
          <strong>02</strong>
          <span>Purchase created</span>
        </div>
      </div>
      <div className="checkout-layout">
        <div className="checkout-groups">
          {cart.stores.map((store) => (
            <section key={store.storeId}>
              <h2>{store.storeName}</h2>
              {store.items.map((item) => (
                <div
                  className={`checkout-line${item.quantity > item.availableStock ? " is-invalid" : ""}`}
                  key={item.variantId}
                >
                  <div>
                    <strong>{item.productName}</strong>
                    <span>
                      {item.variantName} x {item.quantity}
                    </span>
                    {item.quantity > item.availableStock ? (
                      <span className="error-message">
                        Stock changed: {item.availableStock} available.
                      </span>
                    ) : null}
                  </div>
                  <Money value={item.lineTotalMinor} />
                </div>
              ))}
              <div className="checkout-subtotal">
                <span>Store subtotal</span>
                <strong>
                  <Money value={store.subtotalMinor} />
                </strong>
              </div>
            </section>
          ))}
        </div>
        <aside className="order-summary">
          <h2>Final total</h2>
          <dl>
            <div>
              <dt>Items</dt>
              <dd>{cart.totalQuantity}</dd>
            </div>
            <div>
              <dt>Total</dt>
              <dd>
                <Money value={cart.subtotalMinor} />
              </dd>
            </div>
          </dl>
          <p className="notice commerce-notice">
            <strong>Demo checkout</strong>
            <br />
            No address, shipping quote, tax, or real payment is collected in
            this demo.
          </p>
          {message ? (
            <p className="form-message error-message" role="alert">
              {message}
            </p>
          ) : null}
          <button
            className="button button-primary"
            type="button"
            disabled={busy || invalid}
            onClick={() => void submit()}
          >
            {busy ? "Confirming purchase" : "Confirm purchase"}
          </button>
          {invalid ? (
            <p className="error-message">
              Cart contains invalid stock quantity. Return to cart to fix it.
            </p>
          ) : null}
        </aside>
      </div>
    </>
  );
}

export default function CheckoutPage() {
  return (
    <main className="workspace-page shell">
      <RequireRole role="buyer">
        <CheckoutContent />
      </RequireRole>
    </main>
  );
}
