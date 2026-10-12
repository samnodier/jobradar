# Session handoff — reconciliation release

Read this file, `.context/PLAN.md`'s current snapshot, and `docs/RELEASE_REVIEW.md` before continuing. The pasted conversation is no longer needed; it cannot be deleted by the assistant from the chat.

## Delivery state

Work is committed on `feature/release-reconciliation`, created from `dev`. Verified feature work is integrated into `dev`; `main` was promoted to `f6de189` with Sam's explicit one-time approval after automated testing. Future releases still wait for manual review/confirmation. Production now reports the exact embedded revision `f6de189`; its three migrations and pre-deploy backup completed successfully. Source collection and the independent monitor are verified. Check branch tips and runtime version before assuming subsequent work is deployed.

Implemented: bounded explicit-feedback learning; focused AI preparation exports; application follow-ups/current outcomes; SHA-pinned app/migration artifacts and actual running revision; host monitoring plus safe owner-email retries and visible unconfirmed delivery; backup copy to another device; safe cached posting availability checks in digest preparation; separate Adzuna market location; Jobicy cursor traversal; Kenya/South Africa MyJobMag adapters. Existing login recovery and resume data work are preserved.

New migrations end at `20261011000400`. Run them locally before using the new API. Development must keep `DIGEST_ENABLED=false` and `OPS_EMAIL_ENABLED=false`. Private credentials/data are not in Git; do not replace the stable encryption key while restoring existing encrypted provider keys.

The live screenshot's Adzuna failure was provider HTTP 503, confirmed read-only. A distinct bug passed desired residence locations into incompatible Adzuna markets; the new setting fixes that. Refresh on System health does not collect jobs.

## Next actions

1. Follow `docs/RELEASE_REVIEW.md` for local manual checks on the existing Mac checkout; do not delete the folder or restore over an existing populated database.
2. This release was explicitly authorized and pushed to main; confirm production reports `f6de189` and complete the remaining manual checks. Future releases must still wait for Sam's manual review/confirmation.
3. Production source flags and operational-email opt-in are now enabled, with the new monitor timer installed. All five sources are healthy; assess the collected jobs for relevance. Development must keep digest/operational email disabled.
4. Copy a completed backup to another device with `deploy/copy-backup.sh`, preserve private configuration/encryption key separately, and rehearse restore into an isolated empty database.
5. Configure public GitHub OAuth if wanted (existing implementation, account credentials required); verify actual scheduled mail, source relevance, PDF printing and unattended operation.

Verification: backend race/integration and vet; 61 frontend tests/type-check/build/focused lint; six updater simulation tests; all three migrations down/up; packaged image revision/SPA/non-root monitor-state checks. Public feeds were probed read-only; production ingestion and real email delivery remain separate acceptance checks.

Working agreement: direct implementation, focused conventional commits on feature branches from `dev`, learning notes with code links/questions, no quizzes until release works, no assistant attribution. Use agents only when authorized; this task explicitly authorized coordinated parallel agents. Do not send recruiter messages or submit applications automatically.
