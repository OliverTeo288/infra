SHELL := /bin/bash
GO    ?= go

.PHONY: help build install test test-race test-integration lint fmt fmt-check vet tidy clean
.DEFAULT_GOAL := help

help: ## Show available targets
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the infra binary with build-time vars from defaults
	@scripts/build-local.sh build

install: ## Install infra into $GOBIN
	@$(GO) install ./...

test: ## Run unit tests
	@$(GO) test -count=1 ./...

test-race: ## Run unit tests with the race detector
	@$(GO) test -race -count=1 ./...

test-integration: ## Run integration tests against a local floci emulator
	@scripts/test-floci.sh

fmt: ## Format all Go code
	@gofmt -w .

vet: ## Run go vet
	@$(GO) vet ./...

lint: fmt-check vet ## Run all linters (gofmt + vet + staticcheck if available)
	@if command -v staticcheck >/dev/null 2>&1; then \
		staticcheck ./...; \
	else \
		echo "staticcheck not installed; skipping. install: go install honnef.co/go/tools/cmd/staticcheck@latest"; \
	fi

fmt-check: ## Verify all Go code is gofmt'd (exit non-zero otherwise)
	@unformatted=$$(gofmt -l .); \
	if [ -n "$$unformatted" ]; then \
		echo "::error::files need gofmt:"; \
		echo "$$unformatted"; \
		exit 1; \
	fi

tidy: ## Run go mod tidy
	@$(GO) mod tidy

clean: ## Remove build artefacts
	@rm -f infra coverage.txt
