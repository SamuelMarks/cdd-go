@echo off
setlocal enabledelayedexpansion

if "%1" == "" goto help
if "%1" == "help" goto help
if "%1" == "all" goto help
if "%1" == "install_base" goto install_base
if "%1" == "install_deps" goto install_deps
if "%1" == "build_docs" goto build_docs
if "%1" == "build" goto build
if "%1" == "test" goto test
if "%1" == "run" goto run
if "%1" == "build_wasm" goto build_wasm
if "%1" == "build_docker" goto build_docker
if "%1" == "run_docker" goto run_docker

:help
echo Available tasks:
echo   install_base   Install language runtime & tools
echo   install_deps   Install dependencies
echo   build_docs     Build the API docs
echo   build          Build the CLI binary
echo   test           Run tests locally
echo   run            Run the CLI
echo   build_wasm     Build WASM binary
echo   build_docker   Build Alpine and Debian Docker images
echo   run_docker     Run the Docker images
goto end

:install_base
echo Installing base tools...
go version
if %errorlevel% neq 0 (
    echo Please install Go 1.25+
    exit /b 1
)
goto end

:install_deps
go mod tidy
go mod download
goto end

:build_docs
if not exist "docs" mkdir "docs"
call :build
bin\cdd-go.exe to_docs_json -i spec.json -o docs\docs.json
goto end

:build
call :install_deps
if not exist "bin" mkdir "bin"
go build -o bin\cdd-go.exe .\cmd\cdd-go
goto end

:test
go test -v -coverprofile=coverage.out .\...
go tool cover -func=coverage.out
goto end

:run
call :build
bin\cdd-go.exe %2 %3 %4 %5 %6 %7 %8 %9
goto end

:build_wasm
set GOOS=js
set GOARCH=wasm
go build -o bin\cdd-go.wasm .\cmd\cdd-go
goto end

:build_docker
docker build -t cdd-go:alpine -f alpine.Dockerfile .
docker build -t cdd-go:debian -f debian.Dockerfile .
goto end

:run_docker
docker run --rm -p 8082:8082 cdd-go:alpine
goto end

:end
