SHELL := /bin/bash

ifneq (,$(wildcard .env))
	include .env
	export
endif

BINARY_NAME=hv1-link-shortener
TRANSACTION_DATABASE_URL := postgres://$(DB_USER):$(DB_PASSWORD)@$(APP_URL):$(DB_PORT)/$(TRANSACTION_DB)?sslmode=$(DB_SSL)

WORKER_PATH=./services/worker/server/cmd
WORKER_BIN=./bin/worker

APP_PATH=./services/app/server/cmd
APP_BIN=./bin/app

SHARED_DIRECTORY=./shared
APP_DIRECTORY=./services/app
WORKER_DIRECTORY=./services/worker



build-all:
	@echo "Building app services..."
	@go build -o $(APP_BIN) $(APP_PATH)
	@echo "Building worker services..."
	@go build -o $(WORKER_BIN) $(WORKER_PATH)


run-all:
	build-all
	@echo "Running app services and worker services"
	@$(APP_BIN) & $(WORKER_BIN) & wait

tidy-all:
	@echo "Tidying all modules..."
	@cd $(APP_DIRECTORY) && go mod tidy
	@cd $(WORKER_DIRECTORY) && go mod tidy
	@cd $(SHARED_DIRECTORY) && go mod tidy
	@echo "Done tidying all modules"

format-all:
	@echo "Formatting and vetting all modules..."
	@cd $(APP_DIRECTORY) && go fmt ./... && go vet ./...
	@cd $(WORKER_DIRECTORY) && go fmt ./... && go vet ./...
	@cd $(SHARED_DIRECTORY) && go fmt ./... && go vet ./...
	@echo "Done formatting and vetting"


t-db-up:
	@migrate -database "$(TRANSACTION_DATABASE_URL)" -path migrations up

t-db-ver:
	@migrate -database "$(TRANSACTION_DATABASE_URL)" -path migrations version

t-db-down:
	@migrate -database "$(TRANSACTION_DATABASE_URL)" -path migrations down

t-db-force:
	@migrate -database "$(TRANSACTION_DATABASE_URL)" -path migrations force $(v)

db-create:
	@migrate create -ext sql -dir ./shared/migrations -seq $(v)


test-coverage-app:
	@go test -coverprofile=coverage.out ./services/app/internal/...
	@go tool cover -func=coverage.out

test-coverage-worker:
	@go test -coverprofile=coverage.out ./services/worker/internal/...
	@go tool cover -func=coverage.out

test-coverage:
	@go test -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out


test-html-app:
	@go test -coverprofile=coverage.out ./services/app/internal/...
	@go tool cover -html=coverage.out

test-html-worker:
	@go test -coverprofile=coverage.out ./services/worker/internal/...
	@go tool cover -html=coverage.out

test-html:
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out




.DEFAULT_GOAL := help
