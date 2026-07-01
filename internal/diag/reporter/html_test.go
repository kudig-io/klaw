package reporter

import (
	"strings"
	"testing"
	"time"

	"github.com/kudig-io/klaw/internal/diag/types"
)

func TestNewHTMLReporter(t *testing.T) {
	r := NewHTMLReporter()
	if r.Format() != "html" {
		t.Errorf("Expected format 'html', got '%s'", r.Format())
	}
}

func TestHTMLReporterGenerate(t *testing.T) {
	r := NewHTMLReporter()

	issues := []types.Issue{
		{
			Severity:     types.SeverityCritical,
			CNName:       "系统负载过高",
			ENName:       "HIGH_SYSTEM_LOAD",
			Details:      "15分钟平均负载 8.5，超过CPU核心数(2)的4倍",
			Location:     "system_status",
			AnalyzerName: "system.cpu",
			Remediation: &types.Remediation{
				Suggestion: "检查高CPU进程: top -c",
			},
		},
		{
			Severity:     types.SeverityWarning,
			CNName:       "内存使用率偏高",
			ENName:       "ELEVATED_MEMORY_USAGE",
			Details:      "内存使用率 87%",
			Location:     "memory_info",
			AnalyzerName: "system.memory",
		},
		{
			Severity:     types.SeverityInfo,
			CNName:       "Swap未禁用",
			ENName:       "SWAP_NOT_DISABLED",
			Details:      "Kubernetes节点建议禁用swap",
			Location:     "system_info",
			AnalyzerName: "system.swap",
		},
	}

	metadata := NewReportMetadata()
	metadata.Hostname = "test-node-01"
	metadata.Mode = "offline"
	metadata.DiagnosePath = "/tmp/diagnose_test"
	metadata.Summary = types.CalculateSummary(issues)

	output, err := r.Generate(issues, metadata)
	if err != nil {
		t.Fatalf("Failed to generate HTML report: %v", err)
	}

	html := string(output)

	// Check for key elements
	if !strings.Contains(html, "<!DOCTYPE html>") {
		t.Error("HTML report should contain DOCTYPE declaration")
	}

	if !strings.Contains(html, "Kudig Diagnostic Report") {
		t.Error("HTML report should contain title")
	}

	if !strings.Contains(html, "系统负载过高") {
		t.Error("HTML report should contain issue CNName")
	}

	if !strings.Contains(html, "HIGH_SYSTEM_LOAD") {
		t.Error("HTML report should contain issue ENName")
	}

	if !strings.Contains(html, "chart.js") {
		t.Error("HTML report should include Chart.js CDN")
	}

	if !strings.Contains(html, "severityChart") {
		t.Error("HTML report should contain severity chart canvas")
	}

	if !strings.Contains(html, "categoryChart") {
		t.Error("HTML report should contain category chart canvas")
	}

	// Check for severity-specific sections
	if !strings.Contains(html, "Critical Issues") {
		t.Error("HTML report should have Critical Issues section")
	}

	if !strings.Contains(html, "Warning Issues") {
		t.Error("HTML report should have Warning Issues section")
	}

	if !strings.Contains(html, "Info Issues") {
		t.Error("HTML report should have Info Issues section")
	}
}

func TestHTMLReporterGenerateEmpty(t *testing.T) {
	r := NewHTMLReporter()

	issues := []types.Issue{}

	metadata := NewReportMetadata()
	metadata.Hostname = "test-node-02"
	metadata.Summary = types.CalculateSummary(issues)

	output, err := r.Generate(issues, metadata)
	if err != nil {
		t.Fatalf("Failed to generate HTML report: %v", err)
	}

	html := string(output)

	if !strings.Contains(html, "No Issues Found") {
		t.Error("HTML report should show empty state when no issues")
	}
}

func TestFilterIssuesBySeverity(t *testing.T) {
	issues := []types.Issue{
		{Severity: types.SeverityCritical, CNName: "Critical 1"},
		{Severity: types.SeverityWarning, CNName: "Warning 1"},
		{Severity: types.SeverityCritical, CNName: "Critical 2"},
		{Severity: types.SeverityInfo, CNName: "Info 1"},
	}

	critical := filterIssuesBySeverity(issues, types.SeverityCritical)
	if len(critical) != 2 {
		t.Errorf("Expected 2 critical issues, got %d", len(critical))
	}

	warning := filterIssuesBySeverity(issues, types.SeverityWarning)
	if len(warning) != 1 {
		t.Errorf("Expected 1 warning issue, got %d", len(warning))
	}

	info := filterIssuesBySeverity(issues, types.SeverityInfo)
	if len(info) != 1 {
		t.Errorf("Expected 1 info issue, got %d", len(info))
	}
}

func TestIssueCategory(t *testing.T) {
	tests := []struct {
		analyzerName string
		expected     string
	}{
		{"system.cpu", "system"},
		{"network.interface", "network"},
		{"kubernetes.pod", "kubernetes"},
		{"process.kubelet", "process"},
		{"kernel.oom", "kernel"},
		{"runtime.docker", "runtime"},
		{"log.syslog", "log"},
		{"unknown.analyzer", "other"},
		{"", "other"},
	}

	for _, tc := range tests {
		issue := types.Issue{AnalyzerName: tc.analyzerName}
		result := issueCategory(issue)
		if result != tc.expected {
			t.Errorf("issueCategory(%q) = %q, expected %q", tc.analyzerName, result, tc.expected)
		}
	}
}

func TestHTMLReporterWithTimestamp(t *testing.T) {
	r := NewHTMLReporter()

	metadata := NewReportMetadata()
	metadata.Timestamp = time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC)
	metadata.Summary = types.IssueSummary{}

	output, err := r.Generate([]types.Issue{}, metadata)
	if err != nil {
		t.Fatalf("Failed to generate HTML report: %v", err)
	}

	html := string(output)

	// Check that current timestamp is included (since we use time.Now() in template)
	if !strings.Contains(html, "2026-01-15") && !strings.Contains(html, time.Now().Format("2006-01-02")) {
		t.Error("HTML report should contain timestamp")
	}
}
