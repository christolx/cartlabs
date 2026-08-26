"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import type { components } from "@/lib/api/schema";
import { actorHome } from "@/components/session-provider";

type Role = components["schemas"]["Role"];

const links: Record<Role, readonly [string, string][]> = {
  buyer: [
    ["Cart", "/cart"],
    ["Purchases", "/purchases"],
    ["Notifications", "/notifications"],
  ],
  seller: [
    ["Home", "/seller"],
    ["Store", "/seller/store"],
    ["Products", "/seller/products"],
    ["Orders", "/seller/orders"],
    ["Notifications", "/notifications"],
  ],
  admin: [
    ["Overview", "/admin"],
    ["Stores", "/admin/stores"],
    ["Products", "/admin/products"],
    ["Users", "/admin/users"],
    ["Audit", "/admin/audit"],
  ],
};

export function WorkspaceNav({ role }: { role: Role }) {
  const pathname = usePathname();
  const roleLinks = links[role];
  return (
    <nav
      className="workspace-subnav"
      aria-label={`${role} workspace navigation`}
    >
      <Link className="workspace-subnav-label" href={actorHome(role)}>
        <span aria-hidden="true">
          {role === "buyer" ? "01" : role === "seller" ? "02" : "03"}
        </span>
        {role} workspace
      </Link>
      <div className="workspace-subnav-links">
        {roleLinks.map(([label, href]) => {
          const active =
            pathname === href ||
            (href !== actorHome(role) && pathname.startsWith(`${href}/`));
          return (
            <Link
              key={href}
              href={href}
              aria-current={active ? "page" : undefined}
            >
              {label}
            </Link>
          );
        })}
      </div>
    </nav>
  );
}
