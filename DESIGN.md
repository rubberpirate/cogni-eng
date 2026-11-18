# Learning Workspace visual system

Institutional application and public introduction. Audience: staff organizing real work, learners concentrating, families understanding evidence. Primary first action: choose an institution and work context. Character: calm, precise, legible, accountable. Essence: records with clarity.

Use a light teaching studio informed by the owner-supplied references: a white labelled navigation rail, pale teal work surface, rounded panels and restrained peach/yellow/lilac assessment covers. The signature is original response-sheet line artwork paired with direct teaching actions. Calm, approachable, precise, legible; essence: room for discovery. Public content uses the same typography without displaying invented product metrics.

## Tokens

Body: locally bundled IBM Plex Sans; headings use the same family at 600. Base 16px, ratio 1.2: 13,16,19,23,28,33. Spacing unit 4px; component gaps 8/12/16; section gaps 24/32/48. Radii 12px for controls and 24px for panels. Defined light borders, no decorative shadow.

| Role | Light OKLCH | Dark OKLCH |
|---|---|---|
| background | 0.975 0.008 195 | 0.17 0.014 160 |
| surface | 1 0 0 | 0.22 0.015 160 |
| foreground | 0.25 0.025 210 | 0.94 0.008 160 |
| muted | 0.48 0.018 210 | 0.74 0.018 160 |
| border | 0.9 0.012 200 | 0.40 0.02 160 |
| accent | 0.32 0.055 218 | 0.78 0.09 165 |
| accent foreground | 0.99 0 0 | 0.18 0.015 160 |
| success | 0.39 0.09 150 | 0.78 0.10 150 |
| warning | 0.42 0.08 65 | 0.82 0.10 65 |
| error | 0.43 0.15 25 | 0.78 0.10 25 |

Semantic HTML, actual labels, visible focus, 44px interaction targets, text with states, inline errors and a status region. Tables use left-aligned labels and right-aligned tabular numbers; narrow screens stack work forms and scroll tables in labeled regions. Theme and text-size preferences affect all components. Motion limited to 120ms opacity; reduced motion disables it. No stock imagery needed for record workflows.

Audit: rendered desktop, narrow-screen and dark-theme review completed for the current teaching overview/sign-in. WCAG conformity is not yet established.

## Teaching navigation refinement — 2026-11-18

Staff enter Home and see the sequence from questions to assessment to evidence. Direct task links expose AI generation, bank, assessments, scanning, results and sample files; organization and administration use separate disclosures. Learners enter My assessments. Mobile Menu retains keyboard navigation and avoids horizontal overflow. Hash state supports browser back and selected-paper refresh.

Keep each flow's next action close to its input. Show real loading, empty and provider-error states; poll active work without discarding editors. Numerical rules and rubric criteria use normal labelled fields. Uncertain single-choice readings select no radio until an explicit correction. Advanced scoring/layout options are disclosed on demand. Rendered desktop/mobile journeys were inspected and tested; full WCAG qualification remains open.

## Owner reference refinement — 2026-11-18

Transpose the supplied dashboard references through original content and geometry; no borrowed names, logos, illustrations, layouts or fabricated metrics. Show real paper/topic counts and actual job states. Assessment cards expose the saved title, publication state, pinned question count and maximum score, and open the actual assessment. A four-action tool row follows the educational sequence. Responsive forms and scan/review inherit the shared softer panel/control system. Sign-in uses a split introduction/form at desktop and a compact header on phones; all authentication/recovery controls retain their behavior.

Pastel roles: soft teal `0.94 0.027 195`, highlight yellow `0.92 0.115 100`, lilac `0.94 0.025 300`, peach `0.95 0.035 65`, mint `0.94 0.035 155`; dark counterparts are explicit CSS tokens. SVG artwork and icons share the accent and surface tokens; decoration is hidden from assistive technology. Keep text labels beside navigation icons. Provider errors explain the next step in ordinary language, with original technical context in a disclosure and saved-request retry available.

## Model-assisted source review

Document/key upload forms expose the actual selected OCR engine. OMR model comparison is an explicit checkbox beside calibrated reading, with background progress/retry and clear proposal-versus-measurement labels. Preserve pixel readings when a model fails; mark disagreements without substituting a guessed mark. Engine configuration errors and external model claims never appear as delivered academic evidence.
