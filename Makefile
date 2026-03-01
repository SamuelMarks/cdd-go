.PHONY: all help install_base install_deps build_docs build build_wasm test run default

BIN_DIR ?= bin
DOCS_DIR ?= docs

# If the first argument is "run"...
ifeq (run,$(firstword $(MAKECMDGOALS)))
  # use the rest as arguments for "run"
  RUN_ARGS := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
  # ...and turn them into do-nothing targets
  $(eval $(RUN_ARGS) $(ARGS):;@:)
endif

# If the first argument is "build"...
ifeq (build,$(firstword $(MAKECMDGOALS)))
  ifneq (,$(word 2, $(MAKECMDGOALS)))
    BIN_DIR := $(word 2, $(MAKECMDGOALS))
    $(eval $(BIN_DIR):;@:)
  endif
endif

# If the first argument is "build_docs"...
ifeq (build_docs,$(firstword $(MAKECMDGOALS)))
  ifneq (,$(word 2, $(MAKECMDGOALS)))
    DOCS_DIR := $(word 2, $(MAKECMDGOALS))
    $(eval $(DOCS_DIR):;@:)
  endif
endif

default: help
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
	go mod tidy

build_docs:
	@mkdir -p $(DOCS_DIR)
	@echo "Building docs to $(DOCS_DIR)..."
	@go run scripts/doc_cover.go > $(DOCS_DIR)/doc_coverage.txt || echo "doc_cover.go failed"
	@echo "Docs built."

build:
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/cdd-go ./cmd/cdd-go

test:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

run: build
	./$(BIN_DIR)/cdd-go $(RUN_ARGS) $(ARGS)

build_wasm:
	@mkdir -p $(BIN_DIR)
	GOOS=js GOARCH=wasm go build -o $(BIN_DIR)/cdd-go.wasm ./cmd/cdd-go
build_docker:
	docker build -t cdd-go-alpine -f alpine.Dockerfile .
	docker build -t cdd-go-debian -f debian.Dockerfile .

run_docker:
	docker run -d -p 8085:8085 --name cdd-go-test cdd-go-alpine --port 8085 --listen 0.0.0.0
	sleep 2
	curl -X POST -H "Content-Type: application/json" -d "{\"method\":\"version\",\"id\":1}" http://127.0.0.1:8085
	docker stop cdd-go-test
	docker rm cdd-go-test
	docker rmi cdd-go-alpine cdd-go-debian
