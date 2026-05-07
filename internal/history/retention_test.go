package history

import (
	"testing"
	"time"
)

func TestDefaultRetentionPolicy(t *testing.T) {
	p := DefaultRetentionPolicy()
	if p.MaxEntries != 100 {
		t.Errorf("expected MaxEntries=100, got %d", p.MaxEntries)
	}
	if p.MaxAge != 30*24*time.Hour {
		t.Errorf("expected MaxAge=30d, got %v", p.MaxAge)
	}
}

func TestRetentionPolicy_IsZero(t *testing.T) {
	if !(RetentionPolicy{}).IsZero() {
		t.Error("expected zero policy to be zero")
	}
	if DefaultRetentionPolicy().IsZero() {
		t.Error("expected default policy to be non-zero")
	}
}

func TestRetentionPolicy_Validate_OK(t *testing.T) {
	policies := []RetentionPolicy{
		DefaultRetentionPolicy(),
		{MaxEntries: 0, MaxAge: 0},
		{MaxEntries: 50, MaxAge: 7 * 24 * time.Hour},
	}
	for _, p := range policies {
		if err := p.Validate(); err != nil {
			t.Errorf("unexpected error for policy %+v: %v", p, err)
		}
	}
}

func TestRetentionPolicy_Validate_NegativeEntries(t *testing.T) {
	p := RetentionPolicy{MaxEntries: -1}
	err := p.Validate()
	if err == nil {
		t.Fatal("expected error for negative MaxEntries")
	}
	re, ok := err.(*RetentionError)
	if !ok {
		t.Fatalf("expected *RetentionError, got %T", err)
	}
	if re.Field != "max_entries" {
		t.Errorf("unexpected field: %s", re.Field)
	}
}

func TestRetentionPolicy_Validate_NegativeAge(t *testing.T) {
	p := RetentionPolicy{MaxAge: -time.Hour}
	err := p.Validate()
	if err == nil {
		t.Fatal("expected error for negative MaxAge")
	}
	re, ok := err.(*RetentionError)
	if !ok {
		t.Fatalf("expected *RetentionError, got %T", err)
	}
	if re.Field != "max_age" {
		t.Errorf("unexpected field: %s", re.Field)
	}
}

func TestRetentionError_Message(t *testing.T) {
	err := &RetentionError{Field: "max_entries", Value: -5}
	if err.Error() == "" {
		t.Error("expected non-empty error message")
	}
}
