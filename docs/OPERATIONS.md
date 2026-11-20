# Operations — current increment

This is development software, not release-qualified infrastructure. Required operations absent from this document remain unimplemented in the register.

## Environment

`.env.example` documents required names. Startup validates APP_ENV, PUBLIC_ORIGIN, and a 32-byte hex MFA_KEY; database connectivity must succeed before serving. Production requires HTTPS and Secure cookies. Keep MFA_KEY stable and backed up separately from the database, or enrolled authenticator secrets become unreadable. DATABASE_URL must point at PostgreSQL with transport verification in production. The local service passwords are generated, never published defaults.

SMTP_ADDR, SMTP_FROM, SMTP_USERNAME and SMTP_PASSWORD configure the SMTP adapter. Production requires STARTTLS with certificate validation. Development uses Mailpit and explicitly labels captured mail. SMTP acceptance is the strongest observable state; it does not prove delivery. A crash after acceptance but before acknowledgment can repeat a message with the same Message-ID. Invitation/reset business tokens remain single-use. No exactly-once email-delivery claim is made.

SeaweedFS 4.54 runs a private local S3 bucket with environment-provided access keys and host-loopback port exposure. A private profile-image path is implemented and tested; broader RUN-03 document/media intake remains incomplete. Do not add public bucket policies. OBJECT_ENDPOINT is an origin URL (no trailing slash), and OBJECT_BUCKET selects the private bucket. Production requires HTTPS and separately scoped service credentials.

## Bootstrap and recovery

Apply migrations: `go run ./cmd/operator migrate`. On an empty installation, use `go run ./cmd/operator bootstrap` with BOOTSTRAP_EMAIL, BOOTSTRAP_NAME, BOOTSTRAP_INSTITUTION, BOOTSTRAP_CAMPUS, BOOTSTRAP_TIMEZONE and a generated BOOTSTRAP_PASSWORD in the process environment. Bootstrap serializes with a PostgreSQL advisory lock and refuses to run after identities exist. The demo command is rejected outside development. Bootstrap accepts explicit institution/campus names and timezone. Owners can edit core institution settings; full production onboarding remains incomplete.

Members use password recovery from sign-in. The reset requires MFA when enabled and invalidates every session and outstanding security token. There is no automatic MFA bypass. Lost authenticator plus lost recovery codes requires a future verified support workflow; do not weaken the database manually to present a demonstration as complete.

Membership suspension preserves records and revokes all of that person's sessions. Restoring the membership never restores those sessions. Global account-state administration, global account-state recovery remain incomplete. Ownership transfer and organization MFA enforcement now have protected working paths.

## Jobs and messages

Run `go run ./cmd/worker` separately from HTTP. Jobs are inserted with security-token changes in one transaction. Workers claim with SKIP LOCKED and a 60-second lease; SMTP has a 15-second deadline. Transient failures back off and fail after five attempts. Institution owners see sanitized status and can retry failed invitations. Payloads are removed after SMTP acceptance. Global recovery messages have no institution operations console yet. Cancellation, heartbeat, full dead-letter tooling and a maintained general-purpose job mechanism are still required under RUN-01/02; this narrow delivery queue does not complete those requirements.

## Deployment and rollback

For a personal Linux server running the API/worker under PM2, with Cloudflare
Tunnel and a Vercel frontend, see [SELF-HOSTING.md](SELF-HOSTING.md). It includes
production environment and tunnel templates, explicit migration/bootstrap steps,
PM2 reboot persistence and the frontend proxy settings. This path still requires
separately provisioned backing services and production qualification.

Do not expose development ports to the public internet. The Go gateway binds loopback by default. Next proxies versioned API traffic; it contains no business rules. Use a trusted TLS reverse proxy and separate least-privileged PostgreSQL roles before release. The local PostgreSQL user owns migrations; production role separation is not configured. Migrations must be reviewed and backed up before applying. Roll back app releases only while schema compatibility is maintained. Destructive `goose down` is not a production rollback strategy.

## Backup and restore qualification

Targets remain RPO 15 minutes and RTO 2 hours, unmeasured. A consistent backup must include PostgreSQL, object storage, separately protected encryption keys and configuration. A full encrypted scheduled backup workflow and restore drill are not yet delivered. Do not claim recoverability from the presence of Docker volumes. Academic, financial, retention, notification and provider runbooks must be expanded with their working features before release.

## Application containers

After initializing `.env`/migrations/demo with `bun run dev`, stop the host app processes and run:

```sh
docker compose -f compose.yaml -f compose.application.yaml up -d --build
```

API/worker share one Go runtime image. Web uses a standalone Next.js build on Bun. All run without root. This configuration remains a local development deployment with local SMTP; it is not a production activation. The current image builds and a separate browser/API/database/S3 container-network smoke journey passed locally. Full production and release journeys remain unqualified. `API_INTERNAL_URL` is compiled into Next rewrites; rebuild web when changing that destination.

## Enrollment and account changes

Enrollment intake issues expiring bounded codes and queues learner registrations. Email must be verified before placement. Failed/expired verification can be resent; pending requests can be placed or rejected with a reason. Self-registration never accepts a role field. Existing accounts request an institution invitation rather than gaining a second membership from a public code.

Password changes revoke other sessions and unused resets. Email changes retain the old verified email until confirmation, then revoke other sessions. Owners can enforce staff MFA only after enabling their own. Staff can always reach personal MFA enrollment to regain institutional capability. Owner transfer reauthenticates, checks a verified staff successor, serializes the institution change, grants the new owner before archiving the caller's owner membership, and records a reason.

Browser suites create a new explicitly synthetic owner fixture per run. This preserves real shared sign-in rate limits across repeated local test runs. Test fixture creation is prohibited outside development.

## Profile-image lifecycle

The browser crops to a 256-pixel PNG. Go bounds request bytes/dimensions, decodes only PNG/JPEG and re-encodes without original metadata. A staged database record is created before upload. The old profile pointer changes only after successful storage and a second authorized transaction. Failed uploads retain the prior image and expose retryable status. Original selected images never leave the browser.

Images are read only through the current person's authorized API route with `no-store`; no object URL is returned. Authenticated people cannot request another person's image by supplying an ID. History is self-only and contains safe storage state, not object keys. Removing an image immediately ends application reads and retires the object. The worker periodically deletes unused retired/abandoned objects; deletion and its database completion can be replayed safely. An active/in-progress image cannot be manually purged. Configuration/storage failure leaves recoverable state.

Staged uploads expire after one hour; replaced/removed images are immediately eligible for cleanup. These are initial profile-image defaults, not a general academic-record retention policy. Production disk encryption, object backup, document quarantine, resumable uploads and all institutional file policies still require implementation/qualification.

To repeat the bounded container check after building the two local images, run `bun run qualify:containers`. It creates an isolated temporary network and synthetic fixture, uses ports 13000/18080, and removes only its temporary containers/network. Existing backing-service volumes remain intact.
