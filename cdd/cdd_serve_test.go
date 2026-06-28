package cdd

import (
	"testing"
)

func TestServeJSONRPC(t *testing.T) {
	err := ServeJSONRPC(8080, "localhost")
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
}
