"use client";

import { useCallback, useEffect, useState } from "react";
import type { components } from "@/lib/api/schema";
import { RequireRole } from "@/components/require-role";
import { useSession } from "@/components/session-provider";
import { EmptyState, ErrorState, LoadingState } from "@/components/async-state";
import { formatDate, PageHeading } from "@/components/marketplace-ui";
import { errorMessage } from "@/lib/api/browser";

type AuditEvent = components["schemas"]["AuditEvent"];

function AuditContent() {
  const { request } = useSession();
  const [items, setItems] = useState<AuditEvent[] | null>(null);
  const [actorRole, setActorRole] = useState("");
  const [resourceType, setResourceType] = useState("");
  const [action, setAction] = useState("");
  const [error, setError] = useState("");
  const load = useCallback(async () => {
    setError("");
    try {
      setItems(
        (await request<{ items: AuditEvent[] }>("/admin/audit-events")).items,
      );
    } catch (cause) {
      setError(errorMessage(cause, "Audit history unavailable."));
    }
  }, [request]);
  useEffect(() => {
    const timer = window.setTimeout(() => void load(), 0);
    return () => window.clearTimeout(timer);
  }, [load]);
  const filtered = (items ?? []).filter(
    (item) =>
      (!actorRole || item.actorRole === actorRole) &&
      (!resourceType || item.resourceType === resourceType) &&
      (!action ||
        item.action.toLocaleLowerCase().includes(action.toLocaleLowerCase())),
  );
  if (error)
    return (
      <ErrorState
        title="Audit history unavailable"
        message={error}
        retry={() => void load()}
      />
    );
  if (!items) return <LoadingState label="Loading audit history" />;
  return (
    <>
      <PageHeading
        title="Audit history"
        description="Newest immutable trust and fulfillment actions. Filters apply to currently loaded events only."
        family="ADMIN / AUDIT"
        index="05"
        meta={
          <span>
            {filtered.length} of {(items ?? []).length} loaded events
          </span>
        }
      />
      <div className="audit-filter-band">
        <label>
          <span>Actor role</span>
          <select
            value={actorRole}
            onChange={(event) => setActorRole(event.target.value)}
          >
            <option value="">All roles</option>
            <option value="admin">Admin</option>
            <option value="seller">Seller</option>
            <option value="buyer">Buyer</option>
            <option value="system">System</option>
          </select>
        </label>
        <label>
          <span>Resource type</span>
          <input
            value={resourceType}
            onChange={(event) => setResourceType(event.target.value)}
            placeholder="Filter loaded events"
          />
        </label>
        <label>
          <span>Action</span>
          <input
            type="search"
            value={action}
            onChange={(event) => setAction(event.target.value)}
            placeholder="Filter loaded events"
          />
        </label>
        <button
          className="button button-secondary filter-clear"
          type="button"
          disabled={!actorRole && !resourceType && !action}
          onClick={() => {
            setActorRole("");
            setResourceType("");
            setAction("");
          }}
        >
          Clear filters
        </button>
      </div>
      {!filtered.length ? (
        <EmptyState
          title={items.length ? "No matching events" : "No audit events"}
          message={
            items.length
              ? "Change loaded-event filters."
              : "Privileged marketplace actions appear here."
          }
        />
      ) : (
        <div className="audit-timeline">
          {filtered.map((item) => (
            <article key={item.id}>
              <div className="audit-marker">
                <span>{item.resourceType.replaceAll("_", " ")}</span>
                <time dateTime={item.createdAt}>
                  {formatDate(item.createdAt)}
                </time>
              </div>
              <div>
                <h2>{item.action.replaceAll("_", " ")}</h2>
                <p>
                  {item.actorName || "System"} ({item.actorRole}) acted on{" "}
                  {item.resourceType} {item.resourceId}.
                </p>
                {Object.keys(item.data).length ? (
                  <details>
                    <summary>Event metadata</summary>
                    <pre>{JSON.stringify(item.data, null, 2)}</pre>
                  </details>
                ) : (
                  <span className="helper-text">No event metadata.</span>
                )}
              </div>
            </article>
          ))}
        </div>
      )}
    </>
  );
}

export default function AdminAuditPage() {
  return (
    <main className="workspace-page shell">
      <RequireRole role="admin">
        <AuditContent />
      </RequireRole>
    </main>
  );
}
