"use client";

import Link from "next/link";
import { useCallback, useEffect, useState } from "react";
import type { components } from "@/lib/api/schema";
import { RequireRole } from "@/components/require-role";
import { useSession } from "@/components/session-provider";
import { ErrorState, LoadingState } from "@/components/async-state";
import { Money, PageHeading } from "@/components/marketplace-ui";
import { errorMessage } from "@/lib/api/browser";

type Overview = components["schemas"]["AdminOverview"];

function AdminOverviewContent() {
  const { request } = useSession();
  const [overview, setOverview] = useState<Overview | null>(null);
  const [error, setError] = useState("");
  const load = useCallback(async () => {
    setError("");
    try {
      setOverview(await request<Overview>("/admin/overview"));
    } catch (cause) {
      setError(errorMessage(cause, "Marketplace overview unavailable."));
    }
  }, [request]);
  useEffect(() => {
    const timer = window.setTimeout(() => void load(), 0);
    return () => window.clearTimeout(timer);
  }, [load]);
  if (error)
    return (
      <ErrorState
        title="Overview unavailable"
        message={error}
        retry={() => void load()}
      />
    );
  if (!overview) return <LoadingState label="Loading admin overview" />;
  return (
    <>
      <PageHeading
        title="Marketplace overview"
        description={`Current operating totals as of ${new Date().toLocaleString("id-ID")}. No historical trend data.`}
      />
      <div className="admin-metrics">
        <Link href="/admin/users">
          <span>Users</span>
          <strong>{overview.users}</strong>
        </Link>
        <Link href="/admin/stores">
          <span>Approved stores</span>
          <strong>{overview.approvedStores}</strong>
        </Link>
        <Link href="/admin/products">
          <span>Published products</span>
          <strong>{overview.publishedProducts}</strong>
        </Link>
        <div>
          <span>Purchases</span>
          <strong>{overview.purchases}</strong>
        </div>
        <div>
          <span>Active seller orders</span>
          <strong>{overview.activeSellerOrders}</strong>
        </div>
        <div>
          <span>Delivered orders</span>
          <strong>{overview.deliveredOrders}</strong>
        </div>
        <div className="admin-gmv">
          <span>Gross merchandise value</span>
          <strong>
            <Money
              value={overview.grossMerchandiseMinor}
              currency={overview.currency}
            />
          </strong>
        </div>
      </div>
    </>
  );
}

export default function AdminPage() {
  return (
    <main className="workspace-page shell">
      <RequireRole role="admin">
        <AdminOverviewContent />
      </RequireRole>
    </main>
  );
}
