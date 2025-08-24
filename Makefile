SHELL := /bin/bash

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


.DEFAULT_GOAL := help
