SHELL := /bin/bash

ifneq (,$(wildcard .env))
	include .env
	export
endif

BINARY_NAME=app-link-shortener
MAIN_FILE=cmd/server/main.go


build:
	@echo "Building binary..."
	@go build -o ./bin/${BINARY_NAME} ${MAIN_FILE}

run:
	@air

format:
	@go fmt ./...
	@go vet ./...

migrate-up:
	@migrate -database "$(DATABASE_URL)" -path migrations up

migrate-version:
	@migrate -database "$(DATABASE_URL)" -path migrations version

migrate-down:
	@migrate -database "$(DATABASE_URL)" -path migrations down

migrate-force:
	@migrate -database "$(DATABASE_URL)" -path migrations force $(v)

migrate-create:
	@migrate create -ext sql -dir ./migrations -seq $(v)


test-setup:
	@migrate -database "$(DATABASE_URL_TEST)" -path migrations down -all
	@migrate -database "$(DATABASE_URL_TEST)" -path migrations up

test-coverage:
	@go test -coverprofile=coverage.out ./internal/...
	@go tool cover -func=coverage.out

test-integration:
	@go test -v --tags=integration ./...

test-html:
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out

branch-review:
	@git branch --merged main | grep -v -E '^\*|main|master$$'

branch-prune:
	@git branch --merged main | grep -v -E '^\*|main|master$$' | xargs git branch -d



.DEFAULT_GOAL := help
