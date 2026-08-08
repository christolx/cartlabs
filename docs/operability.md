# Operability

Milestone 4 makes the modular monolith observable and recoverable before any
deployment or service extraction work begins.

## Local observability

Start the complete stack with `make compose-full`.

| Surface | Address | Purpose |
| --- | --- | --- |
| Cartlabs | `http://localhost:3000` | Production Next.js build |
| API metrics | `http://localhost:8080/metrics` | Bounded HTTP counters and histograms |
| Worker metrics | `http://localhost:9091` | Relay, delivery, reservation, and backlog signals |
| Prometheus | `http://localhost:9090` | Metrics storage and queries |
| Grafana | `http://localhost:3001` | Provisioned `Cartlabs Overview` dashboard |
| Tempo | `http://localhost:3200` | Trace search and retrieval |

API spans cover HTTP, PostgreSQL operations, payment HTTP calls, outbox relay,
and RabbitMQ publish/consume boundaries. W3C trace context crosses HTTP and AMQP.
Logs are JSON and include bounded route, status, duration, response size, request
ID, and trace ID. SQL text, authorization values, cookies, and request bodies are
not logged. `OTEL_TRACES_SAMPLER_ARG` accepts a ratio greater than zero and at
most one; local Compose deliberately samples every trace.

Dashboard panels track request rate, p95 latency, 5xx rate, pending/dead outbox
events, and notification success/retry/dead-letter outcomes. HTTP metric route
labels use router templates rather than raw paths, preventing ID-driven label
cardinality.

## Recovery policies

- SQL outbox delivery retries five times. Backoff doubles from 2 to 32 seconds,
  then marks the row dead-lettered without losing its payload.
- Notification handling retries transient failures three times through a
  durable two-second delay queue. Invalid input and exhausted deliveries move
  to the durable dead-letter queue.
- Publisher confirms precede outbox publication marks, consumer acknowledgements,
  and dead-letter replay acknowledgements.
- `make replay` moves at most `REPLAY_LIMIT` dead outbox rows and notification
  messages per run. Default is 100; accepted range is 1–1000. Fix poison-message
  cause before replaying.
- API readiness reconnects a lost RabbitMQ dependency. Redis rate-limit calls
  have a two-second operation deadline, so dependency loss returns 503 instead
  of tying up handlers.
- `make outage-test` verifies RabbitMQ durability/recovery, Redis degradation and
  fail-fast behavior, and payment-provider inventory rollback against real
  Compose services.

## Documented limits

| Boundary | Limit |
| --- | --- |
| HTTP headers | 1 MiB |
| Web proxy request body | 1 MiB |
| Web proxy upstream deadline | 10 seconds |
| API read/write timeout | 15/30 seconds |
| API idle timeout | 60 seconds |
| Authentication attempts | 10 per IP-and-subject per minute |
| Authenticated mutations | 120 per user per minute |
| Checkout reservation | 15 minutes |
| Outbox attempts | 5 |
| Notification retries | 3 |
| Replay batch | 1–1000, default 100 |

`make performance` runs a 15-second, 20-client public-catalog baseline and fails
above 250 ms p95 or 1% errors. Local Compose result on 2026-08-08: 201,682
requests, 13,445.0 requests/second, 0 failures, 1.38 ms p50, 2.48 ms p95, and
3.28 ms p99. This is a repeatable development baseline, not a production
capacity claim. Network latency, TLS, shared infrastructure, write contention,
and realistic traffic mixes remain outside this measurement.

## Security controls

API and web responses set content, framing, referrer, and content-type policies.
Web proxy forwards only allowlisted headers, caps body size, and applies an
upstream deadline. Runtime images use an unprivileged user. CI runs
`govulncheck`, `pnpm audit`, repository secret/misconfiguration scanning, and
HIGH/CRITICAL Trivy scans for every runtime image. Local dependency gate is
`make security`.

## Demo reset

`make demo-reset` recreates only local/demo database state from ordered
migrations and deterministic seed data. Reset refuses every other environment.
Install `deploy/cron/cartlabs-demo-reset.cron` on the demo host to run at 03:00
UTC. Reset is destructive to demo data; its schedule and logs should be visible
to demo users and operators.

## Incident checks

1. Check `/api/v1/health/live` and `/api/v1/health/ready`; readiness names the
   unavailable dependency.
2. Open `Cartlabs Overview`; correlate elevated errors or latency with outbox
   backlog and notification outcomes.
3. Search Tempo by service and time, then use trace ID to locate matching JSON
   logs.
4. Restore dependency and confirm readiness plus Prometheus `up` targets.
5. Inspect dead-letter cause. Run `make replay` only after correction.
6. Run `make outage-test`, `make performance`, and `make check` before closing a
   resilience incident caused by code changes.
