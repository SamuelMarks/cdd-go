package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/SamuelMarks/cdd-go/src/mcp"
)

func TestMCPServer(t *testing.T) {
	s := NewMCPServer()

	// Test HandleSSE
	reqSSE, _ := http.NewRequest("GET", "/mcp/sse", nil)
	rrSSE := httptest.NewRecorder()
	s.HandleSSE(rrSSE, reqSSE)

	if rrSSE.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %v", rrSSE.Code)
	}
	if !strings.Contains(rrSSE.Body.String(), "data: /mcp/message") {
		t.Errorf("Expected SSE endpoint data")
	}

	// Test HandleMessage without handler
	msg := mcp.Message{JSONRPC: "2.0"}
	b, _ := json.Marshal(msg)
	reqMsg, _ := http.NewRequest("POST", "/mcp/message", bytes.NewBuffer(b))
	rrMsg := httptest.NewRecorder()
	s.HandleMessage(rrMsg, reqMsg)
	if rrMsg.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %v", rrMsg.Code)
	}

	// Test HandleMessage with Handler
	s.Handler = func(req *mcp.Message) (*mcp.Response, error) {
		return &mcp.Response{JSONRPC: "2.0", ID: req.ID, Result: "success"}, nil
	}
	reqMsg, _ = http.NewRequest("POST", "/mcp/message", bytes.NewBuffer(b))
	rrMsg = httptest.NewRecorder()
	s.HandleMessage(rrMsg, reqMsg)
	if !strings.Contains(rrMsg.Body.String(), `"success"`) {
		t.Errorf("Expected success result")
	}

	// Test HandleMessage error decode
	reqMsgBad, _ := http.NewRequest("POST", "/mcp/message", bytes.NewBuffer([]byte(`{bad}`)))
	rrMsgBad := httptest.NewRecorder()
	s.HandleMessage(rrMsgBad, reqMsgBad)
	if rrMsgBad.Code != http.StatusBadRequest {
		t.Errorf("Expected bad request")
	}

	// Test HandleMessage handler error
	s.Handler = func(req *mcp.Message) (*mcp.Response, error) {
		return nil, fmt.Errorf("internal error")
	}
	reqMsg, _ = http.NewRequest("POST", "/mcp/message", bytes.NewBuffer(b))
	rrMsg = httptest.NewRecorder()
	s.HandleMessage(rrMsg, reqMsg)
	if rrMsg.Code != http.StatusInternalServerError {
		t.Errorf("Expected internal server error")
	}
}
