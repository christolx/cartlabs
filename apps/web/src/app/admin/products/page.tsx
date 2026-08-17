"use client";

import Image from "next/image";
import { useCallback, useEffect, useState, type FormEvent } from "react";
import type { components } from "@/lib/api/schema";
import { RequireRole } from "@/components/require-role";
import { useSession } from "@/components/session-provider";
import { EmptyState, ErrorState, LoadingState } from "@/components/async-state";
import { Money, PageHeading, Status, formatDate } from "@/components/marketplace-ui";
import { errorMessage } from "@/lib/api/browser";

type Product = components["schemas"]["AdminProduct"];
type EnforcementStatus = "suspended" | "published";

function ProductEnforcementContent() {
  const { request } = useSession();
  const [items, setItems] = useState<Product[] | null>(null);
  const [error, setError] = useState("");
  const [message, setMessage] = useState("");
  const [busy, setBusy] = useState("");

  const load = useCallback(async () => {
    setError("");
    try {
      setItems((await request<{ items: Product[] }>("/admin/products")).items);
    } catch (cause) {
      setError(errorMessage(cause, "Listing enforcement unavailable."));
    }
  }, [request]);

  useEffect(() => {
    const timer = window.setTimeout(() => void load(), 0);
    return () => window.clearTimeout(timer);
  }, [load]);

  async function updateStatus(event: FormEvent<HTMLFormElement>, productId: string, status: EnforcementStatus) {
    event.preventDefault();
    const form = event.currentTarget;
    const data = new FormData(form);
    setBusy(productId);
    setMessage("");
    try {
      const next = await request<Product>(`/admin/products/${productId}/status`, {
        method: "PATCH",
        body: JSON.stringify({ status, reason: data.get("reason") }),
      });
      setItems((current) => current?.map((item) => item.id === next.id ? next : item) ?? null);
      setMessage(`${next.name} is ${next.status}.`);
      form.reset();
      form.closest("details")?.removeAttribute("open");
    } catch (cause) {
      setMessage(errorMessage(cause, "Enforcement action failed. Refreshing server state."));
      await load();
    } finally {
      setBusy("");
    }
  }

  if (error) return <ErrorState title="Listing enforcement unavailable" message={error} retry={() => void load()} />;
  if (!items) return <LoadingState label="Loading listings" />;

  return <>
    <PageHeading title="Listing enforcement" description="Suspend unsafe listings or reinstate corrected listings. Every action requires a reason." />
    {message ? <p className="action-message" role="status">{message}</p> : null}
    {!items.length ? <EmptyState title="No enforceable listings" message="Published and suspended products appear here." /> : <div className="moderation-list product-moderation-list">
      {items.map((product) => {
        const target: EnforcementStatus = product.status === "published" ? "suspended" : "published";
        return <article key={product.id}>
          <header>
            <div><h2>{product.name}</h2><span>{product.storeName || "Seller store"} / {product.category.name} / {product.slug}</span></div>
            <Status value={product.status} />
          </header>
          <p>{product.description}</p>
          {product.enforcementReason ? <p className="moderation-note">Latest suspension: {product.enforcementReason}{product.enforcedBy ? ` / ${product.enforcedBy}` : ""}{product.enforcedAt ? ` / ${formatDate(product.enforcedAt)}` : ""}</p> : null}
          <div className="moderation-supply">
            <section><h3>Variants</h3>{product.variants.map((variant) => <div key={variant.id}><div><strong>{variant.name}</strong><span>{variant.sku}. {variant.stock} in stock.</span></div><Money value={variant.priceMinor} currency={variant.currency} /></div>)}</section>
            <section><h3>Images</h3><div className="moderation-images">{product.images.map((image) => <figure key={image.id}><Image src={image.url} alt={image.altText} width={180} height={120} /><figcaption>{image.altText || "No alt text"}</figcaption></figure>)}</div></section>
          </div>
          <div className="moderation-actions"><details><summary>{target === "suspended" ? "Suspend" : "Reinstate"}</summary><form className="stack-form compact-form" onSubmit={(event) => void updateStatus(event, product.id, target)}><p>{target === "suspended" ? "Listing becomes unavailable to catalog, cart additions, and checkout." : "Store approval and product completeness are revalidated."}</p><label><span>Reason</span><textarea name="reason" minLength={1} maxLength={500} rows={3} required /></label><button className={target === "suspended" ? "button button-danger" : "button button-primary"} disabled={busy === product.id}>Confirm {target === "suspended" ? "suspension" : "reinstatement"}</button></form></details></div>
        </article>;
      })}
    </div>}
  </>;
}

export default function AdminProductsPage() {
  return <main className="workspace-page shell"><RequireRole role="admin"><ProductEnforcementContent /></RequireRole></main>;
}
