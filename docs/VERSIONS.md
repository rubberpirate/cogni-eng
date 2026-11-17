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

- Private local S3: SeaweedFS 4.54, official release https://github.com/seaweedfs/seaweedfs/releases/tag/4.54 and authenticated mini setup https://github.com/seaweedfs/seaweedfs#quick-start
- SMTP adapter APIs: https://pkg.go.dev/net/smtp ; local Mailpit container setup: https://mailpit.axllent.org/docs/install/docker/
- GitHub Actions: checkout 7.0.1, setup-go 7.1.0, upload-artifact 7.0.2, verified official release pages. CI source is not proof of a hosted CI run.

- S3 client minio-go v7.4.1: https://pkg.go.dev/github.com/minio/minio-go/v7 and https://github.com/minio/minio-go
- golang.org/x/crypto pinned to v0.59.0 after reviewing the unreferenced OpenPGP module advisory GO-2026-5932: https://pkg.go.dev/vuln/GO-2026-5932 . The updated scan reports no vulnerabilities.
- Quality tools pinned: govulncheck v1.9.0, golangci-lint v2.15.0 (official release https://github.com/golangci/golangci-lint/releases/tag/v2.15.0).
