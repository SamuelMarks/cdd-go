// Package arazzo provides strongly-typed structures, parsers, validators, and emitters for the Arazzo Specification v1.1.0.
package arazzo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/ghodss/yaml"
)

// mergeExtensions merges a map of specification extensions into a raw JSON object map.
// Parameter dst is the destination JSON object map, and ext is the map of "^x-" extensions.
func mergeExtensions(dst map[string]json.RawMessage, ext map[string]json.RawMessage) {
	for k, v := range ext {
		dst[k] = v
	}
}

// MarshalDocument serializes an Arazzo 1.1.0 document into formatted JSON bytes, preserving all specification extensions.
// Parameter doc is the Arazzo specification to serialize, and it returns the indented JSON byte slice and any error.
func MarshalDocument(doc *Arazzo) ([]byte, error) {
	if doc == nil {
		return nil, fmt.Errorf("arazzo document is nil")
	}
	baseBytes, err := json.Marshal(doc)
	if err != nil {
		return nil, err
	}

	if len(doc.Extensions) > 0 || len(doc.Info.Extensions) > 0 {
		var rawRoot map[string]json.RawMessage
		_ = json.Unmarshal(baseBytes, &rawRoot)
		mergeExtensions(rawRoot, doc.Extensions)
		if len(doc.Info.Extensions) > 0 {
			var rawInfo map[string]json.RawMessage
			_ = json.Unmarshal(rawRoot["info"], &rawInfo)
			mergeExtensions(rawInfo, doc.Info.Extensions)
			infoB, _ := json.Marshal(rawInfo)
			rawRoot["info"] = infoB
		}
		merged, errMerge := json.Marshal(rawRoot)
		if errMerge != nil {
			return nil, errMerge
		}
		baseBytes = merged
	}

	var out bytes.Buffer
	_ = json.Indent(&out, baseBytes, "", "  ")
	out.WriteByte(10)
	return out.Bytes(), nil
}

// Emit writes an Arazzo 1.1.0 document as formatted JSON to the given io.Writer.
// Parameter w is the destination writer, and doc is the Arazzo document to emit.
func Emit(w io.Writer, doc *Arazzo) error {
	b, err := MarshalDocument(doc)
	if err != nil {
		return err
	}
	_, err = w.Write(b)
	return err
}

// EmitYAML writes an Arazzo 1.1.0 document as formatted YAML to the given io.Writer.
// Parameter w is the destination writer, and doc is the Arazzo document to emit.
func EmitYAML(w io.Writer, doc *Arazzo) error {
	b, err := MarshalDocument(doc)
	if err != nil {
		return err
	}
	yamlBytes, _ := yaml.JSONToYAML(b)
	_, err = w.Write(yamlBytes)
	return err
}
