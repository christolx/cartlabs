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

type AdminUser = components["schemas"]["AdminUser"];
type UserStatus = components["schemas"]["UserStatus"];

function UserManagementContent() {
  const { user, request } = useSession();
  const [items, setItems] = useState<AdminUser[] | null>(null);
  const [query, setQuery] = useState("");
  const [role, setRole] = useState("");
  const [status, setStatus] = useState("");
  const [error, setError] = useState("");
  const [message, setMessage] = useState("");
  const [busy, setBusy] = useState("");
  const load = useCallback(async () => {
    setError("");
    try {
      setItems((await request<{ items: AdminUser[] }>("/admin/users")).items);
    } catch (cause) {
      setError(errorMessage(cause, "User management unavailable."));
    }
  }, [request]);
  useEffect(() => {
    const timer = window.setTimeout(() => void load(), 0);
    return () => window.clearTimeout(timer);
  }, [load]);
  const filtered = useMemo(
    () =>
      (items ?? []).filter((item) => {
        const normalized = query.trim().toLocaleLowerCase();
        return (
          (!normalized ||
            `${item.displayName} ${item.email}`
              .toLocaleLowerCase()
              .includes(normalized)) &&
          (!role || item.role === role) &&
          (!status || item.status === status)
        );
      }),
    [items, query, role, status],
  );

  async function updateStatus(
    event: FormEvent<HTMLFormElement>,
    userId: string,
    nextStatus: UserStatus,
  ) {
    event.preventDefault();
    const data = new FormData(event.currentTarget);
    setBusy(userId);
    setMessage("");
    try {
      const next = await request<AdminUser>(`/admin/users/${userId}/status`, {
        method: "PATCH",
        body: JSON.stringify({
          status: nextStatus,
          reason: data.get("reason"),
        }),
      });
      setItems(
        (current) =>
          current?.map((item) => (item.id === next.id ? next : item)) ?? null,
      );
      setMessage(`${next.email} is ${next.status}. Audit event recorded.`);
    } catch (cause) {
      setMessage(
        errorMessage(cause, "Status update failed. Refreshing server state."),
      );
      await load();
    } finally {
      setBusy("");
    }
  }

  if (error)
    return (
      <ErrorState
        title="User management unavailable"
        message={error}
        retry={() => void load()}
      />
    );
  if (!items) return <LoadingState label="Loading users" />;
  const activeAdmins = items.filter(
    (item) => item.role === "admin" && item.status === "active",
  ).length;
  return (
    <>
      <PageHeading
        title="User management"
        description="Suspension revokes current access and refresh sessions. Reactivation requires fresh login."
      />
      <div className="notice notice-danger">
        <strong>Account access control</strong>
        <p>
          Reasons are required and stored in audit history. Self-suspension and
          last-active-admin suspension are blocked.
        </p>
      </div>
      <div className="user-filters">
        <label>
          <span>Search</span>
          <input
            type="search"
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            placeholder="Name or email"
          />
        </label>
        <label>
          <span>Role</span>
          <select
            value={role}
            onChange={(event) => setRole(event.target.value)}
          >
            <option value="">All roles</option>
            <option value="buyer">Buyer</option>
            <option value="seller">Seller</option>
            <option value="admin">Admin</option>
          </select>
        </label>
        <label>
          <span>Status</span>
          <select
            value={status}
            onChange={(event) => setStatus(event.target.value)}
          >
            <option value="">All statuses</option>
            <option value="active">Active</option>
            <option value="suspended">Suspended</option>
          </select>
        </label>
      </div>
      {message ? (
        <p className="action-message" role="status">
          {message}
        </p>
      ) : null}
      {!filtered.length ? (
        <EmptyState
          title="No users match"
          message="Change search or filters."
        />
      ) : (
        <div className="user-table" role="table" aria-label="Marketplace users">
          <div className="user-table-head" role="row">
            <span role="columnheader">User</span>
            <span role="columnheader">Role</span>
            <span role="columnheader">Status</span>
            <span role="columnheader">Created</span>
            <span role="columnheader">Action</span>
          </div>
          {filtered.map((item) => {
            const self = item.id === user?.id;
            const lastAdmin =
              item.role === "admin" &&
              item.status === "active" &&
              activeAdmins <= 1;
            const target: UserStatus =
              item.status === "active" ? "suspended" : "active";
            const blocked =
              (self && target === "suspended") ||
              (lastAdmin && target === "suspended");
            return (
              <div className="user-table-row" role="row" key={item.id}>
                <div role="cell">
                  <strong>{item.displayName}</strong>
                  <span>{item.email}</span>
                </div>
                <span role="cell">{item.role}</span>
                <span role="cell">
                  <Status value={item.status} />
                </span>
                <time role="cell" dateTime={item.createdAt}>
                  {formatDate(item.createdAt)}
                </time>
                <div role="cell">
                  {blocked ? (
                    <span className="helper-text">
                      {self ? "Current account" : "Last active admin"}
                    </span>
                  ) : (
                    <details className="user-action">
                      <summary>
                        {target === "suspended" ? "Suspend" : "Reactivate"}
                      </summary>
                      <form
                        className="stack-form compact-form"
                        onSubmit={(event) =>
                          void updateStatus(event, item.id, target)
                        }
                      >
                        <p>
                          {target === "suspended"
                            ? "This immediately invalidates account sessions."
                            : "User must establish a new login session."}
                        </p>
                        <label>
                          <span>Reason</span>
                          <textarea
                            name="reason"
                            minLength={2}
                            maxLength={500}
                            rows={3}
                            required
                          />
                        </label>
                        <button
                          className={
                            target === "suspended"
                              ? "button button-danger"
                              : "button button-primary"
                          }
                          disabled={busy === item.id}
                        >
                          Confirm {target}
                        </button>
                      </form>
                    </details>
                  )}
                </div>
              </div>
            );
          })}
        </div>
      )}
    </>
  );
}

export default function AdminUsersPage() {
  return (
    <main className="workspace-page shell">
      <RequireRole role="admin">
        <UserManagementContent />
      </RequireRole>
    </main>
  );
}
