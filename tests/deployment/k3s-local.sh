#!/bin/sh
set -eu

action=${1:-}
namespace=${K3S_NAMESPACE:-cartlabs}
release=${K3S_RELEASE:-cartlabs}
env_file=${K3S_ENV_FILE:-.env.k3s.local}
chart=deploy/helm
values=$chart/values-k3s.yaml

usage() {
  echo "usage: $0 {up|rebuild|reset|status|logs|down|purge}" >&2
  exit 2
}

acquire_mutation_lock() {
  state_dir=${K3S_STATE_DIR:-${XDG_STATE_HOME:-$HOME/.local/state}/cartlabs}
  mkdir -p "$state_dir"
  chmod 700 "$state_dir"
  lock_file=$state_dir/k3s-$namespace-$release.lock
  exec 9>"$lock_file"
  flock -n 9 || {
    echo "another k3s mutation is running for $namespace/$release" >&2
    exit 1
  }
}

restore_replicas() {
  replicas_file=$1
  restore_failed=0
  while read -r deployment count; do
    [ -n "$deployment" ] || continue
    if ! kubectl -n "$namespace" scale "$deployment" --replicas="$count"; then
      echo "failed to restore $deployment to $count replicas" >&2
      restore_failed=1
    fi
  done < "$replicas_file"
  while read -r deployment count; do
    [ -n "$deployment" ] || continue
    if [ "$count" -gt 0 ] && ! kubectl -n "$namespace" rollout status "$deployment" --timeout=5m; then
      echo "failed waiting for restored $deployment" >&2
      restore_failed=1
    fi
  done < "$replicas_file"
  [ "$restore_failed" -eq 0 ]
}

reset_cluster() {
  replicas_file=$state_dir/k3s-$namespace-$release-reset-replicas
  job_file=$replicas_file-job
  if [ -s "$replicas_file" ]; then
    echo "recovering interrupted reset"
    recovery_result=0
    if [ -s "$job_file" ]; then
      IFS= read -r stale_job < "$job_file"
      if kubectl -n "$namespace" get "job/$stale_job" >/dev/null 2>&1; then
        kubectl -n "$namespace" wait --for=condition=Complete "job/$stale_job" --timeout=6m || recovery_result=$?
        kubectl -n "$namespace" logs "job/$stale_job" --all-containers=true --prefix || true
        delete_result=0
        kubectl -n "$namespace" delete "job/$stale_job" --ignore-not-found=true || delete_result=$?
        if [ "$delete_result" -eq 0 ]; then
          rm -f "$job_file"
        else
          recovery_result=$delete_result
        fi
      else
        rm -f "$job_file"
      fi
    fi
    restore_replicas "$replicas_file" || {
      echo "replica recovery failed; retained $replicas_file" >&2
      exit 1
    }
    rm -f "$replicas_file"
    [ "$recovery_result" -eq 0 ] || {
      echo "interrupted reset Job did not complete cleanly" >&2
      exit "$recovery_result"
    }
    echo "interrupted reset recovered; no new reset started"
    return
  fi

  cronjobs=$(kubectl -n "$namespace" get cronjob \
    -l "app.kubernetes.io/instance=$release,app.kubernetes.io/component=reset" \
    -o name)
  set -- $cronjobs
  [ "$#" -eq 1 ] || {
    echo "expected one reset CronJob for $namespace/$release; found $#" >&2
    exit 1
  }
  reset_cronjob=$1
  reset_cronjob_name=${reset_cronjob#*/}
  [ "$(kubectl -n "$namespace" get "$reset_cronjob" -o jsonpath='{.spec.suspend}')" = true ] || {
    echo "reset CronJob must be suspended: $reset_cronjob" >&2
    exit 1
  }

  replicas_tmp=$replicas_file.tmp.$$
  : > "$replicas_tmp"
  for component in api worker web synthetic-traffic; do
    deployments=$(kubectl -n "$namespace" get deployment \
      -l "app.kubernetes.io/instance=$release,app.kubernetes.io/component=$component" \
      -o name)
    set -- $deployments
    case "$component:$#" in
      api:1|worker:1|web:0|web:1|synthetic-traffic:0|synthetic-traffic:1) ;;
      api:0|worker:0)
        rm -f "$replicas_tmp"
        echo "missing required $component Deployment for $namespace/$release" >&2
        exit 1
        ;;
      *)
        rm -f "$replicas_tmp"
        echo "expected at most one $component Deployment for $namespace/$release; found $#" >&2
        exit 1
        ;;
    esac
    if [ "$#" -eq 1 ]; then
      count=$(kubectl -n "$namespace" get "$1" -o jsonpath='{.spec.replicas}')
      case "$count" in ''|*[!0-9]*) rm -f "$replicas_tmp"; echo "invalid replica count for $1" >&2; exit 1 ;; esac
      printf '%s %s\n' "$1" "$count" >> "$replicas_tmp"
    fi
  done
  mv "$replicas_tmp" "$replicas_file"

  reset_job=
  finish_reset() {
    result=$?
    trap - 0 HUP INT TERM
    if [ -n "$reset_job" ] && ! kubectl -n "$namespace" delete "job/$reset_job" --ignore-not-found=true; then
      [ "$result" -ne 0 ] || result=1
      echo "failed to delete reset Job: $reset_job" >&2
    else
      rm -f "$job_file"
    fi
    if ! restore_replicas "$replicas_file"; then
      [ "$result" -ne 0 ] || result=1
      echo "replica restoration incomplete; retained $replicas_file" >&2
    else
      rm -f "$replicas_file"
    fi
    exit "$result"
  }
  trap finish_reset 0
  trap 'exit 129' HUP
  trap 'exit 130' INT
  trap 'exit 143' TERM

  while read -r deployment count; do
    kubectl -n "$namespace" scale "$deployment" --replicas=0
  done < "$replicas_file"

  pod_selector="app.kubernetes.io/instance=$release,app.kubernetes.io/component in (web,api,worker,synthetic-traffic)"
  kubectl -n "$namespace" wait --for=delete pod -l "$pod_selector" --timeout=3m
  remaining=$(kubectl -n "$namespace" get pod -l "$pod_selector" --no-headers | wc -l | tr -d ' ')
  [ "$remaining" -eq 0 ] || {
    echo "application pods still running after scale-down: $remaining" >&2
    exit 1
  }

  reset_job=cartlabs-reset-$(date +%s)-$$
  printf '%s\n' "$reset_job" > "$job_file"
  kubectl -n "$namespace" create job "$reset_job" --from="cronjob/$reset_cronjob_name"
  job_result=0
  kubectl -n "$namespace" wait --for=condition=Complete "job/$reset_job" --timeout=6m || job_result=$?
  kubectl -n "$namespace" logs "job/$reset_job" --all-containers=true --prefix || true
  [ "$job_result" -eq 0 ] || {
    echo "reset Job failed or timed out: $reset_job" >&2
    exit "$job_result"
  }
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

need_helm_v4() {
  helm_version=$(helm version --template '{{.Version}}')
  case "$helm_version" in
    v4.*) ;;
    *)
      echo "Helm 4 required; found ${helm_version:-unknown}" >&2
      exit 1
      ;;
  esac
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
  K3S_DOMAIN=${K3S_DOMAIN:-cartlabs.christofletjhai.dev}
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
    for command_name in kubectl helm sha256sum flock; do need_command "$command_name"; done
    acquire_mutation_lock
    need_helm_v4
    load_config
    reconcile
    ;;
  rebuild)
    for command_name in kubectl helm sha256sum flock; do need_command "$command_name"; done
    acquire_mutation_lock
    need_helm_v4
    load_config
    reconcile
    kubectl -n "$namespace" rollout restart deployment -l "app.kubernetes.io/instance=$release"
    kubectl -n "$namespace" rollout status deployment -l "app.kubernetes.io/instance=$release" --timeout=5m
    ;;
  reset)
    for command_name in kubectl flock; do need_command "$command_name"; done
    acquire_mutation_lock
    reset_cluster
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
    for command_name in helm flock; do need_command "$command_name"; done
    acquire_mutation_lock
    if helm status "$release" -n "$namespace" >/dev/null 2>&1; then
      helm uninstall "$release" -n "$namespace" --wait
    fi
    echo "release removed; namespace and PVCs retained: $namespace"
    ;;
  purge)
    for command_name in kubectl flock; do need_command "$command_name"; done
    acquire_mutation_lock
    [ "${K3S_PURGE:-0}" = 1 ] || {
      echo 'refusing data deletion; rerun with K3S_PURGE=1 make k3s-purge' >&2
      exit 1
    }
    kubectl delete namespace "$namespace"
    echo "namespace and persistent data deleted: $namespace"
    ;;
  *) usage ;;
esac
