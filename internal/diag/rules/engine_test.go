package rules

import (
	"context"
	"testing"

	"github.com/kudig-io/klaw/internal/diag/types"
)

func TestNewEngine(t *testing.T) {
	loader := NewLoader()
	engine := NewEngine(loader)
	if engine == nil {
		t.Fatal("NewEngine() returned nil")
	}
}

func TestEngineEvaluate_EmptyRules(t *testing.T) {
	loader := NewLoader()
	engine := NewEngine(loader)
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)

	issues, err := engine.Evaluate(ctx, data)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues, got %d", len(issues))
	}
}

func TestEngineEvaluate_FileContains(t *testing.T) {
	loader := NewLoader()
	engine := NewEngine(loader)

	// Add a rule
	loader.ruleSets = append(loader.ruleSets, &RuleSet{
		Rules: []Rule{
			{
				ID:       "TEST_RULE",
				Name:     "Test Rule",
				Category: "test",
				Severity: "warning",
				Enabled:  true,
				Condition: Condition{
					Type:    "file_contains",
					File:    "test.log",
					Pattern: "error",
				},
			},
		},
	})

	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"test.log": []byte("this is an error message"),
	}

	issues, err := engine.Evaluate(ctx, data)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].ENName != "TEST_RULE" {
		t.Errorf("ENName = %v, want %v", issues[0].ENName, "TEST_RULE")
	}
}

func TestEngineEvaluate_FileNotFound(t *testing.T) {
	loader := NewLoader()
	engine := NewEngine(loader)

	loader.ruleSets = append(loader.ruleSets, &RuleSet{
		Rules: []Rule{
			{
				ID:       "TEST_RULE",
				Name:     "Test Rule",
				Category: "test",
				Severity: "warning",
				Enabled:  true,
				Condition: Condition{
					Type:    "file_contains",
					File:    "missing.log",
					Pattern: "error",
				},
			},
		},
	})

	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)

	issues, err := engine.Evaluate(ctx, data)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues, got %d", len(issues))
	}
}

func TestEngineEvaluate_DisabledRule(t *testing.T) {
	loader := NewLoader()
	engine := NewEngine(loader)

	loader.ruleSets = append(loader.ruleSets, &RuleSet{
		Rules: []Rule{
			{
				ID:       "DISABLED_RULE",
				Name:     "Disabled Rule",
				Category: "test",
				Severity: "warning",
				Enabled:  false, // Disabled
				Condition: Condition{
					Type:    "file_contains",
					File:    "test.log",
					Pattern: "error",
				},
			},
		},
	})

	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"test.log": []byte("this is an error message"),
	}

	// Engine only evaluates enabled rules via GetAllRules
	// GetAllRules filters by enabled status
	issues, err := engine.Evaluate(ctx, data)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	// Should have 0 issues because rule is disabled
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues for disabled rule, got %d", len(issues))
	}
}

func TestEngineEvaluate_ContextCancel(t *testing.T) {
	loader := NewLoader()
	engine := NewEngine(loader)

	// Add many rules to ensure context cancellation happens
	var rules []Rule
	for i := 0; i < 100; i++ {
		rules = append(rules, Rule{
			ID:       "RULE_" + string(rune(i)),
			Name:     "Rule",
			Category: "test",
			Severity: "warning",
			Enabled:  true,
			Condition: Condition{
				Type:    "file_contains",
				File:    "test.log",
				Pattern: "error",
			},
		})
	}
	loader.ruleSets = append(loader.ruleSets, &RuleSet{Rules: rules})

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"test.log": []byte("error"),
	}

	_, err := engine.Evaluate(ctx, data)
	if err == nil {
		t.Error("Expected error for cancelled context")
	}
}

func TestEngineEvaluate_MetricThreshold(t *testing.T) {
	loader := NewLoader()
	engine := NewEngine(loader)

	loader.ruleSets = append(loader.ruleSets, &RuleSet{
		Rules: []Rule{
			{
				ID:       "HIGH_LOAD",
				Name:     "High Load",
				Category: "system",
				Severity: "warning",
				Enabled:  true,
				Condition: Condition{
					Type:      "metric_threshold",
					Metric:    "load_avg_1min",
					Operator:  "gt",
					Threshold: 0.8,
				},
			},
		},
	})

	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.SystemMetrics = &types.SystemMetrics{
		CPUCores:     4,
		LoadAvg1Min:  5.0, // High load (5/4 = 1.25 > 0.8)
		LoadAvg5Min:  2.0,
		LoadAvg15Min: 1.0,
	}

	issues, err := engine.Evaluate(ctx, data)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].ENName != "HIGH_LOAD" {
		t.Errorf("ENName = %v, want %v", issues[0].ENName, "HIGH_LOAD")
	}
}

func TestEngineEvaluate_MetricNotMatched(t *testing.T) {
	loader := NewLoader()
	engine := NewEngine(loader)

	loader.ruleSets = append(loader.ruleSets, &RuleSet{
		Rules: []Rule{
			{
				ID:       "HIGH_LOAD",
				Name:     "High Load",
				Category: "system",
				Severity: "warning",
				Enabled:  true,
				Condition: Condition{
					Type:      "metric_threshold",
					Metric:    "load_avg_1min",
					Operator:  "gt",
					Threshold: 0.8,
				},
			},
		},
	})

	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.SystemMetrics = &types.SystemMetrics{
		CPUCores:     4,
		LoadAvg1Min:  1.0, // Low load (1/4 = 0.25 < 0.8)
		LoadAvg5Min:  1.0,
		LoadAvg15Min: 1.0,
	}

	issues, err := engine.Evaluate(ctx, data)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues, got %d", len(issues))
	}
}

// ============ EvaluateByCategory Tests ============

func TestEngineEvaluateByCategory(t *testing.T) {
	loader := NewLoader()
	engine := NewEngine(loader)

	// Add rules in different categories
	loader.ruleSets = append(loader.ruleSets, &RuleSet{
		Rules: []Rule{
			{
				ID:       "SYSTEM_RULE",
				Name:     "System Rule",
				Category: "system",
				Severity: "warning",
				Enabled:  true,
				Condition: Condition{
					Type:    "file_contains",
					File:    "test.log",
					Pattern: "error",
				},
			},
			{
				ID:       "NETWORK_RULE",
				Name:     "Network Rule",
				Category: "network",
				Severity: "warning",
				Enabled:  true,
				Condition: Condition{
					Type:    "file_contains",
					File:    "test.log",
					Pattern: "error",
				},
			},
		},
	})

	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"test.log": []byte("error message"),
	}

	// Evaluate only system category
	issues, err := engine.EvaluateByCategory(ctx, "system", data)
	if err != nil {
		t.Fatalf("EvaluateByCategory() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].ENName != "SYSTEM_RULE" {
		t.Errorf("ENName = %v, want %v", issues[0].ENName, "SYSTEM_RULE")
	}
}

func TestEngineEvaluateByCategory_NoRules(t *testing.T) {
	loader := NewLoader()
	engine := NewEngine(loader)

	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)

	issues, err := engine.EvaluateByCategory(ctx, "nonexistent", data)
	if err != nil {
		t.Fatalf("EvaluateByCategory() error = %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues, got %d", len(issues))
	}
}

func TestEngineEvaluateByCategory_ContextCancel(t *testing.T) {
	loader := NewLoader()
	engine := NewEngine(loader)

	// Add many rules
	var rules []Rule
	for i := 0; i < 100; i++ {
		rules = append(rules, Rule{
			ID:       "RULE_" + string(rune(i)),
			Name:     "Rule",
			Category: "test",
			Severity: "warning",
			Enabled:  true,
			Condition: Condition{
				Type:    "file_contains",
				File:    "test.log",
				Pattern: "error",
			},
		})
	}
	loader.ruleSets = append(loader.ruleSets, &RuleSet{Rules: rules})

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"test.log": []byte("error"),
	}

	_, err := engine.EvaluateByCategory(ctx, "test", data)
	if err == nil {
		t.Error("Expected error for cancelled context")
	}
}

// ============ Regex Match Tests ============

func TestEngineEvaluate_RegexMatch(t *testing.T) {
	loader := NewLoader()
	engine := NewEngine(loader)

	loader.ruleSets = append(loader.ruleSets, &RuleSet{
		Rules: []Rule{
			{
				ID:       "REGEX_RULE",
				Name:     "Regex Rule",
				Category: "test",
				Severity: "warning",
				Enabled:  true,
				Condition: Condition{
					Type:    "regex_match",
					File:    "test.log",
					Pattern: `error\s+(\d+)`,
				},
			},
		},
	})

	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"test.log": []byte("error 500 occurred"),
	}

	issues, err := engine.Evaluate(ctx, data)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].ENName != "REGEX_RULE" {
		t.Errorf("ENName = %v, want %v", issues[0].ENName, "REGEX_RULE")
	}
}

func TestEngineEvaluate_RegexMatch_NoMatch(t *testing.T) {
	loader := NewLoader()
	engine := NewEngine(loader)

	loader.ruleSets = append(loader.ruleSets, &RuleSet{
		Rules: []Rule{
			{
				ID:       "REGEX_RULE",
				Name:     "Regex Rule",
				Category: "test",
				Severity: "warning",
				Enabled:  true,
				Condition: Condition{
					Type:    "regex_match",
					File:    "test.log",
					Pattern: `error\s+(\d+)`,
				},
			},
		},
	})

	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"test.log": []byte("success message"),
	}

	issues, err := engine.Evaluate(ctx, data)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues, got %d", len(issues))
	}
}

func TestEngineEvaluate_RegexMatch_InvalidRegex(t *testing.T) {
	loader := NewLoader()
	engine := NewEngine(loader)

	loader.ruleSets = append(loader.ruleSets, &RuleSet{
		Rules: []Rule{
			{
				ID:       "REGEX_RULE",
				Name:     "Regex Rule",
				Category: "test",
				Severity: "warning",
				Enabled:  true,
				Condition: Condition{
					Type:    "regex_match",
					File:    "test.log",
					Pattern: `[invalid`,
				},
			},
		},
	})

	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"test.log": []byte("test content"),
	}

	// Invalid regex should not match but not panic
	issues, err := engine.Evaluate(ctx, data)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues for invalid regex, got %d", len(issues))
	}
}

func TestEngineEvaluate_RegexMatch_FileNotFound(t *testing.T) {
	loader := NewLoader()
	engine := NewEngine(loader)

	loader.ruleSets = append(loader.ruleSets, &RuleSet{
		Rules: []Rule{
			{
				ID:       "REGEX_RULE",
				Name:     "Regex Rule",
				Category: "test",
				Severity: "warning",
				Enabled:  true,
				Condition: Condition{
					Type:    "regex_match",
					File:    "missing.log",
					Pattern: `error`,
				},
			},
		},
	})

	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)

	issues, err := engine.Evaluate(ctx, data)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues, got %d", len(issues))
	}
}

// ============ AND Condition Tests ============

func TestEngineEvaluate_AndCondition(t *testing.T) {
	loader := NewLoader()
	engine := NewEngine(loader)

	loader.ruleSets = append(loader.ruleSets, &RuleSet{
		Rules: []Rule{
			{
				ID:       "AND_RULE",
				Name:     "And Rule",
				Category: "test",
				Severity: "warning",
				Enabled:  true,
				Condition: Condition{
					Type: "and",
					And: []Condition{
						{
							Type:    "file_contains",
							File:    "test.log",
							Pattern: "error",
						},
						{
							Type:    "file_contains",
							File:    "test.log",
							Pattern: "critical",
						},
					},
				},
			},
		},
	})

	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"test.log": []byte("error: critical failure"),
	}

	issues, err := engine.Evaluate(ctx, data)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].ENName != "AND_RULE" {
		t.Errorf("ENName = %v, want %v", issues[0].ENName, "AND_RULE")
	}
}

func TestEngineEvaluate_AndCondition_FirstFails(t *testing.T) {
	loader := NewLoader()
	engine := NewEngine(loader)

	loader.ruleSets = append(loader.ruleSets, &RuleSet{
		Rules: []Rule{
			{
				ID:       "AND_RULE",
				Name:     "And Rule",
				Category: "test",
				Severity: "warning",
				Enabled:  true,
				Condition: Condition{
					Type: "and",
					And: []Condition{
						{
							Type:    "file_contains",
							File:    "test.log",
							Pattern: "missing",
						},
						{
							Type:    "file_contains",
							File:    "test.log",
							Pattern: "error",
						},
					},
				},
			},
		},
	})

	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"test.log": []byte("error message"),
	}

	issues, err := engine.Evaluate(ctx, data)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues when first AND condition fails, got %d", len(issues))
	}
}

func TestEngineEvaluate_AndCondition_SecondFails(t *testing.T) {
	loader := NewLoader()
	engine := NewEngine(loader)

	loader.ruleSets = append(loader.ruleSets, &RuleSet{
		Rules: []Rule{
			{
				ID:       "AND_RULE",
				Name:     "And Rule",
				Category: "test",
				Severity: "warning",
				Enabled:  true,
				Condition: Condition{
					Type: "and",
					And: []Condition{
						{
							Type:    "file_contains",
							File:    "test.log",
							Pattern: "error",
						},
						{
							Type:    "file_contains",
							File:    "test.log",
							Pattern: "missing",
						},
					},
				},
			},
		},
	})

	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"test.log": []byte("error message"),
	}

	issues, err := engine.Evaluate(ctx, data)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues when second AND condition fails, got %d", len(issues))
	}
}

// ============ OR Condition Tests ============

func TestEngineEvaluate_OrCondition(t *testing.T) {
	loader := NewLoader()
	engine := NewEngine(loader)

	loader.ruleSets = append(loader.ruleSets, &RuleSet{
		Rules: []Rule{
			{
				ID:       "OR_RULE",
				Name:     "Or Rule",
				Category: "test",
				Severity: "warning",
				Enabled:  true,
				Condition: Condition{
					Type: "or",
					Or: []Condition{
						{
							Type:    "file_contains",
							File:    "test.log",
							Pattern: "error",
						},
						{
							Type:    "file_contains",
							File:    "test.log",
							Pattern: "warning",
						},
					},
				},
			},
		},
	})

	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"test.log": []byte("error message"),
	}

	issues, err := engine.Evaluate(ctx, data)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].ENName != "OR_RULE" {
		t.Errorf("ENName = %v, want %v", issues[0].ENName, "OR_RULE")
	}
}

func TestEngineEvaluate_OrCondition_SecondMatches(t *testing.T) {
	loader := NewLoader()
	engine := NewEngine(loader)

	loader.ruleSets = append(loader.ruleSets, &RuleSet{
		Rules: []Rule{
			{
				ID:       "OR_RULE",
				Name:     "Or Rule",
				Category: "test",
				Severity: "warning",
				Enabled:  true,
				Condition: Condition{
					Type: "or",
					Or: []Condition{
						{
							Type:    "file_contains",
							File:    "test.log",
							Pattern: "missing",
						},
						{
							Type:    "file_contains",
							File:    "test.log",
							Pattern: "error",
						},
					},
				},
			},
		},
	})

	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"test.log": []byte("error message"),
	}

	issues, err := engine.Evaluate(ctx, data)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
}

func TestEngineEvaluate_OrCondition_NoneMatch(t *testing.T) {
	loader := NewLoader()
	engine := NewEngine(loader)

	loader.ruleSets = append(loader.ruleSets, &RuleSet{
		Rules: []Rule{
			{
				ID:       "OR_RULE",
				Name:     "Or Rule",
				Category: "test",
				Severity: "warning",
				Enabled:  true,
				Condition: Condition{
					Type: "or",
					Or: []Condition{
						{
							Type:    "file_contains",
							File:    "test.log",
							Pattern: "missing1",
						},
						{
							Type:    "file_contains",
							File:    "test.log",
							Pattern: "missing2",
						},
					},
				},
			},
		},
	})

	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"test.log": []byte("some message"),
	}

	issues, err := engine.Evaluate(ctx, data)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues when no OR condition matches, got %d", len(issues))
	}
}

// ============ Metric Threshold Extended Tests ============

func TestEngineEvaluate_MetricThreshold_NoMetrics(t *testing.T) {
	loader := NewLoader()
	engine := NewEngine(loader)

	loader.ruleSets = append(loader.ruleSets, &RuleSet{
		Rules: []Rule{
			{
				ID:       "METRIC_RULE",
				Name:     "Metric Rule",
				Category: "system",
				Severity: "warning",
				Enabled:  true,
				Condition: Condition{
					Type:      "metric_threshold",
					Metric:    "load_avg_1min",
					Operator:  "gt",
					Threshold: 0.8,
				},
			},
		},
	})

	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	// No SystemMetrics set

	issues, err := engine.Evaluate(ctx, data)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues without metrics, got %d", len(issues))
	}
}

func TestEngineEvaluate_MetricThreshold_AllOperators(t *testing.T) {
	tests := []struct {
		name      string
		operator  string
		loadValue float64
		threshold float64
		wantMatch bool
	}{
		{"gt_true", "gt", 5.0, 0.8, true},
		{"gt_false", "gt", 1.0, 0.8, false},
		{"gte_true", "gte", 3.2, 0.8, true},
		{"lt_true", "lt", 1.0, 0.8, true},
		{"lt_false", "lt", 5.0, 0.8, false},
		{"lte_true", "lte", 3.2, 0.8, true},
		{"eq_true", "eq", 3.2, 0.8, true},  // 3.2/4 = 0.8
		{"eq_false", "eq", 5.0, 0.8, false},
		{"ne_true", "ne", 5.0, 0.8, true},
		{"ne_false", "ne", 3.2, 0.8, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := NewLoader()
			engine := NewEngine(loader)

			loader.ruleSets = append(loader.ruleSets, &RuleSet{
				Rules: []Rule{
					{
						ID:       "METRIC_RULE",
						Name:     "Metric Rule",
						Category: "system",
						Severity: "warning",
						Enabled:  true,
						Condition: Condition{
							Type:      "metric_threshold",
							Metric:    "load_avg_1min",
							Operator:  tt.operator,
							Threshold: tt.threshold,
						},
					},
				},
			})

			ctx := context.Background()
			data := types.NewDiagnosticData(types.ModeOffline)
			data.SystemMetrics = &types.SystemMetrics{
				CPUCores:    4,
				LoadAvg1Min: tt.loadValue,
			}

			issues, err := engine.Evaluate(ctx, data)
			if err != nil {
				t.Fatalf("Evaluate() error = %v", err)
			}

			gotMatch := len(issues) > 0
			if gotMatch != tt.wantMatch {
				t.Errorf("got match = %v, want %v", gotMatch, tt.wantMatch)
			}
		})
	}
}

func TestEngineEvaluate_MetricThreshold_UnknownOperator(t *testing.T) {
	loader := NewLoader()
	engine := NewEngine(loader)

	loader.ruleSets = append(loader.ruleSets, &RuleSet{
		Rules: []Rule{
			{
				ID:       "METRIC_RULE",
				Name:     "Metric Rule",
				Category: "system",
				Severity: "warning",
				Enabled:  true,
				Condition: Condition{
					Type:      "metric_threshold",
					Metric:    "load_avg_1min",
					Operator:  "unknown",
					Threshold: 0.8,
				},
			},
		},
	})

	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.SystemMetrics = &types.SystemMetrics{
		CPUCores:    4,
		LoadAvg1Min: 5.0,
	}

	// Unknown operator should not panic but not match
	issues, err := engine.Evaluate(ctx, data)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues for unknown operator, got %d", len(issues))
	}
}

func TestEngineEvaluate_MetricThreshold_AllMetrics(t *testing.T) {
	tests := []struct {
		name      string
		metric    string
		setupData func() *types.DiagnosticData
		threshold float64
		operator  string
		wantMatch bool
	}{
		{
			name:   "load_avg_5min",
			metric: "load_avg_5min",
			setupData: func() *types.DiagnosticData {
				data := types.NewDiagnosticData(types.ModeOffline)
				data.SystemMetrics = &types.SystemMetrics{
					CPUCores:    4,
					LoadAvg5Min: 5.0, // 5/4 = 1.25
				}
				return data
			},
			threshold: 1.0,
			operator:  "gt",
			wantMatch: true,
		},
		{
			name:   "load_avg_15min",
			metric: "load_avg_15min",
			setupData: func() *types.DiagnosticData {
				data := types.NewDiagnosticData(types.ModeOffline)
				data.SystemMetrics = &types.SystemMetrics{
					CPUCores:     4,
					LoadAvg15Min: 5.0, // 5/4 = 1.25
				}
				return data
			},
			threshold: 1.0,
			operator:  "gt",
			wantMatch: true,
		},
		{
			name:   "mem_used_percent",
			metric: "mem_used_percent",
			setupData: func() *types.DiagnosticData {
				data := types.NewDiagnosticData(types.ModeOffline)
				data.SystemMetrics = &types.SystemMetrics{
					MemTotal:     10000,
					MemAvailable: 2000, // 80% used
				}
				return data
			},
			threshold: 75.0,
			operator:  "gt",
			wantMatch: true,
		},
		{
			name:   "mem_available_percent",
			metric: "mem_available_percent",
			setupData: func() *types.DiagnosticData {
				data := types.NewDiagnosticData(types.ModeOffline)
				data.SystemMetrics = &types.SystemMetrics{
					MemTotal:     10000,
					MemAvailable: 2000, // 20% available
				}
				return data
			},
			threshold: 25.0,
			operator:  "lt",
			wantMatch: true,
		},
		{
			name:   "swap_used_percent",
			metric: "swap_used_percent",
			setupData: func() *types.DiagnosticData {
				data := types.NewDiagnosticData(types.ModeOffline)
				data.SystemMetrics = &types.SystemMetrics{
					SwapTotal: 10000,
					SwapFree:  5000, // 50% used
				}
				return data
			},
			threshold: 40.0,
			operator:  "gt",
			wantMatch: true,
		},
		{
			name:   "disk_used_percent",
			metric: "disk_used_percent",
			setupData: func() *types.DiagnosticData {
				data := types.NewDiagnosticData(types.ModeOffline)
				data.SystemMetrics = &types.SystemMetrics{
					DiskUsage: []types.DiskUsage{
						{MountPoint: "/", UsedPercent: 85.0},
					},
				}
				return data
			},
			threshold: 80.0,
			operator:  "gt",
			wantMatch: true,
		},
		{
			name:   "conntrack_percent",
			metric: "conntrack_percent",
			setupData: func() *types.DiagnosticData {
				data := types.NewDiagnosticData(types.ModeOffline)
				data.SystemMetrics = &types.SystemMetrics{
					ConntrackCurrent: 55000,
					ConntrackMax:     65536, // ~84%
				}
				return data
			},
			threshold: 80.0,
			operator:  "gt",
			wantMatch: true,
		},
		{
			name:   "unknown_metric",
			metric: "unknown_metric",
			setupData: func() *types.DiagnosticData {
				data := types.NewDiagnosticData(types.ModeOffline)
				data.SystemMetrics = &types.SystemMetrics{}
				return data
			},
			threshold: 0.0,
			operator:  "gt",
			wantMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := NewLoader()
			engine := NewEngine(loader)

			loader.ruleSets = append(loader.ruleSets, &RuleSet{
				Rules: []Rule{
					{
						ID:       "METRIC_RULE",
						Name:     "Metric Rule",
						Category: "system",
						Severity: "warning",
						Enabled:  true,
						Condition: Condition{
							Type:      "metric_threshold",
							Metric:    tt.metric,
							Operator:  tt.operator,
							Threshold: tt.threshold,
						},
					},
				},
			})

			ctx := context.Background()
			data := tt.setupData()

			issues, err := engine.Evaluate(ctx, data)
			if err != nil {
				t.Fatalf("Evaluate() error = %v", err)
			}

			gotMatch := len(issues) > 0
			if gotMatch != tt.wantMatch {
				t.Errorf("got match = %v, want %v", gotMatch, tt.wantMatch)
			}
		})
	}
}

// ============ File Contains Extended Tests ============

func TestEngineEvaluate_FileContains_CaseInsensitive(t *testing.T) {
	loader := NewLoader()
	engine := NewEngine(loader)

	loader.ruleSets = append(loader.ruleSets, &RuleSet{
		Rules: []Rule{
			{
				ID:       "CASE_RULE",
				Name:     "Case Rule",
				Category: "test",
				Severity: "warning",
				Enabled:  true,
				Condition: Condition{
					Type:    "file_contains",
					File:    "test.log",
					Pattern: "ERROR",
				},
			},
		},
	})

	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"test.log": []byte("this is an error message"), // lowercase
	}

	issues, err := engine.Evaluate(ctx, data)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue for case-insensitive match, got %d", len(issues))
	}
}

func TestEngineEvaluate_FileContains_WithCount(t *testing.T) {
	loader := NewLoader()
	engine := NewEngine(loader)

	loader.ruleSets = append(loader.ruleSets, &RuleSet{
		Rules: []Rule{
			{
				ID:       "COUNT_RULE",
				Name:     "Count Rule",
				Category: "test",
				Severity: "warning",
				Enabled:  true,
				Condition: Condition{
					Type:    "file_contains",
					File:    "test.log",
					Pattern: "error",
					Count:   3,
				},
			},
		},
	})

	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"test.log": []byte("error1 error2 error3 error4"), // 4 occurrences
	}

	issues, err := engine.Evaluate(ctx, data)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
}

func TestEngineEvaluate_FileContains_CountNotMet(t *testing.T) {
	loader := NewLoader()
	engine := NewEngine(loader)

	loader.ruleSets = append(loader.ruleSets, &RuleSet{
		Rules: []Rule{
			{
				ID:       "COUNT_RULE",
				Name:     "Count Rule",
				Category: "test",
				Severity: "warning",
				Enabled:  true,
				Condition: Condition{
					Type:    "file_contains",
					File:    "test.log",
					Pattern: "error",
					Count:   5,
				},
			},
		},
	})

	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"test.log": []byte("error1 error2"), // Only 2 occurrences
	}

	issues, err := engine.Evaluate(ctx, data)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues when count not met, got %d", len(issues))
	}
}

func TestEngineEvaluate_FileContains_InvalidRegex(t *testing.T) {
	loader := NewLoader()
	engine := NewEngine(loader)

	loader.ruleSets = append(loader.ruleSets, &RuleSet{
		Rules: []Rule{
			{
				ID:       "REGEX_RULE",
				Name:     "Regex Rule",
				Category: "test",
				Severity: "warning",
				Enabled:  true,
				Condition: Condition{
					Type:    "file_contains",
					File:    "test.log",
					Pattern: `[invalid(regex`, // Invalid regex
				},
			},
		},
	})

	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"test.log": []byte("[invalid(regex in content"),
	}

	// Invalid regex should fall back to simple contains
	issues, err := engine.Evaluate(ctx, data)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	// Should match using simple contains
	if len(issues) != 1 {
		t.Errorf("Expected 1 issue with fallback to simple contains, got %d", len(issues))
	}
}

// ============ Unknown Condition Type Tests ============

func TestEngineEvaluate_UnknownConditionType(t *testing.T) {
	loader := NewLoader()
	engine := NewEngine(loader)

	loader.ruleSets = append(loader.ruleSets, &RuleSet{
		Rules: []Rule{
			{
				ID:       "UNKNOWN_RULE",
				Name:     "Unknown Rule",
				Category: "test",
				Severity: "warning",
				Enabled:  true,
				Condition: Condition{
					Type: "unknown_type",
				},
			},
		},
	})

	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)

	// Unknown condition type should not panic
	issues, err := engine.Evaluate(ctx, data)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues for unknown condition type, got %d", len(issues))
	}
}

// ============ Truncate String Tests ============

func TestTruncateString(t *testing.T) {
	tests := []struct {
		input   string
		maxLen  int
		want    string
	}{
		{"short", 100, "short"},
		{"a very long string that needs to be truncated", 10, "a very lon..."},
		{"exactly ten", 11, "exactly ten"},
		{"", 10, ""},
		{"abc", 3, "abc"},
		{"abcd", 3, "abc..."},
	}

	for _, tt := range tests {
		got := truncateString(tt.input, tt.maxLen)
		if got != tt.want {
			t.Errorf("truncateString(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
		}
	}
}

