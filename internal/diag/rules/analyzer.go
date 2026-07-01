package rules

import (
	"context"

	"github.com/kudig-io/klaw/internal/diag/analyzer"
	"github.com/kudig-io/klaw/internal/diag/types"
)

// RuleAnalyzer wraps the rule engine as an analyzer
type RuleAnalyzer struct {
	*analyzer.BaseAnalyzer
	engine *Engine
}

// NewRuleAnalyzer creates a new rule-based analyzer
func NewRuleAnalyzer(loader *Loader) *RuleAnalyzer {
	return &RuleAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer(
			"rules.custom",
			"自定义规则分析器",
			"rules",
			[]types.DataMode{types.ModeOffline, types.ModeOnline},
		),
		engine: NewEngine(loader),
	}
}

// Analyze runs all rules against the diagnostic data
func (a *RuleAnalyzer) Analyze(ctx context.Context, data *types.DiagnosticData) ([]types.Issue, error) {
	return a.engine.Evaluate(ctx, data)
}

// DefaultRuleAnalyzer creates a rule analyzer with built-in rules
func DefaultRuleAnalyzer() *RuleAnalyzer {
	loader := NewLoader()
	loader.LoadBuiltin()
	return NewRuleAnalyzer(loader)
}

// RuleAnalyzerFromDir creates a rule analyzer from a directory of rule files
func RuleAnalyzerFromDir(dir string) (*RuleAnalyzer, error) {
	loader := NewLoader()
	if err := loader.LoadDir(dir); err != nil {
		return nil, err
	}
	return NewRuleAnalyzer(loader), nil
}

// RuleAnalyzerFromFile creates a rule analyzer from a single rule file
func RuleAnalyzerFromFile(file string) (*RuleAnalyzer, error) {
	loader := NewLoader()
	if err := loader.LoadFile(file); err != nil {
		return nil, err
	}
	return NewRuleAnalyzer(loader), nil
}
