# Provider setup and verification matrix

Local configuration is for synthetic demonstrations. Production activation is separate from installation. No charge, refund, external delivery, model prediction, extraction or rendering outcome is inferred from a test double.

| Capability | Source path | Configured local evidence | Production status |
|---|---|---|---|
| Email | internal/delivery/worker.go SMTP adapter | Mailpit SMTP acceptance and captured invitation/verification journey | STARTTLS adapter present; live provider credentials, delivery evidence and security-message completeness pending |
| Private S3 | internal/blob, internal/transport/avatar.go | SeaweedFS put/read/remove, denial and outage/lifecycle tests for profile crops | HTTPS validation present; service credential scopes, backup/encryption and institutional files pending |
| Text/vision/document AI | Unimplemented | None | AIO-01 and associated tutor/content/OCR scope planned |
| Speech/presenter/video | Unimplemented | None | STU-04–09 and AIO-01 planned |
| SMS/push | Unimplemented | None | COM-02 and MOB-06 planned |
| Calendar integration/export | Unimplemented | None | PLN-02 and COM-08 planned |
| Payments/refunds/webhooks | Unimplemented | None | FIN-01–09 planned; no financial activity performed |

SMTP accepts `SMTP_ADDR`, `SMTP_FROM`, `SMTP_USERNAME`, `SMTP_PASSWORD`. APP_ENV=production requires TLS in the adapter. Local development labels all captured messages. Failed operations retain status and retry, with no claim of inbox delivery. A message may be repeated after a crash at SMTP acknowledgment; the underlying security token is single-use.

Object storage accepts OBJECT_ENDPOINT (origin URL), OBJECT_BUCKET, OBJECT_USER, OBJECT_PASSWORD. The local endpoint is host-loopback HTTP; production requires HTTPS. The bucket is initialized privately by SeaweedFS. Profile reads are proxied through the person-authorized Go route, never a published storage URL. Do not expose service credentials to web/native clients.

Official documentation/version checks are recorded in VERSIONS.md. Add each future provider's current official input/output, signature, retry, quota, cancellation and sandbox references with its implemented adapter and tests. This matrix is an honest inventory and does not substitute for the missing adapters.
