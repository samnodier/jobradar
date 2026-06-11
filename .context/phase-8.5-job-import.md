# Phase 8.5 — Job Import from Link (Locked Spec)

Status: Planned — building on `feature/job-import`. Plan agreed 2026-06-10.

## Goal

User pastes a job URL → backend fetches + extracts it with an LLM → user reviews a
pre-filled form → confirm creates a **user-private** job → it flows through the
existing match + enrich pipeline. Highest personal-utility feature: lets Sam score
and track jobs he finds *outside* Jobradar's feed.

## Locked Decisions

### Async model
- **Synchronous (user waits on a spinner):** validate URL → SSRF check → fetch HTML
  → LLM extract → return a **draft** (NOT persisted).
- **Asynchronous (reuses existing pipeline):** on confirm, enqueue match; enrich
  follows automatically via `handleMatchJob`'s existing gate.
- Rationale: extraction is one call the user explicitly initiated — a spinner is
  expected UX, and sync gives an honest immediate error (esp. under Gemini quota)
  instead of a silent DLQ death. Only the match+enrich fan-out stays async.

### Privacy — nullable owner column
- New `created_by_user_id` on `jobs` (nullable FK → users, **ON DELETE CASCADE**).
- `NULL` = public scraped feed (what the scraper writes). Set = private import.
- One filter covers anon / public / owner: `WHERE created_by_user_id IS NULL OR
  created_by_user_id = $userID`. For an anon visitor `$userID` is `uuid.Nil`, which
  matches no private row — so anon sees exactly the public feed.

### Dedup — two partial unique indexes
**Two pre-existing global constraints must BOTH be dropped first** (verified via
`\d jobs`): `jobs_external_id_job_source_key` UNIQUE(external_id, job_source) AND
`jobs_source_url_key` UNIQUE(source_url). The second is the easy-to-miss one — a
*global* unique on `source_url` would reject user B importing a URL user A already
imported, overriding the per-user partial index entirely (any violated unique fails
the insert; there is no "which fires first"). Drop both, then replace with:
```sql
CREATE UNIQUE INDEX idx_external_id_job_source
  ON jobs (external_id, job_source)
  WHERE created_by_user_id IS NULL;        -- public/scraped dedup

CREATE UNIQUE INDEX idx_created_by_user_id_source_url
  ON jobs (created_by_user_id, source_url)
  WHERE created_by_user_id IS NOT NULL;    -- per-user import dedup
```
- Why partial: scraped rows (NULL owner) and imported rows (set owner) have
  *different* natural keys. NULLs in a unique index are treated as distinct, so a
  single combined constraint would break scraped dedup. Partial indexes scope each
  rule to its own population.
- Consequence: imported rows are excluded from the public index, so their
  `external_id` need not be unique (can be the URL/a hash/empty).
- `source_url` is the import dedup anchor → must be present (NOT NULL in practice)
  on every owned row, normalized (strip query/fragment, lowercase host) before use.

### Extraction — new `Extractor`, not `Enricher`
- Separate capability in `internal/llm` (different prompt, input = page text,
  output = all possible job fields, null for missing). Open/closed: new capability =
  new interface, not bending `Enrich`.
- Reuse `classifyGeminiError` / `ErrPermanent` for the error taxonomy.
- `ResponseSchema` with title/company required-ish, other fields optional; defensive
  JSON parsing like the enricher.

### Security — SSRF (the load-bearing one)
The server is about to fetch a user-supplied URL. Defenses:
1. **Primary:** resolve the host and reject non-public IPs (loopback `127/8`,
   private `10/8` `172.16/12` `192.168/16`, link-local `169.254/16`). This is the
   real defense — without it `http://localhost:6379` (Redis) walks straight in.
2. Scheme allowlist: `http`/`https` only (no `file://`, etc.).
3. Redirects **off** for v1 (closes the redirect-bypass sub-vector). Note: legit job
   links often redirect — return a clear "paste the final URL" error, don't fail
   silently.
4. Response size cap + request timeout.
SSRF = controlling *where the request goes*; separate from validating *what comes
back* (content check).

### UX — review-before-commit, always
```
import URL → fetch + extract → pre-filled review form → user confirms/edits
          → THEN create row + enqueue match
```
One screen handles all cases: full success (glance + confirm), partial (top up
blanks), total failure (blank form = manual add — no dead end). Catches LLM
hallucination before persistence. Matches the Phase 5.6 preview/confirm pattern.

### Endpoints (behind RequireAuth — an import needs an owner)
- `POST /api/jobs/import` → SSRF-check, fetch, extract → return draft (no DB write).
- `POST /api/jobs/import/confirm` → validate required fields → **create** the job
  (set `created_by_user_id`) → enqueue match for **the single owner only** (NOT the
  scraper's `GetAllUsers` fan-out).
- Note: this is *create*, not *save*. `POST /api/saved_jobs` bookmarks an existing
  job and cannot create one. Bookmarking an import is optional/separate.

### Required fields
- Extraction success: **title + company** minimum.
- Row to exist at all: **source_url** (the dedup anchor).

## Parked (do NOT block 8.5)
- On-demand enrichment vs higher threshold — stopgap is fine; the structural fix
  (user spends quota deliberately) waits until enrichment does more than a 3-sentence
  summary.
- HTML → text: feed Gemini raw HTML vs strip to text first — decide when building
  the extractor (tokens/noise vs simplicity).

## Build Order (dependency-first)
1. **Migration** — add `created_by_user_id`; drop old constraint; add two partial
   indexes. (Sam just learned partial indexes — natural first task.)
2. **SQLC** — add owner filter to `GetJobs`, `GetJobByID`, `SearchJobs`,
   `GetAllJobIDs` (the re-match fan-out — the sneaky one); add a create-imported-job
   query; regenerate.
3. **Extractor** — interface + gemini impl in `internal/llm`.
4. **SSRF-safe fetcher** — in `internal/fetcher` / `internal/httpx`.
5. **Handlers** — import (extract→draft) + confirm (create→enqueue match).
6. **Frontend** — import button, spinner, pre-filled review form, confirm.
7. Tests → merge to `main`.
