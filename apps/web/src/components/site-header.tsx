"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { usePathname } from "next/navigation";
import { useSession } from "@/components/session-provider";

const publicLinks = [
  ["Home", "/"],
  ["About", "/about"],
  ["Docs", "/docs"],
] as const;

const roleLinks = {
  buyer: [
    ["Cart", "/cart"],
    ["Purchases", "/purchases"],
    ["Notifications", "/notifications"],
  ],
  seller: [
    ["Dashboard", "/seller"],
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
} as const;

export function SiteHeader() {
  const { status, user, logout, request } = useSession();
  const pathname = usePathname();
  const [cartQuantity, setCartQuantity] = useState<number | null>(null);
  useEffect(() => {
    let active = true;
    const onCartUpdate = (event: Event) => {
      const quantity = (event as CustomEvent<{ quantity?: number }>).detail
        ?.quantity;
      if (typeof quantity === "number") setCartQuantity(quantity);
    };
    window.addEventListener("cart:updated", onCartUpdate);
    if (status === "authenticated" && user?.role === "buyer") {
      void request<{ totalQuantity: number }>("/cart")
        .then((cart) => {
          if (active) setCartQuantity(cart.totalQuantity);
        })
        .catch(() => {
          if (active) setCartQuantity(null);
        });
    } else {
      const resetTimer = window.setTimeout(() => {
        if (active) setCartQuantity(null);
      }, 0);
      return () => {
        active = false;
        window.clearTimeout(resetTimer);
        window.removeEventListener("cart:updated", onCartUpdate);
      };
    }
    return () => {
      active = false;
      window.removeEventListener("cart:updated", onCartUpdate);
    };
  }, [request, status, user?.role]);
  function renderNavLink(
    [label, href]: readonly [string, string],
    matchDescendants = true,
  ) {
    return (
      <Link
        key={label}
        href={href}
        aria-current={
          href === "/"
            ? pathname === "/"
              ? "page"
              : undefined
            : pathname === href ||
                (matchDescendants && pathname.startsWith(`${href}/`))
              ? "page"
              : undefined
        }
      >
        {label}
      </Link>
    );
  }

  return (
    <header className="site-header">
      <div className="header-inner">
        <Link className="wordmark" href="/" aria-label="Cartlabs home">
          <span className="wordmark-text">Cartlabs</span>
        </Link>
        <nav aria-label="Primary navigation">
          <div className="nav-global">
            {publicLinks.map((link) => renderNavLink(link))}
          </div>
          {user ? (
            <div
              className="nav-workspace"
              role="group"
              aria-label={`${user.role} workspace`}
            >
              <span className="nav-role-label">{user.role}</span>
              {roleLinks[user.role].map((link, index) =>
                renderNavLink(link, index > 0),
              )}
            </div>
          ) : null}
        </nav>
        <div className="session-nav">
          {status === "loading" ? (
            <span className="session-loading">Loading</span>
          ) : user ? (
            <>
              <span className="session-user">
                <strong>{user.displayName}</strong>
                <small>{user.role}</small>
              </span>
              <button
                className="text-button"
                type="button"
                onClick={() => void logout()}
              >
                Log out
              </button>
            </>
          ) : (
            <Link
              className="header-login"
              href="/login"
              aria-current={pathname === "/login" ? "page" : undefined}
            >
              Sign in
            </Link>
          )}
        </div>
        <Link
          className="cart-link"
          href="/cart"
          aria-label={
            cartQuantity === null ? "Cart" : `Cart, ${cartQuantity} items`
          }
        >
          <svg aria-hidden="true" viewBox="0 0 32 32">
            <path d="M2 4h4l3 17h16l3-13H8M12 27a1.5 1.5 0 1 1-3 0 1.5 1.5 0 0 1 3 0Zm14 0a1.5 1.5 0 1 1-3 0 1.5 1.5 0 0 1 3 0Z" />
          </svg>
          {cartQuantity !== null ? <span>{cartQuantity}</span> : null}
        </Link>
      </div>
    </header>
  );
}
