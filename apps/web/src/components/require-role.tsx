"use client";

import { useEffect, type ReactNode } from "react";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import type { components } from "@/lib/api/schema";
import { LoadingState } from "@/components/async-state";
import { actorHome, useSession } from "@/components/session-provider";
import { WorkspaceNav } from "@/components/workspace-nav";

type Role = components["schemas"]["Role"];

export function RequireRole({
  role,
  children,
}: {
  role: Role;
  children: ReactNode;
}) {
  const { status, user } = useSession();
  const router = useRouter();
  const pathname = usePathname();

  useEffect(() => {
    if (status !== "anonymous") return;
    const query = window.location.search.slice(1);
    const intended = `${pathname}${query ? `?${query}` : ""}`;
    router.replace(`/login?next=${encodeURIComponent(intended)}`);
  }, [pathname, router, status]);

  if (status === "loading" || status === "anonymous")
    return <LoadingState label="Checking access" />;
  if (!user || user.role !== role) {
    return (
      <div className="state-panel forbidden-state">
        <h1>Access restricted</h1>
        <p>
          This page requires {role} access. Signed in as{" "}
          {user?.role ?? "another role"}.
        </p>
        <Link
          className="button button-primary"
          href={user ? actorHome(user.role) : "/login"}
        >
          Go to your workspace
        </Link>
      </div>
    );
  }
  return (
    <>
      <WorkspaceNav role={role} />
      {children}
    </>
  );
}

export function RequireAuth({ children }: { children: ReactNode }) {
  const { status, user } = useSession();
  const router = useRouter();
  const pathname = usePathname();

  useEffect(() => {
    if (status !== "anonymous") return;
    const query = window.location.search.slice(1);
    const intended = `${pathname}${query ? `?${query}` : ""}`;
    router.replace(`/login?next=${encodeURIComponent(intended)}`);
  }, [pathname, router, status]);

  if (status !== "authenticated")
    return <LoadingState label="Checking access" />;
  return (
    <>
      {user ? <WorkspaceNav role={user.role} /> : null}
      {children}
    </>
  );
}
