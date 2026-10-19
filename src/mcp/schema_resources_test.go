package mcp

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestSchemaResourcesStructsAndRoundTrip(t *testing.T) {
	res := Resource{URI: "file://test", Name: "TestRes", Description: "desc", MimeType: "text/plain", Size: 42}
	b, err := json.Marshal(res)
	if err != nil || !strings.Contains(string(b), `"file://test"`) || !strings.Contains(string(b), `"TestRes"`) {
		t.Fatalf("expected resource fields: %v", err)
	}
	var resOut Resource
	if err := json.Unmarshal(b, &resOut); err != nil || !reflect.DeepEqual(res, resOut) {
		t.Errorf("Resource round-trip mismatch: %+v vs %+v (%v)", res, resOut, err)
	}

	tmpl := ResourceTemplate{URITemplate: "file://{path}", Name: "Tmpl"}
	b, err = json.Marshal(tmpl)
	if err != nil || !strings.Contains(string(b), `"file://{path}"`) {
		t.Fatalf("expected resource template: %v", err)
	}
	var tmplOut ResourceTemplate
	if err := json.Unmarshal(b, &tmplOut); err != nil || !reflect.DeepEqual(tmpl, tmplOut) {
		t.Errorf("ResourceTemplate round-trip mismatch: %+v vs %+v (%v)", tmpl, tmplOut, err)
	}

	ref := ResourceReference{Type: "ref/resource", URI: "file://ref"}
	b, err = json.Marshal(ref)
	if err != nil || !strings.Contains(string(b), `"file://ref"`) {
		t.Fatalf("expected resource reference: %v", err)
	}

	readReq := ReadResourceRequest{JSONRPC: "2.0", ID: NewIntRequestID(1), Method: "resources/read"}
	readReq.Params.URI = "file://read"
	b, err = json.Marshal(readReq)
	if err != nil || !strings.Contains(string(b), `"resources/read"`) || !strings.Contains(string(b), `"file://read"`) {
		t.Fatalf("expected read resource request: %v", err)
	}
	var readReqOut ReadResourceRequest
	if err := json.Unmarshal(b, &readReqOut); err != nil || !reflect.DeepEqual(readReq, readReqOut) {
		t.Errorf("ReadResourceRequest round-trip mismatch: %+v vs %+v (%v)", readReq, readReqOut, err)
	}

	readRes := ReadResourceResult{}
	readRes.Contents = append(readRes.Contents, TextResourceContents{URI: "file://read", Text: "content"})
	b, err = json.Marshal(readRes)
	if err != nil || !strings.Contains(string(b), `"content"`) {
		t.Fatalf("expected read resource result: %v", err)
	}

	listReq := ListResourcesRequest{JSONRPC: "2.0", ID: NewIntRequestID(2), Method: "resources/list"}
	listReq.Params.Cursor = "c1"
	b, err = json.Marshal(listReq)
	if err != nil || !strings.Contains(string(b), `"resources/list"`) {
		t.Fatalf("expected list resources request: %v", err)
	}

	listRes := ListResourcesResult{NextCursor: "c1", Resources: []Resource{res}}
	b, err = json.Marshal(listRes)
	if err != nil || !strings.Contains(string(b), `"TestRes"`) {
		t.Fatalf("expected list resources result: %v", err)
	}
	var listResOut ListResourcesResult
	if err := json.Unmarshal(b, &listResOut); err != nil || !reflect.DeepEqual(listRes, listResOut) {
		t.Errorf("ListResourcesResult round-trip mismatch: %+v vs %+v (%v)", listRes, listResOut, err)
	}

	listTmplReq := ListResourceTemplatesRequest{JSONRPC: "2.0", ID: NewIntRequestID(3), Method: "resources/templates/list"}
	b, err = json.Marshal(listTmplReq)
	if err != nil || !strings.Contains(string(b), `"resources/templates/list"`) {
		t.Fatalf("expected list resource templates request: %v", err)
	}

	listTmplRes := ListResourceTemplatesResult{ResourceTemplates: []ResourceTemplate{tmpl}}
	b, err = json.Marshal(listTmplRes)
	if err != nil || !strings.Contains(string(b), `"Tmpl"`) {
		t.Fatalf("expected list resource templates result: %v", err)
	}
	var listTmplResOut ListResourceTemplatesResult
	if err := json.Unmarshal(b, &listTmplResOut); err != nil || !reflect.DeepEqual(listTmplRes, listTmplResOut) {
		t.Errorf("ListResourceTemplatesResult round-trip mismatch: %+v vs %+v (%v)", listTmplRes, listTmplResOut, err)
	}

	not := ResourceListChangedNotification{JSONRPC: "2.0", Method: "notifications/resources/list_changed"}
	b, err = json.Marshal(not)
	if err != nil || !strings.Contains(string(b), `"notifications/resources/list_changed"`) {
		t.Fatalf("expected notification: %v", err)
	}

	upNot := ResourceUpdatedNotification{JSONRPC: "2.0", Method: "notifications/resources/updated"}
	upNot.Params.URI = "file://up"
	b, err = json.Marshal(upNot)
	if err != nil || !strings.Contains(string(b), `"file://up"`) {
		t.Fatalf("expected updated notification: %v", err)
	}

	subReq := SubscribeRequest{JSONRPC: "2.0", ID: NewIntRequestID(4), Method: "resources/subscribe"}
	subReq.Params.URI = "file://sub"
	b, err = json.Marshal(subReq)
	if err != nil || !strings.Contains(string(b), `"resources/subscribe"`) || !strings.Contains(string(b), `"file://sub"`) {
		t.Fatalf("expected subscribe request: %v", err)
	}

	unsubReq := UnsubscribeRequest{JSONRPC: "2.0", ID: NewIntRequestID(5), Method: "resources/unsubscribe"}
	unsubReq.Params.URI = "file://unsub"
	b, err = json.Marshal(unsubReq)
	if err != nil || !strings.Contains(string(b), `"resources/unsubscribe"`) || !strings.Contains(string(b), `"file://unsub"`) {
		t.Fatalf("expected unsubscribe request: %v", err)
	}
}
