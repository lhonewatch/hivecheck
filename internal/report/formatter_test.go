package report_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/example/hivecheck/internal/check"
	"github.com/example/hivecheck/internal/report"
)

func TestNewFormatter_Text(t *testing.T) {
	f, err := report.NewFormatter("text")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f == nil {
		t.Fatal("expected non-nil formatter")
	}
}

func TestNewFormatter_JSON(t *testing.T) {
	f, err := report.NewFormatter("json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f == nil {
		t.Fatal("expected non-nil formatter")
	}
}

func TestNewFormatter_Default(t *testing.T) {
	f, err := report.NewFormatter("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f == nil {
		t.Fatal("expected non-nil formatter")
	}
}

func TestNewFormatter_Unknown(t *testing.T) {
	_, err := report.NewFormatter("xml")
	if err == nil {
		t.Fatal("expected error for unknown format")
	}
}

func TestTextFormatter_Format(t *testing.T) {
	f := &report.TextFormatter{}
	s := report.Summary{
		Overall: check.StatusOK,
		Total:   1,
		Counts:  map[string]int{"ok": 1},
		Results: []check.Result{
			{Name: "db", Status: check.StatusOK, Duration: 5 * time.Millisecond, Message: "connected"},
		},
	}
	var buf bytes.Buffer
	if err := f.Format(&buf, s); err != nil {
		t.Fatalf("Format error: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"HiveCheck", "Overall", "db", "connected", "ok"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\n%s", want, out)
		}
	}
}
