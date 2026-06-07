package mcp

import (
	"errors"
)

type errReader struct{}

func (errReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("mock read error")
}
