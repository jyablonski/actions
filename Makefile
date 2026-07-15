# execute a release after merging a PR to main
# example: `make release VERSION=v1.0.3`
.PHONY: release
release:
	@if [ -z "$(VERSION)" ]; then \
		echo "❌ VERSION required. Usage: make release VERSION=v1.0.3"; \
		exit 1; \
	fi; \
	echo "Creating release $(VERSION)..."; \
	git tag $(VERSION); \
	git push origin $(VERSION); \
	$(MAKE) sync-v1

.PHONY: sync-v1
sync-v1:
	@echo "Syncing v1 branch with latest v1.* tag..."
	@latest_tag=$$(git tag -l "v1.*" --sort=-creatordate | head -n 1); \
	if [ -z "$$latest_tag" ]; then \
		echo "❌ No v1.* tags found"; \
		exit 1; \
	fi; \
	echo "Latest tag: $$latest_tag"; \
	git fetch --all --tags; \
	git checkout tags/$$latest_tag -b tmp-v1; \
	git push origin tmp-v1:v1 --force; \
	git checkout main; \
	git branch -D tmp-v1; \
	echo "v1 branch updated to $$latest_tag"

GO := go
GOLANGCI_LINT_VERSION := v2.12.2
BUILD_DIR := bin

.DEFAULT_GOAL := help

.PHONY: help build test vet deadcode lint ci pre-commit
help:
	@printf '%s\n' \
		'build       Build notify to bin/notify.' \
		'test        Run gotestsum and enforce 90% coverage.' \
		'vet         Run go vet.' \
		'deadcode    Check for unreachable Go code.' \
		'lint        Run the pinned golangci-lint version.' \
		'ci          Run the local equivalent of CI checks.' \
		'pre-commit  Run all pre-commit hooks.'

build:
	mkdir -p $(BUILD_DIR)
	$(GO) build -o $(BUILD_DIR)/notify ./cmd/notify

test:
	scripts/check-go-coverage.sh

vet:
	$(GO) vet ./...

deadcode:
	scripts/check-deadcode.sh

lint:
	$(GO) run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION) run

ci: build vet test deadcode lint

pre-commit:
	pre-commit run --all-files
