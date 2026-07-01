package types

import (
	"testing"
	"time"
)

func TestNewIssue(t *testing.T) {
	issue := NewIssue(SeverityCritical, "测试问题", "TEST_ISSUE", "这是一个测试", "test.go")

	if issue.Severity != SeverityCritical {
		t.Errorf("Severity = %v, want %v", issue.Severity, SeverityCritical)
	}
	if issue.CNName != "测试问题" {
		t.Errorf("CNName = %v, want %v", issue.CNName, "测试问题")
	}
	if issue.ENName != "TEST_ISSUE" {
		t.Errorf("ENName = %v, want %v", issue.ENName, "TEST_ISSUE")
	}
	if issue.Location != "test.go" {
		t.Errorf("Location = %v, want %v", issue.Location, "test.go")
	}
	if issue.Metadata == nil {
		t.Error("Metadata should be initialized")
	}
}

func TestIssueWithRemediation(t *testing.T) {
	issue := NewIssue(SeverityWarning, "问题", "ISSUE", "详情", "loc.go")
	issue.WithRemediation("修复建议")

	if issue.Remediation == nil {
		t.Fatal("Remediation should not be nil")
	}
	if issue.Remediation.Suggestion != "修复建议" {
		t.Errorf("Remediation.Suggestion = %v, want %v", issue.Remediation.Suggestion, "修复建议")
	}
}

func TestIssueWithMetadata(t *testing.T) {
	issue := NewIssue(SeverityInfo, "问题", "ISSUE", "详情", "loc.go")
	issue.WithMetadata("key1", "value1").
		WithMetadata("key2", "value2")

	if issue.Metadata["key1"] != "value1" {
		t.Errorf("Metadata[key1] = %v, want %v", issue.Metadata["key1"], "value1")
	}
	if issue.Metadata["key2"] != "value2" {
		t.Errorf("Metadata[key2] = %v, want %v", issue.Metadata["key2"], "value2")
	}
}

func TestCalculateSummary(t *testing.T) {
	issues := []Issue{
		{Severity: SeverityCritical},
		{Severity: SeverityCritical},
		{Severity: SeverityWarning},
		{Severity: SeverityInfo},
		{Severity: SeverityInfo},
		{Severity: SeverityInfo},
	}

	summary := CalculateSummary(issues)

	if summary.Critical != 2 {
		t.Errorf("Critical = %v, want %v", summary.Critical, 2)
	}
	if summary.Warning != 1 {
		t.Errorf("Warning = %v, want %v", summary.Warning, 1)
	}
	if summary.Info != 3 {
		t.Errorf("Info = %v, want %v", summary.Info, 3)
	}
	if summary.Total != 6 {
		t.Errorf("Total = %v, want %v", summary.Total, 6)
	}
}

func TestCalculateSummaryEmpty(t *testing.T) {
	issues := []Issue{}
	summary := CalculateSummary(issues)

	if summary.Total != 0 {
		t.Errorf("Total = %v, want %v", summary.Total, 0)
	}
}

func TestMaxSeverity(t *testing.T) {
	tests := []struct {
		name   string
		issues []Issue
		want   Severity
	}{
		{
			name:   "Empty",
			issues: []Issue{},
			want:   0,
		},
		{
			name:   "Only Info",
			issues: []Issue{{Severity: SeverityInfo}},
			want:   SeverityInfo,
		},
		{
			name:   "Warning and Info",
			issues: []Issue{{Severity: SeverityInfo}, {Severity: SeverityWarning}},
			want:   SeverityWarning,
		},
		{
			name:   "Critical present",
			issues: []Issue{{Severity: SeverityWarning}, {Severity: SeverityCritical}},
			want:   SeverityCritical,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MaxSeverity(tt.issues); got != tt.want {
				t.Errorf("MaxSeverity() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIssueTimestamp(t *testing.T) {
	before := time.Now()
	issue := NewIssue(SeverityInfo, "问题", "ISSUE", "详情", "loc.go")
	after := time.Now()

	if issue.Timestamp.Before(before) || issue.Timestamp.After(after) {
		t.Error("Timestamp should be set to current time")
	}
}
