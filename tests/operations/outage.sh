#!/usr/bin/env bash
set -euo pipefail

repo_dir=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
compose_file="$repo_dir/deploy/compose/compose.yml"
api_url=${E2E_API_URL:-http://localhost:8080/api/v1}
tmp_dir=$(mktemp -d)

restore() {
  docker compose -f "$compose_file" --profile full start redis rabbitmq mock-payment >/dev/null 2>&1 || true
  docker compose -f "$compose_file" --profile full up -d worker >/dev/null 2>&1 || true
  rm -rf "$tmp_dir"
}
trap restore EXIT

wait_status() {
  local url=$1 expected=$2
  for _ in $(seq 1 30); do
    if [[ $(curl --silent --output /dev/null --write-out '%{http_code}' --max-time 5 "$url" || true) == "$expected" ]]; then
      return 0
    fi
    sleep 1
  done
  echo "timed out waiting for $url status $expected" >&2
  return 1
}

cd "$repo_dir"
make reset >/dev/null

buyer_token=$(curl --fail --silent --show-error "$api_url/auth/demo-login" \
  -H 'Content-Type: application/json' -d '{"role":"buyer"}' | jq -r .accessToken)
auth_header="Authorization: Bearer $buyer_token"

curl --fail --silent --show-error -X PUT "$api_url/cart/items/01989f00-0000-7000-8000-000000000401" \
  -H "$auth_header" -H 'Content-Type: application/json' -d '{"quantity":1}' >/dev/null

docker compose -f "$compose_file" --profile full stop rabbitmq >/dev/null
wait_status "$api_url/health/ready" 503
wait_status "$api_url/health/live" 200

purchase=$(curl --fail --silent --show-error -X POST "$api_url/checkout" \
  -H "$auth_header" -H "Idempotency-Key: outage-rabbit-$(date +%s)")
purchase_id=$(jq -r .id <<<"$purchase")
curl --fail --silent --show-error -X POST "$api_url/purchases/$purchase_id/pay" \
  -H "$auth_header" -H 'Content-Type: application/json' -d '{"outcome":"succeeded"}' >/dev/null

docker compose -f "$compose_file" --profile full start rabbitmq >/dev/null
docker compose -f "$compose_file" --profile full up -d worker >/dev/null
wait_status "$api_url/health/ready" 200
for _ in $(seq 1 30); do
  if curl --fail --silent --show-error "$api_url/notifications" -H "$auth_header" | jq -e '.items | any(.kind == "purchase.paid")' >/dev/null; then
    break
  fi
  sleep 1
done
curl --fail --silent --show-error "$api_url/notifications" -H "$auth_header" | jq -e '.items | any(.kind == "purchase.paid")' >/dev/null

docker compose -f "$compose_file" --profile full stop redis >/dev/null
wait_status "$api_url/health/ready" 503
curl --fail --silent --show-error "$api_url/cart" -H "$auth_header" >/dev/null
login_status=$(curl --silent --output "$tmp_dir/login.json" --write-out '%{http_code}' --max-time 15 \
  -X POST "$api_url/auth/demo-login" -H 'Content-Type: application/json' -d '{"role":"buyer"}' || true)
[[ "$login_status" == 503 ]]
docker compose -f "$compose_file" --profile full start redis >/dev/null
wait_status "$api_url/health/ready" 200

curl --fail --silent --show-error -X PUT "$api_url/cart/items/01989f00-0000-7000-8000-000000000402" \
  -H "$auth_header" -H 'Content-Type: application/json' -d '{"quantity":1}' >/dev/null
stock_before=$(curl --fail --silent --show-error "$api_url/catalog/products/indigo-utility-tray" | jq '.variants[0].stock')
docker compose -f "$compose_file" --profile full stop mock-payment >/dev/null
checkout_status=$(curl --silent --output "$tmp_dir/checkout.json" --write-out '%{http_code}' --max-time 15 \
  -X POST "$api_url/checkout" -H "$auth_header" -H "Idempotency-Key: outage-payment-$(date +%s)" || true)
[[ "$checkout_status" == 503 ]]
stock_after=$(curl --fail --silent --show-error "$api_url/catalog/products/indigo-utility-tray" | jq '.variants[0].stock')
[[ "$stock_before" == "$stock_after" ]]
docker compose -f "$compose_file" --profile full start mock-payment >/dev/null
wait_status http://localhost:8081/health/live 200

echo "outage recovery checks passed"
