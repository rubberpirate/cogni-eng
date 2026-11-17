# Verification report

Date: 2026-11-17. This report is for the current increment, not release qualification.

## Passed

- Go unit tests: password verification and weak-input rejection, authenticated encryption/tamper rejection, role-template default denial.
- PostgreSQL 18.7 integration: cross-institution denial, cross-tenant campus write rollback, learner escalation denial, membership suspension/reactivation with permanent old-session revocation, forged CSRF rejection, invitation single use, reset-session revocation, TOTP enrollment, recovery-code replay rejection, scoped coordinator denial, and transactional audit creation.
- Next.js 16.4.6 production compilation and strict TypeScript.
- Go vet and ordinary Go test suite. Integration tests are skipped unless TEST_DATABASE_URL is supplied; the integration run above explicitly supplied it.
- Local PostgreSQL and mail service startup, Goose migration, one-time synthetic demo initialization; SeaweedFS S3 service startup.

## In progress / failures corrected

- Initial storage image could not be fetched. Replaced with documented pinned SeaweedFS 4.54; startup succeeds.
- Browser test exposed a password label including the visibility-button text. Corrected label association. Both Chromium journeys passed after fixing explicit labels and hoisting form selectors to preserve selections. A service restart was needed after an interrupted session; the complete startup command was then exercised successfully.

## Unexecuted / incomplete

Object storage entitlement/transaction tests, live SMTP verification, complete milestone 1 acceptance, all academic/paper/learning/AI/finance flows, native tests, provider contract/webhook tests, security audit, full WCAG/browser matrix, required reference load, educational/model evaluations and backup/restore drill. No performance, accuracy, availability or production-readiness claim is established.

Observed dataset: small original synthetic local fixtures and isolated test schemas. No 12,000-learner/600-writer workload was run. Local operation logs are not capacity evidence.

Browser run: 2 passed (11.0s). Owner workflow creates campus/course, invites/admit learner, reloads persistence, inspects audit and verifies logout; phone workflow uses keyboard submission and Hindi navigation without page overflow. Family actions need a dedicated journey and are not covered by that test despite the initial broad test title. Screenshots are local `.local/screenshots/` artifacts.

## Expanded trust increment

Passed: PostgreSQL integration/race checks covering password changes and other-session revocation; pending versus confirmed email changes and token replay; preserved learner references and single placement-event history under stale writes; verified successor ownership transfer; staff MFA enforcement with enrollment recovery; registration verification/placement gates, code exhaustion, role-field rejection and policy acknowledgments. Local SMTP acceptance and recovery of an expired delivery lease passed against Mailpit. These establish local behavior, not live inbox delivery.

Latest complete browser run: **4 passed (29.8s)**, including family approval/removal and the actual register → local captured email → verification → staff placement → learner sign-in journey. The profile/settings/campus journey also passed. The email-token effect replay bug and stale work-area table rendering were found by these tests and fixed.

API/worker image and standalone web image builds passed. Hosted CI and full container-network journeys are unexecuted. Earlier test runs exposed and corrected label association, stale worker lifetime and strict-effect token handling; those failures are not counted as passed release gates. All remaining academic, native, provider, load and recovery checks remain unexecuted.
