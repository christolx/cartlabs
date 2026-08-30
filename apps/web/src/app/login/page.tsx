"use client";

import { useEffect, useState, type FormEvent } from "react";
import { useRouter } from "next/navigation";
import type { components } from "@/lib/api/schema";
import { actorHome, useSession } from "@/components/session-provider";
import { errorMessage } from "@/lib/api/browser";

type Role = components["schemas"]["Role"];

const protectedDestinations: { prefix: string; roles: Role[] }[] = [
  { prefix: "/admin", roles: ["admin"] },
  { prefix: "/seller", roles: ["seller"] },
  { prefix: "/cart", roles: ["buyer"] },
  { prefix: "/checkout", roles: ["buyer"] },
  { prefix: "/purchases", roles: ["buyer"] },
  { prefix: "/notifications", roles: ["buyer", "seller"] },
];

function safeDestination(value: string | null, fallback: string, role: Role) {
  if (!value || !value.startsWith("/") || value.startsWith("//"))
    return fallback;
  const pathname = value.split(/[?#]/, 1)[0];
  const owner = protectedDestinations.find(
    (destination) =>
      pathname === destination.prefix ||
      pathname.startsWith(`${destination.prefix}/`),
  );
  return !owner || owner.roles.includes(role) ? value : fallback;
}

export default function LoginPage() {
  const { demoEnabled, status, user, login, demoLogin } = useSession();
  const router = useRouter();
  const [busy, setBusy] = useState(false);
  const [message, setMessage] = useState("");

  useEffect(() => {
    if (status === "authenticated" && user)
      router.replace(
        safeDestination(
          new URLSearchParams(window.location.search).get("next"),
          actorHome(user.role),
          user.role,
        ),
      );
  }, [router, status, user]);

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (busy) return;
    const data = new FormData(event.currentTarget);
    setBusy(true);
    setMessage("");
    try {
      const current = await login(
        String(data.get("email")),
        String(data.get("password")),
      );
      router.replace(
        safeDestination(
          new URLSearchParams(window.location.search).get("next"),
          actorHome(current.role),
          current.role,
        ),
      );
    } catch (error) {
      setMessage(errorMessage(error, "Sign in failed."));
    } finally {
      setBusy(false);
    }
  }

  async function quickLogin(role: Role) {
    if (busy) return;
    setBusy(true);
    setMessage("");
    try {
      const current = await demoLogin(role);
      router.replace(
        safeDestination(
          new URLSearchParams(window.location.search).get("next"),
          actorHome(current.role),
          current.role,
        ),
      );
    } catch (error) {
      setMessage(errorMessage(error, "Demo login failed."));
    } finally {
      setBusy(false);
    }
  }

  return (
    <main className="auth-page shell">
      <section className="auth-copy">
        <p className="eyebrow">Account access</p>
        <h1>Sign in to continue.</h1>
        <p>
          Buyers manage purchases. Sellers manage supply. Admins operate
          marketplace trust.
        </p>
        <div className="role-index-list" aria-label="Account roles">
          <div>
            <span>01</span>
            <strong>Buyer / purchase</strong>
          </div>
          <div>
            <span>02</span>
            <strong>Seller / supply</strong>
          </div>
          <div>
            <span>03</span>
            <strong>Admin / trust</strong>
          </div>
        </div>
      </section>
      <section className="auth-panel" aria-labelledby="login-heading">
        <h2 id="login-heading">Email and password</h2>
        <form className="stack-form" onSubmit={submit}>
          <label>
            <span>Email</span>
            <input
              name="email"
              type="email"
              autoComplete="email"
              required
              disabled={busy}
            />
          </label>
          <label>
            <span>Password</span>
            <input
              name="password"
              type="password"
              autoComplete="current-password"
              required
              disabled={busy}
            />
          </label>
          {message ? (
            <p className="form-message error-message" role="alert">
              {message}
            </p>
          ) : null}
          <button className="button button-primary" disabled={busy}>
            {busy ? "Signing in" : "Sign in"}
          </button>
        </form>
        {demoEnabled ? (
          <div className="demo-login">
            <h2>Demo accounts</h2>
            <p>
              Password for manual sign in: <code>demo-pass-123</code>
            </p>
            <div className="button-row">
              {(["buyer", "seller", "admin"] as Role[]).map((role) => (
                <button
                  key={role}
                  className="button button-secondary"
                  type="button"
                  disabled={busy}
                  onClick={() => void quickLogin(role)}
                >
                  Enter as {role}
                </button>
              ))}
            </div>
          </div>
        ) : null}
      </section>
    </main>
  );
}
