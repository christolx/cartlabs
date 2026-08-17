# 0002 — Extract text search behind candidate-ID contract

Status: accepted

## Context

Milestone 4 measured public catalog listing at 13,445 requests/second with 2.48
ms p95. No performance problem requires distribution. Milestone 6 instead needs
one low-risk boundary that teaches service ownership, RPC, asynchronous state
replication, failure handling, and migration without splitting order or payment
transactions.

Search is read-only, publicly exercised, and can tolerate eventual consistency.
Notifications were considered but remain coupled to user recipients and message
presentation. Orders and payments contain central consistency rules and are
explicitly excluded from first extraction.

## Decision

Extract text candidate retrieval into `apps/search`, an internal protobuf/gRPC
service. Search owns `cartlabs_search`, including `search_documents` and its
processed-event ledger. It returns product IDs only. Catalog PostgreSQL remains
authoritative for publication, listing enforcement, store verification, category, price,
inventory, image, response shape, ordering, and pagination.

Catalog create, content update, publish, archive, suspend, and reinstate
transactions append `catalog.search.upsert.v1` to the
existing outbox. RabbitMQ delivers events to an independently retried search
consumer. Search applies event IDs idempotently and rejects older document
versions. Operator reindex reads a bounded source snapshot, upserts every
document, then prunes stale rows at or before the snapshot cutoff. Concurrent
catalog writes after the cutoff survive pruning.

API uses search candidates only for non-empty text queries. RPC has a 750 ms
deadline and a two-second client circuit breaker. Unavailability, oversized
candidate sets, or RPC failure triggers existing PostgreSQL `ILIKE` search, so
the public REST contract remains compatible. Internal RPCs require a dedicated
bearer token and remain cluster-private. gRPC, PostgreSQL, event, and fallback
paths expose traces and bounded Prometheus metrics.

## Consequences

- Search can evolve its index without owning marketplace truth.
- Eventual index lag cannot expose drafts, archived or suspended products,
  unavailable stock, or unverified stores because catalog re-applies those rules.
- Search and main catalog databases can share one PostgreSQL server for demo
  cost, but use distinct databases, URLs, migrations, repositories, and backups.
- Search outage adds at most one RPC deadline per circuit-breaker interval;
  fallback preserves correctness at lower efficiency.
- Reindex is deliberate migration/repair tooling, not a normal cross-database
  runtime dependency.
- Full-text ranking, typo tolerance, deletion events, and external search engines
  remain future work. Current substring semantics match existing REST behavior.
