"use client";

import { useCallback, useState, type FormEvent } from "react";
import Link from "next/link";
import type { components } from "@/lib/api/schema";

type Role = components["schemas"]["Role"];
type Session = components["schemas"]["Session"];
type Store = components["schemas"]["Store"];
type Product = components["schemas"]["ProductDetail"];
type Category = components["schemas"]["Category"];

const DEMO_PRODUCT_IMAGE_URL = "/images/shared-product.webp";

type Workspace = { store?: Store; stores?: Store[]; products: Product[]; categories: Category[] };

async function request<T>(path: string, token: string, init?: RequestInit): Promise<T> {
  const headers = new Headers(init?.headers);
  headers.set("Accept", "application/json");
  if (token) headers.set("Authorization", `Bearer ${token}`);
  if (init?.body) headers.set("Content-Type", "application/json");
  const response = await fetch(`/api/backend${path}`, { ...init, headers });
  if (!response.ok) {
    const problem = (await response.json().catch(() => ({}))) as { detail?: string };
    throw new Error(problem.detail ?? `Request failed with status ${response.status}`);
  }
  if (response.status === 204) return undefined as T;
  return (await response.json()) as T;
}

export function DemoConsole() {
  const [session, setSession] = useState<Session | null>(null);
  const [workspace, setWorkspace] = useState<Workspace>({ products: [], categories: [] });
  const [busy, setBusy] = useState(false);
  const [message, setMessage] = useState("");

  const loadWorkspace = useCallback(async (current: Session) => {
    const token = current.accessToken;
    if (current.user.role === "seller") {
      const [storeResponse, products, categories] = await Promise.all([
        request<Store>("/seller/store", token).catch((error: Error) => error.message === "resource not found" ? undefined : Promise.reject(error)),
        request<{ items: Product[] }>("/seller/products", token),
        request<{ items: Category[] }>("/categories", token),
      ]);
      setWorkspace({ store: storeResponse, products: products.items, categories: categories.items });
    } else if (current.user.role === "admin") {
      const [stores, products] = await Promise.all([
        request<{ items: Store[] }>("/admin/stores", token),
        request<{ items: Product[] }>("/admin/products", token),
      ]);
      setWorkspace({ stores: stores.items, products: products.items, categories: [] });
    } else {
      setWorkspace({ products: [], categories: [] });
    }
  }, []);

  async function login(role: Role) {
    setBusy(true); setMessage("");
    try {
      const current = await request<Session>("/auth/demo-login", "", { method: "POST", body: JSON.stringify({ role }) });
      setSession(current);
      await loadWorkspace(current);
      setMessage(`Signed in as ${current.user.displayName}.`);
    } catch (error) { setMessage(error instanceof Error ? error.message : "Login failed"); }
    finally { setBusy(false); }
  }

  async function submitStore(event: FormEvent<HTMLFormElement>) {
    event.preventDefault(); if (!session) return;
    setBusy(true); setMessage("");
    const form = event.currentTarget;
    const values = Object.fromEntries(new FormData(form));
    try {
      await request<Store>("/seller/store", session.accessToken, { method: "POST", body: JSON.stringify(values) });
      await loadWorkspace(session); setMessage("Store submitted for admin review."); form.reset();
    } catch (error) { setMessage(error instanceof Error ? error.message : "Store creation failed"); }
    finally { setBusy(false); }
  }

  async function submitProduct(event: FormEvent<HTMLFormElement>) {
    event.preventDefault(); if (!session) return;
    setBusy(true); setMessage("");
    const form = event.currentTarget;
    const data = new FormData(form);
    try {
      const product = await request<Product>("/seller/products", session.accessToken, {
        method: "POST",
        body: JSON.stringify({ categoryId: data.get("categoryId"), name: data.get("name"), slug: data.get("slug"), description: data.get("description") }),
      });
      await request(`/seller/products/${product.id}/variants`, session.accessToken, {
        method: "POST",
        body: JSON.stringify({ sku: data.get("sku"), name: data.get("variantName"), attributes: {}, priceMinor: Number(data.get("priceMinor")), currency: "IDR", stock: Number(data.get("stock")) }),
      });
      await request(`/seller/products/${product.id}/images`, session.accessToken, {
        method: "POST",
        body: JSON.stringify({ url: DEMO_PRODUCT_IMAGE_URL, altText: data.get("name"), position: 0 }),
      });
      await request(`/seller/products/${product.id}/publish`, session.accessToken, { method: "POST" });
      await loadWorkspace(session); setMessage("Product published and waiting for admin review."); form.reset();
    } catch (error) { setMessage(error instanceof Error ? error.message : "Product creation failed"); }
    finally { setBusy(false); }
  }

  async function moderate(kind: "stores" | "products", id: string, status: "approved" | "rejected") {
    if (!session) return; setBusy(true); setMessage("");
    try {
      await request(`/admin/${kind}/${id}/moderation`, session.accessToken, { method: "PATCH", body: JSON.stringify({ status, note: "Reviewed in web demo" }) });
      await loadWorkspace(session); setMessage(`${kind === "stores" ? "Store" : "Product"} ${status}.`);
    } catch (error) { setMessage(error instanceof Error ? error.message : "Moderation failed"); }
    finally { setBusy(false); }
  }

  return (
    <div className="demo-console">
      <div className="role-switcher" role="group" aria-label="Demo role selection">
        {(["buyer", "seller", "admin"] as Role[]).map((role) => (
          <button key={role} className={session?.user.role === role ? "role-active" : ""} aria-pressed={session?.user.role === role} disabled={busy} onClick={() => login(role)}>{role}</button>
        ))}
      </div>
      <p className="demo-message" aria-live="polite">{busy ? "Working..." : message || "Choose a role. Demo mode must be enabled in API configuration."}</p>

      {session?.user.role === "buyer" && (
        <section className="workspace-panel"><h2>Buyer view</h2><p>Public catalog shows only approved products from approved stores.</p><Link className="button button-primary" href="/#catalog">Browse catalog</Link></section>
      )}

      {session?.user.role === "seller" && (
        <div className="workspace-grid">
          <section className="workspace-panel">
            <h2>Seller store</h2>
            {workspace.store ? <><p className="status-line"><strong>{workspace.store.name}</strong><span data-status={workspace.store.status}>{workspace.store.status}</span></p><p>{workspace.store.description}</p></> : (
              <form className="stack-form" onSubmit={submitStore}>
                <Field label="Store name" name="name" minLength={2} required />
                <Field label="Store slug" name="slug" pattern="[a-z0-9]+(?:-[a-z0-9]+)*" required />
                <Field label="Description" name="description" required />
                <button className="button button-primary" disabled={busy}>Submit store</button>
              </form>
            )}
          </section>
          <section className="workspace-panel">
            <h2>Publish product</h2>
            <p>Requires approved store. Product enters admin review after publish.</p>
            <form className="stack-form two-column-form" onSubmit={submitProduct}>
              <Field label="Product name" name="name" required />
              <Field label="Product slug" name="slug" pattern="[a-z0-9]+(?:-[a-z0-9]+)*" required />
              <label><span>Category</span><select name="categoryId" required>{workspace.categories.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}</select></label>
              <Field label="SKU" name="sku" required />
              <Field label="Variant name" name="variantName" required />
              <Field label="Price (IDR)" name="priceMinor" type="number" min={0} required />
              <Field label="Starting stock" name="stock" type="number" min={1} required />
              <label><span>Product image</span><input value="Shared demo image" readOnly /></label>
              <label className="form-wide"><span>Description</span><textarea name="description" rows={3} required /></label>
              <button className="button button-primary form-wide" disabled={busy}>Create and publish</button>
            </form>
          </section>
          <section className="workspace-panel workspace-wide"><h2>Owned products</h2><ProductRows products={workspace.products} /></section>
        </div>
      )}

      {session?.user.role === "admin" && (
        <div className="workspace-grid">
          <section className="workspace-panel"><h2>Store moderation</h2><ModerationRows items={workspace.stores ?? []} kind="stores" busy={busy} moderate={moderate} /></section>
          <section className="workspace-panel"><h2>Product moderation</h2><ModerationRows items={workspace.products} kind="products" busy={busy} moderate={moderate} /></section>
        </div>
      )}
    </div>
  );
}

function Field({ label, ...input }: { label: string } & React.InputHTMLAttributes<HTMLInputElement>) {
  return <label><span>{label}</span><input {...input} /></label>;
}

function ProductRows({ products }: { products: Product[] }) {
  if (!products.length) return <p className="empty-copy">No owned products yet.</p>;
  return <div className="record-list">{products.map((product) => <div className="record-row" key={product.id}><div><strong>{product.name}</strong><span>{product.variants.length} variants</span></div><span data-status={product.moderationStatus}>{product.status}, {product.moderationStatus}</span></div>)}</div>;
}

function ModerationRows({ items, kind, busy, moderate }: { items: (Store | Product)[]; kind: "stores" | "products"; busy: boolean; moderate: (kind: "stores" | "products", id: string, status: "approved" | "rejected") => void }) {
  if (!items.length) return <p className="empty-copy">Nothing to review.</p>;
  return <div className="record-list">{items.map((item) => {
    const status = "moderationStatus" in item ? item.moderationStatus : item.status;
    return <div className="record-row moderation-row" key={item.id}><div><strong>{item.name}</strong><span data-status={status}>{status}</span></div><div className="row-actions"><button disabled={busy || status === "approved"} onClick={() => moderate(kind, item.id, "approved")}>Approve</button><button disabled={busy || status === "rejected"} onClick={() => moderate(kind, item.id, "rejected")}>Reject</button></div></div>;
  })}</div>;
}
