"use client";

import Image from "next/image";
import Link from "next/link";
import { useCallback, useEffect, useState } from "react";
import type { components } from "@/lib/api/schema";
import { RequireRole } from "@/components/require-role";
import { useSession } from "@/components/session-provider";
import { EmptyState, ErrorState, LoadingState } from "@/components/async-state";
import { Money, PageHeading } from "@/components/marketplace-ui";
import { errorMessage } from "@/lib/api/browser";

type Cart = components["schemas"]["Cart"];

function CartContent() {
  const { request } = useSession();
  const [cart, setCart] = useState<Cart | null>(null);
  const [error, setError] = useState("");
  const [busyId, setBusyId] = useState("");
  const [message, setMessage] = useState("");
  const load = useCallback(async () => {
    setError("");
    try { setCart(await request<Cart>("/cart")); }
    catch (cause) { setError(errorMessage(cause, "Cart unavailable.")); }
  }, [request]);
  useEffect(() => { const timer = window.setTimeout(() => void load(), 0); return () => window.clearTimeout(timer); }, [load]);

  async function update(variantId: string, quantity: number) {
    setBusyId(variantId);
    setMessage("");
    try {
      const next = await request<Cart>(`/cart/items/${variantId}`, quantity > 0 ? { method: "PUT", body: JSON.stringify({ quantity }) } : { method: "DELETE" });
      setCart(next);
      setMessage(quantity > 0 ? "Quantity updated." : "Item removed.");
    } catch (cause) { setMessage(errorMessage(cause, "Cart update failed.")); }
    finally { setBusyId(""); }
  }

  if (error) return <ErrorState title="Cart unavailable" message={error} retry={() => void load()} />;
  if (!cart) return <LoadingState label="Loading cart" />;
  return <><PageHeading title="Your cart" description="Live price and inventory are checked again at checkout." actions={cart.totalQuantity ? <Link className="button button-primary" href="/checkout">Continue to checkout</Link> : undefined} />{message ? <p className="action-message" role="status">{message}</p> : null}{!cart.totalQuantity ? <EmptyState title="Cart is empty" message="Browse approved products and choose an in-stock variant." href="/" action="Browse catalog" /> : <div className="cart-layout"><div className="cart-store-list">{cart.stores.map((store) => <section className="cart-store-section" key={store.storeId}><div className="cart-store-heading"><h2>{store.storeName}</h2><strong><Money value={store.subtotalMinor} /></strong></div>{store.items.map((item) => <article className="cart-line" key={item.variantId}>{item.imageUrl ? <Image src={item.imageUrl} alt="" width={96} height={72} /> : <div className="line-image-fallback">No image</div>}<div className="cart-line-copy"><h3><Link href={`/products/${item.productSlug}`}>{item.productName}</Link></h3><p>{item.variantName} / {item.sku}</p><span><Money value={item.unitPriceMinor} /> each. {item.availableStock} available.</span></div><div className="cart-line-actions"><label><span>Quantity</span><input type="number" min="1" max={Math.max(1, item.availableStock)} defaultValue={item.quantity} disabled={busyId === item.variantId} onBlur={(event) => { const value = Number(event.currentTarget.value); if (Number.isInteger(value) && value !== item.quantity) void update(item.variantId, value); }} /></label><strong><Money value={item.lineTotalMinor} /></strong><button className="text-button danger-button" type="button" disabled={busyId === item.variantId} onClick={() => void update(item.variantId, 0)}>Remove</button></div></article>)}</section>)}</div><aside className="order-summary"><h2>Cart total</h2><dl><div><dt>Items</dt><dd>{cart.totalQuantity}</dd></div><div><dt>Total</dt><dd><Money value={cart.subtotalMinor} /></dd></div></dl><Link className="button button-primary" href="/checkout">Continue to checkout</Link></aside></div>}</>;
}

export default function CartPage() {
  return <main className="workspace-page shell"><RequireRole role="buyer"><CartContent /></RequireRole></main>;
}
