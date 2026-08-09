#!/usr/bin/env bash
set -euo pipefail

base_url=${1:?usage: smoke.sh https://cartlabs.example.com}
base_url=${base_url%/}
tmp_dir=$(mktemp -d)
trap 'rm -rf "$tmp_dir"' EXIT

for _ in $(seq 1 60); do
  if curl --fail --silent --show-error --max-time 10 "$base_url/api/v1/health/ready" >"$tmp_dir/ready.json"; then
    break
  fi
  sleep 2
done

jq -e '.status == "ok" and (.dependencies | to_entries | all(.value == "ok"))' "$tmp_dir/ready.json" >/dev/null
curl --fail --silent --show-error --max-time 10 "$base_url/" | grep -q 'Cartlabs'
curl --fail --silent --show-error --max-time 10 -D "$tmp_dir/headers" "$base_url/api/v1/catalog/products" >"$tmp_dir/catalog.json"
grep -qi '^x-request-id:' "$tmp_dir/headers"
jq -e '.items | length >= 2' "$tmp_dir/catalog.json" >/dev/null

buyer_token=$(curl --fail --silent --show-error --max-time 10 "$base_url/api/v1/auth/demo-login" \
  -H 'Content-Type: application/json' -d '{"role":"buyer"}' | jq -r .accessToken)
[[ -n "$buyer_token" && "$buyer_token" != null ]]
curl --fail --silent --show-error --max-time 10 "$base_url/api/v1/cart" \
  -H "Authorization: Bearer $buyer_token" | jq -e '.stores | type == "array"' >/dev/null

echo "deployment smoke checks passed for $base_url"
