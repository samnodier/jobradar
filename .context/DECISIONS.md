## Design Decisions Already Made

- **No ORM** — raw SQL with SQLC for type safety
- **Sessions over JWT** — Redis-backed sessions with HttpOnly cookies
- **Chi router** — lightweight, idiomatic Go
- **Pending signup pattern** — GitHub OAuth creates a Redis entry with TTL, completed on onboarding, written to Postgres atomically
- **SQLC** — generates type-safe Go code from SQL queries
- **Separate `saved_jobs` table** — saved jobs are distinct from applications. A user saves a job first, then optionally creates an application. These are separate concerns with separate tables
- **Username generated from email** — not from GitHub login. Supports future Google OAuth
- **`apiConfig` holds `db`, `rdb`, and `pool`** — refactored to avoid two separate handler configs
- **`Session` struct belongs in a shared package** — currently duplicated between `internal/auth` and `cmd/jobradar`, needs to move to `internal/session` or similar
- **Filter-then-enrich score ownership** — the algorithmic matcher owns `match_score` (cheap, deterministic, runs on every job); the LLM owns only `ai_summary`. The score is the gate for the expensive LLM step, so it can't depend on the LLM — otherwise you'd run the LLM on everything just to decide whether to run the LLM
- **Provider portability via a project-owned `Enricher` interface** — the rest of the app depends on the `Enricher` abstraction, never on a vendor SDK. The Gemini SDK lives behind one implementation; a new provider = a new implementation, not edits to the worker (dependency inversion / open-closed)
- **Enrichment gated at the producer** — `handleMatchJob` enqueues an enrich job only when `score ≥ aiMatchThreshold && has_gemini_key`; don't queue work guaranteed to no-op. The pre-check can go stale, so the enrich handler must still re-check the key defensively before calling the API
- **Narrow enrichment write** — `UpdateMatchEnrichment` is physically unable to touch the score columns, so the LLM stage cannot clobber the authoritative algorithmic scores. No COALESCE: a system write always has its value, so write it directly (a missing value should surface as a bug, not silently preserve stale data)
