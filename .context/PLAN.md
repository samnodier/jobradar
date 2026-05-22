## Current Status

Phase: 2.5 — Resume Profile Vertical Slice
Status: In Progress

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

- **Status:** In Progress
- **Description:** Build out a detailed resume-like profile with Work History, Achievements, and Skills. This serves as the foundation for AI matching and automated resume generation.

### Phase 3 — Frontend MVP

- **Status:** Done
- **Description:** Build out JobsView, ProfileView with real data. Save/unsave buttons on job cards. Applications tracker (Kanban board). Delete account button on profile. Navigation guard working correctly.

### Phase 4 — Job Queue and Worker Engine

- **Status:** Not Started
- **Description:** Redis list as queue (LPUSH/BRPOP). Worker goroutines with graceful shutdown via context cancellation. Exponential backoff (1s→2s→4s→8s, max 3 retries). Dead letter queue. Scheduler with time.Ticker.

### Phase 5 — AI Matching

- **Status:** Not Started
- **Description:** User skills profile. AI scoring worker — job description + skills → JSON with match_score, matched_skills, missing_skills, summary. Prompt engineering for reliable structured output. Store scores in jobs table.

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
