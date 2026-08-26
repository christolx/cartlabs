"use client";

import Link from "next/link";

export function LoadingState({ label = "Loading" }: { label?: string }) {
  return (
    <div
      className="state-panel state-loading"
      aria-live="polite"
      aria-busy="true"
    >
      <span className="state-index" aria-hidden="true">
        00
      </span>
      <span className="skeleton-line" />
      <span className="skeleton-line short" />
      <span className="sr-only">{label}</span>
    </div>
  );
}

export function EmptyState({
  title,
  message,
  href,
  action,
}: {
  title: string;
  message: string;
  href?: string;
  action?: string;
}) {
  return (
    <div className="state-panel">
      <span className="state-index" aria-hidden="true">
        00
      </span>
      <h2>{title}</h2>
      <p>{message}</p>
      {href && action ? (
        <Link className="button button-secondary" href={href}>
          {action}
        </Link>
      ) : null}
    </div>
  );
}

export function ErrorState({
  title = "Something went wrong",
  message,
  retry,
}: {
  title?: string;
  message: string;
  retry?: () => void;
}) {
  return (
    <div className="state-panel" role="alert">
      <span className="state-index" aria-hidden="true">
        !
      </span>
      <h2>{title}</h2>
      <p>{message}</p>
      {retry ? (
        <button
          className="button button-secondary"
          type="button"
          onClick={retry}
        >
          Try again
        </button>
      ) : null}
    </div>
  );
}
