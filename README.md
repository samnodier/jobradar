# Jobradar

A personal job-search application built with Go, PostgreSQL, Redis, Vue 3 and TypeScript. It collects jobs, ranks them against a profile, tracks applications, and sends one or two personal email digests each day. Core matching and notifications work without an AI API or a local model.

## Working features

- GitHub OAuth with Redis sessions and an optional owner-email restriction.
- RemoteOK, Adzuna searches, and a Greenhouse company-board watchlist.
- Saved jobs, application tracking, resume/profile sections, and matching preferences.
- Algorithmic title/skill/experience scoring; optional Groq/Gemini summaries.
- URL import through structured metadata or optional AI, with visible-text/manual fallback; imports are private.
- Copy for AI exports the job and saved resume as reviewable Markdown, without calling a model.
- Profile-controlled thresholds, eligibility filters, best-match views and deduplicated in-app notifications.
- Durable matching recovery from database records, separate AI worker, bounded requests and retry cooldowns.
- Morning/evening email to one configured recipient using Resend or GoHighLevel, persisted delivery history and preview.
- Real service status, paginated job feed, built frontend served by Go.

Scores are heuristics, not hiring probabilities or proof of work authorization. Sources may omit required skills and remote eligibility. Review the original posting before applying.

## Local development

Requires Go 1.25.8+, Node 24, Docker Compose, Goose and SQLC.

1. Copy `.env.example` to `.env`; fill database and GitHub OAuth credentials. Generate `ENCRYPTION_KEY` with `openssl rand -base64 32`. Keep the key stable: changing it prevents decryption of stored AI credentials.
2. `make up`, then `make migrate-up`. Existing `pgdata` is preserved; do not reset it.
3. `make run` starts the API/workers on port 8080.
4. In `frontend`, run `npm ci` and `npm run dev`. The development proxy handles `/api` and OAuth routes.
5. Sign in, complete onboarding, and set desired roles, skills and experience. Jobs without enough profile information cannot receive meaningful scores.

The default AI enrichment flag is off. URL extraction can still use a key configured in Settings; manual import works without one.

## Job sources

RemoteOK is enabled by default and runs every six hours. Set `REMOTEOK_ENABLED=false` to disable it. `SCRAPE_INTERVAL` must be at least one hour.

For Adzuna, obtain your own application ID/key from [Adzuna](https://developer.adzuna.com/overview) and configure `ADZUNA_APP_ID`, `ADZUNA_APP_KEY`, and `ADZUNA_PAGES` (1–3). The personal account supplies up to five desired roles and its first desired location from Profile → Preferences; select the supported search market in Profile → Settings. Runs every 12 hours with at most five searches and a three-second gap between page requests. Every page consumes one API request; retries/restarts consume more. Review [limits and attribution requirements](https://developer.adzuna.com/docs/terms_of_service). Listings retain Adzuna redirect URLs and branding.

For Greenhouse, set `GREENHOUSE_BOARDS` to actual board slugs separated by commas. The public [Job Board API](https://docs.greenhouse.io/job-board.html) is company-specific. This adapter does not discover all companies automatically.

## Personal email

Configure one provider; keep `DIGEST_ENABLED=false` until ready. The recipient must match a registered Jobradar user whose job notifications are enabled.

- Server configuration: `DIGEST_RECIPIENT`, `EMAIL_FROM`, provider credentials and the `DIGEST_ENABLED` safety switch.
- Profile → Settings: minimum score (default 60), timezone, one or two hours (default 8 and 18), maximum jobs per email (default 10), email/in-app opt-outs. The master Job Recommendations switch also applies. Old `DIGEST_HOUR`, `DIGEST_TIMEZONE`, `DIGEST_MIN_SCORE`, and `DIGEST_MAX_JOBS` environment values are no longer read.
- Resend: `EMAIL_PROVIDER=resend`, `RESEND_API_KEY`. Use an allowed/verified sender and recipient according to your account configuration.
- GoHighLevel: `EMAIL_PROVIDER=gohighlevel`, `GHL_API_KEY`, `GHL_CONTACT_ID`. Configure email delivery in your sub-account; token needs contact-read and message-write access. The contact email is checked against the configured recipient before sending. Current adapter uses API Version `v3`.

`make digest-preview` prints HTML without sending or reserving jobs. Dashboard → Preview opens the same preview in a browser. Once configured, enable delivery and run `make digest-once` for an immediate cycle. This still respects opt-out, matching backlog, and one digest per configured local-date/hour slot. If no new qualifying matches exist, it sends nothing. Background scheduling checks once per minute, catches up after startup, and prepares only the latest due slot. The morning jobs are never repeated in the evening.

`sent` means the provider accepted the email. It does not guarantee inbox delivery. Resend retries preserve the exact payload and idempotency key, stopping before the provider's 24-hour deduplication window expires. GoHighLevel ambiguous sends become `uncertain` and require checking provider history. See [deployment and recovery](docs/DEPLOYMENT.md).

## Tests

- `make test`: backend unit tests; integration tests skip unless their explicit test environment is set.
- `cd frontend && npm run build && npm run test:unit -- --run`.
- For integration tests, use isolated PostgreSQL/Redis: `docker compose -p jobradar-test -f compose.test.yml up -d`.
- Apply migrations to `postgres://jobradar_test:jobradar_test@127.0.0.1:55432/jobradar_test?sslmode=disable` with Goose, then `make integration-test`.
- Integration databases must end in `_test`; Redis queue tests flush DB 1 on `TEST_REDIS_ADDR`, never the default development Redis implicitly.

## Deploy

See [docs/DEPLOYMENT.md](docs/DEPLOYMENT.md) and [the personal workflow guide](docs/PERSONAL_WORKFLOW.md). Production Compose creates separate named volumes and binds the app to localhost for an HTTPS reverse proxy. It does not migrate or mount the development database automatically.

## Architecture and learning

```mermaid
flowchart LR
  APIs[RemoteOK / Adzuna / Greenhouse] --> Ingest[Normalize and upsert]
  Ingest --> PG[(PostgreSQL)]
  PG --> Pending[Durable pending matches]
  Pending --> Redis[(Redis delivery)]
  Redis --> Match[Algorithmic matcher]
  Match --> PG
  Match -. optional .-> AI[Separate AI queue]
  AI --> PG
  PG --> Digest[Persisted daily digest]
  Digest --> Email[Email provider]
  Vue[Vue application] <--> Go[Go HTTP API]
  Go <--> PG
```

Project state: [.context/PLAN.md](.context/PLAN.md), [.context/SCHEMA.md](.context/SCHEMA.md), [.context/ROUTES.md](.context/ROUTES.md), [.context/DECISIONS.md](.context/DECISIONS.md).

[LEARNINGS.md](LEARNINGS.md) contains implementation explanations, interview questions, failure scenarios, and a Go/Laravel/TypeScript comparison for a later walkthrough.

Future work includes in-app tailored resume generation (manual Copy for AI is available), more ATS adapters, geographic eligibility beyond text matching, and cross-source deduplication. These do not block a personal daily digest.

## Email and GitHub sign-in

Email/password and GitHub share the same account, profile and saved jobs. After migrations and starting the app:

- Existing GitHub account: sign in with GitHub, then open **Profile → Settings → Sign-in methods** within ten minutes. Add a password, log out, and sign in with your account email and that password.
- Local-only installation or forgotten password: on your own server run `go run ./cmd/jobradar --password-setup you@example.com`. Open the private link printed in your terminal within 15 minutes and choose a password. This command preserves an existing account or provisions a new local one. Never share the link.
- GitHub is optional; leave its client ID and secret empty to disable it. A local account can explicitly connect GitHub in Settings after recent login; its verified primary email must match.
- Password changes sign out other sessions and revoke workflow tokens. Public email registration and email-based password recovery are not enabled. The operator command supplies local recovery without an email provider.
- `PERSONAL_EMAIL`, when set, restricts access regardless of login method. It is separate from `DIGEST_RECIPIENT`, which controls the personal email digest.

Password storage follows [OWASP Argon2id guidance](https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html). See LEARNINGS.md for the implementation and review questions.
