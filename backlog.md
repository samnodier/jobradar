# Jobradar — Task Backlog

Last updated: 2026-05-27
Current phase: **Phase 5 — AI Matching**

---

## Current Phase — Phase 5: AI Matching

### Immediate — Fix main.go bugs (blocking)
- [ ] [Phase 5] Fix duplicate `RegisterHandler` call for `queue.JobMatchJob` in `main.go` (lines 62–67)
- [ ] [Phase 5] Resolve `matchJobWorker` method — either implement it or remove the dead registration
- [ ] [Phase 5] Remove redundant `defer wp.Stop()` (Stop is already called explicitly on shutdown)

### Algorithm improvements
- [ ] [Phase 5] Improve matcher signal: weight skills by proficiency/years_experience from user_skills
- [ ] [Phase 5] Add location preference matching (user_desired_locations vs job.location/is_remote)
- [ ] [Phase 5] Tune Jaro-Winkler threshold (currently 0.55) — validate against real job data

### LLM enrichment (filter-then-enrich)
- [ ] [Phase 5] LLM enrichment only runs on jobs with algorithmic score ≥ threshold (e.g. 60)
- [ ] [Phase 5] Write Gemini prompt for structured JSON: `summary` (1 sentence), semantic skill validation
- [ ] [Phase 5] Defensive JSON parsing for LLM response (model will occasionally break format)
- [ ] [Phase 5] Write LEARNINGS.md entry: how to prompt LLMs for reliable structured JSON output

### User API key management (Phase 5.5)
- [ ] [Phase 5.5] New migration: add `encrypted_gemini_api_key TEXT` to users table
- [ ] [Phase 5.5] Go: AES-256 encrypt/decrypt helpers using `ENCRYPTION_KEY` env var
- [ ] [Phase 5.5] `PATCH /api/users/me/preferences` — add gemini_api_key field
- [ ] [Phase 5.5] Settings page in Vue: input field + link to aistudio.google.com/apikey
- [ ] [Phase 5.5] Write LEARNINGS.md entry: encrypting secrets at rest, key management tradeoffs

### Profile/matching frontend
- [ ] [Phase 5] Display match_score, matched_skills, missing_skills on job cards
- [ ] [Phase 5] Expose `min_match_score` threshold in user preferences UI

---

## Phase 6 — Auto Resume Generation (Not Started)

- [ ] [Phase 6] When job scores above threshold: trigger resume generation worker
- [ ] [Phase 6] LLM-powered tailored resume using user profile + job description
- [ ] [Phase 6] HTML-to-PDF generation
- [ ] [Phase 6] Send resume via email (Resend or Mailgun) with job details and match score attached

---

## Phase 7 — Notifications (Not Started)

- [ ] [Phase 7] Email alert when job scores above `min_match_score` (Resend or Mailgun)
- [ ] [Phase 7] Idempotency — never send the same alert twice (DB flag)
- [ ] [Phase 7] `GET /api/notifications` — list unread notifications for logged-in user
- [ ] [Phase 7] In-app notification bell component in Vue navbar
- [ ] [Phase 7] Write LEARNINGS.md entry: transactional email vs marketing email, idempotency in notification systems

---

## Phase 8 — More Job Sources (Not Started)

- [ ] [Phase 8] Remotive fetcher
- [ ] [Phase 8] Adzuna fetcher
- [ ] [Phase 8] Portal scanner: Greenhouse, Ashby, Lever
- [ ] [Phase 8] Each new source is a new fetcher only — queue and workers stay unchanged
- [ ] [Phase 8] Write LEARNINGS.md entry: Open/Closed Principle in practice

---

## Phase 9 — Discord Bot (Optional)

- [ ] [Phase 9] Register Discord app, use `discordgo`
- [ ] [Phase 9] Auto-post high-match jobs to a channel
- [ ] [Phase 9] `/jobs senior golang` slash command

---

## Phase 10 — Observability and Production Polish (Not Started)

- [ ] [Phase 10] Replace all `fmt.Println` / `log.Printf` with `slog` structured logging
- [ ] [Phase 10] Consistent log fields: `timestamp`, `level`, `source`, `job_id`, `user_id`, `duration_ms`
- [ ] [Phase 10] `GET /api/metrics` — jobs fetched today, processed total, alerts sent, dead letter count, avg match score, active sessions
- [ ] [Phase 10] Token bucket rate limiting — 100 req/hour per user — implement the algorithm, no library
- [ ] [Phase 10] Cache-Control headers
- [ ] [Phase 10] Architecture README with ASCII diagram, local setup, service descriptions, decisions
- [ ] [Phase 10] Deploy to Railway, Render, or DigitalOcean
- [ ] [Phase 10] Write LEARNINGS.md entry: structured logs, observability, token bucket, dev vs prod environments

---

## Technical Debt

- [ ] Move `Session` struct to `internal/session` package — currently duplicated between `internal/auth` and `cmd/jobradar` (tracked in DECISIONS.md)
- [ ] `GET /api/saved-jobs` — route exists in table but not yet built (see ROUTES.md)
- [ ] `DELETE /api/saved-jobs/{id}` — not yet built
- [ ] `PUT /api/applications/{id}/status` — not yet built
- [ ] `PUT /api/applications/{id}/notes` — not yet built
- [ ] `PUT /api/applications/{id}/followup` — not yet built
- [ ] `DELETE /api/applications/{id}` — not yet built
- [ ] Update `.context/SCHEMA.md` to reflect all resume profile tables (skills, user_experiences, user_education, user_projects, user_certifications, user_languages, user_skills, user_desired_locations, user_desired_roles, experience_skills, project_skills)
- [ ] Update `.context/ROUTES.md` to reflect all profile/resume endpoints

---

## Done

- [x] [Phase 1] GitHub OAuth flow with pending signup pattern in Redis
- [x] [Phase 1] Session management (Redis-backed, HttpOnly cookies, TTL)
- [x] [Phase 1] Onboarding flow with atomic DB transaction (user + provider account)
- [x] [Phase 1] `GET /api/users/me`, `POST /auth/logout`, end-to-end auth test
- [x] [Phase 1] Chi router, CORS, Docker Compose, Makefile
- [x] [Phase 1] SQLC, pgx, migrations with Goose timestamps
- [x] [Phase 2] `GET /api/jobs`, `GET /api/jobs/{jobID}` with RequireAuth
- [x] [Phase 2] `GET /api/jobs/stats`
- [x] [Phase 2] RemoteOK fetcher running every 6 hours
- [x] [Phase 2] Saved jobs CRUD (`POST /api/saved_jobs`, hybrid SQL join for is_saved flag)
- [x] [Phase 2] Applications CRUD with Kanban status transitions, notes, follow-up dates
- [x] [Phase 2.5] User profile: Work History (user_experiences + experience_skills)
- [x] [Phase 2.5] User profile: Education, Projects, Certifications, Languages
- [x] [Phase 2.5] Skills table with user_skills junction, proficiency, years_experience
- [x] [Phase 2.5] Desired roles and desired locations tables
- [x] [Phase 3] Vue 3 + TypeScript + Vite + Pinia + Vue Router frontend
- [x] [Phase 3] JobsView, JobDetailView, ProfileView
- [x] [Phase 3] Navigation guard, Pinia auth store
- [x] [Phase 3] Applications Kanban board frontend
- [x] [Phase 3] Linear-inspired design system, CSS variables
- [x] [Phase 3] SPA fallback routing served from Go
- [x] [Phase 4] Redis list queue (LPUSH/BRPOP)
- [x] [Phase 4] Worker goroutines with graceful shutdown (context cancellation + WaitGroup)
- [x] [Phase 4] Exponential backoff (1s→2s→4s→8s, max 3 retries) + dead letter queue
- [x] [Phase 4] Scheduler with `time.Ticker` (every 6 hours)
- [x] [Phase 4] Algorithmic job matcher: Jaro-Winkler title similarity, Aho-Corasick skill matching, token sorting for swapped titles
- [x] [Phase 5] Profile compiler in `worker_match.go`: fetches desired roles, skills, experiences from DB
- [x] [Phase 5] `UpdateJobMatchingResult` writes match_score, matched_skills, missing_skills, ai_summary to jobs table
- [x] [Phase 5] Worker wired to `queue.JobMatchJob` job type (has bugs — fix in progress)
