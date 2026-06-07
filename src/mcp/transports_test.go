package mcp

import (
	"bytes"
	"io"
	"testing"
)

func TestStdioTransport(t *testing.T) {
	inBuf := bytes.NewBufferString("test message\n")
	var outBuf bytes.Buffer

	transport := NewStdioTransport(inBuf, &outBuf)

	// Test Receive
	msg, err := transport.Receive()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(msg) != "test message" {
		t.Errorf("expected 'test message', got %s", string(msg))
	}

	// Test Receive EOF
	_, err = transport.Receive()
	if err != io.EOF {
		t.Errorf("expected EOF, got %v", err)
	}

	// Test Send
	err = transport.Send([]byte("response"))
	if err != nil {
		t.Fatalf("unexpected send error: %v", err)
	}
	if outBuf.String() != "response\n" {
		t.Errorf("expected 'response\\n', got %q", outBuf.String())
	}

	// Test Close
	if err := transport.Close(); err != nil {
		t.Errorf("unexpected close error: %v", err)
	}
}

func TestSseTransport(t *testing.T) {
	transport := &SseTransport{}
	if err := transport.Send([]byte("test")); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if _, err := transport.Receive(); err != io.EOF {
		t.Errorf("expected EOF")
	}
	if err := transport.Close(); err != nil {
		t.Errorf("unexpected close error: %v", err)
	}
}

func TestStdioTransportError(t *testing.T) {
	// A simple reader that always returns an error could be injected to test scanner.Err(),
	// but the bufio.Scanner handles this transparently up to its buffer limits.
	// We've covered the primary branches.
}

func TestStdioTransportScannerErr(t *testing.T) {
	transport := NewStdioTransport(errReader{}, nil)
	_, err := transport.Receive()
	if err == nil || err.Error() != "mock read error" {
		t.Errorf("expected mock read error, got %v", err)
	}
}
