package loganalysis

import "testing"

func TestAnalyzeLogs(t *testing.T) {
	logs := `2026-05-23T10:00:00Z INFO service started
2026-05-23T10:00:01Z WARN retry connection refused
2026-05-23T10:00:02Z ERROR unauthorized request
goroutine 10 [running]:
/api/v1/users GET 820ms`

	analysis := NewAnalyzer().AnalyzeLogs(logs)

	if analysis.TotalLines != 5 {
		t.Fatalf("TotalLines = %d, want 5", analysis.TotalLines)
	}
	if analysis.ErrorCount != 1 {
		t.Fatalf("ErrorCount = %d, want 1", analysis.ErrorCount)
	}
	if analysis.WarningCount != 1 {
		t.Fatalf("WarningCount = %d, want 1", analysis.WarningCount)
	}
	if len(analysis.SecurityEvents) != 1 {
		t.Fatalf("SecurityEvents = %d, want 1", len(analysis.SecurityEvents))
	}
	if len(analysis.StackTraces) == 0 {
		t.Fatal("StackTraces should not be empty")
	}
	if len(analysis.PerformanceMetrics.SlowRequests) != 1 {
		t.Fatalf("SlowRequests = %d, want 1", len(analysis.PerformanceMetrics.SlowRequests))
	}
}
