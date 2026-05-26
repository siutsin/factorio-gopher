GREEN := \033[0;32m
YELLOW := \033[1;33m
RED := \033[0;31m
NC := \033[0m

.DEFAULT_GOAL := help

.PHONY: help
help: ## Show this help message
	@echo "$(GREEN)Available targets:$(NC)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(YELLOW)%-20s$(NC) %s\n", $$1, $$2}'

.PHONY: lint
lint: lint-markdown lint-lua lint-go ## Run all linters

.PHONY: install
install: ## Symlink mod/ into Factorio's mods folder
	@go run ./cmd install

.PHONY: uninstall
uninstall: ## Remove the symlink from Factorio's mods folder
	@go run ./cmd uninstall

.PHONY: build
build: ## Rebuild all derived sprite sheets
	@echo "$(GREEN)Building sprite sheets...$(NC)"
	@command -v go >/dev/null 2>&1 || { \
		echo "$(RED)go not installed. Install with: brew install go$(NC)"; \
		exit 1; \
	}
	@go run ./cmd all
	@echo "$(GREEN)Build complete.$(NC)"

.PHONY: package
package: ## Create Factorio mod zip under build/
	@echo "$(GREEN)Packaging mod...$(NC)"
	@command -v jq >/dev/null 2>&1 || { \
		echo "$(RED)jq not installed. Install with: brew install jq$(NC)"; \
		exit 1; \
	}
	@command -v zip >/dev/null 2>&1 || { \
		echo "$(RED)zip not installed$(NC)"; \
		exit 1; \
	}
	@set -eu; \
	name=$$(jq -r '.name' mod/info.json); \
	version=$$(jq -r '.version' mod/info.json); \
	root="$${name}_$${version}"; \
	out="build/$${root}.zip"; \
	out_abs="$$(pwd)/$${out}"; \
	tmp=$$(mktemp -d); \
	trap 'rm -rf "$${tmp}"' EXIT; \
	mkdir -p build "$${tmp}/$${root}"; \
	cp -R mod/. "$${tmp}/$${root}/"; \
	cp LICENSE "$${tmp}/$${root}/LICENSE"; \
	find "$${tmp}/$${root}" -name AGENTS.md -type f -delete; \
	find "$${tmp}/$${root}" -name .DS_Store -type f -delete; \
	rm -rf "$${tmp}/$${root}/spec"; \
	rm -f "$${out}"; \
	( cd "$${tmp}" && zip -qr "$${out_abs}" "$${root}" ); \
	zip -T "$${out}" >/dev/null; \
	echo "$(GREEN)Wrote $${out}$(NC)"

.PHONY: test
test: test-go test-lua ## Run all tests

.PHONY: test-go
test-go: ## Run Go tests
	@echo "$(GREEN)Running Go tests...$(NC)"
	@command -v go >/dev/null 2>&1 || { \
		echo "$(RED)go not installed. Install with: brew install go$(NC)"; \
		exit 1; \
	}
	@go test -race -cover ./...
	@echo "$(GREEN)Go tests passed!$(NC)"

.PHONY: test-lua
test-lua: ## Run Lua specs with busted (and luacov if installed)
	@echo "$(GREEN)Running Lua specs...$(NC)"
	@command -v busted >/dev/null 2>&1 || { \
		echo "$(RED)busted not installed. Install with: brew install busted$(NC)"; \
		exit 1; \
	}
	@busted -c mod/spec
	@if command -v luacov >/dev/null 2>&1; then \
		luacov && tail -n 5 luacov.report.out; \
	elif [ -x "$$HOME/.luarocks/bin/luacov" ]; then \
		"$$HOME/.luarocks/bin/luacov" && tail -n 5 luacov.report.out; \
	else \
		echo "$(YELLOW)luacov not installed; skipping coverage report.$(NC)"; \
	fi
	@echo "$(GREEN)Lua specs passed!$(NC)"

.PHONY: lint-markdown
lint-markdown: ## Lint Markdown files with markdownlint-cli2
	@echo "$(GREEN)Linting Markdown...$(NC)"
	@command -v markdownlint-cli2 >/dev/null 2>&1 || { \
		echo "$(RED)markdownlint-cli2 not installed. Install with: brew install markdownlint-cli2$(NC)"; \
		exit 1; \
	}
	@markdownlint-cli2 "**/*.md"
	@echo "$(GREEN)Markdown lint passed!$(NC)"

.PHONY: lint-lua
lint-lua: ## Lint Lua files with luacheck
	@echo "$(GREEN)Linting Lua...$(NC)"
	@command -v luacheck >/dev/null 2>&1 || { \
		echo "$(RED)luacheck not installed. Install with: brew install luacheck$(NC)"; \
		exit 1; \
	}
	@luacheck mod/
	@echo "$(GREEN)Lua lint passed!$(NC)"

.PHONY: lint-go
lint-go: ## Lint Go files with golangci-lint
	@echo "$(GREEN)Linting Go...$(NC)"
	@command -v golangci-lint >/dev/null 2>&1 || { \
		echo "$(RED)golangci-lint not installed. Install with: brew install golangci-lint$(NC)"; \
		exit 1; \
	}
	@golangci-lint run ./...
	@echo "$(GREEN)Go lint passed!$(NC)"
