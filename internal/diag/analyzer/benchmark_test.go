package analyzer

import (
	"context"
	"testing"

	"github.com/kudig-io/klaw/internal/diag/types"
)

func BenchmarkExecuteAll(b *testing.B) {
	r := NewRegistry()
	// Register 10 mock analyzers
	for i := 0; i < 10; i++ {
		r.Register(&benchAnalyzer{name: string(rune('a' + i))})
	}

	data := &types.DiagnosticData{Mode: types.ModeOffline}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := r.ExecuteAll(context.Background(), data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkExecuteAllParallel(b *testing.B) {
	r := NewRegistry()
	for i := 0; i < 10; i++ {
		r.Register(&benchAnalyzer{name: string(rune('a' + i))})
	}

	data := &types.DiagnosticData{Mode: types.ModeOffline}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := r.ExecuteAll(context.Background(), data)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkCollectIssues(b *testing.B) {
	results := make([]Result, 100)
	for i := range results {
		issues := make([]types.Issue, 10)
		for j := range issues {
			issues[j] = types.Issue{
				Severity: types.SeverityWarning,
				ENName:   "TEST_ISSUE",
				CNName:   "测试问题",
				Details:  "benchmark test issue",
			}
		}
		results[i] = Result{
			AnalyzerName: "bench",
			Issues:       issues,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CollectIssues(results)
	}
}

func BenchmarkSortByDependencies(b *testing.B) {
	r := NewRegistry()
	// Create a chain: a -> b -> c -> d -> e
	r.Register(&benchAnalyzer{name: "e"})
	r.Register(&benchAnalyzer{name: "d", deps: []string{"e"}})
	r.Register(&benchAnalyzer{name: "c", deps: []string{"d"}})
	r.Register(&benchAnalyzer{name: "b", deps: []string{"c"}})
	r.Register(&benchAnalyzer{name: "a", deps: []string{"b"}})

	data := &types.DiagnosticData{Mode: types.ModeOffline}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := r.ExecuteAll(context.Background(), data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

type benchAnalyzer struct {
	name string
	deps []string
}

func (b *benchAnalyzer) Name() string        { return b.name }
func (b *benchAnalyzer) Description() string { return "benchmark analyzer" }
func (b *benchAnalyzer) Category() string    { return "bench" }
func (b *benchAnalyzer) Analyze(_ context.Context, _ *types.DiagnosticData) ([]types.Issue, error) {
	return nil, nil
}
func (b *benchAnalyzer) SupportedModes() []types.DataMode {
	return []types.DataMode{types.ModeOffline, types.ModeOnline}
}
func (b *benchAnalyzer) Dependencies() []string { return b.deps }
