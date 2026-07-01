package rules

import (
	"testing"

	"github.com/kudig-io/klaw/internal/diag/types"
)

func TestRuleGetSeverity(t *testing.T) {
	tests := []struct {
		name     string
		severity string
		want     types.Severity
	}{
		{"Critical Chinese", "严重", types.SeverityCritical},
		{"Critical English", "critical", types.SeverityCritical},
		{"Warning Chinese", "警告", types.SeverityWarning},
		{"Warning English", "warning", types.SeverityWarning},
		{"Info Chinese", "提示", types.SeverityInfo},
		{"Info English", "info", types.SeverityInfo},
		{"Unknown", "unknown", types.SeverityInfo},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Rule{Severity: tt.severity}
			if got := r.GetSeverity(); got != tt.want {
				t.Errorf("GetSeverity() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRuleToIssue(t *testing.T) {
	r := &Rule{
		ID:          "TEST_RULE",
		Name:        "测试规则",
		Category:    "test",
		Severity:    "warning",
		Remediation: "修复建议",
	}

	issue := r.ToIssue("问题详情", "test.go")

	if issue.ENName != "TEST_RULE" {
		t.Errorf("ENName = %v, want %v", issue.ENName, "TEST_RULE")
	}
	if issue.CNName != "测试规则" {
		t.Errorf("CNName = %v, want %v", issue.CNName, "测试规则")
	}
	if issue.Severity != types.SeverityWarning {
		t.Errorf("Severity = %v, want %v", issue.Severity, types.SeverityWarning)
	}
	if issue.Details != "问题详情" {
		t.Errorf("Details = %v, want %v", issue.Details, "问题详情")
	}
	if issue.Location != "test.go" {
		t.Errorf("Location = %v, want %v", issue.Location, "test.go")
	}
	if issue.Remediation == nil || issue.Remediation.Suggestion != "修复建议" {
		t.Error("Remediation not set correctly")
	}
}

func TestRuleToIssue_Critical(t *testing.T) {
	r := &Rule{
		ID:       "CRITICAL_RULE",
		Name:     "严重问题",
		Category: "security",
		Severity: "critical",
	}

	issue := r.ToIssue("严重问题详情", "critical.go")

	if issue.Severity != types.SeverityCritical {
		t.Errorf("Severity = %v, want %v", issue.Severity, types.SeverityCritical)
	}
}

func TestConditionEmpty(t *testing.T) {
	c := Condition{}
	if c.Type != "" {
		t.Error("Empty condition should have empty type")
	}
}

func TestRuleSetStructure(t *testing.T) {
	rs := RuleSet{
		Version:     "1.0",
		Name:        "Test Rules",
		Description: "For testing",
		Rules: []Rule{
			{
				ID:       "RULE_1",
				Name:     "Rule One",
				Enabled:  true,
				Severity: "warning",
			},
			{
				ID:       "RULE_2",
				Name:     "Rule Two",
				Enabled:  false,
				Severity: "info",
			},
		},
	}

	if rs.Version != "1.0" {
		t.Errorf("Version = %v, want %v", rs.Version, "1.0")
	}
	if len(rs.Rules) != 2 {
		t.Errorf("Rules count = %v, want %v", len(rs.Rules), 2)
	}
	if rs.Rules[0].ID != "RULE_1" {
		t.Error("First rule ID mismatch")
	}
	if !rs.Rules[0].Enabled {
		t.Error("First rule should be enabled")
	}
	if rs.Rules[1].Enabled {
		t.Error("Second rule should be disabled")
	}
}

func TestRuleWithTags(t *testing.T) {
	r := &Rule{
		ID:   "TAGGED_RULE",
		Tags: []string{"kubernetes", "security", "performance"},
	}

	if len(r.Tags) != 3 {
		t.Errorf("Tags count = %v, want %v", len(r.Tags), 3)
	}
	if r.Tags[0] != "kubernetes" {
		t.Errorf("First tag = %v, want %v", r.Tags[0], "kubernetes")
	}
}

func TestConditionWithAnd(t *testing.T) {
	c := Condition{
		Type:    "file_contains",
		File:    "test.log",
		Pattern: "error",
		And: []Condition{
			{
				Type:    "file_contains",
				File:    "test2.log",
				Pattern: "warning",
			},
		},
	}

	if len(c.And) != 1 {
		t.Errorf("And conditions count = %v, want %v", len(c.And), 1)
	}
}

func TestConditionWithOr(t *testing.T) {
	c := Condition{
		Type: "metric_threshold",
		Or: []Condition{
			{
				Metric:    "cpu",
				Operator:  "gt",
				Threshold: 90,
			},
			{
				Metric:    "memory",
				Operator:  "gt",
				Threshold: 80,
			},
		},
	}

	if len(c.Or) != 2 {
		t.Errorf("Or conditions count = %v, want %v", len(c.Or), 2)
	}
}
