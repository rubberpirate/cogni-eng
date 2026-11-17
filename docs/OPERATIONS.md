# Operations — current increment

This is development software, not release-qualified infrastructure. Required operations absent from this document remain unimplemented in the register.

## Environment

`.env.example` documents required names. Startup validates APP_ENV, PUBLIC_ORIGIN, and a 32-byte hex MFA_KEY; database connectivity must succeed before serving. Production requires HTTPS and Secure cookies. Keep MFA_KEY stable and backed up separately from the database, or enrolled authenticator secrets become unreadable. DATABASE_URL must point at PostgreSQL with transport verification in production. The local service passwords are generated, never published defaults.

SMTP_ADDR, SMTP_FROM, SMTP_USERNAME and SMTP_PASSWORD configure the SMTP adapter. Production requires STARTTLS with certificate validation. Development uses Mailpit and explicitly labels captured mail. SMTP acceptance is the strongest observable state; it does not prove delivery. A crash after acceptance but before acknowledgment can repeat a message with the same Message-ID. Invitation/reset business tokens remain single-use. No exactly-once email-delivery claim is made.

SeaweedFS 4.54 runs a private local S3 bucket with environment-provided access keys and host-loopback port exposure. No application file upload/read path is implemented yet; RUN-03 is incomplete. Do not add public bucket policies. Application-level file entitlement and lifecycle tests remain required.

## Bootstrap and recovery

Apply migrations: `go run ./cmd/operator migrate`. On an empty installation, use `go run ./cmd/operator bootstrap` with BOOTSTRAP_EMAIL, BOOTSTRAP_NAME and a generated BOOTSTRAP_PASSWORD in the process environment. Bootstrap serializes with a PostgreSQL advisory lock and refuses to run after identities exist. The demo command is rejected outside development. Bootstrap currently creates an institution with development metadata; institution editing and full production onboarding remain incomplete.

Members use password recovery from sign-in. The reset requires MFA when enabled and invalidates every session and outstanding security token. There is no automatic MFA bypass. Lost authenticator plus lost recovery codes requires a future verified support workflow; do not weaken the database manually to present a demonstration as complete.

Membership suspension preserves records and revokes all of that person's sessions. Restoring the membership never restores those sessions. Global account-state administration, protected owner transfer and organization MFA enforcement remain incomplete.

## Jobs and messages

Run `go run ./cmd/worker` separately from HTTP. Jobs are inserted with security-token changes in one transaction. Workers claim with SKIP LOCKED and a 60-second lease; SMTP has a 15-second deadline. Transient failures back off and fail after five attempts. Institution owners see sanitized status and can retry failed invitations. Payloads are removed after SMTP acceptance. Global recovery messages have no institution operations console yet. Cancellation, heartbeat, full dead-letter tooling and a maintained general-purpose job mechanism are still required under RUN-01/02; this narrow delivery queue does not complete those requirements.

## Deployment and rollback

Do not expose development ports to the public internet. The Go gateway binds loopback by default. Next proxies versioned API traffic; it contains no business rules. Use a trusted TLS reverse proxy and separate least-privileged PostgreSQL roles before release. The local PostgreSQL user owns migrations; production role separation is not configured. Migrations must be reviewed and backed up before applying. Roll back app releases only while schema compatibility is maintained. Destructive `goose down` is not a production rollback strategy.

## Backup and restore qualification

Targets remain RPO 15 minutes and RTO 2 hours, unmeasured. A consistent backup must include PostgreSQL, object storage, separately protected encryption keys and configuration. A full encrypted scheduled backup workflow and restore drill are not yet delivered. Do not claim recoverability from the presence of Docker volumes. Academic, financial, retention, notification and provider runbooks must be expanded with their working features before release.
