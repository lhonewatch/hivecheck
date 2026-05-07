package history

import (
	"bytes"
	"encoding/csv"
	"strings"
	"testing"
	"time"
)

func TestExportCSV_Header(t *testing.T) {
	dir := t.TempDir()
	s, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	var buf bytes.Buffer
	if err := ExportCSV(s, &buf, nil); err != nil {
		t.Fatalf("ExportCSV: %v", err)
	}

	line := strings.SplitN(buf.String(), "\n", 2)[0]
	want := "check,timestamp,status,message,duration_ms"
	if line != want {
		t.Errorf("header = %q; want %q", line, want)
	}
}

func TestExportCSV_Rows(t *testing.T) {
	dir := t.TempDir()
	s, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	now := time.Now().UTC().Truncate(time.Second)
	entry := Entry{
		CheckName:  "api",
		Timestamp:  now,
		Status:     "ok",
		Message:    "all good",
		DurationMs: 42,
	}
	if err := s.Save(entry); err != nil {
		t.Fatalf("Save: %v", err)
	}

	var buf bytes.Buffer
	if err := ExportCSV(s, &buf, []string{"api"}); err != nil {
		t.Fatalf("ExportCSV: %v", err)
	}

	r := csv.NewReader(&buf)
	records, err := r.ReadAll()
	if err != nil {
		t.Fatalf("csv.ReadAll: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("expected 2 records (header+row), got %d", len(records))
	}
	row := records[1]
	if row[0] != "api" {
		t.Errorf("check = %q; want %q", row[0], "api")
	}
	if row[2] != "ok" {
		t.Errorf("status = %q; want %q", row[2], "ok")
	}
	if row[4] != "42" {
		t.Errorf("duration_ms = %q; want %q", row[4], "42")
	}
}

func TestExportCSV_NilStore(t *testing.T) {
	var buf bytes.Buffer
	if err := ExportCSV(nil, &buf, nil); err == nil {
		t.Error("expected error for nil store, got nil")
	}
}
