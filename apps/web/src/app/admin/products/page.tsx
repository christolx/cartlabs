"use client";

import Image from "next/image";
import { useCallback, useEffect, useState, type FormEvent } from "react";
import type { components } from "@/lib/api/schema";
import { RequireRole } from "@/components/require-role";
import { useSession } from "@/components/session-provider";
import { EmptyState, ErrorState, LoadingState } from "@/components/async-state";
import { Money, PageHeading, Status } from "@/components/marketplace-ui";
import { errorMessage } from "@/lib/api/browser";

type Product = components["schemas"]["ProductDetail"];
type ModerationStatus = "approved" | "rejected";

function ProductModerationContent() {
  const { request } = useSession();
  const [items, setItems] = useState<Product[] | null>(null);
  const [error, setError] = useState("");
  const [message, setMessage] = useState("");
  const [busy, setBusy] = useState("");
  const load = useCallback(async () => {
    setError("");
    try { setItems((await request<{ items: Product[] }>("/admin/products")).items); }
    catch (cause) { setError(errorMessage(cause, "Product moderation unavailable.")); }
  }, [request]);
  useEffect(() => { const timer = window.setTimeout(() => void load(), 0); return () => window.clearTimeout(timer); }, [load]);
  async function moderate(event: FormEvent<HTMLFormElement>, productId: string, status: ModerationStatus) {
    event.preventDefault(); const data = new FormData(event.currentTarget);
    setBusy(productId); setMessage("");
    try { const next = await request<Product>(`/admin/products/${productId}/moderation`, { method: "PATCH", body: JSON.stringify({ status, note: data.get("note") }) }); setItems((current) => current?.map((item) => item.id === next.id ? next : item) ?? null); setMessage(`${next.name} is ${next.moderationStatus}.`); }
    catch (cause) { setMessage(errorMessage(cause, "Moderation failed. Refreshing server state.")); await load(); }
    finally { setBusy(""); }
  }
  if (error) return <ErrorState title="Product moderation unavailable" message={error} retry={() => void load()} />;
  if (!items) return <LoadingState label="Loading products" />;
  const sorted = [...items].sort((left, right) => Number(right.moderationStatus === "pending") - Number(left.moderationStatus === "pending"));
  return <><PageHeading title="Product moderation" description="Inspect product identity, variants, inventory, and images before decision." />{message ? <p className="action-message" role="status">{message}</p> : null}{!sorted.length ? <EmptyState title="No products" message="Submitted seller products appear here." /> : <div className="moderation-list product-moderation-list">{sorted.map((product) => <article key={product.id}><header><div><h2>{product.name}</h2><span>{product.storeName || "Seller store"} / {product.category.name} / {product.slug}</span></div><div className="status-pair"><Status value={product.status} /><Status value={product.moderationStatus} /></div></header><p>{product.description}</p>{product.moderationNote ? <p className="moderation-note">Current note: {product.moderationNote}</p> : null}<div className="moderation-supply"><section><h3>Variants</h3>{product.variants.map((variant) => <div key={variant.id}><div><strong>{variant.name}</strong><span>{variant.sku}. {variant.stock} in stock.</span></div><Money value={variant.priceMinor} currency={variant.currency} /></div>)}</section><section><h3>Images</h3><div className="moderation-images">{product.images.map((image) => <figure key={image.id}><Image src={image.url} alt={image.altText} width={180} height={120} /><figcaption>{image.altText || "No alt text"}</figcaption></figure>)}</div></section></div><div className="moderation-actions">{(["approved", "rejected"] as ModerationStatus[]).map((status) => <details key={status}><summary>{status === "approved" ? "Approve" : "Reject"}</summary><form className="stack-form compact-form" onSubmit={(event) => void moderate(event, product.id, status)}><p>Catalog state changes only after server confirms this action.</p><label><span>Moderation note</span><textarea name="note" minLength={2} maxLength={500} rows={3} required /></label><button className={status === "approved" ? "button button-primary" : "button button-danger"} disabled={busy === product.id}>Confirm {status === "approved" ? "approval" : "rejection"}</button></form></details>)}</div></article>)}</div>}</>;
}

export default function AdminProductsPage() {
  return <main className="workspace-page shell"><RequireRole role="admin"><ProductModerationContent /></RequireRole></main>;
}
