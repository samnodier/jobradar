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
- **Per-user multi-provider API keys in `user_api_keys`** — one row per `(user_id, provider)` with a UNIQUE constraint and a DB CHECK allowlist on `provider` (mirrored by `llm.IsKnownProvider` at the handler edge). Replaced the single `users.encrypted_gemini_api_key` column via a data-carrying migration (create table → copy ciphertexts → drop column). Adding a provider = new CHECK value + new `llm` implementation, no schema reshape
- **Fixed server-side provider priority, not a user setting** — when a user has keys for multiple providers, `llm.ProviderPriority` (Groq before Gemini: faster inference, friendlier free tier) picks one. A user-facing "preferred provider" toggle is scope nobody asked for; add it only if a real need appears
- **Provider capability = interfaces implemented** — the Groq client implements both `Extractor` and `Enricher`; `llm.NewExtractor`/`llm.NewEnricher` factories map provider name → implementation, so workers/handlers never mention a vendor
- **Prompt-embedded schema for providers without a schema channel** — Groq's `json_object` mode guarantees valid JSON but not shape (unlike Gemini's `ResponseSchema`), so the expected keys are spelled out in the system prompt. Weaker guarantee, compensated by defensive unmarshalling
- **Imported jobs bypass the match/enrich gates** — the title gate (Jaro-Winkler ≥ 0.55) and the enrich threshold exist to filter the scraped firehose; a user-imported job is explicit intent, so `handleMatchJob` passes `titleThreshold 0` (never skipped, title still contributes its honest score) and enqueues enrichment regardless of `aiMatchThreshold`. The matcher stays pure — the caller owns the policy via the threshold parameter
- **Missing title signal renormalizes, like every other dimension** — with no desired roles, the title dimension drops out of the accumulator (unknown ≠ 0) instead of skipping every job; a profile with zero signal (no roles, skills, or experiences) is skipped, never given a fabricated score
- **Skips are deleted, not written** — `user_job_matches` holds only real matches. On a skip, `DeleteMatch` clears any stale row (a job that dropped below threshold after a profile re-match must not keep its old high score). Prevents the table growing to jobs × users
- **Provider fallback at the call site** — `selectProviderKeys` returns all configured providers decrypted, in `ProviderPriority` order; the extract handler and enrich worker walk the list, falling back to the next provider on any failure. A decrypt failure still fails loud (it's our bug, not the provider's). A user-facing "preferred provider" setting remains future scope

## Personal release decisions — September 2026

- **AI optional:** automatic enrichment defaults off; deterministic matching and email operate without provider keys. Dedicated enrichment worker prevents slow LLM calls from occupying both core workers.
- **PostgreSQL pending-match outbox:** triggers persist pending work in the business transaction. Redis delivery remains simple and replaceable; the dispatcher recovers unacknowledged work after five minutes. Revision checks preserve work created during processing. This is at-least-once processing, not exactly-once delivery.
- **Required-skill coverage:** compare explicit job skills with the profile map; description mentions are display evidence only. Missing skills use the full owned skill set.
- **Source boundaries:** source adapters produce normalized jobs; shared ingestion owns persistence. Adzuna runs twice daily and caps query/page fan-out. Unknown remote status stays unknown.
- **Personal digest only:** one configured recipient matching an opted-in account; no arbitrary-recipient send route. Keep a persisted immutable message and user/job ledger. Resend retries use stable keys for less than 24 hours; ambiguous GoHighLevel delivery stops for review. Provider acceptance is not confirmed inbox delivery.
- **No local model dependency:** small-machine operation and predictable failure recovery take precedence over local inference.
- **Deployment:** preserve current dev pgdata; production Compose uses separate named volumes. Serve built frontend from Go behind an HTTPS reverse proxy. Migrations are an explicit deployment step.
- **Release scope:** finish job collection → useful matching → daily email before resume automation. Knowledge questions and code links are written into LEARNINGS.md for a later guided review.

## Profile-driven personal workflow

- User preferences are the source of truth for roles, locations, score threshold and delivery times. Environment variables retain provider credentials and a global email safety switch.
- Eligibility is a separate deterministic gate with explanations; missing source metadata is user-controlled. No inference of worldwide work authorization from remote=true.
- Notifications are once per user/job, created with the score transaction and filtered against current settings. No external push service in v1.
- Manual AI tailoring uses authenticated Markdown export. No model charges, no autonomous application submission, no promise that chat subscriptions include arbitrary API automation.
- Structured JobPosting data precedes optional model extraction; plain page text is the no-key/outage fallback. All imports require review.
- Two local-time digest slots share the cross-day job delivery ledger; downtime catches up only the latest slot.
- Small focused commits; verify dev locally with Sam before main promotion and deployment.

## Preference scoring and external workflows

- Professional fit (existing role/skill/experience score) has weight 80. Each configured known preference has weight 5 and contributes full or zero credit. Unknown evidence has no numeric weight. Hard eligibility remains separate. Literal industry/category phrases are matched against explicit source/reviewed metadata, without inferring employer industry.
- Read-only workflow tokens authenticate only export GET routes; credential management requires browser sessions. Expiry/revocation/personal-email restrictions are checked on every request.
- Export thresholds override the profile default for a request without changing notification settings. Date semantics are explicit; pagination is keyset-based, not a frozen snapshot.
- Email/password is the default UI login; GitHub is optional. Existing accounts keep the same user ID. No public email registration; a server-owner command provisions or recovers local accounts.
- Password hashes are Argon2id and separate from profile data. Atomic password changes increment a session version and revoke workflow/setup tokens. GitHub linking is explicit and requires a recent session plus the same verified primary email.
- Authentication rate limits use direct peer addresses, not arbitrary forwarded headers. A reverse proxy shares its IP quota unless trusted-proxy handling is designed later.

## September 22 usability review

- Preserve commas within a location tag; Enter/Add location commits it. Locations are literal alternatives, not geocoded regions.
- Share square card/input/button styles and existing accent. Compact settings cards separate match criteria, source market, channels and schedule while saving the matching resource together.
- Optimistic saves retain failed drafts, show persistent retry state and do not replace newer edits with older responses.
- Native drag-and-drop has a status-select alternative. Custom ordering within columns is out of scope.
- Public feeds Jobicy/MyJobMag require explicit opt-in, use conservative polling, preserve attribution and unknown metadata.
- Profile download has an explicit field allowlist and declared limited scope; it is not a database backup.

## Feed threshold clarification — September 22

Sam clarified that Best matches is score-only: include stored scores at or above the saved minimum, regardless of preference eligibility or pending rematching. Source/search constraints still apply. Notifications/digests keep their stricter gates.
