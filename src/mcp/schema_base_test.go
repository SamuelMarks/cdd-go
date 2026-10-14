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
	if !strings.Contains(string(b), `"audience":["dev"]`) {
		t.Errorf("expected audience in annotated")
	}

	// Test BlobResourceContents & TextResourceContents + UnmarshalResourceContents
	blob := BlobResourceContents{URI: "file://blob", MimeType: "application/octet-stream", Blob: "aGVsbG8="}
	blob.IsResourceContents()
	b, _ = json.Marshal(blob)
	rcBlob, err := UnmarshalResourceContents(b)
	if err != nil || rcBlob.(BlobResourceContents).Blob != "aGVsbG8=" {
		t.Errorf("expected round-trip blob resource contents, got %v, %v", rcBlob, err)
	}

	textRes := TextResourceContents{URI: "file://text", MimeType: "text/plain", Text: "hello"}
	textRes.IsResourceContents()
	b, _ = json.Marshal(textRes)
	rcText, err := UnmarshalResourceContents(b)
	if err != nil || rcText.(TextResourceContents).Text != "hello" {
		t.Errorf("expected round-trip text resource contents, got %v, %v", rcText, err)
	}

	if _, err := UnmarshalResourceContents([]byte(`{invalid`)); err == nil {
		t.Errorf("expected error on invalid JSON")
	}
	if _, err := UnmarshalResourceContents([]byte(`{"blob":123}`)); err == nil {
		t.Errorf("expected error on invalid blob field type")
	}
	if _, err := UnmarshalResourceContents([]byte(`{"text":123}`)); err == nil {
		t.Errorf("expected error on invalid text field type")
	}
	if _, err := UnmarshalResourceContents([]byte(`{"uri":"file://unknown"}`)); err == nil {
		t.Errorf("expected error on unknown resource contents")
	}

	// Test EmbeddedResource, TextContent, ImageContent + UnmarshalContent
	emb := EmbeddedResource{Type: "resource", Resource: textRes}
	emb.IsContent()
	b, _ = json.Marshal(emb)
	cEmb, err := UnmarshalContent(b)
	if err != nil || cEmb.(EmbeddedResource).Resource.(TextResourceContents).Text != "hello" {
		t.Errorf("expected round-trip embedded resource content, got %v, %v", cEmb, err)
	}

	tc := TextContent{Type: "text", Text: "some text"}
	tc.IsContent()
	b, _ = json.Marshal(tc)
	cText, err := UnmarshalContent(b)
	if err != nil || cText.(TextContent).Text != "some text" {
		t.Errorf("expected round-trip text content, got %v, %v", cText, err)
	}

	ic := ImageContent{Type: "image", Data: "aGVsbG8=", MimeType: "image/png"}
	ic.IsContent()
	b, _ = json.Marshal(ic)
	cImg, err := UnmarshalContent(b)
	if err != nil || cImg.(ImageContent).MimeType != "image/png" {
		t.Errorf("expected round-trip image content, got %v, %v", cImg, err)
	}

	if _, err := UnmarshalContent([]byte(`{invalid`)); err == nil {
		t.Errorf("expected error on invalid content JSON")
	}
	if _, err := UnmarshalContent([]byte(`{"type":"text","text":123}`)); err == nil {
		t.Errorf("expected error on invalid text content")
	}
	if _, err := UnmarshalContent([]byte(`{"type":"image","data":123}`)); err == nil {
		t.Errorf("expected error on invalid image content")
	}
	if _, err := UnmarshalContent([]byte(`{"type":"resource","annotations":"bad"}`)); err == nil {
		t.Errorf("expected error on invalid embedded resource wrapper")
	}
	if _, err := UnmarshalContent([]byte(`{"type":"resource","resource":{}}`)); err == nil {
		t.Errorf("expected error on invalid inner resource")
	}
	if _, err := UnmarshalContent([]byte(`{"type":"unknown"}`)); err == nil {
		t.Errorf("expected error on unsupported content type")
	}

	// Test Implementation & Role
	impl := Implementation{Name: "cdd-go", Version: "1.0"}
	b, _ = json.Marshal(impl)
	if !strings.Contains(string(b), `"cdd-go"`) {
		t.Errorf("expected implementation name")
	}
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
