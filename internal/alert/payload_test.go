package alert_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/example/hivecheck/internal/alert"
	"github.com/example/hivecheck/internal/check"
)

func makeResult(status check.Status, msg string, val float64) check.Result {
	return check.Result{Status: status, Message: msg, Value: val}
}

func TestNewPayload_Fields(t *testing.T) {
	r := makeResult(check.StatusCritical, "latency high", 3.14)
	p := alert.NewPayload("svc-api", r)

	if p.CheckName != "svc-api" {
		t.Errorf("CheckName: got %q, want %q", p.CheckName, "svc-api")
	}
	if p.Status != check.StatusCritical {
		t.Errorf("Status: got %v, want %v", p.Status, check.StatusCritical)
	}
	if p.Message != "latency high" {
		t.Errorf("Message: got %q", p.Message)
	}
	if p.Value != 3.14 {
		t.Errorf("Value: got %v", p.Value)
	}
	if p.Timestamp.IsZero() {
		t.Error("Timestamp should not be zero")
	}
}

func TestPayload_JSON_RoundTrip(t *testing.T) {
	r := makeResult(check.StatusWarn, "disk usage", 0.87)
	p := alert.NewPayload("disk-check", r)

	b, err := p.JSON()
	if err != nil {
		t.Fatalf("JSON() error: %v", err)
	}

	var decoded alert.Payload
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	if decoded.CheckName != p.CheckName {
		t.Errorf("round-trip CheckName mismatch: %q vs %q", decoded.CheckName, p.CheckName)
	}
	if decoded.Value != p.Value {
		t.Errorf("round-trip Value mismatch: %v vs %v", decoded.Value, p.Value)
	}
}

func TestPayload_String_ContainsFields(t *testing.T) {
	r := makeResult(check.StatusOK, "all good", 0)
	p := alert.NewPayload("heartbeat", r)
	s := p.String()

	for _, want := range []string{"heartbeat", "all good"} {
		if !strings.Contains(s, want) {
			t.Errorf("String() %q missing %q", s, want)
		}
	}
}
