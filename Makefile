.PHONY: all help install_base install_deps build_docs build build_wasm test run

BIN_DIR ?= bin
DOCS_DIR ?= docs

# If the first argument is "run"...
ifeq (run,$(firstword $(MAKECMDGOALS)))
  # use the rest as arguments for "run"
  RUN_ARGS := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
  # ...and turn them into do-nothing targets
  $(eval $(RUN_ARGS) $(ARGS):;@:)
endif

all: help

help:
	@echo "Available targets:"
	@echo "  install_base : install Go runtime (assumes 'go' is already in PATH)"
	@echo "  install_deps : install local dependencies (go mod download)"
	@echo "  build_docs   : build the API docs and put them in the docs directory."
	@echo "  build        : build the CLI binary."
	@echo "  build_wasm   : build the WASM binary."
	@echo "  test         : run tests locally"
	@echo "  run          : run the CLI. Usage: make run --version"
	@echo "  help         : show this help text"

install_base:
	@command -v go >/dev/null 2>&1 || { echo >&2 "Go is not installed. Please install Go 1.21+."; exit 1; }
	@echo "Go is installed."

install_deps:
	go mod download

build_docs:
	@mkdir -p $(DOCS_DIR)
	@echo "Building docs to $(DOCS_DIR)..."
	@go run scripts/doc_cover.go > $(DOCS_DIR)/doc_coverage.txt || echo "doc_cover.go failed"
	@echo "Docs built."

build:
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/cdd-go ./cmd/cdd-go

test:
	go test -v -cover ./...

run: build
	./$(BIN_DIR)/cdd-go $(RUN_ARGS) $(ARGS)

build_wasm:
	@mkdir -p $(BIN_DIR)
	GOOS=js GOARCH=wasm go build -o $(BIN_DIR)/cdd-go.wasm ./cmd/cdd-go
