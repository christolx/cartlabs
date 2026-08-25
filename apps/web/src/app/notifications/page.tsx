"use client";

import { useCallback, useEffect, useState } from "react";
import type { components } from "@/lib/api/schema";
import { RequireAuth } from "@/components/require-role";
import { useSession } from "@/components/session-provider";
import { EmptyState, ErrorState, LoadingState } from "@/components/async-state";
import { formatDate, PageHeading } from "@/components/marketplace-ui";
import { errorMessage } from "@/lib/api/browser";

type Notification = components["schemas"]["Notification"];

function NotificationsContent() {
  const { request } = useSession();
  const [items, setItems] = useState<Notification[] | null>(null);
  const [error, setError] = useState("");
  const load = useCallback(async () => {
    setError("");
    try {
      setItems(
        (await request<{ items: Notification[] }>("/notifications")).items,
      );
    } catch (cause) {
      setError(errorMessage(cause, "Notifications unavailable."));
    }
  }, [request]);
  useEffect(() => {
    const timer = window.setTimeout(() => void load(), 0);
    return () => window.clearTimeout(timer);
  }, [load]);
  if (error)
    return (
      <ErrorState
        title="Notifications unavailable"
        message={error}
        retry={() => void load()}
      />
    );
  if (!items) return <LoadingState label="Loading notifications" />;
  return (
    <>
      <PageHeading
        title="Notifications"
        description="Durable events for current account, newest first."
      />
      {!items.length ? (
        <EmptyState
          title="No notifications"
          message="Marketplace events for your account appear here."
        />
      ) : (
        <div className="notification-feed">
          {items.map((item) => (
            <article key={item.id}>
              <div>
                <span>{item.kind.replaceAll("_", " ")}</span>
                <time dateTime={item.createdAt}>
                  {formatDate(item.createdAt)}
                </time>
              </div>
              <h2>{item.title}</h2>
              <p>{item.body}</p>
            </article>
          ))}
        </div>
      )}
    </>
  );
}

export default function NotificationsPage() {
  return (
    <main className="workspace-page shell">
      <RequireAuth>
        <NotificationsContent />
      </RequireAuth>
    </main>
  );
}
