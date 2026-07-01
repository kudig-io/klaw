// Package rules provides a YAML-based rule engine for klaw
package rules

import (
	"github.com/kudig-io/klaw/internal/diag/types"
)

// Rule defines a diagnostic rule
type Rule struct {
	// ID is the unique identifier for the rule
	ID string `yaml:"id" json:"id"`

	// Name is the display name
	Name string `yaml:"name" json:"name"`

	// Description explains what the rule checks
	Description string `yaml:"description" json:"description"`

	// Category groups related rules
	Category string `yaml:"category" json:"category"`

	// Severity of the issue when triggered
	Severity string `yaml:"severity" json:"severity"`

	// Enabled controls whether the rule is active
	Enabled bool `yaml:"enabled" json:"enabled"`

	// Condition defines when the rule triggers
	Condition Condition `yaml:"condition" json:"condition"`

	// Remediation suggests how to fix the issue
	Remediation string `yaml:"remediation" json:"remediation"`

	// Tags for filtering
	Tags []string `yaml:"tags,omitempty" json:"tags,omitempty"`
}

// Condition defines the rule trigger condition
type Condition struct {
	// Type is the condition type: "file_contains", "metric_threshold", "regex_match", "command_output"
	Type string `yaml:"type" json:"type"`

	// File to check (for file-based conditions)
	File string `yaml:"file,omitempty" json:"file,omitempty"`

	// Pattern to match (regex or string)
	Pattern string `yaml:"pattern,omitempty" json:"pattern,omitempty"`

	// Metric name (for metric-based conditions)
	Metric string `yaml:"metric,omitempty" json:"metric,omitempty"`

	// Operator for comparison: "gt", "lt", "gte", "lte", "eq", "ne"
	Operator string `yaml:"operator,omitempty" json:"operator,omitempty"`

	// Threshold value for comparison
	Threshold float64 `yaml:"threshold,omitempty" json:"threshold,omitempty"`

	// Count minimum occurrences to trigger
	Count int `yaml:"count,omitempty" json:"count,omitempty"`

	// Negate inverts the condition result
	Negate bool `yaml:"negate,omitempty" json:"negate,omitempty"`

	// And combines multiple conditions with AND logic
	And []Condition `yaml:"and,omitempty" json:"and,omitempty"`

	// Or combines multiple conditions with OR logic
	Or []Condition `yaml:"or,omitempty" json:"or,omitempty"`
}

// RuleSet is a collection of rules
type RuleSet struct {
	// Version of the rule set format
	Version string `yaml:"version" json:"version"`

	// Name of the rule set
	Name string `yaml:"name" json:"name"`

	// Description of the rule set
	Description string `yaml:"description,omitempty" json:"description,omitempty"`

	// Rules in the set
	Rules []Rule `yaml:"rules" json:"rules"`
}

// GetSeverity converts string severity to types.Severity
func (r *Rule) GetSeverity() types.Severity {
	switch r.Severity {
	case "critical", "严重":
		return types.SeverityCritical
	case "warning", "警告":
		return types.SeverityWarning
	case "info", "提示":
		return types.SeverityInfo
	default:
		return types.SeverityInfo
	}
}

// ToIssue converts a rule match to an Issue
func (r *Rule) ToIssue(details, location string) *types.Issue {
	issue := types.NewIssue(
		r.GetSeverity(),
		r.Name,
		r.ID,
		details,
		location,
	).WithRemediation(r.Remediation)
	issue.AnalyzerName = "rules." + r.Category
	return issue
}
