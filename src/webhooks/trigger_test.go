package webhooks

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/SamuelMarks/cdd-go/src/openapi"
)

func TestWebhookTriggerAndDispatchE2E(t *testing.T) {
	var receivedBody []byte
	var receivedHeader string
	var receivedWebhookHeader string

	dummyReceiver := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeader = r.Header.Get("X-Custom-Signature")
		receivedWebhookHeader = r.Header.Get("X-Mock-Webhook")
		receivedBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"received":true}`))
	}))
	defer dummyReceiver.Close()

	dispatcher := NewDispatcher(nil, []string{"orderCreated", "paymentUpdated"})
	mockServer := httptest.NewServer(dispatcher)
	defer mockServer.Close()

	reqBody := TriggerWebhookRequest{
		TargetURL: dummyReceiver.URL,
		Method:    "POST",
		Headers:   map[string]string{"X-Custom-Signature": "sig-xyz"},
		Payload:   json.RawMessage(`{"order_id":"ord_123","status":"paid"}`),
	}
	rawReq, _ := json.Marshal(reqBody)

	resp, err := http.Post(mockServer.URL+"/_mock/trigger-webhook/orderCreated", "application/json", bytes.NewReader(rawReq))
	if err != nil {
		t.Fatalf("failed to call trigger endpoint: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d", resp.StatusCode)
	}

	var triggerResp TriggerWebhookResponse
	if err := json.NewDecoder(resp.Body).Decode(&triggerResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if triggerResp.Webhook != "orderCreated" {
		t.Errorf("expected webhook orderCreated, got %s", triggerResp.Webhook)
	}
	if triggerResp.StatusCode != http.StatusAccepted {
		t.Errorf("expected target status 202, got %d", triggerResp.StatusCode)
	}
	if triggerResp.ResponseBody != `{"received":true}` {
		t.Errorf("unexpected target response body: %s", triggerResp.ResponseBody)
	}
	if string(receivedBody) != `{"order_id":"ord_123","status":"paid"}` {
		t.Errorf("dummy receiver got unexpected payload: %s", string(receivedBody))
	}
	if receivedHeader != "sig-xyz" {
		t.Errorf("dummy receiver got unexpected X-Custom-Signature: %s", receivedHeader)
	}
	if receivedWebhookHeader != "orderCreated" {
		t.Errorf("dummy receiver got unexpected X-Mock-Webhook: %s", receivedWebhookHeader)
	}
}

func TestDispatcherDefaultPayloadAndErrors(t *testing.T) {
	dummyReceiver := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(b)
	}))
	defer dummyReceiver.Close()

	d := NewDispatcher(dummyReceiver.Client(), []string{"userSignup"})

	// Default payload and default POST method
	res, err := d.Dispatch("userSignup", TriggerWebhookRequest{TargetURL: dummyReceiver.URL})
	if err != nil {
		t.Fatalf("unexpected Dispatch error: %v", err)
	}
	if res.ResponseBody != `{"webhook":"userSignup"}` {
		t.Errorf("expected default payload, got %s", res.ResponseBody)
	}

	// Empty webhook name
	if _, err := d.Dispatch("", TriggerWebhookRequest{TargetURL: dummyReceiver.URL}); err == nil {
		t.Error("expected error for empty webhook name")
	}

	// Unregistered webhook name
	if _, err := d.Dispatch("unknownHook", TriggerWebhookRequest{TargetURL: dummyReceiver.URL}); err == nil {
		t.Error("expected error for unknown webhook")
	}

	// Empty target URL
	if _, err := d.Dispatch("userSignup", TriggerWebhookRequest{}); err == nil {
		t.Error("expected error for empty target_url")
	}

	// Invalid HTTP method for http.NewRequest
	if _, err := d.Dispatch("userSignup", TriggerWebhookRequest{TargetURL: dummyReceiver.URL, Method: "BAD METHOD"}); err == nil {
		t.Error("expected error for invalid HTTP method")
	}

	// Unreachable target URL
	if _, err := d.Dispatch("userSignup", TriggerWebhookRequest{TargetURL: "http://127.0.0.1:0"}); err == nil {
		t.Error("expected error for unreachable target URL")
	}
}

func TestDispatcherServeHTTPErrors(t *testing.T) {
	d := NewDispatcher(nil, []string{"validHook"})

	// 1. Method not allowed
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/_mock/trigger-webhook/validHook", nil)
	d.ServeHTTP(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}

	// 2. Wrong path prefix
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/other/path", nil)
	d.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}

	// 3. Empty webhook name
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/_mock/trigger-webhook/", nil)
	d.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

	// 4. Unknown webhook name
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/_mock/trigger-webhook/unknown", strings.NewReader(`{}`))
	d.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}

	// 5. Invalid JSON body
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/_mock/trigger-webhook/validHook", strings.NewReader(`{bad`))
	d.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

	// 6. Missing target_url
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/_mock/trigger-webhook/validHook", strings.NewReader(`{}`))
	d.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

	// 7. Dispatch failure (502 Bad Gateway)
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/_mock/trigger-webhook/validHook", strings.NewReader(`{"target_url":"http://127.0.0.1:0"}`))
	d.ServeHTTP(w, req)
	if w.Code != http.StatusBadGateway {
		t.Errorf("expected 502, got %d", w.Code)
	}
}

func TestCollectWebhookNamesAndHasWebhooks(t *testing.T) {
	if HasWebhooksOrCallbacks(nil) {
		t.Error("expected false for nil OpenAPI")
	}
	if HasWebhooksOrCallbacks(&openapi.OpenAPI{}) {
		t.Error("expected false for empty OpenAPI")
	}

	oa := &openapi.OpenAPI{
		Webhooks: map[string]openapi.PathItem{
			"newOrder": {},
		},
		Components: &openapi.Components{
			Callbacks: map[string]openapi.Callback{
				"onPayment": {},
			},
		},
		Paths: openapi.Paths{
			"/subscribe": openapi.PathItem{
				Post: &openapi.Operation{
					Callbacks: map[string]openapi.Callback{
						"onEvent": {},
					},
				},
			},
		},
	}
	if !HasWebhooksOrCallbacks(oa) {
		t.Error("expected true when webhooks and callbacks are present")
	}
	names := CollectWebhookNames(oa)
	if len(names) != 3 || names[0] != "newOrder" || names[1] != "onEvent" || names[2] != "onPayment" {
		t.Errorf("unexpected collected webhook names: %v", names)
	}
}
