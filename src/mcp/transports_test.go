package mcp

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestStdioTransport(t *testing.T) {
	inBuf := bytes.NewBufferString("test message" + string(rune(10)))
	var outBuf bytes.Buffer

	transport := NewStdioTransport(inBuf, &outBuf)

	if err := transport.Send([]byte("response message")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.TrimSpace(outBuf.String()) != "response message" {
		t.Errorf("expected response message, got %q", outBuf.String())
	}

	msg, err := transport.Receive()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(msg) != "test message" {
		t.Errorf("expected 'test message', got %q", string(msg))
	}

	// Next read should be EOF
	_, err = transport.Receive()
	if err != io.EOF {
		t.Errorf("expected EOF, got %v", err)
	}

	if err := transport.Close(); err != nil {
		t.Errorf("unexpected error on close: %v", err)
	}
}

func TestSseTransport(t *testing.T) {
	var outBuf bytes.Buffer
	transport := NewSseTransport(&outBuf)

	if err := transport.Send([]byte(`{"jsonrpc":"2.0"}`)); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !strings.Contains(outBuf.String(), `data: {"jsonrpc":"2.0"}`) {
		t.Errorf("expected SSE frame in outBuf, got %q", outBuf.String())
	}

	// Enqueue and Receive
	if err := transport.Enqueue([]byte(`{"id":1}`)); err != nil {
		t.Fatalf("unexpected enqueue error: %v", err)
	}
	msg, err := transport.Receive()
	if err != nil || string(msg) != `{"id":1}` {
		t.Errorf("expected queued message, got %q, %v", string(msg), err)
	}

	// Empty queue returns EOF
	_, err = transport.Receive()
	if err != io.EOF {
		t.Errorf("expected EOF on empty queue, got %v", err)
	}

	// Nil writer Send
	nilOutTransport := &SseTransport{}
	if err := nilOutTransport.Send([]byte("noop")); err != nil {
		t.Errorf("unexpected error on nil Out Send: %v", err)
	}

	if err := transport.Close(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if err := transport.Enqueue([]byte("after-close")); err != io.ErrClosedPipe {
		t.Errorf("expected ErrClosedPipe on Enqueue after Close, got %v", err)
	}
	if err := transport.Send([]byte("after-close")); err != io.ErrClosedPipe {
		t.Errorf("expected ErrClosedPipe on Send after Close, got %v", err)
	}
}

func TestStdioTransportScannerErr(t *testing.T) {
	transport := NewStdioTransport(errReader{}, nil)
	_, err := transport.Receive()
	if err == nil || err.Error() != "mock read error" {
		t.Errorf("expected mock read error, got %v", err)
	}
}
