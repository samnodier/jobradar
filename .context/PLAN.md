## Current status — personal release

Implementation branch: `feature/profile-workflows`, created from `dev` and fast-forwarded to the existing job-import work. Integration target: `dev`; production promotion to `main` follows live account verification. September 2026 implementation targets a personal, deployable daily job digest. Do not treat the older phase numbering as release dependencies.

### Implemented

- Email/password login and optional GitHub OAuth share Redis sessions; optional PERSONAL_EMAIL restriction.
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
- Live Adzuna startup polling succeeded on September 21 (232 records saved); RemoteOK saved 99. No real email has been sent; provider delivery remains to be tested.

### Remaining configuration for personal use

- Review roles, locations and matching settings in the profile; Adzuna credentials are configured and live polling succeeded.
- Confirm recipient, sender and provider; set timezone/hour in Profile → Settings; configure either Resend or GoHighLevel.
- Sign in, fill matching preferences, preview a digest, then exercise actual provider delivery.
- Deploy behind HTTPS, configure OAuth callback, run migrations, verify backups and one week of scheduled operation.

### Deliberately deferred beyond personal v1

In-app tailored resume generation, resume upload, Discord, broader ATS coverage, public multi-user onboarding, calibrated semantic ranking, cross-source fuzzy deduplication and reliable expired-job detection from partial feeds. Scores do not establish work authorization or geographic eligibility. These are not prerequisites for the personal digest.

### Working agreement

Sam authorized direct implementation and maintenance of learning notes. Record review questions instead of blocking progress for quizzes. Keep commits on feature branches, verify before merging into dev, and promote to main as a release. Do not add assistant attribution to commits, PRs, or contributors.

## Current delivery agreement

Work is committed in small focused increments. The prior unpushed release commit was regrouped into 12 commits with identical final tree; a local backup branch preserves the old tip. No remote history was rewritten. Sam will manually test dev before main promotion/push and deployment. A home server is the recommended zero-hosting-bill option for this always-running scheduler; no local language model is required.

## September 21 follow-up

Implemented preference-aware numeric ranking and job metadata flow; read-only account tokens; filtered/paginated job exports and AI-context access; Settings token management. Details: docs/JOB_EXPORT_API.md. New migrations require application before local UI/API testing. Existing dev database has not been changed by tests.

Email/password login is implemented on feature/email-login: Argon2id credentials, recent-session initial password setup, current-password changes, auth-version session invalidation, workflow-token revocation, throttling, one-time server-owner setup/recovery, optional GitHub and explicit account linking. Development database migrated through 20260921000300 after a private backup on September 21. Sam will test real GitHub consent and email login locally before any push.

Local review servers started on September 21: frontend http://localhost:5173, backend http://localhost:8080, development PostgreSQL/Redis on 5432/6379. HTTP checks passed for the login page, auth methods (both enabled), health, job feed, GitHub authorization redirect and invalid email-login rejection. Automatic AI enrichment and email sending remain disabled. Sam sets their own password after GitHub sign-in; no owner password has been generated or recorded by the implementation.

## September 22 screenshot review

Work developed on `feature/review-polish` from `dev`; merge to `dev` follows successful checks, with main/push still awaiting Sam's manual review. No new migration is required for this review. Sam runs their own Go/Vite servers and sees their logs.

- Fixed serialized preference autosave, nullable salary clearing and concurrent backend PATCH updates; added location tags and visible Save/retry state.
- Split matching/delivery into compact shared-style square cards; preserved edits during saves and added account/token component tests.
- Added scoped profile JSON download; not a full account backup.
- Implemented application status drag/drop with select alternative, rollback and correct delayed-note identity.
- Connected source navigation to SQL filtering before pagination; replaced empty Console navigation with actual Health; improved polling/stale-state UI.
- Added opt-in Jobicy and MyJobMag Kenya sources; flags remain off by default. See docs/JOB_SOURCES.md. Live ingestion of these new feeds has not been verified against Sam's database.
- Full backend race/integration suite and go vet passed against isolated services. Frontend component tests, type check and production build passed; manual visual/touch review remains.
- One authorized Resend test attempted September 22: HTTP 403, domain samnodier.com not verified for the configured key. No test message accepted. Cloudflare DNS alone does not establish Resend verification; fix domain/account setup before another test. Scheduled sending remains disabled.
- Home-server recommendation and Supabase/Render tradeoffs: docs/HOSTING_OPTIONS.md. No deployment or push performed.

## Pagination and status-page follow-up

Developed on `feature/feed-pagination-status` from dev. API limit now 500 with exact X-Has-More metadata; top/bottom UI offers 50/100/200/500, Previous/Next and route-preserved page state. Removed browser filtering after pagination. Live diagnosis found the old running backend ignored source=remoteok and returned Adzuna rows despite 791 RemoteOK database rows. Sam must restart their Go process; it has not been restarted by the agent.

Health now has a summary banner and 30 UTC-day source bars backed by a new daily observation table/trigger. Percentages measure collection success, not uptime; gray means no data. Migration 20260922000100 passed isolated up/down/up but **is not yet applied to Sam's development database**. Sam runs `make migrate-up` before restarting `make run`.

Final verification: 37 frontend tests, production build/typecheck, focused ESLint, backend race/integration suite and go vet passed. No push, production deployment, real email send or server restart in this follow-up.


## Feed review and sorting follow-up — September 22

Implemented on `feature/feed-ui-health` from dev, for local dev integration without pushing:
- Bottom-only pagination with icon arrows; compact responsive rows, medium-weight titles, company beneath title and platform source column.
- Sort menu: newest, oldest, highest score, lowest score. SQL ordering happens before pagination; unmatched scores remain last. Best matches retains threshold, eligibility and pending-work exclusions, now explained in the UI.
- Notification bell with unread count and icon logout.
- Confirmed the local health failure was the missing source-history table. Applied migration 20260922000100 to development; API now preserves current checks when history is unavailable.
- Minimum-only salary already supported; explicit boundary/open-upper-bound regression coverage added.
- Frontend tests/build/typecheck and focused lint passed; backend race/integration suite and vet passed against isolated services. Sam retains control of server restart and push.

## Profile PDF and threshold clarification

- Best matches now uses only the stored score threshold, per Sam's clarification; pending/eligibility gates remain for notifications/digests. Backend integration coverage verifies this distinction.
- Export to PDF loads all saved resume sections and opens a clean browser print document; choose Save as PDF. Popup/request failures are surfaced; no partial export on request failure. This exports the saved profile, without job tailoring.
- 43 frontend tests, production build/typecheck and focused lint passed; backend handler race/integration tests passed. Native print dialog still requires manual browser verification.

## GitHub connection and toolbar follow-up

Settings now reads the signed-in user's non-revoked GitHub account link and shows GitHub connected instead of repeating Connect GitHub. Server availability and account connection are separate. Older API responses show unknown state rather than incorrectly claiming disconnection. Feed Refresh/Import now use icons with tooltips and accessible names. 45 frontend tests, production build/typecheck, focused lint and backend auth/handler race tests plus vet passed. No push performed; restart the Go server to use the updated API.


## Morning digest activated — September 22

Implemented on feature/morning-digest from dev. Table-based Jobradar email follows Sam's supplied visual reference with genuine job data and source attribution. The digest now follows score-only Best matches, with fresh/unexpired/pending/dedup constraints retained. Tests cover top-20 ordering, escaping, delayed start and test-send isolation.

Local configuration: sender jobs@mail.example.com; recipient hello@samnodier.com; enabled, minimum score 60, maximum 20, daily 07:00 America/New_York, first delivery date 2026-09-23. 18 fresh matches currently qualify; no jobs have been reserved by the test. Do not lower the threshold or claim exactly 20 will qualify tomorrow.

Real preview accepted by Resend (message (message id omitted)) and Sam confirmed inbox delivery/rendering. Provider receipt lookup separately returned 401. No need to resend this test.

Installed and enabled ~/.config/systemd/user/jobradar-digest.service; `bin/jobradar --digest-worker` is active, loginctl Linger=yes. Service reads repository .env and runs independently of Go/Vite terminals. Computer must remain awake/online with PostgreSQL available. `DIGEST_START_DATE` prevents today's catch-up send. Rebuild bin/jobradar and restart the service after future digest code changes. Source ingestion still runs in the main application.

## Sender branding, logos and timezone chooser

Sender now uses Jobradar as the display name. Digest includes source-supplied HTTPS company logos with initials fallback; Adzuna does not supply these in the current integration. Settings lists 597 IANA timezone names/aliases (2026c), supports one or two local delivery hours, and preserves the existing schedule until saved. Hosting choices expanded in docs/HOSTING_OPTIONS.md; no hosting purchase/deployment authorized or performed.

## Public homepage and recent discovery

Removed synthetic homepage totals, source count, recommendation scores and employment defaults. Homepage uses anonymous API requests for the public latest list and a separate authenticated score-ranked query for recommendations. Remote CTA now selects the remote filter. A short-lived filter=recent (past month) was removed September 22 at Sam's request; the homepage lists the newest public jobs. All stored history remains accessible. No destructive cleanup. 49 frontend tests/build/typecheck/focused lint and backend handler race/integration tests plus vet passed.


## October 5 — digest outage and machine move

- Root cause of zero scheduled digests: development Compose had no restart policy. After a host reboot PostgreSQL stayed down and the digest worker crash-looped silently (`email_digests` had no rows). Fixed with `restart: unless-stopped`.
- Only one machine runs the digest scheduler at a time; development machines set `DIGEST_ENABLED=false`.
- Moving between machines: `docs/MOVING_MACHINES.md` (dump/restore, `.env` copied privately, full test commands).
- Verified before the move: Go build/vet, frontend type-check/build and unit tests. Backend integration suite to be run on the new machine. Known lint debt in test mocks and `src/env.d.ts`.
- Next: supervise the app on an always-on host, decide public access, schedule off-host backups, and alert when an expected digest is missing.

## October 11 — CI and continuous deployment

- `.github/workflows/ci.yml`: backend integration tests (Postgres/Redis services, race detector), vet, frontend type-check/tests/build on every push and PR; publishes app and migration images to GHCR only from `main` after tests pass.
- `compose.production.yml` uses the published images (local `--build` still works).
- Server: `deploy/update.sh` (pull-based; backup → migrate → recreate → health check), `deploy/backup.sh` (nightly, 14 kept), `deploy/install-systemd.sh` (user timers). See `docs/DEPLOYMENT.md`.
- Development machines keep `DIGEST_ENABLED=false`; only production sends.
- Not yet done: first production deploy and data restore, public access choice, production OAuth app, off-host backup copies, missed-digest alerting.

## October 11 — job retention

- Public postings older than `JOB_RETENTION_DAYS` (default 90; posting date, else first-seen) are deleted hourly in batches; saved, applied-to and imported jobs are never deleted. Collection skips postings already past the cutoff. Code: `cmd/jobradar/retention.go`, `DeleteExpiredPublicJobs`.
- Commit messages now use conventional type prefixes (`feat:`, `fix:`, `docs:`, `test:`, `ci:`, `chore:`).

## October 11 — real profile, login recovery, experience scoring

- Owner profile filled from the public resume/LinkedIn (one-off data change on the server, backed up first; not in the repo).
- Login page recovers when a single-use setup reply is lost or the link is spent.
- Experience scoring uses the best single experience (`BestExperienceMatch`); the union version diluted scores as experiences were added. Replayed on real data: 60+ matches ~7 → ~13.
- Open: scores cluster ~55–65, so the 60 threshold is fragile; calibrate with real outcomes (saved/applied) before further weight changes.
