// Package openapi provides structures and utilities for parsing and emitting OpenAPI descriptions.
package openapi

import (
	"encoding/json"
	"io"
)

// Emit serializes the OpenAPI description into the given writer as JSON.
// Parameter w is the destination writer, oa is the OpenAPI specification, and it returns any serialization error.
func Emit(w io.Writer, oa *OpenAPI) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(oa)
}
