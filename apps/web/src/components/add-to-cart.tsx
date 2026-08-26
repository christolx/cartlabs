"use client";

import { useMemo, useState } from "react";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import type { components } from "@/lib/api/schema";
import { errorMessage } from "@/lib/api/browser";
import { useSession } from "@/components/session-provider";
import { formatMoney } from "@/lib/api/client";

type Product = components["schemas"]["ProductDetail"];
type Cart = components["schemas"]["Cart"];

export function AddToCart({ product }: { product: Product }) {
  const available = useMemo(
    () =>
      product.variants.filter((variant) => variant.active && variant.stock > 0),
    [product.variants],
  );
  const [variantId, setVariantId] = useState(available[0]?.id ?? "");
  const selected = product.variants.find((variant) => variant.id === variantId);
  const [quantity, setQuantity] = useState(1);
  const [busy, setBusy] = useState(false);
  const [message, setMessage] = useState("");
  const [added, setAdded] = useState(false);
  const { status, user, request } = useSession();
  const router = useRouter();
  const pathname = usePathname();

  async function add() {
    setMessage("");
    setAdded(false);
    if (status !== "authenticated") {
      router.push(`/login?next=${encodeURIComponent(pathname)}`);
      return;
    }
    if (user?.role !== "buyer") {
      setMessage(
        "Buyer account required. Seller and admin accounts cannot purchase.",
      );
      return;
    }
    if (!selected || quantity < 1 || quantity > selected.stock) {
      setMessage("Select an in-stock variant and valid quantity.");
      return;
    }
    setBusy(true);
    try {
      const cart = await request<Cart>(`/cart/items/${selected.id}`, {
        method: "PUT",
        body: JSON.stringify({ quantity }),
      });
      window.dispatchEvent(
        new CustomEvent("cart:updated", {
          detail: { quantity: cart.totalQuantity },
        }),
      );
      setAdded(true);
      setMessage(`${quantity} item${quantity === 1 ? "" : "s"} added to cart.`);
    } catch (error) {
      setMessage(errorMessage(error, "Cart update failed."));
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="purchase-box">
      <label>
        <span>Variant</span>
        <select
          value={variantId}
          onChange={(event) => {
            setVariantId(event.target.value);
            setQuantity(1);
          }}
          disabled={!available.length || busy}
        >
          {available.length ? (
            available.map((variant) => (
              <option key={variant.id} value={variant.id}>
                {variant.name} -{" "}
                {formatMoney(variant.priceMinor, variant.currency)} (
                {variant.stock} available)
              </option>
            ))
          ) : (
            <option>Out of stock</option>
          )}
        </select>
      </label>
      <label>
        <span>Quantity</span>
        <input
          type="number"
          min="1"
          max={selected?.stock ?? 1}
          value={quantity}
          onChange={(event) => setQuantity(Number(event.target.value))}
          disabled={!selected || busy}
        />
      </label>
      <button
        className="button button-primary"
        type="button"
        disabled={!selected || busy}
        onClick={() => void add()}
      >
        {busy
          ? "Updating cart"
          : status === "authenticated" && user?.role !== "buyer"
            ? "Buyer access required"
            : "Add to cart"}
      </button>
      {message ? (
        <p
          className={
            added
              ? "form-message success-message"
              : "form-message error-message"
          }
          role="status"
        >
          {message}{" "}
          {added ? (
            <Link className="text-link" href="/cart">
              View cart
            </Link>
          ) : null}
        </p>
      ) : null}
    </div>
  );
}
