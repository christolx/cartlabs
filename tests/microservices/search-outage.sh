#!/usr/bin/env bash
set -euo pipefail

compose_file=${COMPOSE_FILE:-deploy/compose/compose.yml}
api_url=${E2E_API_URL:-http://localhost:8080/api/v1}

baseline=$(curl --fail --silent --show-error "$api_url/catalog/products?q=basket&page=1&pageSize=20" | jq -c '{total,names:[.items[].name]}')
if [[ $(jq -r .total <<<"$baseline") -lt 1 ]]; then
  echo "search baseline returned no product" >&2
  exit 1
fi

restore_search() {
  docker compose -f "$compose_file" --profile full start search >/dev/null
}
trap restore_search EXIT

docker compose -f "$compose_file" --profile full stop search >/dev/null
fallback=$(curl --fail --silent --show-error "$api_url/catalog/products?q=basket&page=1&pageSize=20" | jq -c '{total,names:[.items[].name]}')
if [[ "$fallback" != "$baseline" ]]; then
  echo "compatibility fallback changed response: baseline=$baseline fallback=$fallback" >&2
  exit 1
fi
metrics=$(curl --fail --silent --show-error "${api_url%/api/v1}/metrics")
if ! grep -q 'cartlabs_catalog_search_total{result="fallback"}' <<<"$metrics"; then
  echo "fallback metric missing" >&2
  exit 1
fi

restore_search
for _ in {1..30}; do
  if curl --fail --silent http://localhost:9093/health/ready >/dev/null; then
    break
  fi
  sleep 1
done
recovered=$(curl --fail --silent --show-error "$api_url/catalog/products?q=basket&page=1&pageSize=20" | jq -c '{total,names:[.items[].name]}')
if [[ "$recovered" != "$baseline" ]]; then
  echo "recovered search changed response: baseline=$baseline recovered=$recovered" >&2
  exit 1
fi

trap - EXIT
echo "search outage compatibility passed: $baseline"
