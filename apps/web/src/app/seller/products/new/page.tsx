"use client";

import { useCallback, useEffect, useState, type FormEvent } from "react";
import { useRouter } from "next/navigation";
import type { components } from "@/lib/api/schema";
import { RequireRole } from "@/components/require-role";
import { useSession } from "@/components/session-provider";
import { ErrorState, LoadingState } from "@/components/async-state";
import { PageHeading } from "@/components/marketplace-ui";
import { errorMessage } from "@/lib/api/browser";

type Category = components["schemas"]["Category"];
type Product = components["schemas"]["ProductDetail"];

function NewProductContent() {
  const { request } = useSession();
  const router = useRouter();
  const [categories, setCategories] = useState<Category[] | null>(null);
  const [error, setError] = useState("");
  const [message, setMessage] = useState("");
  const [busy, setBusy] = useState(false);
  const load = useCallback(async () => {
    setError("");
    try { setCategories((await request<{ items: Category[] }>("/categories")).items); }
    catch (cause) { setError(errorMessage(cause, "Categories unavailable.")); }
  }, [request]);
  useEffect(() => { const timer = window.setTimeout(() => void load(), 0); return () => window.clearTimeout(timer); }, [load]);
  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const data = new FormData(event.currentTarget);
    setBusy(true); setMessage("");
    try {
      const product = await request<Product>("/seller/products", { method: "POST", body: JSON.stringify({ categoryId: data.get("categoryId"), name: data.get("name"), slug: data.get("slug"), description: data.get("description") }) });
      router.replace(`/seller/products/${product.id}`);
    } catch (cause) { setMessage(errorMessage(cause, "Product creation failed. Approved store required.")); }
    finally { setBusy(false); }
  }
  if (error) return <ErrorState title="Product form unavailable" message={error} retry={() => void load()} />;
  if (!categories) return <LoadingState label="Loading product form" />;
  return <><PageHeading title="Create product draft" description="Add basic identity first. Variants, images, inventory, and moderation follow on product page." /><section className="form-panel"><form className="stack-form" onSubmit={submit}><label><span>Category</span><select name="categoryId" required disabled={busy}><option value="">Choose category</option>{categories.map((category) => <option key={category.id} value={category.id}>{category.name}</option>)}</select></label><label><span>Product name</span><input name="name" minLength={2} maxLength={160} required disabled={busy} /></label><label><span>Product slug</span><input name="slug" pattern="[a-z0-9]+(?:-[a-z0-9]+)*" maxLength={120} required disabled={busy} /></label><label><span>Description</span><textarea name="description" minLength={2} maxLength={4000} rows={7} required disabled={busy} /></label>{message ? <p className="form-message error-message" role="alert">{message}</p> : null}<button className="button button-primary" disabled={busy}>{busy ? "Creating" : "Create draft"}</button></form></section></>;
}

export default function NewSellerProductPage() {
  return <main className="workspace-page shell"><RequireRole role="seller"><NewProductContent /></RequireRole></main>;
}
