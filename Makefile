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

build:
	CGO_ENABLED=0 GOOS=linux go build -o ./build/app -ldflags="-s -w" ./cmd

compress_binary:
	upx --best --lzma ./build/app

test_binary:
	upx -t ./build/app

.PHONY: dev run test migrations migrate db_tidy db_rollback sqlc build compress_binary test_binary
