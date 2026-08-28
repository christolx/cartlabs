import type { Metadata } from "next";
import { PageHeading } from "@/components/marketplace-ui";

export const metadata: Metadata = {
  title: "About",
  description:
    "An engineering overview of Cartlabs: transactional checkout, role isolation, generated contracts, and observable services.",
};

const transactionPath = [
  [
    "Read authoritative state",
    "Reload published products, approved stores, prices, and available stock from PostgreSQL.",
  ],
  [
    "Reserve under lock",
    "Use row-level locks and atomic updates so concurrent checkout cannot oversell inventory.",
  ],
  [
    "Create the order graph",
    "Create one parent purchase and one seller order per store, with immutable item snapshots.",
  ],
  [
    "Verify payment",
    "Validate the mock provider signature, timestamp, reference, amount, currency, and event ID.",
  ],
  [
    "Publish and fulfill",
    "Commit domain state with an outbox fact, then let workers deliver events while sellers advance owned orders.",
  ],
] as const;

const engineeringHighlights = [
  [
    "Atomic checkout",
    "PostgreSQL transactions reserve SKU inventory and reject underflow before creating purchase state.",
  ],
  [
    "Idempotent mutations",
    "Idempotency keys and provider-event deduplication make retries safe across checkout and payment.",
  ],
  [
    "Seller isolation",
    "Transport and domain checks enforce role boundaries and prevent sellers from crossing store ownership.",
  ],
  [
    "Order state machines",
    "Purchase and seller-order transitions reject invalid moves with explicit conflict responses.",
  ],
  [
    "Resilient search",
    "The gRPC search service retrieves matching product IDs; PostgreSQL applies visibility and remains the fallback.",
  ],
  [
    "Verified reviews",
    "Only buyers with delivered items can review, and duplicate submissions are rejected.",
  ],
] as const;

const platformQualities = [
  [
    "Security",
    "Argon2id passwords, short-lived access tokens, rotating refresh cookies, rate limits, and immutable admin audit facts.",
  ],
  [
    "Contracts",
    "OpenAPI is the source of truth for generated Go and TypeScript types. CI rejects contract drift.",
  ],
  [
    "Observability",
    "OpenTelemetry traces, Prometheus metrics, structured logs, and dependency-aware health probes cover runtime behavior.",
  ],
  [
    "Deployment",
    "Docker images and Helm charts run the stack on a self-hosted single-node k3s cluster with migrations, probes, and optional telemetry.",
  ],
] as const;

export default function AboutPage() {
  return (
    <main className="workspace-page shell info-page about-page">
      <PageHeading
        family="ABOUT / ENGINEERING"
        index="01"
        title="A multi-vendor marketplace case study."
        description="Cartlabs is a deployable full-stack system running on self-hosted Kubernetes, built to explore transactional checkout, authorization, service boundaries, and operational recovery."
        meta={
          <>
            <span>Next.js + Go</span>
            <span>PostgreSQL + RabbitMQ</span>
            <span>Self-hosted k3s</span>
          </>
        }
      />

      <div className="info-stack">
        <section
          className="workspace-panel info-panel"
          aria-labelledby="architecture-title"
        >
          <div className="panel-heading-row">
            <div>
              <span className="route-family">Service boundary</span>
              <h2 id="architecture-title">
                Modular core, measured extraction.
              </h2>
              <p>
                The commerce core stays transactional. Search is the first
                extracted boundary because it tolerates eventual consistency.
              </p>
            </div>
          </div>
          <figure
            className="info-architecture"
            aria-labelledby="architecture-figure-title"
          >
            <figcaption className="info-architecture-caption">
              <div>
                <span className="route-family">Runtime shape</span>
                <h3 id="architecture-figure-title">
                  Request, state, events, extraction.
                </h3>
              </div>
              <p>
                Deployables run inside self-hosted k3s. Synchronous commerce
                stays authoritative in PostgreSQL; projections and delivery move
                through explicit boundaries.
              </p>
            </figcaption>

            <div
              className="info-architecture-legend"
              aria-label="Diagram legend"
            >
              <span>
                <i
                  className="info-architecture-legend-swatch info-architecture-legend-swatch--request"
                  aria-hidden="true"
                />
                HTTP / gRPC
              </span>
              <span>
                <i
                  className="info-architecture-legend-swatch info-architecture-legend-swatch--event"
                  aria-hidden="true"
                />
                versioned events
              </span>
              <span>
                <i
                  className="info-architecture-legend-swatch info-architecture-legend-swatch--authority"
                  aria-hidden="true"
                />
                final authority
              </span>
            </div>

            <div className="info-architecture-runtime">
              <div className="info-architecture-runtime-head">
                <div>
                  <span className="route-family">Deployment boundary</span>
                  <strong>Self-hosted k3s + Helm</strong>
                </div>
                <span className="count-badge">05 deployables</span>
              </div>
              <p className="info-architecture-runtime-components">
                web · api · worker · search · mock payment
              </p>

              <div className="info-architecture-map">
                <div className="info-architecture-lane">
                  <div className="info-architecture-lane-heading">
                    <span className="info-architecture-lane-index">01</span>
                    <div>
                      <h4>Request path</h4>
                      <p>Browser traffic stays REST at platform edge.</p>
                    </div>
                  </div>
                  <div className="info-architecture-lane-body">
                    <div className="info-architecture-flow info-architecture-flow--request">
                      <div className="info-architecture-node">
                        <span>Client</span>
                        <strong>Browser</strong>
                        <small>Same-origin app traffic</small>
                      </div>
                      <span className="info-architecture-connector">
                        <small>HTTP</small>
                        <b aria-hidden="true">→</b>
                      </span>
                      <div className="info-architecture-node">
                        <span>Web</span>
                        <strong>Next.js</strong>
                        <small>Generated TS + REST proxy</small>
                      </div>
                      <span className="info-architecture-connector">
                        <small>REST</small>
                        <b aria-hidden="true">→</b>
                      </span>
                      <div className="info-architecture-node info-architecture-node--accent">
                        <span>Core</span>
                        <strong>Go REST API</strong>
                        <small>Modules + authorization</small>
                      </div>
                    </div>
                  </div>
                </div>

                <div className="info-architecture-lane">
                  <div className="info-architecture-lane-heading">
                    <span className="info-architecture-lane-index">02</span>
                    <div>
                      <h4>State plane</h4>
                      <p>Commerce state and short-lived coordination.</p>
                    </div>
                  </div>
                  <div className="info-architecture-lane-body">
                    <div className="info-architecture-flow info-architecture-flow--dependency">
                      <div className="info-architecture-node">
                        <span>Core dependency</span>
                        <strong>Go REST API</strong>
                        <small>Queries, writes, visibility checks</small>
                      </div>
                      <span className="info-architecture-connector">
                        <small>SQL + cache</small>
                        <b aria-hidden="true">→</b>
                      </span>
                      <div className="info-architecture-targets">
                        <div className="info-architecture-node info-architecture-node--authority">
                          <span>Authority</span>
                          <strong>Catalog PostgreSQL</strong>
                          <small>Orders, inventory, users, stores</small>
                        </div>
                        <div className="info-architecture-node">
                          <span>Coordination</span>
                          <strong>Redis</strong>
                          <small>Cache, limits, short-lived state</small>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>

                <div className="info-architecture-lane">
                  <div className="info-architecture-lane-heading">
                    <span className="info-architecture-lane-index">03</span>
                    <div>
                      <h4>Async delivery</h4>
                      <p>Committed facts leave the request path safely.</p>
                    </div>
                  </div>
                  <div className="info-architecture-lane-body">
                    <div className="info-architecture-flow info-architecture-flow--async">
                      <div className="info-architecture-node">
                        <span>Commit boundary</span>
                        <strong>Transactional outbox</strong>
                        <small>Publishes after database commit</small>
                      </div>
                      <span className="info-architecture-connector">
                        <small>publish</small>
                        <b aria-hidden="true">→</b>
                      </span>
                      <div className="info-architecture-node">
                        <span>Broker</span>
                        <strong>RabbitMQ</strong>
                        <small>Retries + dead-letter queues</small>
                      </div>
                      <span className="info-architecture-connector">
                        <small>consume</small>
                        <b aria-hidden="true">→</b>
                      </span>
                      <div className="info-architecture-node">
                        <span>Runtime</span>
                        <strong>Go worker</strong>
                        <small>Idempotent jobs and events</small>
                      </div>
                    </div>
                  </div>
                </div>

                <div className="info-architecture-lane">
                  <div className="info-architecture-lane-heading">
                    <span className="info-architecture-lane-index">04</span>
                    <div>
                      <h4>Search boundary</h4>
                      <p>
                        Extracted retrieval, PostgreSQL visibility authority.
                      </p>
                    </div>
                  </div>
                  <div className="info-architecture-lane-body">
                    <div className="info-architecture-flow info-architecture-flow--search">
                      <div className="info-architecture-node">
                        <span>Caller</span>
                        <strong>Go REST API</strong>
                        <small>Filters final visible records</small>
                      </div>
                      <span className="info-architecture-connector">
                        <small>protobuf / gRPC</small>
                        <b aria-hidden="true">→</b>
                      </span>
                      <div className="info-architecture-node info-architecture-node--accent-soft">
                        <span>Independent service</span>
                        <strong>Search microservice</strong>
                        <small>Candidate retrieval + projection owner</small>
                      </div>
                      <span className="info-architecture-connector">
                        <small>projection</small>
                        <b aria-hidden="true">→</b>
                      </span>
                      <div className="info-architecture-node">
                        <span>Read model</span>
                        <strong>Search PostgreSQL</strong>
                        <small>Text projection, not commerce authority</small>
                      </div>
                    </div>
                    <p className="info-architecture-note">
                      Worker-published catalog events update the projection. If
                      gRPC is unavailable, API falls back to PostgreSQL text
                      search.
                    </p>
                  </div>
                </div>

                <div className="info-architecture-lane">
                  <div className="info-architecture-lane-heading">
                    <span className="info-architecture-lane-index">05</span>
                    <div>
                      <h4>Payment callback</h4>
                      <p>External behavior stays behind a signed interface.</p>
                    </div>
                  </div>
                  <div className="info-architecture-lane-body">
                    <div className="info-architecture-flow info-architecture-flow--payment">
                      <div className="info-architecture-node">
                        <span>Gateway simulator</span>
                        <strong>Mock payment</strong>
                        <small>Intent completion + retryable callback</small>
                      </div>
                      <span className="info-architecture-connector">
                        <small>signed webhook</small>
                        <b aria-hidden="true">→</b>
                      </span>
                      <div className="info-architecture-node info-architecture-node--accent">
                        <span>Verifier</span>
                        <strong>Go REST API</strong>
                        <small>Signature, amount, event dedupe</small>
                      </div>
                    </div>
                  </div>
                </div>
              </div>

              <div className="info-architecture-ops">
                <div className="info-architecture-ops-item">
                  <span className="route-family">Observability rail</span>
                  <strong>OpenTelemetry + Prometheus/Grafana</strong>
                  <small>
                    Traces, metrics, structured logs, health probes.
                  </small>
                </div>
                <div className="info-architecture-ops-item">
                  <span className="route-family">Operational boundary</span>
                  <strong>Helm-managed recovery</strong>
                  <small>
                    Migrations, readiness checks, and restartable services.
                  </small>
                </div>
              </div>
            </div>
          </figure>
        </section>

        <section
          className="workspace-panel info-panel"
          aria-labelledby="transaction-title"
        >
          <div className="panel-heading-row">
            <div>
              <span className="route-family">Transaction path</span>
              <h2 id="transaction-title">Checkout under contention.</h2>
              <p>
                One checkout crosses inventory, payment, orders, and async
                delivery without losing state between boundaries.
              </p>
            </div>
            <span className="count-badge">05 boundaries</span>
          </div>
          <ol className="record-list info-step-list">
            {transactionPath.map(([title, body], index) => (
              <li className="record-row info-step-row" key={title}>
                <span className="info-step-number" aria-hidden="true">
                  {String(index + 1).padStart(2, "0")}
                </span>
                <div>
                  <strong>{title}</strong>
                  <span>{body}</span>
                </div>
              </li>
            ))}
          </ol>
        </section>

        <div className="info-grid">
          <section
            className="workspace-panel info-panel"
            aria-labelledby="highlights-title"
          >
            <div className="panel-heading-row">
              <div>
                <span className="route-family">Feature surface</span>
                <h2 id="highlights-title">What to inspect</h2>
                <p>The mechanics that make this more than a catalog mockup.</p>
              </div>
              <span className="count-badge">06 areas</span>
            </div>
            <div className="record-list">
              {engineeringHighlights.map(([title, body]) => (
                <article className="record-row info-detail-row" key={title}>
                  <div>
                    <strong>{title}</strong>
                    <span>{body}</span>
                  </div>
                </article>
              ))}
            </div>
          </section>

          <section
            className="workspace-panel info-panel"
            aria-labelledby="qualities-title"
          >
            <div className="panel-heading-row">
              <div>
                <span className="route-family">System qualities</span>
                <h2 id="qualities-title">Built for inspection.</h2>
                <p>
                  Cross-cutting choices keep behavior testable and explainable.
                </p>
              </div>
              <span className="count-badge">04 areas</span>
            </div>
            <div className="record-list">
              {platformQualities.map(([title, body]) => (
                <article className="record-row info-detail-row" key={title}>
                  <div>
                    <strong>{title}</strong>
                    <span>{body}</span>
                  </div>
                </article>
              ))}
            </div>
          </section>
        </div>

        <section
          className="workspace-panel info-boundaries"
          aria-labelledby="boundaries-title"
        >
          <div>
            <span className="route-family">Scope</span>
            <h2 id="boundaries-title">Explicitly not production commerce.</h2>
            <p>
              The project demonstrates credible system behavior while keeping
              external integrations and operational scope honest.
            </p>
          </div>
          <ul>
            <li>
              Payment is simulated; no live financial provider is connected.
            </li>
            <li>
              IDR only; shipping, email, refunds, and promotions are deferred.
            </li>
            <li>
              Search is a read projection; PostgreSQL remains final authority.
            </li>
            <li>Seeded demo data is resettable and quick-login is opt-in.</li>
          </ul>
        </section>
      </div>
    </main>
  );
}
