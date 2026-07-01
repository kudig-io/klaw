package reporter

import (
	"bytes"
	"strings"
	"testing"

	"github.com/kudig-io/klaw/internal/diag/types"
)

func TestTextReporterFormat(t *testing.T) {
	reporter := NewTextReporter(true)
	if reporter.Format() != "text" {
		t.Errorf("Format() = %v, want %v", reporter.Format(), "text")
	}
}

func TestTextReporterGenerateEmpty(t *testing.T) {
	reporter := NewTextReporter(false)
	meta := NewReportMetadata()
	meta.Hostname = "test-node"

	issues := []types.Issue{}

	data, err := reporter.Generate(issues, meta)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	output := string(data)
	if !strings.Contains(output, "未检测到异常") {
		t.Error("Expected '未检测到异常' in output")
	}
	if !strings.Contains(output, "test-node") {
		t.Error("Expected hostname in output")
	}
}

func TestTextReporterGenerateWithIssues(t *testing.T) {
	reporter := NewTextReporter(false)
	meta := NewReportMetadata()
	meta.Hostname = "test-node"

	issues := []types.Issue{
		{
			Severity: types.SeverityCritical,
			CNName:   "严重问题",
			ENName:   "CRITICAL_ISSUE",
			Details:  "问题详情",
			Location: "test.go",
		},
	}

	data, err := reporter.Generate(issues, meta)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	output := string(data)
	if !strings.Contains(output, "严重级别") {
		t.Error("Expected '严重级别' in output")
	}
	if !strings.Contains(output, "严重问题") {
		t.Error("Expected '严重问题' in output")
	}
	if !strings.Contains(output, "CRITICAL_ISSUE") {
		t.Error("Expected 'CRITICAL_ISSUE' in output")
	}
}

func TestTextReporterGenerateWithRemediation(t *testing.T) {
	reporter := NewTextReporter(false)
	meta := NewReportMetadata()

	issue := types.NewIssue(types.SeverityWarning, "警告问题", "WARN_ISSUE", "详情", "loc.go")
	issue.WithRemediation("修复建议")

	issues := []types.Issue{*issue}

	data, err := reporter.Generate(issues, meta)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	output := string(data)
	if !strings.Contains(output, "修复建议") {
		t.Error("Expected remediation suggestion in output")
	}
}

func TestFilterBySeverity(t *testing.T) {
	issues := []types.Issue{
		{Severity: types.SeverityCritical, ENName: "C1"},
		{Severity: types.SeverityCritical, ENName: "C2"},
		{Severity: types.SeverityWarning, ENName: "W1"},
		{Severity: types.SeverityInfo, ENName: "I1"},
	}

	critical := filterBySeverity(issues, types.SeverityCritical)
	if len(critical) != 2 {
		t.Errorf("Critical count = %v, want %v", len(critical), 2)
	}

	warning := filterBySeverity(issues, types.SeverityWarning)
	if len(warning) != 1 {
		t.Errorf("Warning count = %v, want %v", len(warning), 1)
	}
}

func TestDeduplicateIssues(t *testing.T) {
	issues := []types.Issue{
		{Severity: types.SeverityCritical, ENName: "ISSUE1"},
		{Severity: types.SeverityWarning, ENName: "issue1"}, // duplicate (case insensitive)
		{Severity: types.SeverityInfo, ENName: "ISSUE2"},
		{Severity: types.SeverityInfo, ENName: "ISSUE1"}, // duplicate
	}

	result := DeduplicateIssues(issues)
	if len(result) != 2 {
		t.Errorf("Result count = %v, want %v", len(result), 2)
	}
}

func TestSortIssuesBySeverity(t *testing.T) {
	issues := []types.Issue{
		{Severity: types.SeverityInfo, ENName: "INFO1"},
		{Severity: types.SeverityCritical, ENName: "CRIT1"},
		{Severity: types.SeverityWarning, ENName: "WARN1"},
		{Severity: types.SeverityCritical, ENName: "CRIT2"},
	}

	sorted := SortIssuesBySeverity(issues)

	// Should be: CRIT1, CRIT2, WARN1, INFO1
	if sorted[0].ENName != "CRIT1" && sorted[0].ENName != "CRIT2" {
		t.Errorf("First issue should be critical, got %v", sorted[0].ENName)
	}
	if sorted[3].ENName != "INFO1" {
		t.Errorf("Last issue should be info, got %v", sorted[3].ENName)
	}
}

func TestTextReporterWithColor(t *testing.T) {
	reporter := NewTextReporter(true)
	meta := NewReportMetadata()

	issues := []types.Issue{
		{Severity: types.SeverityCritical, CNName: "严重问题", ENName: "CRITICAL"},
	}

	data, err := reporter.Generate(issues, meta)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Check that ANSI color codes are present
	if !bytes.Contains(data, []byte("\033[")) {
		t.Error("Expected ANSI color codes in output")
	}
}

func TestTextReporterWithoutColor(t *testing.T) {
	reporter := NewTextReporter(false)
	meta := NewReportMetadata()

	issues := []types.Issue{
		{Severity: types.SeverityCritical, CNName: "严重问题", ENName: "CRITICAL"},
	}

	data, err := reporter.Generate(issues, meta)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Check that ANSI color codes are NOT present
	if bytes.Contains(data, []byte("\033[")) {
		t.Error("Did not expect ANSI color codes in output")
	}
}
