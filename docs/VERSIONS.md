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

Academic increment, verified 2026-11-17:

- River v0.51.2: https://riverqueue.com/docs ; transactional enqueue https://riverqueue.com/docs/transactional-enqueueing ; migrations https://riverqueue.com/docs/migrations ; isolated search_path https://riverqueue.com/docs/alternate-schema ; pinned API source in go.sum/module cache.
- Excelize v2.11.1: https://xuri.me/excelize/en/workbook.html (`OpenReader`, bounded unzip and raw values).
- Anthropic official Messages API: https://platform.claude.com/docs/en/api/messages/create ; tool use https://platform.claude.com/docs/en/agents-and-tools/tool-use/overview . Request version pinned to 2026-10-21; model is explicit configuration, not an inferred latest model. Contract tests are not live model-delivery evidence.
- Tesseract: https://tesseract-ocr.github.io/tessdoc/Command-Line-Usage.html ; local executable reports 5.5.4/Leptonica 1.87.0. Poppler's installed `pdfinfo -v` reports 26.10.0. Runtime containers install Debian's maintained Poppler/Tesseract packages, so container package versions may differ from the host and require their own qualification.
- golang.org/x/image v0.44.0 provides standard font/raster support; original sheet generation uses its licensed basic font. No external educational assets or inherited product artifacts were used.

Gemini adapter, checked 2026-11-18: official API https://ai.google.dev/api/generate-content and structured output https://ai.google.dev/gemini-api/docs/structured-output (current official redirects to aistudio.google.com). Verified header authentication, systemInstruction, generationConfig.responseMimeType/responseJsonSchema, finishReason, modelVersion and actual usageMetadata fields. User explicitly selected gemini-flash-latest; resolved model versions are recorded per response. Developer Knowledge MCP lookup returned HTTP 403; official documentation retrieval was used instead. Local HTTP contract tests are separate from live provider observations.

Built Go runtime image learning-core:local reports Tesseract 5.3.0/Leptonica 1.82.0 and Poppler 22.12.0; host evaluation uses newer versions listed above. Runtime executable checks passed; full new academic container-flow qualification remains separate.

Sample artifact tool, checked 2026-11-18: ReportLab 5.1.0, official package https://pypi.org/project/reportlab/ . Pinned in the PEP 723 metadata of scripts/build-sample-documents.py; run with uv. It reproduces original public PDFs and is not a production backend dependency.
