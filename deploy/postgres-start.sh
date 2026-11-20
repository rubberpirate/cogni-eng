#!/usr/bin/env bash
set -euo pipefail

# PostgreSQL requires its TLS key to be owned/readable by its own Unix user.
install -d -m 700 -o postgres -g postgres /run/postgresql/tls
install -m 600 -o postgres -g postgres /tls/server.key /run/postgresql/tls/server.key
install -m 644 -o postgres -g postgres /tls/server.crt /run/postgresql/tls/server.crt
exec /usr/local/bin/docker-entrypoint.sh postgres \
  -c ssl=on \
  -c ssl_cert_file=/run/postgresql/tls/server.crt \
  -c ssl_key_file=/run/postgresql/tls/server.key
