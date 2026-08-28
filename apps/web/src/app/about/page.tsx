import type { Metadata } from "next";
import Link from "next/link";

export const metadata: Metadata = {
  title: "About",
  description:
    "How Cartlabs handles multi-seller checkout and marketplace operations.",
};

const decisions = [
  [
    "Order splitting",
    "One purchase becomes one seller order per store. Each seller fulfills only its own lines.",
  ],
  [
    "Inventory reservation",
    "Checkout atomically reserves each variant. Failed, expired, or eligible cancelled purchases release stock once.",
  ],
  [
    "Idempotent payments",
    "Checkout keys prevent duplicate reservation. Signed webhooks validate amount, currency, reference, freshness, and provider event ID.",
  ],
  [
    "Async work",
    "A transactional outbox commits domain facts with state changes. RabbitMQ workers retry delivery and route exhausted messages to dead-letter queues.",
  ],
  [
    "Role isolation",
    "Buyer, seller, and admin routes enforce role and ownership checks at HTTP and domain boundaries.",
  ],
  [
    "Search extraction",
    "A gRPC search service returns candidate IDs. PostgreSQL still decides visibility and provides fallback search when RPC fails.",
  ],
] as const;

export default function AboutPage() {
  return (
    <main className="info-page">
      <header className="info-hero">
        <p className="info-kicker">Marketplace engineering / 01</p>
        <h1>
          One checkout.
          <br />
          Many independent sellers.
        </h1>
        <p className="info-lede">
          Cartlabs keeps buyer payment unified while inventory, fulfillment, and
          ownership remain explicit per seller.
        </p>
      </header>

      <section className="lifecycle" aria-labelledby="lifecycle-title">
        <div className="section-heading">
          <span>01</span>
          <h2 id="lifecycle-title">Checkout lifecycle</h2>
        </div>
        <ol className="lifecycle-track">
          <li>
            <b>01</b>
            <strong>Validate cart</strong>
            <p>
              Reload published products, approved stores, prices, and available
              stock.
            </p>
          </li>
          <li>
            <b>02</b>
            <strong>Reserve inventory</strong>
            <p>
              Lock variants and record time-bound reservations in one
              transaction.
            </p>
          </li>
          <li>
            <b>03</b>
            <strong>Split by seller</strong>
            <p>
              Create one purchase plus independently fulfilled seller orders.
            </p>
          </li>
          <li>
            <b>04</b>
            <strong>Complete payment</strong>
            <p>Mock provider sends signed success or failure webhook.</p>
          </li>
          <li>
            <b>05</b>
            <strong>Fulfill separately</strong>
            <p>
              Each seller moves paid orders through processing, shipped, and
              delivered.
            </p>
          </li>
        </ol>
      </section>

      <section className="decision-section" aria-labelledby="decisions-title">
        <div className="section-heading">
          <span>02</span>
          <h2 id="decisions-title">Implemented decisions</h2>
        </div>
        <div className="decision-grid">
          {decisions.map(([title, body], index) => (
            <article key={title}>
              <b>{String(index + 1).padStart(2, "0")}</b>
              <h3>{title}</h3>
              <p>{body}</p>
            </article>
          ))}
        </div>
      </section>

      <section
        className="architecture-section"
        aria-labelledby="architecture-title"
      >
        <div className="section-heading">
          <span>03</span>
          <h2 id="architecture-title">Platform shape</h2>
        </div>
        <div
          className="architecture-map"
          role="img"
          aria-label="Browser connects to Next.js web and Go API. API uses catalog PostgreSQL, Redis, RabbitMQ workers, and gRPC search. Mock payment sends signed webhooks to API."
        >
          <div className="arch-node arch-browser">Browser</div>
          <span aria-hidden="true">→</span>
          <div className="arch-node">Next.js web</div>
          <span aria-hidden="true">→</span>
          <div className="arch-node arch-core">Go REST API</div>
          <span aria-hidden="true">→</span>
          <div className="arch-branches">
            <div className="arch-node">Catalog PostgreSQL</div>
            <div className="arch-node">Redis</div>
            <div className="arch-node">RabbitMQ + worker</div>
            <div className="arch-node">gRPC search + PostgreSQL</div>
            <div className="arch-node">Mock payment webhook</div>
          </div>
        </div>
      </section>

      <section className="operations" aria-labelledby="operations-title">
        <div className="section-heading">
          <span>04</span>
          <h2 id="operations-title">Operating profile</h2>
        </div>
        <div className="operations-grid">
          <article>
            <h3>Reliability</h3>
            <p>
              Atomic inventory changes, transactional outbox, idempotent
              consumers, bounded retries, health probes, metrics, traces, and
              structured logs.
            </p>
          </article>
          <article>
            <h3>Security</h3>
            <p>
              Argon2id passwords, short-lived access tokens, rotating refresh
              cookies, rate limits, resource ownership, immutable admin audit
              facts.
            </p>
          </article>
          <article>
            <h3>Stack</h3>
            <p>
              Next.js, strict TypeScript, Go, PostgreSQL, Redis, RabbitMQ, MinIO
              or Cloudinary, OpenAPI, protobuf, OpenTelemetry, Prometheus.
            </p>
          </article>
          <article>
            <h3>Deployment</h3>
            <p>
              Docker images and Helm charts target a single-node k3s demo
              platform with ingress, migrations, probes, and optional
              observability.
            </p>
          </article>
        </div>
      </section>

      <section className="boundaries" aria-labelledby="boundaries-title">
        <div>
          <h2 id="boundaries-title">Current boundaries</h2>
          <p>
            This project demonstrates marketplace mechanics, not production
            commerce coverage.
          </p>
        </div>
        <ul>
          <li>Payments use an internal mock provider.</li>
          <li>Money and checkout support IDR only.</li>
          <li>Quick-login exists only when demo mode is explicitly enabled.</li>
        </ul>
      </section>

      <nav className="doc-links" aria-label="Repository documentation">
        <a href="https://github.com/christolx/cartlabs/blob/main/docs/architecture.md">
          Architecture
        </a>
        <a href="https://github.com/christolx/cartlabs/blob/main/docs/fulfillment.md">
          Fulfillment
        </a>
        <a href="https://github.com/christolx/cartlabs/blob/main/docs/e2e-flow.md">
          Role journeys
        </a>
        <Link href="/docs">Use Cartlabs</Link>
      </nav>
    </main>
  );
}
