# Core education increment

This increment delivers usable question/key intake, reviewed authoring, pinned papers, online assessment, registered-sheet OMR review and marking. It also implements configured topic generation and contextual tutoring through Gemini or Anthropic. It does **not** complete milestones 2–5 or the full product contract. Full requirements remain in IMPLEMENTATION-STATUS.md.

## Run and exercise

Run `bun install --frozen-lockfile` then `bun run dev`. PostgreSQL, private S3-compatible storage and Mailpit start through Compose. The operator applies application and River migrations before API/worker startup. Synthetic owner credentials are in the protected, ignored `.local/demo-access.txt`; no credentials are committed.

Open **Home** in the workspace. Use the direct teaching navigation:

1. Create a course/enrollment using the organization tools, then create a topic with a concrete outcome. Instructors see courses/topics within their current teaching responsibilities.
2. Author a question or upload CSV/XLSX/JSON. The original example is `apps/web/public/demo/questions.csv` (CC0-1.0). The first XLSX worksheet is read; formulas are never executed. Type/key aliases are normalized. Structured imports are limited to 8 MiB and 500 rows.
3. Review persistent row errors, correct wording/key/numeric rules/rubric, and commit clean rows. Remaining invalid rows stay editable. Committed rows retain their source evidence; later changes create new bank revisions. Approve the drafts after checking content, solution and rights.
4. Select approved questions, move them into the intended order, compose the paper, review the actual pinned contents/key and publish. Published content has a database immutability guard. Printable booklets and staff marking keys are authorized outputs.
5. Assign a published paper to eligible learners with an opening/closing window and approved extra minutes. Assignments freeze learner identity rather than expanding when a roster changes. Repeated assignment retains the original window. A learner signs in, opens **My assessments**, starts/resumes, answers, saves and submits.
6. For paper work, select the published assessment under **Scan & check**, create a one/two-column layout, select an assigned learner and open their printable sheet. Print the matching booklet separately. Enter the actual test time and upload PDF/PNG/JPEG. The upload group is prepared automatically and review updates as the worker finishes. Earlier groups remain available under the resume disclosure.
7. Compare each page with its reading; confirm the assigned learner. Explicitly correct/confirm ambiguous, missing and written/numerical responses, including blank or invalid states. A campus coordinator/owner confirms whole-sheet identity and commits. The ordinary marking engine creates the sitting/result once. Rubric review and result release are in **Results & review** in the teaching navigation.

`bun run test:e2e tests/browser/academic.spec.ts` creates separate synthetic owner/learner accounts, a course and active enrollment, then exercises the teacher → OMR → learner-online journey. Temporary credentials in `.local/academic-browser-access.json` are generated afresh and must remain private/ignored. This is a test command, not a production enrollment bypass.

## Document transcription

PDF/PNG/JPEG question documents are stored privately and processed asynchronously by Poppler/Tesseract. PDFs are limited to 50 MiB/50 pages, rendered at a bounded 2000-pixel dimension; images are limited to 32 megapixels. OCR preserves the raw extracted text and original file. Its text mapping recognizes numbered questions, `A. option` lines, `Answer: A` / `Answer: A|C`, and `Solution: ...`. Teachers compare, correct, map, then review individual rows. Missing keys remain invalid. No OCR text is treated as an approved answer key automatically.

DOCX, separate-key document matching, figure/equation/table reconstruction, page/crop fidelity and handwriting recognition are **not** delivered by this path. Reprocessing page-level OCR without repeating completed pages is still required. A failed extraction keeps the original source/job and can be retried; a reviewed structured transcription can be uploaded independently.

## Marking and exam continuity

Marks use exact rational decimal arithmetic and integer thousandths, never binary floating-point totals. Selection requires exact sets. Numerical acceptance supports signed exact decimals, accepted values, inclusive ranges and absolute/relative tolerances; nonfinite, exponent and ambiguous locale input are rejected. Written responses remain pending until a staff member approves criterion marks within the pinned rubric. Every counted award/penalty has an explanation.

Optional groups support all / first-N / best-N with deterministic ties and caps. Optional groups currently require equal item marks/penalties to keep denominators unambiguous. Online first-N uses server-assigned first-answer order; paper first-N uses printed order. Optional uncounted blanks never become wrong answers or topic weakness evidence. Maximum-score fixtures include 18×3 + 8×2 = 70.

Saves have an operation reference and expected answer version. Replayed saves are acknowledged without another version; stale writes require explicit comparison/merge. One authenticated session owns the online writer; **Resume here** is an explicit learner takeover. Scheduled River work enforces the deadline while the browser is closed. Save/submit after expiry retains acknowledged answers and does not reopen the sitting. Submission is repeatable and produces one initial result revision.

The browser retains pending work in memory during network failure. It retries the same operation payload after a lost acknowledgment; it does **not** promise protected offline persistence through refresh. The comparison UI lets the learner select local or current saved responses before retrying. Full refresh/offline reconciliation and supervised staff takeover remain SIT-05/SIT-06 work.

Immediate release is formative only. Staff release waits until the assignment window closes and subjective marking is complete. Exposure is recorded and blocks subsequent formal assignment of the same question family. Immediate release is held when the learner already has a current formal assignment containing the same question family. Live answers, keys, numerical rules, rubrics, solutions and authoring metadata are stripped from learner payloads. Tutor access rechecks ownership, eligibility and release every time.

## OMR coverage and evidence

The implemented format is **registered-bubbles-v1**, one column up to 40 items or two columns up to 80, 2–8 A–H options and single/multiple selection. Each sheet contains four registration squares and a machine-readable assignment/layout identity grid. This is an original format, not a universal decoder for arbitrary vendors' OMR sheets.

Recognition locates filled registration components, tries four orientations, estimates a projective transform and samples inner bubble fill independently of the answer key. Fill >=0.65 is filled; >0.12 and <0.65 needs review; lower fill is blank. Multiple fills on a single-selection item need review. These are transparent starting thresholds, **not validated probabilities**. Unreadable identity/layout or missing registration blocks automatic interpretation. Numeric/written areas require explicit transcription. The original normalized scan, machine fill measurements, registration corners, corrections, reason, actor, time and identity changes are retained. Duplicate sources/pages and sitting uniqueness prevent repeated evidence.

The evaluator uses newly authored synthetic images: clean fills, blanks, faint/partial fills, wrong-layout identity, 90° rotation, perspective and changed exposure across both layout arrangements. PostgreSQL/S3 integration runs pixels → queued worker → review → committed marks. This dataset does not establish accuracy on photographs, creases, erasures, handwriting, printer/scanner diversity or all page counts. There is no claimed field accuracy rate. Calibration UI, overlays/crops, resumable upload, subject-delegated review, rejection/quarantine, replacement-scan reconciliation and retention/hold controls remain required.

## Models and jobs

The manual test environment uses Gemini, configured privately in the ignored `.env`. The user-selected `gemini-flash-latest` alias is explicit; actual response `modelVersion` is recorded so output is attributable even when the alias changes. Set on the worker:

```
ACADEMIC_MODEL_PROVIDER=gemini
GEMINI_API_KEY=<private worker secret>
GEMINI_MODEL=gemini-flash-latest
```

Gemini uses `X-goog-api-key` header authentication and `generateContent` with JSON Schema output for draft questions. Truncated, blocked, malformed and invalid output never becomes approved material. The same private released-context tutor boundary applies. Missing credentials/model access are configuration failures; rate/quota and availability failures remain recoverable jobs. `go run ./cmd/operator model-check` makes one real generation and one tutor request using synthetic material, saving observations privately to `.local/model-check.json` only if both succeed. It is an explicit live check and may consume provider quota. Ordinary automated tests use a local HTTP contract server, never silently spend live quota.

For the alternate Anthropic adapter, set these on the worker:

```
ACADEMIC_MODEL_PROVIDER=anthropic
ANTHROPIC_API_KEY=<private worker secret>
ANTHROPIC_MODEL=<supported model identifier selected by the institution>
```

The actual adapter calls the official Messages endpoint using API version `2026-10-21`. Topic generation sends curriculum fields, never learner records, and requires structured `submit_questions` output. Every generated item passes the Go validator and remains a draft until teacher approval. Model output and measured token usage (where returned) are retained; absent usage is unknown. Provider content/error bodies and secrets are not copied into user errors.

The tutor receives only the selected released question, pinned solution, this learner's response and this private conversation. Replies are labelled AI guidance. It has no document search or cross-institution retrieval. It does not diagnose emotion or claim mastery/readiness. Missing credentials show **configuration_required** and preserve the request for retry. Network/provider failure does not erase assessment results or scanned evidence. Live provider delivery/quality has not been qualified without configured credentials.

For explicit deterministic local generation only:

```
ACADEMIC_MODEL_PROVIDER=development_fixture
```

This is forbidden outside development/test. It produces visibly labelled original arithmetic fixtures, supports mathematics only, and does not prove topic grounding or AI quality. It provides no fabricated tutor replies. Unconfigured operation is the default; deterministic success is never silently substituted for the real provider.

River v0.51.2 supplies PostgreSQL transactional enqueue, concurrent workers, scheduling, retry, cancellation and abandoned-background tasks. `academic_job` exposes input/progress/attempts/provider/output/error through authorized operations UI. Model/scan/document jobs attempt at most three automatic executions; expiry jobs at most ten. Retry reuses the River job rather than creating competing executions. Completed model output is persisted before draft materialization so downstream retries can reuse it. Provider success followed by an unrecoverable network failure can still incur an unknown external charge; provider-wide idempotency/complete cost accounting is not claimed.

After fixing configuration or quota, retry in **Background tasks** (or the learner's tutor panel). Cancellation retains existing evidence. Run migrations before upgrading/starting workers. Back up River's schema along with academic tables. General worker-crash drills, complete budgets and instruction-management UI remain required. Document processing executables are installed in the Go runtime container; local hosts need `pdfinfo`, `pdftoppm` and Tesseract language packs.

## Evidence and remaining scope

Released topic summaries are calculated from the current result revision, with counted/answered/correct/wrong/required-blank/pending counts and an insufficient-evidence state. They do not invent timing, emotion, ranks, readiness or mastery. Staff can inspect retained marking revisions. Full longitudinal learning, revision journal, targeted lessons, interventions, adaptive practice and paper-key correction/appeals remain outstanding.

Private document assets currently retain evidence rather than applying automated retention deletion. Failed transaction uploads can leave unreachable private object artifacts; an academic orphan/retention reconciliation workflow is required before production qualification. No public object URL is returned. Account/avatar lifecycle remains separately implemented. Full file malware quarantine, complete delegated subject-review assignment and production storage/lifecycle qualification remain RUN-03/PPR-12 work.

Native/mobile offline study, family academic reports, finance, collaboration, media, forecasts and the remaining platform requirements are still required by PRODUCT-SPEC.md. No milestone or production release is declared complete by this increment.

## Manual accounts

With the local environment loaded, run `go run ./cmd/operator manual-demo`. It requires development mode and creates an independent **Manual Learning Institute** with two campuses, seven actors, active enrollments, subject responsibilities, family links, four original subject banks and three published practice papers. Random per-account credentials and a role-by-role walkthrough are stored at `.local/manual-access.md` (mode 0600), with a private JSON manifest alongside it. The command is idempotent and preserves manual changes/passwords; it never silently resets an account. All seven accounts were checked through actual sign-in, identity and workspace requests. Guardian and reviewer academic workspaces remain incomplete; these accounts currently exercise account/security and permission boundaries.

Live Gemini observation (2026-11-18): the configured alias returned HTTP 503, then timed out on a bounded retry; a minimal request also timed out. Live generation/tutoring is therefore **not verified**. Re-run the explicit model check after provider availability recovers; the application retains failed job inputs and allows authorized retry. Do not infer a successful AI response from configuration or the local contract fixture.

## Original practice kit and simpler navigation — 2026-11-18

Start with **Home** for the workflow overview, or **Try sample files** for a complete prepared test. Staff choose an authorized course and enrolled learner; Go publishes and assigns a new practice paper and creates its registered layout. Download the authenticated ZIP, then follow **Go to Scan & check**. No result is created until the learner submits or staff commit reviewed evidence.

The six original CC0 questions cover arithmetic and introductory science. Each awards 2 marks, deducts 0.5 for a wrong response and awards 0 for blank; maximum 12. The ZIP includes the paper/key, importable CSV/JSON, expected results, matching blank/filled PDF/PNG sheets and a duplicate-page PDF. Correct = 12/12; mixed = 5/12 (three correct, two wrong, one blank). The uncertain example marks B+C on question 1; explicitly choose B before committing for 12/12. Use a fresh kit for each committed variant, since one assignment cannot acquire competing sittings. Do not upload another learner's or another kit's sheet.

Generic paper/key previews are under `/samples/`. Structured CSV/JSON imports include the key; PDF OCR is a reviewable transcription path and does not automatically reconcile a separately uploaded key. Public samples are prohibited in formal exams. Teacher edits to reused sample keys/marks are preserved and block creation of a misleading kit until resolved.

AI generation now keeps topic selection, progress, provider errors and recovery on one screen. The bank supports search/status filters; numerical acceptance and written rubrics use labelled fields. Scanning automatically prepares the upload group, polls worker progress and presents identity/correction/commit steps. Earlier groups and custom layouts remain available in disclosures. Direct hash navigation supports browser back and selected-paper refresh. Organization and administration tools are grouped separately; mobile navigation has a keyboard-operable Menu.

The latest live Gemini check again returned HTTP 503. The provider configuration is retained privately; neither the sample kit nor HTTP test fixtures prove live generation. Six required blanks retain an insufficient-evidence status rather than claiming observed responses. See TEST-REPORT.md for current evidence and remaining qualification.
