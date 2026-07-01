// Package types defines core data structures for klaw
package types

import (
	"encoding/json"
	"strings"
)

// Severity represents the severity level of an issue
type Severity int

const (
	// SeverityCritical indicates a critical issue that needs immediate attention
	SeverityCritical Severity = iota + 1
	// SeverityWarning indicates a warning that should be investigated
	SeverityWarning
	// SeverityInfo indicates an informational finding
	SeverityInfo
)

// String returns the Chinese string representation of severity
func (s Severity) String() string {
	switch s {
	case SeverityCritical:
		return "严重"
	case SeverityWarning:
		return "警告"
	case SeverityInfo:
		return "提示"
	default:
		return "未知"
	}
}

// EnglishString returns the English string representation of severity
func (s Severity) EnglishString() string {
	switch s {
	case SeverityCritical:
		return "critical"
	case SeverityWarning:
		return "warning"
	case SeverityInfo:
		return "info"
	default:
		return "unknown"
	}
}

// MarshalJSON implements json.Marshaler
func (s Severity) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

// UnmarshalJSON implements json.Unmarshaler
func (s *Severity) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	*s = ParseSeverity(str)
	return nil
}

// ParseSeverity parses a string to Severity
func ParseSeverity(s string) Severity {
	switch strings.ToLower(s) {
	case "严重", "critical":
		return SeverityCritical
	case "警告", "warning":
		return SeverityWarning
	case "提示", "info":
		return SeverityInfo
	default:
		return SeverityInfo
	}
}

// ExitCode returns the exit code for a given severity
// 0 = no issues, 1 = warning/info, 2 = critical
func (s Severity) ExitCode() int {
	switch s {
	case SeverityCritical:
		return 2
	case SeverityWarning, SeverityInfo:
		return 1
	default:
		return 0
	}
}
