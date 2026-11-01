# MOAJ (Mentor of all Jacks) - Build, Test, Boundary & License Verification Makefile
# Usage:
#   make build           - Build Community Edition binary (./bin/moaj)
#   make build-ee        - Build Enterprise Edition binary with -tags ee (./bin/moaj-ee)
#   make test            - Run unit tests (Community build)
#   make test-ee         - Run unit tests (Enterprise -tags ee build)
#   make test-race       - Run unit tests with Go race detector (Community & EE)
#   make vet             - Run go vet on both Community and Enterprise build tags
#   make fmt-check       - Verify all Go files are formatted with gofmt
#   make boundary-check  - Verify default Community build imports zero packages from ee/
#   make license-check   - Verify all go.mod dependencies match approved permissive allowlist
#   make generate        - Run SQLc code generation (Phase 2+)
#   make ui              - Build static SvelteKit frontend bundle (Phase 4+)
#   make clean           - Remove build artifacts in ./bin

GO          ?= go
GOFMT       ?= gofmt
SQLC        ?= sqlc
GO_LICENSES ?= go-licenses
BIN_DIR     := bin
BINARY      := $(BIN_DIR)/moaj
BINARY_EE   := $(BIN_DIR)/moaj-ee
MODULE      := github.com/MOAJ-Mentor-of-all-Jacks/MOAJ

# Approved permissive runtime module allowlist (see docs/governance/DEPENDENCY_LICENSES.md)
APPROVED_MODULES := \
	github.com/MOAJ-Mentor-of-all-Jacks/MOAJ \
	github.com/prometheus/client_golang \
	github.com/beorn7/perks \
	github.com/cespare/xxhash/v2 \
	github.com/munnerz/goautoneg \
	github.com/prometheus/client_model \
	github.com/prometheus/common \
	github.com/prometheus/procfs \
	golang.org/x/sys \
	google.golang.org/protobuf

.PHONY: all build build-ee test test-ee test-race vet fmt-check boundary-check license-check generate ui clean

all: fmt-check vet boundary-check license-check test test-ee build build-ee

build:
	@mkdir -p $(BIN_DIR)
	$(GO) build -trimpath -o $(BINARY) ./cmd/moaj

build-ee:
	@mkdir -p $(BIN_DIR)
	$(GO) build -trimpath -tags ee -o $(BINARY_EE) ./cmd/moaj

test:
	$(GO) test ./...

test-ee:
	$(GO) test -tags ee ./...

test-race:
	$(GO) test -race ./...
	$(GO) test -race -tags ee ./...

vet:
	$(GO) vet ./...
	$(GO) vet -tags ee ./...

fmt-check:
	@unformatted=$$($(GOFMT) -l .); \
	if [ -n "$$unformatted" ]; then \
		echo "ERROR: The following Go files are not formatted with gofmt:"; \
		echo "$$unformatted"; \
		exit 1; \
	fi
	@echo "OK: All Go files are properly formatted."

boundary-check:
	@ee_deps=$$($(GO) list -deps ./cmd/moaj | grep '^$(MODULE)/ee' || true); \
	if [ -n "$$ee_deps" ]; then \
		echo "ERROR: Community build imports proprietary ee/ packages:"; \
		echo "$$ee_deps"; \
		exit 1; \
	fi
	@echo "OK: Community build dependency graph contains zero ee/ packages."

license-check:
	@for mod in $$($(GO) list -m -f '{{.Path}}' all); do \
		allowed=0; \
		for ok in $(APPROVED_MODULES); do \
			if [ "$$mod" = "$$ok" ]; then allowed=1; break; fi; \
		done; \
		if [ "$$allowed" -eq 0 ]; then \
			echo "ERROR: Unapproved module '$$mod' found in go.mod."; \
			echo "Verify permissive SPDX license (NFR-03) and update docs/governance/DEPENDENCY_LICENSES.md and Makefile."; \
			exit 1; \
		fi; \
	done
	@if command -v $(GO_LICENSES) >/dev/null 2>&1; then \
		$(GO_LICENSES) check ./cmd/moaj --disallowed_types=forbidden,restricted,reciprocal,unknown; \
	fi
	@echo "OK: All distributed Go modules satisfy the permissive license allowlist (NFR-03)."

generate:
	@if command -v $(SQLC) >/dev/null 2>&1; then \
		$(SQLC) generate; \
	else \
		echo "INFO: sqlc not installed; skipping code generation (configured in Phase 2)."; \
	fi

ui:
	@if [ -d "ui/node_modules" ]; then \
		cd ui && npm run build; \
	else \
		echo "INFO: ui/node_modules not present; UI build activates in Phase 4."; \
	fi

clean:
	rm -rf $(BIN_DIR)
