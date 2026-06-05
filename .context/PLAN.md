## Current Status

Phase: Phase 5 — AI Matching
Status: In Progress — enrichment scaffolding built (Enricher interface + gemini client + queue plumbing); live wiring pending

## The Vision — What Jobradar Becomes

Jobradar is a full job search command center inspired by career-ops (<https://github.com/santifer/career-ops>):

1. **Fetch jobs** from multiple APIs and portal scanners
2. **Score jobs** with AI — match score, matched/missing skills, one-line summary
3. **Auto-generate a tailored resume** for high-match jobs — ATS-optimized PDF customized per job description — delivered to your email
4. **Notify you** when a high-match job appears — email with score, summary, and resume attached
5. **Track applications** — Kanban board, status, follow-up dates, notes
6. **Scan portals automatically** — Greenhouse, Ashby, Lever, company pages
7. **Serve everything** through a clean Linear-inspired dashboard

## Build Phases

### Phase 2 — Saved Jobs and Applications API

- **Status:** Done
- **Description:** Go handlers for all saved_jobs and applications routes are built. Frontend tracking implemented with hybrid SQL views and auto-saving notes.

### Phase 2.5 — Resume Profile Vertical Slice

- **Status:** Done
- **Description:** Build out a detailed resume-like profile with Work History, Achievements, and Skills. This serves as the foundation for AI matching and automated resume generation.

### Phase 3 — Frontend MVP

- **Status:** Done
- **Description:** Build out JobsView, ProfileView with real data. Save/unsave buttons on job cards. Applications tracker (Kanban board). Delete account button on profile. Navigation guard working correctly.

### Phase 4 — Job Queue and Worker Engine

- **Status:** Done
- **Description:** Redis list as queue (LPUSH/BRPOP). Worker goroutines with graceful shutdown via context cancellation. Exponential backoff (1s→2s→4s→8s, max 3 retries). Dead letter queue. Scheduler with time.Ticker.

### Phase 5 — AI Matching

- **Status:** In Progress
- **Built so far:**
  - Algorithmic matcher in `internal/matcher/`: Jaro-Winkler title similarity, Aho-Corasick skill matching, token sorting for swapped word order
  - `user_job_matches` table — per-user match scores replacing the old per-job columns
  - Scraper fans out one `JobMatchJob` per registered user after each new job saved
  - Worker payload carries `{job_id, user_id}`; writes to `UpsertMatch` (atomic insert-or-update)
  - `GetJobs` and `GetJobByID` LEFT JOIN `user_job_matches` — match data flows to API automatically
  - `SCRAPE_INTERVAL` configurable via env var (default 6h)
  - `UpdateMatchEnrichment` query — narrow write (`ai_summary` + `is_enriched = TRUE` only); no COALESCE (system write is authoritative)
  - `internal/llm` package — provider-agnostic `Enricher` interface, `EnrichmentInput`/`EnrichmentResult` types, and `geminiEnricher` implementing it via `google.golang.org/genai` with structured output (`ResponseSchema`) and defensive JSON parsing
  - Enrich-job queue plumbing — `JobEnrichMatch` type + dedicated `EnrichJobPayload`; `handleMatchJob` gates enqueue at the producer (`score ≥ aiMatchThreshold && has_gemini_key`); `handleEnrichJob` stub registered in `main.go` (Gemini call not yet wired)
- **Approach decided:** Filter-then-enrich — algorithm runs on every job (fast, free); LLM enrichment only on jobs scoring ≥ threshold
- **LLM provider:** Gemini (user-provided API key, not baked in), accessed through a project-owned `Enricher` interface so other providers can be added later without touching the worker
- **Enrichment scope (now):** LLM returns only what we can persist today — a one-line `ai_summary`. The gemini package returns its own result type (`EnrichmentResult{Summary}`); the worker maps it into `UpdateMatchEnrichment` and sets `is_enriched = TRUE` itself (the LLM never knows about persistence state).
- **Enrichment scope (future, not built):** (a) a semantic-validation signal — LLM judges whether the algorithmic match actually holds up — needs a new column before it can be stored; (b) richer match inputs (location, etc.) as more scrapers supply those fields, enabling the matcher/enricher to reason over more of the user profile.
- **Remaining:**
  - Wire `Enrich` into `handleEnrichJob` — decrypt key (`crypto.Decrypt`, handle no-key), build `EnrichmentInput`, call `Enrich`, persist via `UpdateMatchEnrichment`
  - Failure taxonomy in the enrich handler — bad/revoked key → drop (no retry); 429/timeout → return err and let backoff retry
  - Frontend: display match scores, matched/missing skills, and `ai_summary` on job cards
  - Improve algorithmic matcher signal quality (weight skills by proficiency)

### Phase 5.5 — User API Key Management

- **Status:** Done
- **Description:** Let users provide their own Gemini API key from the Settings page. Store encrypted at rest (AES-256, server-side key from env). Never log. Link to aistudio.google.com/apikey. Schema: add `encrypted_gemini_api_key TEXT` column to users.
- **Built:**
  - `encrypted_gemini_api_key TEXT` column on `users` (migration `20260531153511`)
  - `internal/crypto` — AES-256-GCM service; `ENCRYPTION_KEY` (base64, 32 bytes) loaded from env, `log.Fatalf` on missing/invalid key (fail-loud)
  - `PUT /api/users/me/gemini-key` — validates, encrypts, stores ciphertext; never logs the key
  - `SetGeminiKeyByUserID` (store) and `GetGeminiKeyByUserID` (read for worker) queries
  - `has_gemini_key` boolean derived in SQL (`encrypted_gemini_api_key IS NOT NULL`) — exposes "key set" state to the frontend without ever sending the ciphertext
  - `SettingsTab.vue` — password-masked input, "Key Configured / Replace" state, optimistic `has_gemini_key` flip, link to aistudio.google.com/apikey
- **Note:** Decryption is NOT part of 5.5 — it happens where the key is *consumed*, in the Phase 5 LLM enrichment worker (`crypto.Decrypt`, handle the no-key case).

### Phase 6 — Auto Resume Generation

- **Status:** Not Started
- **Description:** When a job scores above threshold: generate tailored resume PDF using LLM + HTML template, send via email with job details and match score attached. This is the career-ops feature baked into Jobradar.

### Phase 7 — Notifications

- **Status:** Not Started
- **Description:** Email alerts via Resend or Mailgun. In-app notification bell. `GET /api/notifications`. Idempotency — never send the same alert twice.

### Phase 5.6 — AI Resume Import

- **Status:** Not Started
- **Description:** User uploads a resume PDF or paste raw text. LLM extracts and categorizes: skills (with proficiency), work experiences, education, projects, certifications. Populates all Phase 2.5 profile tables in a single DB transaction. Triggers re-matching on all existing jobs for that user. UI: upload button on Profile page with a preview/confirm step before committing.

### Phase 8 — More Job Sources

- **Status:** Not Started
- **Description:** Remotive, Adzuna, portal scanner (Greenhouse/Ashby/Lever). Each is a new fetcher only — queue and workers unchanged. Open/closed principle in practice.

### Phase 8.5 — User-Submitted Job Import

- **Status:** Not Started
- **Description:** User pastes a URL to any job posting (LinkedIn, Greenhouse, Ashby, Lever, company page). Backend fetches and scrapes the page, extracts title/company/description/location/skills using LLM. Creates a job record flagged as user-private (visible only to that user). Enqueues match + enrichment automatically. This gives users matching and resume generation for jobs they find outside Jobradar's feed.

### Phase 9 — Discord Bot (Optional)

- **Status:** Not Started
- **Description:** Posts high-match jobs to a channel. Slash commands. Just another input/output to the existing pipeline.

### Phase 10 — Observability and Production Polish

- **Status:** Not Started
- **Description:** Structured logging with slog. `GET /api/metrics`. Token bucket rate limiting (implement the algorithm). Cache-Control headers. Architecture README. Deploy to Railway/Render/DigitalOcean.
