#!/bin/sh
set -eu

work_dir=$(mktemp -d)
trap 'rm -rf "$work_dir"' EXIT HUP INT TERM
mkdir -p "$work_dir/bin" "$work_dir/state"

cat > "$work_dir/bin/kubectl" <<'EOF'
#!/bin/sh
set -eu
[ "$1" = -n ] && shift 2
args=$*
case "$args" in
  "get cronjob "*"-o name")
    [ "${FAKE_MISSING_CRONJOB:-0}" = 1 ] || echo cronjob.batch/cartlabs-reset
    ;;
  "get cronjob.batch/cartlabs-reset "*) echo true ;;
  "get deployment "*"component=api"*) echo deployment.apps/cartlabs-api ;;
  "get deployment "*"component=worker"*) echo deployment.apps/cartlabs-worker ;;
  "get deployment "*"component=web"*) ;;
  "get deployment "*"component=synthetic-traffic"*) echo deployment.apps/cartlabs-synthetic-traffic ;;
  "get deployment.apps/cartlabs-api "*) echo 2 ;;
  "get deployment.apps/cartlabs-worker "*) echo 1 ;;
  "get deployment.apps/cartlabs-synthetic-traffic "*) echo 0 ;;
  "get job/cartlabs-reset-stale") ;;
  "get pod "*) ;;
  "scale "*|"rollout status "*|"create job "*|"logs job/"*|"delete job/"*)
    echo "$args" >> "$FAKE_KUBECTL_LOG"
    ;;
  "wait --for=delete pod "*) ;;
  "wait --for=condition=Complete job/"*)
    echo "$args" >> "$FAKE_KUBECTL_LOG"
    [ "${FAKE_FAIL_JOB:-0}" != 1 ]
    ;;
  *)
    echo "unexpected kubectl call: $args" >&2
    exit 1
    ;;
esac
EOF
chmod +x "$work_dir/bin/kubectl"

run_reset() {
  PATH="$work_dir/bin:$PATH" \
    HOME="$work_dir" \
    K3S_STATE_DIR="$work_dir/state" \
    FAKE_KUBECTL_LOG="$work_dir/kubectl.log" \
    "$@" sh tests/deployment/k3s-local.sh reset
}

: > "$work_dir/kubectl.log"
run_reset
grep -q '^scale deployment.apps/cartlabs-api --replicas=0$' "$work_dir/kubectl.log"
grep -q '^scale deployment.apps/cartlabs-api --replicas=2$' "$work_dir/kubectl.log"
grep -q '^create job cartlabs-reset-.* --from=cronjob/cartlabs-reset$' "$work_dir/kubectl.log"
grep -q '^logs job/cartlabs-reset-.* --all-containers=true --prefix$' "$work_dir/kubectl.log"
grep -q '^delete job/cartlabs-reset-.* --ignore-not-found=true$' "$work_dir/kubectl.log"
[ ! -e "$work_dir/state/k3s-cartlabs-cartlabs-reset-replicas" ]

: > "$work_dir/kubectl.log"
if run_reset env FAKE_FAIL_JOB=1 >"$work_dir/job-failure.log" 2>&1; then
  echo 'failed reset Job must fail command' >&2
  exit 1
fi
grep -q '^logs job/cartlabs-reset-' "$work_dir/kubectl.log"
grep -q '^scale deployment.apps/cartlabs-worker --replicas=1$' "$work_dir/kubectl.log"
[ ! -e "$work_dir/state/k3s-cartlabs-cartlabs-reset-replicas" ]

printf '%s %s\n' deployment.apps/cartlabs-api 2 > "$work_dir/state/k3s-cartlabs-cartlabs-reset-replicas"
printf '%s\n' cartlabs-reset-stale > "$work_dir/state/k3s-cartlabs-cartlabs-reset-replicas-job"
: > "$work_dir/kubectl.log"
run_reset
grep -q '^wait --for=condition=Complete job/cartlabs-reset-stale --timeout=6m$' "$work_dir/kubectl.log"
grep -q '^scale deployment.apps/cartlabs-api --replicas=2$' "$work_dir/kubectl.log"
if grep -q '^create job ' "$work_dir/kubectl.log"; then
  echo 'recovery must not start another reset' >&2
  exit 1
fi
[ ! -e "$work_dir/state/k3s-cartlabs-cartlabs-reset-replicas" ]
[ ! -e "$work_dir/state/k3s-cartlabs-cartlabs-reset-replicas-job" ]

if run_reset env FAKE_MISSING_CRONJOB=1 >"$work_dir/missing-cronjob.log" 2>&1; then
  echo 'missing reset CronJob must fail command' >&2
  exit 1
fi

exec 8>"$work_dir/state/k3s-cartlabs-cartlabs.lock"
flock -n 8
if run_reset >"$work_dir/lock-contention.log" 2>&1; then
  echo 'concurrent reset must fail command' >&2
  exit 1
fi
flock -u 8
exec 8>&-

echo 'k3s reset regression checks passed.'
