#!/bin/sh
set -eu

action=${1:-}
namespace=${K3S_NAMESPACE:-cartlabs}
release=${K3S_RELEASE:-cartlabs}
env_file=${K3S_ENV_FILE:-.env.k3s.local}
chart=deploy/helm
values=$chart/values-k3s.yaml

usage() {
  echo "usage: $0 {up|rebuild|status|logs|down|purge}" >&2
  exit 2
}

valid_name() {
  case "$1" in
    ''|*[!a-z0-9-]*|-*|*-) return 1 ;;
  esac
}

need_command() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "missing command: $1" >&2
    exit 1
  }
}

need_value() {
  eval "value=\${$1-}"
  [ -n "$value" ] || {
    echo "missing required value in $env_file: $1" >&2
    exit 1
  }
}

load_config() {
  [ -f "$env_file" ] || {
    echo "missing $env_file; copy .env.k3s.local.example and fill it" >&2
    exit 1
  }
  requested_deployment_mode=${K3S_DEPLOYMENT_MODE-}
  set -a
  # shellcheck disable=SC1090
  . "$env_file"
  set +a

  K3S_DEPLOYMENT_MODE=${requested_deployment_mode:-${K3S_DEPLOYMENT_MODE:-full}}
  case "$K3S_DEPLOYMENT_MODE" in
    full) web_enabled=true; ingress_target=web ;;
    backend) web_enabled=false; ingress_target=api ;;
    *) echo 'K3S_DEPLOYMENT_MODE must be full or backend' >&2; exit 1 ;;
  esac

  required='K3S_IMAGE_TAG GHCR_USERNAME GHCR_PULL_TOKEN ACCESS_TOKEN_SECRET PAYMENT_WEBHOOK_SECRET MOCK_PAYMENT_API_KEY POSTGRES_PASSWORD RABBITMQ_DEFAULT_PASS SEARCH_SERVICE_TOKEN TRUSTED_PROXY_TOKEN CLOUDINARY_CLOUD_NAME CLOUDINARY_API_KEY CLOUDINARY_API_SECRET GRAFANA_ADMIN_USER GRAFANA_ADMIN_PASSWORD'
  if [ "$K3S_DEPLOYMENT_MODE" = full ]; then
    required="$required NEXT_PUBLIC_CLOUDINARY_CLOUD_NAME NEXT_PUBLIC_CLOUDINARY_UPLOAD_PRESET"
  fi
  for key in $required; do need_value "$key"; done
  case "$K3S_IMAGE_TAG" in
    sha-*[!0-9a-f]*|sha-) valid_tag=0 ;;
    sha-*) valid_tag=1 ;;
    *) valid_tag=0 ;;
  esac
  [ "$valid_tag" -eq 1 ] && [ "${#K3S_IMAGE_TAG}" -eq 44 ] || {
    echo 'K3S_IMAGE_TAG must be sha- followed by full 40-character lowercase commit' >&2
    exit 1
  }
  [ "${#ACCESS_TOKEN_SECRET}" -ge 32 ] || { echo 'ACCESS_TOKEN_SECRET must contain at least 32 bytes' >&2; exit 1; }
  [ "${#PAYMENT_WEBHOOK_SECRET}" -ge 32 ] || { echo 'PAYMENT_WEBHOOK_SECRET must contain at least 32 bytes' >&2; exit 1; }
  [ "${#MOCK_PAYMENT_API_KEY}" -ge 24 ] || { echo 'MOCK_PAYMENT_API_KEY must contain at least 24 bytes' >&2; exit 1; }
  [ "${#SEARCH_SERVICE_TOKEN}" -ge 32 ] || { echo 'SEARCH_SERVICE_TOKEN must contain at least 32 bytes' >&2; exit 1; }
  [ "${#TRUSTED_PROXY_TOKEN}" -ge 32 ] || { echo 'TRUSTED_PROXY_TOKEN must contain at least 32 bytes' >&2; exit 1; }
  case "$POSTGRES_PASSWORD$RABBITMQ_DEFAULT_PASS" in
    *[!A-Za-z0-9._~-]*) echo 'database and RabbitMQ passwords must use URL-safe characters' >&2; exit 1 ;;
  esac
  NEXT_PUBLIC_CLOUDINARY_CLOUD_NAME=${NEXT_PUBLIC_CLOUDINARY_CLOUD_NAME:-}
  NEXT_PUBLIC_CLOUDINARY_UPLOAD_PRESET=${NEXT_PUBLIC_CLOUDINARY_UPLOAD_PRESET:-}
  [ -z "$NEXT_PUBLIC_CLOUDINARY_CLOUD_NAME" ] || [ "$CLOUDINARY_CLOUD_NAME" = "$NEXT_PUBLIC_CLOUDINARY_CLOUD_NAME" ] || {
    echo 'CLOUDINARY_CLOUD_NAME and NEXT_PUBLIC_CLOUDINARY_CLOUD_NAME must match' >&2
    exit 1
  }

  K3S_IMAGE_REGISTRY=${K3S_IMAGE_REGISTRY:-ghcr.io/christolx}
  case "$K3S_IMAGE_REGISTRY" in ghcr.io/*) ;; *) echo 'K3S_IMAGE_REGISTRY must use ghcr.io' >&2; exit 1 ;; esac
  K3S_DOMAIN=${K3S_DOMAIN:-cartlabs.christofle.dev}
  K3S_SYNTHETIC_TRAFFIC=${K3S_SYNTHETIC_TRAFFIC:-false}
  case "$K3S_SYNTHETIC_TRAFFIC" in true|false) ;; *) echo 'K3S_SYNTHETIC_TRAFFIC must be true or false' >&2; exit 1 ;; esac
}

reconcile() {
  database_url="postgres://cartlabs:$POSTGRES_PASSWORD@$release-postgresql:5432/cartlabs?sslmode=disable"
  search_database_url="postgres://cartlabs:$POSTGRES_PASSWORD@$release-postgresql:5432/cartlabs_search?sslmode=disable"
  rabbitmq_url="amqp://cartlabs:$RABBITMQ_DEFAULT_PASS@$release-rabbitmq:5672/"

  kubectl create namespace "$namespace" --dry-run=client -o yaml | kubectl apply -f - >/dev/null
  kubectl -n "$namespace" create secret docker-registry ghcr-pull \
    --docker-server=ghcr.io --docker-username="$GHCR_USERNAME" --docker-password="$GHCR_PULL_TOKEN" \
    --dry-run=client -o yaml | kubectl apply -f - >/dev/null
  kubectl -n "$namespace" create secret generic cartlabs-secrets \
    --from-literal=ACCESS_TOKEN_SECRET="$ACCESS_TOKEN_SECRET" \
    --from-literal=PAYMENT_WEBHOOK_SECRET="$PAYMENT_WEBHOOK_SECRET" \
    --from-literal=MOCK_PAYMENT_API_KEY="$MOCK_PAYMENT_API_KEY" \
    --from-literal=DATABASE_URL="$database_url" \
    --from-literal=POSTGRES_PASSWORD="$POSTGRES_PASSWORD" \
    --from-literal=RABBITMQ_URL="$rabbitmq_url" \
    --from-literal=RABBITMQ_DEFAULT_PASS="$RABBITMQ_DEFAULT_PASS" \
    --from-literal=SEARCH_DATABASE_URL="$search_database_url" \
    --from-literal=SEARCH_SERVICE_TOKEN="$SEARCH_SERVICE_TOKEN" \
    --from-literal=TRUSTED_PROXY_TOKEN="$TRUSTED_PROXY_TOKEN" \
    --from-literal=CLOUDINARY_CLOUD_NAME="$CLOUDINARY_CLOUD_NAME" \
    --from-literal=NEXT_PUBLIC_CLOUDINARY_CLOUD_NAME="$NEXT_PUBLIC_CLOUDINARY_CLOUD_NAME" \
    --from-literal=NEXT_PUBLIC_CLOUDINARY_UPLOAD_PRESET="$NEXT_PUBLIC_CLOUDINARY_UPLOAD_PRESET" \
    --from-literal=BACKUP_S3_ENDPOINT= \
    --from-literal=BACKUP_S3_ACCESS_KEY= \
    --from-literal=BACKUP_S3_SECRET_KEY= \
    --from-literal=BACKUP_S3_BUCKET= \
    --from-literal=GRAFANA_ADMIN_USER="$GRAFANA_ADMIN_USER" \
    --from-literal=GRAFANA_ADMIN_PASSWORD="$GRAFANA_ADMIN_PASSWORD" \
    --dry-run=client -o yaml | kubectl apply -f - >/dev/null
  kubectl -n "$namespace" create secret generic cartlabs-cloudinary-cleanup \
    --from-literal=CLOUDINARY_CLOUD_NAME="$CLOUDINARY_CLOUD_NAME" \
    --from-literal=CLOUDINARY_API_KEY="$CLOUDINARY_API_KEY" \
    --from-literal=CLOUDINARY_API_SECRET="$CLOUDINARY_API_SECRET" \
    --dry-run=client -o yaml | kubectl apply -f - >/dev/null

  secret_revision=$(sha256sum "$env_file" | awk '{print $1}')
  helm upgrade --install "$release" "$chart" \
    --namespace "$namespace" --values "$values" \
    --set-string global.imageRegistry="$K3S_IMAGE_REGISTRY" \
    --set-string global.imageTag="$K3S_IMAGE_TAG" \
    --set-string global.domain="$K3S_DOMAIN" \
    --set-string global.secretRevision="$secret_revision" \
    --set web.enabled="$web_enabled" \
    --set-string ingress.target="$ingress_target" \
    --set syntheticTraffic.enabled="$K3S_SYNTHETIC_TRAFFIC" \
    --rollback-on-failure --wait --timeout 12m
  helm test "$release" --namespace "$namespace" --logs --timeout 3m
}

valid_name "$namespace" || { echo "invalid K3S_NAMESPACE: $namespace" >&2; exit 1; }
valid_name "$release" || { echo "invalid K3S_RELEASE: $release" >&2; exit 1; }

case "$action" in
  up)
    for command_name in kubectl helm sha256sum; do need_command "$command_name"; done
    load_config
    reconcile
    ;;
  rebuild)
    for command_name in kubectl helm sha256sum; do need_command "$command_name"; done
    load_config
    reconcile
    kubectl -n "$namespace" rollout restart deployment -l "app.kubernetes.io/instance=$release"
    kubectl -n "$namespace" rollout status deployment -l "app.kubernetes.io/instance=$release" --timeout=5m
    ;;
  status)
    for command_name in kubectl helm; do need_command "$command_name"; done
    helm status "$release" -n "$namespace"
    kubectl -n "$namespace" get pods,jobs,ingress,pvc
    ;;
  logs)
    need_command kubectl
    if [ -n "${K3S_LOG_COMPONENT:-}" ]; then
      selector="app.kubernetes.io/instance=$release,app.kubernetes.io/component=$K3S_LOG_COMPONENT"
    else
      selector="app.kubernetes.io/instance=$release"
    fi
    kubectl -n "$namespace" logs -l "$selector" --all-containers=true --prefix --tail=200 -f --max-log-requests=30
    ;;
  down)
    need_command helm
    if helm status "$release" -n "$namespace" >/dev/null 2>&1; then
      helm uninstall "$release" -n "$namespace" --wait
    fi
    echo "release removed; namespace and PVCs retained: $namespace"
    ;;
  purge)
    need_command kubectl
    [ "${K3S_PURGE:-0}" = 1 ] || {
      echo 'refusing data deletion; rerun with K3S_PURGE=1 make k3s-purge' >&2
      exit 1
    }
    kubectl delete namespace "$namespace"
    echo "namespace and persistent data deleted: $namespace"
    ;;
  *) usage ;;
esac
