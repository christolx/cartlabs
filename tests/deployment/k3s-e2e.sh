#!/bin/sh
set -eu

namespace=${K3S_NAMESPACE:-cartlabs-e2e}
release=cartlabs
chart=deploy/helm
values=$chart/values-local.yaml
k3s_bin=${K3S_BIN:-/usr/local/bin/k3s}
build_images=${K3S_BUILD:-1}
keep_namespace=${K3S_KEEP_NAMESPACE:-0}
created_namespace=0
work_dir=$(mktemp -d)

cleanup() {
  if [ "$created_namespace" -eq 1 ] && [ "$keep_namespace" -ne 1 ]; then
    kubectl delete namespace "$namespace" --wait=false >/dev/null 2>&1 || true
  fi
  rm -rf "$work_dir"
}
trap cleanup EXIT HUP INT TERM

case "$namespace" in
  ''|*[!a-z0-9-]*|-*|*-) echo "invalid K3S_NAMESPACE: $namespace" >&2; exit 1 ;;
esac

for command_name in docker kubectl helm openssl; do
  command -v "$command_name" >/dev/null 2>&1 || {
    echo "missing command: $command_name" >&2
    exit 1
  }
done
[ -x "$k3s_bin" ] || { echo "missing k3s binary: $k3s_bin" >&2; exit 1; }

if kubectl get namespace "$namespace" >/dev/null 2>&1; then
  echo "namespace already exists; refusing destructive reuse: $namespace" >&2
  exit 1
fi

go_apps='api worker mock-payment migrate seed reset search search-migrate search-reindex synthetic-traffic'
images=''
for app in $go_apps; do
  image_name="cartlabs-$app:local"
  if [ "$build_images" -eq 1 ]; then
    docker build --file deploy/docker/go.Dockerfile --build-arg "APP=$app" --tag "$image_name" .
  else
    docker image inspect "$image_name" >/dev/null
  fi
  images="$images $image_name"
done

if [ "$build_images" -eq 1 ]; then
  docker build --file deploy/docker/web.Dockerfile --tag cartlabs-web:local .
else
  docker image inspect cartlabs-web:local >/dev/null
fi
images="$images cartlabs-web:local"

image_archive=$work_dir/cartlabs-images.tar
# shellcheck disable=SC2086
docker image save --output "$image_archive" $images
if "$k3s_bin" ctr images list >/dev/null 2>&1; then
  "$k3s_bin" ctr images import "$image_archive"
else
  command -v pkexec >/dev/null 2>&1 || {
    echo 'k3s containerd needs root; install pkexec or run with permitted K3S_BIN' >&2
    exit 1
  }
  pkexec "$k3s_bin" ctr images import "$image_archive"
fi

kubectl create namespace "$namespace"
created_namespace=1

postgres_password=$(openssl rand -hex 16)
rabbitmq_password=$(openssl rand -hex 16)
kubectl --namespace "$namespace" create secret generic cartlabs-secrets \
  --from-literal=ACCESS_TOKEN_SECRET="$(openssl rand -hex 32)" \
  --from-literal=PAYMENT_WEBHOOK_SECRET="$(openssl rand -hex 32)" \
  --from-literal=MOCK_PAYMENT_API_KEY="$(openssl rand -hex 24)" \
  --from-literal=DATABASE_URL="postgres://cartlabs:$postgres_password@cartlabs-postgresql:5432/cartlabs?sslmode=disable" \
  --from-literal=POSTGRES_PASSWORD="$postgres_password" \
  --from-literal=RABBITMQ_URL="amqp://cartlabs:$rabbitmq_password@cartlabs-rabbitmq:5672/" \
  --from-literal=RABBITMQ_DEFAULT_PASS="$rabbitmq_password" \
  --from-literal=SEARCH_DATABASE_URL="postgres://cartlabs:$postgres_password@cartlabs-postgresql:5432/cartlabs_search?sslmode=disable" \
  --from-literal=SEARCH_SERVICE_TOKEN="$(openssl rand -hex 24)" \
  --from-literal=TRUSTED_PROXY_TOKEN= \
  --from-literal=CLOUDINARY_CLOUD_NAME= \
  --from-literal=NEXT_PUBLIC_CLOUDINARY_CLOUD_NAME= \
  --from-literal=NEXT_PUBLIC_CLOUDINARY_UPLOAD_PRESET= \
  --from-literal=BACKUP_S3_ENDPOINT=http://unused.invalid \
  --from-literal=BACKUP_S3_ACCESS_KEY=unused \
  --from-literal=BACKUP_S3_SECRET_KEY=unused \
  --from-literal=BACKUP_S3_BUCKET=unused \
  --from-literal=GRAFANA_ADMIN_USER=admin \
  --from-literal=GRAFANA_ADMIN_PASSWORD="$(openssl rand -hex 16)"

helm upgrade --install "$release" "$chart" \
  --namespace "$namespace" \
  --values "$values" \
  --set ingress.enabled=false \
  --set observability.enabled=true \
  --set syntheticTraffic.enabled=true \
  --set syntheticTraffic.journeyInterval=30s \
  --rollback-on-failure --wait --timeout 12m

helm test "$release" --namespace "$namespace" --logs --timeout 3m

helm upgrade "$release" "$chart" \
  --namespace "$namespace" \
  --values "$values" \
  --set ingress.enabled=false \
  --set observability.enabled=true \
  --set syntheticTraffic.enabled=true \
  --set syntheticTraffic.journeyInterval=20s \
  --rollback-on-failure --wait --timeout 12m

kubectl --namespace "$namespace" wait \
  --for=condition=available deployment --all --timeout=5m
helm test "$release" --namespace "$namespace" --logs --timeout 3m

journey_seen=0
attempt=0
while [ "$attempt" -lt 45 ]; do
  if kubectl --namespace "$namespace" logs deployment/cartlabs-synthetic-traffic --since=3m 2>/dev/null | \
      grep -Eq '"cycle":"journey"|cycle=journey'; then
    journey_seen=1
    break
  fi
  attempt=$((attempt + 1))
  sleep 2
done
[ "$journey_seen" -eq 1 ] || { echo 'synthetic journey did not complete' >&2; exit 1; }

check_pod=cartlabs-observability-check
kubectl --namespace "$namespace" run "$check_pod" \
  --image=curlimages/curl@sha256:463eaf6072688fe96ac64fa623fe73e1dbe25d8ad6c34404a669ad3ce1f104b6 \
  --restart=Never --command -- sh -ec '
    attempt=0
    while [ "$attempt" -lt 45 ]; do
      prometheus=$(curl --fail --silent --show-error --get \
        --data-urlencode "query=count(up{job=~\"cartlabs-(api|worker|search|synthetic-traffic)\"} == 1)" \
        http://cartlabs-prometheus:9090/api/v1/query) || prometheus=
      tempo=$(curl --fail --silent --show-error --get \
        --data-urlencode "tags=service.name=cartlabs-synthetic-traffic" \
        http://cartlabs-tempo:3200/api/search) || tempo=
      if echo "$prometheus" | grep -Fq "\"4\"" && echo "$tempo" | grep -q traceID; then
        exit 0
      fi
      attempt=$((attempt + 1))
      sleep 2
    done
    echo "Prometheus response: $prometheus" >&2
    echo "Tempo response: $tempo" >&2
    exit 1
  '
if ! kubectl --namespace "$namespace" wait --for=jsonpath='{.status.phase}'=Succeeded "pod/$check_pod" --timeout=2m; then
  kubectl --namespace "$namespace" logs "$check_pod" >&2 || true
  exit 1
fi
kubectl --namespace "$namespace" logs "$check_pod"

echo "k3s install, upgrade, smoke, metrics, traces, and synthetic journey passed: $namespace"
