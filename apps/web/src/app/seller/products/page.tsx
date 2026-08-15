"use client";

import Image from "next/image";
import Link from "next/link";
import { useCallback, useEffect, useState } from "react";
import type { components } from "@/lib/api/schema";
import { RequireRole } from "@/components/require-role";
import { useSession } from "@/components/session-provider";
import { EmptyState, ErrorState, LoadingState } from "@/components/async-state";
import { Money, PageHeading, Status } from "@/components/marketplace-ui";
import { errorMessage } from "@/lib/api/browser";

type Product = components["schemas"]["ProductDetail"];

function ProductsContent() {
  const { request } = useSession();
  const [items, setItems] = useState<Product[] | null>(null);
  const [error, setError] = useState("");
  const load = useCallback(async () => {
    setError("");
    try { setItems((await request<{ items: Product[] }>("/seller/products")).items); }
    catch (cause) { setError(errorMessage(cause, "Products unavailable.")); }
  }, [request]);
  useEffect(() => { const timer = window.setTimeout(() => void load(), 0); return () => window.clearTimeout(timer); }, [load]);
  if (error) return <ErrorState title="Products unavailable" message={error} retry={() => void load()} />;
  if (!items) return <LoadingState label="Loading seller products" />;
  return <><PageHeading title="Products" description="Owned product lifecycle, moderation, variants, and current stock." actions={<Link className="button button-primary" href="/seller/products/new">Create draft</Link>} />{!items.length ? <EmptyState title="No products" message="Approved store owners can create first product draft." href="/seller/products/new" action="Create draft" /> : <div className="seller-product-list">{items.map((product) => { const stock = product.variants.reduce((total, variant) => total + variant.stock, 0); const lowest = product.variants.reduce<number | null>((value, variant) => value === null || variant.priceMinor < value ? variant.priceMinor : value, null); return <article key={product.id}>{product.images[0] ? <Image src={product.images[0].url} alt={product.images[0].altText} width={120} height={90} /> : <div className="line-image-fallback">No image</div>}<div className="seller-product-copy"><div className="heading-status"><h2>{product.name}</h2><div><Status value={product.status} /> <Status value={product.moderationStatus} /></div></div><p>{product.category.name}. {product.variants.length} variants. {stock} units in stock{lowest !== null ? <>. From <Money value={lowest} /></> : null}.</p>{product.moderationStatus === "rejected" ? <p className="moderation-note">Admin note: {product.moderationNote || "No note provided."}</p> : null}<Link className="text-link" href={`/seller/products/${product.id}`}>Manage product</Link></div></article>; })}</div>}</>;
}

export default function SellerProductsPage() {
  return <main className="workspace-page shell"><RequireRole role="seller"><ProductsContent /></RequireRole></main>;
}
