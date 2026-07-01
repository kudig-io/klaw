package reporter

import (
	"testing"
	"time"

	"github.com/kudig-io/klaw/internal/diag/types"
)

func TestNewReportMetadata(t *testing.T) {
	meta := NewReportMetadata()

	if meta.ReportVersion != "2.0" {
		t.Errorf("ReportVersion = %v, want %v", meta.ReportVersion, "2.0")
	}
	if meta.Engine != "go" {
		t.Errorf("Engine = %v, want %v", meta.Engine, "go")
	}
	if meta.Timestamp.IsZero() {
		t.Error("Timestamp should be set")
	}
}

func TestNewReport(t *testing.T) {
	meta := NewReportMetadata()
	meta.Hostname = "test-node"
	meta.DiagnosePath = "/tmp/test"

	issues := []types.Issue{
		{Severity: types.SeverityCritical, CNName: "严重问题"},
		{Severity: types.SeverityWarning, CNName: "警告问题"},
		{Severity: types.SeverityInfo, CNName: "提示信息"},
	}

	report := NewReport(issues, meta)

	if report.Hostname != "test-node" {
		t.Errorf("Hostname = %v, want %v", report.Hostname, "test-node")
	}
	if report.DiagnoseDir != "/tmp/test" {
		t.Errorf("DiagnoseDir = %v, want %v", report.DiagnoseDir, "/tmp/test")
	}
	if len(report.Anomalies) != 3 {
		t.Errorf("Anomalies count = %v, want %v", len(report.Anomalies), 3)
	}
	if report.Summary.Critical != 1 {
		t.Errorf("Summary.Critical = %v, want %v", report.Summary.Critical, 1)
	}
	if report.Summary.Total != 3 {
		t.Errorf("Summary.Total = %v, want %v", report.Summary.Total, 3)
	}
}

func TestReporterFactory(t *testing.T) {
	factory := NewReporterFactory()

	// Create a mock reporter
	mockReporter := &mockReporter{format: "test"}
	factory.Register(mockReporter)

	// Test Get
	got, ok := factory.Get("test")
	if !ok {
		t.Error("Expected to find reporter")
	}
	if got.Format() != "test" {
		t.Errorf("Format = %v, want %v", got.Format(), "test")
	}

	// Test List
	formats := factory.List()
	if len(formats) != 1 {
		t.Errorf("List length = %v, want %v", len(formats), 1)
	}

	// Test not found
	_, ok = factory.Get("nonexistent")
	if ok {
		t.Error("Should not find nonexistent reporter")
	}
}

func TestDefaultFactory(t *testing.T) {
	mockReporter := &mockReporter{format: "mock"}
	RegisterReporter(mockReporter)

	got, ok := GetReporter("mock")
	if !ok {
		t.Error("Expected to find reporter in default factory")
	}
	if got.Format() != "mock" {
		t.Errorf("Format = %v, want %v", got.Format(), "mock")
	}
}

// mockReporter is a test double
type mockReporter struct {
	format string
}

func (m *mockReporter) Generate(_ []types.Issue, _ *ReportMetadata) ([]byte, error) {
	return []byte("mock report"), nil
}

func (m *mockReporter) Format() string {
	return m.format
}

func TestReportTimestamp(t *testing.T) {
	meta := NewReportMetadata()
	meta.Timestamp = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	report := NewReport(nil, meta)

	expectedTimestamp := "2024-01-01T00:00:00Z"
	if report.Timestamp != expectedTimestamp {
		t.Errorf("Timestamp = %v, want %v", report.Timestamp, expectedTimestamp)
	}
}
