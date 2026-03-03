.PHONY: all default help install_base install_deps build_docs build test run build_wasm build_docker run_docker

default: all

all: help

help:
	@echo "Available tasks:"
	@echo "  install_base   Install language runtime & tools"
	@echo "  install_deps   Install dependencies"
	@echo "  build_docs     Build the API docs (specify DOCS_DIR=... for alternative)"
	@echo "  build          Build the CLI binary (specify BIN_DIR=... for alternative)"
	@echo "  test           Run tests locally"
	@echo "  run            Run the CLI (builds if not built, pass args via ARGS=\"...\")"
	@echo "  build_wasm     Build WASM binary"
	@echo "  build_docker   Build Alpine and Debian Docker images"
	@echo "  run_docker     Run the Docker images"

install_base:
	@echo "Installing base tools..."
	go version || (echo "Please install Go 1.25+"; exit 1)

install_deps:
	go mod tidy
	go mod download

DOCS_DIR ?= docs
build_docs: build
	mkdir -p $(DOCS_DIR)
	./bin/cdd-go to_docs_json -i spec.json -o $(DOCS_DIR)/docs.json || true

BIN_DIR ?= bin
build: install_deps
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/cdd-go ./cmd/cdd-go

test:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

run: build
	./$(BIN_DIR)/cdd-go $(ARGS)

build_wasm:
	GOOS=js GOARCH=wasm go build -o $(BIN_DIR)/cdd-go.wasm ./cmd/cdd-go

build_docker:
	docker build -t cdd-go:alpine -f alpine.Dockerfile .
	docker build -t cdd-go:debian -f debian.Dockerfile .

run_docker:
	docker run --rm -p 8082:8082 cdd-go:alpine
