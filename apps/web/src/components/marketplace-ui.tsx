import type { ReactNode } from "react";
import { formatMoney } from "@/lib/api/client";

export function PageHeading({
  title,
  description,
  actions,
  family = "WORKSPACE",
  index = "01",
  meta,
  compact = false,
}: {
  title: string;
  description: string;
  actions?: ReactNode;
  family?: string;
  index?: string;
  meta?: ReactNode;
  compact?: boolean;
}) {
  return (
    <div
      className={`page-heading route-masthead${compact ? " route-masthead-compact" : ""}`}
    >
      <span className="route-index" aria-hidden="true">
        {index}
      </span>
      <div className="route-masthead-main">
        <span className="route-family">{family}</span>
        <h1>{title}</h1>
        <p>{description}</p>
        {meta ? <div className="route-meta">{meta}</div> : null}
      </div>
      {actions ? <div className="page-actions">{actions}</div> : null}
    </div>
  );
}

export function Status({ value }: { value: string }) {
  const normalized = value.toLowerCase();
  return (
    <span className="status-label" data-status={normalized}>
      <i aria-hidden="true" />
      {normalized.replaceAll("_", " ")}
    </span>
  );
}

export function Money({
  value,
  currency = "IDR",
}: {
  value: number;
  currency?: string;
}) {
  return <>{formatMoney(value, currency)}</>;
}

export function formatDate(
  value: string,
  precision: "date" | "datetime" = "datetime",
) {
  return new Intl.DateTimeFormat("id-ID", {
    dateStyle: "medium",
    ...(precision === "datetime" ? { timeStyle: "short" } : {}),
  }).format(new Date(value));
}
