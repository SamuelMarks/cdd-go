package mcp

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestSchemaBaseStructsAndRoundTrip(t *testing.T) {
	ann := Annotated{Annotations: &Annotations{Audience: []string{"dev"}, Priority: 1.0}}
	b, err := json.Marshal(ann)
	if err != nil || !strings.Contains(string(b), `"audience":["dev"]`) {
		t.Fatalf("expected audience dev: %v", err)
	}
	var annOut Annotated
	if err := json.Unmarshal(b, &annOut); err != nil || !reflect.DeepEqual(ann, annOut) {
		t.Errorf("Annotated round-trip mismatch: %+v vs %+v (%v)", ann, annOut, err)
	}

	blob := BlobResourceContents{URI: "file://blob", MimeType: "application/octet-stream", Blob: "base64data"}
	blob.IsResourceContents()
	b, err = json.Marshal(blob)
	if err != nil || !strings.Contains(string(b), `"blob":"base64data"`) {
		t.Fatalf("expected blob data: %v", err)
	}

	textRes := TextResourceContents{URI: "file://text", MimeType: "text/plain", Text: "hello"}
	textRes.IsResourceContents()
	b, err = json.Marshal(textRes)
	if err != nil || !strings.Contains(string(b), `"text":"hello"`) {
		t.Fatalf("expected text hello: %v", err)
	}

	emb := EmbeddedResource{Type: "resource", Resource: textRes}
	emb.IsContent()
	b, err = json.Marshal(emb)
	if err != nil || !strings.Contains(string(b), `"type":"resource"`) {
		t.Fatalf("expected embedded resource: %v", err)
	}
	var embOut EmbeddedResource
	if err := json.Unmarshal(b, &embOut); err != nil || !reflect.DeepEqual(emb, embOut) {
		t.Errorf("EmbeddedResource round-trip mismatch: %+v vs %+v (%v)", emb, embOut, err)
	}
	if err := json.Unmarshal([]byte(`{invalid`), &embOut); err == nil {
		t.Errorf("expected error unmarshaling invalid EmbeddedResource JSON")
	}

	tc := TextContent{Type: "text", Text: "msg"}
	tc.IsContent()
	b, err = json.Marshal(tc)
	if err != nil || !strings.Contains(string(b), `"text":"msg"`) {
		t.Fatalf("expected text content: %v", err)
	}

	ic := ImageContent{Type: "image", Data: "imgdata", MimeType: "image/png"}
	ic.IsContent()
	b, err = json.Marshal(ic)
	if err != nil || !strings.Contains(string(b), `"mimeType":"image/png"`) {
		t.Fatalf("expected image content: %v", err)
	}

	// Test UnmarshalResourceContents & UnmarshalContent
	rcBlob, err := UnmarshalResourceContents([]byte(`{"uri":"file://b","blob":"aGVsbG8="}`))
	if err != nil || rcBlob.(BlobResourceContents).Blob != "aGVsbG8=" {
		t.Errorf("expected BlobResourceContents, got %v, %v", rcBlob, err)
	}
	if _, err := UnmarshalResourceContents([]byte(`{invalid`)); err == nil {
		t.Errorf("expected error for invalid json")
	}
	if _, err := UnmarshalResourceContents([]byte(`{"blob":123}`)); err == nil {
		t.Errorf("expected error for invalid blob type")
	}
	if _, err := UnmarshalResourceContents([]byte(`{"text":123}`)); err == nil {
		t.Errorf("expected error for invalid text type")
	}
	if _, err := UnmarshalResourceContents([]byte(`{"uri":"file://empty"}`)); err == nil {
		t.Errorf("expected error for missing blob/text")
	}

	cText, err := UnmarshalContent([]byte(`{"type":"text","text":"hi"}`))
	if err != nil || cText.(TextContent).Text != "hi" {
		t.Errorf("expected TextContent, got %v, %v", cText, err)
	}
	cImg, err := UnmarshalContent([]byte(`{"type":"image","data":"abc","mimeType":"image/png"}`))
	if err != nil || cImg.(ImageContent).MimeType != "image/png" {
		t.Errorf("expected ImageContent, got %v, %v", cImg, err)
	}
	cEmb, err := UnmarshalContent([]byte(`{"type":"resource","resource":{"uri":"file://t","text":"body"}}`))
	if err != nil || cEmb.(EmbeddedResource).Resource.(TextResourceContents).Text != "body" {
		t.Errorf("expected EmbeddedResource, got %v, %v", cEmb, err)
	}
	if _, err := UnmarshalContent([]byte(`{invalid`)); err == nil {
		t.Errorf("expected error for invalid content json")
	}
	if _, err := UnmarshalContent([]byte(`{"type":"text","text":123}`)); err == nil {
		t.Errorf("expected error for invalid text content")
	}
	if _, err := UnmarshalContent([]byte(`{"type":"image","data":123}`)); err == nil {
		t.Errorf("expected error for invalid image content")
	}
	if _, err := UnmarshalContent([]byte(`{"type":"resource","resource":{"uri":"x"}}`)); err == nil {
		t.Errorf("expected error for invalid embedded resource content")
	}
	if _, err := UnmarshalContent([]byte(`{"type":"unknown"}`)); err == nil {
		t.Errorf("expected error for unknown content type")
	}

	impl := Implementation{Name: "cdd-go", Version: "0.0.4"}
	b, err = json.Marshal(impl)
	if err != nil || !strings.Contains(string(b), `"name":"cdd-go"`) {
		t.Fatalf("expected implementation name: %v", err)
	}
	var implOut Implementation
	if err := json.Unmarshal(b, &implOut); err != nil || implOut != impl {
		t.Errorf("Implementation round-trip failed: %+v, %v", implOut, err)
	}

	if RoleUser != "user" || RoleAssistant != "assistant" {
		t.Errorf("expected correct roles")
	}

	er := EmptyResult{}
	b, _ = json.Marshal(er)
	if string(b) != "{}" {
		t.Errorf("expected empty object")
	}
}
