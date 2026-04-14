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

# run 'make migrate-up'
migrate-up:
	goose -dir sql/schema postgres "${DB_URL}" up
migrate-down:
	goose -dir sql/schema postgres "${DB_URL}" down

sqlc:
	sqlc generate
# end
