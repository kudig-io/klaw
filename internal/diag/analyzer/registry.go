// Package analyzer provides the analysis engine
package analyzer

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/kudig-io/klaw/internal/diag/types"
)

// Registry manages all registered analyzers
type Registry struct {
	analyzers map[string]Analyzer
	mu        sync.RWMutex
}

// DefaultRegistry is the global default analyzer registry
var DefaultRegistry = NewRegistry()

// NewRegistry creates a new analyzer registry
func NewRegistry() *Registry {
	return &Registry{
		analyzers: make(map[string]Analyzer),
	}
}

// Register adds an analyzer to the registry
func (r *Registry) Register(analyzer Analyzer) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := analyzer.Name()
	if _, exists := r.analyzers[name]; exists {
		return fmt.Errorf("analyzer %q already registered", name)
	}

	r.analyzers[name] = analyzer
	return nil
}

// Get retrieves an analyzer by name
func (r *Registry) Get(name string) (Analyzer, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	analyzer, ok := r.analyzers[name]
	return analyzer, ok
}

// List returns all registered analyzers
func (r *Registry) List() []Analyzer {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]Analyzer, 0, len(r.analyzers))
	for _, a := range r.analyzers {
		result = append(result, a)
	}
	return result
}

// ListByCategory returns analyzers filtered by category
func (r *Registry) ListByCategory(category string) []Analyzer {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []Analyzer
	for _, a := range r.analyzers {
		if a.Category() == category {
			result = append(result, a)
		}
	}
	return result
}

// ListByMode returns analyzers that support a specific mode
func (r *Registry) ListByMode(mode types.DataMode) []Analyzer {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []Analyzer
	for _, a := range r.analyzers {
		for _, m := range a.SupportedModes() {
			if m == mode {
				result = append(result, a)
				break
			}
		}
	}
	return result
}

// ExecuteAll runs all registered analyzers and returns results
func (r *Registry) ExecuteAll(ctx context.Context, data *types.DiagnosticData) ([]Result, error) {
	analyzers := r.ListByMode(data.Mode)
	return r.executeAnalyzers(ctx, analyzers, data)
}

// ExecuteByCategory runs analyzers of a specific category
func (r *Registry) ExecuteByCategory(
	ctx context.Context,
	category string,
	data *types.DiagnosticData,
) ([]Result, error) {
	analyzers := r.ListByCategory(category)
	return r.executeAnalyzers(ctx, analyzers, data)
}

// ExecuteByMode runs analyzers that support a specific mode
func (r *Registry) ExecuteByMode(
	ctx context.Context,
	data *types.DiagnosticData,
	mode types.DataMode,
) ([]Result, error) {
	analyzers := r.ListByMode(mode)
	return r.executeAnalyzers(ctx, analyzers, data)
}

// ExecuteByNames runs specific analyzers by name
func (r *Registry) ExecuteByNames(
	ctx context.Context,
	names []string,
	data *types.DiagnosticData,
) ([]Result, error) {
	r.mu.RLock()
	var analyzers []Analyzer
	for _, name := range names {
		if a, ok := r.analyzers[name]; ok {
			analyzers = append(analyzers, a)
		}
	}
	r.mu.RUnlock()

	return r.executeAnalyzers(ctx, analyzers, data)
}

// executeAnalyzers runs a list of analyzers
func (r *Registry) executeAnalyzers(
	ctx context.Context,
	analyzers []Analyzer,
	data *types.DiagnosticData,
) ([]Result, error) {
	// Sort by dependencies (topological sort)
	sorted, err := r.sortByDependencies(analyzers)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve dependencies: %w", err)
	}

	results := make([]Result, 0, len(sorted))

	for _, analyzer := range sorted {
		select {
		case <-ctx.Done():
			return results, ctx.Err()
		default:
		}

		result := r.executeOne(ctx, analyzer, data)
		results = append(results, result)
	}

	return results, nil
}

// executeOne runs a single analyzer
func (r *Registry) executeOne(ctx context.Context, analyzer Analyzer, data *types.DiagnosticData) Result {
	start := time.Now()

	issues, err := analyzer.Analyze(ctx, data)

	return Result{
		AnalyzerName: analyzer.Name(),
		Issues:       issues,
		Error:        err,
		DurationMs:   time.Since(start).Milliseconds(),
	}
}

// sortByDependencies performs topological sort based on dependencies
func (r *Registry) sortByDependencies(analyzers []Analyzer) ([]Analyzer, error) {
	// Build dependency graph
	nameToAnalyzer := make(map[string]Analyzer)
	for _, a := range analyzers {
		nameToAnalyzer[a.Name()] = a
	}

	// Simple topological sort
	visited := make(map[string]bool)
	visiting := make(map[string]bool)
	var sorted []Analyzer

	var visit func(name string) error
	visit = func(name string) error {
		if visited[name] {
			return nil
		}
		if visiting[name] {
			return fmt.Errorf("circular dependency detected at %q", name)
		}

		analyzer, ok := nameToAnalyzer[name]
		if !ok {
			// Dependency not in our list, skip
			return nil
		}

		visiting[name] = true
		for _, dep := range analyzer.Dependencies() {
			if err := visit(dep); err != nil {
				return err
			}
		}
		visiting[name] = false
		visited[name] = true
		sorted = append(sorted, analyzer)
		return nil
	}

	for _, a := range analyzers {
		if err := visit(a.Name()); err != nil {
			return nil, err
		}
	}

	return sorted, nil
}

// Register registers an analyzer with the default registry
func Register(analyzer Analyzer) error {
	return DefaultRegistry.Register(analyzer)
}

// MustRegister registers analyzers and panics on error.
// 供 init() 使用：重名等错误属于程序缺陷，应快速失败而非静默丢弃。
func MustRegister(analyzers ...Analyzer) {
	for _, a := range analyzers {
		if err := DefaultRegistry.Register(a); err != nil {
			panic(err)
		}
	}
}

// CollectIssues extracts all issues from analyzer results
func CollectIssues(results []Result) []types.Issue {
	var issues []types.Issue
	for _, r := range results {
		if r.Error == nil && len(r.Issues) > 0 {
			issues = append(issues, r.Issues...)
		}
	}
	return issues
}
