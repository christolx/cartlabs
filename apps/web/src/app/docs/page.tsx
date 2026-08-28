import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Docs",
  description: "Role workflows, statuses, and demo limits for Cartlabs.",
};

const sections = [
  ["getting-started", "Getting started"],
  ["buyer", "Buyer workflow"],
  ["seller", "Seller workflow"],
  ["admin", "Admin workflow"],
  ["statuses", "Order and payment statuses"],
  ["demo", "Demo behavior and limitations"],
] as const;

export default function DocsPage() {
  return (
    <main className="info-page docs-page">
      <header className="docs-hero">
        <p className="info-kicker">Product guide / 02</p>
        <h1>Use Cartlabs by role.</h1>
        <p>
          Six focused sections cover local setup, marketplace workflows, state
          transitions, and demo constraints.
        </p>
      </header>
      <div className="docs-layout">
        <nav className="docs-toc" aria-label="On this page">
          <strong>On this page</strong>
          {sections.map(([id, label], i) => (
            <a key={id} href={`#${id}`}>
              <span>{String(i + 1).padStart(2, "0")}</span>
              {label}
            </a>
          ))}
        </nav>
        <article className="docs-content">
          <section id="getting-started">
            <h2>Getting started</h2>
            <ol>
              <li>
                Copy <code>.env.example</code> to <code>.env</code>.
              </li>
              <li>
                Run <code>make setup</code>, then <code>make compose-up</code>.
              </li>
              <li>
                Run <code>make migrate</code> and <code>make dev</code>.
              </li>
              <li>
                After search starts, run <code>make seed</code> in another
                terminal.
              </li>
              <li>
                Open storefront, browse catalog, or sign in with a seeded
                account.
              </li>
            </ol>
            <p>
              Use <code>make check</code> for generated-contract, test, lint,
              and build checks.
            </p>
          </section>
          <section id="buyer">
            <h2>Buyer workflow</h2>
            <ol>
              <li>Browse or filter published products from approved stores.</li>
              <li>Choose variant and quantity, then add item to cart.</li>
              <li>Review cart grouped by seller.</li>
              <li>
                Checkout with recipient and delivery address. Inventory becomes
                reserved.
              </li>
              <li>Complete mock payment with success or failure.</li>
              <li>Track purchase and each seller order independently.</li>
              <li>
                Cancel an eligible purchase or review a delivered item once.
              </li>
            </ol>
          </section>
          <section id="seller">
            <h2>Seller workflow</h2>
            <ol>
              <li>Sign in as seller and create or update store profile.</li>
              <li>Create products, variants, inventory, and product images.</li>
              <li>
                Publish complete products after store approval; archive when
                unavailable.
              </li>
              <li>See only orders belonging to owned store.</li>
              <li>Move paid orders to processing, shipped, then delivered.</li>
              <li>
                Cancel paid or processing orders with a reason. Eligible stock
                is restored.
              </li>
            </ol>
          </section>
          <section id="admin">
            <h2>Admin workflow</h2>
            <ol>
              <li>
                Review platform totals and active gross merchandise value.
              </li>
              <li>Approve or reject stores.</li>
              <li>Suspend or reinstate product listings.</li>
              <li>
                Suspend or reactivate users while preserving one active admin.
              </li>
              <li>
                Inspect append-only audit timeline across moderation and
                fulfillment.
              </li>
            </ol>
          </section>
          <section id="statuses">
            <h2>Order and payment statuses</h2>
            <div className="status-groups">
              <div>
                <h3>Purchase</h3>
                <dl>
                  <dt>
                    <code>pending_payment</code>
                  </dt>
                  <dd>Inventory reserved; payment unresolved.</dd>
                  <dt>
                    <code>paid</code>
                  </dt>
                  <dd>Payment verified and reservations converted.</dd>
                  <dt>
                    <code>payment_failed</code>
                  </dt>
                  <dd>Payment failed; reservations released.</dd>
                  <dt>
                    <code>expired</code>
                  </dt>
                  <dd>Reservation window elapsed before payment.</dd>
                  <dt>
                    <code>cancelled</code>
                  </dt>
                  <dd>Buyer cancellation completed where eligible.</dd>
                </dl>
              </div>
              <div>
                <h3>Payment</h3>
                <dl>
                  <dt>
                    <code>pending</code>
                  </dt>
                  <dd>Provider intent exists.</dd>
                  <dt>
                    <code>succeeded</code>
                  </dt>
                  <dd>Signed success event accepted.</dd>
                  <dt>
                    <code>failed</code>
                  </dt>
                  <dd>Signed failure event accepted.</dd>
                  <dt>
                    <code>expired</code>
                  </dt>
                  <dd>Reservation expired.</dd>
                  <dt>
                    <code>cancelled</code>
                  </dt>
                  <dd>Pending purchase cancelled.</dd>
                </dl>
              </div>
              <div>
                <h3>Seller order</h3>
                <dl>
                  <dt>
                    <code>pending_payment</code>
                  </dt>
                  <dd>Waiting for parent payment.</dd>
                  <dt>
                    <code>paid</code>
                  </dt>
                  <dd>Ready for seller action.</dd>
                  <dt>
                    <code>processing</code>
                  </dt>
                  <dd>Seller preparing order.</dd>
                  <dt>
                    <code>shipped</code>
                  </dt>
                  <dd>Order dispatched; cancellation closed.</dd>
                  <dt>
                    <code>delivered</code>
                  </dt>
                  <dd>Fulfillment complete; review eligible.</dd>
                  <dt>
                    <code>cancelled</code>
                  </dt>
                  <dd>
                    Seller order stopped and stock restored where eligible.
                  </dd>
                </dl>
              </div>
            </div>
          </section>
          <section id="demo">
            <h2>Demo behavior and limitations</h2>
            <ul>
              <li>
                Role quick-login appears only with <code>DEMO_MODE=true</code>{" "}
                or matching public web flag.
              </li>
              <li>
                Production fallback-token configuration prevents demo mode.
              </li>
              <li>
                Seeded buyer, seller, and admin accounts share local-only
                password <code>demo-pass-123</code>.
              </li>
              <li>
                Mock payment lets user choose success or failure; no money
                moves.
              </li>
              <li>Checkout currency is IDR.</li>
              <li>
                Seed reset can replace demo changes. Treat local content as
                disposable.
              </li>
            </ul>
          </section>
        </article>
      </div>
    </main>
  );
}
