# Session handoff — 2026-09-20

Sam authorized direct implementation, full maintenance of LEARNINGS.md, deferred knowledge checks, and no assistant attribution. Preserve feature → dev → main workflow. Historical mentorship-only instructions are superseded by that agreement.

Implemented personal release on `feature/personal-release`, based on dev with the existing job-import commits preserved. Read PLAN.md, SCHEMA.md, ROUTES.md and DECISIONS.md for current behavior. LEARNINGS.md contains explanations, review questions and code links, including cross-language full-stack fundamentals.

Major changes: Adzuna/Greenhouse/RemoteOK ingestion; PostgreSQL-backed recoverable matching; independent optional AI workers; private-job authorization; paginated frontend; persisted daily personal digest with Resend and GoHighLevel adapters; real status/preview UI; production Docker packaging and deployment documentation.

Automated verification and running production-image smoke checks passed. See docs/RELEASE_VERIFICATION.md. No real email was sent, no live Adzuna account was queried, and existing development data was not migrated by this work. Account configuration remains the next step, not another broad implementation phase.

Sam still needs to choose/provide: country/location/roles, Adzuna account configuration, recipient, sender/provider, timezone/hour and deployment destination. Secrets belong only in local/deployment environment configuration. DIGEST_ENABLED stays false until reviewed configuration. Resend has a simple API and idempotency; an already-working GoHighLevel sender is also supported. Verify costs in the actual account.

Keep main as the production promotion step after live verification. Do not describe mocks or synthetic sessions as evidence of actual OAuth or inbox delivery. Do not reintroduce a local model requirement. After deployment works, use the learning log for a guided code walkthrough and answer the recorded questions together.
