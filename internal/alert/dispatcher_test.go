package alert_test

import (
	"errors"
	"testing"

	"github.com/example/hivecheck/internal/alert"
	"github.com/example/hivecheck/internal/check"
)

type mockHook struct {
	called  bool
	payload alert.Payload
	err     error
}

func (m *mockHook) Send(p alert.Payload) error {
	m.called = true
	m.payload = p
	return m.err
}

func TestDispatcher_Dispatch_AboveThreshold(t *testing.T) {
	hook := &mockHook{}
	d := alert.NewDispatcher(hook, alert.WithMinStatus(check.StatusWarn))

	p := alert.Payload{CheckName: "db", Status: check.StatusWarn, Message: "slow"}
	if err := d.Dispatch(p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hook.called {
		t.Error("expected hook to be called")
	}
}

func TestDispatcher_Dispatch_BelowThreshold(t *testing.T) {
	hook := &mockHook{}
	d := alert.NewDispatcher(hook, alert.WithMinStatus(check.StatusCritical))

	p := alert.Payload{CheckName: "db", Status: check.StatusWarn, Message: "slow"}
	if err := d.Dispatch(p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hook.called {
		t.Error("hook should not be called below min status")
	}
}

func TestDispatcher_Dispatch_HookError_LogFailure(t *testing.T) {
	hook := &mockHook{err: errors.New("timeout")}
	d := alert.NewDispatcher(hook,
		alert.WithMinStatus(check.StatusWarn),
		alert.WithLogFailure(true),
	)

	p := alert.Payload{CheckName: "api", Status: check.StatusCritical, Message: "down"}
	// WithLogFailure suppresses the error
	if err := d.Dispatch(p); err != nil {
		t.Fatalf("expected suppressed error, got: %v", err)
	}
}

func TestDispatcher_Dispatch_HookError_Propagated(t *testing.T) {
	hook := &mockHook{err: errors.New("timeout")}
	d := alert.NewDispatcher(hook, alert.WithMinStatus(check.StatusWarn))

	p := alert.Payload{CheckName: "api", Status: check.StatusCritical, Message: "down"}
	if err := d.Dispatch(p); err == nil {
		t.Fatal("expected error to propagate")
	}
}
