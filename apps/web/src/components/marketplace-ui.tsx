import type { ReactNode } from "react";
import { formatMoney } from "@/lib/api/client";

export function PageHeading({
  title,
  description,
  actions,
}: {
  title: string;
  description: string;
  actions?: ReactNode;
}) {
  return (
    <div className="page-heading">
      <div>
        <h1>{title}</h1>
        <p>{description}</p>
      </div>
      {actions ? <div className="page-actions">{actions}</div> : null}
    </div>
  );
}

export function Status({ value }: { value: string }) {
  return (
    <span className="status-label" data-status={value}>
      {value.replaceAll("_", " ")}
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

export function formatDate(value: string) {
  return new Intl.DateTimeFormat("id-ID", {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(new Date(value));
}
