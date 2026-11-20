#!/usr/bin/env bash
set -euo pipefail

root_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$root_dir"

usage() {
  echo 'Usage: bash scripts/server.sh build|migrate|bootstrap|start|restart|tunnel|status'
}

load_env() {
  if [[ ! -f .env.production ]]; then
    echo 'Create .env.production from deploy/backend.env.example first.' >&2
    exit 1
  fi
  set -a
  # This is a private, operator-owned Bash file; never source an untrusted file.
  source .env.production
  set +a
  [[ ${APP_ENV:-} == production ]] || { echo 'APP_ENV must be production.' >&2; exit 1; }
  [[ ${HTTP_ADDR:-} == 127.0.0.1:8080 ]] || { echo 'HTTP_ADDR must be 127.0.0.1:8080 for this tunnel setup.' >&2; exit 1; }
  [[ ${PUBLIC_ORIGIN:-} =~ ^https://[^/?#[:space:]]+$ ]] || { echo 'PUBLIC_ORIGIN must be the frontend HTTPS origin without a trailing slash.' >&2; exit 1; }
  [[ ${MFA_KEY:-} =~ ^[[:xdigit:]]{64}$ ]] || { echo 'MFA_KEY must contain 32 random bytes encoded as hex.' >&2; exit 1; }
  : "${DATABASE_URL:?Set DATABASE_URL}" "${OBJECT_BUCKET:?Set OBJECT_BUCKET}"
  : "${OBJECT_USER:?Set OBJECT_USER}" "${OBJECT_PASSWORD:?Set OBJECT_PASSWORD}"
  : "${SMTP_ADDR:?Set SMTP_ADDR}" "${SMTP_FROM:?Set SMTP_FROM}"
  [[ ${OBJECT_ENDPOINT:-} =~ ^https://[^/?#[:space:]]+$ ]] || { echo 'OBJECT_ENDPOINT must be an HTTPS origin without a trailing slash.' >&2; exit 1; }
}

case "${1:-}" in
  build)
    mkdir -p .local/server/releases
    release_dir="$(mktemp -d "$root_dir/.local/server/releases/build.XXXXXXXX")"
    trap 'rm -rf -- "$release_dir"' EXIT
    for binary in gateway worker operator; do
      CGO_ENABLED=0 go build -trimpath -o "$release_dir/$binary" "./cmd/$binary"
    done
    # Switch only after every binary builds; running processes keep their old executable.
    ln -s "${release_dir#"$root_dir/.local/server/"}" "$release_dir.link"
    mv -Tf "$release_dir.link" .local/server/current
    trap - EXIT
    echo "Built backend: $release_dir"
    ;;
  migrate|bootstrap)
    load_env
    if [[ $1 == migrate && -n ${MIGRATION_DATABASE_URL:-} ]]; then
      export DATABASE_URL="$MIGRATION_DATABASE_URL"
    fi
    exec .local/server/current/operator "$1"
    ;;
  start|restart)
    # Validate in a subshell so application secrets do not enter PM2's environment dump.
    (load_env)
    for binary in gateway worker; do
      [[ -x .local/server/current/$binary ]] || { echo 'Run build first.' >&2; exit 1; }
    done
    pm2 startOrRestart deploy/ecosystem.config.cjs --only learning-api,learning-worker
    ;;
  run)
    case "${2:-}" in
      gateway|worker) ;;
      *) usage >&2; exit 1 ;;
    esac
    load_env
    unset MIGRATION_DATABASE_URL BOOTSTRAP_PASSWORD
    exec ".local/server/current/$2"
    ;;
  tunnel)
    command -v cloudflared >/dev/null
    [[ -f "$HOME/.cloudflared/config.yml" ]] || { echo 'Create ~/.cloudflared/config.yml using docs/SELF-HOSTING.md.' >&2; exit 1; }
    pm2 startOrRestart deploy/ecosystem.config.cjs --only learning-tunnel
    ;;
  run-tunnel)
    exec cloudflared tunnel --config "$HOME/.cloudflared/config.yml" run
    ;;
  status)
    pm2 status
    ;;
  -h|--help) usage ;;
  *) usage >&2; exit 1 ;;
esac
