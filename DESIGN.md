# Learning Workspace visual system

Institutional application and public introduction. Audience: staff organizing real work, learners concentrating, families understanding evidence. Primary first action: choose an institution and work context. Character: calm, precise, legible, accountable. Essence: records with clarity.

Choose a document-workbench direction over a marketing dashboard or colorful classroom metaphor. A compact dark pine navigation rail anchors a wide light workspace. The signature is a numbered work ledger: actions have clear ownership, state, and an audit reference. Public content uses the same typography without displaying invented product metrics.

## Tokens

Body: locally bundled IBM Plex Sans; headings use the same family at 600. Base 16px, ratio 1.2: 13,16,19,23,28,33. Spacing unit 4px; component gaps 8/12/16; section gaps 24/32/48. Radii 4px and 8px. Defined borders, no decorative shadow.

| Role | Light OKLCH | Dark OKLCH |
|---|---|---|
| background | 0.97 0.008 160 | 0.17 0.014 160 |
| surface | 0.995 0.002 160 | 0.22 0.015 160 |
| foreground | 0.24 0.025 160 | 0.94 0.008 160 |
| muted | 0.46 0.025 160 | 0.74 0.018 160 |
| border | 0.82 0.018 160 | 0.40 0.02 160 |
| accent | 0.37 0.065 165 | 0.78 0.09 165 |
| accent foreground | 0.99 0 0 | 0.18 0.015 160 |
| success | 0.39 0.09 150 | 0.78 0.10 150 |
| warning | 0.42 0.08 65 | 0.82 0.10 65 |
| error | 0.43 0.15 25 | 0.78 0.10 25 |

Semantic HTML, actual labels, visible focus, 44px interaction targets, text with states, inline errors and a status region. Tables use left-aligned labels and right-aligned tabular numbers; narrow screens stack work forms and scroll tables in labeled regions. Theme and text-size preferences affect all components. Motion limited to 120ms opacity; reduced motion disables it. No stock imagery needed for record workflows.

Audit: source-level token and semantic review pending rendered verification. WCAG conformity is not yet established.

## Teaching navigation refinement — 2026-11-18

Staff enter Home and see the sequence from questions to assessment to evidence. Direct task links expose AI generation, bank, assessments, scanning, results and sample files; organization and administration use separate disclosures. Learners enter My assessments. Mobile Menu retains keyboard navigation and avoids horizontal overflow. Hash state supports browser back and selected-paper refresh.

Keep each flow's next action close to its input. Show real loading, empty and provider-error states; poll active work without discarding editors. Numerical rules and rubric criteria use normal labelled fields. Uncertain single-choice readings select no radio until an explicit correction. Advanced scoring/layout options are disclosed on demand. Rendered desktop/mobile journeys were inspected and tested; full WCAG qualification remains open.
