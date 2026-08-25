"use client";

import { useCallback, useEffect, useState, type FormEvent } from "react";
import type { components } from "@/lib/api/schema";
import { RequireRole } from "@/components/require-role";
import { useSession } from "@/components/session-provider";
import { EmptyState, ErrorState, LoadingState } from "@/components/async-state";
import { formatDate, PageHeading, Status } from "@/components/marketplace-ui";
import { errorMessage } from "@/lib/api/browser";

type Store = components["schemas"]["Store"];
type ModerationStatus = "approved" | "rejected";

function StoreModerationContent() {
  const { request } = useSession();
  const [items, setItems] = useState<Store[] | null>(null);
  const [error, setError] = useState("");
  const [message, setMessage] = useState("");
  const [busy, setBusy] = useState("");
  const load = useCallback(async () => {
    setError("");
    try {
      setItems((await request<{ items: Store[] }>("/admin/stores")).items);
    } catch (cause) {
      setError(errorMessage(cause, "Store verification unavailable."));
    }
  }, [request]);
  useEffect(() => {
    const timer = window.setTimeout(() => void load(), 0);
    return () => window.clearTimeout(timer);
  }, [load]);
  async function moderate(
    event: FormEvent<HTMLFormElement>,
    storeId: string,
    status: ModerationStatus,
  ) {
    event.preventDefault();
    const data = new FormData(event.currentTarget);
    setBusy(storeId);
    setMessage("");
    try {
      const next = await request<Store>(`/admin/stores/${storeId}/moderation`, {
        method: "PATCH",
        body: JSON.stringify({ status, note: data.get("note") }),
      });
      setItems(
        (current) =>
          current?.map((item) => (item.id === next.id ? next : item)) ?? null,
      );
      setMessage(`${next.name} is ${next.status}.`);
    } catch (cause) {
      setMessage(
        errorMessage(cause, "Verification failed. Refreshing server state."),
      );
      await load();
    } finally {
      setBusy("");
    }
  }
  if (error)
    return (
      <ErrorState
        title="Store verification unavailable"
        message={error}
        retry={() => void load()}
      />
    );
  if (!items) return <LoadingState label="Loading stores" />;
  const sorted = [...items].sort(
    (left, right) =>
      Number(right.status === "pending") - Number(left.status === "pending"),
  );
  return (
    <>
      <PageHeading
        title="Store verification"
        description="Pending stores appear first. Previous decisions remain visible."
      />
      {message ? (
        <p className="action-message" role="status">
          {message}
        </p>
      ) : null}
      {!sorted.length ? (
        <EmptyState
          title="No stores"
          message="Seller stores appear here after creation."
        />
      ) : (
        <div className="moderation-list">
          {sorted.map((store) => (
            <article key={store.id}>
              <header>
                <div>
                  <h2>{store.name}</h2>
                  <span>
                    /{store.slug} / created {formatDate(store.createdAt)}
                  </span>
                </div>
                <Status value={store.status} />
              </header>
              <p>{store.description}</p>
              {store.moderationNote ? (
                <p className="moderation-note">
                  Verification note: {store.moderationNote}
                </p>
              ) : null}
              <div className="moderation-actions">
                {(["approved", "rejected"] as ModerationStatus[]).map(
                  (status) => (
                    <details key={status}>
                      <summary>
                        {status === "approved" ? "Verify" : "Reject"}
                      </summary>
                      <form
                        className="stack-form compact-form"
                        onSubmit={(event) =>
                          void moderate(event, store.id, status)
                        }
                      >
                        <p>
                          This updates public store availability immediately
                          after server confirmation.
                        </p>
                        <label>
                          <span>Verification note</span>
                          <textarea
                            name="note"
                            minLength={2}
                            maxLength={500}
                            rows={3}
                            required={status === "rejected"}
                          />
                        </label>
                        <button
                          className={
                            status === "approved"
                              ? "button button-primary"
                              : "button button-danger"
                          }
                          disabled={busy === store.id}
                        >
                          Confirm{" "}
                          {status === "approved" ? "verification" : "rejection"}
                        </button>
                      </form>
                    </details>
                  ),
                )}
              </div>
            </article>
          ))}
        </div>
      )}
    </>
  );
}

export default function AdminStoresPage() {
  return (
    <main className="workspace-page shell">
      <RequireRole role="admin">
        <StoreModerationContent />
      </RequireRole>
    </main>
  );
}
