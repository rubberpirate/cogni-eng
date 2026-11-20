# Implementation status

All scope is required. `planned` means unimplemented; `building` means partial source exists; `verified` requires backend, permissions, persistence, UI, recovery and acceptance evidence. External production verification is separate. No requirement is verified at initialization.

| Requirement | Milestone | State | Evidence | Relevant test | Blocker |
|---|---|---|---|---|---|
| [ID-01](#id-01) — Registration and invitations | 1 | building | Registration code, verification, invitation and placement APIs/UI; policy acknowledgments | TestLearnerRegistrationRequiresVerificationAndStaffPlacement; registration browser journey | Phone verification, configurable approval policy and production email verification outstanding |
| [ID-02](#id-02) — Credential sign-in | 1 | building | internal/access; web access form with pre-hydration credential guard and POST fallback | TestPasswordAndSecretBoundaries; browser sign-in | Full accessible validation/localization and credential remediation pending |
| [ID-03](#id-03) — Password recovery/change | 1 | building | Recovery/reset, password change and verified email-change APIs/UI | TestAccountChangesPreserveVerificationAndRevokeOtherSessions | Security notifications and complete recovery browser journeys outstanding |
| [ID-04](#id-04) — MFA | 1 | building | TOTP enrollment, single-use recovery codes, institutional enforcement and security UI | TestStaffMFAEnforcementAllowsEnrollmentRecoveryPath | Lost-device support recovery and configured alternative factors outstanding |
| [ID-05](#id-05) — Session management | 1 | building | Server sessions; session list/revoke/logout UI | TestInstitutionIsolationAndRevocation | Rotation, approximate context and complete browser recovery checks outstanding |
| [ID-06](#id-06) — Account state | 1 | building | Membership suspension/restoration with audit and session revocation | TestInstitutionIsolationAndRevocation | Global account-state administration and complete recovery workflow outstanding |
| [ID-07](#id-07) — Abuse controls | 1 | building | PostgreSQL sign-in/recovery/token throttles | Integration suite exercises requests | Remaining request categories and concurrent abuse tests outstanding |
| [ID-08](#id-08) — Profile | 1 | building | Profile/contact/activity and cropped private image APIs/UI with cleanup history | TestPrivateAvatarStorageAndLifecycle; profile crop browser journey | Staff-controlled academic fields and complete profile/localization qualification outstanding |
| [ID-09](#id-09) — Safe bootstrap | 1 | building | cmd/operator one-time bootstrap and development guard | Local startup exercised | Compromised-credential remediation and bootstrap acceptance test outstanding |
| [ORG-01](#org-01) — Institution lifecycle | 1 | building | Institution creation/settings, timezone/locale/support/year and MFA policy UI | Settings browser journey; MFA policy integration test | Branding, lifecycle, closure and subscription entitlements incomplete |
| [ORG-02](#org-02) — Campus management | 1 | building | Scoped campus creation/edit/closure checks with optimistic version UI | Campus editing browser journey; scoped coordinator tests | Contacts, responsible staff and active-date/reassignment workflows incomplete |
| [ORG-03](#org-03) — Courses and groups | 1 | building | Campus-bound course creation/list UI | Browser course workflow | Groups, cohorts, sections and timetable management outstanding |
| [ORG-04](#org-04) — Learner admission | 1 | building | Single admission and verified registration placement queue with permanent references | Registration/placement and admission browser journeys | CSV/XLSX, eligibility/grade and target exam metadata outstanding |
| [ORG-05](#org-05) — Human references | 1 | building | Nonrecycled sequence-backed learner references | Browser admission journey | Paper references, aliases, external IDs and correction workflow outstanding |
| [ORG-06](#org-06) — Transfers and withdrawal | 1 | building | Transfer/withdrawal API/UI with placement journal and session revocation | TestPlacementReferenceHistoryAndStaleWrite | Complete history viewing, eligibility/assignment effects and scheduled transfer handling outstanding |
| [ORG-07](#org-07) — Guardian relationships | 1 | building | Approve/remove/reapprove multiple family links, with retained audit and UI | Family-link browser creation/removal journey | Guardian verification, child workspace and custody controls outstanding |
| [ORG-08](#org-08) — Staff and teaching assignments | 1 | building | Course/subject/effective-period staff responsibilities UI | Integration permission boundaries | Professional profile, roster editing and pagination outstanding |
| [ORG-09](#org-09) — Permission administration | 1 | building | Role templates, scoped checks, membership state UI and password/MFA-protected owner transfer | TestOwnershipTransferNeverOrphansInstitution | Custom scoped capability grants and bulk preview outstanding |
| [ORG-10](#org-10) — Import/reconciliation tools | 1 | planned | None | Not executed | Implementation outstanding |
| [CAT-01](#cat-01) — Curriculum models | 2 | building | Topic/course outcomes; typed authoring/review; revision snapshots; private diagrams; JSON/CSV/XLSX preview/commit UI | TestAcademicImportPreviewDedupAndScanPixelsToMarks; academic browser journey | Hierarchy/versioned curriculum, rich math/chemistry, image bundles, bank operations and complete acceptance remain; see docs/CORE-EDUCATION.md |
| [CAT-02](#cat-02) — Question types | 2 | building | Topic/course outcomes; typed authoring/review; revision snapshots; private diagrams; JSON/CSV/XLSX preview/commit UI ; task navigation and original practice kit (docs/CORE-EDUCATION.md) | TestAcademicImportPreviewDedupAndScanPixelsToMarks; academic browser journey ; 10 browser journeys; sample PDF/PG integration (docs/TEST-REPORT.md) | Hierarchy/versioned curriculum, rich math/chemistry, image bundles, bank operations and complete acceptance remain; see docs/CORE-EDUCATION.md |
| [CAT-03](#cat-03) — Rich question content | 2 | building | Topic/course outcomes; typed authoring/review; revision snapshots; private diagrams; JSON/CSV/XLSX preview/commit UI | TestAcademicImportPreviewDedupAndScanPixelsToMarks; academic browser journey | Hierarchy/versioned curriculum, rich math/chemistry, image bundles, bank operations and complete acceptance remain; see docs/CORE-EDUCATION.md |
| [CAT-04](#cat-04) — Academic metadata | 2 | building | Topic/course outcomes; typed authoring/review; revision snapshots; private diagrams; JSON/CSV/XLSX preview/commit UI | TestAcademicImportPreviewDedupAndScanPixelsToMarks; academic browser journey | Hierarchy/versioned curriculum, rich math/chemistry, image bundles, bank operations and complete acceptance remain; see docs/CORE-EDUCATION.md |
| [CAT-05](#cat-05) — Lifecycle and revision | 2 | building | Topic/course outcomes; typed authoring/review; revision snapshots; private diagrams; JSON/CSV/XLSX preview/commit UI | TestAcademicImportPreviewDedupAndScanPixelsToMarks; academic browser journey | Hierarchy/versioned curriculum, rich math/chemistry, image bundles, bank operations and complete acceptance remain; see docs/CORE-EDUCATION.md |
| [CAT-06](#cat-06) — Bank management | 2 | building | Topic/course outcomes; typed authoring/review; revision snapshots; private diagrams; JSON/CSV/XLSX preview/commit UI | TestAcademicImportPreviewDedupAndScanPixelsToMarks; academic browser journey | Hierarchy/versioned curriculum, rich math/chemistry, image bundles, bank operations and complete acceptance remain; see docs/CORE-EDUCATION.md |
| [CAT-07](#cat-07) — Duplicate management | 2 | building | Topic/course outcomes; typed authoring/review; revision snapshots; private diagrams; JSON/CSV/XLSX preview/commit UI | TestAcademicImportPreviewDedupAndScanPixelsToMarks; academic browser journey | Hierarchy/versioned curriculum, rich math/chemistry, image bundles, bank operations and complete acceptance remain; see docs/CORE-EDUCATION.md |
| [CAT-08](#cat-08) — Bulk authoring/import | 2 | building | Topic/course outcomes; typed authoring/review; revision snapshots; private diagrams; JSON/CSV/XLSX preview/commit UI | TestAcademicImportPreviewDedupAndScanPixelsToMarks; academic browser journey | Hierarchy/versioned curriculum, rich math/chemistry, image bundles, bank operations and complete acceptance remain; see docs/CORE-EDUCATION.md |
| [CAT-09](#cat-09) — Review and provenance | 2 | building | Topic/course outcomes; typed authoring/review; revision snapshots; private diagrams; JSON/CSV/XLSX preview/commit UI | TestAcademicImportPreviewDedupAndScanPixelsToMarks; academic browser journey | Hierarchy/versioned curriculum, rich math/chemistry, image bundles, bank operations and complete acceptance remain; see docs/CORE-EDUCATION.md |
| [CAT-10](#cat-10) — Quality and item analysis | 2 | planned | None | Not executed | Implementation outstanding |
| [DOC-01](#doc-01) — Source intake | 2 | building | Private PDF/image intake; durable Poppler/Tesseract extraction; source/text/row review and approval gates ; ordered inline question/option editing, source numbers, reviewed separate-key upload  Local LightOn OCR and pinned original mark-classifier adapters; ordered faculty review, private inference provenance and recovery (docs/CORE-EDUCATION.md); | TestActualPDFRasterizationAndBounds; import/scan integration tests ; two-column and image-key browser journey; key authorization/version integration  live local 10-question/40-option and 180-position/720-bubble sample evaluation; model boundary/recovery tests (docs/TEST-REPORT.md); | DOCX, general key-layout matching, figures/equation fidelity, full document acceptance outstanding; broader field/model evaluation and inference-container qualification remain open. |
| [DOC-02](#doc-02) — Durable extraction workspace | 2 | building | Private PDF/image intake; durable Poppler/Tesseract extraction; source/text/row review and approval gates ; ordered inline question/option editing, source numbers, reviewed separate-key upload  Local LightOn OCR and pinned original mark-classifier adapters; ordered faculty review, private inference provenance and recovery (docs/CORE-EDUCATION.md);; shared visible processing states, upload steps, recovery and direct upload navigation | TestActualPDFRasterizationAndBounds; import/scan integration tests ; two-column and image-key browser journey; key authorization/version integration  live local 10-question/40-option and 180-position/720-bubble sample evaluation; model boundary/recovery tests (docs/TEST-REPORT.md);; processing-feedback browser transition/retry/mobile/reduced-motion check | DOCX, general key-layout matching, figures/equation fidelity, full document acceptance outstanding; broader field/model evaluation and inference-container qualification remain open. |
| [DOC-03](#doc-03) — Faithful extraction | 2 | building | Private PDF/image intake; durable Poppler/Tesseract extraction; source/text/row review and approval gates ; ordered inline question/option editing, source numbers, reviewed separate-key upload  Local LightOn OCR and pinned original mark-classifier adapters; ordered faculty review, private inference provenance and recovery (docs/CORE-EDUCATION.md); | TestActualPDFRasterizationAndBounds; import/scan integration tests ; two-column and image-key browser journey; key authorization/version integration  live local 10-question/40-option and 180-position/720-bubble sample evaluation; model boundary/recovery tests (docs/TEST-REPORT.md); | DOCX, general key-layout matching, figures/equation fidelity, full document acceptance outstanding; broader field/model evaluation and inference-container qualification remain open. |
| [DOC-04](#doc-04) — Key matching | 2 | building | Private PDF/image intake; durable Poppler/Tesseract extraction; source/text/row review and approval gates ; ordered inline question/option editing, source numbers, reviewed separate-key upload  Local LightOn OCR and pinned original mark-classifier adapters; ordered faculty review, private inference provenance and recovery (docs/CORE-EDUCATION.md); | TestActualPDFRasterizationAndBounds; import/scan integration tests ; two-column and image-key browser journey; key authorization/version integration  live local 10-question/40-option and 180-position/720-bubble sample evaluation; model boundary/recovery tests (docs/TEST-REPORT.md); | DOCX, general key-layout matching, figures/equation fidelity, full document acceptance outstanding; broader field/model evaluation and inference-container qualification remain open. |
| [DOC-05](#doc-05) — Reviewer tools | 2 | building | Private PDF/image intake; durable Poppler/Tesseract extraction; source/text/row review and approval gates ; ordered inline question/option editing, source numbers, reviewed separate-key upload  Local LightOn OCR and pinned original mark-classifier adapters; ordered faculty review, private inference provenance and recovery (docs/CORE-EDUCATION.md);; shared visible processing states, upload steps, recovery and direct upload navigation | TestActualPDFRasterizationAndBounds; import/scan integration tests ; two-column and image-key browser journey; key authorization/version integration  live local 10-question/40-option and 180-position/720-bubble sample evaluation; model boundary/recovery tests (docs/TEST-REPORT.md);; processing-feedback browser transition/retry/mobile/reduced-motion check | DOCX, general key-layout matching, figures/equation fidelity, full document acceptance outstanding; broader field/model evaluation and inference-container qualification remain open. |
| [DOC-06](#doc-06) — Completeness checks | 2 | building | Private PDF/image intake; durable Poppler/Tesseract extraction; source/text/row review and approval gates ; ordered inline question/option editing, source numbers, reviewed separate-key upload  Local LightOn OCR and pinned original mark-classifier adapters; ordered faculty review, private inference provenance and recovery (docs/CORE-EDUCATION.md); | TestActualPDFRasterizationAndBounds; import/scan integration tests ; two-column and image-key browser journey; key authorization/version integration  live local 10-question/40-option and 180-position/720-bubble sample evaluation; model boundary/recovery tests (docs/TEST-REPORT.md); | DOCX, general key-layout matching, figures/equation fidelity, full document acceptance outstanding; broader field/model evaluation and inference-container qualification remain open. |
| [DOC-07](#doc-07) — Publication | 2 | building | Private PDF/image intake; durable Poppler/Tesseract extraction; source/text/row review and approval gates ; ordered inline question/option editing, source numbers, reviewed separate-key upload  Local LightOn OCR and pinned original mark-classifier adapters; ordered faculty review, private inference provenance and recovery (docs/CORE-EDUCATION.md); | TestActualPDFRasterizationAndBounds; import/scan integration tests ; two-column and image-key browser journey; key authorization/version integration  live local 10-question/40-option and 180-position/720-bubble sample evaluation; model boundary/recovery tests (docs/TEST-REPORT.md); | DOCX, general key-layout matching, figures/equation fidelity, full document acceptance outstanding; broader field/model evaluation and inference-container qualification remain open. |
| [DOC-08](#doc-08) — Retry and cost | 2 | building | Private PDF/image intake; durable Poppler/Tesseract extraction; source/text/row review and approval gates ; ordered inline question/option editing, source numbers, reviewed separate-key upload  Local LightOn OCR and pinned original mark-classifier adapters; ordered faculty review, private inference provenance and recovery (docs/CORE-EDUCATION.md);; shared visible processing states, upload steps, recovery and direct upload navigation | TestActualPDFRasterizationAndBounds; import/scan integration tests ; two-column and image-key browser journey; key authorization/version integration  live local 10-question/40-option and 180-position/720-bubble sample evaluation; model boundary/recovery tests (docs/TEST-REPORT.md);; processing-feedback browser transition/retry/mobile/reduced-motion check | DOCX, general key-layout matching, figures/equation fidelity, full document acceptance outstanding; broader field/model evaluation and inference-container qualification remain open. |
| [ASM-01](#asm-01) — Assessment purposes | 2 | building | Ordered pinned paper composition, publishing guards, assignments/windows/extra time, printable booklet/key, evidence UI | TestAcademicOnlineIsolationReplayImmutableAndExpiry; teacher/scan/online browser journey | Multiple-section UI, full blueprints/groups/retakes/tools/conduct and complete publication/scheduling acceptance outstanding |
| [ASM-02](#asm-02) — Composition | 2 | building | Ordered pinned paper composition, publishing guards, assignments/windows/extra time, printable booklet/key, evidence UI | TestAcademicOnlineIsolationReplayImmutableAndExpiry; teacher/scan/online browser journey | Multiple-section UI, full blueprints/groups/retakes/tools/conduct and complete publication/scheduling acceptance outstanding |
| [ASM-03](#asm-03) — Sections and scoring | 2 | building | Ordered pinned paper composition, publishing guards, assignments/windows/extra time, printable booklet/key, evidence UI | TestAcademicOnlineIsolationReplayImmutableAndExpiry; teacher/scan/online browser journey | Multiple-section UI, full blueprints/groups/retakes/tools/conduct and complete publication/scheduling acceptance outstanding |
| [ASM-04](#asm-04) — Lifecycle | 2 | building | Ordered pinned paper composition, publishing guards, assignments/windows/extra time, printable booklet/key, evidence UI | TestAcademicOnlineIsolationReplayImmutableAndExpiry; teacher/scan/online browser journey | Multiple-section UI, full blueprints/groups/retakes/tools/conduct and complete publication/scheduling acceptance outstanding |
| [ASM-05](#asm-05) — Eligibility/recipients | 2 | building | Ordered pinned paper composition, publishing guards, assignments/windows/extra time, printable booklet/key, evidence UI | TestAcademicOnlineIsolationReplayImmutableAndExpiry; teacher/scan/online browser journey | Multiple-section UI, full blueprints/groups/retakes/tools/conduct and complete publication/scheduling acceptance outstanding |
| [ASM-06](#asm-06) — Accommodations | 2 | building | Ordered pinned paper composition, publishing guards, assignments/windows/extra time, printable booklet/key, evidence UI | TestAcademicOnlineIsolationReplayImmutableAndExpiry; teacher/scan/online browser journey | Multiple-section UI, full blueprints/groups/retakes/tools/conduct and complete publication/scheduling acceptance outstanding |
| [ASM-07](#asm-07) — Delivery configuration | 2 | building | Ordered pinned paper composition, publishing guards, assignments/windows/extra time, printable booklet/key, evidence UI | TestAcademicOnlineIsolationReplayImmutableAndExpiry; teacher/scan/online browser journey | Multiple-section UI, full blueprints/groups/retakes/tools/conduct and complete publication/scheduling acceptance outstanding |
| [ASM-08](#asm-08) — Reuse and extra assignment | 2 | building | Ordered pinned paper composition, publishing guards, assignments/windows/extra time, printable booklet/key, evidence UI | TestAcademicOnlineIsolationReplayImmutableAndExpiry; teacher/scan/online browser journey | Multiple-section UI, full blueprints/groups/retakes/tools/conduct and complete publication/scheduling acceptance outstanding |
| [ASM-09](#asm-09) — Staff monitoring | 2 | building | Ordered pinned paper composition, publishing guards, assignments/windows/extra time, printable booklet/key, evidence UI | TestAcademicOnlineIsolationReplayImmutableAndExpiry; teacher/scan/online browser journey | Multiple-section UI, full blueprints/groups/retakes/tools/conduct and complete publication/scheduling acceptance outstanding |
| [ASM-10](#asm-10) — Paper output | 2 | building | Ordered pinned paper composition, publishing guards, assignments/windows/extra time, printable booklet/key, evidence UI | TestAcademicOnlineIsolationReplayImmutableAndExpiry; teacher/scan/online browser journey | Multiple-section UI, full blueprints/groups/retakes/tools/conduct and complete publication/scheduling acceptance outstanding |
| [ASM-11](#asm-11) — Publishing checks | 2 | building | Ordered pinned paper composition, publishing guards, assignments/windows/extra time, printable booklet/key, evidence UI | TestAcademicOnlineIsolationReplayImmutableAndExpiry; teacher/scan/online browser journey | Multiple-section UI, full blueprints/groups/retakes/tools/conduct and complete publication/scheduling acceptance outstanding |
| [SIT-01](#sit-01) — Eligible start | 3 | building | Learner player; exact server deadline; versioned/deduplicated saves; conflict comparison, explicit resume, scheduled expiry and result receipt | TestAcademicOnlineIsolationReplayImmutableAndExpiry; 390px keyboard answer browser journey | Protected refresh/offline persistence, supervised takeover, full palette/conduct/tools/accessibility matrix and acceptance outstanding |
| [SIT-02](#sit-02) — Player | 3 | building | Learner player; exact server deadline; versioned/deduplicated saves; conflict comparison, explicit resume, scheduled expiry and result receipt | TestAcademicOnlineIsolationReplayImmutableAndExpiry; 390px keyboard answer browser journey | Protected refresh/offline persistence, supervised takeover, full palette/conduct/tools/accessibility matrix and acceptance outstanding |
| [SIT-03](#sit-03) — Palette and progress | 3 | building | Learner player; exact server deadline; versioned/deduplicated saves; conflict comparison, explicit resume, scheduled expiry and result receipt | TestAcademicOnlineIsolationReplayImmutableAndExpiry; 390px keyboard answer browser journey | Protected refresh/offline persistence, supervised takeover, full palette/conduct/tools/accessibility matrix and acceptance outstanding |
| [SIT-04](#sit-04) — Authoritative time | 3 | building | Learner player; exact server deadline; versioned/deduplicated saves; conflict comparison, explicit resume, scheduled expiry and result receipt | TestAcademicOnlineIsolationReplayImmutableAndExpiry; 390px keyboard answer browser journey | Protected refresh/offline persistence, supervised takeover, full palette/conduct/tools/accessibility matrix and acceptance outstanding |
| [SIT-05](#sit-05) — Durable saving | 3 | building | Learner player; exact server deadline; versioned/deduplicated saves; conflict comparison, explicit resume, scheduled expiry and result receipt | TestAcademicOnlineIsolationReplayImmutableAndExpiry; 390px keyboard answer browser journey | Protected refresh/offline persistence, supervised takeover, full palette/conduct/tools/accessibility matrix and acceptance outstanding |
| [SIT-06](#sit-06) — Reconnection and concurrent clients | 3 | building | Learner player; exact server deadline; versioned/deduplicated saves; conflict comparison, explicit resume, scheduled expiry and result receipt | TestAcademicOnlineIsolationReplayImmutableAndExpiry; 390px keyboard answer browser journey | Protected refresh/offline persistence, supervised takeover, full palette/conduct/tools/accessibility matrix and acceptance outstanding |
| [SIT-07](#sit-07) — Timing evidence | 3 | planned | None | Not executed | Implementation outstanding |
| [SIT-08](#sit-08) — Submission | 3 | building | Learner player; exact server deadline; versioned/deduplicated saves; conflict comparison, explicit resume, scheduled expiry and result receipt | TestAcademicOnlineIsolationReplayImmutableAndExpiry; 390px keyboard answer browser journey | Protected refresh/offline persistence, supervised takeover, full palette/conduct/tools/accessibility matrix and acceptance outstanding |
| [SIT-09](#sit-09) — Final state | 3 | building | Learner player; exact server deadline; versioned/deduplicated saves; conflict comparison, explicit resume, scheduled expiry and result receipt | TestAcademicOnlineIsolationReplayImmutableAndExpiry; 390px keyboard answer browser journey | Protected refresh/offline persistence, supervised takeover, full palette/conduct/tools/accessibility matrix and acceptance outstanding |
| [SIT-10](#sit-10) — Accessibility and continuity | 3 | building | Learner player; exact server deadline; versioned/deduplicated saves; conflict comparison, explicit resume, scheduled expiry and result receipt | TestAcademicOnlineIsolationReplayImmutableAndExpiry; 390px keyboard answer browser journey | Protected refresh/offline persistence, supervised takeover, full palette/conduct/tools/accessibility matrix and acceptance outstanding |
| [INT-01](#int-01) — Configurable conduct modes | 3 | planned | None | Not executed | Implementation outstanding |
| [INT-02](#int-02) — Browser observations | 3 | planned | None | Not executed | Implementation outstanding |
| [INT-03](#int-03) — Events and deduplication | 3 | planned | None | Not executed | Implementation outstanding |
| [INT-04](#int-04) — Escalation | 3 | planned | None | Not executed | Implementation outstanding |
| [INT-05](#int-05) — Answer secrecy | 3 | building | Unreleased key/rule/rubric/solution stripping, staff-only bank/export, release-checked tutor and formal exposure gate | TestAcademicOnlineIsolationReplayImmutableAndExpiry | Complete multi-entry answer-leak adversarial suite and full exposure/reuse policy acceptance outstanding |
| [INT-06](#int-06) — Security versus monitoring | 3 | planned | None | Not executed | Implementation outstanding |
| [GRD-01](#grd-01) — Objective correctness | 3 | building | Exact selection/decimal/range/tolerance grading; first/best/all/caps; human rubric review, retained result revisions and gated release UI | TestOptionalAccountingAndExactDecimals; TestNumericalAndManualRules; academic integration/browser journeys | Partial credit, corrected paper keys/special cases, appeals/double marking, full reports and complete acceptance outstanding |
| [GRD-02](#grd-02) — Numerical acceptance | 3 | building | Exact selection/decimal/range/tolerance grading; first/best/all/caps; human rubric review, retained result revisions and gated release UI | TestOptionalAccountingAndExactDecimals; TestNumericalAndManualRules; academic integration/browser journeys | Partial credit, corrected paper keys/special cases, appeals/double marking, full reports and complete acceptance outstanding |
| [GRD-03](#grd-03) — Optional groups | 3 | building | Exact selection/decimal/range/tolerance grading; first/best/all/caps; human rubric review, retained result revisions and gated release UI | TestOptionalAccountingAndExactDecimals; TestNumericalAndManualRules; academic integration/browser journeys | Partial credit, corrected paper keys/special cases, appeals/double marking, full reports and complete acceptance outstanding |
| [GRD-04](#grd-04) — Special cases | 3 | planned | None | Not executed | Implementation outstanding |
| [GRD-05](#grd-05) — Subjective marking | 3 | building | Exact selection/decimal/range/tolerance grading; first/best/all/caps; human rubric review, retained result revisions and gated release UI | TestOptionalAccountingAndExactDecimals; TestNumericalAndManualRules; academic integration/browser journeys | Partial credit, corrected paper keys/special cases, appeals/double marking, full reports and complete acceptance outstanding |
| [GRD-06](#grd-06) — Immediate outcomes | 3 | building | Exact selection/decimal/range/tolerance grading; first/best/all/caps; human rubric review, retained result revisions and gated release UI | TestOptionalAccountingAndExactDecimals; TestNumericalAndManualRules; academic integration/browser journeys | Partial credit, corrected paper keys/special cases, appeals/double marking, full reports and complete acceptance outstanding |
| [GRD-07](#grd-07) — Release policy | 3 | building | Exact selection/decimal/range/tolerance grading; first/best/all/caps; human rubric review, retained result revisions and gated release UI | TestOptionalAccountingAndExactDecimals; TestNumericalAndManualRules; academic integration/browser journeys | Partial credit, corrected paper keys/special cases, appeals/double marking, full reports and complete acceptance outstanding |
| [GRD-08](#grd-08) — Learner review | 3 | building | Exact selection/decimal/range/tolerance grading; first/best/all/caps; human rubric review, retained result revisions and gated release UI | TestOptionalAccountingAndExactDecimals; TestNumericalAndManualRules; academic integration/browser journeys | Partial credit, corrected paper keys/special cases, appeals/double marking, full reports and complete acceptance outstanding |
| [GRD-09](#grd-09) — Appeals | 3 | planned | None | Not executed | Implementation outstanding |
| [GRD-10](#grd-10) — Regrading | 3 | building | Exact selection/decimal/range/tolerance grading; first/best/all/caps; human rubric review, retained result revisions and gated release UI | TestOptionalAccountingAndExactDecimals; TestNumericalAndManualRules; academic integration/browser journeys | Partial credit, corrected paper keys/special cases, appeals/double marking, full reports and complete acceptance outstanding |
| [GRD-11](#grd-11) — Rankings | 3 | planned | None | Not executed | Implementation outstanding |
| [PPR-01](#ppr-01) — Layout studio | 4 | building | Original registered layouts; real projective pixel reader; private async batches; uncertainty/identity/correction review; coordinator commit through ordinary marking ; task navigation and original practice kit (docs/CORE-EDUCATION.md) ; calibrated staff diagnostic with overlay/export, 180-position private evaluation  Local LightOn OCR and pinned original mark-classifier adapters; ordered faculty review, private inference provenance and recovery (docs/CORE-EDUCATION.md); | TestRegisteredOMRRealPixelsAndAmbiguity; TestOMRPerspectiveAndExposure; actual PostgreSQL/S3 scan/browser journey ; 10 browser journeys; sample PDF/PG integration (docs/TEST-REPORT.md) ; TestCalibratedGridSeparatesBlankAndMultipleWithoutAKey; authorized image evaluation  live local 10-question/40-option and 180-position/720-bubble sample evaluation; model boundary/recovery tests (docs/TEST-REPORT.md); | Arbitrary formats, real-photo dataset, scored arbitrary-layout calibration/reconciliation, delegated review, quarantine/replacement/retention and complete acceptance outstanding; broader field/model evaluation and inference-container qualification remain open. |
| [PPR-02](#ppr-02) — Format flexibility | 4 | building | Original registered layouts; real projective pixel reader; private async batches; uncertainty/identity/correction review; coordinator commit through ordinary marking ; calibrated staff diagnostic with overlay/export, 180-position private evaluation  Local LightOn OCR and pinned original mark-classifier adapters; ordered faculty review, private inference provenance and recovery (docs/CORE-EDUCATION.md); | TestRegisteredOMRRealPixelsAndAmbiguity; TestOMRPerspectiveAndExposure; actual PostgreSQL/S3 scan/browser journey ; TestCalibratedGridSeparatesBlankAndMultipleWithoutAKey; authorized image evaluation  live local 10-question/40-option and 180-position/720-bubble sample evaluation; model boundary/recovery tests (docs/TEST-REPORT.md); | Arbitrary formats, real-photo dataset, scored arbitrary-layout calibration/reconciliation, delegated review, quarantine/replacement/retention and complete acceptance outstanding; broader field/model evaluation and inference-container qualification remain open. |
| [PPR-03](#ppr-03) — Intake session | 4 | building | Original registered layouts; real projective pixel reader; private async batches; uncertainty/identity/correction review; coordinator commit through ordinary marking | TestRegisteredOMRRealPixelsAndAmbiguity; TestOMRPerspectiveAndExposure; actual PostgreSQL/S3 scan/browser journey | Arbitrary formats, real-photo dataset, calibration/overlays, delegated review, quarantine/replacement/retention and complete acceptance outstanding |
| [PPR-04](#ppr-04) — Recognition and validation | 4 | building | Original registered layouts; real projective pixel reader; private async batches; uncertainty/identity/correction review; coordinator commit through ordinary marking ; task navigation and original practice kit (docs/CORE-EDUCATION.md) ; calibrated staff diagnostic with overlay/export, 180-position private evaluation  Local LightOn OCR and pinned original mark-classifier adapters; ordered faculty review, private inference provenance and recovery (docs/CORE-EDUCATION.md); | TestRegisteredOMRRealPixelsAndAmbiguity; TestOMRPerspectiveAndExposure; actual PostgreSQL/S3 scan/browser journey ; 10 browser journeys; sample PDF/PG integration (docs/TEST-REPORT.md) ; TestCalibratedGridSeparatesBlankAndMultipleWithoutAKey; authorized image evaluation  live local 10-question/40-option and 180-position/720-bubble sample evaluation; model boundary/recovery tests (docs/TEST-REPORT.md); | Arbitrary formats, real-photo dataset, scored arbitrary-layout calibration/reconciliation, delegated review, quarantine/replacement/retention and complete acceptance outstanding; broader field/model evaluation and inference-container qualification remain open. |
| [PPR-05](#ppr-05) — Confidence and uncertainty | 4 | building | Original registered layouts; real projective pixel reader; private async batches; uncertainty/identity/correction review; coordinator commit through ordinary marking ; task navigation and original practice kit (docs/CORE-EDUCATION.md) ; calibrated staff diagnostic with overlay/export, 180-position private evaluation  Local LightOn OCR and pinned original mark-classifier adapters; ordered faculty review, private inference provenance and recovery (docs/CORE-EDUCATION.md); | TestRegisteredOMRRealPixelsAndAmbiguity; TestOMRPerspectiveAndExposure; actual PostgreSQL/S3 scan/browser journey ; 10 browser journeys; sample PDF/PG integration (docs/TEST-REPORT.md) ; TestCalibratedGridSeparatesBlankAndMultipleWithoutAKey; authorized image evaluation  live local 10-question/40-option and 180-position/720-bubble sample evaluation; model boundary/recovery tests (docs/TEST-REPORT.md); | Arbitrary formats, real-photo dataset, scored arbitrary-layout calibration/reconciliation, delegated review, quarantine/replacement/retention and complete acceptance outstanding; broader field/model evaluation and inference-container qualification remain open. |
| [PPR-06](#ppr-06) — Review console | 4 | building | Original registered layouts; real projective pixel reader; private async batches; uncertainty/identity/correction review; coordinator commit through ordinary marking ; task navigation and original practice kit (docs/CORE-EDUCATION.md) ; calibrated staff diagnostic with overlay/export, 180-position private evaluation  Local LightOn OCR and pinned original mark-classifier adapters; ordered faculty review, private inference provenance and recovery (docs/CORE-EDUCATION.md);; shared visible processing states, upload steps, recovery and direct upload navigation | TestRegisteredOMRRealPixelsAndAmbiguity; TestOMRPerspectiveAndExposure; actual PostgreSQL/S3 scan/browser journey ; 10 browser journeys; sample PDF/PG integration (docs/TEST-REPORT.md) ; TestCalibratedGridSeparatesBlankAndMultipleWithoutAKey; authorized image evaluation  live local 10-question/40-option and 180-position/720-bubble sample evaluation; model boundary/recovery tests (docs/TEST-REPORT.md);; processing-feedback browser transition/retry/mobile/reduced-motion check | Arbitrary formats, real-photo dataset, scored arbitrary-layout calibration/reconciliation, delegated review, quarantine/replacement/retention and complete acceptance outstanding; broader field/model evaluation and inference-container qualification remain open. |
| [PPR-07](#ppr-07) — Reviewer scope | 4 | building | Original registered layouts; real projective pixel reader; private async batches; uncertainty/identity/correction review; coordinator commit through ordinary marking | TestRegisteredOMRRealPixelsAndAmbiguity; TestOMRPerspectiveAndExposure; actual PostgreSQL/S3 scan/browser journey | Arbitrary formats, real-photo dataset, calibration/overlays, delegated review, quarantine/replacement/retention and complete acceptance outstanding |
| [PPR-08](#ppr-08) — Evidence audit | 4 | building | Original registered layouts; real projective pixel reader; private async batches; uncertainty/identity/correction review; coordinator commit through ordinary marking | TestRegisteredOMRRealPixelsAndAmbiguity; TestOMRPerspectiveAndExposure; actual PostgreSQL/S3 scan/browser journey | Arbitrary formats, real-photo dataset, calibration/overlays, delegated review, quarantine/replacement/retention and complete acceptance outstanding |
| [PPR-09](#ppr-09) — Duplicate safety | 4 | building | Original registered layouts; real projective pixel reader; private async batches; uncertainty/identity/correction review; coordinator commit through ordinary marking ; task navigation and original practice kit (docs/CORE-EDUCATION.md) | TestRegisteredOMRRealPixelsAndAmbiguity; TestOMRPerspectiveAndExposure; actual PostgreSQL/S3 scan/browser journey ; 10 browser journeys; sample PDF/PG integration (docs/TEST-REPORT.md) | Arbitrary formats, real-photo dataset, calibration/overlays, delegated review, quarantine/replacement/retention and complete acceptance outstanding |
| [PPR-10](#ppr-10) — Commit | 4 | building | Original registered layouts; real projective pixel reader; private async batches; uncertainty/identity/correction review; coordinator commit through ordinary marking ; task navigation and original practice kit (docs/CORE-EDUCATION.md) | TestRegisteredOMRRealPixelsAndAmbiguity; TestOMRPerspectiveAndExposure; actual PostgreSQL/S3 scan/browser journey ; 10 browser journeys; sample PDF/PG integration (docs/TEST-REPORT.md) | Arbitrary formats, real-photo dataset, calibration/overlays, delegated review, quarantine/replacement/retention and complete acceptance outstanding |
| [PPR-11](#ppr-11) — Failure recovery | 4 | building | Original registered layouts; real projective pixel reader; private async batches; uncertainty/identity/correction review; coordinator commit through ordinary marking; shared visible processing states, upload steps, recovery and direct upload navigation | TestRegisteredOMRRealPixelsAndAmbiguity; TestOMRPerspectiveAndExposure; actual PostgreSQL/S3 scan/browser journey; processing-feedback browser transition/retry/mobile/reduced-motion check | Arbitrary formats, real-photo dataset, calibration/overlays, delegated review, quarantine/replacement/retention and complete acceptance outstanding |
| [PPR-12](#ppr-12) — Private access and retention | 4 | building | Original registered layouts; real projective pixel reader; private async batches; uncertainty/identity/correction review; coordinator commit through ordinary marking | TestRegisteredOMRRealPixelsAndAmbiguity; TestOMRPerspectiveAndExposure; actual PostgreSQL/S3 scan/browser journey | Arbitrary formats, real-photo dataset, calibration/overlays, delegated review, quarantine/replacement/retention and complete acceptance outstanding |
| [WRT-01](#wrt-01) — Script intake | 4 | planned | None | Not executed | Implementation outstanding |
| [WRT-02](#wrt-02) — Transcription review | 4 | planned | None | Not executed | Implementation outstanding |
| [WRT-03](#wrt-03) — Rubric assistance | 4 | planned | None | Not executed | Implementation outstanding |
| [WRT-04](#wrt-04) — Human approval | 4 | planned | None | Not executed | Implementation outstanding |
| [WRT-05](#wrt-05) — Student feedback | 4 | planned | None | Not executed | Implementation outstanding |
| [WRT-06](#wrt-06) — Quality evaluation | 4 | planned | None | Not executed | Implementation outstanding |
| [EVD-01](#evd-01) — Explainable error observations | 5 | building | Released topic observations from current result revisions; explicit missing timing and insufficient evidence; staff revision history UI | Marking accounting fixtures; academic integration/browser journey | Longitudinal recalculation/evidence policies, follow-up learning and full acceptance outstanding |
| [EVD-02](#evd-02) — Consistent measures | 5 | building | Released topic observations from current result revisions; explicit missing timing and insufficient evidence; staff revision history UI | Marking accounting fixtures; academic integration/browser journey | Longitudinal recalculation/evidence policies, follow-up learning and full acceptance outstanding |
| [EVD-03](#evd-03) — Topic evidence | 5 | building | Released topic observations from current result revisions; explicit missing timing and insufficient evidence; staff revision history UI | Marking accounting fixtures; academic integration/browser journey | Longitudinal recalculation/evidence policies, follow-up learning and full acceptance outstanding |
| [EVD-04](#evd-04) — Mastery transitions | 5 | planned | None | Not executed | Implementation outstanding |
| [EVD-05](#evd-05) — Behavioral indicators | 5 | planned | None | Not executed | Implementation outstanding |
| [EVD-06](#evd-06) — Missing evidence | 5 | building | Released topic observations from current result revisions; explicit missing timing and insufficient evidence; staff revision history UI ; task navigation and original practice kit (docs/CORE-EDUCATION.md) | Marking accounting fixtures; academic integration/browser journey ; 10 browser journeys; sample PDF/PG integration (docs/TEST-REPORT.md) | Longitudinal recalculation/evidence policies, follow-up learning and full acceptance outstanding |
| [EVD-07](#evd-07) — Focus priorities | 5 | planned | None | Not executed | Implementation outstanding |
| [EVD-08](#evd-08) — Readiness | 5 | planned | None | Not executed | Implementation outstanding |
| [EVD-09](#evd-09) — Historical corrections | 5 | building | Released topic observations from current result revisions; explicit missing timing and insufficient evidence; staff revision history UI | Marking accounting fixtures; academic integration/browser journey | Longitudinal recalculation/evidence policies, follow-up learning and full acceptance outstanding |
| [EVD-10](#evd-10) — Cohort comparisons | 5 | planned | None | Not executed | Implementation outstanding |
| [EVD-11](#evd-11) — Educational policy controls | 5 | planned | None | Not executed | Implementation outstanding |
| [REV-01](#rev-01) — Revision collection | 5 | planned | None | Not executed | Implementation outstanding |
| [REV-02](#rev-02) — States and scheduling | 5 | planned | None | Not executed | Implementation outstanding |
| [REV-03](#rev-03) — Discovery | 5 | planned | None | Not executed | Implementation outstanding |
| [REV-04](#rev-04) — One tutor conversation | 5 | building | Private per-released-question tutor thread, real bounded Gemini/Anthropic adapters, ownership/eligibility/release checks and job retry/cancel UI | Provider HTTP contract tests; released tutor browser surface | Gemini configured privately but live check failed (503/timeouts); evaluations, journal association, abuse budgets and complete tutor acceptance outstanding; not production-verified |
| [REV-05](#rev-05) — Grounded help | 5 | building | Private per-released-question tutor thread, real bounded Gemini/Anthropic adapters, ownership/eligibility/release checks and job retry/cancel UI | Provider HTTP contract tests; released tutor browser surface | Gemini configured privately but live check failed (503/timeouts); evaluations, journal association, abuse budgets and complete tutor acceptance outstanding; not production-verified |
| [REV-06](#rev-06) — Interaction quality | 5 | building | Private per-released-question tutor thread, real bounded Gemini/Anthropic adapters, ownership/eligibility/release checks and job retry/cancel UI | Provider HTTP contract tests; released tutor browser surface | Gemini configured privately but live check failed (503/timeouts); evaluations, journal association, abuse budgets and complete tutor acceptance outstanding; not production-verified |
| [REV-07](#rev-07) — Release enforcement | 5 | building | Private per-released-question tutor thread, real bounded Gemini/Anthropic adapters, ownership/eligibility/release checks and job retry/cancel UI | Provider HTTP contract tests; released tutor browser surface | Gemini configured privately but live check failed (503/timeouts); evaluations, journal association, abuse budgets and complete tutor acceptance outstanding; not production-verified |
| [REV-08](#rev-08) — Abuse/privacy controls | 5 | building | Private per-released-question tutor thread, real bounded Gemini/Anthropic adapters, ownership/eligibility/release checks and job retry/cancel UI | Provider HTTP contract tests; released tutor browser surface | Gemini configured privately but live check failed (503/timeouts); evaluations, journal association, abuse budgets and complete tutor acceptance outstanding; not production-verified |
| [REV-09](#rev-09) — Teacher escalation | 5 | planned | None | Not executed | Implementation outstanding |
| [PLN-01](#pln-01) — Goals | 5 | planned | None | Not executed | Implementation outstanding |
| [PLN-02](#pln-02) — Calendar | 5 | planned | None | Not executed | Implementation outstanding |
| [PLN-03](#pln-03) — Instruction design | 5 | planned | None | Not executed | Implementation outstanding |
| [PLN-04](#pln-04) — Personal context | 5 | planned | None | Not executed | Implementation outstanding |
| [PLN-05](#pln-05) — Consumption | 5 | planned | None | Not executed | Implementation outstanding |
| [PLN-06](#pln-06) — Completion and understanding | 5 | planned | None | Not executed | Implementation outstanding |
| [PLN-07](#pln-07) — Exercise responses | 5 | planned | None | Not executed | Implementation outstanding |
| [PLN-08](#pln-08) — Content reuse/versioning | 5 | planned | None | Not executed | Implementation outstanding |
| [PLN-09](#pln-09) — Instructor assignments | 5 | planned | None | Not executed | Implementation outstanding |
| [PLN-10](#pln-10) — Adaptive papers | 5 | planned | None | Not executed | Implementation outstanding |
| [PLN-11](#pln-11) — Thin-bank behavior | 5 | planned | None | Not executed | Implementation outstanding |
| [PLN-12](#pln-12) — Live adaptive practice | 5 | planned | None | Not executed | Implementation outstanding |
| [PLN-13](#pln-13) — Continuity | 5 | planned | None | Not executed | Implementation outstanding |
| [PLN-14](#pln-14) — Motivation | 5 | planned | None | Not executed | Implementation outstanding |
| [STU-01](#stu-01) — Resource library | 6 | planned | None | Not executed | Implementation outstanding |
| [STU-02](#stu-02) — Content workflows | 6 | planned | None | Not executed | Implementation outstanding |
| [STU-03](#stu-03) — Generated material | 6 | planned | None | Not executed | Implementation outstanding |
| [STU-04](#stu-04) — Narration | 6 | planned | None | Not executed | Implementation outstanding |
| [STU-05](#stu-05) — Presentation-to-video | 6 | planned | None | Not executed | Implementation outstanding |
| [STU-06](#stu-06) — Presenter generation | 6 | planned | None | Not executed | Implementation outstanding |
| [STU-07](#stu-07) — Captions and exports | 6 | planned | None | Not executed | Implementation outstanding |
| [STU-08](#stu-08) — Concept animations | 6 | planned | None | Not executed | Implementation outstanding |
| [STU-09](#stu-09) — Job operations | 6 | planned | None | Not executed | Implementation outstanding |
| [STU-10](#stu-10) — Topic delivery controls | 6 | planned | None | Not executed | Implementation outstanding |
| [STU-11](#stu-11) — External resources | 6 | planned | None | Not executed | Implementation outstanding |
| [TCH-01](#tch-01) — Teaching home | 5 | planned | None | Not executed | Implementation outstanding |
| [TCH-02](#tch-02) — Learner detail | 5 | planned | None | Not executed | Implementation outstanding |
| [TCH-03](#tch-03) — Class insight | 5 | planned | None | Not executed | Implementation outstanding |
| [TCH-04](#tch-04) — Interventions | 5 | planned | None | Not executed | Implementation outstanding |
| [TCH-05](#tch-05) — Recommendations | 5 | planned | None | Not executed | Implementation outstanding |
| [TCH-06](#tch-06) — Teaching outcomes | 5 | planned | None | Not executed | Implementation outstanding |
| [TCH-07](#tch-07) — Professional directory | 5 | planned | None | Not executed | Implementation outstanding |
| [TCH-08](#tch-08) — Roster and review tasks | 5 | planned | None | Not executed | Implementation outstanding |
| [FAM-01](#fam-01) — Child workspace | 7 | planned | None | Not executed | Implementation outstanding |
| [FAM-02](#fam-02) — Useful reports | 7 | planned | None | Not executed | Implementation outstanding |
| [FAM-03](#fam-03) — Communication | 7 | planned | None | Not executed | Implementation outstanding |
| [FAM-04](#fam-04) — Consent and preferences | 7 | planned | None | Not executed | Implementation outstanding |
| [FAM-05](#fam-05) — Engagement | 7 | planned | None | Not executed | Implementation outstanding |
| [FAM-06](#fam-06) — Print/download | 7 | planned | None | Not executed | Implementation outstanding |
| [ATT-01](#att-01) — Attendance capture | 7 | planned | None | Not executed | Implementation outstanding |
| [ATT-02](#att-02) — Timetables | 7 | planned | None | Not executed | Implementation outstanding |
| [ATT-03](#att-03) — Follow-up | 7 | planned | None | Not executed | Implementation outstanding |
| [ATT-04](#att-04) — Analysis | 7 | planned | None | Not executed | Implementation outstanding |
| [ATT-05](#att-05) — Activity definitions | 7 | planned | None | Not executed | Implementation outstanding |
| [COM-01](#com-01) — Notification inbox | 7 | planned | None | Not executed | Implementation outstanding |
| [COM-02](#com-02) — Delivery channels | 7 | building | Real STARTTLS SMTP adapter; local Mailpit acceptance and invitation/verification messages | TestSMTPAcceptanceAndExpiredLeaseRecovery; captured-email browser journey | Live SMTP, SMS, push and actual-delivery qualification outstanding |
| [COM-03](#com-03) — Preferences | 7 | planned | None | Not executed | Implementation outstanding |
| [COM-04](#com-04) — Messaging | 7 | planned | None | Not executed | Implementation outstanding |
| [COM-05](#com-05) — Class discussions | 7 | planned | None | Not executed | Implementation outstanding |
| [COM-06](#com-06) — Announcements | 7 | planned | None | Not executed | Implementation outstanding |
| [COM-07](#com-07) — Templates and retries | 7 | building | Transactional invitation/reset queue and owner retry UI | Invitation transaction test | Templates, localization, complete deduplication/recovery outstanding |
| [COM-08](#com-08) — Meetings | 7 | planned | None | Not executed | Implementation outstanding |
| [COM-09](#com-09) — Support desk | 7 | planned | None | Not executed | Implementation outstanding |
| [RPT-01](#rpt-01) — Personal insight | 7 | planned | None | Not executed | Implementation outstanding |
| [RPT-02](#rpt-02) — Institution oversight | 7 | planned | None | Not executed | Implementation outstanding |
| [RPT-03](#rpt-03) — Filters and comparisons | 7 | planned | None | Not executed | Implementation outstanding |
| [RPT-04](#rpt-04) — Report builder | 7 | planned | None | Not executed | Implementation outstanding |
| [RPT-05](#rpt-05) — Export center | 7 | planned | None | Not executed | Implementation outstanding |
| [RPT-06](#rpt-06) — Scheduled reporting | 7 | planned | None | Not executed | Implementation outstanding |
| [RPT-07](#rpt-07) — Custom dashboards | 7 | planned | None | Not executed | Implementation outstanding |
| [RPT-08](#rpt-08) — Leaderboards and benchmarking | 7 | planned | None | Not executed | Implementation outstanding |
| [RPT-09](#rpt-09) — Evidence traceability | 7 | planned | None | Not executed | Implementation outstanding |
| [RPT-10](#rpt-10) — Readable sharing | 7 | planned | None | Not executed | Implementation outstanding |
| [INS-01](#ins-01) — Performance trajectories | 8 | planned | None | Not executed | Implementation outstanding |
| [INS-02](#ins-02) — Rank forecasting | 8 | planned | None | Not executed | Implementation outstanding |
| [INS-03](#ins-03) — What-if tools | 8 | planned | None | Not executed | Implementation outstanding |
| [INS-04](#ins-04) — Confidence and reflection | 8 | planned | None | Not executed | Implementation outstanding |
| [INS-05](#ins-05) — Supportive coaching | 8 | planned | None | Not executed | Implementation outstanding |
| [INS-06](#ins-06) — Human attention flags | 8 | planned | None | Not executed | Implementation outstanding |
| [INS-07](#ins-07) — Evaluation workspace | 8 | planned | None | Not executed | Implementation outstanding |
| [INS-08](#ins-08) — Production gates | 8 | planned | None | Not executed | Implementation outstanding |
| [FIN-01](#fin-01) — Plans and entitlements | 8 | planned | None | Not executed | Implementation outstanding |
| [FIN-02](#fin-02) — Checkout/payment | 8 | planned | None | Not executed | Implementation outstanding |
| [FIN-03](#fin-03) — Financial records | 8 | planned | None | Not executed | Implementation outstanding |
| [FIN-04](#fin-04) — History and billing UI | 8 | planned | None | Not executed | Implementation outstanding |
| [FIN-05](#fin-05) — Refunds/cancellation | 8 | planned | None | Not executed | Implementation outstanding |
| [FIN-06](#fin-06) — Webhook/reconciliation | 8 | planned | None | Not executed | Implementation outstanding |
| [FIN-07](#fin-07) — Budgets and usage | 8 | planned | None | Not executed | Implementation outstanding |
| [FIN-08](#fin-08) — Financial insight | 8 | planned | None | Not executed | Implementation outstanding |
| [FIN-09](#fin-09) — Provider setup and access | 8 | planned | None | Not executed | Implementation outstanding |
| [UX-01](#ux-01) — Public site | 1 | building | Original public introduction, sign-in, help and development policies | Next production build | Public feature/about/contact/pricing workflows outstanding |
| [UX-02](#ux-02) — Honest content | 1 | planned | None | Not executed | Implementation outstanding |
| [UX-03](#ux-03) — Original visual system | 1 | building | DESIGN.md; shared tokens; theme and text size controls ; task navigation and original practice kit (docs/CORE-EDUCATION.md) ; reference-guided light studio, original SVG artwork, assessment cards, plain-language service recovery; shared visible processing states, upload steps, recovery and direct upload navigation | Build; narrow-screen browser test ; 10 browser journeys; sample PDF/PG integration (docs/TEST-REPORT.md); processing-feedback browser transition/retry/mobile/reduced-motion check | Full visual state audit outstanding |
| [UX-04](#ux-04) — Responsive behavior | 1 | building | Responsive work forms, scrollable tables and print rules ; task navigation and original practice kit (docs/CORE-EDUCATION.md) ; reference-guided light studio, original SVG artwork, assessment cards, plain-language service recovery; shared visible processing states, upload steps, recovery and direct upload navigation | Narrow-screen browser test ; 10 browser journeys; sample PDF/PG integration (docs/TEST-REPORT.md); processing-feedback browser transition/retry/mobile/reduced-motion check | All required device/workflow qualification outstanding |
| [UX-05](#ux-05) — Accessibility | 1 | building | Semantic labels, focus and keyboard forms ; task navigation and original practice kit (docs/CORE-EDUCATION.md) ; reference-guided light studio, original SVG artwork, assessment cards, plain-language service recovery; shared visible processing states, upload steps, recovery and direct upload navigation | Browser label defects corrected ; 10 browser journeys; sample PDF/PG integration (docs/TEST-REPORT.md); processing-feedback browser transition/retry/mobile/reduced-motion check | WCAG 2.2 AA audit and complete journeys outstanding |
| [UX-06](#ux-06) — Localization | 1 | building | English/Hindi workspace dictionary and timezone dates | Hindi narrow-screen browser test | Complete translations outstanding |
| [UX-07](#ux-07) — State quality | 1 | building | Loading/empty/error/saved states; request cancellation ; task navigation and original practice kit (docs/CORE-EDUCATION.md) ; reference-guided light studio, original SVG artwork, assessment cards, plain-language service recovery | Browser workflow ; 10 browser journeys; sample PDF/PG integration (docs/TEST-REPORT.md) | Offline persistence and all feature states outstanding |
| [UX-08](#ux-08) — Discoverability | 1 | building | Institution/campus selector, people search, access help ; task navigation and original practice kit (docs/CORE-EDUCATION.md) | Browser workflow ; 10 browser journeys; sample PDF/PG integration (docs/TEST-REPORT.md) | Global search, child switcher and complete cross-workspace deep links outstanding |
| [UX-09](#ux-09) — Public performance/metadata | 1 | planned | None | Not executed | Implementation outstanding |
| [MOB-01](#mob-01) — Native applications | 9 | planned | None | Not executed | Implementation outstanding |
| [MOB-02](#mob-02) — Secure local state | 9 | planned | None | Not executed | Implementation outstanding |
| [MOB-03](#mob-03) — Offline study | 9 | planned | None | Not executed | Implementation outstanding |
| [MOB-04](#mob-04) — Synchronization | 9 | planned | None | Not executed | Implementation outstanding |
| [MOB-05](#mob-05) — Assessment network policy | 9 | planned | None | Not executed | Implementation outstanding |
| [MOB-06](#mob-06) — Push/deep links | 9 | planned | None | Not executed | Implementation outstanding |
| [MOB-07](#mob-07) — Web installation | 9 | planned | None | Not executed | Implementation outstanding |
| [MOB-08](#mob-08) — Qualification | 9 | planned | None | Not executed | Implementation outstanding |
| [AIO-01](#aio-01) — Capability adapters | 5 | building | Actual Gemini/Anthropic question/tutor adapters, Go draft validation, generation usage fields, durable jobs and recovery UI; opt-in labelled development fixture; Local LightOn OCR and pinned original mark-classifier adapters; ordered faculty review, private inference provenance and recovery (docs/CORE-EDUCATION.md); | TestModelContractKeepsOutputDraftAndReportsProviderFailures; TestAcademicGenerationRequiresConfigurationAndReviewBeforeScoring  live local 10-question/40-option and 180-position/720-bubble sample evaluation; model boundary/recovery tests (docs/TEST-REPORT.md); | Gemini live check returned HTTP 503/timeouts; other provider adapters, full budgets/consumption/evaluation and acceptance outstanding; broader field/model evaluation and inference-container qualification remain open. |
| [AIO-02](#aio-02) — Instruction management | 5 | planned | None | Not executed | Implementation outstanding |
| [AIO-03](#aio-03) — Output validation | 5 | building | Actual Gemini/Anthropic question/tutor adapters, Go draft validation, generation usage fields, durable jobs and recovery UI; opt-in labelled development fixture; Local LightOn OCR and pinned original mark-classifier adapters; ordered faculty review, private inference provenance and recovery (docs/CORE-EDUCATION.md); | TestModelContractKeepsOutputDraftAndReportsProviderFailures; TestAcademicGenerationRequiresConfigurationAndReviewBeforeScoring  live local 10-question/40-option and 180-position/720-bubble sample evaluation; model boundary/recovery tests (docs/TEST-REPORT.md); | Gemini live check returned HTTP 503/timeouts; other provider adapters, full budgets/consumption/evaluation and acceptance outstanding; broader field/model evaluation and inference-container qualification remain open. |
| [AIO-04](#aio-04) — Reuse/invalidation | 5 | planned | None | Not executed | Implementation outstanding |
| [AIO-05](#aio-05) — Complete consumption accounting | 5 | building | Actual Gemini/Anthropic question/tutor adapters, Go draft validation, generation usage fields, durable jobs and recovery UI; opt-in labelled development fixture | TestModelContractKeepsOutputDraftAndReportsProviderFailures; TestAcademicGenerationRequiresConfigurationAndReviewBeforeScoring | Gemini live check returned HTTP 503/timeouts; other provider adapters, full budgets/consumption/evaluation and acceptance outstanding |
| [AIO-06](#aio-06) — Budgets/limits | 5 | planned | None | Not executed | Implementation outstanding |
| [AIO-07](#aio-07) — Resilience | 5 | building | Actual Gemini/Anthropic question/tutor adapters, Go draft validation, generation usage fields, durable jobs and recovery UI; opt-in labelled development fixture; Local LightOn OCR and pinned original mark-classifier adapters; ordered faculty review, private inference provenance and recovery (docs/CORE-EDUCATION.md);; shared visible processing states, upload steps, recovery and direct upload navigation | TestModelContractKeepsOutputDraftAndReportsProviderFailures; TestAcademicGenerationRequiresConfigurationAndReviewBeforeScoring  live local 10-question/40-option and 180-position/720-bubble sample evaluation; model boundary/recovery tests (docs/TEST-REPORT.md);; processing-feedback browser transition/retry/mobile/reduced-motion check | Gemini live check returned HTTP 503/timeouts; other provider adapters, full budgets/consumption/evaluation and acceptance outstanding; broader field/model evaluation and inference-container qualification remain open. |
| [AIO-08](#aio-08) — Quality and feedback | 5 | building | Local LightOn OCR and pinned original mark-classifier adapters; ordered faculty review, private inference provenance and recovery (docs/CORE-EDUCATION.md); | Live local 10-question/40-option and 180-position/720-bubble sample evaluation; model boundary/recovery tests (docs/TEST-REPORT.md); | Implementation outstanding; broader field/model evaluation and inference-container qualification remain open. |
| [AIO-09](#aio-09) — Data boundaries | 5 | building | Actual Gemini/Anthropic question/tutor adapters, Go draft validation, generation usage fields, durable jobs and recovery UI; opt-in labelled development fixture | TestModelContractKeepsOutputDraftAndReportsProviderFailures; TestAcademicGenerationRequiresConfigurationAndReviewBeforeScoring | Gemini live check returned HTTP 503/timeouts; other provider adapters, full budgets/consumption/evaluation and acceptance outstanding |
| [AIO-10](#aio-10) — Operator console | 5 | building | Actual Gemini/Anthropic question/tutor adapters, Go draft validation, generation usage fields, durable jobs and recovery UI; opt-in labelled development fixture; Local LightOn OCR and pinned original mark-classifier adapters; ordered faculty review, private inference provenance and recovery (docs/CORE-EDUCATION.md);; shared visible processing states, upload steps, recovery and direct upload navigation | TestModelContractKeepsOutputDraftAndReportsProviderFailures; TestAcademicGenerationRequiresConfigurationAndReviewBeforeScoring  live local 10-question/40-option and 180-position/720-bubble sample evaluation; model boundary/recovery tests (docs/TEST-REPORT.md);; processing-feedback browser transition/retry/mobile/reduced-motion check | Gemini live check returned HTTP 503/timeouts; other provider adapters, full budgets/consumption/evaluation and acceptance outstanding; broader field/model evaluation and inference-container qualification remain open. |
| [RUN-01](#run-01) — Durable execution | 1 | building | River academic jobs with transactional enqueue/schedule/retry/cancel and recovery UI; existing SMTP queue retained | Invitation handoff test | General worker-crash drills, full operator recovery/retention and complete durability acceptance outstanding |
| [RUN-02](#run-02) — Operational console | 1 | building | Invitation delivery list and authorized retry UI; shared visible processing states, upload steps, recovery and direct upload navigation | Permission tests; browser pending; processing-feedback browser transition/retry/mobile/reduced-motion check | Complete operational console outstanding |
| [RUN-03](#run-03) — File lifecycle | 1 | building | Private bounded profile-image intake, staged/active/retired/purged records and recovery UI | Actual PostgreSQL/SeaweedFS authorization, failure and lifecycle test | General document/media quarantine, resumable upload, provenance and retention controls outstanding |
| [RUN-04](#run-04) — Audit | 1 | building | Transactionally inserted audit, bounded institution audit UI | TestScopedCoordinatorAndAuditAtomicity | Filters, export, full audit coverage and database role protection outstanding |
| [RUN-05](#run-05) — Application security | 1 | building | CSRF/session/input/tenant controls plus bounded private file access | Isolation/revocation/S3 negative tests; vulnerability scan | Full security audit, webhook and export controls outstanding |
| [RUN-06](#run-06) — Privacy/consent | 1 | planned | None | Not executed | Implementation outstanding |
| [RUN-07](#run-07) — Resilience/health | 1 | building | Liveness/readiness and structured request logging | Local startup | Metrics, traces, alerting and operator dashboard outstanding |
| [RUN-08](#run-08) — Backup and recovery | 1 | planned | None | Not executed | Implementation outstanding |
| [RUN-09](#run-09) — Environment/deployment | 1 | building | Local Compose, API/worker/web container builds, migrations, environment validation and CI source | Bun startup; Docker API and web builds; unit/integration/browser checks | Hosted CI execution, production deployment, maintenance and rollback qualification outstanding |
| [RUN-10](#run-10) — Performance/concurrency | 1 | planned | None | Not executed | Implementation outstanding |
| [RUN-11](#run-11) — Support recovery | 1 | planned | None | Not executed | Implementation outstanding |
| [RUN-12](#run-12) — Data portability | 1 | planned | None | Not executed | Implementation outstanding |
| [RUN-13](#run-13) — Delivery evidence | 1 | building | Complete register, test report, operations/provenance documentation | Register count 233 | Release evidence incomplete |

## Expanded acceptance checklists

Every source detail is retained below. Checklist clauses group closely related behavior; each listed subfeature must be tested before checking a clause. Cross-cutting completion checks apply to every requirement.

### ID-01

Registration and invitations · Milestone 1

- [ ] Support learner registration and staff/guardian invitations, institution enrollment codes, verified email, optional phone verification, and explicit approval where the institution requires it.
- [ ] Collect only needed personal/academic information
- [ ] record terms/privacy acknowledgments with policy version and time.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### ID-02

Credential sign-in · Milestone 1

- [ ] Normalize email, validate inputs, offer password visibility and clear accessible validation, support persistent versus short sessions, and route the user to the correct permitted workspace.
- [ ] Do not reveal account existence through recovery responses.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### ID-03

Password recovery/change · Milestone 1

- [ ] Provide expiring single-use reset links/codes, authenticated password change, verified email-change confirmation, and relevant security notifications.
- [ ] Revoke sessions according to the recovery policy.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### ID-04

MFA · Milestone 1

- [ ] Implement authenticator-based MFA, recovery codes, enrollment confirmation, lost-device recovery, and organization-level enforcement for privileged staff.
- [ ] Offer email/SMS factors only when their provider and policy are configured.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### ID-05

Session management · Milestone 1

- [ ] Show current and other signed-in devices/sessions, last activity and approximate session context
- [ ] revoke one or all.
- [ ] Rotate/reissue safely and expire idle/absolute sessions.
- [ ] Logout invalidates the applicable server session.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### ID-06

Account state · Milestone 1

- [ ] Support invited, active, suspended, and archived access states, with reason, actor, time, and recovery workflow.
- [ ] Suspension immediately invalidates access while preserving lawful academic/financial history.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### ID-07

Abuse controls · Milestone 1

- [ ] Rate-limit sign-in, recovery, registration, verification, invitations, uploads, and paid interactions across all application instances.
- [ ] Avoid locking an entire institution behind a shared network address because one person makes mistakes.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### ID-08

Profile · Milestone 1

- [ ] Edit permitted name/contact/academic fields and a cropped profile image
- [ ] retain verified identifiers and staff-controlled fields.
- [ ] Support emergency contacts and user-facing account activity history.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### ID-09

Safe bootstrap · Milestone 1

- [ ] Create the first organization owner through a secure one-time operator process.
- [ ] Development fixtures must be synthetic and unusable as production credentials.
- [ ] Provide weak/compromised credential remediation.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### ORG-01

Institution lifecycle · Milestone 1

- [ ] Create, configure, activate, suspend, and close organizations with branding, timezone, locale, support contacts, academic-year settings, and subscription entitlements.
- [ ] Separate organization closure from deletion of retained records.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### ORG-02

Campus management · Milestone 1

- [ ] Maintain named campuses with address, codes, contacts, responsible staff and active dates.
- [ ] Support staff responsible for more than one campus, institution-wide oversight, and scoped reassignment with audit.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### ORG-03

Courses and groups · Milestone 1

- [ ] Manage programs, academic years, grades, courses, cohorts, class groups, sections, and teaching timetables.
- [ ] A batch date is not a replacement for managed class membership.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### ORG-04

Learner admission · Milestone 1

- [ ] Enroll singly or through validated CSV/XLSX import, with preview, duplicates, row errors, eligibility/grade, target exams, campus, cohort, start/end dates, and admission state.
- [ ] Support an intake queue awaiting placement.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### ORG-05

Human references · Milestone 1

- [ ] Allocate unique readable learner and paper reference codes, allow appropriate externally issued identifiers, preserve aliases after deliberate corrections, prevent reuse/races, and support manual lookup.
- [ ] Do not encode mutable campus or role state into permanent identity.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### ORG-06

Transfers and withdrawal · Milestone 1

- [ ] Transfer campus/class/course, preserve academic history and identifier continuity, update affected staff scope, and record effective dates.
- [ ] Withdrawal ends access/assignments according to policy without destroying evidence.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### ORG-07

Guardian relationships · Milestone 1

- [ ] Invite and verify guardians, link multiple children and multiple guardians, distinguish relationship/permission level, approve/remove links, and show an unambiguous child switcher.
- [ ] Handle custody/contact restrictions through explicit institutional controls.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### ORG-08

Staff and teaching assignments · Milestone 1

- [ ] Create/edit staff professional profiles, subjects, qualifications/experience, campus/class responsibilities and effective periods.
- [ ] Replace/add/remove teaching rosters through searchable selectors without losing selections across pagination.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### ORG-09

Permission administration · Milestone 1

- [ ] Provide role templates, a capability matrix, custom scoped roles, grant/revoke workflows, protected owner transfer, and bulk changes with preview.
- [ ] Prevent self-escalation and orphaned ownership.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### ORG-10

Import/reconciliation tools · Milestone 1

- [ ] Reconcile duplicate people and conflicting external IDs without silently merging unrelated records.
- [ ] Offer dry-run reports, explicit resolution, reversible administrative corrections, and an audit trail.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### CAT-01

Curriculum models · Milestone 2

- [ ] Maintain versioned curricula, exams, subjects, chapters, topics/subtopics, outcomes, prerequisite relationships, language and academic level.
- [ ] Seed original representative mathematics, physics, chemistry, and biology content
- [ ] allow institution-defined subjects.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### CAT-02

Question types · Milestone 2

- [ ] Support single selection, multiple selection, integer/decimal numerical, short written response, and extended written response.
- [ ] Provide rubric-based manual marking for written types and extensible type contracts without inventing a general plugin system.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### CAT-03

Rich question content · Milestone 2

- [ ] Author readable text, mathematical/chemical notation, diagrams, tables, attachments, image-only stems, option images, alt text, and worked solutions.
- [ ] Rendering must agree in authoring preview, online/native player, paper export, review, and tutor context.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### CAT-04

Academic metadata · Milestone 2

- [ ] Track intended outcome/topic, difficulty, source/license/provenance, language, estimated time, tags, default marks, penalties, numeric acceptance and accessibility alternatives.
- [ ] Permit weighted multi-topic attribution when explicitly authored.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### CAT-05

Lifecycle and revision · Milestone 2

- [ ] Support draft, review, approved, retired and archived content with history, reviewer comments, comparison and restoration as a new revision.
- [ ] A correction cannot rewrite an already delivered question snapshot.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### CAT-06

Bank management · Milestone 2

- [ ] Search/filter/preview, bulk tag/move/archive, edit, import/export selected records, inspect use history, and report deficient topic/difficulty coverage.
- [ ] Protect questions referenced by live or past assessments from destructive deletion.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### CAT-07

Duplicate management · Milestone 2

- [ ] Detect exact and likely duplicate text/images with explainable matches
- [ ] provide human merge/keep decisions that preserve references and original authorship.
- [ ] Do not silently replace similar scientific questions.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### CAT-08

Bulk authoring/import · Milestone 2

- [ ] Support structured JSON, CSV, and XLSX import with type/answer alias normalization, image bundles, validation preview, row errors, repeat-import safety, and explicit handling of unknown answers.
- [ ] Incomplete questions cannot be published into scored work.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### CAT-09

Review and provenance · Milestone 2

- [ ] Distinguish generated, imported, authored, and externally licensed content.
- [ ] Require staff approval before AI-generated questions/solutions become scored material.
- [ ] Record source attribution and permitted reuse.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### CAT-10

Quality and item analysis · Milestone 2

- [ ] Show total/draft/approved/retired counts, topic/difficulty/type distribution, bank depth, freshness, exposure, observed facility/difficulty, distractor selection, discrimination where statistically supported, and flagged quality concerns.
- [ ] Provide evidence counts and avoid treating sparse statistics as authoritative.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### DOC-01

Source intake · Milestone 2

- [ ] Accept image photographs, scanned/text PDFs, and DOCX papers, plus separate answer-key files and worked-solution documents.
- [ ] Validate content type, size, page limits, encryption, supported languages and readability.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### DOC-02

Durable extraction workspace · Milestone 2

- [ ] Create a persistent import job with source previews, page-level progress, recoverable drafts, cancellation and resume.
- [ ] Closing the browser must not lose the import or its reviewed edits.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### DOC-03

Faithful extraction · Milestone 2

- [ ] Recover complete stems, options, figures, option images, tables, equations, printed question numbers, section instructions, and supplied solutions.
- [ ] Preserve provenance/crop/page coordinates
- [ ] flag missing or uncertain content instead of inventing it.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### DOC-04

Key matching · Milestone 2

- [ ] Extract separate or embedded keys and match by paper revision, section and printed number.
- [ ] Report unmatched, duplicated, conflicting, multi-answer, numerical and withdrawn-question entries for review.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### DOC-05

Reviewer tools · Milestone 2

- [ ] Edit wording/figures/options, keys and rubric, assign topic/difficulty, apply tags/rules to ranges, compare original and draft, and approve pages/questions.
- [ ] Support keyboard navigation and autosaved reviewer changes.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### DOC-06

Completeness checks · Milestone 2

- [ ] Reconcile declared versus extracted question counts, section boundaries, numbering gaps, cross-page continuations, repeated pages and option completeness.
- [ ] Partial success remains usable but cannot be silently advertised as complete.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### DOC-07

Publication · Milestone 2

- [ ] Validate every mandatory topic, key/rubric and scoring instruction before publishing a versioned assessment or selected bank entries.
- [ ] Preserve the exact intended order and retain the reviewed source record.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### DOC-08

Retry and cost · Milestone 2

- [ ] Retry only failed extraction work without discarding reviewed corrections or paying again for completed work unnecessarily.
- [ ] Show consumption and useful quota/configuration failures to authorized operators.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### ASM-01

Assessment purposes · Milestone 2

- [ ] Support baseline diagnostics, subject quizzes, practice, mock exams, formal exams, homework and adaptive practice.
- [ ] Keep purpose separate from delivery channel and grading/release policy.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### ASM-02

Composition · Milestone 2

- [ ] Select explicit questions in order or generate from an approved blueprint using subjects/topics, count, difficulty, coverage, expected duration and exposure constraints.
- [ ] Preview the actual paper, not just a claimed count.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### ASM-03

Sections and scoring · Milestone 2

- [ ] Configure sections, marks/penalties, maximum contribution, optional groups, numerical rules, partial-credit policies and rubrics.
- [ ] Derive total maximum marks from the selected rules and validate impossible configurations.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### ASM-04

Lifecycle · Milestone 2

- [ ] Draft, review, publish, schedule, open, close and archive
- [ ] duplicate a paper as a new revision
- [ ] distinguish corrections from new papers.
- [ ] Pin a published revision for each assigned sitting.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### ASM-05

Eligibility/recipients · Milestone 2

- [ ] Assign individually or by course/class/cohort, with searchable bulk selection, exclusions, due dates, windows, timezones, attempt counts and prerequisites.
- [ ] Freeze or explicitly manage later group-membership changes.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### ASM-06

Accommodations · Milestone 2

- [ ] Set extra time, approved breaks, accessible presentation, alternative delivery, conduct exemptions, and individual scheduling.
- [ ] Approved accommodations affect the actual authoritative policy and are not misconduct.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### ASM-07

Delivery configuration · Milestone 2

- [ ] Choose online, paper, or mixed delivery, fixed order versus seeded randomization, section timing, permitted calculator/reference material, and conduct policy.
- [ ] Prevent inconsistent settings between instructions and player behavior.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### ASM-08

Reuse and extra assignment · Milestone 2

- [ ] Add eligible recipients to a published paper without duplicates
- [ ] support reassignment/retakes under explicit version and result-selection policy.
- [ ] Preserve who assigned what and when.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### ASM-09

Staff monitoring · Milestone 2

- [ ] List/search papers and show assigned/not started/live/submitted/awaiting review/released counts, deadlines, aggregate results, per-recipient state and actionable exceptions.
- [ ] Respect live-answer visibility policy.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### ASM-10

Paper output · Milestone 2

- [ ] Export printable question booklets, marking instructions, approved answer keys and compatible response sheets, with clear version/reference identity and accessible alternatives.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### ASM-11

Publishing checks · Milestone 2

- [ ] Detect missing keys, insufficient pool, duplicate questions, missing figure assets, invalid optional counts, unreconciled section totals, and broken schedules before delivery.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### SIT-01

Eligible start · Milestone 3

- [ ] Show instructions, applicable accessibility/conduct requirements, technical readiness and release policy.
- [ ] Start only an eligible assigned/allowed paper within its window
- [ ] resume the authorized sitting rather than creating accidental duplicates.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### SIT-02

Player · Milestone 3

- [ ] Provide readable rich questions, figures/options, numerical/written answer inputs, section context, marks/penalties, approved tools, zoom, and accessible navigation.
- [ ] Save answer, clear, move back/forward, jump, bookmark/mark for review and filter the palette.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### SIT-03

Palette and progress · Milestone 3

- [ ] Distinguish unseen, visited blank, answered, marked for review, and answered-and-marked states.
- [ ] Show section/optional-group completion, time remaining and meaningful approaching-deadline warnings.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### SIT-04

Authoritative time · Milestone 3

- [ ] Establish deadlines from server time and approved allowances, independent of client clock and browser presence.
- [ ] Recover countdown after refresh
- [ ] expire abandoned sittings through durable scheduled work.
- [ ] Define authorized pause/break and window-close behavior.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### SIT-05

Durable saving · Milestone 3

- [ ] Autosave versioned responses, display saved/saving/offline/failed status, retry safely, and reconcile acknowledgments.
- [ ] Keep pending changes through refresh where permitted
- [ ] never label a lost update saved.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### SIT-06

Reconnection and concurrent clients · Milestone 3

- [ ] Resume after connection loss, reconcile local/server edits, reject stale writes, and detect a second active device/session.
- [ ] Adopt one active exam writer by default with an explicit supervised takeover flow.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### SIT-07

Timing evidence · Milestone 3

- [ ] Accumulate per-item active viewing/answer interaction with return visits and evidence quality.
- [ ] Exclude background/inactive time appropriately, retain uncertainty after outages, and treat client timing as an observation rather than trusted proof of ability or misconduct.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### SIT-08

Submission · Milestone 3

- [ ] Confirm voluntary submission with unanswered/optional-section summary
- [ ] safely finalize the latest acknowledged responses and explicitly handle pending unsent work.
- [ ] Submission, deadline expiry and conduct submission converge on one idempotent finalization outcome.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### SIT-09

Final state · Milestone 3

- [ ] Submitted content cannot change through answer-save routes.
- [ ] Offer a receipt, effective finish time, recorded reason, score/review status and reconnectable status lookup when the submit response is lost.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### SIT-10

Accessibility and continuity · Milestone 3

- [ ] Support keyboard/screen-reader operation, large text, narrow/landscape layouts and approved accommodations without trapping the learner in fullscreen or inaccessible controls.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### INT-01

Configurable conduct modes · Milestone 3

- [ ] Provide open practice, monitored browser, and institution-approved supervised delivery policies.
- [ ] Explain exactly what is observed and what consequence follows
- [ ] obtain required acknowledgments.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### INT-02

Browser observations · Milestone 3

- [ ] Observe fullscreen exit, focus/visibility changes, supported display information, copy/paste/context-menu restrictions and observable screenshot shortcuts when the selected policy enables them.
- [ ] Detect capabilities and document unsupported cases.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### INT-03

Events and deduplication · Milestone 3

- [ ] Retain event category, observed/server times, device context, actor and policy version
- [ ] deduplicate related browser events and network replays.
- [ ] Provide an accessible warning and remaining allowance.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### INT-04

Escalation · Milestone 3

- [ ] Apply configured warning, instructor review, temporary hold or finalization consequences on the server.
- [ ] Support staff adjudication, justified overrides and appeals, with retained evidence.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### INT-05

Answer secrecy · Milestone 3

- [ ] Prevent unreleased answers/solutions from appearing in assessment payloads, search, tutor history, content suggestions, attachments, notifications, caches and exports.
- [ ] Exposure history must affect future formal-paper selection or access policy.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### INT-06

Security versus monitoring · Milestone 3

- [ ] Never claim the browser can prevent OS recording, external devices or collaboration.
- [ ] Webcam/biometric collection is not required by default.
- [ ] Any later collection needs a separately approved policy and implementation.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### GRD-01

Objective correctness · Milestone 3

- [ ] Determine single selection, exact-set multiple selection and numerical correctness from the approved immutable key/rule version.
- [ ] Support defined all-or-nothing and explicitly authored partial-credit policies
- [ ] distinguish blank, invalid and answered wrong.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### GRD-02

Numerical acceptance · Milestone 3

- [ ] Support integer/decimal input, signs, bounds, accepted values/ranges, absolute/relative tolerances, units where authored, and explicit rounding/tie policy.
- [ ] Reject NaN/infinity, malformed input, and ambiguous locale formats with useful messages.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### GRD-03

Optional groups · Milestone 3

- [ ] Support count-all, first-N-attempted and best-N selection.
- [ ] Record exactly which answers contribute, why, and the resulting maximum.
- [ ] For offline first-N, use published printed-order selection because chronological answer order is unobservable
- [ ] do not claim to reconstruct it.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### GRD-04

Special cases · Milestone 3

- [ ] Support cancelled/withdrawn questions, corrected keys, bonus/full-credit policies, absent/excused sittings and section caps.
- [ ] Show the applied policy and its effect rather than silently altering totals.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### GRD-05

Subjective marking · Milestone 3

- [ ] Provide versioned rubrics, criterion/step marks, annotated script review, marker assignment, moderation, adjudication, double marking when required and reviewer workload.
- [ ] AI suggestions must preserve the script evidence and distinguish proposal from approved marks.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### GRD-06

Immediate outcomes · Milestone 3

- [ ] Return objective score/counts promptly after finalization even if commentary/media is unavailable.
- [ ] Separately expose pending subjective review and provisional/released results
- [ ] do not imply a provisional score is final.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### GRD-07

Release policy · Milestone 3

- [ ] Configure when scores, rank, item correctness, answer keys, full solutions and tutor access become available.
- [ ] Default formal exams to cohort closure
- [ ] practice can release immediately.
- [ ] Guardians receive only released child information.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### GRD-08

Learner review · Milestone 3

- [ ] Once released, show every permitted question, answer, correctness, marks, worked solution, reason/uncertainty and observed time quality
- [ ] distinguish optional blank, mandatory blank and invalid mark.
- [ ] Include subject/section breakdowns and offline/online provenance.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### GRD-09

Appeals · Milestone 3

- [ ] Allow a learner/instructor to flag a problem and submit evidence, route it to an authorized reviewer, track acknowledgment/resolution and notify affected people.
- [ ] Preserve the original submitted answer.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### GRD-10

Regrading · Milestone 3

- [ ] Preview affected sittings and score changes, require an authorized correction reason, publish a new marking revision, and recompute results, aggregates, learning indicators, rankings and applicable practice plans consistently.
- [ ] Keep old/new results and notifications auditable.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### GRD-11

Rankings · Milestone 3

- [ ] Calculate cohort ranks/percentiles only for comparable released papers/sitting policies.
- [ ] Define ties, eligible population and retake selection
- [ ] do not compare raw scores from different maximums or different personalized papers as if equivalent.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### PPR-01

Layout studio · Milestone 4

- [ ] Create/version response-sheet layouts with learner/paper identity areas, response kinds, section/question mapping, registration/orientation marks and machine-readable template identity.
- [ ] Preview printable output and validate calibration against the generated blank sheet.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### PPR-02

Format flexibility · Milestone 4

- [ ] Support different counts, sections, option labels, multi-select groups and numeric write-ins, including signed/decimal or four-plus-digit values when the selected format permits.
- [ ] Never apply one template's assumptions to another sheet silently.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### PPR-03

Intake session · Milestone 4

- [ ] Create a campus/paper-bound batch with the actual exam date/time, template version, upload permissions and reviewers.
- [ ] Accept images and multi-page PDFs through resumable asynchronous intake.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### PPR-04

Recognition and validation · Milestone 4

- [ ] Normalize perspective/orientation/exposure where supported, match layout identity, read references and responses, and validate learner eligibility/paper association.
- [ ] Uncertain identity or mismatched paper must block commitment.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### PPR-05

Confidence and uncertainty · Milestone 4

- [ ] Identify blanks, multiple marks, ambiguous fills and uncertain handwriting.
- [ ] Keep transcription independent of the answer key
- [ ] retain confidence/evidence and disagreements rather than replacing them with guessed certainty.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### PPR-06

Review console · Milestone 4

- [ ] Prioritize uncertain sheets/items, filter by status/subject/reviewer, display original/aligned/cropped evidence and overlays, and support keyboard-first correction.
- [ ] Allow explicit blank, invalid/wrong, corrected value, identity correction, flag acknowledgment with reason, and reject/quarantine.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### PPR-07

Reviewer scope · Milestone 4

- [ ] Assign subject-limited reviews without exposing unrelated private material unnecessarily.
- [ ] A designated coordinator approves whole-sheet identity and final commit.
- [ ] Detect conflicting simultaneous reviewer edits.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### PPR-08

Evidence audit · Milestone 4

- [ ] Preserve machine readings, original scans, confidence/geometry, reviewer changes, reasons, actor/time and template revision.
- [ ] A re-decode can propose updated reads but cannot erase approved corrections
- [ ] compare/reconcile conflicts visibly.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### PPR-09

Duplicate safety · Milestone 4

- [ ] Deduplicate repeated files/pages, repeated uploads across requests, and repeated learner/paper commit races.
- [ ] Offer an explicit corrected/replacement scan workflow rather than generating a second academic result.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### PPR-10

Commit · Milestone 4

- [ ] Commit eligible reviewed sheets once
- [ ] allow selected clean sheets while other sheets remain in review, making remaining work explicit.
- [ ] Each committed sheet creates the ordinary marking/learning/reporting flow, with actual sitting time and no invented pace.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### PPR-11

Failure recovery · Milestone 4

- [ ] Isolate bad pages, restart failed recognition, handle quota/provider failure as manual-review work, retain partial progress, and report rejected/failed/committed counts.
- [ ] Batch completion must not mean every sheet succeeded.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### PPR-12

Private access and retention · Milestone 4

- [ ] Protect source scans/signatures and derived crops through entitlement checks, expiring access and audit.
- [ ] Apply configured retention/hold rules while retaining required marking evidence.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### WRT-01

Script intake · Milestone 4

- [ ] Upload scanned written scripts or digital response attachments, associate them with the right sitting, map pages to questions and detect missing/duplicate pages and unreadable material.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### WRT-02

Transcription review · Milestone 4

- [ ] Extract handwriting/text/diagrams where supported, retain page coordinates and uncertainty, and allow corrections without overwriting the original evidence.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### WRT-03

Rubric assistance · Milestone 4

- [ ] Propose criterion-level marks and explanations grounded in the approved rubric and visible response.
- [ ] Show quoted/cropped supporting evidence and uncertainty
- [ ] never manufacture student working.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### WRT-04

Human approval · Milestone 4

- [ ] Allow manual marking when transcription/AI fails, require review under the institution's policy, support moderation and disagreement resolution, and finalize through the ordinary marking/release process.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### WRT-05

Student feedback · Milestone 4

- [ ] Release approved annotations, criteria, improvement suggestions and appeals with private document access.
- [ ] Keep rubric/key secrecy until the result-release policy allows it.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### WRT-06

Quality evaluation · Milestone 4

- [ ] Measure agreement with qualified reviewers on representative original/licensed scripts, report per-criterion disagreement, and support switching to mandatory human review when performance is inadequate.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### EVD-01

Explainable error observations · Milestone 5

- [ ] Distinguish observed incorrect/blank/invalid answers from suggested causes such as concept gaps, rushed work, time shortage, or uncertainty.
- [ ] Show the evidence and confidence
- [ ] retain unknown-cause states.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### EVD-02

Consistent measures · Milestone 5

- [ ] Define accuracy as correct among genuinely answered/scored responses, attempt coverage among required opportunities, and score percentage from applicable maximum.
- [ ] Publish alternative weighting explicitly.
- [ ] Do not silently include optional blanks or unmeasured time in denominators.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### EVD-03

Topic evidence · Milestone 5

- [ ] Accumulate per-topic/outcome observations over time with evidence count, recency, question quality, mode and revision provenance.
- [ ] Insufficient evidence is a first-class state.
- [ ] Apply multi-topic weights without double counting.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### EVD-04

Mastery transitions · Milestone 5

- [ ] Use a documented, versioned, reproducible policy for emerging/developing/secure understanding and regression.
- [ ] Require adequate diverse evidence for promotion
- [ ] expose history and allow review.
- [ ] Choose thresholds/calibration independently and validate changes before applying them.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### EVD-05

Behavioral indicators · Milestone 5

- [ ] Describe observed pace, accuracy habits, unanswered patterns and strategy needs across recent comparable work.
- [ ] Give practical coaching recommendations rather than permanent personality labels
- [ ] retain indicator/history and the source window.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### EVD-06

Missing evidence · Milestone 5

- [ ] Paper sittings, interrupted sessions and uncertain timing do not produce pace values.
- [ ] They cannot prove a prior timing-related difficulty has vanished.
- [ ] Show coverage and excluded observations.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### EVD-07

Focus priorities · Milestone 5

- [ ] Rank manageable learning priorities using strength of evidence, prerequisite relevance, exam/course goals and subject balance
- [ ] explain why an item is recommended and permit instructor adjustment with reason.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### EVD-08

Readiness · Milestone 5

- [ ] Provide a visible preparation summary with component definitions, comparable assessment coverage, confidence, goals and history.
- [ ] Omit unsupported dimensions or show insufficient evidence rather than assigning an invented neutral measurement.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### EVD-09

Historical corrections · Milestone 5

- [ ] Rebuild evidence chronologically after regrading, deletion permitted by policy, late paper ingestion or revised scoring.
- [ ] A corrected old result must not simply be appended as a new learning event.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### EVD-10

Cohort comparisons · Milestone 5

- [ ] Compare pace/accuracy only against an appropriate eligible cohort with minimum sample size, anonymity and published population/date.
- [ ] Personal practice scores must not masquerade as formal peer rank.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### EVD-11

Educational policy controls · Milestone 5

- [ ] Provide an authorized settings workspace for evidence minimums, mastery criteria, historical weighting, focus limits, adaptive coverage/difficulty, retest intervals and numerical policy defaults.
- [ ] Validate ranges and relationships, preview effects on representative records, publish a version, reset/restore defaults and document whether existing results are recalculated.
- [ ] Every setting must affect the behavior its description promises.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### REV-01

Revision collection · Milestone 5

- [ ] Collect released wrong answers, required skips and learner bookmarks, grouped by course/topic/reason
- [ ] preserve repeated encounters and origins while avoiding duplicate cards for the same learning target.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### REV-02

States and scheduling · Milestone 5

- [ ] Track new, practicing, demonstrated and reopened revision items.
- [ ] Allow personal notes and instructor guidance
- [ ] schedule spaced review and require fresh evidence before calling a misunderstanding resolved.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### REV-03

Discovery · Milestone 5

- [ ] Search/filter by wording, topic, subject, assessment, date, response type and resolution state.
- [ ] Show figures/options/solutions, learner working, prior help and next practice action.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### REV-04

One tutor conversation · Milestone 5

- [ ] Provide a consistent persistent tutor experience from result review, revision, lessons and practice, with appropriate context and clearly separated threads/learner privacy.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### REV-05

Grounded help · Milestone 5

- [ ] Supply released question/figure/options, approved solution/rubric, actual learner response and relevant lesson sources.
- [ ] Explain concepts, show steps, offer hints and fresh practice, and cite institution-approved sources when retrieved.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### REV-06

Interaction quality · Milestone 5

- [ ] Support streaming with interruption/retry, starter prompts, follow-up, feedback/reporting, mathematical rendering, useful transcript search, accessible input and recovery of unsent messages.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### REV-07

Release enforcement · Milestone 5

- [ ] Re-check entitlement on every history read, message, retrieval and attachment.
- [ ] Previously released questions must be excluded from confidential scored reuse for that learner by default
- [ ] policy exceptions need an explicit assessment design decision.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### REV-08

Abuse/privacy controls · Milestone 5

- [ ] Enforce per-person/institution request and spend limits, bounded context, attachment protections, prompt-injection resistance and approved knowledge-source scope.
- [ ] Never execute student-provided code or tool instructions in a privileged environment.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### REV-09

Teacher escalation · Milestone 5

- [ ] Allow a learner to escalate unresolved doubt to an assigned instructor with selected context, status and notification
- [ ] private tutor history is shared only through the applicable consent/policy.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### PLN-01

Goals · Milestone 5

- [ ] Set target exam/course outcomes, target date, weekly availability and achievable goals
- [ ] show progress and allow instructor/learner revisions without fabricated rank promises.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### PLN-02

Calendar · Milestone 5

- [ ] Generate/edit daily, weekly and monthly study plans with lessons, timed drills, revision and assessments
- [ ] reschedule missed tasks, avoid overload and show dependencies/estimated effort.
- [ ] Export calendar events with timezone correctness.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### PLN-03

Instruction design · Milestone 5

- [ ] Offer reviewed topic lessons, concise notes, formulas, worked examples, misconception checks and unsolved exercises.
- [ ] Adapt explanation style/depth and accessibility to observed needs and learner preference.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### PLN-04

Personal context · Milestone 5

- [ ] Show “what to work on” grounded in the learner's released work without leaking personal examples into shared lessons.
- [ ] Include useful next steps and clear rationale.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### PLN-05

Consumption · Milestone 5

- [ ] Provide section navigation, reader progress, bookmarks, notes, readable maths/figures, downloads when licensed, audio/video/transcript views and remembered position.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### PLN-06

Completion and understanding · Milestone 5

- [ ] Distinguish opened, viewed, activity completed and understanding demonstrated.
- [ ] Retain section progress and allow unmark/revisit
- [ ] completion alone does not establish mastery.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### PLN-07

Exercise responses · Milestone 5

- [ ] Let learners answer embedded practice/check questions, see policy-appropriate hints/feedback, and retain attempts.
- [ ] Connect reviewed practice evidence to learning indicators using its declared lower/high-stakes weighting.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### PLN-08

Content reuse/versioning · Milestone 5

- [ ] Reuse approved material by topic/outcome/language/level/accessibility/version and compatible needs.
- [ ] Refresh after approved changes while retaining completed historical versions and personal progress mappings.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### PLN-09

Instructor assignments · Milestone 5

- [ ] Assign learning packs or individual resources to classes/learners, set due dates, see progress and understanding checks, and intervene without overwriting a learner's private notes.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### PLN-10

Adaptive papers · Milestone 5

- [ ] Generate personalized papers emphasizing current gaps and prerequisite needs while maintaining breadth, choosing suitable difficulty and avoiding recent/exposed questions.
- [ ] Persist the selection rationale, coverage and fallback decisions.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### PLN-11

Thin-bank behavior · Milestone 5

- [ ] Show unavailable coverage, offer reviewed alternatives or explicitly disclosed repeat practice, and request authoring work.
- [ ] Never silently weaken a formal paper's secrecy or pretend target coverage was met.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### PLN-12

Live adaptive practice · Milestone 5

- [ ] Implement a distinct practice mode that chooses subsequent questions from observed responses and uncertainty, with transparent stopping conditions.
- [ ] Do not mix its score/rank with fixed formal papers without validated normalization.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### PLN-13

Continuity · Milestone 5

- [ ] Replace only obsolete unstarted recommended work, preserve started assignments, re-evaluate priorities after new evidence, and rotate focus after sustained improvement or regression.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### PLN-14

Motivation · Milestone 5

- [ ] Offer optional streaks, achievements and progress celebrations, with rest/accommodation-aware rules.
- [ ] Peer leaderboards are opt-in and cohort-scoped
- [ ] avoid exposing children's identities publicly.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### STU-01

Resource library · Milestone 6

- [ ] Upload and organize documents, lesson notes, presentations, video/audio, worksheets and links by course/topic/outcome/language/license.
- [ ] Search, preview, version, approve, assign, hide/unpublish and archive with access controls.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### STU-02

Content workflows · Milestone 6

- [ ] Separate authoring, review, publication and learner visibility
- [ ] retain feedback/version comparison and dependencies.
- [ ] Broken links, superseded material and missing accessibility assets are visible maintenance work.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### STU-03

Generated material · Milestone 6

- [ ] Generate original drafts of notes, slide lessons, teaching scripts, examples and exercises from approved sources/outcomes.
- [ ] Require academic review where the content may teach or assess a learner
- [ ] keep source/version/model attribution.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### STU-04

Narration · Milestone 6

- [ ] Generate speech from approved scripts with voice/language/rate controls, pronunciation overrides and math-aware phrasing
- [ ] provide transcript and downloadable media where licensed.
- [ ] Audio processing must not delay usable text.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### STU-05

Presentation-to-video · Milestone 6

- [ ] Accept PPTX/PDF slides and a custom or generated per-slide script.
- [ ] Select slide range, voice, presenter asset/layout, captions, output quality and target language
- [ ] validate script/slide alignment.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### STU-06

Presenter generation · Milestone 6

- [ ] Create a talking-presenter video using consented/licensed presenter assets, with static-presenter or narration-only alternatives explicitly identified.
- [ ] Support preview and control of placement/size/background without impersonating an unconsenting person.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### STU-07

Captions and exports · Milestone 6

- [ ] Generate/edit synchronized captions, provide transcript/caption-file export, render a short preview, inspect duration/resolution, then publish/download the full result.
- [ ] Caption editing creates a traceable revision.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### STU-08

Concept animations · Milestone 6

- [ ] Create educational animations through controlled templates/parameters and reviewed scripts, including a representative original mathematical/scientific explanation.
- [ ] Generated animation instructions are untrusted
- [ ] render in an isolated bounded environment without privileged execution.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### STU-09

Job operations · Milestone 6

- [ ] Queue single/batch generation, show stages/history/thumbnails, estimate available cost/time, save presets, cancel, retry, recover interrupted work, download results and purge unneeded artifacts under retention policy.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### STU-10

Topic delivery controls · Milestone 6

- [ ] Independently control generation, publication and visibility of media/resources
- [ ] uploaded and generated material can coexist.
- [ ] Hiding output must not falsely display a generating state or delete its audit history.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### STU-11

External resources · Milestone 6

- [ ] Validate resource links, preview supported embeds safely, show source attribution, and provide accessible/open-in-new-tab alternatives.
- [ ] Link access still respects institutional publication scope.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### TCH-01

Teaching home · Milestone 5

- [ ] Show assigned classes/learners, upcoming teaching and assessments, recent released results, review queues, unresolved doubts and actionable attention items.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### TCH-02

Learner detail · Milestone 5

- [ ] Inspect permitted subject/outcome evidence, comparable trends, required skips/errors, revision history, study engagement, assessments and attendance.
- [ ] Cross-subject insight may be granted for assigned learners without granting unrelated authoring privileges.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### TCH-03

Class insight · Milestone 5

- [ ] Provide student-by-topic heatmaps, subject/date/class filters, evidence counts, common misconceptions, question quality concerns, top/improving/struggling groups and exports.
- [ ] An empty cell is unobserved, not weak.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### TCH-04

Interventions · Milestone 5

- [ ] Create a private academic note or structured intervention with reason, learner/group, assigned staff, tasks/resources, due date and follow-up outcome.
- [ ] Distinguish staff-only notes from family/learner-visible guidance.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### TCH-05

Recommendations · Milestone 5

- [ ] Propose remedial groups, teaching sequences, targeted resources and reassessment based on actual evidence.
- [ ] Allow teachers to accept/edit/reject and record why
- [ ] preserve human responsibility.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### TCH-06

Teaching outcomes · Milestone 5

- [ ] Show cohort progress and instructional engagement with comparable populations and evidence limitations.
- [ ] Do not present unadjusted student scores as proof that one teacher caused improvement.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### TCH-07

Professional directory · Milestone 5

- [ ] Maintain approved biography, qualifications/experience, teaching subjects, institutional contact and relevant availability.
- [ ] Default directory access to assigned families/learners and appropriate staff
- [ ] public publication requires explicit institution/teacher policy.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### TCH-08

Roster and review tasks · Milestone 5

- [ ] Search/filter assigned learners, retain bulk selections, mark/review delegated work and produce a useful student report.
- [ ] Remove access when the teaching assignment ends, subject to policy for retained grading obligations.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### FAM-01

Child workspace · Milestone 7

- [ ] Provide a clear child switcher and read-only released progress, topic strengths/gaps, learning recommendations, assessment history, rank context where permitted, upcoming deadlines and attendance.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### FAM-02

Useful reports · Milestone 7

- [ ] Show readable summaries and drill-down without overwhelming jargon, with explanation of unknown/insufficient evidence, actual dates, preparation goals and appropriate ways to support the child.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### FAM-03

Communication · Milestone 7

- [ ] Contact assigned staff through permitted threads, view announcements, acknowledge requested notices and arrange meetings where configured.
- [ ] Do not expose staff private phone/email details by default.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### FAM-04

Consent and preferences · Milestone 7

- [ ] Manage notifications, approved child data-sharing/optional-feature choices, relationship/contact verification and receipt of relevant policy changes.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### FAM-05

Engagement · Milestone 7

- [ ] Record actual portal/notice engagement for authorized institution oversight, distinguishing login from meaningful activity.
- [ ] Show transparent definitions and avoid making judgments about parenting from visit frequency.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### FAM-06

Print/download · Milestone 7

- [ ] Print or download a child's released report with clean pagination, readable charts, date range and reference/version.
- [ ] Missing or revoked family relationships must immediately remove private access.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### ATT-01

Attendance capture · Milestone 7

- [ ] Record class attendance, late/excused/absent states, dates, sessions and authorized corrections.
- [ ] Support manual and validated bulk import
- [ ] retain who recorded or changed each entry.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### ATT-02

Timetables · Milestone 7

- [ ] Maintain recurring sessions, rooms/online links, holidays, substitutions and schedule exceptions
- [ ] show appropriate learner/teacher/family calendars and handle timezone/display correctly.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### ATT-03

Follow-up · Milestone 7

- [ ] Notify configured recipients of missing attendance or recurring absence, with correction/acknowledgment workflow and safeguards against duplicate notifications.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### ATT-04

Analysis · Milestone 7

- [ ] Relate attendance and learning participation to observed academic outcomes with date/subject filters and explicit population
- [ ] describe correlation without claiming causation.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### ATT-05

Activity definitions · Milestone 7

- [ ] Define enrollment, active learner, lesson activity, assessment completion and guardian engagement consistently
- [ ] show metric windows and avoid inflated activity from polling or background processes.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### COM-01

Notification inbox · Milestone 7

- [ ] Provide categorized unread/read/actioned notices for assessments, reports, lessons, messages, account security, billing and review exceptions
- [ ] include deep links authorized when opened.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### COM-02

Delivery channels · Milestone 7

- [ ] Implement in-app, email and mobile/web push, plus configured SMS for selected events.
- [ ] Track queued/sent/provider-accepted/delivered/failed where actually observable
- [ ] keep security messages reliable.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### COM-03

Preferences · Milestone 7

- [ ] Offer per-event/channel preferences, digest frequency, timezone/quiet hours, unsubscribe where applicable, and institution-mandatory notices with explicit justification.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### COM-04

Messaging · Milestone 7

- [ ] Provide scoped instructor–learner and guardian–staff threads, attachments, read-state policy, search, conversation assignment and escalation.
- [ ] Avoid unrestricted discovery/contact between minors or unrelated adults.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### COM-05

Class discussions · Milestone 7

- [ ] Support topic/class forums, questions/replies, helpful answers, attachments, reporting, moderation, locking and archival.
- [ ] Make participation policies and staff responsibility visible.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### COM-06

Announcements · Milestone 7

- [ ] Draft/review/publish scheduled organization/campus/class announcements, select recipients, request acknowledgment and report delivery/engagement without exposing other recipients' private data.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### COM-07

Templates and retries · Milestone 7

- [ ] Version message templates, localize, preview sample data, enforce deduplication, retry transient delivery failures and provide dead-letter recovery.
- [ ] Do not leak answer keys through previews or notification snippets.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### COM-08

Meetings · Milestone 7

- [ ] Offer appointment requests/time slots, confirmation/cancellation, reminders and calendar files
- [ ] integrate an external meeting/calendar provider only through a configured adapter.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### COM-09

Support desk · Milestone 7

- [ ] Submit a help request, attach consented evidence, track status/owner/replies and link an approved expiring support grant.
- [ ] Provide searchable user help and contextual guidance.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### RPT-01

Personal insight · Milestone 7

- [ ] Show completed/assigned assessments, average/best normalized scores, accuracy, required coverage, trends, time quality, readiness, focus priorities, study progress and revision outcomes.
- [ ] Every tile drills into its underlying evidence.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### RPT-02

Institution oversight · Milestone 7

- [ ] Compare authorized campuses/classes/courses/subjects with enrollment, participation, completion, learning evidence, staff responsibility and review backlog.
- [ ] Use consistent definitions and comparable populations.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### RPT-03

Filters and comparisons · Milestone 7

- [ ] Support date range, course, campus, subject, class, instructor, assessment type/channel and evidence sufficiency
- [ ] retain filter state in permitted shareable links.
- [ ] Compare periods with stated denominators.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### RPT-04

Report builder · Milestone 7

- [ ] Select student/class/institution report scope, date range, metrics, sections, grouping/sorting, charts and privacy mode
- [ ] preview, save templates and generate reports with provenance/version.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### RPT-05

Export center · Milestone 7

- [ ] Produce real PDF, CSV and XLSX outputs
- [ ] export allowed charts as PNG/SVG and heatmaps as appropriate.
- [ ] Large exports run in the background with access-controlled, expiring downloads and spreadsheet-formula injection protection.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### RPT-06

Scheduled reporting · Milestone 7

- [ ] Schedule report generation/delivery to authorized recipients, honor timezone, permission changes and retention, expose run history, retry safely and stop when recipients lose entitlement.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### RPT-07

Custom dashboards · Milestone 7

- [ ] Provide a constrained interactive chart builder with permitted dimensions/metrics/types/layout, accessible color choices, saved dashboards and drill-down.
- [ ] Do not expose arbitrary SQL or unrestricted query execution.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### RPT-08

Leaderboards and benchmarking · Milestone 7

- [ ] Offer comparable assessment/cohort rankings, institution-approved anonymous peer percentiles and optional learner participation.
- [ ] Show cohort size, ties, eligibility and retake policy
- [ ] no universal rank fabricated from adaptive practice.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### RPT-09

Evidence traceability · Milestone 7

- [ ] Include calculation policy, source revision and as-of time
- [ ] show incomplete/stale recalculation states.
- [ ] PDF/Excel/UI totals must agree for the same filters and revision.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### RPT-10

Readable sharing · Milestone 7

- [ ] Create clean print layouts and explicitly authorized expiring report shares with audit and revocation.
- [ ] Never make a child's report publicly discoverable for convenience.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### INS-01

Performance trajectories · Milestone 8

- [ ] Implement explainable trend/improvement estimates from comparable historical work, with data sufficiency, uncertainty intervals, influencing factors and target gap.
- [ ] Do not equate extrapolation with a guaranteed exam result.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### INS-02

Rank forecasting · Milestone 8

- [ ] Implement a forecast pipeline with an appropriate licensed outcome/calibration dataset, evaluation split, versioned model and out-of-sample calibration.
- [ ] If the necessary exam/cohort data are absent, return a structured insufficient-evidence result and explain what is needed.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### INS-03

What-if tools · Milestone 8

- [ ] Offer hypothetical study/score scenarios clearly labeled as simulations, constrain inputs and show assumptions/uncertainty.
- [ ] Hypothetical scenarios must never update real learner evidence.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### INS-04

Confidence and reflection · Milestone 8

- [ ] Provide voluntary pre/post assessment confidence, perceived time pressure, and study reflections.
- [ ] Separate self-report from timing inference, permit correction/withdrawal, and respect guardian/institution consent policies.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### INS-05

Supportive coaching · Milestone 8

- [ ] Offer nonclinical reflection prompts, preparation/break strategies and instructor follow-up based on consented observations.
- [ ] Never diagnose anxiety, mental illness or personality from a score, camera, or click pattern
- [ ] visibly distinguish learner statements from inferred suggestions.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### INS-06

Human attention flags · Milestone 8

- [ ] Show explainable academic/support flags to authorized staff, with evidence, review, dismissal, follow-up and correction.
- [ ] No automated punitive or clinical decision based on a predictive score.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### INS-07

Evaluation workspace · Milestone 8

- [ ] Manage original/licensed evaluation datasets and labels, model/prompt versions, test runs, drift, per-subgroup error and calibration, comparison/rollback and approval.
- [ ] Keep evaluation data scoped/private.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### INS-08

Production gates · Milestone 8

- [ ] Expose evaluated versus experimental insight and fail closed to honest missing evidence when input distribution or model support is inadequate.
- [ ] Deliver the functional model training/evaluation/inference path even when a deployment still needs its real institutional dataset.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### FIN-01

Plans and entitlements · Milestone 8

- [ ] Manage subscription plans/trials, seat/campus/storage/AI allowances, feature access, upgrades/downgrades, grace periods, renewal/cancellation and institution billing identity.
- [ ] Preserve learners' access to entitled historical records when a plan changes.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### FIN-02

Checkout/payment · Milestone 8

- [ ] Implement a real payment-provider adapter with sandbox support, hosted/tokenized payment collection, state tracking, signed callbacks and idempotent settlement.
- [ ] The platform must not store raw payment-card details.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### FIN-03

Financial records · Milestone 8

- [ ] Represent money in currency-appropriate minor units with explicit currency, rounding and discount/tax configuration.
- [ ] Issue immutable numbered invoices/receipts and correction/credit records instead of editing settled totals.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### FIN-04

History and billing UI · Milestone 8

- [ ] Show invoices, credits, payments, failed collection, due dates, subscription state and downloadable documents to authorized finance staff and payers.
- [ ] Explain pending versus completed payment.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### FIN-05

Refunds/cancellation · Milestone 8

- [ ] Provide permissioned refund requests, reason/approval, full/partial refund execution, provider confirmation and failure/reconciliation.
- [ ] Prevent concurrent or replayed requests from refunding more than the settled refundable amount.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### FIN-06

Webhook/reconciliation · Milestone 8

- [ ] Verify authenticity, persist event IDs, tolerate replay/out-of-order delivery, reconcile provider status and expose discrepancies.
- [ ] A checkout-success page is not proof of settlement.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### FIN-07

Budgets and usage · Milestone 8

- [ ] Meter billable seats/services and AI consumption transparently, show period limits and actual-versus-estimated amounts, alert budget owners and define behavior at the limit without losing exam answers.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### FIN-08

Financial insight · Milestone 8

- [ ] Report collected/outstanding/refunded revenue, subscription changes and usage trends, scoped by organization/campus where meaningful.
- [ ] Separate accrual/invoice views from cash receipts and document definitions.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### FIN-09

Provider setup and access · Milestone 8

- [ ] Validate configuration, provide test-mode indicators and controlled production activation, mask secrets, separate finance permissions, and audit refunds/plan changes.
- [ ] Never charge or issue refunds during automated demo seeding.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### UX-01

Public site · Milestone 1

- [ ] Provide original home, feature/audience explanations, company/about, pricing where configured, contact/demo request, help, terms/privacy and sign-in/enrollment entries.
- [ ] A demo booking/contact form must actually persist and route requests.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### UX-02

Honest content · Milestone 1

- [ ] Publish only supported features and verified outcomes
- [ ] show real availability/configuration rather than invented testimonials, customer numbers, prediction accuracy or locked-browser claims.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### UX-03

Original visual system · Milestone 1

- [ ] Define coherent typography, spacing, color, density, components and motion
- [ ] support light/dark/system preference and text-size settings.
- [ ] Keep complex staff tools legible and learner work calm and focused.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### UX-04

Responsive behavior · Milestone 1

- [ ] Support phones, tablets, desktop and print
- [ ] dense tables need usable filtering/column choices or alternate views.
- [ ] Maintain accessible touch targets and avoid horizontal overflow for core workflows.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### UX-05

Accessibility · Milestone 1

- [ ] Target WCAG 2.2 AA through keyboard navigation, visible focus, semantic labels, error summaries, screen-reader context, contrast, reduced motion, captions/transcripts and accessible mathematical content.
- [ ] Test complete workflows, not only component snapshots.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### UX-06

Localization · Milestone 1

- [ ] Provide an internationalization foundation and complete English/Hindi UI translations for delivered workflows, locale-aware dates/numbers and content-language selection.
- [ ] Keep technical identifiers stable and permit additional languages without rewriting business logic.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### UX-07

State quality · Milestone 1

- [ ] Every important view needs truthful loading, empty, partial, offline, permission-denied, recoverable error and success states, including saved feedback and destructive-action confirmation.
- [ ] Preserve user input across retryable failures.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### UX-08

Discoverability · Milestone 1

- [ ] Implement entitled global search, navigation breadcrumbs, remembered workspace context, accessible dialogs, contextual help and clear account/child switches.
- [ ] Do not expose records through search snippets beyond the user's scope.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### UX-09

Public performance/metadata · Milestone 1

- [ ] Provide appropriate discoverability metadata, sitemap/canonical handling and fast public pages.
- [ ] Exclude private learner pages and reports from indexing/cache leakage.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### MOB-01

Native applications · Milestone 9

- [ ] Deliver Android/iOS source/build configuration with sign-in/MFA/recovery, learner/family home, lessons, tutor, revision, eligible assessments, notifications, goals and released reports.
- [ ] Reuse business contracts and design tokens, not browser-only components by force.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### MOB-02

Secure local state · Milestone 9

- [ ] Use OS-protected credential storage, per-user/institution caches, explicit downloads and eviction, local data encryption appropriate to sensitivity, and sign-out/revocation cleanup.
- [ ] Do not leave another child's records accessible after switching accounts.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### MOB-03

Offline study · Milestone 9

- [ ] Download permitted notes/media/exercises and continue reading/progress/notes offline
- [ ] show download rights, size, availability, expiry and synchronization state.
- [ ] Protected formal answer keys are never cached early.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### MOB-04

Synchronization · Milestone 9

- [ ] Use retryable versioned mutations, stable operation IDs, conflict review/resolution and server acknowledgments.
- [ ] Preserve edits, handle partial upload and stop sync when membership/permissions change.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### MOB-05

Assessment network policy · Milestone 9

- [ ] Formal examinations require server-backed eligibility/deadlines and a documented connectivity policy.
- [ ] Offline practice is supported
- [ ] offline formal sittings are allowed only through an explicitly configured supervised delivery design, never by silently trusting device time.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### MOB-06

Push/deep links · Milestone 9

- [ ] Register/revoke push tokens by installation, navigate to authorized content, respect child/workspace context and preferences, and handle stale or revoked links safely.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### MOB-07

Web installation · Milestone 9

- [ ] Provide installable web access and offline study where browser capabilities permit, with truthful limits compared with native apps.
- [ ] Service workers must not cache unreleased assessments/results or private responses across users.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### MOB-08

Qualification · Milestone 9

- [ ] Verify interruption/resume, updates, permission revocation, accessibility, low bandwidth, storage exhaustion and basic Android/iOS flows
- [ ] document store-signing/release configuration separately from source completion.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### AIO-01

Capability adapters · Milestone 5

- [ ] Route text, vision/document, retrieval, speech and presenter generation through Go-controlled adapters with explicit model/provider capabilities, timeouts, cancellation and supported input limits.
- [ ] At least one functional adapter is required for each delivered AI capability.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### AIO-02

Instruction management · Milestone 5

- [ ] Edit/review/version generation instructions, describe variables/output contracts, preview/test against fixtures, compare history, publish/reset/rollback and retain editor/approver identity.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### AIO-03

Output validation · Milestone 5

- [ ] Validate structured outputs, academic references, allowed values, completeness and asset relationships before business use.
- [ ] Preserve malformed/failed outcomes safely for authorized debugging without exposing private data unnecessarily.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### AIO-04

Reuse/invalidation · Milestone 5

- [ ] Key reusable generation by approved input, instruction, model, language and policy revisions
- [ ] prevent duplicate concurrent generation
- [ ] refresh dependent material after approval while preserving completed historical content.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### AIO-05

Complete consumption accounting · Milestone 5

- [ ] Record requests, model/provider, feature, tenant/person scope, input/output/reasoning/cached tokens where observable, retries/failures, latency, monetary estimates and provider-reported actual cost.
- [ ] Distinguish missing cost data from zero.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### AIO-06

Budgets/limits · Milestone 5

- [ ] Enforce per-feature/person/organization ceilings and shared provider quotas, reserve/reconcile spend for concurrent jobs, and provide warnings and admin overrides with audit.
- [ ] Unauthorized requests must not burn a learner's allowance.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### AIO-07

Resilience · Milestone 5

- [ ] Use capability-compatible fallback and bounded retry/backoff
- [ ] distinguish transient overload, authentication failure, exhausted budget and daily quota.
- [ ] Do not repeatedly retry a permanent error or discard approved human review.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### AIO-08

Quality and feedback · Milestone 5

- [ ] Collect content/tutor reports, run repeatable evaluations, compare provider/instruction revisions and gate publication on actual evidence.
- [ ] Track hallucination/diagram extraction/marking failures as supportable incidents.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### AIO-09

Data boundaries · Milestone 5

- [ ] Isolate retrieval/content caches by tenancy and permitted sharing
- [ ] minimize personal data, redact logs and control provider retention settings.
- [ ] Treat document/student content as data, not instructions to access tools or secrets.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### AIO-10

Operator console · Milestone 5

- [ ] Show current configuration/capabilities, consumption, budgets, failures, evaluation results, generation history and recoverable jobs.
- [ ] Secrets are set through controlled configuration and never revealed through the ordinary console/API.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### RUN-01

Durable execution · Milestone 1

- [ ] Persist job intent with the triggering business transaction
- [ ] support leases/heartbeat, retry attempts, exponential backoff where appropriate, cancellation, dead-letter inspection and crash recovery.
- [ ] Exactly-once business outcomes must survive at-least-once delivery.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### RUN-02

Operational console · Milestone 1

- [ ] Search/filter jobs, inspect safe error context, replay failed work with authorization, cancel or reschedule, see backlog/progress, and identify orphaned records/assets.
- [ ] Every administrative recovery action is audited.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### RUN-03

File lifecycle · Milestone 1

- [ ] Store originals and derivatives privately, validate file signatures and limits, quarantine unsafe/unscannable uploads, use bounded document/media processing, preserve provenance, handle resumable uploads and remove abandoned artifacts safely.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### RUN-04

Audit · Milestone 1

- [ ] Provide append-only security/business audit records with actor, tenant, scope, operation, reason, object/revision, correlation ID and time.
- [ ] Search/export through appropriate permissions
- [ ] logs must not contain passwords, session tokens, raw card data or unrestricted tutor transcripts.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### RUN-05

Application security · Milestone 1

- [ ] Protect CSRF/session cookies, request boundaries, injection/SSRF, upload execution, object access, cross-tenant reads/writes, export abuse, output rendering and webhook integrity.
- [ ] Encrypt transport and restrict service credentials
- [ ] use centralized secret management appropriate to deployment.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### RUN-06

Privacy/consent · Milestone 1

- [ ] Provide data-access/export/correction/deletion-request workflows, guardian/optional-feature consent records, revocable shares, retention schedules and legal holds.
- [ ] Implement the institution's configured jurisdiction/policy requirements without claiming automatic compliance from a checkbox.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### RUN-07

Resilience/health · Milestone 1

- [ ] Offer liveness/readiness, structured logs, request/job traces and metrics for API latency/errors, database saturation, job age, marking delay, file processing, provider quota and financial reconciliation.
- [ ] Expose safe operator dashboards and actionable alerts.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### RUN-08

Backup and recovery · Milestone 1

- [ ] Schedule encrypted database/object backups, define RPO/RTO and retention, document restore, and perform a timed restore drill including academic versions, files, financial records and access controls.
- [ ] A backup file alone is not recovery evidence.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### RUN-09

Environment/deployment · Milestone 1

- [ ] Provide local containers, migrations, validated secrets/provider settings, CI checks, staged releases, rollback procedures, maintenance mode and safe schema evolution.
- [ ] Reboots and deployments must not lose private files or acknowledged responses.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### RUN-10

Performance/concurrency · Milestone 1

- [ ] Adopt the reference workload in section F, load-test actual behavior, bound expensive work and stream/queue large operations.
- [ ] Do not hardcode success claims without an observed run.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### RUN-11

Support recovery · Milestone 1

- [ ] Resolve account/session issues, failed import/scan jobs, missing files, disputed marks, notification failures and payment discrepancies through authorized tools/runbooks.
- [ ] Operator access to private records requires a scoped support grant and audit.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### RUN-12

Data portability · Milestone 1

- [ ] Export/import documented institution-owned data formats, validate mappings and reconcile counts/revisions in a fresh installation.
- [ ] Do not require any external historical application or its data model.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.

### RUN-13

Delivery evidence · Milestone 1

- [ ] Maintain requirement status, test results, screen/state coverage, accessibility observations, provider setup and known limitations.
- [ ] Distinguish source completion, local/sandbox verification and production verification.
- [ ] Authorized API, schema/data effects, functional UI, recoverable failures and meaningful acceptance evidence.



## Teaching desk refinement — 2026-11-19

Added direct manual authoring alongside upload/AI, inclusive range/list marking controls using versioned Go APIs, multi-file OMR intake with group recovery, and searchable student marks/correct-question tables. Empty OCR extraction now exposes recovery instead of a ready success state. Existing permission checks, human review gates and pinned assessment revisions remain authoritative. Evidence: `tests/browser/teaching-desk.spec.ts`, updated import/academic/usability journeys, `tests/question-selection.test.ts`, and `docs/TEST-REPORT.md`.

These are additional implementation evidence for CAT-02/04/08, DOC-02/05/06, ASM-03, PPR-01/04/10/11 and grading/result presentation; they do not change those requirements to verified. OMR requires registered layouts and encoded assignment identity. Arbitrary handwritten-name recognition, universal OMR layouts, production provider/model quality, full permissions/recovery qualification and full acceptance layers remain outstanding.


Reference-dashboard follow-up: replaced the welcome hero with live two-by-two counts, compact question-entry/saved-assessment lists, real search and job states, and working OMR/results/account links. Authenticated browser coverage is in `tests/browser/reference-dashboard.spec.ts`. No API, business rules, schema or requirement-verification status changed.


Light-only owner refinement: all dark/system theme branches and the Appearance selector are removed; saved preferences migrate to light. Manrope, frosted translucent panels, requested pastel tints, an original local illustration, and live top-bar job/AI shortcuts use the shared existing routes. Dashboard coverage now asserts dark-OS/old-dark-setting migration and decoded illustration pixels. No academic API, grading rules or requirement verification status changes.

## Assessment-first drafting increment — 2026-11-19

Added migration 00009 and scoped/versioned `assessment_draft` persistence, generated contracts, explicit review/publish/assign transaction, separate editable revisions of published papers, five primary faculty routes, three-method setup, course-backed batch/roster/date metadata, multi-group generation, common question/key/marking editor and source-key identity matching. The existing course represents the batch; a full separate cohort/group model remains outstanding. Student admission permissions are unchanged.

Question generation and OCR normalization now select the loopback-only Ollama adapter (default Qwen3.5-9B Q4_K_M); no hosted generation fallback. OCR text and normalized chunks survive failures/retries. Primary-source research and hardware assessment are in docs/LOCAL-ASSESSMENT-MODEL.md. Subsequent live testing installed/connected the runtime and weights, exercised real JEE/NEET generation, corrected one worked solution through faculty review and published to a synthetic learner. Digital-PDF text, typed extraction output, literal printed-key reconciliation and an automated generation consistency pass address observed failures. These functional runs do not establish general scientific correctness, exam difficulty or production performance. New backend/browser evidence is recorded in docs/TEST-REPORT.md; no broad requirement is promoted to verified.

Live OMR recognition matches all 190 independently checked positions in the two supplied images, including six flagged blanks. Their real student identities and grades are not inferred. Separate registered known-key sample and batch workflows verify authoritative grading and uncertain-response review. Universal layouts, arbitrary handwritten student names, disjoint numerical acceptance ranges and source partial-credit schemes beyond the existing grading policies remain outstanding.

The three supplied full PDFs now pass functional upload checks for 48/51/51 questions, source order/types/keys/marks, retained missing-key review states and reload persistence. A missing-question failure exercised source-completeness rejection and selective cached-chunk regeneration before its passing retry. A known formula defect was corrected through the actual draft editor; complex mathematical and diagram fidelity remains unqualified, so these counts do not imply unattended import correctness or full requirement acceptance.

## Native OMRChecker integration — 2026-11-20

Added a scoped native-template reading endpoint, generated contracts, isolated headless CPU adapter and an upload/result panel under Upload OMR answer sheets. It calls the installed OMRChecker checkout directly for configured preprocessing, page/marker alignment and adaptive bubble reading, returning detected fields, combined template values, an annotated image and downloadable JSON. Recognition does not invoke upstream grading or create academic results. An original filled synthetic sheet and its template/expected responses are provided for repeatable testing.

The initial engine-first delivery deferred visual configuration. The following increment adds it. Native PDF/batch intake through this new endpoint, reliable handwriting recognition, arbitrary-template discovery, full field-scan accuracy and full product acceptance remain outstanding. See `omr/README.md` and `docs/TEST-REPORT.md`; no requirement is promoted to fully verified.

## Guided native OMR layout setup — 2026-11-20

The default OMRChecker flow now builds its template, reference and tuning from visual name/roll snippets and two clicked rows per answer column. Remaining bubble rows interpolate automatically, with editable counts, numbering, bubble size and blocks. Native reference alignment replaces the need for a separate marker upload. Layouts can be downloaded as one reusable file, imported after reload and reused on subsequent sheets. Existing native-file upload remains available; legacy pixel diagnostics are collapsed.

Go validates optional identity rectangles and the native adapter performs local Tesseract OCR on the aligned snippets. Identity text is editable and uncertain text stays flagged; no automatic roster match or academic grade is created. Portable layouts do not introduce database storage or require a migration. Real-engine, saved-layout and keyboard/mobile evidence is recorded in `docs/TEST-REPORT.md`. General handwritten identity accuracy, field-scan diversity and complete product acceptance remain unqualified.

The subsequent whole-sheet refinement replaces the step-based editor with a draw-then-label workspace. Registration/name/answer frames can be added in any order, with no premature circles. Whole-sheet fit, an expanded native dialog, explicit bubble-spacing setup with automatic selected-column zoom, lettered/numbered option labels and optional saved bubble guides simplify the CTET-style layout. Pointer drawing now updates a transient SVG outline on animation frames without rebuilding the React field tree. The duplicate legacy pixel editor is no longer mounted; coordinate diagnostics remain available separately. Native recognition on the supplied CTET image matched all 150 independently checked responses through a visually generated five-block layout. Existing profile compatibility and grading authority are unchanged; this is functional evidence rather than full field/accessibility acceptance.

## Saved-template OMR checking by registration — 2026-11-20

The faculty scanner now has one ordered flow: published assessment → saved named template → one/multiple PNG/JPEG or PDF uploads. Calibration and template creation sit below the scanner/results. Raw native-template upload, the pixel-reader diagnostic and the registered-sheet scanner are no longer alternative faculty UI options. Migration 00010 adds course-scoped named templates, durable source uploads and account-free per-page grades; resources use private object storage. Question/option counts and supported selection-grid types are checked against the selected paper before intake.

The native worker reads independently of the key and passes mapped responses to the existing exact Go grader with the published paper's marks, deductions and section policies. Registration numbers retain leading zeros; name is optional. No student account, assignment or linked result revision is manufactured. Draft publication now permits zero online recipients. Uncertain identity/answers stay provisional until versioned review with confirmation/reason. Original readings, source images, marked evidence and previous review answers/grades are retained. Duplicate files reuse checks; duplicate graded registrations cannot overwrite or merge grades, and a page can be explicitly excluded. Per-file errors, partial PDF progress, cancellation and runtime recovery remain visible. A retry race with a still-running River attempt now returns a recoverable conflict instead of falsely leaving academic work queued.

Results are also visible under Results, ahead of account-linked submissions. Cursor pagination retains access to older uploads. Real-engine/private-storage/PostgreSQL, browser and geometry evidence is recorded in `docs/TEST-REPORT.md`; this extends PPR-01/03/04/05/06/08/09/11/12 implementation evidence without marking full requirements verified. Numerical digit grids, automatic geometry discovery, account reconciliation, full reviewer/retention policy acceptance, field handwriting/scan diversity and production/container qualification remain outstanding.


## Reusable OMR template library — 2026-11-20

Templates now describe physical sheet layouts independently of assessments and are listed in an institution-wide staff library. Faculty can calibrate/save without selecting an assessment, then reuse the same saved template across assessments and teaching batches. Existing templates remain available; their course/campus records are private-asset provenance only. Checks still require the selected assessment's full staff/subject access and use its pinned questions, answer key, marks and policies. Library metadata does not grant access to another course's private reference image. No schema migration is needed.

A template may have more answer rows than the assessment: only the first N rows are graded, and unused blanks/multiple marks do not affect the result. Too few rows, unsupported question types and incompatible option counts on used rows are rejected. The scanner retains assessment → saved template → individual/batch upload, with calibration below it. Real-engine integration and connected browser evidence in `docs/TEST-REPORT.md` cover creation without an assessment, reuse across two assessments, cross-batch instructor access, independent keys, PDF/batch intake and private-image denial. No broad requirement is promoted to verified.
