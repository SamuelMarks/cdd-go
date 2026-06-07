package mcp

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSchemaBaseStructs(t *testing.T) {
	// Test Annotated
	ann := Annotated{Annotations: &Annotations{Audience: []string{"dev"}, Priority: 1.0}}
	b, _ := json.Marshal(ann)
	if !strings.Contains(string(b), `"audience":["dev"]`) || !strings.Contains(string(b), `"priority":1`) {
		t.Errorf("expected annotations")
	}

	// Test TextContent
	txt := TextContent{Type: "text", Text: "Hello"}
	txt.Annotations = &Annotations{Priority: 0.5}
	b, _ = json.Marshal(txt)
	if !strings.Contains(string(b), `"text":"Hello"`) {
		t.Errorf("expected text content")
	}

	// Test ImageContent
	img := ImageContent{Type: "image", Data: "base64", MimeType: "image/png"}
	b, _ = json.Marshal(img)
	if !strings.Contains(string(b), `"image/png"`) {
		t.Errorf("expected image mime type")
	}

	// Test Implementation
	impl := Implementation{Name: "test", Version: "1.0"}
	b, _ = json.Marshal(impl)
	if !strings.Contains(string(b), `"test"`) {
		t.Errorf("expected implementation name")
	}

	// Test EmbeddedResource
	res := EmbeddedResource{Type: "resource", Resource: TextResourceContents{URI: "file://test", Text: "data"}}
	b, _ = json.Marshal(res)
	if !strings.Contains(string(b), `"file://test"`) {
		t.Errorf("expected embedded resource uri")
	}

	// Test BlobResourceContents
	blob := BlobResourceContents{URI: "file://blob", Blob: "YWJj"}
	b, _ = json.Marshal(blob)
	if !strings.Contains(string(b), `"YWJj"`) {
		t.Errorf("expected blob contents")
	}

	// Test Roles
	if RoleUser != "user" || RoleAssistant != "assistant" {
		t.Errorf("expected correct roles")
	}
}
func TestEmptyResult(t *testing.T) {
	er := EmptyResult{}
	b, _ := json.Marshal(er)
	if string(b) != "{}" {
		t.Errorf("expected empty object")
	}
}
