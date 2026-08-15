SHELL := /bin/sh

COMPOSE ?= docker compose
COMPOSE_FILE := deploy/compose/compose.yml
E2E_API_URL ?= http://localhost:8080/api/v1
HURL ?= hurl

.DEFAULT_GOAL := help

.PHONY: help setup dev web-dev api-dev worker-dev payment-dev search-dev search-migrate search-reindex \
	compose-up compose-full compose-down compose-logs migrate seed reset \
	generate fmt lint test e2e-api outage-test performance build check security replay demo-reset \
	helm-check helm-regression k3s-e2e infra-check platform-check deployment-smoke microservice-test search-outage clean

help: ## Show available commands
	@awk 'BEGIN {FS = ":.*## "; printf "Cartlabs commands:\n"} /^[a-zA-Z_-]+:.*## / {printf "  %-16s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

setup: ## Install dependencies and generate API types
	pnpm install --frozen-lockfile
	go mod download
	$(MAKE) generate

dev: ## Run web, API, worker, search, and mock payment locally
	@set -a; \
	if [ -f .env ]; then . ./.env || exit 1; fi; \
	set +a; \
	exec $(MAKE) -j5 web-dev api-dev worker-dev search-dev payment-dev

web-dev: ## Run Next.js development server
	pnpm web:dev

api-dev: ## Run Go API
	go run ./apps/api

worker-dev: ## Run Go worker
	go run ./apps/worker

payment-dev: ## Run mock payment service
	go run ./apps/mock-payment

search-dev: ## Run search gRPC service
	go run ./apps/search

compose-up: ## Start local infrastructure dependencies
	$(COMPOSE) -f $(COMPOSE_FILE) up -d

compose-full: ## Build and start complete containerized stack
	$(COMPOSE) -f $(COMPOSE_FILE) --profile full up -d --build

compose-down: ## Stop complete stack, preserving volumes
	$(COMPOSE) -f $(COMPOSE_FILE) --profile full down

compose-logs: ## Follow complete stack logs
	$(COMPOSE) -f $(COMPOSE_FILE) --profile full logs -f

migrate: ## Apply pending PostgreSQL migrations
	@set -a; [ ! -f .env ] || . ./.env; set +a; go run ./apps/migrate && go run ./apps/search-migrate

search-migrate: ## Create and migrate search-owned database
	@set -a; [ ! -f .env ] || . ./.env; set +a; go run ./apps/search-migrate

search-reindex: ## Rebuild search documents from authoritative catalog
	@set -a; [ ! -f .env ] || . ./.env; set +a; go run ./apps/search-reindex

seed: ## Apply deterministic demo seed
	@set -a; [ ! -f .env ] || . ./.env; set +a; go run ./apps/seed; \
		[ -z "$$SEARCH_GRPC_ADDR" ] || go run ./apps/search-reindex

reset: ## Rebuild local/demo database from migrations and seed
	@set -a; [ ! -f .env ] || . ./.env; set +a; go run ./apps/reset

demo-reset: ## Run recoverable demo reset job through Compose
	$(COMPOSE) -f $(COMPOSE_FILE) --profile full --profile ops run --rm reset

replay: ## Replay bounded SQL and RabbitMQ dead letters
	$(COMPOSE) -f $(COMPOSE_FILE) --profile full --profile ops run --rm replay

generate: ## Generate Go and TypeScript code from OpenAPI and protobuf contracts
	go tool oapi-codegen -config api/openapi/oapi-codegen.yaml api/openapi/openapi.yaml
	pnpm openapi:generate:ts
	protoc -I . --plugin=protoc-gen-go="$$(go tool -n protoc-gen-go)" \
		--plugin=protoc-gen-go-grpc="$$(go tool -n protoc-gen-go-grpc)" \
		--go_out=. --go_opt=module=github.com/christolx/cartlabs \
		--go-grpc_out=. --go-grpc_opt=module=github.com/christolx/cartlabs \
		api/proto/search/v1/search.proto

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

outage-test: ## Exercise Redis, RabbitMQ, and payment-provider recovery
	bash tests/operations/outage.sh

search-outage: ## Prove search fallback and recovery against Compose
	bash tests/microservices/search-outage.sh

microservice-test: ## Run independent search persistence and catalog-event integration tests
	INTEGRATION_DATABASE_URL="$${INTEGRATION_DATABASE_URL:-postgres://cartlabs:cartlabs@localhost:5432/cartlabs?sslmode=disable}" \
	INTEGRATION_SEARCH_DATABASE_URL="$${INTEGRATION_SEARCH_DATABASE_URL:-postgres://cartlabs:cartlabs@localhost:5432/cartlabs_search?sslmode=disable}" \
		go test -tags=integration ./internal/catalog ./internal/search

performance: ## Run repeatable public catalog performance baseline
	go run ./tests/performance -url '$(E2E_API_URL)/catalog/products?pageSize=24' -duration 15s -concurrency 20 -max-p95 250ms

build: ## Build all runtime applications
	go build ./apps/...
	pnpm web:build

security: ## Scan Go and frontend dependency vulnerabilities
	go tool govulncheck ./...
	pnpm audit --audit-level high

helm-regression: ## Check rendered k3s behavior that schema validation cannot prove
	bash tests/deployment/helm-regression.sh

helm-check: helm-regression ## Lint and render the k3s Helm chart
	helm lint deploy/helm --values deploy/helm/values-local.yaml
	helm lint deploy/helm --values deploy/helm/values-demo.yaml
	helm template cartlabs deploy/helm --namespace cartlabs --values deploy/helm/values-local.yaml | \
		kubeconform -strict -summary -kubernetes-version 1.36.0 -ignore-missing-schemas
	helm template cartlabs deploy/helm --namespace cartlabs --values deploy/helm/values-demo.yaml | \
		kubeconform -strict -summary -kubernetes-version 1.36.0 -ignore-missing-schemas

infra-check: ## Format and validate Terraform and Ansible
	terraform -chdir=infra/terraform fmt -check -recursive
	terraform -chdir=infra/terraform init -backend=false
	terraform -chdir=infra/terraform validate
	cd infra/ansible && ansible-lint playbook.yml
	cd infra/ansible && ansible-playbook -i inventory.example.yml playbook.yml --syntax-check \
		-e cert_manager_email=operator@example.com -e demo_domain=demo.cartlabs.example.com

platform-check: helm-check infra-check ## Validate deployment and infrastructure assets

deployment-smoke: ## Smoke-test an already deployed demo URL
	bash tests/deployment/smoke.sh "$${DEMO_URL:?set DEMO_URL}"

k3s-e2e: ## Build, import, install, upgrade, and verify an isolated local k3s release
	bash tests/deployment/k3s-e2e.sh

check: generate lint test build ## Run full local verification
	@git diff --exit-code -- internal/contract/openapi.gen.go apps/web/src/lib/api/schema.d.ts

clean: ## Remove generated build output
	rm -rf apps/web/.next coverage bin
