//go:build !js && !wasip1 && !wasm

// Package webhooks provides parsing, emitting, and mock trigger dispatching of OpenAPI Webhooks and Callbacks.
package webhooks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// TriggerWebhookRequest represents the JSON payload sent to the administrative webhook trigger endpoint.
type TriggerWebhookRequest struct {
	// TargetURL is the destination URL where the outgoing webhook payload will be dispatched.
	TargetURL string `json:"target_url"`
	// Method is the optional HTTP verb used when dispatching the webhook (defaults to POST).
	Method string `json:"method,omitempty"`
	// Headers contains optional HTTP headers to attach to the outgoing webhook request.
	Headers map[string]string `json:"headers,omitempty"`
	// Payload is the raw JSON body to dispatch to the target URL.
	Payload json.RawMessage `json:"payload,omitempty"`
}

// TriggerWebhookResponse represents the JSON response returned by the administrative webhook trigger endpoint.
type TriggerWebhookResponse struct {
	// Webhook is the name of the triggered webhook or callback.
	Webhook string `json:"webhook"`
	// TargetURL is the destination URL that received the dispatched payload.
	TargetURL string `json:"target_url"`
	// StatusCode is the HTTP status code returned by the target receiver.
	StatusCode int `json:"status_code"`
	// ResponseBody is the response payload returned by the target receiver.
	ResponseBody string `json:"response_body,omitempty"`
}

// Dispatcher manages registered OpenAPI webhooks/callbacks and dispatches outgoing HTTP payloads.
type Dispatcher struct {
	// Client is the HTTP client used to send outgoing webhook requests.
	Client *http.Client
	// Registered maps valid webhook or callback names to true when restricted to spec-defined names.
	Registered map[string]bool
}

// NewDispatcher creates a new webhook Dispatcher for the provided webhook/callback names.
// Parameter client is the optional HTTP client (defaults to http.DefaultClient if nil), and names is the slice of registered webhook names.
func NewDispatcher(client *http.Client, names []string) *Dispatcher {
	if client == nil {
		client = &http.Client{}
	}
	reg := make(map[string]bool, len(names))
	for _, n := range names {
		reg[n] = true
	}
	return &Dispatcher{
		Client:     client,
		Registered: reg,
	}
}

// Dispatch sends an outgoing webhook HTTP request to req.TargetURL and returns the receiver's response details.
// Parameter webhookName is the identifier of the webhook being triggered, and req is the trigger configuration.
func (d *Dispatcher) Dispatch(webhookName string, req TriggerWebhookRequest) (*TriggerWebhookResponse, error) {
	if strings.TrimSpace(webhookName) == "" {
		return nil, fmt.Errorf("webhook name is required")
	}
	if len(d.Registered) > 0 && !d.Registered[webhookName] {
		return nil, fmt.Errorf("unknown webhook: %s", webhookName)
	}
	if strings.TrimSpace(req.TargetURL) == "" {
		return nil, fmt.Errorf("target_url is required")
	}
	method := strings.ToUpper(strings.TrimSpace(req.Method))
	if method == "" {
		method = http.MethodPost
	}
	payload := req.Payload
	if len(payload) == 0 {
		payload = json.RawMessage(fmt.Sprintf(`{"webhook":%q}`, webhookName))
	}

	httpReq, err := http.NewRequest(method, req.TargetURL, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Mock-Webhook", webhookName)
	for k, v := range req.Headers {
		httpReq.Header.Set(k, v)
	}

	resp, err := d.Client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	return &TriggerWebhookResponse{
		Webhook:      webhookName,
		TargetURL:    req.TargetURL,
		StatusCode:   resp.StatusCode,
		ResponseBody: string(bodyBytes),
	}, nil
}

// ServeHTTP handles administrative `POST /_mock/trigger-webhook/{webhook_name}` requests.
// Parameter w is the HTTP response writer, and r is the incoming HTTP request.
func (d *Dispatcher) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	const prefix = "/_mock/trigger-webhook/"
	if !strings.HasPrefix(r.URL.Path, prefix) {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	webhookName := strings.TrimPrefix(r.URL.Path, prefix)
	if webhookName == "" {
		http.Error(w, `{"error":"webhook name is required"}`, http.StatusBadRequest)
		return
	}
	if len(d.Registered) > 0 && !d.Registered[webhookName] {
		http.Error(w, `{"error":"unknown webhook"}`, http.StatusNotFound)
		return
	}

	var triggerReq TriggerWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&triggerReq); err != nil {
		http.Error(w, `{"error":"invalid JSON body"}`, http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(triggerReq.TargetURL) == "" {
		http.Error(w, `{"error":"target_url is required"}`, http.StatusBadRequest)
		return
	}

	res, err := d.Dispatch(webhookName, triggerReq)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(res)
}
