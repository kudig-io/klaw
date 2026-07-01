package rca

import (
	"context"
	"testing"

	"github.com/kudig-io/klaw/internal/diag/types"
)

func TestNewEngine(t *testing.T) {
	e := NewEngine()
	if e == nil {
		t.Fatal("Expected non-nil engine")
	}
	if len(e.rules) == 0 {
		t.Error("Expected default rules to be loaded")
	}
}

func TestEngine_AddRule(t *testing.T) {
	e := NewEngine()
	initialCount := len(e.rules)
	e.AddRule(Rule{
		ID:   "test.rule",
		Name: "Test Rule",
		MatchConditions: []MatchCondition{
			{CodePattern: "TEST_*", MinSeverity: types.SeverityWarning},
		},
		MinConfidence: 0.5,
		Result:        RootCause{Title: "Test Root Cause"},
	})
	if len(e.rules) != initialCount+1 {
		t.Errorf("Expected %d rules after adding, got %d", initialCount+1, len(e.rules))
	}
}

func TestEngine_Analyze_EmptyIssues(t *testing.T) {
	e := NewEngine()
	rootCauses := e.Analyze(context.Background(), nil)
	if len(rootCauses) != 0 {
		t.Errorf("Expected 0 root causes for empty issues, got %d", len(rootCauses))
	}
}

func TestEngine_Analyze_DNSFailure(t *testing.T) {
	e := NewEngine()
	issues := []types.Issue{
		{ENName: "DNS_RESOLUTION_FAILED", Severity: types.SeverityWarning, AnalyzerName: "network"},
		{ENName: "NETWORK_DNS_TIMEOUT", Severity: types.SeverityWarning, AnalyzerName: "network"},
	}
	rootCauses := e.Analyze(context.Background(), issues)
	found := false
	for _, rc := range rootCauses {
		if rc.ID == "rca.dns_failure" {
			found = true
			if rc.Confidence < 0.7 {
				t.Errorf("Expected confidence >= 0.7, got %f", rc.Confidence)
			}
			if rc.Category != "network" {
				t.Errorf("Expected category 'network', got '%s'", rc.Category)
			}
			if len(rc.SuggestedActions) == 0 {
				t.Error("Expected suggested actions")
			}
		}
	}
	if !found {
		t.Error("Expected rca.dns_failure root cause to be detected")
	}
}

func TestEngine_Analyze_MemoryPressure(t *testing.T) {
	e := NewEngine()
	issues := []types.Issue{
		{ENName: "HIGH_MEMORY_USAGE", Severity: types.SeverityCritical, AnalyzerName: "system"},
		{ENName: "OOM_KILLER_TRIGGERED", Severity: types.SeverityCritical, AnalyzerName: "kernel"},
		{ENName: "SYSTEM_MEMORY_HIGH", Severity: types.SeverityWarning, AnalyzerName: "system"},
	}
	rootCauses := e.Analyze(context.Background(), issues)
	found := false
	for _, rc := range rootCauses {
		if rc.ID == "rca.memory_pressure" {
			found = true
		}
	}
	if !found {
		t.Error("Expected rca.memory_pressure root cause to be detected")
	}
}

func TestEngine_Analyze_CancelledContext(t *testing.T) {
	e := NewEngine()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	issues := []types.Issue{
		{ENName: "DNS_RESOLUTION_FAILED", Severity: types.SeverityWarning},
	}
	rootCauses := e.Analyze(ctx, issues)
	if len(rootCauses) != 0 {
		t.Error("Expected 0 root causes with cancelled context")
	}
}

func TestMatchGlob(t *testing.T) {
	tests := []struct {
		pattern  string
		s        string
		expected bool
	}{
		{"*", "anything", true},
		{"", "anything", true},
		{"DNS_*", "DNS_RESOLUTION_FAILED", true},
		{"NETWORK_*", "DNS_RESOLUTION_FAILED", false},
		{"NODE_TAINT", "NODE_TAINT", true},
		{"EXACT", "EXACT", true},
		{"EXACT", "OTHER", false},
		{"PREFIX*", "PREFIX_MATCH", true},
		{"PREFIX*", "NO_MATCH", false},
		{"*SUFFIX", "MY_SUFFIX", true},
		{"PRE*SUF", "PRE_MID_SUF", true},
	}
	for _, tc := range tests {
		result := matchGlob(tc.pattern, tc.s)
		if result != tc.expected {
			t.Errorf("matchGlob(%q, %q) = %v, expected %v", tc.pattern, tc.s, result, tc.expected)
		}
	}
}

func TestFormatRootCauses_Empty(t *testing.T) {
	result := FormatRootCauses(nil)
	if result != "未发现明确的根因模式" {
		t.Errorf("Unexpected empty result: %s", result)
	}
}

func TestFormatRootCauses_WithCauses(t *testing.T) {
	rootCauses := []RootCause{
		{
			Title:           "Test Root Cause",
			Confidence:      0.85,
			Category:        "test",
			Description:     "Test description",
			RelatedIssues:   []string{"ISSUE_1"},
			SuggestedActions: []string{"Do something"},
		},
	}
	result := FormatRootCauses(rootCauses)
	if result == "" {
		t.Error("Expected non-empty formatted result")
	}
}
