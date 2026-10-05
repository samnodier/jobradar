# Moving to a new machine

Use this to set up Jobradar on another development machine or an always-on host, carrying over your existing data. It works on macOS and Linux.

Secrets and data never go through git. Copy them privately (USB drive, or `scp` on your own network): your `.env` file and a database dump.

## 1. On the old machine: export

```sh
make up
docker compose exec -T db sh -c 'pg_dump -U "$POSTGRES_USER" -d "$POSTGRES_DB" -Fc --no-owner --no-acl' > jobradar.dump
```

Copy `jobradar.dump` and `.env` to the new machine. If the old machine runs the digest scheduler, stop it once the new one takes over: only one machine should send email.

## 2. On the new machine: tools

- Git, Make, Go (version in `go.mod`), Node (current LTS), and Docker with the Compose plugin
- Goose: `go install github.com/pressly/goose/v3/cmd/goose@latest`, with `$(go env GOPATH)/bin` on your PATH
- Linux only: a C compiler (for example `build-essential`) for Go's race detector

## 3. Code, configuration and data

```sh
git clone git@github.com:samnodier/jobradar.git
cd jobradar
git checkout dev
cp /path/to/.env .env && chmod 600 .env
go mod download
(cd frontend && npm ci)
make up
until docker compose exec -T db pg_isready -q; do sleep 1; done
docker compose exec -T db sh -c 'pg_restore -U "$POSTGRES_USER" -d "$POSTGRES_DB" --no-owner --no-acl' < /path/to/jobradar.dump
make migrate-up
```

- Keep `ENCRYPTION_KEY` unchanged; stored provider keys cannot be decrypted with a different key.
- Restore into an empty database only; restoring twice produces duplicate-key errors.
- Compare a few row counts (`jobs`, `users`, `user_job_matches`) with the old machine.
- On a machine that should not send email, set `DIGEST_ENABLED=false`.

## 4. Run every test

```sh
docker compose -p jobradar-test -f compose.test.yml up -d
until docker compose -p jobradar-test -f compose.test.yml exec -T db pg_isready -q; do sleep 1; done
goose -dir sql/schema postgres 'postgres://jobradar_test:jobradar_test@127.0.0.1:55432/jobradar_test?sslmode=disable' up
make integration-test
go vet ./cmd/... ./internal/...
(cd frontend && npx vitest run && npm run build)
docker compose -p jobradar-test -f compose.test.yml down
```

The test stack uses separate ports and in-memory storage, so it never touches development data. Use `./cmd/... ./internal/...` instead of `./...`, because the development `pgdata/` directory inside the repository is unreadable to Go tooling.

Known lint debt (no behavior impact): `npx oxlint .` reports `require-mock-type-parameters` findings in test files; `npx eslint .` reports findings in Vite's stock `src/env.d.ts` shim.

## 5. Run

```sh
make run                     # API, source collection, matching and (if enabled) the digest scheduler
cd frontend && npm run dev   # http://localhost:5173
```

The main process runs the digest scheduler when `DIGEST_ENABLED=true`; a separate `--digest-worker` process is optional. For an always-on host, see `docs/DEPLOYMENT.md`. A laptop used as a server must be configured not to suspend when its lid closes.
