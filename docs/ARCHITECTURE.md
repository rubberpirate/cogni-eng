# Architecture and trust boundaries

Working identity: Learning Workspace. This is a neutral descriptive identity.

One Go application owns access, organization, academic evidence, learning, communication, commerce and operations. `cmd/gateway` serves HTTP; `cmd/worker` executes durable work; `cmd/operator` performs explicit offline administration. Dependencies are passed by constructors. PostgreSQL stores authoritative state. Next.js renders public pages and authenticated tools and proxies `/api/v1` to Go. Expo will consume the same generated contracts; it is required and remains planned until implemented.

## Vocabulary

Person is a global identity. Membership grants capabilities in an institution, optionally at a campus. Enrollment is a learner's academic placement, independent of membership. Family link is an approved relationship and never grants answering permission. Responsibility scopes staff access. Paper revision is immutable delivered content; assignment is recipient eligibility; sitting is an actual examination; marking revision is a reproducible result; release grants visibility. A job is durable intent; a provider attempt is not a completed side effect.

## Authorization matrix

| Capability | Owner | Coordinator | Instructor | Learner | Guardian | Operator |
|---|---|---|---|---|---|---|
| organization.manage | Institution | No | No | No | No | No |
| campus.manage | Institution | Assigned campus | No | No | No | No |
| people.manage | Institution | Assigned campus | No | No | No | No |
| people.read | Institution | Assigned campus | Assigned roster | Self | Linked children | No |
| access.manage | Institution; no self escalation | No | No | No | No | No |
| audit.read | Institution | No | No | Own security activity | Own security activity | Sanitized service events |
| academic.author | Explicit subject grant | Explicit subject grant | Assigned subjects | No | No | No |
| academic.answer | Own assignment only | Own assignment only | Own assignment only | Own assignment | No | No |
| finance.manage | Separately granted | Separately granted | No | No | Payer documents | No |

Role templates are membership defaults, not universal trust. Each resource operation must check person state, live session, institution state, membership capability and campus/relationship. A null campus scope means institution-wide only when explicitly granted. Operator access to child records requires an expiring justified support grant; there is no operator bypass. Unsupported scopes fail closed.

## Data and transactions

Explicit SQL migrations with Goose; pgx typed scanning. Composite foreign keys prevent cross-institution campus references. Person IDs and public references do not encode campus. Mutating access operations, audit and durable work intent commit together. Suspension revokes sessions; reactivation never restores old sessions. Owner changes serialize on the institution row and may not orphan ownership. No cascading removal of academic/financial history.

## API conventions

`/api/v1`, JSON, UTC RFC3339 instants and explicit institution timezone. Opaque HttpOnly SameSite cookies, CSRF header matching the issued session nonce and strict Origin on mutations. Cookie credentials never enter localStorage. Native authentication requires a separately designed secure credential flow; browser CSRF exemptions are prohibited. Structured errors contain code, message and request ID. Unknown JSON fields and oversized bodies fail. List operations are bounded. Resource versions protect concurrent edits. Business idempotency uses durable unique keys. Long jobs expose persisted status and pollable progress; streaming will supplement polling. OpenAPI generates shared TypeScript types; a single transport consumes them.

## Initial operational scope

No external messages or charges are claimed. Development recovery/invitation tokens are visible only through a local operator command and a clearly labeled local inbox. Production requires a real delivery adapter and integration evidence. Object storage is private and isolated; routes must check entitlement before streaming any object. No file feature is considered delivered merely because storage runs.
