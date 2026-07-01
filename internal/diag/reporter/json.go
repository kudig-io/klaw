// Package reporter provides report generation functionality
package reporter

import (
	"encoding/json"

	"github.com/kudig-io/klaw/internal/diag/types"
)

// JSONReporter generates JSON format reports
type JSONReporter struct {
	// Indent controls whether to pretty-print the JSON
	Indent bool
}

// NewJSONReporter creates a new JSON reporter
func NewJSONReporter(indent bool) *JSONReporter {
	return &JSONReporter{Indent: indent}
}

// Format returns the output format name
func (r *JSONReporter) Format() string {
	return "json"
}

// Generate creates a JSON report from the issues
func (r *JSONReporter) Generate(issues []types.Issue, metadata *ReportMetadata) ([]byte, error) {
	report := NewReport(issues, metadata)

	if r.Indent {
		return json.MarshalIndent(report, "", "  ")
	}
	return json.Marshal(report)
}

// init registers the JSON reporter
func init() {
	RegisterReporter(NewJSONReporter(true))
}
