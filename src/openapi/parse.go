package openapi

import (
	"encoding/json"
	"io"

	"github.com/ghodss/yaml"
)

// Parse reads an OpenAPI description from the given reader and unmarshals it into an OpenAPI object.
// It supports both JSON and YAML formats.
func Parse(r io.Reader) (*OpenAPI, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	var oa OpenAPI

	// Try JSON first
	errJSON := json.Unmarshal(data, &oa)
	if errJSON == nil {
		if len(oa.Definitions) > 0 {
			if oa.Components == nil {
				oa.Components = &Components{}
			}
			if oa.Components.Schemas == nil {
				oa.Components.Schemas = make(map[string]Schema)
			}
			for k, v := range oa.Definitions {
				oa.Components.Schemas[k] = v
			}
			oa.Definitions = nil
		}
		return &oa, nil
	}

	// Try YAML if JSON fails
	errYAML := yaml.Unmarshal(data, &oa)
	if errYAML == nil {
		if len(oa.Definitions) > 0 {
			if oa.Components == nil {
				oa.Components = &Components{}
			}
			if oa.Components.Schemas == nil {
				oa.Components.Schemas = make(map[string]Schema)
			}
			for k, v := range oa.Definitions {
				oa.Components.Schemas[k] = v
			}
			oa.Definitions = nil
		}
		return &oa, nil
	}

	// If both fail, return JSON error as it's the default
	return nil, errJSON
}
