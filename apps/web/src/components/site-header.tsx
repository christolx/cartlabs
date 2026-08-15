"use client";

import Link from "next/link";
import { actorHome, useSession } from "@/components/session-provider";

const roleLinks = {
  buyer: [["Cart", "/cart"], ["Purchases", "/purchases"], ["Notifications", "/notifications"]],
  seller: [["Home", "/seller"], ["Store", "/seller/store"], ["Products", "/seller/products"], ["Orders", "/seller/orders"], ["Notifications", "/notifications"]],
  admin: [["Overview", "/admin"], ["Stores", "/admin/stores"], ["Products", "/admin/products"], ["Users", "/admin/users"], ["Audit", "/admin/audit"], ["Notifications", "/notifications"]],
} as const;

export function SiteHeader() {
  const { status, user, logout } = useSession();
  const links = user ? roleLinks[user.role] : [["Catalog", "/#catalog"], ["Demo", "/demo"]] as const;
  return <header className="site-header"><div className="shell header-inner">
    <Link className="wordmark" href={user ? actorHome(user.role) : "/"} aria-label="Cartlabs home">cartlabs<span>/market</span></Link>
    <nav aria-label="Primary navigation">
      {links.map(([label, href]) => <Link key={href} href={href}>{label}</Link>)}
    </nav>
    <div className="session-nav">
      {status === "loading" ? <span className="session-loading">Loading session</span> : user ? <><span className="session-user"><strong>{user.displayName}</strong><small>{user.role}</small></span><button className="text-button" type="button" onClick={() => void logout()}>Log out</button></> : <Link className="button button-primary header-login" href="/login">Sign in</Link>}
    </div>
  </div></header>;
}
