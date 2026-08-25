"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import type { components } from "@/lib/api/schema";
import { actorHome, useSession } from "@/components/session-provider";
import { errorMessage } from "@/lib/api/browser";

type Role = components["schemas"]["Role"];

const accounts: { role: Role; email: string; purpose: string }[] = [
  {
    role: "buyer",
    email: "buyer@demo.cartlabs.local",
    purpose: "Discover products, checkout, pay, track, and review.",
  },
  {
    role: "seller",
    email: "seller@demo.cartlabs.local",
    purpose: "Create store, products, inventory, and fulfill orders.",
  },
  {
    role: "admin",
    email: "admin@demo.cartlabs.local",
    purpose: "Approve supply, manage users, and inspect audit history.",
  },
];

export function DemoEntry() {
  const { demoEnabled, demoLogin } = useSession();
  const router = useRouter();
  const [busy, setBusy] = useState<Role | "">("");
  const [message, setMessage] = useState("");

  async function enter(role: Role) {
    setBusy(role);
    setMessage("");
    try {
      const current = await demoLogin(role);
      router.push(actorHome(current.role));
    } catch (cause) {
      setMessage(errorMessage(cause, "Demo login failed."));
      setBusy("");
    }
  }

  return (
    <>
      <section className="demo-flow" aria-label="Recommended demo order">
        {accounts.map((account, index) => (
          <article key={account.role}>
            <span>{index + 1}</span>
            <h2>{account.role}</h2>
            <p>{account.purpose}</p>
            <dl>
              <div>
                <dt>Email</dt>
                <dd>{account.email}</dd>
              </div>
              <div>
                <dt>Password</dt>
                <dd>demo-pass-123</dd>
              </div>
            </dl>
            {demoEnabled ? (
              <button
                className="button button-primary"
                type="button"
                disabled={Boolean(busy)}
                onClick={() => void enter(account.role)}
              >
                {busy === account.role ? "Opening" : `Enter ${account.role}`}
              </button>
            ) : (
              <p className="helper-text">
                Quick login disabled. Use credentials on sign-in page if account
                access is configured.
              </p>
            )}
          </article>
        ))}
      </section>
      <section className="demo-note">
        <h2>Temporary demo data</h2>
        <p>
          Reset jobs can replace marketplace state. Business actions happen only
          on normal actor pages, never on this guide.
        </p>
      </section>
      {message ? (
        <p className="form-message error-message" role="alert">
          {message}
        </p>
      ) : null}
    </>
  );
}
