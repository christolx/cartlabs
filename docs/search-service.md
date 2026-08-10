# Search service

## Ownership

| Concern | Owner |
| --- | --- |
| Product name, description, category text projection | Search database |
| Product lifecycle, moderation, store approval | Catalog database |
| Variants, price, currency, inventory, images | Catalog database |
| Public REST response and pagination | API/catalog |
| Protobuf RPC schema | `api/proto/search/v1/search.proto` |
| Projection event | `catalog.search.upsert.v1` |

Search returns candidate product IDs. Candidate IDs are never authorization or
visibility decisions.

## Data flow

```text
seller write
  -> catalog PostgreSQL transaction
       -> product row
       -> catalog.search.upsert.v1 outbox row
  -> worker relay -> RabbitMQ cartlabs.events
  -> search consumer (bounded retry/DLQ)
  -> authenticated gRPC UpsertProduct
  -> cartlabs_search.search_documents

buyer text query
  -> REST API -> gRPC SearchProducts -> candidate IDs
  -> catalog PostgreSQL applies visibility/price/stock/filter rules
  -> unchanged REST response
```

Events are idempotent by UUID. Document updates apply only when `updated_at` is
not older than stored version. Retry queue delays two seconds and dead-letters
after three failed deliveries. `make search-reindex` repairs missing/dead-letter
state and safely removes stale documents.

## Compatibility and failure behavior

- Empty query keeps original catalog SQL path.
- Search RPC deadline: 750 ms.
- First transport failure opens local API circuit for two seconds.
- RPC failure, open circuit, or truncated candidate response uses legacy SQL
  substring search.
- API increments `cartlabs_catalog_search_total{result="fallback"}` and logs one
  structured warning for fallback requests.
- Worker delivery exposes `cartlabs_search_delivery_total`; service exposes
  request latency/results, upsert results, and document count.
- Search readiness checks owned PostgreSQL. gRPC and PostgreSQL calls emit
  OpenTelemetry spans.

Public behavior remains correct during projection lag or service outage because
catalog database is final authority. Search failure does not affect empty-query
catalog browsing, product detail, checkout, orders, or payments.

## Local operation

```bash
make compose-up
make migrate
make seed
make search-dev
make search-reindex
```

Complete containers:

```bash
make compose-full
docker compose -f deploy/compose/compose.yml --profile full --profile ops run --rm search-reindex
make microservice-test
make search-outage
```

Required settings:

- `SEARCH_GRPC_ADDR`: client endpoint; empty disables extraction and keeps legacy search.
- `SEARCH_GRPC_SERVER_ADDR`: gRPC listener.
- `SEARCH_METRICS_ADDR`: HTTP health/metrics listener.
- `SEARCH_DATABASE_URL`: search-owned database URL.
- `SEARCH_SERVICE_TOKEN`: minimum 32-byte internal bearer token; local default is rejected outside local/test.

## Migration, repair, rollback

1. Deploy search migration and service while API `SEARCH_GRPC_ADDR` remains empty.
2. Run reindex; compare search document count with catalog product count.
3. Enable worker consumer, then API search client.
4. Watch fallback, delivery retry/DLQ, upsert failure, and document-count metrics.
5. Repair drift or dead letters with `make search-reindex`.

Rollback needs no REST or catalog schema rollback: clear `SEARCH_GRPC_ADDR` on API
and worker. Legacy SQL search immediately remains authoritative. Search service,
queue, and database can stay for investigation, then be removed after retention
requirements. Never point `SEARCH_DATABASE_URL` at `postgres`, a template
database, or main `cartlabs`; migration command rejects unsafe system targets.

Current reindex accepts at most 100,000 IDs in one prune request. Above that
size, replace prune contract with streamed generations before scaling catalog.
