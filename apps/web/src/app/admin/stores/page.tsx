"use client";

import {
  useCallback,
  useEffect,
  useMemo,
  useState,
  type FormEvent,
} from "react";
import type { components } from "@/lib/api/schema";
import { RequireRole } from "@/components/require-role";
import { useSession } from "@/components/session-provider";
import { EmptyState, ErrorState, LoadingState } from "@/components/async-state";
import { formatDate, PageHeading, Status } from "@/components/marketplace-ui";
import { errorMessage } from "@/lib/api/browser";

type Store = components["schemas"]["Store"];
type ModerationStatus = "approved" | "rejected";
type StoreStatusFilter = "all" | Store["status"];

function StoreModerationContent() {
  const { request } = useSession();
  const [items, setItems] = useState<Store[] | null>(null);
  const [query, setQuery] = useState("");
  const [statusFilter, setStatusFilter] = useState<StoreStatusFilter>("all");
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
  const sorted = useMemo(
    () =>
      [...(items ?? [])].sort(
        (left, right) =>
          Number(right.status === "pending") -
          Number(left.status === "pending"),
      ),
    [items],
  );
  const filtered = useMemo(() => {
    const normalized = query.trim().toLocaleLowerCase();
    return sorted.filter(
      (store) =>
        (!normalized ||
          `${store.name} ${store.slug} ${store.sellerId}`
            .toLocaleLowerCase()
            .includes(normalized)) &&
        (statusFilter === "all" || store.status === statusFilter),
    );
  }, [query, sorted, statusFilter]);
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
  return (
    <>
      <PageHeading
        title="Store verification"
        description="Pending stores appear first. Previous decisions remain visible."
        family="ADMIN / STORES"
        index="03"
        meta={
          <span>
            {filtered.filter((store) => store.status === "pending").length}{" "}
            pending / {filtered.length} of {sorted.length} loaded stores
          </span>
        }
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
        <>
          <div className="workspace-filter-band store-filter-band">
            <label>
              <span>Search stores</span>
              <input
                type="search"
                value={query}
                onChange={(event) => setQuery(event.target.value)}
                placeholder="Name, slug, or seller ID"
              />
            </label>
            <label>
              <span>Status</span>
              <select
                value={statusFilter}
                onChange={(event) =>
                  setStatusFilter(event.target.value as StoreStatusFilter)
                }
              >
                <option value="all">All statuses</option>
                <option value="pending">Pending</option>
                <option value="approved">Approved</option>
                <option value="rejected">Rejected</option>
              </select>
            </label>
            {query || statusFilter !== "all" ? (
              <button
                className="button button-secondary filter-clear"
                type="button"
                onClick={() => {
                  setQuery("");
                  setStatusFilter("all");
                }}
              >
                Clear filters
              </button>
            ) : null}
          </div>
          {!filtered.length ? (
            <EmptyState
              title="No stores match"
              message="Change search or status filters."
            />
          ) : (
            <div className="moderation-list">
              {filtered.map((store) => (
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
                    {store.status === "approved" ? (
                      <span className="helper-text">
                        Approved store. No moderation action required.
                      </span>
                    ) : (
                      (
                        [
                          "approved",
                          ...(store.status === "pending" ? ["rejected"] : []),
                        ] as ModerationStatus[]
                      ).map((status) => (
                        <details key={status}>
                          <summary>
                            {status === "approved"
                              ? store.status === "rejected"
                                ? "Re-verify"
                                : "Verify"
                              : "Reject"}
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
                              {status === "approved"
                                ? "verification"
                                : "rejection"}
                            </button>
                          </form>
                        </details>
                      ))
                    )}
                  </div>
                </article>
              ))}
            </div>
          )}
        </>
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
