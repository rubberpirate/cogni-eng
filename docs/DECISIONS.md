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
