#!/usr/bin/env bash
set -euo pipefail

# Ensure ~/.local/bin is in PATH for PM2 and other user-installed CLI tools
if [[ -d "$HOME/.local/bin" ]]; then
  export PATH="$PATH:$HOME/.local/bin"
fi

root_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$root_dir"

usage() {
  cat <<'USAGE'
Usage: bash scripts/server.sh <command> [options]

Core Operations:
  check                       Audit prerequisites, tools, ports, TLS, and services
  init <origin>               Initialize .env.production and TLS certificates for frontend origin
  restart-scratch [origin]    Restart full backend from scratch (stops conflicts, runs infra, OMR, build, migrate, start)
  status                      Show status of PM2 processes, backing containers, and health check
  logs [service|all]          View PM2 logs (default: learning-api; or learning-worker, learning-tunnel, all)

Component Management:
  infra <up|status|stop|down|logs> Manage Docker backing services (PostgreSQL & SeaweedFS)
  omr                         Set up frozen Python OMR virtual environment with uv
  build                       Build Go binaries (gateway, worker, operator) atomically
  migrate                     Run database migrations using operator
  bootstrap                   Interactively bootstrap initial institution owner account
  start                       Start API and worker in PM2
  stop                        Stop API and worker in PM2
  restart                     Restart API and worker in PM2
  tunnel                      Start Cloudflare Tunnel in PM2
  run <gateway|worker>        Run binary in foreground (invoked by PM2)
  run-tunnel                  Run cloudflared tunnel in foreground (invoked by PM2)
USAGE
}

load_env() {
  if [[ ! -f .env.production ]]; then
    echo 'Create .env.production first (run: bash scripts/server.sh init https://YOUR_FRONTEND_DOMAIN or see deploy/backend.env.example).' >&2
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

release_dir=""

build_binaries() {
  mkdir -p .local/server/releases
  release_dir="$(mktemp -d "$root_dir/.local/server/releases/build.XXXXXXXX")"
  trap '[[ -n "${release_dir:-}" ]] && rm -rf -- "$release_dir"' EXIT
  for binary in gateway worker operator; do
    CGO_ENABLED=0 go build -trimpath -o "$release_dir/$binary" "./cmd/$binary"
  done
  # Switch only after every binary builds; running processes keep their old executable.
  ln -s "${release_dir#"$root_dir/.local/server/"}" "$release_dir.link"
  mv -Tf "$release_dir.link" .local/server/current
  trap - EXIT
  echo "Built backend: $release_dir"
}

check_tool() {
  local cmd="$1"
  local name="$2"
  local install_cmd="$3"
  if command -v "$cmd" >/dev/null 2>&1; then
    echo "  [✓] $name ($cmd): installed"
  else
    echo "  [✗] $name ($cmd): NOT FOUND"
    echo "      Install with: $install_cmd"
  fi
}

check_port() {
  local port="$1"
  local desc="$2"
  local in_use=0
  if (exec 3<>/dev/tcp/127.0.0.1/"$port") 2>/dev/null; then
    exec 3>&-
    in_use=1
  fi
  if (( in_use == 1 )); then
    echo "  [*] Port $port ($desc): in use"
  else
    echo "  [ ] Port $port ($desc): free"
  fi
}

cmd_check() {
  echo "=========================================================="
  echo "   Backend Preflight & Environment Diagnostic Check       "
  echo "=========================================================="

  echo
  echo "--- 1. Tools & CLI Utilities ---"
  check_tool "go" "Go compiler" "sudo pacman -S go"
  if command -v go >/dev/null 2>&1; then
    local go_ver
    go_ver=$(go version | awk '{print $3}')
    echo "       Version: $go_ver"
  fi

  check_tool "docker" "Docker container engine" "sudo pacman -S docker && sudo systemctl enable --now docker"
  if command -v docker >/dev/null 2>&1; then
    if docker info >/dev/null 2>&1; then
      echo "       Daemon: running"
    else
      echo "       Daemon: NOT running (run: sudo systemctl start docker)"
    fi
  fi

  check_tool "node" "Node.js" "sudo pacman -S nodejs"
  check_tool "pm2" "PM2 process manager" "npm install --global --prefix \"\$HOME/.local\" pm2 && export PATH=\"\$HOME/.local/bin:\$PATH\""
  check_tool "cloudflared" "Cloudflare Tunnel daemon" "sudo pacman -S cloudflared"
  check_tool "uv" "Python package installer (for OMR)" "sudo pacman -S uv"
  check_tool "tesseract" "Tesseract OCR engine" "sudo pacman -S tesseract tesseract-data-eng"
  check_tool "openssl" "OpenSSL cryptography" "sudo pacman -S openssl"

  echo
  echo "--- 2. Port Usage & Conflicts ---"
  check_port 8080 "Backend Go API (127.0.0.1:8080)"
  check_port 55432 "PostgreSQL (127.0.0.1:55432)"
  if command -v docker >/dev/null 2>&1; then
    if docker ps --format '{{.Names}}' 2>/dev/null | grep -q '^cogni-fresh-database-1$'; then
      echo "  [!] WARNING: Development container 'cogni-fresh-database-1' is running and occupies port 55432."
      echo "      Run 'docker compose -f compose.yaml stop' to avoid collision with production database."
    fi
  fi
  check_port 59443 "SeaweedFS S3 Storage (127.0.0.1:59443)"

  echo
  echo "--- 3. Configuration & TLS Certificates ---"
  if [[ -f .env.production ]]; then
    echo "  [✓] .env.production found"
    local mode
    mode=$(stat -c "%a" .env.production 2>/dev/null || stat -f "%Lp" .env.production 2>/dev/null || echo "unknown")
    echo "      Permissions: $mode (recommended: 600)"
    (
      set +e
      source .env.production 2>/dev/null
      echo "      APP_ENV: ${APP_ENV:-<unset>}"
      echo "      HTTP_ADDR: ${HTTP_ADDR:-<unset>}"
      echo "      PUBLIC_ORIGIN: ${PUBLIC_ORIGIN:-<unset>}"
      if [[ -n "${MFA_KEY:-}" ]]; then echo "      MFA_KEY: [Configured - 64 hex chars]"; else echo "      MFA_KEY: [MISSING]"; fi
      if [[ -n "${DATABASE_URL:-}" ]]; then echo "      DATABASE_URL: [Configured]"; else echo "      DATABASE_URL: [MISSING]"; fi
      if [[ -n "${OBJECT_ENDPOINT:-}" ]]; then echo "      OBJECT_ENDPOINT: ${OBJECT_ENDPOINT}"; else echo "      OBJECT_ENDPOINT: [MISSING]"; fi
      if [[ "${SMTP_ADDR:-}" =~ example\.com ]]; then
        echo "      SMTP_ADDR: ${SMTP_ADDR} (Warning: placeholder address - configure real SMTP provider)"
      else
        echo "      SMTP_ADDR: ${SMTP_ADDR:-<unset>}"
      fi
    )
  else
    echo "  [✗] .env.production NOT found (run: bash scripts/server.sh init https://YOUR_FRONTEND_DOMAIN)"
  fi

  if [[ -d .local/server/tls ]]; then
    echo "  [✓] .local/server/tls directory found"
    for f in ca.crt server.crt server.key trust.pem; do
      if [[ -f ".local/server/tls/$f" ]]; then
        echo "      - $f: present"
      else
        echo "      - $f: [MISSING]"
      fi
    done
    if [[ -f .local/server/tls/server.crt ]] && command -v openssl >/dev/null 2>&1; then
      local expiry
      expiry=$(openssl x509 -in .local/server/tls/server.crt -noout -enddate 2>/dev/null | cut -d= -f2)
      echo "      Certificate expires: $expiry"
    fi
  else
    echo "  [✗] .local/server/tls NOT found"
  fi

  if [[ -f "$HOME/.cloudflared/config.yml" ]]; then
    echo "  [✓] Cloudflared config found: ~/.cloudflared/config.yml"
  else
    echo "  [-] Cloudflared config not yet created at ~/.cloudflared/config.yml"
  fi

  echo
  echo "--- 4. Backend Services & Health ---"
  if command -v docker >/dev/null 2>&1 && [[ -f .env.production ]]; then
    echo "  Docker backing containers (compose.server.yaml):"
    docker compose --env-file .env.production -f compose.server.yaml ps 2>/dev/null || echo "      (not running)"
  fi

  if command -v pm2 >/dev/null 2>&1; then
    echo "  PM2 processes:"
    pm2 jlist 2>/dev/null | grep -o '"name":"[^"]*"' | tr '\n' ' ' || echo "      (none)"
    echo
  fi

  local ready_status="000"
  if (exec 3<>/dev/tcp/127.0.0.1/8080) 2>/dev/null; then
    exec 3>&-
    ready_status=$(curl -s -o /dev/null -w "%{http_code}" http://127.0.0.1:8080/health/ready 2>/dev/null || echo "000")
    if [[ "$ready_status" == "200" ]]; then
      echo "  [✓] Backend readiness check: OK (HTTP 200)"
    else
      echo "  [!] Backend readiness check: HTTP $ready_status (expected 200)"
    fi
  else
    echo "  [-] Backend not listening on 127.0.0.1:8080"
  fi
  echo "=========================================================="
}

restart_scratch() {
  local origin=""
  local reset_data=0
  for arg in "${@:2}"; do
    if [[ "$arg" == "--reset-data" || "$arg" == "--clean" ]]; then
      reset_data=1
    elif [[ "$arg" =~ ^https?:// ]]; then
      origin="$arg"
    fi
  done

  echo "=========================================================="
  echo "   Starting Cogni Backend Scratch / Restart Sequence      "
  echo "=========================================================="

  # 1. Stop conflicting development docker stack if running
  if command -v docker >/dev/null 2>&1; then
    if docker ps --format '{{.Names}}' 2>/dev/null | grep -q '^cogni-fresh-database-1$'; then
      echo "--> [1/9] Stopping conflicting dev containers on port 55432..."
      docker compose -f compose.yaml down >/dev/null 2>&1 || docker compose -f compose.yaml stop >/dev/null 2>&1 || true
    fi
  fi

  # 2. Stop existing PM2 processes
  if command -v pm2 >/dev/null 2>&1; then
    echo "--> [2/9] Stopping existing backend PM2 processes..."
    pm2 stop learning-api learning-worker >/dev/null 2>&1 || true
  fi

  # 3. Ensure configuration and TLS exist
  if [[ ! -f .env.production || ! -d .local/server/tls ]]; then
    if [[ -n "$origin" ]]; then
      echo "--> [3/9] Initializing environment and TLS certificates with origin: $origin"
      if [[ -f scripts/server-infra.sh ]]; then
        bash scripts/server-infra.sh init "$origin"
      else
        echo "Missing scripts/server-infra.sh" >&2
        exit 1
      fi
    else
      echo "--> Error: .env.production or .local/server/tls missing." >&2
      echo "    Run: bash scripts/server.sh init https://YOUR_FRONTEND_DOMAIN" >&2
      echo "    or pass origin: bash scripts/server.sh restart-scratch https://YOUR_FRONTEND_DOMAIN" >&2
      exit 1
    fi
  else
    echo "--> [3/9] Verified .env.production and TLS certificates"
  fi

  # 4. Load & validate environment in subshell (prevent leaking secrets to PM2)
  (load_env)

  # 5. Start Docker backing services
  if (( reset_data == 1 )); then
    echo "--> [4/9] Resetting production data volumes (--reset-data requested)..."
    docker compose --env-file .env.production -f compose.server.yaml down -v >/dev/null 2>&1 || true
  fi
  echo "--> [4/9] Starting backing services (PostgreSQL & SeaweedFS) via compose.server.yaml..."
  if command -v docker >/dev/null 2>&1; then
    docker compose --env-file .env.production -f compose.server.yaml up -d --wait
  else
    echo "Error: docker command not found" >&2
    exit 1
  fi

  # 6. Set up OMR virtual environment if uv is present and omr dir exists
  if command -v uv >/dev/null 2>&1 && [[ -d omr ]]; then
    echo "--> [5/9] Syncing OMR frozen Python environment via uv..."
    uv sync --project omr --frozen
  else
    echo "--> [5/9] Skipping omr virtual environment sync (uv or omr/ not present)"
  fi

  # 7. Build binaries atomically
  echo "--> [6/9] Building Go binaries (gateway, worker, operator)..."
  build_binaries

  # 8. Apply database migrations (scoped in subshell)
  echo "--> [7/9] Applying database migrations..."
  (
    load_env
    if [[ -n ${MIGRATION_DATABASE_URL:-} ]]; then
      export DATABASE_URL="$MIGRATION_DATABASE_URL"
    fi
    .local/server/current/operator migrate
  )

  # 9. Bootstrap admin account if requested via env
  if [[ -n "${BOOTSTRAP_EMAIL:-}" && -n "${BOOTSTRAP_PASSWORD:-}" ]]; then
    echo "--> Bootstrapping owner identity ($BOOTSTRAP_EMAIL)..."
    (
      load_env
      .local/server/current/operator bootstrap || true
    )
  fi

  # 10. Start PM2 services (secrets remain inside child process wrappers)
  echo "--> [8/9] Starting backend API and Worker with PM2..."
  if command -v pm2 >/dev/null 2>&1; then
    pm2 startOrRestart deploy/ecosystem.config.cjs --only learning-api,learning-worker
  else
    echo "Error: pm2 not found in PATH." >&2
    echo "Install PM2: npm install --global --prefix \"\$HOME/.local\" pm2" >&2
    exit 1
  fi

  # 11. Verify health endpoint
  echo "--> [9/9] Verifying backend health ready endpoint..."
  local attempts=0
  local ready=0
  while (( attempts < 15 )); do
    if curl -s -f http://127.0.0.1:8080/health/ready >/dev/null 2>&1; then
      ready=1
      break
    fi
    sleep 1
    (( attempts++ ))
  done

  echo
  if (( ready == 1 )); then
    echo "=========================================================="
    echo "  [✓] Backend successfully restarted & ready on 127.0.0.1:8080!"
    echo "=========================================================="
  else
    echo "=========================================================="
    echo "  [!] Warning: Health endpoint not responding yet."
    echo "      Check logs with: bash scripts/server.sh logs"
    echo "=========================================================="
  fi

  pm2 status
}

case "${1:-}" in
  check)
    cmd_check
    ;;
  init)
    if [[ -f "$root_dir/scripts/server-infra.sh" ]]; then
      bash "$root_dir/scripts/server-infra.sh" init "${2:-}"
    else
      echo "Missing scripts/server-infra.sh" >&2
      exit 1
    fi
    ;;
  restart-scratch|scratch|fresh-start)
    restart_scratch "$@"
    ;;
  infra)
    case "${2:-}" in
      up|status|stop|down)
        if [[ -f "$root_dir/scripts/server-infra.sh" ]]; then
          bash "$root_dir/scripts/server-infra.sh" "$2"
        else
          [[ -f .env.production ]] || { echo 'Run init first.' >&2; exit 1; }
          case $2 in
            up) docker compose --env-file .env.production -f compose.server.yaml up -d --wait ;;
            status) docker compose --env-file .env.production -f compose.server.yaml ps ;;
            stop) docker compose --env-file .env.production -f compose.server.yaml stop ;;
            down) docker compose --env-file .env.production -f compose.server.yaml down ;;
          esac
        fi
        ;;
      logs)
        load_env
        docker compose --env-file .env.production -f compose.server.yaml logs --tail 50 "${3:-}"
        ;;
      *)
        echo 'Usage: bash scripts/server.sh infra up|status|stop|down|logs' >&2
        exit 1
        ;;
    esac
    ;;
  omr)
    command -v uv >/dev/null || { echo 'Install uv: sudo pacman -S uv' >&2; exit 1; }
    uv sync --project omr --frozen
    echo 'Synced frozen Python environment for OMR in omr/.venv'
    ;;
  build)
    build_binaries
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
  stop)
    pm2 stop learning-api learning-worker
    ;;
  logs)
    target="${2:-learning-api}"
    if [[ "$target" == "all" ]]; then
      pm2 logs --lines 40 --nostream
    else
      pm2 logs "$target" --lines 40 --nostream
    fi
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
