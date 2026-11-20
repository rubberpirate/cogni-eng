#!/usr/bin/env bash
set -euo pipefail
umask 077
root_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$root_dir"

case "${1:-}" in
  init)
    origin=${2:-}
    [[ $origin =~ ^https://[a-zA-Z0-9.-]+(:[0-9]+)?$ ]] || {
      echo 'Usage: bash scripts/server-infra.sh init https://YOUR_FRONTEND_DOMAIN' >&2; exit 1;
    }
    [[ ! -e .env.production && ! -e .local/server/tls ]] || {
      echo 'Existing .env.production or TLS files found. Retain them; see docs/SERVER-DATABASE.md before changing credentials.' >&2; exit 1;
    }
    mkdir -p .local/server
    staging=$(mktemp -d "$root_dir/.local/server/infra.XXXXXXXX")
    trap 'rm -rf -- "$staging"' EXIT
    openssl req -x509 -newkey rsa:3072 -nodes -days 3650 \
      -subj '/CN=Learning Server Local CA' -addext 'basicConstraints=critical,CA:TRUE' \
      -keyout "$staging/ca.key" -out "$staging/ca.crt" 2>/dev/null
    openssl req -newkey rsa:3072 -nodes -subj '/CN=localhost' \
      -keyout "$staging/server.key" -out "$staging/server.csr" 2>/dev/null
    cat > "$staging/extensions" <<'EXT'
subjectAltName=IP:127.0.0.1,DNS:localhost
basicConstraints=critical,CA:FALSE
keyUsage=critical,digitalSignature,keyEncipherment
extendedKeyUsage=serverAuth
EXT
    openssl x509 -req -in "$staging/server.csr" -CA "$staging/ca.crt" \
      -CAkey "$staging/ca.key" -CAcreateserial -days 825 \
      -extfile "$staging/extensions" -out "$staging/server.crt" 2>/dev/null
    # Retain public roots for SMTP/model providers alongside the private local CA.
    cat /etc/ssl/certs/ca-certificates.crt "$staging/ca.crt" > "$staging/trust.pem"
    database_password=$(openssl rand -hex 24)
    app_password=$(openssl rand -hex 24)
    tls_dir="$root_dir/.local/server/tls"
    cat > "$staging/env" <<ENV
APP_ENV=production
HTTP_ADDR=127.0.0.1:8080
PUBLIC_ORIGIN=$origin
DATABASE_PASSWORD=$database_password
APP_DATABASE_PASSWORD=$app_password
DATABASE_URL='postgres://workspace:$app_password@127.0.0.1:55432/workspace?sslmode=verify-full'
MIGRATION_DATABASE_URL='postgres://workspace_owner:$database_password@127.0.0.1:55432/workspace?sslmode=verify-full'
SERVER_TLS_DIR='$tls_dir'
PGSSLROOTCERT='$tls_dir/ca.crt'
SSL_CERT_FILE='$tls_dir/trust.pem'
MFA_KEY=$(openssl rand -hex 32)
OBJECT_ENDPOINT=https://127.0.0.1:59443
OBJECT_BUCKET=private-workspace
OBJECT_USER=workspace-server
OBJECT_PASSWORD=$(openssl rand -hex 32)
SMTP_ADDR=smtp.example.com:587
SMTP_FROM=security@example.com
SMTP_USERNAME=''
SMTP_PASSWORD=''
ACADEMIC_MODEL_PROVIDER=gemini
GEMINI_API_KEY=''
GEMINI_MODEL=gemini-flash-latest
OCR_PROVIDER=tesseract
OMR_MODEL_PROVIDER=bubble_classifier
OMRCHECKER_ROOT=''
OMRCHECKER_PYTHON=''
OMRCHECKER_ADAPTER=''
ENV
    # Avoid creating shell syntax from an unusual checkout path.
    [[ $root_dir != *"'"* && $root_dir != *$'\n'* ]] || { echo 'Use a checkout path without single quotes/newlines.' >&2; exit 1; }
    (set -o noclobber; cat "$staging/env" > .env.production)
    mv "$staging" "$tls_dir"
    trap - EXIT
    # The environment belongs in the repo root, not in the certificate directory.
    rm -- "$tls_dir/env"
    echo 'Created .env.production and .local/server/tls. Edit SMTP settings before starting the worker.'
    ;;
  up|status|stop|down)
    [[ -f .env.production ]] || { echo 'Run init first.' >&2; exit 1; }
    case $1 in
      up) docker compose --env-file .env.production -f compose.server.yaml up -d --wait ;;
      status) docker compose --env-file .env.production -f compose.server.yaml ps ;;
      stop) docker compose --env-file .env.production -f compose.server.yaml stop ;;
      down) docker compose --env-file .env.production -f compose.server.yaml down ;;
    esac
    ;;
  logs)
    [[ -f .env.production ]] || { echo 'Run init first.' >&2; exit 1; }
    docker compose --env-file .env.production -f compose.server.yaml logs --tail 50 "${2:-}"
    ;;
  *) echo 'Usage: bash scripts/server-infra.sh init https://FRONTEND_DOMAIN | up | status | stop | down | logs' >&2; exit 1 ;;
esac
