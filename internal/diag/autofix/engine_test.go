package autofix

import (
	"context"
	"testing"

	"github.com/kudig-io/klaw/internal/diag/types"
)

func TestNewEngine(t *testing.T) {
	e := NewEngine(false)
	if e == nil {
		t.Fatal("Expected non-nil engine")
	}
	if e.dryRun {
		t.Error("Expected dryRun to be false")
	}
}

func TestNewEngine_DryRun(t *testing.T) {
	e := NewEngine(true)
	if !e.dryRun {
		t.Error("Expected dryRun to be true")
	}
}

func TestEngine_CanFix_KnownIssue(t *testing.T) {
	e := NewEngine(false)
	issue := types.Issue{ENName: "SWAP_NOT_DISABLED"}
	action, canFix := e.CanFix(issue)
	if !canFix {
		t.Error("Expected CanFix to return true for SWAP_NOT_DISABLED")
	}
	if action.IssueCode != "SWAP_NOT_DISABLED" {
		t.Errorf("Expected issue code 'SWAP_NOT_DISABLED', got '%s'", action.IssueCode)
	}
}

func TestEngine_CanFix_UnknownIssue(t *testing.T) {
	e := NewEngine(false)
	issue := types.Issue{ENName: "UNKNOWN_ISSUE"}
	_, canFix := e.CanFix(issue)
	if canFix {
		t.Error("Expected CanFix to return false for unknown issue")
	}
}

func TestEngine_Fix_DryRun(t *testing.T) {
	e := NewEngine(true)
	issue := types.Issue{ENName: "SWAP_NOT_DISABLED"}
	result := e.Fix(context.Background(), issue)
	if !result.Success {
		t.Errorf("Expected success in dry-run mode, got: %s", result.Message)
	}
	if result.IssueCode != "SWAP_NOT_DISABLED" {
		t.Errorf("Expected issue code 'SWAP_NOT_DISABLED', got '%s'", result.IssueCode)
	}
}

func TestEngine_Fix_Unfixable(t *testing.T) {
	e := NewEngine(false)
	issue := types.Issue{ENName: "UNKNOWN_ISSUE"}
	result := e.Fix(context.Background(), issue)
	if result.Success {
		t.Error("Expected failure for unfixable issue")
	}
}

func TestEngine_GetFixableIssues(t *testing.T) {
	e := NewEngine(false)
	issues := []types.Issue{
		{ENName: "SWAP_NOT_DISABLED"},
		{ENName: "UNKNOWN_ISSUE"},
		{ENName: "COREDNS_RESTART"},
	}
	fixable := e.GetFixableIssues(issues)
	if len(fixable) != 2 {
		t.Errorf("Expected 2 fixable issues, got %d", len(fixable))
	}
}

func TestEngine_FixAll(t *testing.T) {
	e := NewEngine(true)
	issues := []types.Issue{
		{ENName: "SWAP_NOT_DISABLED"},
		{ENName: "UNKNOWN_ISSUE"},
	}
	results := e.FixAll(context.Background(), issues)
	if len(results) != 1 {
		t.Errorf("Expected 1 result (only fixable), got %d", len(results))
	}
}

func TestFormatResults_Empty(t *testing.T) {
	result := FormatResults(nil)
	if result != "没有可自动修复的问题" {
		t.Errorf("Unexpected empty result: %s", result)
	}
}

func TestFormatResults_WithResults(t *testing.T) {
	results := []FixResult{
		{IssueCode: "TEST", Success: true, Message: "ok"},
	}
	result := FormatResults(results)
	if result == "" {
		t.Error("Expected non-empty formatted result")
	}
}
