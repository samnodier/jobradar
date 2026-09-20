## Current status — personal release

Implementation branch: `feature/profile-workflows`, created from `dev` and fast-forwarded to the existing job-import work. Integration target: `dev`; production promotion to `main` follows live account verification. September 2026 implementation targets a personal, deployable daily job digest. Do not treat the older phase numbering as release dependencies.

### Implemented

- GitHub OAuth, Redis sessions, optional PERSONAL_EMAIL login restriction.
- PostgreSQL migrations and SQLC; profile/resume sections, preferences, saved jobs, application tracking.
- Public-or-owned job authorization, including save/application insertion and indirect listing.
- RemoteOK, optional Adzuna (12-hour interval), optional Greenhouse board watchlist.
- Shared normalized ingestion, source health, publication dates, update-aware upserts.
- Algorithmic matching independent of AI; corrected required-skill/missing-skill semantics.
- Transactional pending-match records created by job/profile triggers; version-aware completion and recovery after lost delivery.
- Separate optional AI queue; bounded requests/prompts and provider cooldown-aware retries.
- URL import and manual entry; Groq/Gemini extraction fallback.
- Profile matching settings: score threshold, work mode, unknown-field policy, supported Adzuna market, timezone, 1–2 delivery hours, max jobs and separate channel toggles.
- Profile-driven Adzuna queries; role/skill/experience ranking plus eligibility gates for work mode, desired locations, employment type and compatible salary ranges.
- Best Matches feed/profile tab, durable deduplicated in-app notifications and authorized deep links.
- Markdown Copy for AI with safe profile export; structured JobPosting extraction and no-key text fallback.
- Personal email digest: fixed recipient, opt-out, timezone, threshold, persisted payload/ledger, stable Resend idempotency keys, conservative GoHighLevel ambiguous-send handling.
- Digest preview CLI/endpoint; dashboard shows real delivery state; status page reads real dependencies/source health.
- Paginated/searchable feed, empty-list normalization, import refresh, production SPA serving.
- Production application/migration Dockerfiles, Compose examples, deployment guide.
- LEARNINGS.md: explanations, failure scenarios, questions and implementation links for later walkthrough.

### Release verification

- Clean migration up and new-migration down/up tested against isolated PostgreSQL 16.
- Backend unit/integration suite with race detector and go vet passed after implementation.
- Frontend production build, type checks and shell test passed with updated dependencies.
- Compatible npm security updates applied; audit reported zero vulnerabilities.
- Production application and migration images built. Running-image HTTP checks passed for SPA routing, empty feeds, validation, authentication guards, manual import, private visibility, save, search, pagination and status endpoints. See docs/RELEASE_VERIFICATION.md.
- Live Adzuna and live email require Sam's account configuration. No real email has been sent during implementation.

### Remaining configuration for personal use

- Sam fills roles, locations and matching settings in the profile; set Adzuna application credentials in the environment.
- Confirm recipient, sender and provider; set timezone/hour in Profile → Settings; configure either Resend or GoHighLevel.
- Sign in, fill matching preferences, preview a digest, then exercise actual provider delivery.
- Deploy behind HTTPS, configure OAuth callback, run migrations, verify backups and one week of scheduled operation.

### Deliberately deferred beyond personal v1

In-app tailored resume generation, resume upload, Discord, broader ATS coverage, public multi-user onboarding, calibrated semantic ranking, cross-source fuzzy deduplication and reliable expired-job detection from partial feeds. Scores do not establish work authorization or geographic eligibility. These are not prerequisites for the personal digest.

### Working agreement

Sam authorized direct implementation and maintenance of learning notes. Record review questions instead of blocking progress for quizzes. Keep commits on feature branches, verify before merging into dev, and promote to main as a release. Do not add assistant attribution to commits, PRs, or contributors.

## Current delivery agreement

Work is committed in small focused increments. The prior unpushed release commit was regrouped into 12 commits with identical final tree; a local backup branch preserves the old tip. No remote history was rewritten. Sam will manually test dev before main promotion/push and deployment. A home server is the recommended zero-hosting-bill option for this always-running scheduler; no local language model is required.
