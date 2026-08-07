SHELL := /bin/sh

COMPOSE ?= docker compose
COMPOSE_FILE := deploy/compose/compose.yml
E2E_API_URL ?= http://localhost:8080/api/v1
HURL ?= hurl

.DEFAULT_GOAL := help

.PHONY: help setup dev web-dev api-dev worker-dev payment-dev \
	compose-up compose-full compose-down compose-logs migrate seed reset \
	generate fmt lint test e2e-api build check clean

help: ## Show available commands
	@awk 'BEGIN {FS = ":.*## "; printf "Cartlabs commands:\n"} /^[a-zA-Z_-]+:.*## / {printf "  %-16s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

setup: ## Install dependencies and generate API types
	pnpm install --frozen-lockfile
	go mod download
	$(MAKE) generate

dev: ## Run web, API, worker, and mock payment locally
	$(MAKE) -j4 web-dev api-dev worker-dev payment-dev

web-dev: ## Run Next.js development server
	pnpm web:dev

api-dev: ## Run Go API
	go run ./apps/api

worker-dev: ## Run Go worker
	go run ./apps/worker

payment-dev: ## Run mock payment service
	go run ./apps/mock-payment

compose-up: ## Start local infrastructure dependencies
	$(COMPOSE) -f $(COMPOSE_FILE) up -d

compose-full: ## Build and start complete containerized stack
	$(COMPOSE) -f $(COMPOSE_FILE) --profile full up -d --build

compose-down: ## Stop complete stack, preserving volumes
	$(COMPOSE) -f $(COMPOSE_FILE) --profile full down

compose-logs: ## Follow complete stack logs
	$(COMPOSE) -f $(COMPOSE_FILE) --profile full logs -f

migrate: ## Apply pending PostgreSQL migrations
	@set -a; [ ! -f .env ] || . ./.env; set +a; go run ./apps/migrate

seed: ## Apply deterministic demo seed
	@set -a; [ ! -f .env ] || . ./.env; set +a; go run ./apps/seed

reset: ## Rebuild local/demo database from migrations and seed
	@set -a; [ ! -f .env ] || . ./.env; set +a; go run ./apps/reset

generate: ## Generate Go and TypeScript code from OpenAPI
	go tool oapi-codegen -config api/openapi/oapi-codegen.yaml api/openapi/openapi.yaml
	pnpm openapi:generate:ts

fmt: ## Format Go source
	gofmt -w $$(find apps internal -name '*.go' -type f)

lint: ## Run format, static, frontend, and OpenAPI checks
	@test -z "$$(gofmt -l $$(find apps internal -name '*.go' -type f))"
	go vet ./...
	pnpm web:lint
	pnpm openapi:lint

test: ## Run backend tests and frontend type checking
	go test ./...
	pnpm web:typecheck

e2e-api: ## Run black-box API workflow tests against a running API
	$(HURL) --test --jobs 1 --error-format long --retry 10 \
		--variable base_url=$(E2E_API_URL) tests/e2e/api/*.hurl

build: ## Build all runtime applications
	go build ./apps/...
	pnpm web:build

check: generate lint test build ## Run full local verification
	@git diff --exit-code -- internal/contract/openapi.gen.go apps/web/src/lib/api/schema.d.ts

clean: ## Remove generated build output
	rm -rf apps/web/.next coverage bin
