"use client";

import Link from "next/link";
import { useCallback, useEffect, useState } from "react";
import type { components } from "@/lib/api/schema";
import { RequireRoles } from "@/components/require-role";
import { useSession } from "@/components/session-provider";
import { EmptyState, ErrorState, LoadingState } from "@/components/async-state";
import { formatDate, PageHeading } from "@/components/marketplace-ui";
import { errorMessage } from "@/lib/api/browser";

type Notification = components["schemas"]["Notification"];

function NotificationsContent() {
  const { request, user } = useSession();
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
        family="ACCOUNT / EVENTS"
        index="01"
        meta={
          <span>
            {items.length} event{items.length === 1 ? "" : "s"}
          </span>
        }
      />
      {!items.length ? (
        <EmptyState
          title="No notifications"
          message="Marketplace events for your account appear here."
        />
      ) : (
        <div className="notification-feed">
          <h2 className="notification-day">
            Newest first / {items.length} event{items.length === 1 ? "" : "s"}
          </h2>
          {items.map((item) => (
            <article className={item.readAt ? "" : "is-unread"} key={item.id}>
              <div className="notification-meta">
                <span>{item.kind.replaceAll("_", " ")}</span>
                <time dateTime={item.createdAt}>
                  {formatDate(item.createdAt)}
                </time>
              </div>
              <div className="notification-content">
                <h2>{item.title}</h2>
                <p>{item.body}</p>
                {item.href ? (
                  <Link className="notification-link" href={item.href}>
                    {user?.role === "seller"
                      ? "View seller orders"
                      : "View purchase"}
                  </Link>
                ) : null}
              </div>
              <span className="notification-read-state">
                {item.readAt ? "Read" : "New"}
              </span>
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
      <RequireRoles roles={["buyer", "seller"]}>
        <NotificationsContent />
      </RequireRoles>
    </main>
  );
}
