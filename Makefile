SHELL := /bin/sh

export GOTOOLCHAIN := local

GO := $(shell command -v go)
GO_PACKAGES := $(shell $(GO) list ./... 2>/dev/null)

.PHONY: format format-check lint typecheck test test-race vuln quality verify verify-build

format:
	$(GO) tool gofumpt -w .

format-check:
	@test -z "$$($(GO) tool gofumpt -l .)" || { \
		echo "Go files are not formatted. Run 'make format'."; \
		$(GO) tool gofumpt -l .; \
		exit 1; \
	}

lint:
	@if test -n "$(GO_PACKAGES)"; then \
		$(GO) tool staticcheck ./...; \
	else \
		echo "lint: no Go packages to check"; \
	fi

typecheck:
	@if test -n "$(GO_PACKAGES)"; then \
		$(GO) vet ./...; \
	else \
		echo "typecheck: no Go packages to check"; \
	fi

test:
	@if test -n "$(GO_PACKAGES)"; then \
		CGO_ENABLED=0 $(GO) test ./...; \
	else \
		echo "test: no Go packages to test"; \
	fi

test-race:
	@if test "$$($(GO) env GOOS)/$$($(GO) env GOARCH)" != "linux/amd64"; then \
		echo "test-race requires Linux/amd64"; \
		exit 1; \
	elif test -n "$(GO_PACKAGES)"; then \
		CGO_ENABLED=1 $(GO) test -race ./...; \
	else \
		echo "test-race: no Go packages to test"; \
	fi

vuln:
	@if test -n "$(GO_PACKAGES)"; then \
		$(GO) tool govulncheck ./...; \
	else \
		echo "vuln: no Go packages to check"; \
	fi

quality: format-check lint typecheck test

verify-build:
	@if test -n "$(GO_PACKAGES)"; then \
		for arch in amd64 arm64; do \
			echo "verify-build: linux/$$arch"; \
			CGO_ENABLED=0 GOOS=linux GOARCH=$$arch $(GO) build ./...; \
		done; \
	else \
		echo "verify-build: no Go packages to build"; \
	fi

verify: quality vuln verify-build
