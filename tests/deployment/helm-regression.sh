#!/bin/sh
set -eu

chart=deploy/helm
values=$chart/values-local.yaml
work_dir=$(mktemp -d)
trap 'rm -rf "$work_dir"' EXIT HUP INT TERM

helm template cartlabs "$chart" --namespace cartlabs --values "$values" \
  --show-only templates/networkpolicy.yaml >"$work_dir/networkpolicy.yaml"
helm template cartlabs "$chart" --namespace cartlabs --values "$values" \
  --show-only templates/search-jobs.yaml >"$work_dir/search-jobs.yaml"
helm template cartlabs "$chart" --namespace cartlabs --values "$values" \
  --set observability.enabled=true \
  --show-only templates/observability.yaml >"$work_dir/observability.yaml"

if helm lint "$chart" --values "$values" \
  --set-string observability.traceSampleRatio=0 >"$work_dir/zero-trace-ratio.log" 2>&1; then
  echo 'trace sampling ratio must reject zero' >&2
  exit 1
fi

if grep -q 'namespaceSelector:' "$work_dir/networkpolicy.yaml"; then
  echo 'network policies must use release pod selectors for same-namespace traffic' >&2
  exit 1
fi

selector_count=$(grep -c 'app.kubernetes.io/instance: cartlabs' "$work_dir/networkpolicy.yaml")
if [ "$selector_count" -lt 8 ]; then
  echo 'dependency policies do not select both target and allowed release pods' >&2
  exit 1
fi

wait_count=$(grep -c 'name: wait-postgresql' "$work_dir/search-jobs.yaml")
if [ "$wait_count" -ne 2 ] || ! grep -q 'pg_isready.*SEARCH_DATABASE_URL' "$work_dir/search-jobs.yaml"; then
  echo 'search hooks must wait for their PostgreSQL database' >&2
  exit 1
fi

if ! grep -A5 'startupProbe:' "$work_dir/observability.yaml" | grep -q 'failureThreshold: 30'; then
  echo 'Tempo startup probe grace missing' >&2
  exit 1
fi

echo 'Helm behavior regression checks passed.'
