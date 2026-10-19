// Package mcp provides strongly-typed Model Context Protocol (MCP) 1.0.0 and JSON-RPC 2.0 schemas, transports, and handlers.
package mcp

import (
	"bufio"
	"fmt"
	"io"
	"sync"
)

// Transport defines the interface for MCP message passing.
type Transport interface {
	// Send writes a serialized JSON-RPC message to the underlying transport.
	Send(msg []byte) error
	// Receive reads the next serialized JSON-RPC message from the underlying transport.
	Receive() ([]byte, error)
	// Close terminates the transport and releases any associated resources.
	Close() error
}

// StdioTransport implements Transport over standard input/output.
type StdioTransport struct {
	// In is the underlying input reader.
	In io.Reader
	// Out is the underlying output writer.
	Out io.Writer
	// Scanner reads newline-delimited messages from In.
	Scanner *bufio.Scanner
}

// NewStdioTransport creates a new stdio transport.
// Parameter in is the input stream, out is the output stream, and it returns the initialized StdioTransport.
func NewStdioTransport(in io.Reader, out io.Writer) *StdioTransport {
	return &StdioTransport{
		In:      in,
		Out:     out,
		Scanner: bufio.NewScanner(in),
	}
}

// Send writes a message to stdio.
// Parameter msg is the JSON-RPC payload bytes, and it returns any write error.
func (t *StdioTransport) Send(msg []byte) error {
	_, err := t.Out.Write(append(msg, byte(10)))
	return err
}

// Receive reads a message from stdio.
// It returns the next line bytes or io.EOF when the input stream ends.
func (t *StdioTransport) Receive() ([]byte, error) {
	if t.Scanner.Scan() {
		return t.Scanner.Bytes(), nil
	}
	if err := t.Scanner.Err(); err != nil {
		return nil, err
	}
	return nil, io.EOF
}

// Close closes the transport.
func (t *StdioTransport) Close() error {
	return nil
}

// SseTransport implements a bidirectional Server-Sent Events (SSE) and HTTP POST message queue transport.
type SseTransport struct {
	// Out is the optional SSE writer where server-sent event frames are written.
	Out io.Writer
	// mu protects closed state and queue operations.
	mu sync.Mutex
	// queue buffers incoming messages posted to the SSE message endpoint.
	queue [][]byte
	// closed indicates whether Close has been called.
	closed bool
}

// NewSseTransport creates a new SseTransport writing SSE frames to out.
// Parameter out is the optional writer for outbound SSE frames, and it returns the initialized SseTransport.
func NewSseTransport(out io.Writer) *SseTransport {
	return &SseTransport{
		Out: out,
	}
}

// Enqueue pushes an inbound client message (e.g., from an HTTP POST to /mcp/message) into the transport receive queue.
// Parameter msg is the raw JSON-RPC message bytes, and it returns an error if the transport is closed.
func (t *SseTransport) Enqueue(msg []byte) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.closed {
		return io.ErrClosedPipe
	}
	cp := make([]byte, len(msg))
	copy(cp, msg)
	t.queue = append(t.queue, cp)
	return nil
}

// Send formats and writes a message as an SSE data frame if Out is configured, or queues it for local loopback.
// Parameter msg is the JSON-RPC message bytes, and it returns any write error.
func (t *SseTransport) Send(msg []byte) error {
	t.mu.Lock()
	if t.closed {
		t.mu.Unlock()
		return io.ErrClosedPipe
	}
	out := t.Out
	t.mu.Unlock()

	if out != nil {
		_, err := fmt.Fprintf(out, `data: %s`+string([]byte{10, 10}), string(msg))
		return err
	}
	return nil
}

// Receive reads the next queued message from the SSE transport or returns io.EOF when empty or closed.
func (t *SseTransport) Receive() ([]byte, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if len(t.queue) > 0 {
		msg := t.queue[0]
		t.queue = t.queue[1:]
		return msg, nil
	}
	return nil, io.EOF
}

// Close marks the SSE transport as closed and clears any unread messages.
func (t *SseTransport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.closed = true
	t.queue = nil
	return nil
}
