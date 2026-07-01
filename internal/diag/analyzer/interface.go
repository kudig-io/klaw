// Package analyzer defines the analysis engine layer interfaces
package analyzer

import (
	"context"

	"github.com/kudig-io/klaw/internal/diag/types"
)

// Analyzer is the interface that all analyzers must implement
type Analyzer interface {
	// Name returns the unique identifier of the analyzer (e.g., "system.cpu")
	Name() string

	// Description returns a human-readable description of what this analyzer checks
	Description() string

	// Category returns the category of this analyzer (e.g., "system", "network", "kubernetes")
	Category() string

	// Analyze performs the analysis and returns detected issues
	Analyze(ctx context.Context, data *types.DiagnosticData) ([]types.Issue, error)

	// SupportedModes returns the data modes this analyzer supports
	SupportedModes() []types.DataMode

	// Dependencies returns the names of analyzers that must run before this one
	Dependencies() []string
}

// BaseAnalyzer provides common functionality for analyzers
type BaseAnalyzer struct {
	name        string
	description string
	category    string
	modes       []types.DataMode
	deps        []string
}

// NewBaseAnalyzer creates a new BaseAnalyzer
func NewBaseAnalyzer(name, description, category string, modes []types.DataMode) *BaseAnalyzer {
	return &BaseAnalyzer{
		name:        name,
		description: description,
		category:    category,
		modes:       modes,
		deps:        []string{},
	}
}

// Name returns the analyzer name
func (b *BaseAnalyzer) Name() string {
	return b.name
}

// Description returns the analyzer description
func (b *BaseAnalyzer) Description() string {
	return b.description
}

// Category returns the analyzer category
func (b *BaseAnalyzer) Category() string {
	return b.category
}

// SupportedModes returns supported data modes
func (b *BaseAnalyzer) SupportedModes() []types.DataMode {
	return b.modes
}

// Dependencies returns analyzer dependencies
func (b *BaseAnalyzer) Dependencies() []string {
	return b.deps
}

// SupportsMode checks if the analyzer supports a specific mode
func (b *BaseAnalyzer) SupportsMode(mode types.DataMode) bool {
	for _, m := range b.modes {
		if m == mode {
			return true
		}
	}
	return false
}

// Config holds configuration for analyzers.
type Config struct {
	// Enabled indicates if the analyzer is enabled
	Enabled bool `json:"enabled" yaml:"enabled"`

	// Thresholds holds configurable threshold values
	Thresholds map[string]interface{} `json:"thresholds,omitempty" yaml:"thresholds,omitempty"`

	// ExtraConfig holds additional configuration
	ExtraConfig map[string]interface{} `json:"extra_config,omitempty" yaml:"extra_config,omitempty"`
}

// Result wraps the result of an analyzer execution.
type Result struct {
	// AnalyzerName is the name of the analyzer
	AnalyzerName string `json:"analyzer_name" yaml:"analyzer_name"`

	// Issues contains the detected issues
	Issues []types.Issue `json:"issues" yaml:"issues"`

	// Error contains any error that occurred during analysis
	Error error `json:"error,omitempty" yaml:"error,omitempty"`

	// Duration is how long the analysis took
	DurationMs int64 `json:"duration_ms" yaml:"duration_ms"`
}
