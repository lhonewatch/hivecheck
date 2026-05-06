// Package alert provides alerting hooks and dispatch logic for hivecheck.
package alert

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/example/hivecheck/internal/check"
)

// Payload is the structured data sent to an alerting hook.
type Payload struct {
	// CheckName is the identifier of the health check that triggered the alert.
	CheckName string `json:"check_name"`
	// Status is the evaluated status of the check result.
	Status check.Status `json:"status"`
	// Message is a human-readable description of the check outcome.
	Message string `json:"message"`
	// Value is the raw measured value from the check.
	Value float64 `json:"value"`
	// Timestamp records when the alert was generated.
	Timestamp time.Time `json:"timestamp"`
}

// NewPayload constructs a Payload from a check Result, stamping it with
// the current UTC time.
func NewPayload(name string, r check.Result) Payload {
	return Payload{
		CheckName: name,
		Status:    r.Status,
		Message:   r.Message,
		Value:     r.Value,
		Timestamp: time.Now().UTC(),
	}
}

// JSON serialises the Payload to a compact JSON byte slice.
func (p Payload) JSON() ([]byte, error) {
	b, err := json.Marshal(p)
	if err != nil {
		return nil, fmt.Errorf("alert: marshal payload: %w", err)
	}
	return b, nil
}

// String returns a short human-readable representation of the Payload.
func (p Payload) String() string {
	return fmt.Sprintf("[%s] %s: %s (value=%.4g)",
		p.Status, p.CheckName, p.Message, p.Value)
}
