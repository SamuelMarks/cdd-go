package cdd

import (
	"testing"
)

func TestServeJSONRPC(t *testing.T) {
	err := ServeJSONRPC(8080, "localhost")
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
	if err := ServeJSONRPC(0, "localhost"); err == nil {
		t.Errorf("Expected error for invalid port 0")
	}
	if err := ServeJSONRPC(8080, ""); err == nil {
		t.Errorf("Expected error for empty listen address")
	}
}
