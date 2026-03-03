.PHONY: run build lint fmt test test-int mock tidy

run:
	go run cmd/main.go

build:
	go build -o bin/chat-service ./cmd

lint:
	golangci-lint run

fmt:
	golangci-lint fmt

test:
	go test ./internal/...

test-int:
	go test -tags=integration -count=1 ./tests/integration/...

mock:
	go generate ./...

tidy:
	go mod tidy
