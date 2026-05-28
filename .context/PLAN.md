## Current Status

Phase: Phase 5 — AI Matching
Status: In Progress — algorithmic matcher built, LLM integration pending

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
  - Profile compiler in `worker_match.go`: fetches desired roles, skills, experiences from DB and passes to matcher
  - `UpdateJobMatchingResult` writes `match_score`, `matched_skills`, `missing_skills`, `ai_summary` back to jobs table
  - Worker wired to `queue.JobMatchJob` job type
- **Approach decided:** Filter-then-enrich — algorithm runs on every job (fast, free); LLM enrichment only on jobs scoring ≥ threshold
- **LLM provider:** Gemini (user-provided API key, not baked in)
- **Remaining:**
  - Fix bugs in `main.go` worker registration (duplicate handler, missing `matchJobWorker` method)
  - Improve algorithmic matcher signal quality
  - Add `encrypted_gemini_api_key` to user schema for user-provided API keys
  - LLM enrichment worker for high-scoring jobs (summary + semantic validation)
  - `GET/PUT /api/users/me/profile` for AI matching preferences
  - Frontend: display match scores, matched/missing skills on job cards

### Phase 5.5 — User API Key Management

- **Status:** Not Started
- **Description:** Let users provide their own Gemini API key from the Settings page. Store encrypted at rest (AES-256, server-side key from env). Never log. Link to aistudio.google.com/apikey. Schema: add `encrypted_gemini_api_key TEXT` column to users.

### Phase 6 — Auto Resume Generation

- **Status:** Not Started
- **Description:** When a job scores above threshold: generate tailored resume PDF using LLM + HTML template, send via email with job details and match score attached. This is the career-ops feature baked into Jobradar.

### Phase 7 — Notifications

- **Status:** Not Started
- **Description:** Email alerts via Resend or Mailgun. In-app notification bell. `GET /api/notifications`. Idempotency — never send the same alert twice.

### Phase 8 — More Job Sources

- **Status:** Not Started
- **Description:** Remotive, Adzuna, portal scanner (Greenhouse/Ashby/Lever). Each is a new fetcher only — queue and workers unchanged. Open/closed principle in practice.

### Phase 9 — Discord Bot (Optional)

- **Status:** Not Started
- **Description:** Posts high-match jobs to a channel. Slash commands. Just another input/output to the existing pipeline.

### Phase 10 — Observability and Production Polish

- **Status:** Not Started
- **Description:** Structured logging with slog. `GET /api/metrics`. Token bucket rate limiting (implement the algorithm). Cache-Control headers. Architecture README. Deploy to Railway/Render/DigitalOcean.
