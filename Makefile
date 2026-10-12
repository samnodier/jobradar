##
# jobradar
#
# @file
# @version 0.1

-include .env

# run 'make up' instead of the full docker command
up:
	docker compose up -d
down:
	docker compose down

build:
	go build ./cmd/jobradar/...
run:
	go run ./cmd/jobradar/...

# run 'make migrate-up'
migrate-up:
	goose -dir sql/schema postgres "${DB_URL}" up
migrate-down:
	goose -dir sql/schema postgres "${DB_URL}" down
migrate-reset:
	goose -dir sql/schema postgres "${DB_URL}" reset
migrate-create:
	goose -dir sql/schema create $(name) sql

sqlc:
	sqlc generate

exec:
	docker compose exec db psql -U ${DB_USER} -d ${DB_NAME}

test:
	go test ./internal/... ./cmd/...

.PHONY: up down migrate-up migrate-down migrate-create sqlc exec test
# end

frontend-build:
	cd frontend && npm ci && npm run build

digest-preview:
	go run ./cmd/jobradar --digest-preview

digest-once:
	go run ./cmd/jobradar --digest-once

integration-test:
	TEST_DATABASE_URL='postgres://jobradar_test:jobradar_test@127.0.0.1:55432/jobradar_test?sslmode=disable' TEST_REDIS_ADDR=127.0.0.1:56379 go test -race -p 1 ./internal/... ./cmd/...

.PHONY: build run frontend-build digest-preview digest-once integration-test migrate-reset
