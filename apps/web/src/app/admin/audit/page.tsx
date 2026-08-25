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
        description="Newest immutable trust and fulfillment actions. Current API provides no search, export, or pagination."
      />
      {!items.length ? (
        <EmptyState
          title="No audit events"
          message="Privileged marketplace actions appear here."
        />
      ) : (
        <div className="audit-timeline">
          {items.map((item) => (
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
