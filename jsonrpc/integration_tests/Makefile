# JSON-RPC Integration Tests Makefile

# =============================================================================
# CONFIGURATION
# =============================================================================

.PHONY: all test run-test test-quick clean help deps

# Determine git root and makefile directory
GIT_ROOT := $(shell git rev-parse --show-toplevel)
MAKEFILE_DIR := $(GIT_ROOT)/jsonrpc/integration_tests

# Command wrapper to run in the correct directory
RUN = cd $(MAKEFILE_DIR) &&

# Default test timeout
TIMEOUT ?= 5m

# Test package
PKG = ./tests

# Coverage output
COVERAGE_OUT = coverage.out
COVERAGE_HTML = coverage.html

# Extract test name from command line arguments for run-test target
ifeq (run-test,$(firstword $(MAKECMDGOALS)))
  RUN_ARGS := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
  $(eval $(RUN_ARGS):;@:)
endif

# =============================================================================
# MAIN TARGETS (Most Important)
# =============================================================================

# Default target
all: test

# Run all integration tests
test: deps
	$(call show_progress,all integration tests)
	$(call run_test,Test.*)
	$(call show_result,All integration tests)

# Run a specific test with verbose output and preserved artifacts
run-test: deps
	@if [ -z "$(TEST)" ] && [ -z "$(RUN_ARGS)" ]; then \
		echo "Usage: make run-test <test_name_or_pattern>"; \
		echo ""; \
		echo "Examples:"; \
		echo "  make run-test TestHTTPBasic"; \
		echo "  make run-test TestWebSocketServerStreaming"; \
		echo "  make run-test 'TestHTTP.*'"; \
		echo "  make run-test 'TestHTTPBasic/http_basic'"; \
		echo ""; \
		echo "Available tests:"; \
		cd $(MAKEFILE_DIR)/tests && go test -list . 2>/dev/null | grep '^Test' | sort | sed 's/^/  /'; \
		echo ""; \
		echo "Common patterns:"; \
		echo "  'TestHTTP.*'         - All HTTP transport tests"; \
		echo "  'TestWebSocket.*'    - All WebSocket tests"; \
		echo "  'TestSSE.*'          - All Server-Sent Events tests"; \
		echo "  '.*Error.*'          - All error handling tests"; \
		echo "  '.*Validation.*'     - All validation tests"; \
		echo ""; \
		exit 1; \
	fi
	$(eval RUN_TEST := $(if $(RUN_ARGS),$(RUN_ARGS),$(TEST)))
	@echo "Running specific test: $(RUN_TEST)"
	@echo "Timeout: $(TIMEOUT)"
	@echo "Artifacts will be preserved in: $(MAKEFILE_DIR)/tests/testdata/runs/"
	@echo ""
	@$(call run_test,$(RUN_TEST),KEEP_ARTIFACTS=1); \
	test_result=$$?; \
	echo ""; \
	echo "Test execution completed."; \
	echo ""; \
	echo "Generated artifacts preserved in:"; \
	echo "  $(MAKEFILE_DIR)/tests/testdata/runs/"; \
	echo ""; \
	echo "Recent test runs:"; \
	ls -la $(MAKEFILE_DIR)/tests/testdata/runs/ 2>/dev/null | tail -5 || echo "  (no artifacts found)"; \
	exit $$test_result

# Run quick subset of tests (short mode)
test-quick: deps
	$(call show_progress,quick integration tests (short mode))
	@$(RUN) go test -short -timeout 30s -v $(PKG) -run Test
	$(call show_result,Quick tests)

# =============================================================================
# HELPER FUNCTIONS
# =============================================================================

# Progress indicator functions
define show_progress
	@echo "Running $(1)..."
	@echo "Timeout: $(TIMEOUT)"
	@echo ""
endef

define show_result
	@echo ""
	@echo "$(1) completed."
endef

# Run go test with raw output - simple and clean
# Usage: $(call run_test,test_pattern,optional_env_vars,optional_timeout)
define run_test
	$(RUN) $(if $(2),$(2)) go test -timeout $(if $(3),$(3),$(TIMEOUT)) -v $(PKG) -run '$(1)'
endef



# =============================================================================
# UTILITY TARGETS
# =============================================================================

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	@$(RUN) go mod download && $(RUN) go mod tidy
	@echo "Dependencies ready."

# Clean test artifacts
clean:
	@echo "Cleaning test artifacts..."
	@$(RUN) rm -f $(COVERAGE_OUT) $(COVERAGE_HTML) 2>/dev/null || true
	@$(RUN) rm -rf tests/testdata/runs/* 2>/dev/null || true
	@echo "Clean completed."

# Show help information
help:
	@echo "JSON-RPC Integration Test Targets:"
	@echo "Note: This Makefile can be run from anywhere in the git repository."
	@echo ""
	@echo "MAIN TARGETS:"
	@echo "  make test            - Run all integration tests"
	@echo "  make run-test        - Run a specific test with verbose output and preserved artifacts"
	@echo "  make test-quick      - Run quick subset of tests"
	@echo ""
	@echo "UTILITY TARGETS:"
	@echo "  make clean           - Clean test artifacts"
	@echo "  make deps            - Download dependencies"
	@echo "  make help            - Show this help"
	@echo ""
	@echo "EXAMPLES:"
	@echo "  make run-test TestHTTPBasic"
	@echo "  make run-test 'TestHTTP.*'"
	@echo "  make run-test 'TestHTTPBasic/http_basic'"
	@echo ""
	@echo "OPTIONS:"
	@echo "  TIMEOUT=5m          - Set test timeout (default: 5m)"
	@echo ""
