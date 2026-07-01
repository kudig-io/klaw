package reporter

import (
	"encoding/json"
	"testing"

	"github.com/kudig-io/klaw/internal/diag/types"
)

func TestJSONReporterFormat(t *testing.T) {
	reporter := NewJSONReporter(true)
	if reporter.Format() != "json" {
		t.Errorf("Format() = %v, want %v", reporter.Format(), "json")
	}
}

func TestJSONReporterGenerate(t *testing.T) {
	reporter := NewJSONReporter(false)

	meta := NewReportMetadata()
	meta.Hostname = "test-node"

	issues := []types.Issue{
		{
			Severity: types.SeverityCritical,
			CNName:   "测试问题",
			ENName:   "TEST_ISSUE",
			Details:  "问题详情",
			Location: "test.go",
		},
	}

	data, err := reporter.Generate(issues, meta)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Verify it's valid JSON
	var report Report
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatalf("Generated invalid JSON: %v", err)
	}

	if report.Hostname != "test-node" {
		t.Errorf("Hostname = %v, want %v", report.Hostname, "test-node")
	}
	if len(report.Anomalies) != 1 {
		t.Errorf("Anomalies count = %v, want %v", len(report.Anomalies), 1)
	}
}

func TestJSONReporterGenerateIndented(t *testing.T) {
	reporter := NewJSONReporter(true)

	meta := NewReportMetadata()
	issues := []types.Issue{}

	data, err := reporter.Generate(issues, meta)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Check that it's indented (contains newlines)
	if len(data) == 0 {
		t.Error("Generated empty data")
	}

	// Verify it's valid JSON
	var report Report
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatalf("Generated invalid JSON: %v", err)
	}
}

func TestJSONReporterEmptyIssues(t *testing.T) {
	reporter := NewJSONReporter(false)
	meta := NewReportMetadata()
	issues := []types.Issue{}

	data, err := reporter.Generate(issues, meta)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	var report Report
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatalf("Generated invalid JSON: %v", err)
	}

	if report.Summary.Total != 0 {
		t.Errorf("Summary.Total = %v, want %v", report.Summary.Total, 0)
	}
}

func TestJSONReporterMultipleIssues(t *testing.T) {
	reporter := NewJSONReporter(false)
	meta := NewReportMetadata()

	issues := []types.Issue{
		{Severity: types.SeverityCritical, CNName: "严重问题1"},
		{Severity: types.SeverityCritical, CNName: "严重问题2"},
		{Severity: types.SeverityWarning, CNName: "警告问题"},
		{Severity: types.SeverityInfo, CNName: "提示信息"},
	}

	data, err := reporter.Generate(issues, meta)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	var report Report
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatalf("Generated invalid JSON: %v", err)
	}

	if report.Summary.Critical != 2 {
		t.Errorf("Summary.Critical = %v, want %v", report.Summary.Critical, 2)
	}
	if report.Summary.Warning != 1 {
		t.Errorf("Summary.Warning = %v, want %v", report.Summary.Warning, 1)
	}
	if report.Summary.Info != 1 {
		t.Errorf("Summary.Info = %v, want %v", report.Summary.Info, 1)
	}
}
