@ECHO OFF
SETLOCAL ENABLEDELAYEDEXPANSION

IF "%1" == "" GOTO help
IF "%1" == "help" GOTO help
IF "%1" == "all" GOTO help
IF "%1" == "install_base" GOTO install_base
IF "%1" == "install_deps" GOTO install_deps
IF "%1" == "build_docs" GOTO build_docs
IF "%1" == "build" GOTO build
IF "%1" == "build_wasm" GOTO build_wasm
IF "%1" == "test" GOTO test
IF "%1" == "run" GOTO run

ECHO Unknown target: %1
GOTO help

:help
ECHO Available targets:
ECHO   install_base : install Go runtime (assumes 'go' is already in PATH)
ECHO   install_deps : install local dependencies (go mod download)
ECHO   build_docs   : build the API docs and put them in the docs directory. Usage: make.bat build_docs [DOCS_DIR]
ECHO   build        : build the CLI binary. Usage: make.bat build [BIN_DIR]
ECHO   build_wasm   : build the WASM binary. Usage: make.bat build_wasm [BIN_DIR]
ECHO   test         : run tests locally
ECHO   run          : run the CLI. Usage: make.bat run [ARGS]
ECHO   help         : show this help text
ECHO   all          : show this help text
GOTO :EOF

:install_base
go version >nul 2>&1
IF %ERRORLEVEL% NEQ 0 (
    ECHO Go is not installed. Please install Go 1.21+.
    EXIT /B 1
)
ECHO Go is installed.
GOTO :EOF

:install_deps
go mod download
GOTO :EOF

:build_docs
SET DOCS_DIR=%2
IF "%DOCS_DIR%"=="" SET DOCS_DIR=docs
IF NOT EXIST "%DOCS_DIR%" MKDIR "%DOCS_DIR%"
ECHO Building docs to %DOCS_DIR%...
go run scripts\doc_cover.go > "%DOCS_DIR%\doc_coverage.txt" 2>nul
ECHO Docs built.
GOTO :EOF

:build
SET BIN_DIR=%2
IF "%BIN_DIR%"=="" SET BIN_DIR=bin
IF NOT EXIST "%BIN_DIR%" MKDIR "%BIN_DIR%"
go build -o "%BIN_DIR%\cdd-go.exe" .\cmd\cdd-go
GOTO :EOF

:test
go test -v -cover .\...
GOTO :EOF

:run
SET BIN_DIR=%2
IF "%BIN_DIR%"=="" SET BIN_DIR=bin
IF NOT EXIST "%BIN_DIR%\cdd-go.exe" (
    CALL :build %BIN_DIR%
)
REM Shift args to pass the rest to the command
SHIFT
SHIFT
SET CMD_ARGS=
:loop
IF "%1"=="" GOTO end_loop
SET CMD_ARGS=!CMD_ARGS! %1
SHIFT
GOTO loop
:end_loop
"%BIN_DIR%\cdd-go.exe" %CMD_ARGS%
GOTO :EOF

:build_wasm
SET BIN_DIR=%2
IF "%BIN_DIR%"=="" SET BIN_DIR=bin
IF NOT EXIST "%BIN_DIR%" MKDIR "%BIN_DIR%"
SET GOOS=js
SET GOARCH=wasm
go build -o "%BIN_DIR%\cdd-go.wasm" .\cmd\cdd-go
SET GOOS=
SET GOARCH=
ECHO Built WASM to %BIN_DIR%\cdd-go.wasm
GOTO :EOF

:build_docker
docker build -t cdd-go-alpine -f alpine.Dockerfile .
docker build -t cdd-go-debian -f debian.Dockerfile .
GOTO :EOF

:run_docker
docker run -d -p 8085:8085 --name cdd-go-test cdd-go-alpine --port 8085 --listen 0.0.0.0
timeout /t 2
curl -X POST -H "Content-Type: application/json" -d "{\"method\":\"version\",\"id\":1}" http://127.0.0.1:8085
docker stop cdd-go-test
docker rm cdd-go-test
docker rmi cdd-go-alpine cdd-go-debian
GOTO :EOF
