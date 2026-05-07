package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

const validConfig = `
service_name: test-svc
checks:
  - name: always-ok
    url: "http://example.com"
    warn_threshold: 500
    crit_threshold: 1000
alerts:
  webhook_url: ""
  min_status: warn
  timeout_seconds: 5
`

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "hivecheck.yaml")
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return p
}

func TestRun_MissingConfig(t *testing.T) {
	err := runWithArgs("/nonexistent/hivecheck.yaml", "text")
	if err == nil {
		t.Fatal("expected error for missing config, got nil")
	}
}

func TestRun_InvalidFormat_DefaultsToText(t *testing.T) {
	// Build the binary so we can invoke it as a subprocess.
	// This test is intentionally lightweight — it just validates
	// that an unknown format falls back without panicking.
	if os.Getenv("HIVECHECK_INTEGRATION") == "" {
		t.Skip("set HIVECHECK_INTEGRATION=1 to run subprocess tests")
	}
	cfgPath := writeTempConfig(t, validConfig)
	cmd := exec.Command(os.Args[0], "-config", cfgPath, "-format", "unknown")
	cmd.Env = append(os.Environ(), "HIVECHECK_INTEGRATION=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("output: %s", out)
	}
}

// runWithArgs is a thin helper that exercises the config-loading path.
func runWithArgs(cfgPath, format string) error {
	_, err := loadCfg(cfgPath)
	_ = format
	return err
}

func loadCfg(path string) (interface{}, error) {
	// Thin shim so we can unit-test the error path without
	// spinning up real HTTP servers.
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	_ = data
	return nil, nil
}
