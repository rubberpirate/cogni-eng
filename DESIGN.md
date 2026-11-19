# Learning Workspace visual system

Institutional application and public introduction. Audience: staff organizing real work, learners concentrating, families understanding evidence. Primary first action: choose an institution and work context. Character: calm, precise, legible, accountable. Essence: records with clarity.

Use a light teaching studio informed by the owner-supplied references: a white labelled navigation rail, pale blue work surface, rounded panels and restrained peach/yellow/lilac assessment covers. The signature is original response-sheet line artwork paired with direct teaching actions. Calm, approachable, precise, legible; essence: room for discovery. Public content uses the same typography without displaying invented product metrics.

## Current tokens — light glass (owner specification)

The 2026-11-19 light-only specification supersedes the earlier opaque/dark visual system and historical notes below. Locally bundled Manrope, weights 400/500/600; base 16px, stat numerals 28–32px, labels 12–13px, headings 16–21px. Spacing unit 4px; gutters 14–24px. Panels 24–26px, inner rows 16px, pills/circular avatar controls. A washed-out sage/lavender/blush gradient sits below a 68% milky overlay. Glass surfaces use white at 48–55%, 18–22px backdrop blur, translucent white borders and `0 8px 30px rgba(31,42,55,.055)` shadow.

| Role | Value |
|---|---|
| Background base | #eef2f3 |
| Surface fallback | #f8faf9 |
| Glass | rgba(255,255,255,.52) |
| Text | #1f2a37 |
| Secondary text | #586878 |
| Accent | #3f6f83 |
| Accent text | #f8fbfc |
| Sage | #bfd9d6 at 72–76% |
| Lavender | #d3d0ee at 72–76% |
| Blush | #f3c6da at 72% |
| Sky | #bcd8ee at 65–76% |
| Butter | #f6ebb4 at 64–76% |
| Positive badge | #bfe8a8 / #3d6330 |

Semantic HTML, actual labels, visible focus, 44px interaction targets, text with states, inline errors and a status region. Tables use left-aligned labels and right-aligned tabular numbers; narrow screens stack work forms and scroll tables in labeled regions. Language and text-size preferences remain available; dark/system theme selection is removed. Controls respond over 180ms with a gentle lift/shadow; reduced motion disables movement. No stock imagery needed for record workflows.

Audit: rendered desktop, narrow-screen and dark-theme review completed for the current teaching overview/sign-in. WCAG conformity is not yet established.

## Teaching navigation refinement — 2026-11-18

Staff enter Home and see the sequence from questions to assessment to evidence. Direct task links expose AI generation, bank, assessments, scanning, results and sample files; organization and administration use separate disclosures. Learners enter My assessments. Mobile Menu retains keyboard navigation and avoids horizontal overflow. Hash state supports browser back and selected-paper refresh.

Keep each flow's next action close to its input. Show real loading, empty and provider-error states; poll active work without discarding editors. Numerical rules and rubric criteria use normal labelled fields. Uncertain single-choice readings select no radio until an explicit correction. Advanced scoring/layout options are disclosed on demand. Rendered desktop/mobile journeys were inspected and tested; full WCAG qualification remains open.

## Owner reference refinement — 2026-11-18

Transpose the supplied dashboard references through original content and geometry; no borrowed names, logos, illustrations, layouts or fabricated metrics. Show real paper/topic counts and actual job states. Assessment cards expose the saved title, publication state, pinned question count and maximum score, and open the actual assessment. A four-action tool row follows the educational sequence. Responsive forms and scan/review inherit the shared softer panel/control system. Sign-in uses a split introduction/form at desktop and a compact header on phones; all authentication/recovery controls retain their behavior.

Pastel roles: soft teal `0.94 0.027 195`, highlight yellow `0.92 0.115 100`, lilac `0.94 0.025 300`, peach `0.95 0.035 65`, mint `0.94 0.035 155`; dark counterparts are explicit CSS tokens. SVG artwork and icons share the accent and surface tokens; decoration is hidden from assistive technology. Keep text labels beside navigation icons. Provider errors explain the next step in ordinary language, with original technical context in a disclosure and saved-request retry available.

## Model-assisted source review

Document/key upload forms expose the actual selected OCR engine. OMR model comparison is an explicit checkbox beside calibrated reading, with background progress/retry and clear proposal-versus-measurement labels. Preserve pixel readings when a model fails; mark disagreements without substituting a guessed mark. Engine configuration errors and external model claims never appear as delivered academic evidence.

## Processing clarity — 2026-11-18

Put current work before setup: paper upload shows Upload → Extract → Review → Save, collapses completed setup, and puts the extraction panel above the review. Distinguish upload transfer from saved queued/running work. Use one labelled spinner/progress pattern for paper extraction, generation, keys and sheet reading/model comparison. Only display percentages reported by Go; otherwise use indeterminate progress and no invented time estimate. Completion gives a concrete next action; failures retain source and expose recovery. Announce state changes politely, respect reduced motion, and give Upload paper its own selected navigation entry. Background tasks refresh automatically; transient save notices clear when changing sections.


## Teaching desk — 2026-11-19

Artifact: staff dashboard/data tool. Audience: coordinators, instructors and other authorized teaching staff. Character: calm, approachable, direct, legible. Essence: questions into evidence. Transpose the supplied fourth reference into a wide preparation area and compact workspace/task column, with pastel blue/yellow/lilac surfaces for the three real creation methods. Keep the bundled IBM Plex Sans and existing 1.2 type scale. Signature: numbered question-entry methods beside the next grading tasks. Retain original response-sheet artwork, real counts, scoped records and labelled navigation.

The home page exposes Upload a paper, Enter manually and Create with AI. Manual entry has a dedicated deep-linkable view and explicitly selected answer keys. Existing OCR normalization and editable source/option review remain connected. Empty extraction is a recovery state, with a direct link to an opened transcription editor; zero rows cannot be saved or advertised as ready.

Marking rules select printed import numbers or the displayed bank numbers using inclusive ranges and comma lists. Their preview counts selected questions, not an invented score. Each versioned Go write validates and persists marks; published assessment snapshots retain their earlier rules. Save edits before bulk updates. Bank updates create drafts that need approval again. Bulk updates use existing sequential revision writes; if interrupted, already saved changes persist and remaining work can be retried.

Individual/batch OMR controls accept real files, show durable job progress and per-file upload failures, and restore available jobs on resuming a group. Registered sheet identity maps to assigned students; ambiguous responses and identities retain explicit review. The results table displays server scores, pending-mark labels, correct/wrong counts and numbered correct questions from the pinned assessment order. Search filters students; desktop tables remain horizontally scrollable on phones.

Audit: inspected light desktop, 390px mobile, dark desktop and student-table renders. Labelled forms, visible focus, keyboard navigation, reduced-motion behavior and no viewport overflow exercised in browser journeys. No fabricated metrics or live AI-success claims. Full WCAG and arbitrary handwriting/field-scan qualification remain open.


## Reference dashboard with Taste — 2026-11-19

Software: authenticated assessment workspace for teaching staff. Follow the owner's supplied compact dashboard rather than the skill's marketing-page hero, full-screen sections, random palette deals, mock UI fragments, shaders and promotional moments. Voice: direct product labels. Real material: scoped Go paper/topic/job records and authenticated person name. Primary action: upload a paper. No assumed metrics, calendar deadlines, chat, upgrade claims, avatars or external stock photographs.

The dashboard now has a two-by-two count grid (saved/published/draft assessments and topics), pastel method rows, compact saved-paper rows and a right tools column. Counts are computed from the current scoped API records; paper rows display their actual title/reference/question count/maximum and open the saved paper. Search matches title/reference. Background work lists actual active/failed/configuration-required jobs and opens the existing job view. OMR/results shortcuts and an account chip open the real screens. Sample files remain an explicitly labelled working feature.

Keep IBM Plex Sans, the shared blue accent and explicit dark tokens. Add rose `oklch(0.94 0.038 345)` / dark `oklch(0.3 0.028 345)` for draft counts; retain soft blue, lilac and pale yellow. Use 12px row/control and 16px panel radii, flat fills and aligned edges. Typography: 13–16px labels, 26px counts, 21px page heading. Hover lifts only shortcut cards by 2px over 160ms; reduced motion suppresses it. Responsive rail/menu, stacked tools and horizontally contained content keep features available at 1440/1024/768/390.

Audit: authenticated browser checks compare displayed counts to real API responses, search/open real papers and navigate to OMR. Four widths and dark mode are rendered and inspected. Taste's unauthenticated generic gate redirects to the existing sign-in page: its marketing-layout findings do not establish dashboard quality. See docs/TEST-REPORT.md for the raw gate outcome and limits. Full WCAG certification remains open.


## Milky light-only refinement — 2026-11-19

The owner explicitly requests a soft frosted-glass interface. All dark CSS branches are removed, the server HTML starts with `data-theme=light`, and saved dark/system settings are migrated to light on mount. OS dark preference cannot change the palette. Keep the two-by-two overview and three-zone layout, soft outlined icon tiles, dotted stat connectors, low-opacity paper watermarks, pastel recent-paper strips, and the right illustration/tools column. Accent fills are reserved for selected navigation and primary actions. The positive publication percentage is derived from the loaded saved/published paper counts; no invented learning/score percentages are shown.

`public/art/assessment-desk.webp` is an original generated flat editorial illustration (900×600, ~39KB) for the real grading shortcut; no reference character/logo copied. It is decorative, with empty alt text, and never supplies academic data. The calendar/chat/promo examples remain omitted because those features are not connected. Top-bar job and AI actions open the existing permitted routes. Manrope replaces the older font dependency. Authenticated tests emulate a dark OS and seed an old dark preference, verify migration and reload, confirm the image decodes, compare API counts, search/open real papers, and check four screen widths.

## Assessment-first navigation — 2026-11-19

The five primary faculty sections are Draft assessment, Active assessments, Upload OMR answer sheets, Students, and Results. Draft is the faculty starting page; the logo still opens the existing overview. Separate upload/bank/AI links no longer compete in primary navigation. Older routes remain available for existing deep links; sample/background work moves into Tools & resources. Administrative functions retain their existing role restrictions.

Draft assessment initially presents exactly three creation methods: manual, upload, AI. A closed saved-drafts disclosure provides recovery. The selected method leads to assessment name, course-backed student batch/selected roster, date/time, duration, paper type, instructions and main subject/topic. A persistent server draft then owns questions, source import and generation job IDs. Upload shows a file/drop target; OCR/local-model progress and errors stay on the same paper. AI generation uses repeatable groups with subject, topic/range, count and difficulty. The common editor supports manual/AI additions, answer selection, question editing/removal/order, all/range/list marking and explicit save/review/publication. Published edits create new drafts instead of rewriting already assigned content. Student rosters are scoped to existing faculty responsibilities; admission administration remains restricted.
