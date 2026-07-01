package analyzer

import (
	"context"
	"fmt"
	"testing"

	"github.com/kudig-io/klaw/internal/diag/types"
)

func TestNewRegistry(t *testing.T) {
	r := NewRegistry()
	if r == nil {
		t.Fatal("NewRegistry() returned nil")
	}
	if len(r.List()) != 0 {
		t.Error("New registry should be empty")
	}
}

func TestRegistryRegister(t *testing.T) {
	r := NewRegistry()
	mock := newMockAnalyzer()

	err := r.Register(mock)
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	if len(r.List()) != 1 {
		t.Errorf("List() length = %v, want %v", len(r.List()), 1)
	}
}

func TestRegistryRegisterDuplicate(t *testing.T) {
	r := NewRegistry()
	mock := newMockAnalyzer()

	if err := r.Register(mock); err != nil {
		t.Fatalf("First Register() error = %v", err)
	}

	// Try to register again
	err := r.Register(mock)
	if err == nil {
		t.Error("Register() should return error for duplicate")
	}
}

func TestRegistryGet(t *testing.T) {
	r := NewRegistry()
	mock := newMockAnalyzer()
	r.Register(mock)

	got, ok := r.Get("mock.analyzer")
	if !ok {
		t.Error("Get() should find registered analyzer")
	}
	if got.Name() != "mock.analyzer" {
		t.Errorf("Name = %v, want %v", got.Name(), "mock.analyzer")
	}

	_, ok = r.Get("nonexistent")
	if ok {
		t.Error("Get() should not find nonexistent analyzer")
	}
}

func TestRegistryList(t *testing.T) {
	r := NewRegistry()

	// Register multiple analyzers
	r.Register(newMockAnalyzer())
	r.Register(&mockAnalyzer2{})

	list := r.List()
	if len(list) != 2 {
		t.Errorf("List() length = %v, want %v", len(list), 2)
	}
}

func TestRegistryListByCategory(t *testing.T) {
	r := NewRegistry()

	// Register analyzers with different categories
	r.Register(newMockAnalyzer()) // category: "test"
	r.Register(&mockAnalyzer2{})  // category: "test2"

	testAnalyzers := r.ListByCategory("test")
	if len(testAnalyzers) != 1 {
		t.Errorf("ListByCategory('test') length = %v, want %v", len(testAnalyzers), 1)
	}

	test2Analyzers := r.ListByCategory("test2")
	if len(test2Analyzers) != 1 {
		t.Errorf("ListByCategory('test2') length = %v, want %v", len(test2Analyzers), 1)
	}

	noneAnalyzers := r.ListByCategory("nonexistent")
	if len(noneAnalyzers) != 0 {
		t.Errorf("ListByCategory('nonexistent') length = %v, want %v", len(noneAnalyzers), 0)
	}
}

func TestRegistryListByMode(t *testing.T) {
	r := NewRegistry()
	r.Register(newMockAnalyzer()) // supports both modes
	r.Register(&mockAnalyzer2{})  // supports only offline

	offlineAnalyzers := r.ListByMode(types.ModeOffline)
	if len(offlineAnalyzers) != 2 {
		t.Errorf("ListByMode(Offline) length = %v, want %v", len(offlineAnalyzers), 2)
	}

	onlineAnalyzers := r.ListByMode(types.ModeOnline)
	if len(onlineAnalyzers) != 1 {
		t.Errorf("ListByMode(Online) length = %v, want %v", len(onlineAnalyzers), 1)
	}
}

func TestCollectIssues(t *testing.T) {
	results := []Result{
		{
			AnalyzerName: "analyzer1",
			Issues: []types.Issue{
				{Severity: types.SeverityCritical, ENName: "ISSUE1"},
			},
		},
		{
			AnalyzerName: "analyzer2",
			Issues: []types.Issue{
				{Severity: types.SeverityWarning, ENName: "ISSUE2"},
				{Severity: types.SeverityInfo, ENName: "ISSUE3"},
			},
		},
	}

	issues := CollectIssues(results)
	if len(issues) != 3 {
		t.Errorf("CollectIssues length = %v, want %v", len(issues), 3)
	}
}

func TestCollectIssuesEmpty(t *testing.T) {
	results := []Result{}
	issues := CollectIssues(results)
	if len(issues) != 0 {
		t.Errorf("CollectIssues length = %v, want %v", len(issues), 0)
	}
}

func TestCollectIssuesSkipsErrors(t *testing.T) {
	results := []Result{
		{
			AnalyzerName: "ok.analyzer",
			Issues: []types.Issue{
				{Severity: types.SeverityCritical, ENName: "ISSUE1"},
			},
		},
		{
			AnalyzerName: "err.analyzer",
			Error:        fmt.Errorf("analyzer failed"),
			Issues: []types.Issue{
				{Severity: types.SeverityWarning, ENName: "SHOULD_BE_SKIPPED"},
			},
		},
	}
	issues := CollectIssues(results)
	if len(issues) != 1 {
		t.Fatalf("CollectIssues length = %v, want %v", len(issues), 1)
	}
	if issues[0].ENName != "ISSUE1" {
		t.Errorf("ENName = %v, want %v", issues[0].ENName, "ISSUE1")
	}
}

func TestRegistryExecuteAll(t *testing.T) {
	r := NewRegistry()
	r.Register(newMockAnalyzer())
	r.Register(&mockAnalyzer2{})

	data := &types.DiagnosticData{Mode: types.ModeOffline}
	results, err := r.ExecuteAll(context.Background(), data)
	if err != nil {
		t.Fatalf("ExecuteAll() error = %v", err)
	}

	names := make(map[string]bool)
	for _, res := range results {
		names[res.AnalyzerName] = true
	}
	if !names["mock.analyzer"] {
		t.Error("ExecuteAll should include mock.analyzer")
	}
	if !names["mock.analyzer2"] {
		t.Error("ExecuteAll should include mock.analyzer2")
	}
}

func TestRegistryExecuteAllOnlineFilter(t *testing.T) {
	r := NewRegistry()
	r.Register(newMockAnalyzer())  // supports both modes
	r.Register(&mockAnalyzer2{})   // supports only offline

	data := &types.DiagnosticData{Mode: types.ModeOnline}
	results, err := r.ExecuteAll(context.Background(), data)
	if err != nil {
		t.Fatalf("ExecuteAll() error = %v", err)
	}
	if len(results) != 1 {
		t.Errorf("ExecuteAll(Online) should return 1 result, got %d", len(results))
	}
}

func TestRegistryExecuteByMode(t *testing.T) {
	r := NewRegistry()
	r.Register(newMockAnalyzer())
	r.Register(&mockAnalyzer2{})

	results, err := r.ExecuteByMode(context.Background(), &types.DiagnosticData{}, types.ModeOffline)
	if err != nil {
		t.Fatalf("ExecuteByMode() error = %v", err)
	}
	if len(results) != 2 {
		t.Errorf("ExecuteByMode(Offline) should return 2 results, got %d", len(results))
	}
}

func TestRegistryExecuteByCategory(t *testing.T) {
	r := NewRegistry()
	r.Register(newMockAnalyzer())
	r.Register(&mockAnalyzer2{})

	results, err := r.ExecuteByCategory(context.Background(), "test", &types.DiagnosticData{})
	if err != nil {
		t.Fatalf("ExecuteByCategory() error = %v", err)
	}
	if len(results) != 1 {
		t.Errorf("ExecuteByCategory('test') should return 1 result, got %d", len(results))
	}
}

func TestRegistryExecuteByNames(t *testing.T) {
	r := NewRegistry()
	r.Register(newMockAnalyzer())
	r.Register(&mockAnalyzer2{})

	results, err := r.ExecuteByNames(context.Background(), []string{"mock.analyzer"}, &types.DiagnosticData{})
	if err != nil {
		t.Fatalf("ExecuteByNames() error = %v", err)
	}
	if len(results) != 1 {
		t.Errorf("ExecuteByNames should return 1 result, got %d", len(results))
	}
	if results[0].AnalyzerName != "mock.analyzer" {
		t.Errorf("AnalyzerName = %v, want mock.analyzer", results[0].AnalyzerName)
	}
}

func TestRegistryExecuteByNamesUnknown(t *testing.T) {
	r := NewRegistry()
	r.Register(newMockAnalyzer())

	results, err := r.ExecuteByNames(context.Background(), []string{"nonexistent"}, &types.DiagnosticData{})
	if err != nil {
		t.Fatalf("ExecuteByNames() error = %v", err)
	}
	if len(results) != 0 {
		t.Errorf("ExecuteByNames should return 0 results for unknown names, got %d", len(results))
	}
}

func TestRegistryExecuteCancelled(t *testing.T) {
	r := NewRegistry()
	r.Register(newMockAnalyzer())
	r.Register(&mockAnalyzer2{})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := r.ExecuteAll(ctx, &types.DiagnosticData{Mode: types.ModeOffline})
	if err == nil {
		t.Error("ExecuteAll with cancelled context should return error")
	}
}

func TestRegistrySortByDependencies(t *testing.T) {
	r := NewRegistry()
	dep := &depAnalyzer{name: "base"}
	child := &depAnalyzer{name: "child", deps: []string{"base"}}
	r.Register(dep)
	r.Register(child)

	data := &types.DiagnosticData{Mode: types.ModeOffline}
	results, err := r.ExecuteAll(context.Background(), data)
	if err != nil {
		t.Fatalf("ExecuteAll() with deps error = %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("ExecuteAll should return 2 results, got %d", len(results))
	}
	if results[0].AnalyzerName != "base" {
		t.Errorf("First result should be 'base', got %v", results[0].AnalyzerName)
	}
}

func TestRegistrySortByCircularDependency(t *testing.T) {
	r := NewRegistry()
	a := &depAnalyzer{name: "a", deps: []string{"b"}}
	b := &depAnalyzer{name: "b", deps: []string{"a"}}
	r.Register(a)
	r.Register(b)

	_, err := r.ExecuteAll(context.Background(), &types.DiagnosticData{Mode: types.ModeOffline})
	if err == nil {
		t.Error("ExecuteAll with circular deps should return error")
	}
}

// mockAnalyzer2 is another test double with different category
type mockAnalyzer2 struct {
	BaseAnalyzer
}

func (m *mockAnalyzer2) Name() string        { return "mock.analyzer2" }
func (m *mockAnalyzer2) Description() string { return "Mock analyzer 2" }
func (m *mockAnalyzer2) Category() string    { return "test2" }
func (m *mockAnalyzer2) Analyze(_ context.Context, _ *types.DiagnosticData) ([]types.Issue, error) {
	return nil, nil
}
func (m *mockAnalyzer2) SupportedModes() []types.DataMode { return []types.DataMode{types.ModeOffline} }
func (m *mockAnalyzer2) Dependencies() []string           { return nil }

func (m *mockAnalyzer2) initBase() {
	m.BaseAnalyzer = *NewBaseAnalyzer(
		"mock.analyzer2",
		"Mock analyzer 2",
		"test2",
		[]types.DataMode{types.ModeOffline},
	)
}

// depAnalyzer is a mock that supports dependency ordering tests
type depAnalyzer struct {
	BaseAnalyzer
	name string
	deps []string
}

func (d *depAnalyzer) Name() string        { return d.name }
func (d *depAnalyzer) Description() string { return d.name }
func (d *depAnalyzer) Category() string    { return "dep-test" }
func (d *depAnalyzer) Analyze(_ context.Context, _ *types.DiagnosticData) ([]types.Issue, error) {
	return nil, nil
}
func (d *depAnalyzer) SupportedModes() []types.DataMode {
	return []types.DataMode{types.ModeOffline, types.ModeOnline}
}
func (d *depAnalyzer) Dependencies() []string { return d.deps }
