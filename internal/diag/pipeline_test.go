package diag

import "testing"

func TestRegisteredAnalyzerCount(t *testing.T) {
	count := RegisteredAnalyzerCount()
	if count < 10 {
		t.Errorf("RegisteredAnalyzerCount() = %d, want >= 10 (70+ analyzers expected)", count)
	}
	t.Logf("registered analyzers: %d", count)
}
