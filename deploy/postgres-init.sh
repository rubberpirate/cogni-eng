#!/usr/bin/env bash
set -euo pipefail

psql --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" \
  --set ON_ERROR_STOP=1 --set app_password="$APP_DATABASE_PASSWORD" <<'SQL'
CREATE ROLE workspace LOGIN PASSWORD :'app_password';
GRANT CONNECT ON DATABASE workspace TO workspace;
GRANT USAGE ON SCHEMA public TO workspace;
ALTER DEFAULT PRIVILEGES FOR ROLE workspace_owner IN SCHEMA public
  GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO workspace;
ALTER DEFAULT PRIVILEGES FOR ROLE workspace_owner IN SCHEMA public
  GRANT USAGE, SELECT, UPDATE ON SEQUENCES TO workspace;
SQL
