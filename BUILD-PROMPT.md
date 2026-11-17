# Build a complete learning and assessment platform

You are responsible for implementing a new production application from an empty repository. Read the accompanying `PRODUCT-SPEC.md` completely before making changes. These two documents are sufficient to begin; there are no required reference repositories, screenshots, datasets, inherited assets, credentials, or prior application artifacts.

Build the complete scope, including administration, operational tools, integrations, mobile access, and recovery paths. Deliver working software through ordered milestones. A milestone is an implementation sequence, not permission to discard later requirements or finish with only a prototype.

## 1. Product outcome

Create an institution-oriented platform where learners can take assessments online or on paper, receive defensible results and useful feedback, study targeted material, ask a contextual tutor for help, and practice until improvement is demonstrated. Teachers need tools to author, review, intervene, and teach. Families need an appropriate view of progress. Institutional staff need enrollment, communication, reporting, payments, and reliable operations across campuses.

Treat the supplied specification as the product contract. Do not substitute attractive dashboard screenshots for functioning workflows. A control must perform the operation it promises, enforce permissions, persist the result, and recover sensibly from failure.

Use a neutral working identity until the owner supplies a brand. Create original copy, visual assets, terminology, navigation, domain model, contracts, and code. Use newly authored, licensed demonstration material. Do not copy an existing application's name, layout, route tree, package organization, schema, prompts, assets, seed records, credentials, or source code. Standard framework conventions and ordinary educational terms are permitted.

## 2. Fixed technical decisions

- **Backend:** Go. All application business rules, authorization, assessment scheduling and grading, learning calculations, payments, AI orchestration, and job coordination belong here.
- **Web:** Next.js App Router with React and strict TypeScript. Use server rendering for public/discoverable content and appropriate initial views; use client components for the exam player, editors, chat, offline synchronization, and other interactive work.
- **Native mobile:** React Native with Expo and TypeScript for Android and iOS, using the same Go contracts. This is required in the full scope, alongside responsive web and installable web access.
- **Primary data:** PostgreSQL, explicit SQL migrations, and typed database access. Redis may be used for shared rate limiting, presence, short-lived cache, and fan-out; it must not be the sole durable home of examination or financial records.
- **Files:** private object storage with an S3-compatible local implementation. Deliver application file authorization and lifecycle policies, not public guessed object URLs.
- **Contracts:** versioned HTTP API documented with OpenAPI; generate typed clients. Use server-sent events or WebSockets where progress/chat/presence benefits from live updates, with reconnection and polling fallback where appropriate.
- **Background work:** durable jobs, retry policy, scheduling, cancellation, progress, dead-letter/recovery UI, and transactional handoff from committed business changes. Implement these in Go. Select the smallest maintained mechanism that meets the required durability and concurrency semantics.
- **External capabilities:** provider adapters for models, OCR/document processing, speech, video generation, email, SMS/push, calendar export, and payment collection. Implement real adapters plus explicit development test doubles. Verify current official provider documentation before integration.
- **Delivery:** reproducible local environment, containers, CI, environment validation, automated checks, telemetry, backup/restore procedures, and deploy/runbooks.

Use current supported stable releases verified against official documentation at implementation time, then pin versions and commit lockfiles. Do not infer a framework API from memory when the installed version has authoritative local or official documentation.

Keep a small, capability-oriented Go application with explicit transaction boundaries and separately runnable background execution. Avoid speculative microservices, a generic plugin platform, or a second TypeScript business backend. Frontend server handlers may bridge authentication/rendering but must not become another source of business truth. Native graphics/OCR/media tools or hosted inference services may be called through bounded interfaces; they do not justify reimplementing the product backend in another language.

The rationale for the web choice is that one product needs discoverable public pages and highly interactive authenticated tools. Next.js supports both in one React application. It does not replace the Go backend. See the [Next.js component guidance](https://nextjs.org/docs/app/getting-started/server-and-client-components), [React framework guidance](https://react.dev/learn/creating-a-react-app), and [Go module guidance](https://go.dev/doc/modules/layout).

## 3. First actions

1. Inspect only the new workspace and the two supplied documents. Identify environment constraints without searching unrelated directories for product code.
2. Build a requirement register from every `AREA-NN` requirement in the specification. Expand the detail sentences into acceptance checklists so a broad requirement cannot hide incomplete subfeatures.
3. Create `IMPLEMENTATION-STATUS.md` with requirement, milestone, state, evidence, relevant test, and blocker. All requirements start unimplemented. Use `planned`, `building`, `verified`, or `blocked`; never call a mock-backed integration production-verified.
4. Record a concise original architecture, authorization matrix, domain terminology, API conventions, and decision log. Choose package names and folder boundaries from the new design. Do not begin by recreating a conventional role-by-role clone of some other product.
5. Establish the runnable environment, migrations, deterministic fixtures, authentication, tenant boundaries, contract generation, and CI. Then implement vertical workflows in dependency order.
6. Design the interface from the product's work: selecting an institution/campus, a learner's day, preparing an assessment, reviewing evidence, teaching, and resolving operational issues. Build an accessible visual system; do not scatter unrelated component styles across screens.

Proceed with reasonable reversible defaults documented in the decision log. Ask only for missing information that actually blocks an irreversible action or a product-critical decision. Continue independent authorized work while an external configuration item is pending. Do not stop at a plan or ask whether to begin.

## 4. Rules that may never be compromised

- Every request is authorized by actor, institution, resource relationship, and capability. Hidden controls are not access control.
- A platform operator has no automatic right to read a child's private records. Emergency access is exceptional, justified, time-limited, and audited.
- Grade fixed-answer assessments deterministically. Explain every awarded/deducted mark. Subjective AI marks are reviewable proposals until the applicable approval policy releases them.
- Published assessment revisions are immutable. Question wording, assets, answer keys, marking policy, accommodations, and release policy are pinned for each sitting.
- Submission, expiry, duplicate uploads, replayed webhooks, regrading, and retried jobs must not duplicate academic evidence or money movements.
- A server deadline remains valid while a browser is closed. A missing timing measurement remains unknown. An optional blank is not evidence of ignorance.
- AI-generated interpretation is distinguishable from observed fact. Never fabricate emotional diagnoses, model validation, predicted ranks, accuracy rates, or outcomes.
- Learners must not obtain unreleased answers through tutor history, document search, exports, media URLs, notifications, or repeated question exposure.
- Personal records must not contaminate shared generated material or other institutions' search/AI contexts.
- An external failure must not destroy saved work or silently convert uncertain reading into a mark. Preserve evidence, expose useful status, and provide human recovery.
- Keep disabled accounts' academic history, while ending their access immediately. Retention/deletion follows explicit policies rather than incidental cascading deletes.
- Do not display invented billing totals or claim an unconfigured provider has sent, charged, refunded, rendered, or delivered anything.

## 5. Implementation order

Implement all milestones; update the status register after each meaningful increment.

| Milestone | Demonstrable outcome |
|---|---|
| 1 — Trust and organization | Institution/campus setup, memberships, learner/family enrollment, staff responsibilities, recovery/MFA, sessions, audit, accessibility foundations. |
| 2 — Academic authoring | Syllabus, versioned question library, imports, rich content, paper/key review, ordered assessment composition, schedules, accommodations, assignment. |
| 3 — Complete online assessment | Reliable player, saved answers, reconnection, conduct policy, server expiry, deterministic grading, controlled result release, appeals/regrading. |
| 4 — Evidence from paper | Multiple sheet layouts, calibration, asynchronous batch intake, uncertain-read review, correction audit, safe commit, subjective-script review. |
| 5 — Learning and support | Defensible topic evidence, learning indicators, revision journal, contextual tutor, teacher interventions, goals/calendar, generated and authored lessons, spaced practice. |
| 6 — Content and adaptive practice | Reviewed content workflows, narration/video studio, captions, concept animations, personalized follow-up papers, live-adaptive practice mode, media operations. |
| 7 — Institution collaboration | Cross-campus reporting, guardian workspace, attendance, messaging/forums, announcements, notifications, report builder and scheduled exports. |
| 8 — Business and advanced insight | Subscriptions/payments/refunds, financial reconciliation, analytical dashboards, explainable readiness/forecasts, self-reported confidence support, model evaluations. |
| 9 — Mobile and release qualification | Native apps and offline study/sync, push delivery, web parity/accessibility, load/failure testing, security checks, restore drill, operator/support documentation. |

Build shared foundations when first needed, without postponing their user-facing operations screens indefinitely. Every backend-only capability that staff need to run the institution needs a complete authorized UI by completion.

## 6. Verification expectations

Use tests that establish behavior, especially across boundaries. Cover:

- Go unit/property tests for marking, numerical acceptance, optional sections, event ordering, eligibility, learning-evidence accounting, deduplication, and monetary calculations.
- Integration tests against actual PostgreSQL and local object storage for transactions, tenant isolation, authorization, job recovery, imports, attachment access, and result revisions.
- Provider contract/webhook tests, including replay, signature failure, timeout, quota exhaustion, cancellation, and refund reconciliation. Never use a fake provider response as proof of live delivery.
- Browser tests for complete learner, teacher, family, institution, scanning/review, and finance journeys, including keyboard-only use and narrow screens.
- Mobile tests for authentication, protected offline storage, network interruption, synchronization conflicts, accessibility, and child switching.
- Adversarial tests for object access, revoked sessions, upload abuse, prompt injection, answer leakage, malformed answers, client-clock manipulation, and duplicate financial events.
- Worker-crash tests at transaction boundaries, retry after partial success, and consistent recalculation after an older result is corrected.
- Evaluation fixtures for OCR fidelity, generated content grounding, tutor quality, forecast calibration, and behavior with insufficient data. Report the dataset and limitations.

Do not write a huge collection of shallow implementation-mirroring tests. Write enough meaningful tests to establish the contract. Run targeted tests during development and the complete required gates before milestone/release completion.

## 7. What “implemented” means

A requirement is verified only when its backend behavior, permissions, migrations/data effects, user interface, error/recovery path, and meaningful acceptance evidence are complete. For an external feature, also deliver its actual adapter, configuration validation, sandbox/integration evidence, and operator instructions. Track pending production credentials separately from code completion and production verification.

No placeholder buttons, fake reports, hardcoded production credentials, decorative AI claims, silent no-op settings, or unimplemented “coming soon” screens count as delivered features. A required feature may explicitly report insufficient evidence or missing provider configuration; it must still have a functional configured path and a tested recovery workflow.

All listed scope is required. Sequence complexity through milestones; do not silently relabel native apps, payment processing, subjective review, video generation, forecasting, or collaboration as optional. Features needing a validated dataset must include the actual implementation/evaluation path and an honest insufficient-data state, never fabricated predictions.

## 8. Required handoff

Deliver:

1. Original application source, database migrations, typed API contract/clients, workers, web and native apps.
2. A repeatable local startup command and documented environment values. A developer should not need production secrets to run deterministic demonstrations.
3. A requirement-by-requirement status register with concrete evidence and remaining external setup.
4. Meaningful automated checks and a test/evaluation report separating passed, failed, skipped, and unexecuted checks.
5. Original representative demo content with appropriate licensing/provenance and synthetic identities. No private institutional records.
6. Operations documentation covering deployment, providers, jobs, backup/restore, retention, access recovery, payment reconciliation, audit, and support.
7. A short decision log and known limitations tied to requirements. Include observed performance and the tested workload rather than unsupported production claims.

At each milestone, report what works, how it was verified, and the next concrete work. At the final handoff, state completion honestly against the entire register. Do not claim completion because a homepage renders, a build passes, or a provider test double returns success.

**Begin by reading `PRODUCT-SPEC.md`, creating the complete requirement register, and implementing the first runnable milestone. Continue through the full scope.**
