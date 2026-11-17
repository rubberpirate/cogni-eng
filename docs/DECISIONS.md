# Decision log

2026-11-17

- Follow the supplied Go/Next.js/PostgreSQL/Expo architecture. Use one Go module and manual constructors; no framework is needed for routing or dependency injection. This is already authorized by the build prompt's reversible defaults, overriding skill preferences to ask architecture questions.
- Keep capability packages small and expand only with working vertical flows. No role-specific backends.
- Use Bun 1.4.4. Pin exact dependencies and retain bun.lock/go.sum. Official version references are in VERSIONS.md.
- Start with opaque hashed server sessions, PBKDF2-HMAC-SHA256 passwords (600,000 iterations), PostgreSQL rate limits, TOTP and hashed single-use recovery codes. Encrypt TOTP secrets using an environment-supplied AES-GCM key. No reusable browser token in local storage.
- Use PostgreSQL-backed transactional work. Queue choice must meet leases, retries, cancellation and recovery before claiming RUN-01; initial outbox schema alone does not meet that requirement.
- Initial demo identities are synthetic. Demo creation requires APP_ENV=development and a random local password, never a committed credential. Production bootstrap is explicit, one time and locked transactionally.
- Use a restrained record-workspace design: calm, precise, legible, accountable. Dense operations use tables and work forms; learner screens prioritize reading. No unsupported analytics or outcome claims.
- UI translations are centralized English/Hindi dictionaries. Dates display in the selected institution's timezone. Incomplete translations remain tracked under UX-06.
- No acceptance state is promoted by a build alone. All 233 requirements begin planned and retain the full source detail. Partial features remain building until every subcheck is demonstrated.

Core education priority (2026-11-17):

- Implement the user's OMR/question-key upload/topic-generation priority as connected vertical workflows across milestones 2–5. Preserve every later requirement in the register; no milestone is declared complete by this increment.
- Add `academic` for pure exact marking/import/registered-sheet geometry and `scholar` for tenant-authorized transactional workflows. Go remains the sole assessment business authority. DTO schemas are generated from Go JSON-tagged structs into the versioned OpenAPI contract, then TypeScript types.
- Select maintained River v0.51.2 over expanding the narrow mail queue. Apply River's own SQL migrator using the same PostgreSQL search_path; enqueue in the committing business transaction. Keep separately runnable workers and authorized job recovery UI.
- Publish immutable paper JSON snapshots of exact question revisions. Database triggers prevent content replacement. Correct bank material through new question revisions; never mutate delivered keys. Signed session writer ownership, server-assigned answer order and server deadlines are authoritative.
- Use original registered response sheets with a public documented geometry/identity grid and honest confidence thresholds. Keep measurement independent of the key; human review gates uncertainty. Do not claim arbitrary-sheet or handwriting accuracy.
- Use real bounded Poppler/Tesseract processing and the official Anthropic Messages adapter. Development generation is explicitly labelled and opt-in; no provider credentials means recoverable configuration status. Production delivery/evaluation is separate from fixture/contract evidence.
- Optional groups currently require equal item weights/penalties; full weighted/partial-credit policies remain tracked. In-memory exam reconnection is delivered; encrypted offline persistence remains required. See CORE-EDUCATION.md for limits and recovery.

2026-11-18 manual testing and Gemini:

- The owner explicitly selected Gemini and supplied a key. Store it only in ignored mode-0600 `.env`, authenticate in an HTTP header and provide worker-only container secret environment wiring. No committed credential or frontend key.
- Implement the verified official `generateContent` JSON Schema contract with an explicit configured model alias and retain the actual resolved response model version. Reject incomplete/blocked output and require teacher approval. HTTP contract tests distinguish themselves from a live model check.
- Create independent development-only manual actors with random unique passwords and a protected ignored manifest. Seed active relationships and original content without bypassing authorization in application routes. Preserve fixture changes on rerun. Family/reviewer workspaces are not misrepresented as complete.
- Exact question duplicates are scoped to institution and topic, so a North subject cannot accidentally deduplicate into a South course or disclose another subject's record. Serialize concurrent duplicate creation inside the business transaction.
