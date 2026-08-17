"use client";

import { useCallback, useEffect, useState, type FormEvent } from "react";
import type { components } from "@/lib/api/schema";
import { RequireRole } from "@/components/require-role";
import { useSession } from "@/components/session-provider";
import { ErrorState, LoadingState } from "@/components/async-state";
import { PageHeading, Status } from "@/components/marketplace-ui";
import { BrowserAPIError, errorMessage } from "@/lib/api/browser";

type Store = components["schemas"]["Store"];

function StoreContent() {
  const { request } = useSession();
  const [store, setStore] = useState<Store | null | undefined>(undefined);
  const [error, setError] = useState("");
  const [message, setMessage] = useState("");
  const [busy, setBusy] = useState(false);
  const load = useCallback(async () => {
    setError("");
    try { setStore(await request<Store>("/seller/store")); }
    catch (cause) { if (cause instanceof BrowserAPIError && cause.status === 404) setStore(null); else setError(errorMessage(cause, "Store unavailable.")); }
  }, [request]);
  useEffect(() => { const timer = window.setTimeout(() => void load(), 0); return () => window.clearTimeout(timer); }, [load]);

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const data = new FormData(event.currentTarget);
    setBusy(true); setMessage("");
    try {
      const next = await request<Store>("/seller/store", { method: store ? "PATCH" : "POST", body: JSON.stringify({ name: data.get("name"), slug: data.get("slug"), description: data.get("description") }) });
      setStore(next); setMessage(store ? "Store saved." : "Store created and submitted for verification.");
    } catch (cause) { setMessage(errorMessage(cause, "Store save failed.")); }
    finally { setBusy(false); }
  }
  if (error) return <ErrorState title="Store unavailable" message={error} retry={() => void load()} />;
  if (store === undefined) return <LoadingState label="Loading store" />;
  return <><PageHeading title={store ? "Manage store" : "Create store"} description="Verified store status is required before product supply can be created." actions={store ? <Status value={store.status} /> : undefined} />{store?.status === "rejected" ? <div className="notice notice-danger"><strong>Verification rejected</strong><p>{store.moderationNote || "No verification note provided."}</p></div> : store?.status === "pending" ? <div className="notice"><strong>Verification pending</strong><p>Products cannot be created until admin verifies this store.</p></div> : null}<section className="form-panel"><form className="stack-form" onSubmit={submit}><label><span>Store name</span><input name="name" minLength={2} maxLength={120} defaultValue={store?.name} required disabled={busy} /></label><label><span>Store slug</span><input name="slug" pattern="[a-z0-9]+(?:-[a-z0-9]+)*" maxLength={80} defaultValue={store?.slug} required disabled={busy} /></label><label><span>Description</span><textarea name="description" minLength={2} maxLength={2000} rows={6} defaultValue={store?.description} required disabled={busy} /></label>{store ? <p className="helper-text">Approved stores stay public after profile edits. Rejected edits return to pending verification.</p> : null}{message ? <p className="form-message" role="status">{message}</p> : null}<button className="button button-primary" disabled={busy}>{busy ? "Saving" : store ? "Save store" : "Create store"}</button></form></section></>;
}

export default function SellerStorePage() {
  return <main className="workspace-page shell"><RequireRole role="seller"><StoreContent /></RequireRole></main>;
}
