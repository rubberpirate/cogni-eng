# Verification report

2026-11-17 · partial milestone 1 · **not release qualification**.

Test environment: Ryoku Linux x86-64 host; Go 1.27.4; Bun 1.4.4; PostgreSQL 18.7 and SeaweedFS 4.54 in Docker; Mailpit 1.28.0; Next.js 16.4.6; Playwright Chromium. Data consists of small original synthetic fixtures and isolated database schemas. No real institutional data or live deliveries/charges were used.

## Core education increment — 2026-11-18

This is qualification of connected core workflows, **not completion of milestones 2–5 or a production release**.

Passed:

- Exact marking fixtures: selection sets, penalties, exact decimal boundary/tolerance acceptance, malformed numbers, first/best optional contribution, uncounted blank treatment, rubric proposals/overflow, and 18×3 + 8×2 = 70. Number parsing also has seed fuzz checks.
- Actual image pixels across original one/two-column registered OMR layouts: filled/blank/partial circles, mismatched identity, rotation, projective distortion and changed exposure. These are synthetic fixtures; no field accuracy percentage is inferred.
- Actual PostgreSQL/private S3: import errors and repeated-file deduplication, immutable published key after bank revision, answer secrecy, stale/replayed saves, duplicate submission, browser-closed worker expiry, pixels → durable scan job → reviewed commit → ordinary marks, and repeated commit safety. Added cross-subject private asset denial, ended-teaching denial and no leakage of a new bank diagram into an older pinned sitting.
- Actual River worker recovery from missing model configuration into an explicit development fixture; generated output remains draft and cannot enter a scored paper without approval. This demonstrates recovery/approval behavior, **not live AI delivery**.
- Anthropic/Gemini HTTP contract tests for headers, model/version handling, structured output, real usage fields, omitted internal thought parts, malformed/truncated/blocked output, signature-free API-key authentication failures, quota/HTTP errors, cancellation and timeouts. No provider response body or key is exposed in application errors.
- Actual Poppler multi-page rasterization/cancellation/malformed-PDF handling. Actual host Tesseract transcription of one original printed PDF fixture, followed by application text mapping. This does not qualify handwriting, photographs, equation/figure reconstruction or the complete OCR import journey.
- Latest full normal PostgreSQL/S3/SMTP suite passed (transport 35.3s). Full core/foundation race suite passed earlier in this increment (transport 192.0s); the newer targeted academic race suite passed (61.6s). The final subject-file entitlement check passed separately (1.8s); it was not included in that earlier full race run.
- Latest complete browser suite: **6 passed (1.0m)**. The new teacher journey imports a key, approves a question, composes/publishes/assigns papers, uploads a real filled PNG, confirms review and obtains marks. A separate learner uses a 390px viewport and keyboard selection, saves/submits and sees the released result/solution and guarded tutor surface. The final focused academic rerun passed (16.5s), including topic evidence and the compact answer controls. Screenshots were inspected locally.
- Seven manual synthetic accounts passed actual sign-in, identity and institution checks; staff saw the expected four subject banks. Credentials are protected ignored local files, not report content. Account creation retains passwords/work on rerun.
- Contract generation, strict TypeScript, Go vet/tests, production Next build and focused lint (0 issues). `govulncheck` found **0 reachable vulnerabilities**, 0 affected imported packages and 1 module-only advisory in uncalled code.
- Updated Go runtime image built successfully (`learning-core:local`); native executable checks inside it report Tesseract 5.3.0/Leptonica 1.82.0 and Poppler 22.12.0. Host native evaluation used Tesseract 5.5.4/Poppler 26.10.0. This is a build/executable check; the full new academic browser flow was tested on host processes with real local services, not rerun inside this image.

Failed / external:

- Live Gemini `gemini-flash-latest` structured generation using the owner-supplied local key returned HTTP 503 on the first attempt, then timed out waiting for response headers on a bounded retry. A minimal request matching the owner's example also timed out. Consequently the live generation → tutoring check did **not** pass. No live success, token total, charge, quality or resolved model version is asserted. Configuration is retained privately for a later retry. Contract tests do not replace this evidence.
- Google Developer Knowledge MCP returned HTTP 403; current official Gemini API/reference pages were retrieved directly instead.

Skipped / unexecuted:

- Tagged integration tests skip when service variables are absent; reported runs provided PostgreSQL/private S3 and, in the complete run, SMTP values. Native tests explicitly skip if tools are absent; reported native runs found and executed them.
- Live Anthropic checks, successful live Gemini generation/tutoring, full model/OCR educational evaluation datasets, worker-process crash injection, protected refresh/offline exam synchronization, generalized sheet calibration, full document import acceptance, complete subjective moderation/appeals/regrading and the remaining product register are unexecuted/incomplete. No reference workload/load, recovery SLO or full WCAG/device matrix has been qualified.

## Earlier trust-foundation passed checks

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

Hosted CI execution; complete milestone 1 acceptance; complete English/Hindi translations; full WCAG 2.2 AA, screen-reader and browser/device matrix; production SMTP/S3 provider qualification; generalized job crash/cancellation tests; backup/restore drill; reference load (12,000 learners/600 writers); availability/recovery targets; complete academic milestone acceptance beyond the core evidence above; full AI/model/media educational evaluations; collaboration/finance/provider-webhook tests; native/offline tests.

No performance, availability, educational accuracy, production readiness or full-platform completion claim is established. All required scope remains in IMPLEMENTATION-STATUS.md.

Rendering: public pages use server rendering; authenticated workspace identity/context is fetched from Go per request with no shared cache. Controls are disabled until hydration. Private image tests also assert decoded image pixels, not just an image element.

## Teaching usability and original practice kit — 2026-11-18

Passed:

- Latest complete browser suite: **10 passed (1.6m)**. Covers foundation journeys plus teacher import/publish/scan/online assessment; real private ZIP download and mixed PDF to 5/12; ambiguous sheet rejection then explicit correction to 12/12; numerical authoring through normal fields with backend persistence; generation error/retry surface, browser back and narrow-screen keyboard Menu. The generation-error response in that UI test is an explicit HTTP fixture, not live-provider evidence.
- Complete actual PostgreSQL/S3/SMTP integration suite passed (transport 37.144s). Updated academic integration passed (11.446s); final sample-key-edit/preset guards passed (4.060s). Tests cover private kit authorization, actual image recognition, immutable published revisions, preserved teacher edits and insufficient evidence for required blanks.
- Go race checks passed: academic 112.879s and scholar 10.359s. Registration fixtures include the known dense-header case and 32 randomized headers across four rotations, alongside perspective/exposure fixtures. These are synthetic rasters, not field-scan accuracy evidence or application throughput measurements.
- Generated sample PDF round-trip through actual Poppler preserves sheet identity and answers. Original paper/key and OMR PDFs were rendered and visually inspected. Six original questions, correct/mixed/uncertain/blank and duplicate-page variants form a deliberately small functional dataset.
- Strict TypeScript, Go vet/unit checks and focused lint (0 issues) passed. Latest production Next build passed after correcting the initial selected-results prop. UI screenshots were inspected at desktop and narrow widths.

Failed / external: a fresh explicit live Gemini model check again returned **HTTP 503**. Successful live generation/tutoring, model quality and resolved model version remain unverified. No successful provider response or delivery is fabricated.

Unexecuted / remaining: real-photo/erasures/printer diversity evaluation, general calibration and replacement reconciliation, full WCAG/device/localization qualification, production model delivery/quality, native/offline scope and the remaining register. The current runtime image was not rebuilt/requalified for this increment. These changes complete neither the full academic milestones nor the full application.

## Reference-guided interface refinement — 2026-11-18

Passed: complete browser suite **10 passed (1.4m)** with the new light shell, task overview, assessment cards and sign-in design. The generation failure fixture now specifically exercises HTTP 503: plain-language guidance, technical-details disclosure and retry are visible; browser-back and mobile keyboard navigation still pass. Existing real OMR PDF/PNG correction/commit, online learner, numerical authoring and organization journeys passed. Strict TypeScript and the optimized production Next build passed. Final desktop (1600px), phone (390px) and dark-theme screenshots of Home, generation and sign-in were inspected; the final phone render has no document overflow. All overview counts derive from authorized API data.

External failure: one explicit live Gemini generation check using the existing private configuration timed out awaiting response headers. No successful live generation/tutoring or provider quality is established. The 503 UI test is an HTTP fixture and does not prove provider delivery.

No Go business rules, database schema or provider adapter changed in this visual increment; new Go/database qualification was not required or rerun. Full WCAG/screen-reader/device qualification and the remaining product register are still open. Screenshots are private local review artifacts under `.local/screenshots/reference-*`.

## Owner-supplied image and faculty review increment — 2026-11-18

Passed: Go unit/vet and strict TypeScript; full PostgreSQL/S3/SMTP transport integration **43.049s**; focused grid/text/key race tests **1.071s**; focused lint **0 issues**. Complete browser suite **11 passed (1.7m)**. The added journey rasterizes newly authored two-column questions, uploads the real image, edits wording and a faculty selection, uploads a real key image through Tesseract, confirms its preview, and verifies four persisted bank drafts. Its final narrow-screen/raw-header-preservation rerun passed **11.3s**. A prior harness attempt targeted a nonexistent sidebar item; it was corrected before the passing runs. Test suites now run sequentially to avoid shared Playwright artifact collisions.

Actual owner-supplied images were also exercised through authenticated application uploads, not a mocked response. The calibrated Go pixel reader matched **180/180** independent visually transcribed positions: **179 single fills, blank question 94**, with that blank flagged for confirmation. The registered-sheet baseline appropriately refused the unregistered image. Native paper OCR/order parsing produced **10 questions, 40 options**, with no solved-paper answers silently trusted. Raw OCR had wording defects (e.g. missing r in Artificial and a malformed more); two corrections were saved manually through the UI. Headers describe 50 questions, but only 1–10 are visible. Private reports/source hashes and visual overlays remain in `.local/image-evaluation`; user-provided images are not committed or assigned a redistribution license.

Integration tests establish explicit key application, learner denial, stale-version and foreign-import rejection, durable key extraction and bank persistence. Grid inspection rejects learners/cross-institution courses and creates no result revisions. Existing registered OMR/online workflows still pass. No field accuracy percentage, universal layout support, subjective OCR reliability or production qualification is claimed. This diagnostic does not commit arbitrary third-party sheets as academic evidence.

## Local OCR and reviewed mark models — 2026-11-18

Actual model evidence, independent of HTTP fixtures:

- An authenticated question-image upload ran through private object storage, the Go durable worker and revision-pinned **LightOnOCR-2-1B**. It produced **10 ordered editable questions and 40 options**. All 50 stem/option fields matched a prewritten visually checked reference after Unicode/case/punctuation normalization. Only questions 1–10 are visible despite the header's 50-question count. Faculty keys remained unset. The final run recorded **27,126 ms**, 2,157 input/676 output tokens; end-to-end polling wall observation was 27,558 ms on the local RTX 4060 Laptop 8 GB. These are individual sample observations, not load measurements or general OCR accuracy.
- The original CPU mark classifier matched all **720 independently checked bubble classes** on the supplied calibrated image. Its full authorized Go job agreed with the pixel reader at **180/180 response positions: 179 single fills and question 94 blank**. Direct classifier observation was 0.259 s; durable-job polling wall observation 2,099 ms. Blank 94 remains flagged for confirmation. No identity/grade/result was inferred. The actual browser displayed the overlay, restored the editable paper via its URL, and downloaded a report containing calibration, measured/model selections, local corrections and actual model/revision/runtime observations.
- Training data: 10,000 newly authored synthetic circle/digit patches; separate 2,000-patch synthetic holdout, 1,999 correct in the recorded run, seed 42017, 16 epochs. No supplied image or answer transcription was used for training. Weight digest and recipe are committed in inference/. This is limited synthetic validation, not calibrated confidence or field accuracy.

Failed candidate: actual **Qwen3.5-2B** OMR probes returned incorrect choices; larger five-row crops also produced wrong/incomplete structured answers. The first job was cancelled with its source/partial evidence retained. That adapter remains explicit experimental evaluation; the default is the task-specific classifier. No universal best-model claim is made. Native Tesseract's earlier wording errors and the registered reader's appropriate unknown-format refusal remain useful baselines.

Passed automated checks for this increment: strict TypeScript, Go vet/unit tests; actual PostgreSQL/private S3/SMTP integration suite (transport **37.199s**); Go race suites (academic **102.871s**, scholar **9.416s**); eight focused Go linters (**0 issues**); two Python runtime/artifact checks (**0.033s**). Model HTTP fixtures test wrong models, truncated output, busy/auth/cancellation and malformed/missing/invented OMR rows. Integration fixtures test durable OCR/mark handoff, private authorization, revision validation and retained recovery inputs. These fixtures establish orchestration contracts, not live-model quality.

A first browser regression run had two failures: the native document journey inherited the newly configured model default, and responsive navigation raced its pre-hydration layout. The native journey now selects Tesseract explicitly and passed its focused rerun (**11.5s**). Regression tracing also exposed native GET sign-in submission before hydration; the credential form now uses a disabled-until-hydrated fieldset and POST fallback. A JavaScript-disabled browser verified disabled credential controls and the POST method. A subsequent unfinished browser run was interrupted for that fix; it is not counted as passed.

Private evaluation sources, checked labels, job output, source SHA-256 hashes, response CSV and report are retained in `.local/image-evaluation/`; they are intentionally excluded from Git. The owner's files have no verified redistribution license. Original classifier weights/fixtures are CC0; third-party models retain their own licenses.

Unexecuted / incomplete: broader independent real-photo/printer/pen/erasures/calibration datasets, general layout detection and scored arbitrary-sheet reconciliation, inference GPU-container build/run, production deployment/load/failure qualification, full mobile/WCAG matrix, live successful Gemini generation/tutoring and the remaining 233-requirement register. Existing Gemini HTTP 503/timeouts are unresolved by local document OCR. No milestone or production readiness is declared complete.

The managed and standalone inference startup paths were checked for exclusion of application database/object/MFA/Gemini credentials; both passed. Standalone authenticated health/startup on a separate temporary port passed and that process was stopped. The primary local runtime remains running.

Final complete browser regression: **12 passed (1.8m)** after the explicit engine selection and hydration fixes. The final strict TypeScript/Go vet/unit gate also passed. Live local model UI checks and JavaScript-disabled authentication checks were separate real browser runs.

The final optimized Next.js production build passed after the browser regression and credential guard. The optional GPU inference container was not built or qualified.

## Educational processing feedback — 2026-11-18

The upload workspace now puts current processing above setup, collapses the original form after a saved upload, and shows an Upload → Extract → Review → Save sequence. Paper extraction, question generation, separate keys, registered sheet reading and calibrated/model inspection share a labelled spinner/progress panel. Unknown progress is indeterminate; only Go-reported stage percentages are displayed. Extraction failures offer direct retry, completion points to editable questions, and the upload screen has its own selected navigation entry. Background tasks refresh automatically. Generation status survives switching workspace sections, and old save notices clear on navigation.

Acceptance evidence includes a browser-controlled delayed upload and queued → running (35 reported stage percent) → failed → retry → editable-ready transition. This UI fixture verifies state rendering, recovery requests, step/navigation state, setup collapse, 390px overflow and reduced-motion behavior; it is explicitly **not** model inference evidence. It runs as a step in the real native question-image/separate-key import journey using the existing authenticated actor. This avoids exceeding the normal 12-attempt sign-in throttle; security limits were not weakened. Existing generation checks also assert indeterminate progress and retained status after navigation.

Initial checks caught a null job access during the new pre-job sheet-upload indicator; the guard was corrected. An independently signed-in thirteenth browser test hit the existing rate limit; it was combined with the authenticated import journey. Neither failing run is reported as passing. Strict TypeScript, Go vet and ordinary Go tests passed after the changes. New backend/database/provider qualification is not claimed by this frontend increment; no Go rules, schema or contracts changed. Full WCAG/screen-reader/device and production provider qualification remain open.

Final complete browser regression: **12 passed (1.9m)**, including the real paper/key import and registered OMR scoring journeys plus the delayed processing-state fixture. Light/dark and 390px processing screenshots were inspected; no horizontal document overflow and reduced-motion spinner behavior passed. The final small generation-state retention and upload-status focus changes are included in that run.

The final optimized production Next build passed. No new provider delivery or full accessibility certification is asserted by these UI checks.

## Minimal teaching desk — 2026-11-19

Passed: strict web TypeScript, production Next build, two Bun selection tests (nine assertions), and ten connected academic/import/usability/teaching-desk browser journeys (1.5m). Real Go/PostgreSQL/private-object/worker paths cover manual single-choice authoring and selected key, CSV range/list marking with fractional penalties and reload persistence, scanned two-column question editing, separate image answer key application, publishing/assignment/online work, PDF/PNG OMR review and authoritative marks. The student table agrees with the actual mixed sample: 5/12, correct questions 1,3,5, three correct/two wrong/one blank; name search also passes. Multiple uploaded OMR files preserve registered assignment matches and ambiguous-response review. A final targeted batch check includes an identical file twice to exercise source/job deduplication.

Empty extraction recovery uses an explicit HTTP fixture to reproduce zero rows; it is not OCR accuracy evidence. Generation-provider errors and model-comparison recovery remain explicit fixtures. Selection tests reject missing, malformed, descending and excessive ranges and deduplicate overlapping selections. Rendered home/light/dark/mobile and student-result screenshots were inspected; no mobile viewport overflow was found in the exercised journeys.

No new backend, migration or contract is introduced. Go retains validation, scoped access, immutable published revisions and score calculation. Range writes reuse existing versioned APIs sequentially; interruption can leave an already saved subset, retained for retry. A bulk change to bank questions produces new drafts requiring approval. Registered OMR templates/encoded identity are required; arbitrary handwritten-name reading and universal sheet recognition are not implemented. Live AI quality, real-photo/handwriting accuracy, full accessibility/role/recovery acceptance and production container qualification were not established by this increment. Requirement states remain partial.


## Reference dashboard / Taste audit — 2026-11-19

GATE: FAIL 21 (round 1 of 3). The skill's fresh unauthenticated browser follows /workspace to the existing sign-in page. The generic marketing audit therefore checks sign-in rather than the dashboard: hero-fill/screen, CTA, copy-headline, heading-scale, moment-none, section tile/height/gaps/edges, text-only, focus-ring and the development Next portal focus finding remain open in that report. The layout checker additionally reports `no Chrome found (set CHROME_PATH)`. No checker files were changed, findings waived, or authenticated-access checks bypassed. These results are not presented as a passed dashboard gate. Marketing-only hero/deal/effects/mock-fragment requirements yield to the owner's explicit working-dashboard reference and no-mock requirement.

Independent authenticated application verification: strict TypeScript and production build passed. New dashboard browser coverage asserts API-backed paper/topic counts, real title/reference search, opening a saved assessment, navigation to OMR, no viewport overflow at 1440/1024/768/390, and dark rendering. Manual/range authoring, empty-extraction recovery and multi-file OMR tests passed in the same four-test run (36.2s). The subsequent dashboard/usability run passed all six journeys (46.2s), checking the final compact layout plus existing generation, numerical authoring, OMR ambiguity/recovery and result-table journeys. Screenshots remain in ignored .local/screenshots/dashboard-reference-*.png; audit artifacts remain in ignored .taste-skill. No credentials or local records are committed. No live-model quality or complete accessibility certification is claimed.
