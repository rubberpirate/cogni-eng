# Verification report

2026-11-17 · partial milestone 1 · **not release qualification**.

Test environment: Ryoku Linux x86-64 host; Go 1.27.4; Bun 1.4.4; PostgreSQL 18.7 and SeaweedFS 4.54 in Docker; Mailpit 1.28.0; Next.js 16.4.6; Playwright Chromium. Data consists of small original synthetic fixtures and isolated database schemas. No real institutional data or live deliveries/charges were used.

## Passed

- Unit checks: password hashing/verification and weak-input rejection; AES-GCM tamper rejection; default-deny role capabilities; bounded image decoding and malformed/non-square/oversized image rejection.
- PostgreSQL integration: institution/campus authorization; cross-tenant write rollback; learner escalation denial; CSRF rejection; permanent old-session revocation after membership suspension/reactivation; invitation/reset single use; reset revocation; TOTP and recovery-code replay; audit/business transaction consistency.
- Account/organization integration: password change and other-session revocation; pending/confirmed email changes; ownership succession with verified staff; MFA enforcement with an available enrollment path; retained learner reference and one historical placement event under stale writes.
- Registration integration: explicit policy acknowledgments; no role injection; email verification before placement; code exhaustion; duplicate placement denial; one learner membership and enrollment.
- Local SMTP: actual Mailpit acceptance, expired-lease recovery, one durable security token and sensitive payload removal after acceptance. This proves local SMTP acceptance, **not live inbox delivery**.
- Private storage: actual S3 put/read/remove; anonymous object access denied; unauthenticated and other-person application access denied; no-store responses; malformed/failed upload preserves prior image; safe deletion/retry and fresh-staged cleanup rejection. Storage outage uses an unreachable endpoint, not a fabricated successful response.
- Latest complete browser suite: **5 passed (28.5s)**. Campus/course setup, invitation/admission, family approval/removal, persistence/audit/logout; keyboard sign-in/Hindi navigation at 390px width; profile/settings/campus edits; registration → captured verification mail → staff placement → learner sign-in; private crop save/read/remove.
- Strict TypeScript (with Next route type generation), Go vet, ordinary Go tests, Next.js production build and eight focused linters (0 issues).
- Full PostgreSQL/S3/SMTP integration suite passed with race detection and shuffled test order (transport package 122.3s). Targeted image authorization/lifecycle checks passed after final cleanup guards.
- Updated API/worker and standalone web container builds passed. An isolated container-network browser journey passed through web → Go → PostgreSQL/private S3, including authenticated image put/read/remove.
- govulncheck v1.9.0: **no vulnerabilities found** after pinning golang.org/x/crypto v0.59.0. The prior module-only OpenPGP advisory was not reachable and was removed by this update.

## Failed checks corrected

The initial storage image was unavailable and was replaced by pinned SeaweedFS. Browser tests caught ambiguous control labels, remounted form selections, stale work-area rows, development-effect replay clearing email tokens and controls accepting clicks before hydration. A stale `go run` worker consumed a new message kind incorrectly; startup now manages compiled binaries directly. Subsequent tests above passed. No corrected historical failure is represented as an unexecuted passing gate.

## Skipped checks

Go integration tests are excluded without the `integration` tag. Tagged tests explicitly skip when required TEST_DATABASE_URL, TEST_SMTP_ADDR or OBJECT_ENDPOINT settings are absent. The reported service runs supplied those settings. A default unit-only `go test` result is not integration evidence.

## Unexecuted / incomplete

Hosted CI execution; complete milestone 1 acceptance; complete English/Hindi translations; full WCAG 2.2 AA, screen-reader and browser/device matrix; production SMTP/S3 provider qualification; generalized job crash/cancellation tests; backup/restore drill; reference load (12,000 learners/600 writers); availability/recovery targets; academic authoring and online/paper assessment tests; all AI/model/media educational evaluations; collaboration/finance/provider-webhook tests; native/offline tests.

No performance, availability, educational accuracy, production readiness or full-platform completion claim is established. All required scope remains in IMPLEMENTATION-STATUS.md.

Rendering: public pages use server rendering; authenticated workspace identity/context is fetched from Go per request with no shared cache. Controls are disabled until hydration. Private image tests also assert decoded image pixels, not just an image element.
