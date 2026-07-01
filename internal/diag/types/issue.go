// Package types defines core data structures for klaw
package types

import (
	"time"
)

// Issue represents a detected anomaly or problem
type Issue struct {
	// Severity indicates the severity level (Critical/Warning/Info)
	Severity Severity `json:"severity" yaml:"severity"`

	// CNName is the Chinese name of the issue
	CNName string `json:"cn_name" yaml:"cn_name"`

	// ENName is the English identifier of the issue (e.g., HIGH_SYSTEM_LOAD)
	ENName string `json:"en_name" yaml:"en_name"`

	// Details provides detailed description of the issue
	Details string `json:"details" yaml:"details"`

	// Location indicates where the issue was found (e.g., file path)
	Location string `json:"location" yaml:"location"`

	// Timestamp is when the issue was detected
	Timestamp time.Time `json:"timestamp,omitempty" yaml:"timestamp,omitempty"`

	// Remediation provides fix suggestions (optional)
	Remediation *Remediation `json:"remediation,omitempty" yaml:"remediation,omitempty"`

	// Metadata stores additional key-value pairs
	Metadata map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty"`

	// AnalyzerName indicates which analyzer detected this issue
	AnalyzerName string `json:"analyzer_name,omitempty" yaml:"analyzer_name,omitempty"`
}

// Remediation represents fix suggestions for an issue
type Remediation struct {
	// Suggestion is the human-readable fix suggestion
	Suggestion string `json:"suggestion" yaml:"suggestion"`

	// Command is the command to execute for automatic fix (optional)
	Command string `json:"command,omitempty" yaml:"command,omitempty"`

	// AutoFix indicates if this can be automatically fixed
	AutoFix bool `json:"auto_fix,omitempty" yaml:"auto_fix,omitempty"`

	// Risk indicates the risk level of the remediation (low/medium/high)
	Risk string `json:"risk,omitempty" yaml:"risk,omitempty"`
}

// NewIssue creates a new Issue with the given parameters
func NewIssue(severity Severity, cnName, enName, details, location string) *Issue {
	return &Issue{
		Severity:  severity,
		CNName:    cnName,
		ENName:    enName,
		Details:   details,
		Location:  location,
		Timestamp: time.Now(),
		Metadata:  make(map[string]string),
	}
}

// WithRemediation adds remediation suggestion to the issue
func (i *Issue) WithRemediation(suggestion string) *Issue {
	i.Remediation = &Remediation{
		Suggestion: suggestion,
	}
	return i
}

// WithMetadata adds metadata to the issue
func (i *Issue) WithMetadata(key, value string) *Issue {
	if i.Metadata == nil {
		i.Metadata = make(map[string]string)
	}
	i.Metadata[key] = value
	return i
}

// IssueSummary provides statistics about issues
type IssueSummary struct {
	Critical int `json:"critical" yaml:"critical"`
	Warning  int `json:"warning" yaml:"warning"`
	Info     int `json:"info" yaml:"info"`
	Total    int `json:"total" yaml:"total"`
}

// CalculateSummary calculates summary statistics from a list of issues
func CalculateSummary(issues []Issue) IssueSummary {
	summary := IssueSummary{}
	for _, issue := range issues {
		switch issue.Severity {
		case SeverityCritical:
			summary.Critical++
		case SeverityWarning:
			summary.Warning++
		case SeverityInfo:
			summary.Info++
		}
		summary.Total++
	}
	return summary
}

// MaxSeverity returns the highest severity from a list of issues
func MaxSeverity(issues []Issue) Severity {
	maxSev := SeverityInfo
	for _, issue := range issues {
		if issue.Severity < maxSev {
			maxSev = issue.Severity
		}
	}
	if len(issues) == 0 {
		return 0 // No issues
	}
	return maxSev
}
