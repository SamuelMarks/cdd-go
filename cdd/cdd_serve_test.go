package cdd

import (
	"testing"
)

func TestServeJsonRpc(t *testing.T) {
	err := ServeJsonRpc(8080, "localhost")
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
}
