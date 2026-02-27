// Package openapi provides structures and utilities for parsing and emitting OpenAPI descriptions.
package openapi

import (
	"encoding/json"
	"io"
)

// Parse reads an OpenAPI description from the given reader and unmarshals it into an OpenAPI object.
func Parse(r io.Reader) (*OpenAPI, error) {
	var oa OpenAPI
	decoder := json.NewDecoder(r)
	if err := decoder.Decode(&oa); err != nil {
		return nil, err
	}
	return &oa, nil
}
