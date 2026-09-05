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
helm template cartlabs "$chart" --namespace cartlabs --values deploy/helm/values-k3s.yaml \
  --show-only templates/deployments.yaml >"$work_dir/k3s-deployments.yaml"
helm template cartlabs "$chart" --namespace cartlabs --values deploy/helm/values-k3s.yaml \
  --show-only templates/services.yaml >"$work_dir/full-services.yaml"
helm template cartlabs "$chart" --namespace cartlabs --values deploy/helm/values-k3s.yaml \
  --show-only templates/ingress.yaml >"$work_dir/full-ingress.yaml"
helm template cartlabs "$chart" --namespace cartlabs --values deploy/helm/values-k3s.yaml \
  --show-only templates/tests/smoke-test.yaml >"$work_dir/full-smoke.yaml"
helm template cartlabs "$chart" --namespace cartlabs --values deploy/helm/values-k3s.yaml \
  --set web.enabled=false --set-string ingress.target=api \
  --show-only templates/deployments.yaml >"$work_dir/backend-deployments.yaml"
helm template cartlabs "$chart" --namespace cartlabs --values deploy/helm/values-k3s.yaml \
  --set web.enabled=false --set-string ingress.target=api \
  --show-only templates/services.yaml >"$work_dir/backend-services.yaml"
helm template cartlabs "$chart" --namespace cartlabs --values deploy/helm/values-k3s.yaml \
  --set web.enabled=false --set-string ingress.target=api \
  --show-only templates/ingress.yaml >"$work_dir/backend-ingress.yaml"
helm template cartlabs "$chart" --namespace cartlabs --values deploy/helm/values-k3s.yaml \
  --set web.enabled=false --set-string ingress.target=api \
  --show-only templates/tests/smoke-test.yaml >"$work_dir/backend-smoke.yaml"

if helm lint "$chart" --values "$values" \
  --set-string observability.traceSampleRatio=0 >"$work_dir/zero-trace-ratio.log" 2>&1; then
  echo 'trace sampling ratio must reject zero' >&2
  exit 1
fi

if helm lint "$chart" --values deploy/helm/values-k3s.yaml \
  --set web.enabled=false >"$work_dir/invalid-topology.log" 2>&1; then
  echo 'web-disabled topology must require API ingress target' >&2
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

if ! grep -q 'name: COOKIE_SECURE' "$work_dir/k3s-deployments.yaml" || \
    ! grep -A1 'name: COOKIE_SECURE' "$work_dir/k3s-deployments.yaml" | grep -q 'value: "true"'; then
  echo 'external HTTPS must enable secure cookies independently of origin TLS' >&2
  exit 1
fi

if ! grep -A1 'name: DEMO_MODE' "$work_dir/k3s-deployments.yaml" | grep -q 'value: "true"'; then
  echo 'web deployment must receive demo mode' >&2
  exit 1
fi

for full_manifest in "$work_dir/k3s-deployments.yaml" "$work_dir/full-services.yaml"; do
  if ! grep -q 'name: cartlabs-web' "$full_manifest"; then
    echo 'full mode must include web Deployment and Service' >&2
    exit 1
  fi
done
if ! grep -A1 'name: cartlabs-web' "$work_dir/full-ingress.yaml" | grep -q 'number: 3000'; then
  echo 'full mode ingress must route to web port 3000' >&2
  exit 1
fi
if ! grep -q 'cartlabs-web:3000' "$work_dir/full-smoke.yaml"; then
  echo 'full mode smoke test must check web' >&2
  exit 1
fi

for backend_manifest in "$work_dir/backend-deployments.yaml" "$work_dir/backend-services.yaml"; do
  if grep -q 'name: cartlabs-web' "$backend_manifest"; then
    echo 'backend mode must omit web Deployment and Service' >&2
    exit 1
  fi
done
for component in api worker search mock-payment; do
  if ! grep -q "name: cartlabs-$component" "$work_dir/backend-deployments.yaml"; then
    echo "backend mode must retain $component Deployment" >&2
    exit 1
  fi
done
if ! grep -A1 'name: cartlabs-api' "$work_dir/backend-ingress.yaml" | grep -q 'number: 8080'; then
  echo 'backend mode ingress must route to API port 8080' >&2
  exit 1
fi
if grep -q 'cartlabs-web:3000' "$work_dir/backend-smoke.yaml"; then
  echo 'backend mode smoke test must omit web check' >&2
  exit 1
fi
if ! grep -q 'cartlabs-api:8080/api/v1/health/ready' "$work_dir/backend-smoke.yaml"; then
  echo 'backend mode smoke test must retain API check' >&2
  exit 1
fi

echo 'Helm behavior regression checks passed.'
