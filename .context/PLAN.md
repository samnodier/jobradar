## Current Status

Phase: Phase 8.5 — User-Submitted Job Import (plus multi-provider key restructure)
Status: Import happy path verified live 2026-07-07 (Groq extract → confirm → job created). Same day: fixed why imported jobs never matched/enriched (title gate + enrich threshold now bypassed for imports; enrich worker could not see private jobs due to a uuid.Nil lookup), added provider fallback, delete-key UI, and resolved the matcher/worker TODOs. Degradation paths (no key → 422, bad key → 422/manual entry) still need a manual pass.

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

- **Status:** Done — restructured 2026-07-06 into multi-provider `user_api_keys` (one row per user+provider; gemini and groq). `PUT/DELETE /api/users/me/api-keys/{provider}` replaced the gemini-key route; `has_gemini_key` became `configured_providers: string[]` on the user payload; SettingsTab renders a provider list. Provider selection at runtime: `llm.ProviderPriority` (groq first), decrypt in `selectProviderKey` (cmd/jobradar/apikeys.go).
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
- **Cadence (captured 2026-06-10):** daily morning digest of *new* matches scoring ≥ threshold, batched per user (LinkedIn-job-alert style), not one email per job.
- **Default policy:** notify on matches scoring ≥ 50% by default; make the threshold (and on/off, and cadence) user-customizable later — start with the sane default, add settings when needed.
- **Idempotency mechanism (this is the "more architecture" Sam worried about — it's bounded, not new scope):** a sent-notification ledger keyed by `(user_id, job_id, notification_type)`; check-before-send. One small table + a uniqueness guard. This is exactly the Phase-7 idempotency requirement, already planned — not extra architecture.
- **North-star reference:** career-ops.org — the product Jobradar is modelled on; revisit for feature parity ideas (esp. resume tailoring + alerts).

### Phase 5.6 — AI Resume Import

- **Status:** Not Started
- **Description:** User uploads a resume PDF or paste raw text. LLM extracts and categorizes: skills (with proficiency), work experiences, education, projects, certifications. Populates all Phase 2.5 profile tables in a single DB transaction. Triggers re-matching on all existing jobs for that user. UI: upload button on Profile page with a preview/confirm step before committing.

### Phase 8 — More Job Sources

- **Status:** Not Started
- **Description:** Remotive, Adzuna, portal scanner (Greenhouse/Ashby/Lever). Each is a new fetcher only — queue and workers unchanged. Open/closed principle in practice.

### Phase 8.5 — User-Submitted Job Import

- **Status:** Built — pending end-to-end verification
- **Description:** User pastes a URL to any job posting (LinkedIn, Greenhouse, Ashby, Lever, company page). Backend fetches and scrapes the page, extracts title/company/description/location/skills using LLM. Creates a job record flagged as user-private (visible only to that user). Enqueues match + enrichment automatically. This gives users matching and resume generation for jobs they find outside Jobradar's feed.
- **Built:** SSRF-safe fetcher (dial-time IP guard), HTML-to-text, `Extractor` interface with Gemini + Groq implementations (Ollama experiment removed 2026-07-06), extract/confirm endpoints, private-job ownership + dedup upsert, import panel with chips/inline errors/manual-entry fallback, match enqueue on confirm. Verified live 2026-07-07 with a Groq key. Imported jobs bypass the title gate and enrich threshold (user intent); enrich worker sees private jobs (uuid.Nil bug fixed).
- **Remaining:** manual pass on degradation paths (no key → 422 — now testable from the UI via the Remove button; bad key → 422 + manual-entry fallback), and a second import to confirm the match+enrich chain end-to-end.
- **Future (deliberately not built yet):** user-selectable preferred provider (a `preferred_provider` user setting overriding `llm.ProviderPriority` — fixed priority + automatic fallback covers today's need); UI nudge "set a desired role to get match scores" when the profile has no desired roles (belongs with the match-display UI).

### Phase 9 — Discord Bot (Optional)

- **Status:** Not Started
- **Description:** Posts high-match jobs to a channel. Slash commands. Just another input/output to the existing pipeline.

### Phase 10 — Observability and Production Polish

- **Status:** Not Started
- **Description:** Structured logging with slog. `GET /api/metrics`. Token bucket rate limiting (implement the algorithm). Cache-Control headers. Architecture README. Deploy to Railway/Render/DigitalOcean.
