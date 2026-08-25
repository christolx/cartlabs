"use client";

import Link from "next/link";
import { actorHome, useSession } from "@/components/session-provider";

const roleLinks = {
  buyer: [["Cart", "/cart"], ["Purchases", "/purchases"], ["Notifications", "/notifications"]],
  seller: [["Home", "/seller"], ["Store", "/seller/store"], ["Products", "/seller/products"], ["Orders", "/seller/orders"], ["Notifications", "/notifications"]],
  admin: [["Overview", "/admin"], ["Stores", "/admin/stores"], ["Products", "/admin/products"], ["Users", "/admin/users"], ["Audit", "/admin/audit"]],
} as const;

export function SiteHeader() {
  const { status, user, logout } = useSession();
  const links = user ? roleLinks[user.role] : [["Browse", "/#catalog"], ["Stores", "/#catalog"], ["Demo", "/demo"]] as const;
  return <header className="site-header"><div className="header-inner">
    <Link className="wordmark" href={user ? actorHome(user.role) : "/"} aria-label="Cartlabs home"><span className="wordmark-text">Cartlabs</span></Link>
    <nav aria-label="Primary navigation">
      {links.map(([label, href]) => <Link key={label} href={href}>{label}</Link>)}
    </nav>
    <div className="session-nav">
      {status === "loading" ? <span className="session-loading">Loading session</span> : user ? <><span className="session-user"><strong>{user.displayName}</strong><small>{user.role}</small></span><button className="text-button" type="button" onClick={() => void logout()}>Log out</button></> : <Link className="header-login" href="/login">Sign in</Link>}
    </div>
    <Link className="cart-link" href="/cart" aria-label="Cart, 0 items"><svg aria-hidden="true" viewBox="0 0 32 32"><path d="M2 4h4l3 17h16l3-13H8M12 27a1.5 1.5 0 1 1-3 0 1.5 1.5 0 0 1 3 0Zm14 0a1.5 1.5 0 1 1-3 0 1.5 1.5 0 0 1 3 0Z" /></svg><span>0</span></Link>
  </div></header>;
}
