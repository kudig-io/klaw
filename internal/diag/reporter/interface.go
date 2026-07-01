// Package reporter defines the reporting layer interfaces
package reporter

import (
	"time"

	"github.com/kudig-io/klaw/internal/diag/types"
)

// Reporter is the interface for report generators
type Reporter interface {
	// Generate creates a report from the issues
	Generate(issues []types.Issue, metadata *ReportMetadata) ([]byte, error)

	// Format returns the output format name (e.g., "json", "text", "yaml")
	Format() string
}

// ReportMetadata contains metadata for the report
type ReportMetadata struct {
	// ReportVersion is the version of the report format
	ReportVersion string `json:"report_version" yaml:"report_version"`

	// Timestamp when the report was generated
	Timestamp time.Time `json:"timestamp" yaml:"timestamp"`

	// Hostname of the diagnosed node
	Hostname string `json:"hostname" yaml:"hostname"`

	// DiagnosePath is the path to diagnostic directory
	DiagnosePath string `json:"diagnose_dir" yaml:"diagnose_dir"`

	// Mode indicates the analysis mode used
	Mode string `json:"mode,omitempty" yaml:"mode,omitempty"`

	// Engine indicates the analysis engine (go or bash)
	Engine string `json:"engine,omitempty" yaml:"engine,omitempty"`

	// Summary contains issue statistics
	Summary types.IssueSummary `json:"summary" yaml:"summary"`
}

// NewReportMetadata creates a new ReportMetadata with defaults
func NewReportMetadata() *ReportMetadata {
	return &ReportMetadata{
		ReportVersion: "2.0",
		Timestamp:     time.Now(),
		Engine:        "go",
	}
}

// Report is the complete diagnostic report
type Report struct {
	// Metadata contains report metadata
	Metadata *ReportMetadata `json:"metadata,omitempty" yaml:"metadata,omitempty"`

	// Legacy fields for backward compatibility with klaw.sh JSON output
	ReportVersion string             `json:"report_version" yaml:"report_version"`
	Timestamp     string             `json:"timestamp" yaml:"timestamp"`
	Hostname      string             `json:"hostname" yaml:"hostname"`
	DiagnoseDir   string             `json:"diagnose_dir" yaml:"diagnose_dir"`
	Anomalies     []types.Issue      `json:"anomalies" yaml:"anomalies"`
	Summary       types.IssueSummary `json:"summary" yaml:"summary"`
}

// NewReport creates a new Report from issues and metadata
func NewReport(issues []types.Issue, metadata *ReportMetadata) *Report {
	summary := types.CalculateSummary(issues)

	return &Report{
		Metadata:      metadata,
		ReportVersion: metadata.ReportVersion,
		Timestamp:     metadata.Timestamp.UTC().Format(time.RFC3339),
		Hostname:      metadata.Hostname,
		DiagnoseDir:   metadata.DiagnosePath,
		Anomalies:     issues,
		Summary:       summary,
	}
}

// ReporterFactory manages reporter instances
type ReporterFactory struct {
	reporters map[string]Reporter
}

// NewReporterFactory creates a new factory
func NewReporterFactory() *ReporterFactory {
	return &ReporterFactory{
		reporters: make(map[string]Reporter),
	}
}

// Register adds a reporter
func (f *ReporterFactory) Register(reporter Reporter) {
	f.reporters[reporter.Format()] = reporter
}

// Get returns a reporter by format
func (f *ReporterFactory) Get(format string) (Reporter, bool) {
	r, ok := f.reporters[format]
	return r, ok
}

// List returns all available formats
func (f *ReporterFactory) List() []string {
	formats := make([]string, 0, len(f.reporters))
	for format := range f.reporters {
		formats = append(formats, format)
	}
	return formats
}

// DefaultFactory is the global reporter factory
var DefaultFactory = NewReporterFactory()

// RegisterReporter registers a reporter with the default factory
func RegisterReporter(reporter Reporter) {
	DefaultFactory.Register(reporter)
}

// GetReporter gets a reporter from the default factory
func GetReporter(format string) (Reporter, bool) {
	return DefaultFactory.Get(format)
}
