package types

import (
	"encoding/json"
	"testing"
)

func TestSeverityString(t *testing.T) {
	tests := []struct {
		name     string
		severity Severity
		want     string
	}{
		{"Critical", SeverityCritical, "严重"},
		{"Warning", SeverityWarning, "警告"},
		{"Info", SeverityInfo, "提示"},
		{"Unknown", Severity(99), "未知"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.severity.String(); got != tt.want {
				t.Errorf("Severity.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSeverityEnglishString(t *testing.T) {
	tests := []struct {
		name     string
		severity Severity
		want     string
	}{
		{"Critical", SeverityCritical, "critical"},
		{"Warning", SeverityWarning, "warning"},
		{"Info", SeverityInfo, "info"},
		{"Unknown", Severity(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.severity.EnglishString(); got != tt.want {
				t.Errorf("Severity.EnglishString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseSeverity(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  Severity
	}{
		{"Chinese Critical", "严重", SeverityCritical},
		{"English Critical", "critical", SeverityCritical},
		{"Chinese Warning", "警告", SeverityWarning},
		{"English Warning", "warning", SeverityWarning},
		{"Chinese Info", "提示", SeverityInfo},
		{"English Info", "info", SeverityInfo},
		{"Mixed Case", "CRITICAL", SeverityCritical},
		{"Unknown", "unknown", SeverityInfo},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseSeverity(tt.input); got != tt.want {
				t.Errorf("ParseSeverity(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestSeverityExitCode(t *testing.T) {
	tests := []struct {
		name     string
		severity Severity
		want     int
	}{
		{"Critical", SeverityCritical, 2},
		{"Warning", SeverityWarning, 1},
		{"Info", SeverityInfo, 1},
		{"Unknown", Severity(0), 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.severity.ExitCode(); got != tt.want {
				t.Errorf("Severity.ExitCode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSeverityMarshalJSON(t *testing.T) {
	data, err := json.Marshal(SeverityCritical)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	want := `"严重"`
	if string(data) != want {
		t.Errorf("MarshalJSON() = %s, want %s", string(data), want)
	}
}

func TestSeverityUnmarshalJSON(t *testing.T) {
	var s Severity
	if err := json.Unmarshal([]byte(`"警告"`), &s); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if s != SeverityWarning {
		t.Errorf("UnmarshalJSON() = %v, want %v", s, SeverityWarning)
	}
}
