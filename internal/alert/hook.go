// Package alert provides alerting hook interfaces and implementations
// for notifying external systems when health check thresholds are breached.
package alert

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Severity represents the urgency level of an alert notification.
type Severity string

const (
	// SeverityWarning indicates a non-critical threshold was breached.
	SeverityWarning Severity = "warning"
	// SeverityCritical indicates a critical threshold was breached.
	SeverityCritical Severity = "critical"
	// SeverityError indicates the check itself failed to execute.
	SeverityError Severity = "error"
)

// Payload is the structured data sent to an alerting hook.
type Payload struct {
	// Service is the name of the service that was checked.
	Service string `json:"service"`
	// CheckName is the identifier of the specific check that triggered the alert.
	CheckName string `json:"check_name"`
	// Severity is the level of the alert.
	Severity Severity `json:"severity"`
	// Message contains a human-readable description of the alert.
	Message string `json:"message"`
	// Value is the observed metric value that triggered the alert.
	Value float64 `json:"value"`
	// Threshold is the configured limit that was breached.
	Threshold float64 `json:"threshold"`
	// OccurredAt is the UTC timestamp when the alert was generated.
	OccurredAt time.Time `json:"occurred_at"`
}

// Hook is the interface that all alerting backends must implement.
type Hook interface {
	// Send delivers an alert payload to the backend.
	// Implementations should respect context cancellation.
	Send(ctx context.Context, p Payload) error
}

// WebhookHook sends alert payloads as JSON POST requests to a configured URL.
type WebhookHook struct {
	// URL is the endpoint that receives alert payloads.
	URL string
	// Client is the HTTP client used for requests. If nil, a default client
	// with a 10-second timeout is used.
	Client *http.Client
	// Headers contains optional HTTP headers added to every request.
	Headers map[string]string
}

// NewWebhookHook creates a WebhookHook with sensible defaults.
func NewWebhookHook(url string, headers map[string]string) *WebhookHook {
	return &WebhookHook{
		URL: url,
		Client: &http.Client{Timeout: 10 * time.Second},
		Headers: headers,
	}
}

// Send marshals the payload to JSON and POSTs it to the configured URL.
// It returns an error if the request fails or the server responds with a
// non-2xx status code.
func (w *WebhookHook) Send(ctx context.Context, p Payload) error {
	body, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("alert: marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, w.URL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("alert: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range w.Headers {
		req.Header.Set(k, v)
	}

	client := w.Client
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("alert: send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("alert: webhook returned non-2xx status: %d", resp.StatusCode)
	}
	return nil
}

// NoopHook is a Hook implementation that discards all alerts.
// It is useful in tests or when alerting is explicitly disabled.
type NoopHook struct{}

// Send discards the payload and always returns nil.
func (n *NoopHook) Send(_ context.Context, _ Payload) error { return nil }
