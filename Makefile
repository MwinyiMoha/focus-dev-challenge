dev:
	air

run:
	go run ./cmd

test:
	go test -v -cover ./...

.PHONY: dev run test