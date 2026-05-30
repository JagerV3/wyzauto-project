DATABASE_URL ?= postgres://wyzauto:wyzauto@localhost:5432/wyzauto?sslmode=disable

.PHONY: setup fmt vet lint test test-unit test-integration run docker-up docker-down hooks

setup:
	go mod download

fmt:
	gofmt -w $$(find . -name '*.go' -not -path './vendor/*')

vet:
	go vet ./...

lint:
	golangci-lint run ./...

test-unit:
	go test ./...

test-integration: docker-up
	RUN_INTEGRATION=1 DATABASE_URL="$(DATABASE_URL)" go test ./internal/service -run Integration -count=1

test: docker-up fmt vet test-unit test-integration

run:
	DATABASE_URL="$(DATABASE_URL)" go run ./cmd/server

docker-up:
	docker compose up -d postgres
	docker compose exec postgres pg_isready -U wyzauto -d wyzauto

docker-down:
	docker compose down -v

hooks:
	git config core.hooksPath .githooks
