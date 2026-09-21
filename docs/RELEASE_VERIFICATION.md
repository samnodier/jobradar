# Personal release verification — 2026-09-20

Implementation branch: `feature/personal-release`. All database tests used a separate PostgreSQL 16 container/database ending in `_test`; Redis tests used a separate container and test databases. Existing development data and real provider accounts were not used.

## Passed

- All migrations applied to an empty database. The new personal-release migration was rolled down and up successfully.
- `TEST_DATABASE_URL=... TEST_REDIS_ADDR=... go test -race ./internal/... ./cmd/...`: unit and integration tests passed. Coverage includes private-job authorization, filtered feeds, revision acknowledgement, lost-delivery recovery, actual deterministic matching, email opt-out/deduplication, exact-payload retries and ambiguous-send handling.
- `go vet ./internal/... ./cmd/...`: passed.
- Frontend production build/type checking and Vitest application-shell test: passed.
- Compatible frontend dependency fixes applied; npm audit reported zero vulnerabilities at verification time.
- Production application image and Goose migration image built. Migration image successfully read the applied migration history.
- Running production-image HTTP checks passed: health, empty JSON array, invalid pagination rejection, unknown API 404, protected endpoint 401, Vue deep-link fallback, PostgreSQL/Redis status, disabled digest/history, manual import, authenticated private visibility, anonymous private-job rejection, saving, search and pagination.
- `git diff --check`: passed.

The HTTP check used a synthetic user and Redis session in the isolated environment. It validates authenticated application behavior but does not validate GitHub's external OAuth exchange. No provider keys or actual recipient were needed. The synthetic job never reached an email provider.

## Still requires account/deployment verification

- Actual GitHub OAuth on the deployment's HTTPS domain.
- Adzuna credentials, country and search terms; account quotas and results.
- Real RemoteOK/Greenhouse job availability and desired watchlist.
- Verified email sender and the chosen recipient, provider, timezone and hour.
- Digest preview followed by actual inbox delivery and provider-history inspection.
- Several days of scheduled fetching/delivery, backup and restore rehearsal on the deployment host.

The application builds and runs without an AI provider or local language model. Tests do not establish that any remote provider is available, that credentials are valid, or that an email arrives in the inbox. Automatic AI enrichment remains optional and disabled by default.

Nonblocking frontend build warnings: the vendor bundle exceeds the default 500 kB warning threshold, and a dependency contains misplaced pure annotations. Neither prevented type checking or production asset generation; route-level code splitting can be a later improvement.


## Profile workflows verification

The follow-up implementation on `feature/profile-workflows` was verified with:

- Backend unit/integration suite with the race detector, including preference gates, notification deduplication/ownership, safe Markdown export, invalid settings and ownership enforcement, morning/evening slots and cross-slot job deduplication. No real provider sends.
- Go vet and gopls diagnostics on the originally flagged files and new handlers: no reported diagnostics. Unsaved editor buffers are outside this check.
- Vue type checks, production build and Vitest shell test: passed. A production-only inline event-expression failure was fixed in its own commit.
- Full production Docker image build: passed.
- Running-container HTTP workflow: save profile settings and roles/location; import Nairobi and London jobs; background worker scores both but recommends only the desired location; exactly one eligible notification, mark read; Markdown includes profile/job facts and excludes primary contact email; anonymous export/private-job access rejected; preview includes only the eligible job; digest status uses profile hours/threshold; in-app opt-out hides notifications; deep-link SPA route serves correctly.
- All migrations on a fresh isolated database, followed by rollback/reapply of migrations 20260920000200 through 20260920000500: passed. The digest-slot rollback intentionally refuses if data would violate the old one-digest-per-day constraint.

These checks used synthetic sessions and fake providers. Sam's manual browser review, actual OAuth, live Adzuna results, sender setup and inbox delivery remain necessary before production promotion. No changes were applied to the existing development database.

Editor follow-up: `go list ./...` reproduced a pgdata permission/traversal error; package-targeted checks remain green. The repository now includes a syntax-checked LazyVim `.lazy.lua` gopls exclusion. The user's running editor must reload it; existing PostgreSQL ownership was preserved.

## Preference scoring and workflow export follow-up

- Preference scorer unit tests and worker/database integration verify industry changes enqueue work, lower stored score and reject eligibility.
- Source/import fixtures verify Adzuna category/contract-time and structured industry/employment extraction.
- Token integration verifies account identity, expiry, revocation, cross-owner rejection and read-only/personal-instance restrictions.
- Export integration verifies minimum score, private ownership, eligibility override, pending-score exclusion, deterministic pagination and request validation; DST unit test checks a 23-hour local day.
- Full backend suite passed with race detector against isolated PostgreSQL/Redis; go vet passed. Frontend production build/typecheck and existing shell test passed. New migration down/down/up passed on the isolated database.
- No live external AI/provider call, real email or browser manual acceptance test performed for this follow-up. Existing development database was not migrated.
