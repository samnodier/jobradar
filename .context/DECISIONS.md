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
