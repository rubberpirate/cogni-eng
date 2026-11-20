# Personal server backend, Cloudflare Tunnel and Vercel frontend

These instructions target your **CachyOS/Arch Linux x86_64 server**, logged in as
`ladmin` (`/home/ladmin`). Use that normal user for the application and PM2.
Examples use `~/cogni-fresh`, frontend `https://learn.example.com` and backend
`https://api.example.com`; replace those domains with your own. This guide adds a
hosting path for the current increment; production release qualification remains
open in [IMPLEMENTATION-STATUS.md](../IMPLEMENTATION-STATUS.md).

```text
Browser → Vercel (learn.example.com)
        → /api/v1/* rewrite → Cloudflare (api.example.com)
        → cloudflared on your PC → Go API (127.0.0.1:8080)
                                  ↕ PostgreSQL / private S3
                                  Go worker → scans / jobs / SMTP
```

## 1. Prepare the server

Your login shell is Fish. Open a Bash session for this guide so the bootstrap
password prompt, exports and environment-file syntax work as written. This does
not change your default shell. First run these commands **in Fish**:

```fish
mkdir -p "$HOME/.local/bin"
fish_add_path "$HOME/.local/bin"
bash
```

Run **all remaining commands in Bash**, as `ladmin` unless a command explicitly
uses `sudo`. Type `exit` when finished to return to Fish. Fish has different
[variable and builtin syntax](https://fishshell.com/docs/current/fish_for_bash_users.html).

Install dependencies with a full Arch upgrade:

```sh
sudo pacman -Syu --needed git bash go nodejs npm cloudflared curl openssl ca-certificates tzdata poppler tesseract tesseract-data-eng tesseract-data-hin uv
go version
node --version
npm --version
cloudflared --version
tesseract --list-langs
```

If you already use a supported Node.js LTS package, keep it and omit `nodejs`
from the install command. Review pacman's upgrade/conflict prompts before
confirming. The backend requires Go **1.27.4** or newer (see `go.mod`). Arch's
package versions change; check `go version` after installation. The native OMR
environment uses its frozen Python dependencies through `uv`, rather than the
system Python packages. Bun **1.4.4** is needed on machines building the frontend;
the server script needs Go, Bash and PM2.

Arch provides [cloudflared](https://archlinux.org/packages/extra/x86_64/cloudflared/),
[Go](https://archlinux.org/packages/extra/x86_64/go/),
[uv](https://archlinux.org/packages/extra/x86_64/uv/) and
[Tesseract language data](https://archlinux.org/groups/x86_64/tesseract-data/)
through its package repositories; no AUR helper is needed for these dependencies.
Install PM2 into your user's home directory:

```sh
npm install --global --prefix "$HOME/.local" pm2
export PATH="$HOME/.local/bin:$PATH"
pm2 --version
```

This uses npm's [user-owned prefix approach](https://docs.npmjs.com/resolving-eacces-permissions-errors-when-installing-packages-globally/)
without `sudo npm`. The earlier `fish_add_path` keeps PM2 available in future
Fish sessions, and PM2's startup command below records the path for systemd.

Use a domain with DNS managed by Cloudflare for the named tunnel. The server must
stay powered on with sleep disabled and outbound Internet access (including
Cloudflare port 7844); no router port forwarding or public IP is required.
In KDE Plasma, open **System Settings → Power Management** and turn off automatic
suspend for the power profile used by this server. A sleeping PC cannot serve
requests through the tunnel.

Keep the checkout, Python environment and any local database/object data on the
server's Linux filesystem, for example `/home/ladmin/cogni-fresh` on the Btrfs
root disk. The mounted NTFS dataset drive is unnecessary for this deployment.

Clone this repository on the server, then `cd ~/cogni-fresh`. Do not run
`bun run dev` or copy your development `.env` here.

For Docker-hosted PostgreSQL and private storage on this same PC, follow
[SERVER-DATABASE.md](SERVER-DATABASE.md) first. It generates `.env.production`,
passwords and local TLS certificates and starts the server containers. Continue
with the OMR setup and migrations below once the environment is ready.

Alternatively, provision these backing services with a provider:

| Service                       | Required configuration                                                                                                                                                                                                                   |
| ----------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| PostgreSQL                    | A fresh database, runtime credentials and a migration owner. Use TLS verification for network connections, e.g. `postgres://USER:URL_ENCODED_PASSWORD@DB_HOST:5432/workspace?sslmode=verify-full`. Supply the provider CA when required. |
| Private S3-compatible storage | Existing private bucket, HTTPS endpoint and credentials scoped to that bucket. Objects are accessed through Go; no public bucket or separate browser CORS setup is needed.                                                               |
| SMTP                          | Sender address and server supporting STARTTLS with a trusted certificate, normally port 587. Mailpit is only for development.                                                                                                            |

The repository's `compose.yaml` is a **development** stack. It has no production
TLS setup, backup automation or reboot policy. The dedicated `compose.server.yaml`
provides local PostgreSQL/S3 with TLS and reboot policies; SMTP credentials still
come from your provider. A managed database/private object store is also
compatible with a personal PC running the API and worker. Keep
database, storage, SMTP and optional model ports private; only the API goes into
the tunnel.

## 2. Configure the backend

If you used `server-infra.sh init`, `.env.production` is already populated.
Keep its generated database/storage/TLS/MFA values and edit SMTP/model settings
as described in [SERVER-DATABASE.md](SERVER-DATABASE.md). Skip copying the
template and generating another MFA key below. The following manual setup is
for externally provisioned services.

```sh
umask 077
cp -n deploy/backend.env.example .env.production
chmod 600 .env.production
openssl rand -hex 32
```

Edit `.env.production` locally. Paste that random hex string into `MFA_KEY`, fill
in your backing-service credentials and set:

```sh
APP_ENV=production
HTTP_ADDR=127.0.0.1:8080
PUBLIC_ORIGIN=https://learn.example.com
```

`PUBLIC_ORIGIN` is the **frontend** origin, without a trailing slash. This can be
your stable `https://YOUR_PROJECT.vercel.app` address if you have no frontend
custom domain yet. `OBJECT_ENDPOINT` must be an HTTPS origin without a path or
trailing slash. Values follow Bash syntax: single-quote secrets and values with
spaces or shell characters. Only source this trusted, privately edited file.
The file is ignored by Git. Keep `MFA_KEY` stable across releases and back it up
separately; changing it makes existing enrolled MFA secrets unreadable.

Set `MIGRATION_DATABASE_URL` if migrations use a separate database owner. Grant
the runtime role the table/sequence access required by the API and River worker
after migrations. Do not put bootstrap credentials in this file. AI keys stay on
the server; missing model configuration leaves AI jobs unavailable.

For **saved-template OMR scanning**, the pinned engine source is already bundled
in `omr/engine`. Install its frozen CPU environment from the project root:

```sh
uv sync --project omr --frozen
omr/.venv/bin/python -B -m unittest discover -s omr -p 'test_*.py'
```

Leave `OMRCHECKER_ROOT`, `OMRCHECKER_PYTHON` and `OMRCHECKER_ADAPTER` empty in
`.env.production`. PM2 runs from the project root and finds `omr/engine`,
`omr/.venv/bin/python` and `omr/bridge.py` there. Remove any old overrides pointing
to a separate OMRChecker checkout. No second repository, submodule or Git command
is needed at runtime. See [OMR setup](../omr/README.md).
Optional GPU OCR/Ollama services have their own [runtime setup](../inference/README.md);
PM2 here manages the API, worker and tunnel, so separately supervise any model
service you enable. Use loopback model endpoints.

## 3. Build, migrate and create the owner

```sh
bash scripts/server.sh build
bash scripts/server.sh migrate
```

`build` compiles only the three Go binaries and switches `.local/server/current`
after all builds succeed. `migrate` explicitly applies both application and River
migrations; it does not start the API. Back up an existing database and review
migrations before running it.

On a **new, empty database**, bootstrap once from the **Bash session** opened
above:

```sh
export BOOTSTRAP_EMAIL='you@example.com'
export BOOTSTRAP_NAME='Your Name'
export BOOTSTRAP_INSTITUTION='Your Institution'
export BOOTSTRAP_CAMPUS='Main Campus'
export BOOTSTRAP_TIMEZONE='Asia/Kolkata'
read -r -s -p 'Initial owner password (at least 14 characters): ' BOOTSTRAP_PASSWORD
echo
export BOOTSTRAP_PASSWORD
bash scripts/server.sh bootstrap
unset BOOTSTRAP_PASSWORD BOOTSTRAP_EMAIL BOOTSTRAP_NAME BOOTSTRAP_INSTITUTION BOOTSTRAP_CAMPUS BOOTSTRAP_TIMEZONE
```

The password is read without terminal echo and is not placed in shell history.
Use a unique random password from your password manager. Bootstrap refuses an
already initialized database. No demo accounts are created by this setup.

## 4. Start the API and worker with PM2

```sh
bash scripts/server.sh start
pm2 logs learning-api --lines 30 --nostream
pm2 logs learning-worker --lines 30 --nostream
curl --retry 15 --retry-connrefused --retry-delay 2 --fail http://127.0.0.1:8080/health/ready
```

Both processes must stay online without restart loops. The readiness check tests
database connectivity; it does not establish SMTP, bucket, OMR or worker health.
PM2 runs Bash wrappers that load `.env.production` then `exec` the Go binaries.
Edits to that file take effect after `bash scripts/server.sh restart`. Run PM2
from a clean shell without exported application secrets; the wrapper loads them
inside the child process rather than recording them in the PM2 config.

## 5. Create a named Cloudflare Tunnel

Run as the **same server user** that runs PM2:

```sh
cloudflared tunnel login
cloudflared tunnel create learning-backend
mkdir -p "$HOME/.cloudflared"
cp -n deploy/cloudflared.yml.example "$HOME/.cloudflared/config.yml"
```

On a headless server, open the login URL on your own computer and authorize the
correct Cloudflare zone. Edit `~/.cloudflared/config.yml`: replace the tunnel UUID,
the credentials file's absolute path
(`/home/ladmin/.cloudflared/YOUR_TUNNEL_UUID.json`) and `api.example.com`. The origin service
remains **HTTP** `http://127.0.0.1:8080`; Cloudflare handles public HTTPS.
The catch-all 404 rule is required. Keep tunnel credentials outside the repo.

```sh
chmod 700 "$HOME/.cloudflared"
chmod 600 "$HOME/.cloudflared/config.yml" "$HOME/.cloudflared/"*.json
cloudflared tunnel --config "$HOME/.cloudflared/config.yml" ingress validate
cloudflared tunnel route dns learning-backend api.example.com
bash scripts/server.sh tunnel
pm2 logs learning-tunnel --lines 30 --nostream
curl --retry 15 --retry-delay 2 --fail https://api.example.com/health/ready
```

If the hostname already has a DNS record, resolve that conflict in Cloudflare DNS
before creating the route. Use this stable named tunnel for deployment. Keep
Cloudflare caching bypassed for API traffic and avoid browser challenges or an
interactive Cloudflare Access login on this hostname: Vercel's proxy must be able
to reach it. Authentication/authorization remain enforced by Go. This setup makes
the API hostname publicly reachable; the tunnel does not itself restrict callers
to Vercel. See [Cloudflare's named-tunnel guide](https://developers.cloudflare.com/cloudflare-one/networks/connectors/cloudflare-tunnel/do-more-with-tunnels/local-management/create-local-tunnel/).

## 6. Restore everything after reboot

```sh
pm2 startup systemd -u ladmin --hp /home/ladmin
# Execute the exact sudo command PM2 prints for your server user and PATH.
pm2 save
systemctl status pm2-ladmin --no-pager
```

Run `pm2 startup` as `ladmin` first; only its printed installation command
uses sudo. Do not also install `cloudflared` as a system service when this PM2
entry manages it. Configure your backing services to start at boot too. Reboot
the server and check `pm2 status`, all three logs and the public readiness URL.
After changing your Node installation, regenerate PM2's startup integration.
See [PM2 startup and persistence](https://pm2.keymetrics.io/docs/usage/startup/).
If startup fails, inspect `journalctl -u pm2-ladmin -b --no-pager` and verify the
generated service includes `/home/ladmin/.local/bin` in its PATH. Do not run
`sudo pm2 start` or `sudo pm2 save`; that creates a separate root-owned PM2 setup.

Keep logs bounded:

```sh
pm2 install pm2-logrotate
pm2 set pm2-logrotate:max_size 10M
pm2 set pm2-logrotate:retain 7
pm2 save
```

## 7. Deploy only the frontend on Vercel

Import this Git repository into Vercel and configure the project:

| Setting                                     | Value                                                 |
| ------------------------------------------- | ----------------------------------------------------- |
| Framework preset                            | Next.js                                               |
| Root directory                              | `apps/web`                                            |
| Include source files outside Root Directory | Enabled (the app imports shared `clients/` contracts) |
| Install command                             | `cd ../.. && bun install --frozen-lockfile`           |
| Build command                               | `bun run build`                                       |
| Output directory                            | Framework default                                     |
| Production environment variable             | `API_INTERNAL_URL=https://api.example.com`            |

`API_INTERNAL_URL` has no trailing slash or `/api/v1` suffix. Set it **before
building**, because [next.config.ts](../apps/web/next.config.ts) compiles it into
the existing `/api/v1/:path*` external rewrite. Redeploy Vercel after changing it.
This value is not a `NEXT_PUBLIC_` variable. Do not copy the backend environment
or credentials into Vercel.

Assign `learn.example.com` to the Vercel project (or use its stable production
`vercel.app` address), ensure backend `PUBLIC_ORIGIN` matches it exactly and
restart the backend if you changed that setting. Browsers keep making
same-origin `/api/v1` requests through Vercel, preserving the host-only Secure,
HttpOnly session cookie and CSRF checks. No browser CORS changes are required.
Production mutations from a different preview deployment domain are intentionally
rejected; give previews a separate backend/database and matching origin if needed.
See [Vercel monorepos](https://vercel.com/docs/monorepos) and
[external rewrites](https://vercel.com/docs/routing/rewrites).

Verify from the actual frontend domain:

```sh
# An anonymous request should return 401 JSON, not an HTML error/Cloudflare page.
curl -i https://learn.example.com/api/v1/me
```

Then sign in in the browser, reload, sign out, and test an invitation, private
upload and queued OMR job. Confirm the worker completes the job and the result
persists after a PM2 restart. Check large real uploads against your Cloudflare and
Vercel plan limits; this setup does not bypass their upload/time limits.

## Updates and troubleshooting

Deploy one reviewed commit at a time. Back up PostgreSQL, private objects and
configuration first; save the previous commit ID and do not store backups in Git.
For a maintenance deployment:

```sh
git pull --ff-only
bash scripts/server.sh build
pm2 stop learning-api learning-worker
bash scripts/server.sh migrate
bash scripts/server.sh start
curl --fail http://127.0.0.1:8080/health/ready
pm2 save
```

This causes a brief API outage. If migration fails, keep the application stopped,
inspect the failure and repair or restore the database before restarting. Old
binaries remain in `.local/server/releases/`; roll back to the previous reviewed
source commit and rebuild/restart only if it is compatible with the current
schema. Never use destructive down migrations as an automatic rollback.

| Symptom                                      | Check                                                                                                                     |
| -------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------- |
| Cloudflare 502                               | `learning-api` logs and local `/health/ready`; tunnel service must be `http://127.0.0.1:8080`.                            |
| Vercel API 404/502 or local connection error | Correct production `API_INTERNAL_URL`, then rebuild/redeploy the frontend.                                                |
| `origin_denied` on sign-in or writes         | `PUBLIC_ORIGIN` exactly equals the frontend browser origin, including scheme and without a trailing slash.                |
| Login is lost after reload                   | Test through the frontend domain; inspect the `/api/v1` response's `Set-Cookie`. Disable caching/challenges on API paths. |
| Job stays queued                             | `pm2 logs learning-worker`; check database, private bucket and selected provider/native runtime configuration.            |
| Processes absent after reboot                | PM2 startup service, `pm2 save`, correct user/PATH, backing-service startup and server sleep settings.                    |

The guide and scripts do not implement automated backups or a restore drill.
The recovery and production qualification limits in [OPERATIONS.md](OPERATIONS.md)
still apply; volumes and a green readiness endpoint are not recovery evidence.
