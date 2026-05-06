package alert_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/example/hivecheck/internal/alert"
	"github.com/example/hivecheck/internal/check"
)

func TestWebhookHook_Send_OK(t *testing.T) {
	var received map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Errorf("failed to decode body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	h := alert.NewWebhookHook(server.URL, 5*time.Second)
	result := check.Result{
		Name:    "latency",
		Status:  check.StatusCritical,
		Message: "latency too high",
		Value:   350.0,
	}

	if err := h.Send(result); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if received["name"] != "latency" {
		t.Errorf("expected name=latency, got %v", received["name"])
	}
	if received["status"] != "CRITICAL" {
		t.Errorf("expected status=CRITICAL, got %v", received["status"])
	}
}

func TestWebhookHook_Send_Non2xx(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	h := alert.NewWebhookHook(server.URL, 5*time.Second)
	result := check.Result{
		Name:   "disk",
		Status: check.StatusWarn,
	}

	if err := h.Send(result); err == nil {
		t.Fatal("expected error for non-2xx response, got nil")
	}
}

func TestWebhookHook_Send_InvalidURL(t *testing.T) {
	h := alert.NewWebhookHook("http://127.0.0.1:0/no-server", 1*time.Second)
	result := check.Result{
		Name:   "memory",
		Status: check.StatusCritical,
	}

	if err := h.Send(result); err == nil {
		t.Fatal("expected connection error, got nil")
	}
}
