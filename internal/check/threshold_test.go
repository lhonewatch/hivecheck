package check

import (
	"testing"
)

func TestThresholdEvaluate(t *testing.T) {
	th := Threshold{Warn: 80, Critical: 95}

	cases := []struct {
		value    float64
		wantStatus Status
	}{
		{50, StatusOK},
		{79.9, StatusOK},
		{80, StatusWarn},
		{90, StatusWarn},
		{95, StatusCritical},
		{100, StatusCritical},
	}

	for _, tc := range cases {
		got := th.Evaluate(tc.value)
		if got != tc.wantStatus {
			t.Errorf("Evaluate(%.1f) = %s, want %s", tc.value, got, tc.wantStatus)
		}
	}
}

func TestStatusString(t *testing.T) {
	cases := []struct {
		status Status
		want   string
	}{
		{StatusOK, "OK"},
		{StatusWarn, "WARN"},
		{StatusCritical, "CRITICAL"},
		{StatusUnknown, "UNKNOWN"},
	}
	for _, tc := range cases {
		if got := tc.status.String(); got != tc.want {
			t.Errorf("Status(%d).String() = %q, want %q", tc.status, got, tc.want)
		}
	}
}

func TestResultString(t *testing.T) {
	r := Result{
		Name:      "cpu_usage",
		Status:    StatusWarn,
		Message:   "CPU usage elevated",
		Value:     82.5,
		Threshold: Threshold{Warn: 80, Critical: 95},
	}
	s := r.String()
	if s == "" {
		t.Error("Result.String() returned empty string")
	}
}
