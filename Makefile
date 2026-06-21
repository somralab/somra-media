# Somra — top-level Makefile
#
# This file owns the developer entry-point. CI mirrors most of these targets;
# anything CI executes directly (scripts/, .github/workflows/ci.yml) is the
# source of truth for the gate logic — keep them aligned when editing.

SHELL := /bin/bash

GO          ?= go
PNPM        ?= pnpm
BIN_DIR     := bin
BINARY      := $(BIN_DIR)/somra
GO_COVER    := coverage.go.out
WEB_DIR     := web
WEB_DIST    := $(WEB_DIR)/dist
WEB_COVER   := $(WEB_DIR)/coverage/coverage-summary.json
OPENAPI     := api/openapi.yaml

IMAGE       ?= ghcr.io/somralab/somra-media
TAG         ?= dev
VERSION     ?= $(shell git describe --tags --always 2>/dev/null || echo dev)
COMMIT      ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
BUILT_AT    ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS := -s -w \
  -X main.version=$(VERSION) \
  -X main.commit=$(COMMIT) \
  -X main.builtAt=$(BUILT_AT)

.DEFAULT_GOAL := help

GO_TAGS       ?=
ACQ_TAGS      := -tags acquisition

.PHONY: help dev dev-backend dev-frontend build build-go build-acquisition build-web \
        test test-go test-acquisition test-web lint lint-go lint-web \
        migrate coverage coverage-go coverage-web coverage-gate \
        i18n-check docker docker-acquisition docker-multiarch openapi-types docs-api e2e \
        profile bundle-check soak-test clean

## help: show this list
help:
	@printf "Somra — make targets\n\n"
	@awk '/^## / {sub(/^## /,""); printf "  %s\n", $$0}' $(MAKEFILE_LIST)

## dev: run backend + frontend dev servers concurrently
dev:
	@trap 'kill 0' EXIT INT TERM; \
	set -a; [ -f .env ] && . ./.env; set +a; \
	echo ">> $(GO) run ./cmd/somra (background)"; \
	$(GO) run ./cmd/somra & \
	echo ">> $(PNPM) --dir $(WEB_DIR) run dev"; \
	$(PNPM) --dir $(WEB_DIR) run dev & \
	wait

## dev-backend: run only the Go backend
dev-backend:
	@set -a; [ -f .env ] && . ./.env; set +a; \
	$(GO) run ./cmd/somra

## dev-frontend: run only the Vite dev server
dev-frontend:
	$(PNPM) --dir $(WEB_DIR) run dev

## build: build everything (backend + frontend)
build: build-go build-web

## build-go: compile the Go binary (CGO disabled, core — no acquisition adapters)
build-go:
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 $(GO) build -trimpath -ldflags '$(LDFLAGS)' -o $(BINARY) ./cmd/somra

## build-acquisition: compile with acquisition plugin adapters (full image)
build-acquisition:
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 $(GO) build $(ACQ_TAGS) -trimpath -ldflags '$(LDFLAGS)' -o $(BINARY)-acquisition ./cmd/somra

## build-web: produce the SPA static bundle
build-web:
	$(PNPM) --dir $(WEB_DIR) install --frozen-lockfile
	$(PNPM) --dir $(WEB_DIR) run build

## test: run backend + frontend tests
test: test-go test-web

## test-go: run all Go unit tests (CGO-free; no -race — see CI unit-test job)
test-go:
	$(GO) test ./... -count=1

## test-acquisition: run Go tests including acquisition-tagged packages
test-acquisition:
	$(GO) test $(ACQ_TAGS) ./... -count=1

## test-web: run frontend tests once
test-web:
	$(PNPM) --dir $(WEB_DIR) test --run

## lint: run all linters (Go + frontend)
lint: lint-go lint-web

## lint-go: run gofmt + golangci-lint
lint-go:
	@diff=$$(gofmt -l .); \
	if [ -n "$$diff" ]; then \
	  echo "gofmt found unformatted files:"; echo "$$diff"; exit 1; \
	fi
	@if command -v golangci-lint >/dev/null 2>&1; then \
	  echo ">> golangci-lint run ./..."; \
	  golangci-lint run ./...; \
	else \
	  echo "lint-go: golangci-lint not installed — install with: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b \$$($(GO) env GOPATH)/bin"; \
	  exit 1; \
	fi

## lint-web: run frontend lint + prettier check
lint-web:
	$(PNPM) --dir $(WEB_DIR) run lint
	$(PNPM) --dir $(WEB_DIR) run format:check
	$(PNPM) --dir $(WEB_DIR) run typecheck

## migrate: apply database migrations (no-op for now; runs at server startup)
migrate:
	@echo "migrate: migrations are applied automatically on server startup."

## coverage: produce backend + frontend coverage reports
coverage: coverage-go coverage-web

## coverage-go: run Go tests with coverage and print summary
coverage-go:
	$(GO) test ./... -count=1 -coverprofile=$(GO_COVER)
	$(GO) tool cover -func=$(GO_COVER) | tail -5

## coverage-web: run frontend tests with coverage (v8, json-summary)
coverage-web:
	$(PNPM) --dir $(WEB_DIR) run test:coverage

## coverage-gate: enforce DoD §4.1 coverage thresholds
coverage-gate: coverage
	bash scripts/coverage-gate.sh

## i18n-check: validate translation key parity (en-US / tr-TR)
i18n-check:
	bash scripts/i18n-check.sh

## e2e-fixture: generate playback test media via ffmpeg
e2e-fixture:
	bash scripts/gen-e2e-media.sh

## docker: build the container image for the local architecture (core)
docker:
	docker buildx build \
	  --platform $$(uname -m | sed 's/x86_64/linux\/amd64/;s/aarch64/linux\/arm64/;s/arm64/linux\/arm64/') \
	  --build-arg VERSION=$(VERSION) \
	  --build-arg COMMIT=$(COMMIT) \
	  --build-arg BUILT_AT=$(BUILT_AT) \
	  -f deploy/Dockerfile \
	  -t $(IMAGE):$(TAG) \
	  --load .

## docker-acquisition: build image with acquisition adapters enabled
docker-acquisition:
	docker buildx build \
	  --platform $$(uname -m | sed 's/x86_64/linux\/amd64/;s/aarch64/linux\/arm64/;s/arm64/linux\/arm64/') \
	  --build-arg VERSION=$(VERSION) \
	  --build-arg COMMIT=$(COMMIT) \
	  --build-arg BUILT_AT=$(BUILT_AT) \
	  --build-arg BUILD_TAGS=acquisition \
	  -f deploy/Dockerfile \
	  -t $(IMAGE):$(TAG)-acquisition \
	  --load .

## docker-multiarch: build + push multi-arch (amd64 + arm64) image
docker-multiarch:
	docker buildx build \
	  --platform linux/amd64,linux/arm64 \
	  --build-arg VERSION=$(VERSION) \
	  --build-arg COMMIT=$(COMMIT) \
	  --build-arg BUILT_AT=$(BUILT_AT) \
	  -f deploy/Dockerfile \
	  -t $(IMAGE):$(TAG) \
	  --push .

## openapi-types: regenerate frontend TypeScript types from the OpenAPI spec
openapi-types:
	bash scripts/gen-openapi-types.sh

## docs-api: generate static Redoc HTML from the OpenAPI spec
docs-api:
	bash scripts/gen-api-docs.sh

## e2e: install playwright browsers and run e2e specs
e2e: e2e-fixture build-web
	$(PNPM) --dir $(WEB_DIR) exec playwright install --with-deps chromium
	$(PNPM) --dir $(WEB_DIR) exec playwright test

## profile: capture Go CPU/memory profiles for hot packages (Sprint 10 A1)
profile:
	bash scripts/profile.sh

## bundle-check: enforce frontend gzip bundle budget after build-web (Sprint 10 C1)
bundle-check: build-web
	bash scripts/bundle-budget.sh

## soak-test: run shortened server soak against local binary (Sprint 10 D1)
soak-test: build-go
	bash scripts/soak-test.sh

## clean: remove build/test artefacts
clean:
	rm -rf $(BIN_DIR) $(GO_COVER) $(WEB_DIST) $(WEB_DIR)/coverage \
	  $(WEB_DIR)/playwright-report $(WEB_DIR)/test-results
