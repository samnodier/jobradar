##
# jobradar
#
# @file
# @version 0.1

include .env

# run 'make up' instead of the full docker command
up:
	docker compose up -d
down:
	docker compose down

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

.PHONY: up down migrate-up migrate-down migrate-create sqlc
# end
