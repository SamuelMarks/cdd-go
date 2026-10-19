package cdd

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestServeJSONRPC(t *testing.T) {
	oldListen := cddHTTPListenAndServe
	defer func() { cddHTTPListenAndServe = oldListen }()

	var capturedAddr string
	var capturedHandler http.Handler
	cddHTTPListenAndServe = func(addr string, handler http.Handler) error {
		capturedAddr = addr
		capturedHandler = handler
		return nil
	}

	if err := ServeJSONRPC(8080, "localhost"); err != nil {
		t.Fatalf("Expected nil error, got %v", err)
	}
	if capturedAddr != "localhost:8080" || capturedHandler == nil {
		t.Fatalf("Expected localhost:8080 and non-nil handler, got %q, %v", capturedAddr, capturedHandler)
	}

	if err := ServeJSONRPC(0, "localhost"); err == nil {
		t.Errorf("Expected error for invalid port 0")
	}
	if err := ServeJSONRPC(8080, ""); err == nil {
		t.Errorf("Expected error for empty listen address")
	}

	// Test the HTTP JSON-RPC handler end-to-end
	reqGet := httptest.NewRequest(http.MethodGet, "/", nil)
	rrGet := httptest.NewRecorder()
	capturedHandler.ServeHTTP(rrGet, reqGet)
	if rrGet.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected 405 for GET, got %d", rrGet.Code)
	}

	reqBad := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("{bad"))
	rrBad := httptest.NewRecorder()
	capturedHandler.ServeHTTP(rrBad, reqBad)
	if rrBad.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 for invalid JSON, got %d", rrBad.Code)
	}

	reqVer := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"jsonrpc":"2.0","id":1,"method":"version"}`))
	rrVer := httptest.NewRecorder()
	capturedHandler.ServeHTTP(rrVer, reqVer)
	if !strings.Contains(rrVer.Body.String(), `"0.0.4"`) {
		t.Errorf("Expected version 0.0.4, got %s", rrVer.Body.String())
	}

	reqBadParams := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"jsonrpc":"2.0","id":2,"method":"cdd_generate_sdk","params":123}`))
	rrBadParams := httptest.NewRecorder()
	capturedHandler.ServeHTTP(rrBadParams, reqBadParams)
	if !strings.Contains(rrBadParams.Body.String(), `"Invalid params"`) {
		t.Errorf("Expected Invalid params error, got %s", rrBadParams.Body.String())
	}

	reqUnknown := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"jsonrpc":"2.0","id":3,"method":"unknown_tool","params":{}}`))
	rrUnknown := httptest.NewRecorder()
	capturedHandler.ServeHTTP(rrUnknown, reqUnknown)
	if !strings.Contains(rrUnknown.Body.String(), "unknown generator tool") {
		t.Errorf("Expected unknown generator tool error, got %s", rrUnknown.Body.String())
	}

	tmpDir := t.TempDir()
	inPath := filepath.Join(tmpDir, "spec.json")
	outDir := filepath.Join(tmpDir, "sdk")
	os.WriteFile(inPath, []byte(`{"openapi":"3.2.0","info":{"title":"T","version":"1.0"}}`), 0644)

	reqGen := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"jsonrpc":"2.0","id":4,"method":"cdd_generate_sdk","params":{"input":"`+inPath+`","output":"`+outDir+`"}}`))
	rrGen := httptest.NewRecorder()
	capturedHandler.ServeHTTP(rrGen, reqGen)
	if !strings.Contains(rrGen.Body.String(), `"SDK Generated"`) {
		t.Errorf("Expected SDK Generated result, got %s", rrGen.Body.String())
	}
}
