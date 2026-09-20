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
