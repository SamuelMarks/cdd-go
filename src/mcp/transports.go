package mcp

import (
	"bufio"
	"io"
)

// Transport defines the interface for MCP message passing.
type Transport interface {
	Send(msg []byte) error
	Receive() ([]byte, error)
	Close() error
}

// StdioTransport implements Transport over standard input/output.
type StdioTransport struct {
	In      io.Reader
	Out     io.Writer
	Scanner *bufio.Scanner
}

// NewStdioTransport creates a new stdio transport.
func NewStdioTransport(in io.Reader, out io.Writer) *StdioTransport {
	return &StdioTransport{
		In:      in,
		Out:     out,
		Scanner: bufio.NewScanner(in),
	}
}

// Send writes a message to stdio.
func (t *StdioTransport) Send(msg []byte) error {
	_, err := t.Out.Write(append(msg, '\n'))
	return err
}

// Receive reads a message from stdio.
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

// SseTransport is a stub for Server-Sent Events transport.
type SseTransport struct {
	// Typically this would hold HTTP response writers and flusher channels.
}

// Send writes a message to sse.
func (t *SseTransport) Send(msg []byte) error {
	// Stub: In reality this formats "data: <msg>\n\n"
	return nil
}

// Receive reads a message from sse.
func (t *SseTransport) Receive() ([]byte, error) {
	// Stub: In reality this reads from an HTTP POST to /mcp/message
	return nil, io.EOF
}

// Close closes the transport.
func (t *SseTransport) Close() error {
	return nil
}
