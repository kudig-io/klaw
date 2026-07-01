package cost

import (
	"context"
	"testing"

	"github.com/kudig-io/klaw/internal/diag/types"
)

func TestNewCostAnalyzer(t *testing.T) {
	a := NewCostAnalyzer()
	if a == nil {
		t.Fatal("Expected non-nil analyzer")
	}
	if a.CPUPricePerCore != 0.05 {
		t.Errorf("Expected CPU price 0.05, got %f", a.CPUPricePerCore)
	}
	if a.MemoryPricePerGB != 0.01 {
		t.Errorf("Expected memory price 0.01, got %f", a.MemoryPricePerGB)
	}
}

func TestCostAnalyzer_Analyze_NoData(t *testing.T) {
	a := NewCostAnalyzer()
	data := &types.DiagnosticData{}
	result, err := a.Analyze(context.Background(), data)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
	if len(result.Resources) != 0 {
		t.Errorf("Expected 0 resources without system metrics, got %d", len(result.Resources))
	}
}

func TestCostAnalyzer_Analyze_WithSystemMetrics(t *testing.T) {
	a := NewCostAnalyzer()
	data := &types.DiagnosticData{
		NodeInfo: types.NodeInfo{Hostname: "test-node"},
		SystemMetrics: &types.SystemMetrics{
			CPUCores:    8,
			LoadAvg1Min: 4.5,
			MemTotal:    32 * 1024 * 1024, // 32 GB in KB
		},
	}
	result, err := a.Analyze(context.Background(), data)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(result.Resources) != 1 {
		t.Fatalf("Expected 1 resource, got %d", len(result.Resources))
	}
	r := result.Resources[0]
	if r.Name != "test-node" {
		t.Errorf("Expected name 'test-node', got '%s'", r.Name)
	}
	if r.Type != "node" {
		t.Errorf("Expected type 'node', got '%s'", r.Type)
	}
	if r.DailyCost <= 0 {
		t.Errorf("Expected positive daily cost, got %f", r.DailyCost)
	}
	if r.MonthlyCost != r.DailyCost*30 {
		t.Errorf("Monthly cost calculation mismatch")
	}
	if r.YearlyCost != r.DailyCost*365 {
		t.Errorf("Yearly cost calculation mismatch")
	}
	if result.TotalDailyCost != r.DailyCost {
		t.Errorf("Total daily cost mismatch")
	}
}

func TestCostAnalyzer_Analyze_Recommendations(t *testing.T) {
	a := NewCostAnalyzer()
	data := &types.DiagnosticData{
		NodeInfo: types.NodeInfo{Hostname: "test-node"},
		SystemMetrics: &types.SystemMetrics{
			LoadAvg1Min: 4.0,
			MemTotal:    16 * 1024 * 1024,
		},
	}
	result, err := a.Analyze(context.Background(), data)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(result.Recommendations) == 0 {
		t.Error("Expected recommendations to be generated")
	}
}

func TestFormatResult(t *testing.T) {
	result := &AnalysisResult{
		Resources: []ResourceCost{
			{
				Name:       "test-node",
				Type:       "node",
				CPUCores:   4.0,
				MemoryGB:   16.0,
				DailyCost:  0.36,
			},
		},
		TotalDailyCost:   0.36,
		TotalMonthlyCost: 10.80,
		TotalYearlyCost:  131.40,
		Recommendations:  []string{"建议启用自动伸缩以优化成本"},
	}
	formatted := FormatResult(result)
	if formatted == "" {
		t.Error("Expected non-empty formatted result")
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"short", 10, "short"},
		{"this is a very long string", 10, "this is..."},
	}
	for _, tc := range tests {
		result := truncate(tc.input, tc.maxLen)
		if result != tc.expected {
			t.Errorf("truncate(%q, %d) = %q, expected %q", tc.input, tc.maxLen, result, tc.expected)
		}
	}
}
