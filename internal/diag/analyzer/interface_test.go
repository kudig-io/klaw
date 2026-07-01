package analyzer

import (
	"context"
	"testing"

	"github.com/kudig-io/klaw/internal/diag/types"
)

// mockAnalyzer is a test double for Analyzer interface
type mockAnalyzer struct {
	BaseAnalyzer
}

func newMockAnalyzer() *mockAnalyzer {
	return &mockAnalyzer{
		BaseAnalyzer: *NewBaseAnalyzer(
			"mock.analyzer",
			"Mock analyzer for testing",
			"test",
			[]types.DataMode{types.ModeOffline, types.ModeOnline},
		),
	}
}

func (m *mockAnalyzer) Analyze(_ context.Context, _ *types.DiagnosticData) ([]types.Issue, error) {
	return []types.Issue{
		{
			Severity: types.SeverityInfo,
			CNName:   "测试问题",
			ENName:   "MOCK_ISSUE",
		},
	}, nil
}

func TestNewBaseAnalyzer(t *testing.T) {
	modes := []types.DataMode{types.ModeOffline}
	ba := NewBaseAnalyzer("test.name", "Test description", "test", modes)

	if ba.name != "test.name" {
		t.Errorf("name = %v, want %v", ba.name, "test.name")
	}
	if ba.description != "Test description" {
		t.Errorf("description = %v, want %v", ba.description, "Test description")
	}
	if ba.category != "test" {
		t.Errorf("category = %v, want %v", ba.category, "test")
	}
	if len(ba.modes) != 1 || ba.modes[0] != types.ModeOffline {
		t.Error("modes not set correctly")
	}
}

func TestBaseAnalyzerName(t *testing.T) {
	ba := NewBaseAnalyzer("test.name", "Description", "category", nil)
	if got := ba.Name(); got != "test.name" {
		t.Errorf("Name() = %v, want %v", got, "test.name")
	}
}

func TestBaseAnalyzerDescription(t *testing.T) {
	ba := NewBaseAnalyzer("test.name", "Test description", "category", nil)
	if got := ba.Description(); got != "Test description" {
		t.Errorf("Description() = %v, want %v", got, "Test description")
	}
}

func TestBaseAnalyzerCategory(t *testing.T) {
	ba := NewBaseAnalyzer("test.name", "Description", "test-category", nil)
	if got := ba.Category(); got != "test-category" {
		t.Errorf("Category() = %v, want %v", got, "test-category")
	}
}

func TestBaseAnalyzerSupportedModes(t *testing.T) {
	modes := []types.DataMode{types.ModeOffline, types.ModeOnline}
	ba := NewBaseAnalyzer("test.name", "Description", "category", modes)

	got := ba.SupportedModes()
	if len(got) != 2 {
		t.Errorf("SupportedModes() length = %v, want %v", len(got), 2)
	}
}

func TestBaseAnalyzerDependencies(t *testing.T) {
	ba := NewBaseAnalyzer("test.name", "Description", "category", nil)

	// Initially empty
	if len(ba.Dependencies()) != 0 {
		t.Error("Dependencies should be empty initially")
	}

	// Add dependencies
	ba.deps = []string{"dep1", "dep2"}
	if len(ba.Dependencies()) != 2 {
		t.Errorf("Dependencies length = %v, want %v", len(ba.Dependencies()), 2)
	}
}

func TestBaseAnalyzerWithDependencies(t *testing.T) {
	ba := NewBaseAnalyzer("test.name", "Description", "category", nil)
	ba.deps = []string{"analyzer.a", "analyzer.b"}

	deps := ba.Dependencies()
	if len(deps) != 2 {
		t.Errorf("len(Dependencies) = %v, want %v", len(deps), 2)
	}
}

func TestMockAnalyzer(t *testing.T) {
	mock := newMockAnalyzer()

	if mock.Name() != "mock.analyzer" {
		t.Errorf("Name() = %v, want %v", mock.Name(), "mock.analyzer")
	}

	ctx := context.Background()
	data := &types.DiagnosticData{}

	issues, err := mock.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}

	if len(issues) != 1 {
		t.Errorf("len(issues) = %v, want %v", len(issues), 1)
	}

	if issues[0].ENName != "MOCK_ISSUE" {
		t.Errorf("ENName = %v, want %v", issues[0].ENName, "MOCK_ISSUE")
	}
}
