# Version verification

Checked 2026-11-17 against official documentation and package registries:

- Go stable 1.27.4: https://go.dev/dl/?mode=json (host initially 1.27.3).
- Next.js 16.4.6: https://nextjs.org/docs/app/getting-started/installation
- React 19.3.1: registry exact version; https://react.dev/learn/creating-a-react-app
- Bun 1.4.4: https://bun.sh/docs/installation
- PostgreSQL 18.7: https://www.postgresql.org/docs/current/
- Expo SDK 58: https://docs.expo.dev/versions/latest/
- pgx v5: https://pkg.go.dev/github.com/jackc/pgx/v5/pgxpool
- Goose: https://pkg.go.dev/github.com/pressly/goose/v3
- sqlc: https://docs.sqlc.dev/en/stable/tutorials/getting-started-postgresql.html

Exact installed Go dependency versions live in go.mod/go.sum; JS in package manifests/bun.lock. Public documentation determines supported APIs; installed source is authoritative for exact dependency signatures. Provider integrations must add their official sources here when implemented. No integration is live-verified by this document.
