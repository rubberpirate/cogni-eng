# Learning and assessment platform — product and engineering specification

Version 1.0 · 2026-11-17

This document defines a complete new product. It is self-contained and is intended to travel with `BUILD-PROMPT.md` into an empty workspace. It does not depend on any source repository, existing brand, installed application, institutional database, sample paper, or proprietary asset.

Every numbered requirement below is in scope. Delivery is incremental; the final scope includes the advanced capabilities. “Unavailable,” “needs review,” “insufficient data,” and “provider not configured” are valid truthful states, but not substitutes for implementing the configured path.

## A. Purpose, audience, and product boundaries

Enable schools, tutoring organizations, and examination-preparation institutions to assess learners, understand learning gaps, deliver appropriate teaching/practice, and measure progress. Support entrance-exam preparation, school curricula, and institution-defined courses without embedding one exam's syllabus, paper format, or terminology into business rules.

The complete learning journey is:

**Enroll → establish a baseline → assess → mark → explain evidence → plan study → learn/practice → reassess → verify improvement.**

The institution's companion journey is:

**Organize people and curriculum → prepare/review material → deliver assessment → resolve exceptions → teach/intervene → communicate/report → manage access, finance, and operations.**

Business roles are permission templates, not separate identities. A person may teach one class, manage a campus, and be a guardian in another context. Keep user identity separate from institutional membership, learner enrollment, family relationship, and staff assignment.

### A.1 Roles and responsibility

| Responsibility | Default permitted view/actions |
|---|---|
| Learner | Own enrolled courses, assessments, released results, revision, lessons, tutor conversations, goals, and permitted class communication. |
| Guardian | Linked children, released progress/attendance, relevant staff communication, consent and notification choices. No assessment answering on a child's behalf. |
| Instructor | Assigned learners/classes, permitted cross-subject learner insight, teaching content, assessments within assigned subjects, grading/review tasks and intervention work. |
| Campus coordinator | Assigned campus enrollment, staff responsibilities, assessments, scan intake, family links, local reporting and authorized finance tasks. |
| Organization owner/academic lead | Institution-wide configuration, campuses, delegated permissions, curriculum, analytics, financial administration and audit. |
| Reviewer | Explicitly delegated paper, answer-key, written-script, content, or scan-review work. Scope may be limited by subject and campus. |
| Support/platform operator | Service health and approved support operations. No blanket academic-record access. Sensitive access requires an audited, expiring support grant. |

A user can hold multiple roles with explicit scopes. Financial, safeguarding, grading, and platform administration are separately grantable permissions. Self-registration never grants staff power.

### A.2 Product areas

Organize navigation around work rather than reproducing role-named portal trees. Suggested product areas are **My day, Assessments, Learning, People, Teaching studio, Conversations, Insights, Organization, Billing, and Service desk**. Final route names, package names, schemas, and layouts must be designed independently.

Provide role-appropriate starting screens, a clear institution/campus context switcher, global search scoped to entitlement, and deep links to permitted work. One underlying result must mean the same thing wherever it is displayed.

## B. Delivery contract

- Build with Go backend, Next.js/React TypeScript web, PostgreSQL, private object storage, durable Go workers, and native React Native/Expo clients as described in the build prompt.
- Use standards where helpful: OpenAPI contracts, accessible semantic HTML, calendar exports, CSV/XLSX/PDF reports, and signed provider callbacks.
- Implement one source of business truth. Browser/native clients collect interaction and render state; they do not decide eligibility, final scores, financial settlement, or staff scope.
- Define authoritative policy versions for marking, answer release, deadlines, learning evidence, moderation, privacy, and billing. A configurable field needs validation, documented effect, and an acceptance test.
- All asynchronous features need visible progress, cancellation semantics, failure explanation, retry/recovery, and an audit reference.
- Required external integrations need at least one actual documented adapter and deterministic local test implementations. The test implementation is visibly identified and prohibited from claiming live financial or communication outcomes.
- Preserve a complete requirement register. Broad entries below must expand into their detailed subchecks during implementation.

## C. Functional requirements

### 1. Identity, sign-in, recovery, and sessions

- **ID-01 — Registration and invitations.** Support learner registration and staff/guardian invitations, institution enrollment codes, verified email, optional phone verification, and explicit approval where the institution requires it. Collect only needed personal/academic information; record terms/privacy acknowledgments with policy version and time.
- **ID-02 — Credential sign-in.** Normalize email, validate inputs, offer password visibility and clear accessible validation, support persistent versus short sessions, and route the user to the correct permitted workspace. Do not reveal account existence through recovery responses.
- **ID-03 — Password recovery/change.** Provide expiring single-use reset links/codes, authenticated password change, verified email-change confirmation, and relevant security notifications. Revoke sessions according to the recovery policy.
- **ID-04 — MFA.** Implement authenticator-based MFA, recovery codes, enrollment confirmation, lost-device recovery, and organization-level enforcement for privileged staff. Offer email/SMS factors only when their provider and policy are configured.
- **ID-05 — Session management.** Show current and other signed-in devices/sessions, last activity and approximate session context; revoke one or all. Rotate/reissue safely and expire idle/absolute sessions. Logout invalidates the applicable server session.
- **ID-06 — Account state.** Support invited, active, suspended, and archived access states, with reason, actor, time, and recovery workflow. Suspension immediately invalidates access while preserving lawful academic/financial history.
- **ID-07 — Abuse controls.** Rate-limit sign-in, recovery, registration, verification, invitations, uploads, and paid interactions across all application instances. Avoid locking an entire institution behind a shared network address because one person makes mistakes.
- **ID-08 — Profile.** Edit permitted name/contact/academic fields and a cropped profile image; retain verified identifiers and staff-controlled fields. Support emergency contacts and user-facing account activity history.
- **ID-09 — Safe bootstrap.** Create the first organization owner through a secure one-time operator process. Development fixtures must be synthetic and unusable as production credentials. Provide weak/compromised credential remediation.

**Acceptance:** invite an instructor, enable MFA, recover access, revoke an active session, suspend/re-enable the account, and verify old sessions never regain access. A learner cannot obtain staff privileges by modifying a request.

### 2. Institutions, campuses, permissions, and enrollment

- **ORG-01 — Institution lifecycle.** Create, configure, activate, suspend, and close organizations with branding, timezone, locale, support contacts, academic-year settings, and subscription entitlements. Separate organization closure from deletion of retained records.
- **ORG-02 — Campus management.** Maintain named campuses with address, codes, contacts, responsible staff and active dates. Support staff responsible for more than one campus, institution-wide oversight, and scoped reassignment with audit.
- **ORG-03 — Courses and groups.** Manage programs, academic years, grades, courses, cohorts, class groups, sections, and teaching timetables. A batch date is not a replacement for managed class membership.
- **ORG-04 — Learner admission.** Enroll singly or through validated CSV/XLSX import, with preview, duplicates, row errors, eligibility/grade, target exams, campus, cohort, start/end dates, and admission state. Support an intake queue awaiting placement.
- **ORG-05 — Human references.** Allocate unique readable learner and paper reference codes, allow appropriate externally issued identifiers, preserve aliases after deliberate corrections, prevent reuse/races, and support manual lookup. Do not encode mutable campus or role state into permanent identity.
- **ORG-06 — Transfers and withdrawal.** Transfer campus/class/course, preserve academic history and identifier continuity, update affected staff scope, and record effective dates. Withdrawal ends access/assignments according to policy without destroying evidence.
- **ORG-07 — Guardian relationships.** Invite and verify guardians, link multiple children and multiple guardians, distinguish relationship/permission level, approve/remove links, and show an unambiguous child switcher. Handle custody/contact restrictions through explicit institutional controls.
- **ORG-08 — Staff and teaching assignments.** Create/edit staff professional profiles, subjects, qualifications/experience, campus/class responsibilities and effective periods. Replace/add/remove teaching rosters through searchable selectors without losing selections across pagination.
- **ORG-09 — Permission administration.** Provide role templates, a capability matrix, custom scoped roles, grant/revoke workflows, protected owner transfer, and bulk changes with preview. Prevent self-escalation and orphaned ownership.
- **ORG-10 — Import/reconciliation tools.** Reconcile duplicate people and conflicting external IDs without silently merging unrelated records. Offer dry-run reports, explicit resolution, reversible administrative corrections, and an audit trail.

**Acceptance:** create two institutions with two campuses each and a person with several legitimate memberships. Verify every read, write, export, file, search, notification, and background operation stays within the selected authorized context.

### 3. Academic catalog and content bank

- **CAT-01 — Curriculum models.** Maintain versioned curricula, exams, subjects, chapters, topics/subtopics, outcomes, prerequisite relationships, language and academic level. Seed original representative mathematics, physics, chemistry, and biology content; allow institution-defined subjects.
- **CAT-02 — Question types.** Support single selection, multiple selection, integer/decimal numerical, short written response, and extended written response. Provide rubric-based manual marking for written types and extensible type contracts without inventing a general plugin system.
- **CAT-03 — Rich question content.** Author readable text, mathematical/chemical notation, diagrams, tables, attachments, image-only stems, option images, alt text, and worked solutions. Rendering must agree in authoring preview, online/native player, paper export, review, and tutor context.
- **CAT-04 — Academic metadata.** Track intended outcome/topic, difficulty, source/license/provenance, language, estimated time, tags, default marks, penalties, numeric acceptance and accessibility alternatives. Permit weighted multi-topic attribution when explicitly authored.
- **CAT-05 — Lifecycle and revision.** Support draft, review, approved, retired and archived content with history, reviewer comments, comparison and restoration as a new revision. A correction cannot rewrite an already delivered question snapshot.
- **CAT-06 — Bank management.** Search/filter/preview, bulk tag/move/archive, edit, import/export selected records, inspect use history, and report deficient topic/difficulty coverage. Protect questions referenced by live or past assessments from destructive deletion.
- **CAT-07 — Duplicate management.** Detect exact and likely duplicate text/images with explainable matches; provide human merge/keep decisions that preserve references and original authorship. Do not silently replace similar scientific questions.
- **CAT-08 — Bulk authoring/import.** Support structured JSON, CSV, and XLSX import with type/answer alias normalization, image bundles, validation preview, row errors, repeat-import safety, and explicit handling of unknown answers. Incomplete questions cannot be published into scored work.
- **CAT-09 — Review and provenance.** Distinguish generated, imported, authored, and externally licensed content. Require staff approval before AI-generated questions/solutions become scored material. Record source attribution and permitted reuse.
- **CAT-10 — Quality and item analysis.** Show total/draft/approved/retired counts, topic/difficulty/type distribution, bank depth, freshness, exposure, observed facility/difficulty, distractor selection, discrimination where statistically supported, and flagged quality concerns. Provide evidence counts and avoid treating sparse statistics as authoritative.

**Acceptance:** create a multi-answer question with image options, revise its solution after use, and verify an earlier sitting remains reproducible. Repeat an import without duplicating questions or files.

### 4. Document digitization and publication

- **DOC-01 — Source intake.** Accept image photographs, scanned/text PDFs, and DOCX papers, plus separate answer-key files and worked-solution documents. Validate content type, size, page limits, encryption, supported languages and readability.
- **DOC-02 — Durable extraction workspace.** Create a persistent import job with source previews, page-level progress, recoverable drafts, cancellation and resume. Closing the browser must not lose the import or its reviewed edits.
- **DOC-03 — Faithful extraction.** Recover complete stems, options, figures, option images, tables, equations, printed question numbers, section instructions, and supplied solutions. Preserve provenance/crop/page coordinates; flag missing or uncertain content instead of inventing it.
- **DOC-04 — Key matching.** Extract separate or embedded keys and match by paper revision, section and printed number. Report unmatched, duplicated, conflicting, multi-answer, numerical and withdrawn-question entries for review.
- **DOC-05 — Reviewer tools.** Edit wording/figures/options, keys and rubric, assign topic/difficulty, apply tags/rules to ranges, compare original and draft, and approve pages/questions. Support keyboard navigation and autosaved reviewer changes.
- **DOC-06 — Completeness checks.** Reconcile declared versus extracted question counts, section boundaries, numbering gaps, cross-page continuations, repeated pages and option completeness. Partial success remains usable but cannot be silently advertised as complete.
- **DOC-07 — Publication.** Validate every mandatory topic, key/rubric and scoring instruction before publishing a versioned assessment or selected bank entries. Preserve the exact intended order and retain the reviewed source record.
- **DOC-08 — Retry and cost.** Retry only failed extraction work without discarding reviewed corrections or paying again for completed work unnecessarily. Show consumption and useful quota/configuration failures to authorized operators.

**Acceptance:** import an original multi-page paper containing an image-only question and separate key, close/reopen, fix one extraction error, publish, and verify paper order and figures in the delivered assessment.

### 5. Assessment design, schedule, and distribution

- **ASM-01 — Assessment purposes.** Support baseline diagnostics, subject quizzes, practice, mock exams, formal exams, homework and adaptive practice. Keep purpose separate from delivery channel and grading/release policy.
- **ASM-02 — Composition.** Select explicit questions in order or generate from an approved blueprint using subjects/topics, count, difficulty, coverage, expected duration and exposure constraints. Preview the actual paper, not just a claimed count.
- **ASM-03 — Sections and scoring.** Configure sections, marks/penalties, maximum contribution, optional groups, numerical rules, partial-credit policies and rubrics. Derive total maximum marks from the selected rules and validate impossible configurations.
- **ASM-04 — Lifecycle.** Draft, review, publish, schedule, open, close and archive; duplicate a paper as a new revision; distinguish corrections from new papers. Pin a published revision for each assigned sitting.
- **ASM-05 — Eligibility/recipients.** Assign individually or by course/class/cohort, with searchable bulk selection, exclusions, due dates, windows, timezones, attempt counts and prerequisites. Freeze or explicitly manage later group-membership changes.
- **ASM-06 — Accommodations.** Set extra time, approved breaks, accessible presentation, alternative delivery, conduct exemptions, and individual scheduling. Approved accommodations affect the actual authoritative policy and are not misconduct.
- **ASM-07 — Delivery configuration.** Choose online, paper, or mixed delivery, fixed order versus seeded randomization, section timing, permitted calculator/reference material, and conduct policy. Prevent inconsistent settings between instructions and player behavior.
- **ASM-08 — Reuse and extra assignment.** Add eligible recipients to a published paper without duplicates; support reassignment/retakes under explicit version and result-selection policy. Preserve who assigned what and when.
- **ASM-09 — Staff monitoring.** List/search papers and show assigned/not started/live/submitted/awaiting review/released counts, deadlines, aggregate results, per-recipient state and actionable exceptions. Respect live-answer visibility policy.
- **ASM-10 — Paper output.** Export printable question booklets, marking instructions, approved answer keys and compatible response sheets, with clear version/reference identity and accessible alternatives.
- **ASM-11 — Publishing checks.** Detect missing keys, insufficient pool, duplicate questions, missing figure assets, invalid optional counts, unreconciled section totals, and broken schedules before delivery.

**Acceptance:** publish a mixed-type paper, assign two classes and one learner with extra time, then verify each recipient receives the correct pinned version, window, deadline and maximum score.

### 6. Online and mobile examination experience

- **SIT-01 — Eligible start.** Show instructions, applicable accessibility/conduct requirements, technical readiness and release policy. Start only an eligible assigned/allowed paper within its window; resume the authorized sitting rather than creating accidental duplicates.
- **SIT-02 — Player.** Provide readable rich questions, figures/options, numerical/written answer inputs, section context, marks/penalties, approved tools, zoom, and accessible navigation. Save answer, clear, move back/forward, jump, bookmark/mark for review and filter the palette.
- **SIT-03 — Palette and progress.** Distinguish unseen, visited blank, answered, marked for review, and answered-and-marked states. Show section/optional-group completion, time remaining and meaningful approaching-deadline warnings.
- **SIT-04 — Authoritative time.** Establish deadlines from server time and approved allowances, independent of client clock and browser presence. Recover countdown after refresh; expire abandoned sittings through durable scheduled work. Define authorized pause/break and window-close behavior.
- **SIT-05 — Durable saving.** Autosave versioned responses, display saved/saving/offline/failed status, retry safely, and reconcile acknowledgments. Keep pending changes through refresh where permitted; never label a lost update saved.
- **SIT-06 — Reconnection and concurrent clients.** Resume after connection loss, reconcile local/server edits, reject stale writes, and detect a second active device/session. Adopt one active exam writer by default with an explicit supervised takeover flow.
- **SIT-07 — Timing evidence.** Accumulate per-item active viewing/answer interaction with return visits and evidence quality. Exclude background/inactive time appropriately, retain uncertainty after outages, and treat client timing as an observation rather than trusted proof of ability or misconduct.
- **SIT-08 — Submission.** Confirm voluntary submission with unanswered/optional-section summary; safely finalize the latest acknowledged responses and explicitly handle pending unsent work. Submission, deadline expiry and conduct submission converge on one idempotent finalization outcome.
- **SIT-09 — Final state.** Submitted content cannot change through answer-save routes. Offer a receipt, effective finish time, recorded reason, score/review status and reconnectable status lookup when the submit response is lost.
- **SIT-10 — Accessibility and continuity.** Support keyboard/screen-reader operation, large text, narrow/landscape layouts and approved accommodations without trapping the learner in fullscreen or inaccessible controls.

**Acceptance:** answer and mark items, disconnect, refresh, reconnect, and submit while an expiry worker runs concurrently. Exactly one final sitting and one marking record exist, with a visible account of any unsent answers.

### 7. Assessment conduct and integrity

- **INT-01 — Configurable conduct modes.** Provide open practice, monitored browser, and institution-approved supervised delivery policies. Explain exactly what is observed and what consequence follows; obtain required acknowledgments.
- **INT-02 — Browser observations.** Observe fullscreen exit, focus/visibility changes, supported display information, copy/paste/context-menu restrictions and observable screenshot shortcuts when the selected policy enables them. Detect capabilities and document unsupported cases.
- **INT-03 — Events and deduplication.** Retain event category, observed/server times, device context, actor and policy version; deduplicate related browser events and network replays. Provide an accessible warning and remaining allowance.
- **INT-04 — Escalation.** Apply configured warning, instructor review, temporary hold or finalization consequences on the server. Support staff adjudication, justified overrides and appeals, with retained evidence.
- **INT-05 — Answer secrecy.** Prevent unreleased answers/solutions from appearing in assessment payloads, search, tutor history, content suggestions, attachments, notifications, caches and exports. Exposure history must affect future formal-paper selection or access policy.
- **INT-06 — Security versus monitoring.** Never claim the browser can prevent OS recording, external devices or collaboration. Webcam/biometric collection is not required by default. Any later collection needs a separately approved policy and implementation.

**Acceptance:** replay one observed focus event, use an approved accommodation, and attempt to retrieve a live question's solution through several entry points. Only the real deduplicated policy event is recorded and answer access remains denied.

### 8. Marking, moderation, release, and appeals

- **GRD-01 — Objective correctness.** Determine single selection, exact-set multiple selection and numerical correctness from the approved immutable key/rule version. Support defined all-or-nothing and explicitly authored partial-credit policies; distinguish blank, invalid and answered wrong.
- **GRD-02 — Numerical acceptance.** Support integer/decimal input, signs, bounds, accepted values/ranges, absolute/relative tolerances, units where authored, and explicit rounding/tie policy. Reject NaN/infinity, malformed input, and ambiguous locale formats with useful messages.
- **GRD-03 — Optional groups.** Support count-all, first-N-attempted and best-N selection. Record exactly which answers contribute, why, and the resulting maximum. For offline first-N, use published printed-order selection because chronological answer order is unobservable; do not claim to reconstruct it.
- **GRD-04 — Special cases.** Support cancelled/withdrawn questions, corrected keys, bonus/full-credit policies, absent/excused sittings and section caps. Show the applied policy and its effect rather than silently altering totals.
- **GRD-05 — Subjective marking.** Provide versioned rubrics, criterion/step marks, annotated script review, marker assignment, moderation, adjudication, double marking when required and reviewer workload. AI suggestions must preserve the script evidence and distinguish proposal from approved marks.
- **GRD-06 — Immediate outcomes.** Return objective score/counts promptly after finalization even if commentary/media is unavailable. Separately expose pending subjective review and provisional/released results; do not imply a provisional score is final.
- **GRD-07 — Release policy.** Configure when scores, rank, item correctness, answer keys, full solutions and tutor access become available. Default formal exams to cohort closure; practice can release immediately. Guardians receive only released child information.
- **GRD-08 — Learner review.** Once released, show every permitted question, answer, correctness, marks, worked solution, reason/uncertainty and observed time quality; distinguish optional blank, mandatory blank and invalid mark. Include subject/section breakdowns and offline/online provenance.
- **GRD-09 — Appeals.** Allow a learner/instructor to flag a problem and submit evidence, route it to an authorized reviewer, track acknowledgment/resolution and notify affected people. Preserve the original submitted answer.
- **GRD-10 — Regrading.** Preview affected sittings and score changes, require an authorized correction reason, publish a new marking revision, and recompute results, aggregates, learning indicators, rankings and applicable practice plans consistently. Keep old/new results and notifications auditable.
- **GRD-11 — Rankings.** Calculate cohort ranks/percentiles only for comparable released papers/sitting policies. Define ties, eligible population and retake selection; do not compare raw scores from different maximums or different personalized papers as if equivalent.

**Acceptance:** mark a paper with optional questions, correct a key after release, approve a regrade, and verify the full downstream history updates once while the original result remains inspectable.

### 9. Paper response sheets and batch review

- **PPR-01 — Layout studio.** Create/version response-sheet layouts with learner/paper identity areas, response kinds, section/question mapping, registration/orientation marks and machine-readable template identity. Preview printable output and validate calibration against the generated blank sheet.
- **PPR-02 — Format flexibility.** Support different counts, sections, option labels, multi-select groups and numeric write-ins, including signed/decimal or four-plus-digit values when the selected format permits. Never apply one template's assumptions to another sheet silently.
- **PPR-03 — Intake session.** Create a campus/paper-bound batch with the actual exam date/time, template version, upload permissions and reviewers. Accept images and multi-page PDFs through resumable asynchronous intake.
- **PPR-04 — Recognition and validation.** Normalize perspective/orientation/exposure where supported, match layout identity, read references and responses, and validate learner eligibility/paper association. Uncertain identity or mismatched paper must block commitment.
- **PPR-05 — Confidence and uncertainty.** Identify blanks, multiple marks, ambiguous fills and uncertain handwriting. Keep transcription independent of the answer key; retain confidence/evidence and disagreements rather than replacing them with guessed certainty.
- **PPR-06 — Review console.** Prioritize uncertain sheets/items, filter by status/subject/reviewer, display original/aligned/cropped evidence and overlays, and support keyboard-first correction. Allow explicit blank, invalid/wrong, corrected value, identity correction, flag acknowledgment with reason, and reject/quarantine.
- **PPR-07 — Reviewer scope.** Assign subject-limited reviews without exposing unrelated private material unnecessarily. A designated coordinator approves whole-sheet identity and final commit. Detect conflicting simultaneous reviewer edits.
- **PPR-08 — Evidence audit.** Preserve machine readings, original scans, confidence/geometry, reviewer changes, reasons, actor/time and template revision. A re-decode can propose updated reads but cannot erase approved corrections; compare/reconcile conflicts visibly.
- **PPR-09 — Duplicate safety.** Deduplicate repeated files/pages, repeated uploads across requests, and repeated learner/paper commit races. Offer an explicit corrected/replacement scan workflow rather than generating a second academic result.
- **PPR-10 — Commit.** Commit eligible reviewed sheets once; allow selected clean sheets while other sheets remain in review, making remaining work explicit. Each committed sheet creates the ordinary marking/learning/reporting flow, with actual sitting time and no invented pace.
- **PPR-11 — Failure recovery.** Isolate bad pages, restart failed recognition, handle quota/provider failure as manual-review work, retain partial progress, and report rejected/failed/committed counts. Batch completion must not mean every sheet succeeded.
- **PPR-12 — Private access and retention.** Protect source scans/signatures and derived crops through entitlement checks, expiring access and audit. Apply configured retention/hold rules while retaining required marking evidence.

**Acceptance:** process several new layouts, upload the same multi-page PDF twice, correct an ambiguous answer, re-run recognition, commit concurrently and verify one result per accepted sitting with corrections preserved.

### 10. Written answer scripts and assisted evaluation

- **WRT-01 — Script intake.** Upload scanned written scripts or digital response attachments, associate them with the right sitting, map pages to questions and detect missing/duplicate pages and unreadable material.
- **WRT-02 — Transcription review.** Extract handwriting/text/diagrams where supported, retain page coordinates and uncertainty, and allow corrections without overwriting the original evidence.
- **WRT-03 — Rubric assistance.** Propose criterion-level marks and explanations grounded in the approved rubric and visible response. Show quoted/cropped supporting evidence and uncertainty; never manufacture student working.
- **WRT-04 — Human approval.** Allow manual marking when transcription/AI fails, require review under the institution's policy, support moderation and disagreement resolution, and finalize through the ordinary marking/release process.
- **WRT-05 — Student feedback.** Release approved annotations, criteria, improvement suggestions and appeals with private document access. Keep rubric/key secrecy until the result-release policy allows it.
- **WRT-06 — Quality evaluation.** Measure agreement with qualified reviewers on representative original/licensed scripts, report per-criterion disagreement, and support switching to mandatory human review when performance is inadequate.

**Acceptance:** an uncertain handwritten derivation is manually corrected, AI assistance is compared with a rubric, a reviewer adjusts marks, and only the approved result is released.

### 11. Learning evidence, progress, and readiness

- **EVD-01 — Explainable error observations.** Distinguish observed incorrect/blank/invalid answers from suggested causes such as concept gaps, rushed work, time shortage, or uncertainty. Show the evidence and confidence; retain unknown-cause states.
- **EVD-02 — Consistent measures.** Define accuracy as correct among genuinely answered/scored responses, attempt coverage among required opportunities, and score percentage from applicable maximum. Publish alternative weighting explicitly. Do not silently include optional blanks or unmeasured time in denominators.
- **EVD-03 — Topic evidence.** Accumulate per-topic/outcome observations over time with evidence count, recency, question quality, mode and revision provenance. Insufficient evidence is a first-class state. Apply multi-topic weights without double counting.
- **EVD-04 — Mastery transitions.** Use a documented, versioned, reproducible policy for emerging/developing/secure understanding and regression. Require adequate diverse evidence for promotion; expose history and allow review. Choose thresholds/calibration independently and validate changes before applying them.
- **EVD-05 — Behavioral indicators.** Describe observed pace, accuracy habits, unanswered patterns and strategy needs across recent comparable work. Give practical coaching recommendations rather than permanent personality labels; retain indicator/history and the source window.
- **EVD-06 — Missing evidence.** Paper sittings, interrupted sessions and uncertain timing do not produce pace values. They cannot prove a prior timing-related difficulty has vanished. Show coverage and excluded observations.
- **EVD-07 — Focus priorities.** Rank manageable learning priorities using strength of evidence, prerequisite relevance, exam/course goals and subject balance; explain why an item is recommended and permit instructor adjustment with reason.
- **EVD-08 — Readiness.** Provide a visible preparation summary with component definitions, comparable assessment coverage, confidence, goals and history. Omit unsupported dimensions or show insufficient evidence rather than assigning an invented neutral measurement.
- **EVD-09 — Historical corrections.** Rebuild evidence chronologically after regrading, deletion permitted by policy, late paper ingestion or revised scoring. A corrected old result must not simply be appended as a new learning event.
- **EVD-10 — Cohort comparisons.** Compare pace/accuracy only against an appropriate eligible cohort with minimum sample size, anonymity and published population/date. Personal practice scores must not masquerade as formal peer rank.
- **EVD-11 — Educational policy controls.** Provide an authorized settings workspace for evidence minimums, mastery criteria, historical weighting, focus limits, adaptive coverage/difficulty, retest intervals and numerical policy defaults. Validate ranges and relationships, preview effects on representative records, publish a version, reset/restore defaults and document whether existing results are recalculated. Every setting must affect the behavior its description promises.

**Acceptance:** add an older paper result after several online sittings, then correct it. Verify chronological learning history, absence of fake timing and stable reproducible current recommendations.

### 12. Revision journal and contextual tutor

- **REV-01 — Revision collection.** Collect released wrong answers, required skips and learner bookmarks, grouped by course/topic/reason; preserve repeated encounters and origins while avoiding duplicate cards for the same learning target.
- **REV-02 — States and scheduling.** Track new, practicing, demonstrated and reopened revision items. Allow personal notes and instructor guidance; schedule spaced review and require fresh evidence before calling a misunderstanding resolved.
- **REV-03 — Discovery.** Search/filter by wording, topic, subject, assessment, date, response type and resolution state. Show figures/options/solutions, learner working, prior help and next practice action.
- **REV-04 — One tutor conversation.** Provide a consistent persistent tutor experience from result review, revision, lessons and practice, with appropriate context and clearly separated threads/learner privacy.
- **REV-05 — Grounded help.** Supply released question/figure/options, approved solution/rubric, actual learner response and relevant lesson sources. Explain concepts, show steps, offer hints and fresh practice, and cite institution-approved sources when retrieved.
- **REV-06 — Interaction quality.** Support streaming with interruption/retry, starter prompts, follow-up, feedback/reporting, mathematical rendering, useful transcript search, accessible input and recovery of unsent messages.
- **REV-07 — Release enforcement.** Re-check entitlement on every history read, message, retrieval and attachment. Previously released questions must be excluded from confidential scored reuse for that learner by default; policy exceptions need an explicit assessment design decision.
- **REV-08 — Abuse/privacy controls.** Enforce per-person/institution request and spend limits, bounded context, attachment protections, prompt-injection resistance and approved knowledge-source scope. Never execute student-provided code or tool instructions in a privileged environment.
- **REV-09 — Teacher escalation.** Allow a learner to escalate unresolved doubt to an assigned instructor with selected context, status and notification; private tutor history is shared only through the applicable consent/policy.

**Acceptance:** revisit a conversation on a diagram question, request a hint and similar problem, resolve it through fresh work, then verify the same answer cannot leak into a confidential later sitting.

### 13. Study plans, lessons, and practice

- **PLN-01 — Goals.** Set target exam/course outcomes, target date, weekly availability and achievable goals; show progress and allow instructor/learner revisions without fabricated rank promises.
- **PLN-02 — Calendar.** Generate/edit daily, weekly and monthly study plans with lessons, timed drills, revision and assessments; reschedule missed tasks, avoid overload and show dependencies/estimated effort. Export calendar events with timezone correctness.
- **PLN-03 — Instruction design.** Offer reviewed topic lessons, concise notes, formulas, worked examples, misconception checks and unsolved exercises. Adapt explanation style/depth and accessibility to observed needs and learner preference.
- **PLN-04 — Personal context.** Show “what to work on” grounded in the learner's released work without leaking personal examples into shared lessons. Include useful next steps and clear rationale.
- **PLN-05 — Consumption.** Provide section navigation, reader progress, bookmarks, notes, readable maths/figures, downloads when licensed, audio/video/transcript views and remembered position.
- **PLN-06 — Completion and understanding.** Distinguish opened, viewed, activity completed and understanding demonstrated. Retain section progress and allow unmark/revisit; completion alone does not establish mastery.
- **PLN-07 — Exercise responses.** Let learners answer embedded practice/check questions, see policy-appropriate hints/feedback, and retain attempts. Connect reviewed practice evidence to learning indicators using its declared lower/high-stakes weighting.
- **PLN-08 — Content reuse/versioning.** Reuse approved material by topic/outcome/language/level/accessibility/version and compatible needs. Refresh after approved changes while retaining completed historical versions and personal progress mappings.
- **PLN-09 — Instructor assignments.** Assign learning packs or individual resources to classes/learners, set due dates, see progress and understanding checks, and intervene without overwriting a learner's private notes.
- **PLN-10 — Adaptive papers.** Generate personalized papers emphasizing current gaps and prerequisite needs while maintaining breadth, choosing suitable difficulty and avoiding recent/exposed questions. Persist the selection rationale, coverage and fallback decisions.
- **PLN-11 — Thin-bank behavior.** Show unavailable coverage, offer reviewed alternatives or explicitly disclosed repeat practice, and request authoring work. Never silently weaken a formal paper's secrecy or pretend target coverage was met.
- **PLN-12 — Live adaptive practice.** Implement a distinct practice mode that chooses subsequent questions from observed responses and uncertainty, with transparent stopping conditions. Do not mix its score/rank with fixed formal papers without validated normalization.
- **PLN-13 — Continuity.** Replace only obsolete unstarted recommended work, preserve started assignments, re-evaluate priorities after new evidence, and rotate focus after sustained improvement or regression.
- **PLN-14 — Motivation.** Offer optional streaks, achievements and progress celebrations, with rest/accommodation-aware rules. Peer leaderboards are opt-in and cohort-scoped; avoid exposing children's identities publicly.

**Acceptance:** create a realistic study week, complete a lesson/check, miss a task, reschedule it, take adaptive practice and demonstrate improvement that changes the next plan without deleting prior work.

### 14. Teaching library and media studio

- **STU-01 — Resource library.** Upload and organize documents, lesson notes, presentations, video/audio, worksheets and links by course/topic/outcome/language/license. Search, preview, version, approve, assign, hide/unpublish and archive with access controls.
- **STU-02 — Content workflows.** Separate authoring, review, publication and learner visibility; retain feedback/version comparison and dependencies. Broken links, superseded material and missing accessibility assets are visible maintenance work.
- **STU-03 — Generated material.** Generate original drafts of notes, slide lessons, teaching scripts, examples and exercises from approved sources/outcomes. Require academic review where the content may teach or assess a learner; keep source/version/model attribution.
- **STU-04 — Narration.** Generate speech from approved scripts with voice/language/rate controls, pronunciation overrides and math-aware phrasing; provide transcript and downloadable media where licensed. Audio processing must not delay usable text.
- **STU-05 — Presentation-to-video.** Accept PPTX/PDF slides and a custom or generated per-slide script. Select slide range, voice, presenter asset/layout, captions, output quality and target language; validate script/slide alignment.
- **STU-06 — Presenter generation.** Create a talking-presenter video using consented/licensed presenter assets, with static-presenter or narration-only alternatives explicitly identified. Support preview and control of placement/size/background without impersonating an unconsenting person.
- **STU-07 — Captions and exports.** Generate/edit synchronized captions, provide transcript/caption-file export, render a short preview, inspect duration/resolution, then publish/download the full result. Caption editing creates a traceable revision.
- **STU-08 — Concept animations.** Create educational animations through controlled templates/parameters and reviewed scripts, including a representative original mathematical/scientific explanation. Generated animation instructions are untrusted; render in an isolated bounded environment without privileged execution.
- **STU-09 — Job operations.** Queue single/batch generation, show stages/history/thumbnails, estimate available cost/time, save presets, cancel, retry, recover interrupted work, download results and purge unneeded artifacts under retention policy.
- **STU-10 — Topic delivery controls.** Independently control generation, publication and visibility of media/resources; uploaded and generated material can coexist. Hiding output must not falsely display a generating state or delete its audit history.
- **STU-11 — External resources.** Validate resource links, preview supported embeds safely, show source attribution, and provide accessible/open-in-new-tab alternatives. Link access still respects institutional publication scope.

**Acceptance:** turn an original slide deck into an approved narrated/captioned video with a chosen presenter, preview and revise captions, publish to a topic, then hide it without deleting its generation history.

### 15. Instructor work and interventions

- **TCH-01 — Teaching home.** Show assigned classes/learners, upcoming teaching and assessments, recent released results, review queues, unresolved doubts and actionable attention items.
- **TCH-02 — Learner detail.** Inspect permitted subject/outcome evidence, comparable trends, required skips/errors, revision history, study engagement, assessments and attendance. Cross-subject insight may be granted for assigned learners without granting unrelated authoring privileges.
- **TCH-03 — Class insight.** Provide student-by-topic heatmaps, subject/date/class filters, evidence counts, common misconceptions, question quality concerns, top/improving/struggling groups and exports. An empty cell is unobserved, not weak.
- **TCH-04 — Interventions.** Create a private academic note or structured intervention with reason, learner/group, assigned staff, tasks/resources, due date and follow-up outcome. Distinguish staff-only notes from family/learner-visible guidance.
- **TCH-05 — Recommendations.** Propose remedial groups, teaching sequences, targeted resources and reassessment based on actual evidence. Allow teachers to accept/edit/reject and record why; preserve human responsibility.
- **TCH-06 — Teaching outcomes.** Show cohort progress and instructional engagement with comparable populations and evidence limitations. Do not present unadjusted student scores as proof that one teacher caused improvement.
- **TCH-07 — Professional directory.** Maintain approved biography, qualifications/experience, teaching subjects, institutional contact and relevant availability. Default directory access to assigned families/learners and appropriate staff; public publication requires explicit institution/teacher policy.
- **TCH-08 — Roster and review tasks.** Search/filter assigned learners, retain bulk selections, mark/review delegated work and produce a useful student report. Remove access when the teaching assignment ends, subject to policy for retained grading obligations.

**Acceptance:** identify an evidence-backed misconception cluster, assign a short remedial sequence to that group, review a later comparable assessment and record the intervention outcome.

### 16. Family experience

- **FAM-01 — Child workspace.** Provide a clear child switcher and read-only released progress, topic strengths/gaps, learning recommendations, assessment history, rank context where permitted, upcoming deadlines and attendance.
- **FAM-02 — Useful reports.** Show readable summaries and drill-down without overwhelming jargon, with explanation of unknown/insufficient evidence, actual dates, preparation goals and appropriate ways to support the child.
- **FAM-03 — Communication.** Contact assigned staff through permitted threads, view announcements, acknowledge requested notices and arrange meetings where configured. Do not expose staff private phone/email details by default.
- **FAM-04 — Consent and preferences.** Manage notifications, approved child data-sharing/optional-feature choices, relationship/contact verification and receipt of relevant policy changes.
- **FAM-05 — Engagement.** Record actual portal/notice engagement for authorized institution oversight, distinguishing login from meaningful activity. Show transparent definitions and avoid making judgments about parenting from visit frequency.
- **FAM-06 — Print/download.** Print or download a child's released report with clean pagination, readable charts, date range and reference/version. Missing or revoked family relationships must immediately remove private access.

**Acceptance:** one guardian switches between two children at different campuses, receives only allowed notices and cannot retrieve a third child's report through a guessed link.

### 17. Attendance, scheduling, and institutional activity

- **ATT-01 — Attendance capture.** Record class attendance, late/excused/absent states, dates, sessions and authorized corrections. Support manual and validated bulk import; retain who recorded or changed each entry.
- **ATT-02 — Timetables.** Maintain recurring sessions, rooms/online links, holidays, substitutions and schedule exceptions; show appropriate learner/teacher/family calendars and handle timezone/display correctly.
- **ATT-03 — Follow-up.** Notify configured recipients of missing attendance or recurring absence, with correction/acknowledgment workflow and safeguards against duplicate notifications.
- **ATT-04 — Analysis.** Relate attendance and learning participation to observed academic outcomes with date/subject filters and explicit population; describe correlation without claiming causation.
- **ATT-05 — Activity definitions.** Define enrollment, active learner, lesson activity, assessment completion and guardian engagement consistently; show metric windows and avoid inflated activity from polling or background processes.

**Acceptance:** change one attendance entry after an import, issue a single corrected notification and see consistent values in the child report and campus summary.

### 18. Communication, notifications, and collaboration

- **COM-01 — Notification inbox.** Provide categorized unread/read/actioned notices for assessments, reports, lessons, messages, account security, billing and review exceptions; include deep links authorized when opened.
- **COM-02 — Delivery channels.** Implement in-app, email and mobile/web push, plus configured SMS for selected events. Track queued/sent/provider-accepted/delivered/failed where actually observable; keep security messages reliable.
- **COM-03 — Preferences.** Offer per-event/channel preferences, digest frequency, timezone/quiet hours, unsubscribe where applicable, and institution-mandatory notices with explicit justification.
- **COM-04 — Messaging.** Provide scoped instructor–learner and guardian–staff threads, attachments, read-state policy, search, conversation assignment and escalation. Avoid unrestricted discovery/contact between minors or unrelated adults.
- **COM-05 — Class discussions.** Support topic/class forums, questions/replies, helpful answers, attachments, reporting, moderation, locking and archival. Make participation policies and staff responsibility visible.
- **COM-06 — Announcements.** Draft/review/publish scheduled organization/campus/class announcements, select recipients, request acknowledgment and report delivery/engagement without exposing other recipients' private data.
- **COM-07 — Templates and retries.** Version message templates, localize, preview sample data, enforce deduplication, retry transient delivery failures and provide dead-letter recovery. Do not leak answer keys through previews or notification snippets.
- **COM-08 — Meetings.** Offer appointment requests/time slots, confirmation/cancellation, reminders and calendar files; integrate an external meeting/calendar provider only through a configured adapter.
- **COM-09 — Support desk.** Submit a help request, attach consented evidence, track status/owner/replies and link an approved expiring support grant. Provide searchable user help and contextual guidance.

**Acceptance:** a coordinator schedules a class notice, one family receives a digest, an invalid push token fails without losing the inbox notice, and moderation prevents an unauthorized attachment read.

### 19. Reports, analytics, and exports

- **RPT-01 — Personal insight.** Show completed/assigned assessments, average/best normalized scores, accuracy, required coverage, trends, time quality, readiness, focus priorities, study progress and revision outcomes. Every tile drills into its underlying evidence.
- **RPT-02 — Institution oversight.** Compare authorized campuses/classes/courses/subjects with enrollment, participation, completion, learning evidence, staff responsibility and review backlog. Use consistent definitions and comparable populations.
- **RPT-03 — Filters and comparisons.** Support date range, course, campus, subject, class, instructor, assessment type/channel and evidence sufficiency; retain filter state in permitted shareable links. Compare periods with stated denominators.
- **RPT-04 — Report builder.** Select student/class/institution report scope, date range, metrics, sections, grouping/sorting, charts and privacy mode; preview, save templates and generate reports with provenance/version.
- **RPT-05 — Export center.** Produce real PDF, CSV and XLSX outputs; export allowed charts as PNG/SVG and heatmaps as appropriate. Large exports run in the background with access-controlled, expiring downloads and spreadsheet-formula injection protection.
- **RPT-06 — Scheduled reporting.** Schedule report generation/delivery to authorized recipients, honor timezone, permission changes and retention, expose run history, retry safely and stop when recipients lose entitlement.
- **RPT-07 — Custom dashboards.** Provide a constrained interactive chart builder with permitted dimensions/metrics/types/layout, accessible color choices, saved dashboards and drill-down. Do not expose arbitrary SQL or unrestricted query execution.
- **RPT-08 — Leaderboards and benchmarking.** Offer comparable assessment/cohort rankings, institution-approved anonymous peer percentiles and optional learner participation. Show cohort size, ties, eligibility and retake policy; no universal rank fabricated from adaptive practice.
- **RPT-09 — Evidence traceability.** Include calculation policy, source revision and as-of time; show incomplete/stale recalculation states. PDF/Excel/UI totals must agree for the same filters and revision.
- **RPT-10 — Readable sharing.** Create clean print layouts and explicitly authorized expiring report shares with audit and revocation. Never make a child's report publicly discoverable for convenience.

**Acceptance:** a filtered class report produces matching UI, PDF and spreadsheet values, delivers once on schedule, and becomes inaccessible to a revoked recipient.

### 20. Forecasts, confidence support, and model evaluation

- **INS-01 — Performance trajectories.** Implement explainable trend/improvement estimates from comparable historical work, with data sufficiency, uncertainty intervals, influencing factors and target gap. Do not equate extrapolation with a guaranteed exam result.
- **INS-02 — Rank forecasting.** Implement a forecast pipeline with an appropriate licensed outcome/calibration dataset, evaluation split, versioned model and out-of-sample calibration. If the necessary exam/cohort data are absent, return a structured insufficient-evidence result and explain what is needed.
- **INS-03 — What-if tools.** Offer hypothetical study/score scenarios clearly labeled as simulations, constrain inputs and show assumptions/uncertainty. Hypothetical scenarios must never update real learner evidence.
- **INS-04 — Confidence and reflection.** Provide voluntary pre/post assessment confidence, perceived time pressure, and study reflections. Separate self-report from timing inference, permit correction/withdrawal, and respect guardian/institution consent policies.
- **INS-05 — Supportive coaching.** Offer nonclinical reflection prompts, preparation/break strategies and instructor follow-up based on consented observations. Never diagnose anxiety, mental illness or personality from a score, camera, or click pattern; visibly distinguish learner statements from inferred suggestions.
- **INS-06 — Human attention flags.** Show explainable academic/support flags to authorized staff, with evidence, review, dismissal, follow-up and correction. No automated punitive or clinical decision based on a predictive score.
- **INS-07 — Evaluation workspace.** Manage original/licensed evaluation datasets and labels, model/prompt versions, test runs, drift, per-subgroup error and calibration, comparison/rollback and approval. Keep evaluation data scoped/private.
- **INS-08 — Production gates.** Expose evaluated versus experimental insight and fail closed to honest missing evidence when input distribution or model support is inadequate. Deliver the functional model training/evaluation/inference path even when a deployment still needs its real institutional dataset.

**Acceptance:** insufficient historical data yields an honest missing-evidence state; a versioned held-out fixture exercises forecast inference and calibration; a self-report withdrawal removes optional reflection use without deleting unrelated exam marks.

### 21. Plans, payments, invoices, and finance

- **FIN-01 — Plans and entitlements.** Manage subscription plans/trials, seat/campus/storage/AI allowances, feature access, upgrades/downgrades, grace periods, renewal/cancellation and institution billing identity. Preserve learners' access to entitled historical records when a plan changes.
- **FIN-02 — Checkout/payment.** Implement a real payment-provider adapter with sandbox support, hosted/tokenized payment collection, state tracking, signed callbacks and idempotent settlement. The platform must not store raw payment-card details.
- **FIN-03 — Financial records.** Represent money in currency-appropriate minor units with explicit currency, rounding and discount/tax configuration. Issue immutable numbered invoices/receipts and correction/credit records instead of editing settled totals.
- **FIN-04 — History and billing UI.** Show invoices, credits, payments, failed collection, due dates, subscription state and downloadable documents to authorized finance staff and payers. Explain pending versus completed payment.
- **FIN-05 — Refunds/cancellation.** Provide permissioned refund requests, reason/approval, full/partial refund execution, provider confirmation and failure/reconciliation. Prevent concurrent or replayed requests from refunding more than the settled refundable amount.
- **FIN-06 — Webhook/reconciliation.** Verify authenticity, persist event IDs, tolerate replay/out-of-order delivery, reconcile provider status and expose discrepancies. A checkout-success page is not proof of settlement.
- **FIN-07 — Budgets and usage.** Meter billable seats/services and AI consumption transparently, show period limits and actual-versus-estimated amounts, alert budget owners and define behavior at the limit without losing exam answers.
- **FIN-08 — Financial insight.** Report collected/outstanding/refunded revenue, subscription changes and usage trends, scoped by organization/campus where meaningful. Separate accrual/invoice views from cash receipts and document definitions.
- **FIN-09 — Provider setup and access.** Validate configuration, provide test-mode indicators and controlled production activation, mask secrets, separate finance permissions, and audit refunds/plan changes. Never charge or issue refunds during automated demo seeding.

**Acceptance:** a sandbox payment with duplicated/out-of-order webhooks settles once, generates one invoice, supports an approved partial refund and reconciles the remaining amount correctly.

### 22. Public experience, appearance, language, and accessibility

- **UX-01 — Public site.** Provide original home, feature/audience explanations, company/about, pricing where configured, contact/demo request, help, terms/privacy and sign-in/enrollment entries. A demo booking/contact form must actually persist and route requests.
- **UX-02 — Honest content.** Publish only supported features and verified outcomes; show real availability/configuration rather than invented testimonials, customer numbers, prediction accuracy or locked-browser claims.
- **UX-03 — Original visual system.** Define coherent typography, spacing, color, density, components and motion; support light/dark/system preference and text-size settings. Keep complex staff tools legible and learner work calm and focused.
- **UX-04 — Responsive behavior.** Support phones, tablets, desktop and print; dense tables need usable filtering/column choices or alternate views. Maintain accessible touch targets and avoid horizontal overflow for core workflows.
- **UX-05 — Accessibility.** Target WCAG 2.2 AA through keyboard navigation, visible focus, semantic labels, error summaries, screen-reader context, contrast, reduced motion, captions/transcripts and accessible mathematical content. Test complete workflows, not only component snapshots.
- **UX-06 — Localization.** Provide an internationalization foundation and complete English/Hindi UI translations for delivered workflows, locale-aware dates/numbers and content-language selection. Keep technical identifiers stable and permit additional languages without rewriting business logic.
- **UX-07 — State quality.** Every important view needs truthful loading, empty, partial, offline, permission-denied, recoverable error and success states, including saved feedback and destructive-action confirmation. Preserve user input across retryable failures.
- **UX-08 — Discoverability.** Implement entitled global search, navigation breadcrumbs, remembered workspace context, accessible dialogs, contextual help and clear account/child switches. Do not expose records through search snippets beyond the user's scope.
- **UX-09 — Public performance/metadata.** Provide appropriate discoverability metadata, sitemap/canonical handling and fast public pages. Exclude private learner pages and reports from indexing/cache leakage.

**Acceptance:** a keyboard-only user can enroll, take an accommodated assessment and read feedback in either UI language; a small-screen guardian can switch children and download a readable report.

### 23. Native mobile and offline access

- **MOB-01 — Native applications.** Deliver Android/iOS source/build configuration with sign-in/MFA/recovery, learner/family home, lessons, tutor, revision, eligible assessments, notifications, goals and released reports. Reuse business contracts and design tokens, not browser-only components by force.
- **MOB-02 — Secure local state.** Use OS-protected credential storage, per-user/institution caches, explicit downloads and eviction, local data encryption appropriate to sensitivity, and sign-out/revocation cleanup. Do not leave another child's records accessible after switching accounts.
- **MOB-03 — Offline study.** Download permitted notes/media/exercises and continue reading/progress/notes offline; show download rights, size, availability, expiry and synchronization state. Protected formal answer keys are never cached early.
- **MOB-04 — Synchronization.** Use retryable versioned mutations, stable operation IDs, conflict review/resolution and server acknowledgments. Preserve edits, handle partial upload and stop sync when membership/permissions change.
- **MOB-05 — Assessment network policy.** Formal examinations require server-backed eligibility/deadlines and a documented connectivity policy. Offline practice is supported; offline formal sittings are allowed only through an explicitly configured supervised delivery design, never by silently trusting device time.
- **MOB-06 — Push/deep links.** Register/revoke push tokens by installation, navigate to authorized content, respect child/workspace context and preferences, and handle stale or revoked links safely.
- **MOB-07 — Web installation.** Provide installable web access and offline study where browser capabilities permit, with truthful limits compared with native apps. Service workers must not cache unreleased assessments/results or private responses across users.
- **MOB-08 — Qualification.** Verify interruption/resume, updates, permission revocation, accessibility, low bandwidth, storage exhaustion and basic Android/iOS flows; document store-signing/release configuration separately from source completion.

**Acceptance:** download a lesson, study offline, add a note, reconnect with a simultaneous server edit, resolve the conflict, sign out and verify protected records are unavailable to the next user.

### 24. AI control, instruction versions, and provider operations

- **AIO-01 — Capability adapters.** Route text, vision/document, retrieval, speech and presenter generation through Go-controlled adapters with explicit model/provider capabilities, timeouts, cancellation and supported input limits. At least one functional adapter is required for each delivered AI capability.
- **AIO-02 — Instruction management.** Edit/review/version generation instructions, describe variables/output contracts, preview/test against fixtures, compare history, publish/reset/rollback and retain editor/approver identity.
- **AIO-03 — Output validation.** Validate structured outputs, academic references, allowed values, completeness and asset relationships before business use. Preserve malformed/failed outcomes safely for authorized debugging without exposing private data unnecessarily.
- **AIO-04 — Reuse/invalidation.** Key reusable generation by approved input, instruction, model, language and policy revisions; prevent duplicate concurrent generation; refresh dependent material after approval while preserving completed historical content.
- **AIO-05 — Complete consumption accounting.** Record requests, model/provider, feature, tenant/person scope, input/output/reasoning/cached tokens where observable, retries/failures, latency, monetary estimates and provider-reported actual cost. Distinguish missing cost data from zero.
- **AIO-06 — Budgets/limits.** Enforce per-feature/person/organization ceilings and shared provider quotas, reserve/reconcile spend for concurrent jobs, and provide warnings and admin overrides with audit. Unauthorized requests must not burn a learner's allowance.
- **AIO-07 — Resilience.** Use capability-compatible fallback and bounded retry/backoff; distinguish transient overload, authentication failure, exhausted budget and daily quota. Do not repeatedly retry a permanent error or discard approved human review.
- **AIO-08 — Quality and feedback.** Collect content/tutor reports, run repeatable evaluations, compare provider/instruction revisions and gate publication on actual evidence. Track hallucination/diagram extraction/marking failures as supportable incidents.
- **AIO-09 — Data boundaries.** Isolate retrieval/content caches by tenancy and permitted sharing; minimize personal data, redact logs and control provider retention settings. Treat document/student content as data, not instructions to access tools or secrets.
- **AIO-10 — Operator console.** Show current configuration/capabilities, consumption, budgets, failures, evaluation results, generation history and recoverable jobs. Secrets are set through controlled configuration and never revealed through the ordinary console/API.

**Acceptance:** an invalid vision reply is rejected, the approved reviewer edit survives retry, a shared quota limit is respected under concurrent jobs, and actual/unknown cost remains distinguishable in reports.

### 25. Reliability, security, privacy, and service operations

- **RUN-01 — Durable execution.** Persist job intent with the triggering business transaction; support leases/heartbeat, retry attempts, exponential backoff where appropriate, cancellation, dead-letter inspection and crash recovery. Exactly-once business outcomes must survive at-least-once delivery.
- **RUN-02 — Operational console.** Search/filter jobs, inspect safe error context, replay failed work with authorization, cancel or reschedule, see backlog/progress, and identify orphaned records/assets. Every administrative recovery action is audited.
- **RUN-03 — File lifecycle.** Store originals and derivatives privately, validate file signatures and limits, quarantine unsafe/unscannable uploads, use bounded document/media processing, preserve provenance, handle resumable uploads and remove abandoned artifacts safely.
- **RUN-04 — Audit.** Provide append-only security/business audit records with actor, tenant, scope, operation, reason, object/revision, correlation ID and time. Search/export through appropriate permissions; logs must not contain passwords, session tokens, raw card data or unrestricted tutor transcripts.
- **RUN-05 — Application security.** Protect CSRF/session cookies, request boundaries, injection/SSRF, upload execution, object access, cross-tenant reads/writes, export abuse, output rendering and webhook integrity. Encrypt transport and restrict service credentials; use centralized secret management appropriate to deployment.
- **RUN-06 — Privacy/consent.** Provide data-access/export/correction/deletion-request workflows, guardian/optional-feature consent records, revocable shares, retention schedules and legal holds. Implement the institution's configured jurisdiction/policy requirements without claiming automatic compliance from a checkbox.
- **RUN-07 — Resilience/health.** Offer liveness/readiness, structured logs, request/job traces and metrics for API latency/errors, database saturation, job age, marking delay, file processing, provider quota and financial reconciliation. Expose safe operator dashboards and actionable alerts.
- **RUN-08 — Backup and recovery.** Schedule encrypted database/object backups, define RPO/RTO and retention, document restore, and perform a timed restore drill including academic versions, files, financial records and access controls. A backup file alone is not recovery evidence.
- **RUN-09 — Environment/deployment.** Provide local containers, migrations, validated secrets/provider settings, CI checks, staged releases, rollback procedures, maintenance mode and safe schema evolution. Reboots and deployments must not lose private files or acknowledged responses.
- **RUN-10 — Performance/concurrency.** Adopt the reference workload in section F, load-test actual behavior, bound expensive work and stream/queue large operations. Do not hardcode success claims without an observed run.
- **RUN-11 — Support recovery.** Resolve account/session issues, failed import/scan jobs, missing files, disputed marks, notification failures and payment discrepancies through authorized tools/runbooks. Operator access to private records requires a scoped support grant and audit.
- **RUN-12 — Data portability.** Export/import documented institution-owned data formats, validate mappings and reconcile counts/revisions in a fresh installation. Do not require any external historical application or its data model.
- **RUN-13 — Delivery evidence.** Maintain requirement status, test results, screen/state coverage, accessibility observations, provider setup and known limitations. Distinguish source completion, local/sandbox verification and production verification.

**Acceptance:** kill a worker after a side effect but before acknowledgment, restart it, restore a backup into a clean environment and verify one academic/financial outcome, authorized file access and intact audit history.

## D. Cross-feature business rules

### D.1 Institutional boundaries

Every data access and asynchronous execution carries an explicit institution context. Membership does not imply access to every campus or child. Files, search indexes, generated material, message recipients, exports and live subscriptions obey the same rule as ordinary records.

An operator may see job identifiers and sanitized status without seeing student content. A signed download link is a temporary capability: scope, lifetime, revocation policy and leak risk must be deliberate. Answer assets use the applicable release policy.

### D.2 Immutable assessment evidence

Separate authored content revisions, published paper revisions, recipient assignments, learner sittings, response observations, marking revisions and released reports. These are product concepts, not a prescribed database schema.

Retain the exact content, assets, key/rubric, order/random seed, policy and accommodation snapshot needed to reproduce an awarded mark. A later teacher edit does not change a past result until an explicit correction/regrading workflow approves it.

### D.3 Optional sections and denominators

Example acceptance fixture: an original paper has 18 mandatory selection questions at 3 marks each, and a numerical section containing 12 questions of which 8 count at 2 marks each. Its maximum is **70**, not 78. Test first-N and best-N independently, including penalties, too many responses, blank groups, cancelled questions and exact ties.

For learning evidence, distinguish required-but-skipped from not-required-to-answer. An answered but uncounted optional item may contribute to formative evidence only under the published policy; it must not be counted twice as scored evidence. Maintain consistent definitions in all dashboards, exports and predictions.

### D.4 Event time and historical correction

Use actual sitting date for academic chronology and a separate received/processed date for operations. Late paper intake changes academic history in its proper chronological position. Regrading an older sitting may require rebuilding later derived evidence; ensure a coherent published revision rather than a mix of old/new aggregates.

### D.5 Submission and offline synchronization

Server acknowledgments identify the accepted response revision. The UI distinguishes a locally stored answer from one accepted by the server. At expiry, finalize the authoritative accepted state, retain evidence of unresolved client edits if policy permits, and explain what happened. Do not accept unlimited backdated answers based on a user-controlled clock.

Explicit supervised offline examinations require their own trusted operational policy. Ordinary offline study/practice must not weaken formal examination rules.

### D.6 AI boundaries and unknown evidence

AI may extract, explain, propose or generate. Fixed-answer marks, permissions, financial settlement, final publication and authoritative deadlines are deterministic or approved human decisions.

Represent observed fact, human correction, inferred explanation and learner self-report separately. Unknown pace is not zero pace; unreadable handwriting is not a blank; an inferred concept gap is not a clinical or personality finding. Every derived measure should expose enough evidence for a teacher to challenge it.

### D.7 Safe answer reuse

Maintain learner exposure to released question/solution material. By default, confidential formal papers exclude items whose solutions that learner could already access. If a paper intentionally reuses formative material, label its purpose and disable misleading confidentiality/rank claims. A tutor's historical thread must never become an indirect key-delivery channel for a live confidential sitting.

### D.8 Money, notifications, and side effects

Use stable idempotency references for payment requests, callbacks, refunds, admissions, document commits, submissions and notification dispatch. Keep external provider attempt records separate from completed business outcomes. Recovery reconciles actual state rather than repeating a side effect blindly.

### D.9 Completion and deletion

Opening a lesson is not understanding it. Starting a payment is not settlement. Queuing a message is not delivery. Receiving scans is not releasing results. A batch can finish operationally with failures that remain visible.

Archiving a person, paper or course ends future activity without rewriting past evidence. Deletion/privacy requests use a deliberate retention/hold workflow covering objects, derivatives, search, providers, backups and shares.

### D.10 Starting defaults for a runnable product

Use these defaults unless a documented decision improves them. They are intended to avoid blocking implementation on routine choices, and are configurable where the product needs institutional variation:

| Decision | Starting behavior |
|---|---|
| Browser authentication | Opaque server-side sessions using secure HttpOnly cookies, CSRF protection, rotation and revocation. Do not store a reusable browser bearer credential in ordinary local storage. |
| Enrollment | Verified registration joins an admission queue or a valid invitation; staff approve class/campus placement. Privileged roles require an authorized invitation/grant. |
| Family access | Multiple verified guardians/children are supported; newly linked guardians see released information only and receive explicit institution approval. |
| Identity | Permanent internal person identity plus nonrecycled human reference/alias. Campus transfer does not rewrite permanent identity. |
| Assessment revision | Publishing pins an immutable paper revision; an academic correction uses a distinct reviewed marking revision. |
| Retakes | One scored sitting per formal assignment unless authorized staff grant another; practice may repeat. Formal ranking uses the policy pinned at assignment time. |
| Result release | Practice releases on completion; formal answers/solutions release after the cohort closes and marking/moderation completes. |
| Conduct | Practice is unmonitored by default; formal browser monitoring requires an explicit policy. Warnings default to staff review, with automatic finalization available only when configured and communicated. |
| Optional questions | Score only the configured count. Exclude not-required blanks from avoidance measures. Publish offline first-N selection in printed question order. |
| Marking | Exact-set multi-selection by default, partial credit only where explicitly authored. Objective numeric answers use authored decimal/range/tolerance rules; no guessed global tolerance or rounding. |
| Scan commitment | Commit selected clean sheets, keep unresolved sheets open, and require explicit review of identity/template/paper conflicts. |
| Shared generated material | Only reviewed material can be shared; strip personal examples and include institution sharing/license scope in reuse decisions. |
| Forecasts | Show actual readiness dimensions and data sufficiency; rank forecasts remain unavailable until their required calibrated model/data are present. The configured inference/evaluation path is still required. |
| Communications | Class/relationship-based discovery only; no open cross-institution messaging or public child ranking. |
| Billing | Sandbox mode locally; explicit authorized activation for production provider credentials. No real side effects from fixtures or routine test runs. |
| Languages | English and Hindi UI, organization-selected timezone, and separately tagged academic content language. |
| Repository shape | A new workspace containing the Go application, web client, native client, contracts and deployment assets; choose actual directory/package names after recording the new domain design. |

Before the learning-evidence milestone, document and test an independently chosen reproducible mastery/indicator policy, including its formulas and data minimums. This is a required design deliverable rather than an invitation to hide arbitrary numbers in code. Before external launch, institution owners must configure retention, financial/tax details, consent requirements and service commitments appropriate to their actual operation.

## E. Engineering constraints and interfaces

### E.1 Backend and web separation

Use Go for all business capabilities and durable orchestration. Next.js owns presentation/rendering and secure API consumption. Background generation, scanning, expiry, reporting and provider callbacks must remain functional without a web request keeping a process alive.

Prefer explicit capability boundaries and a small deployable application with separately runnable workers. Choose fresh domain/package names and an original schema based on these requirements. Do not prescribe or recreate a role-by-role package/route tree. Introduce more services only after a measured isolation or deployment need.

### E.2 API contract

Document authentication, versioning, pagination/cursors, filters, structured errors, request correlation, optimistic concurrency, idempotency, file permissions, streaming progress and callback signatures. Supply machine-readable contracts and generated web/native clients; don't maintain several incompatible handwritten copies.

Use consistent timestamps/timezones and bounded resource limits. Long operations return a job reference and status rather than blocking until an external renderer finishes. Validate output as rigorously as input, especially numerical marking and OCR/AI payloads.

### E.3 Data integrity

Use transactions, database constraints and concurrency control for business invariants. Durable work handoff must occur with committed state. Choose a maintained queue/outbox approach and prove recovery; an in-memory goroutine launched after a request is not durable orchestration.

Store grades and money with appropriate exactness; numerical-answer acceptance requires deliberate decimal/rounding behavior. Maintain append-only change/audit history where required without building an unnecessary universal event-sourcing framework.

### E.4 Providers

At implementation time, choose and document maintained providers that satisfy required modality, privacy, geographic and sandbox needs. Verify official documentation and supported input/output formats. Provider selection is a configuration/adapter decision; no specific model name or output quality is assumed by this specification.

Implement configuration validation and a clear provider capability matrix. Local fixtures exercise workflows without real credentials; configured sandbox/live integration tests establish what has actually been verified. Production activation is an operational action, not a default of installing dependencies.

### E.5 Educational and rendering assets

Use original/licensed material, synthetic identities and consented presenter assets. Create representative science/math content and scanned/script fixtures to validate figures, handwriting uncertainty, option formats and different paper structures. Track license and provenance.

For media/OCR, a Go worker may invoke maintained native tools or hosted services through isolated resource-limited processing. No application backend in another language is required. Do not execute unreviewed generated code on the application host.

## F. Nonfunctional acceptance targets

These are starting engineering targets, not claims of measured achievement. Record hardware, dataset, provider limits and observed results; report failures honestly. Adjust only with a documented product/engineering decision.

| Area | Target / required proof |
|---|---|
| Reference load | At least 3 institutions, 8 campuses, 12,000 synthetic enrolled learners, 600 concurrent exam writers and 60 concurrent staff views. Load data must have realistic distributions. |
| Exam durability | No loss of acknowledged responses and no duplicate final results under retry, worker crash, competing expiry/submission and ordinary rolling restart. |
| Response latency | On the documented reference environment, p95 ordinary reads under 700 ms and exam-save acknowledgment under 500 ms, excluding simulated poor networks/external providers. Objective marking for a 120-item paper should normally return within 3 seconds. |
| Paper intake | A 400-page mixed-validity batch is accepted durably without blocking a single request for the recognition duration; progress and failures remain recoverable. Actual throughput depends on documented provider/CPU limits. |
| Availability | Define the production service objective and alert policy; target at least 99.9% monthly availability for the core API after qualification. Do not claim the target has been achieved without observation. |
| Recovery | Starting target: database RPO 15 minutes and core recovery RTO 2 hours, with documented file/backup consistency and an observed restore drill. |
| Accessibility | Verify representative complete web journeys against WCAG 2.2 AA and native accessibility conventions; document remaining defects. |
| Browser/device | Support current and previous major Chrome/Edge/Firefox/Safari where practical; capability-dependent monitoring must degrade transparently. Test representative Android/iOS devices. |
| Privacy isolation | Automated negative tests across institutions, campuses, guardian links, rosters, exports, files, streams, provider contexts and background jobs. |
| Monetary correctness | Exact minor-unit calculations and no duplicate settlement/refund under replay/concurrency; ledger-to-provider reconciliation exercised. |
| Evidence quality | Report OCR/question extraction fidelity, reviewer agreement, tutor/content evaluation and forecast calibration using declared representative datasets; no unsupported universal accuracy threshold. |
| Observability | Trace a sitting from saved response through grading, learning updates, generated assignments and notification; trace a payment through webhook and reconciliation without revealing secrets. |

## G. Release journeys and required demonstrations

All journeys below must have repeatable fixtures and appropriate automated or documented manual evidence.

1. **Admission and access:** establish an institution/campus/course, invite staff and a guardian, admit two children, issue references, enable MFA and exercise role/scoped access changes.
2. **Paper authoring:** import a multi-page original paper and separate key, preserve figures, correct drafts, apply optional-section scoring, approve and publish a pinned revision.
3. **Online examination:** assign a formal paper with an accommodation, save and resume through network interruption, handle duplicate clients, finalize at deadline and release only after policy conditions.
4. **Paper examination:** intake several layouts, resolve ambiguous bubbles/handwriting, reject a wrong-paper scan, preserve reviewer corrections across recognition retry and commit each accepted sitting once.
5. **Written evaluation:** ingest a handwritten script, review uncertain transcription, propose rubric marks, moderate and release only approved feedback.
6. **Learning loop:** explain evidence and uncertainty, assign topic lessons and timed practice, revisit revision items, reassess and demonstrate a justified priority/readiness change.
7. **Tutor safety:** ask about a released diagram question, retain a follow-up thread, escalate a doubt to a teacher and prove unreleased material remains inaccessible.
8. **Teaching:** inspect a class heatmap, form a remedial group, assign learning tasks, record a private intervention and measure later outcomes with appropriate limits.
9. **Media studio:** produce an original presentation-based lesson with narration, presenter, captions and a concept animation; preview, approve, publish, retry a failure and revoke publication.
10. **Family support:** switch children, inspect released progress/attendance, acknowledge a notice, message appropriate staff and download an accurate readable report.
11. **Organization insight:** compare campuses and subjects, build a custom filtered dashboard/report, export matching PDF/XLSX values and schedule one permission-checked delivery.
12. **Advanced insight:** show insufficient data honestly, run a versioned forecast evaluation, exercise a clearly hypothetical scenario and manage consented self-reported confidence.
13. **Commerce:** change a plan, settle a sandbox payment through replayed/out-of-order callbacks, issue documents, refund partially and reconcile balances/entitlements.
14. **Mobile/offline:** download permitted lessons, work offline, reconcile a conflict, receive a push/deep link and revoke a device without cross-user cache leakage.
15. **Correction:** approve an answer-key change, regrade affected sittings, rebuild chronological evidence and reports, and retain previous revisions and issued notices.
16. **Disaster/support recovery:** crash a worker, recover a partially completed operation without duplicating effects, use an expiring support grant, restore data/files and demonstrate consistent authorized access.

## H. Completion and handoff

The implementation is complete only when the full requirement register and expanded subchecks have demonstrable evidence. Publish the following in the new repository:

- Setup/run instructions and reproducible local services.
- API contracts, migrations, provider adapter/setup guide and environment validation.
- Web/native application builds and accessible representative screenshots/recordings.
- Requirement register with implementation and verification states, actual evidence and explicit unresolved dependencies.
- Unit, integration, contract, end-to-end, load, security, recovery and educational/model evaluation results, distinguishing unexecuted checks.
- Operations runbooks, backup/restore evidence, support workflows, retention policy configuration and financial reconciliation instructions.
- Original/licensed fixtures and a demonstration guide covering section G without production credentials.
- Recorded design decisions, known limitations, measured targets and production activation requirements.

Do not call required native, billing, written-review, predictive, collaboration or media features “later” at final handoff. If an external service or real dataset remains unavailable, state precisely which configured/production verification is blocked, retain the requirement as incomplete at that level, and continue all independent authorized implementation work.
