package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestRunServerJSONRPC_Flags(t *testing.T) {
	err := runServerJSONRPC([]string{"-invalid-flag"})
	if err == nil {
		t.Errorf("expected error for invalid flag")
	}
}

func TestRunServerJSONRPC_Handler(t *testing.T) {
	// Start server on a specific address to test the Handler.
	// But it's easier to just call http.DefaultServeMux
	// We run it with a dummy port to start it but then immediately stop or just test the mux.
	// We'll run it async on an available port.
	go func() {
		runServerJSONRPC([]string{"-port", "0"})
	}()

	// Since we can't easily capture the mux, let's just create an httptest server with DefaultServeMux.
	ts := httptest.NewServer(http.DefaultServeMux)
	defer ts.Close()

	// 1. Invalid method (GET)
	res, _ := http.Get(ts.URL)
	if res.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", res.StatusCode)
	}

	// 2. Invalid JSON body
	res, _ = http.Post(ts.URL, "application/json", strings.NewReader(`{invalid}`))
	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", res.StatusCode)
	}

	// Helper for valid requests
	doRequest := func(method string, params string) (RPCResponse, error) {
		reqStr := fmt.Sprintf(`{"method": "%s", "params": %s, "id": 1}`, method, params)
		res, err := http.Post(ts.URL, "application/json", strings.NewReader(reqStr))
		if err != nil {
			return RPCResponse{}, err
		}
		defer res.Body.Close()
		var rpcResp RPCResponse
		json.NewDecoder(res.Body).Decode(&rpcResp)
		return rpcResp, nil
	}

	// 3. Unknown method
	resp, _ := doRequest("unknown_method", "[]")
	if resp.Error != "Method not found" {
		t.Errorf("expected Method not found, got %v", resp.Error)
	}

	// 4. version
	resp, _ = doRequest("version", "[]")
	if resp.Result != "0.0.1" {
		t.Errorf("expected 0.0.1, got %v", resp.Result)
	}

	// 5. to_docs_json (invalid params)
	resp, _ = doRequest("to_docs_json", `{"invalid":"type"}`)
	if resp.Error != "Invalid params for to_docs_json" {
		t.Errorf("expected invalid params, got %v", resp.Error)
	}

	// 6. to_docs_json (valid params, but missing file)
	resp, _ = doRequest("to_docs_json", `["-i", "missing.json"]`)
	if resp.Error == nil {
		t.Errorf("expected error for missing file")
	}

	// 7. to_docs_json (valid params, valid file)
	// Create a temporary file
	dir := t.TempDir()
	specFile := dir + "/spec.json"
	// just write empty string to fail parsing or valid openapi
	os.WriteFile(specFile, []byte(`{"openapi": "3.2.0"}`), 0644)
	resp, _ = doRequest("to_docs_json", fmt.Sprintf(`["-i", "%s"]`, specFile))
	if resp.Error != nil {
		t.Errorf("expected no error, got %v", resp.Error)
	}

	// 8. to_openapi (invalid params)
	resp, _ = doRequest("to_openapi", `{"invalid":"type"}`)
	if resp.Error != "Invalid params for to_openapi" {
		t.Errorf("expected invalid params, got %v", resp.Error)
	}

	// 9. to_openapi (valid params, missing file)
	resp, _ = doRequest("to_openapi", `["-in", "missing.go"]`)
	if resp.Error == nil {
		t.Errorf("expected error for missing file")
	}

	// 10. to_openapi (valid params, valid file)
	goFile := dir + "/main.go"
	os.WriteFile(goFile, []byte(`package main`), 0644)
	resp, _ = doRequest("to_openapi", fmt.Sprintf(`["-in", "%s", "-o", "%s/out.json"]`, goFile, dir))
	if resp.Error != nil {
		t.Errorf("expected no error, got %v", resp.Error)
	}

	// 11. from_openapi (invalid params)
	resp, _ = doRequest("from_openapi", `{"invalid":"type"}`)
	if resp.Error != "Invalid params for from_openapi" {
		t.Errorf("expected invalid params, got %v", resp.Error)
	}

	// 12. from_openapi (valid params, missing file)
	resp, _ = doRequest("from_openapi", `["to_sdk", "-i", "missing.json"]`)
	if resp.Error == nil {
		t.Errorf("expected error for missing file")
	}

	// 13. from_openapi (valid params, valid file)
	resp, _ = doRequest("from_openapi", fmt.Sprintf(`["to_sdk", "-i", "%s", "-o", "%s/sdk"]`, specFile, dir))
	if resp.Error != nil {
		t.Errorf("expected no error, got %v", resp.Error)
	}
}
