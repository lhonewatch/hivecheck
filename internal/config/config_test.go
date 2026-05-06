package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/example/hivecheck/internal/config"
)

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = f.WriteString(content)
	_ = f.Close()
	return f.Name()
}

func TestLoad_Valid(t *testing.T) {
	path := writeTemp(t, `
checks:
  - name: api
    url: http://localhost:8080/health
    interval: 30s
    timeout: 5s
    warn_at: 200
    crit_at: 500
alerts:
  - url: http://hooks.example.com/notify
`)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Checks) != 1 {
		t.Fatalf("expected 1 check, got %d", len(cfg.Checks))
	}
	if cfg.Checks[0].Name != "api" {
		t.Errorf("expected name 'api', got %q", cfg.Checks[0].Name)
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := config.Load(filepath.Join(t.TempDir(), "nonexistent.yaml"))
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestValidate_NoChecks(t *testing.T) {
	cfg := &config.Config{}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for empty checks")
	}
}

func TestValidate_WarnGtCrit(t *testing.T) {
	cfg := &config.Config{
		Checks: []config.CheckConfig{
			{Name: "svc", URL: "http://example.com", WarnAt: 600, CritAt: 500},
		},
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error when warn_at > crit_at")
	}
}

func TestValidate_MissingURL(t *testing.T) {
	cfg := &config.Config{
		Checks: []config.CheckConfig{
			{Name: "svc"},
		},
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for missing url")
	}
}
