dev:
	air

run:
	go run ./cmd

test:
	go test -v -cover ./...

migrations:
	migrate create -ext sql -dir schema/migrations -seq $(NAME)

migrate:
	migrate -path schema/migrations -database ${DATABASE_URL} -verbose up

db_tidy:
	migrate -path schema/migrations -database ${DATABASE_URL} force $(REVISIONS)

db_rollback:
	migrate -path schema/migrations -database ${DATABASE_URL} -verbose down $(REVISIONS)

sqlc:
	sqlc generate

.PHONY: dev run test migrations migrate db_tidy db_rollback sqlc
