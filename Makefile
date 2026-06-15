.DEFAULT_GOAL := help
SHELL := /bin/bash

BACKEND := backend
DB_URL ?= postgres://dios:dios@localhost:5432/dios?sslmode=disable
MIGRATIONS := $(BACKEND)/migrations

.PHONY: help
help: ## Show this help
	@grep -hE '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
	awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

.PHONY: generate
generate: ## Run code generation (sqlc + gqlgen)
	cd $(BACKEND) && sqlc generate
	cd $(BACKEND) && go run github.com/99designs/gqlgen generate

.PHONY: lint
lint: ## Run linters
	cd $(BACKEND) && gofmt -l . | tee /dev/stderr | (! read)
	cd $(BACKEND) && golangci-lint run ./...

.PHONY: test
test: ## Run tests
	cd $(BACKEND) && go test ./...

.PHONY: test-race
test-race: ## Run tests with race detector and coverage
	cd $(BACKEND) && go test -race -coverprofile=coverage.out -covermode=atomic ./...

.PHONY: build
build: ## Build the API binary
	cd $(BACKEND) && go build -o bin/api ./cmd/api

.PHONY: run
run: ## Run the API locally
	cd $(BACKEND) && go run ./cmd/api

.PHONY: migrate-up
migrate-up: ## Apply all migrations
	migrate -path $(MIGRATIONS) -database "$(DB_URL)" up

.PHONY: migrate-down
migrate-down: ## Roll back one migration
	migrate -path $(MIGRATIONS) -database "$(DB_URL)" down 1

.PHONY: migrate-create
migrate-create: ## Create a migration: make migrate-create name=add_widgets
	migrate create -ext sql -dir $(MIGRATIONS) -seq $(name)

.PHONY: up
up: ## Start the full stack (docker compose)
	docker compose -f infra/docker-compose.yml up --build

.PHONY: down
down: ## Stop the stack
	docker compose -f infra/docker-compose.yml down
