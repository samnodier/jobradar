# Session handoff — reconciliation release

Read this file, `.context/PLAN.md`'s current snapshot, and `docs/RELEASE_REVIEW.md` before continuing. The pasted conversation is no longer needed; it cannot be deleted by the assistant from the chat.

## Delivery state

Work is committed on `feature/release-reconciliation`, created from `dev`. Verified feature work is integrated into `dev`; `main` stays at the current personal release until Sam's manual review. Check `git status`, branch tips and remote CI before assuming any release is deployed.

Implemented: bounded explicit-feedback learning; focused AI preparation exports; application follow-ups/current outcomes; SHA-pinned app/migration artifacts and actual running revision; host monitoring plus safe owner-email retries and visible unconfirmed delivery; backup copy to another device; safe cached posting availability checks in digest preparation; separate Adzuna market location; Jobicy cursor traversal; Kenya/South Africa MyJobMag adapters. Existing login recovery and resume data work are preserved.

New migrations end at `20261011000400`. Run them locally before using the new API. Development must keep `DIGEST_ENABLED=false` and `OPS_EMAIL_ENABLED=false`. Private credentials/data are not in Git; do not replace the stable encryption key while restoring existing encrypted provider keys.

The live screenshot's Adzuna failure was provider HTTP 503, confirmed read-only. A distinct bug passed desired residence locations into incompatible Adzuna markets; the new setting fixes that. Refresh on System health does not collect jobs.

## Next actions

1. Follow `docs/RELEASE_REVIEW.md` for local manual checks on the existing Mac checkout; do not delete the folder or restore over an existing populated database.
2. After review, promote tested `dev` to `main`; CI publishes both SHA-tagged images, then the server updater deploys them together. Confirm the running revision, not merely HTTP 200.
3. Set production source flags and operational-email opt-in privately, recreate the exact reviewed app image and rerun `deploy/install-systemd.sh` to install the monitor timer. These are not activated by editing `.env.example`.
4. Copy a completed backup to another device with `deploy/copy-backup.sh`, preserve private configuration/encryption key separately, and rehearse restore into an isolated empty database.
5. Configure public GitHub OAuth if wanted (existing implementation, account credentials required); verify actual scheduled mail, source relevance, PDF printing and unattended operation.

Verification: backend race/integration and vet; 61 frontend tests/type-check/build/focused lint; six updater simulation tests; all three migrations down/up; packaged image revision/SPA/non-root monitor-state checks. Public feeds were probed read-only; production ingestion and real email delivery remain separate acceptance checks.

Working agreement: direct implementation, focused conventional commits on feature branches from `dev`, learning notes with code links/questions, no quizzes until release works, no assistant attribution. Use agents only when authorized; this task explicitly authorized coordinated parallel agents. Do not send recruiter messages or submit applications automatically.
