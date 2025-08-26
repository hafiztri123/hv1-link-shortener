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
	@go run ${MAIN_FILE}

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

.DEFAULT_GOAL := help
