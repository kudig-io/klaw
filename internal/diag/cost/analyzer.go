// Package cost provides Kubernetes resource cost analysis
package cost

import (
	"context"
	"fmt"
	"strings"

	"github.com/kudig-io/klaw/internal/diag/types"
)

// CostAnalyzer analyzes resource costs
type CostAnalyzer struct {
	// Pricing configuration (USD per unit)
	CPUPricePerCore     float64
	MemoryPricePerGB    float64
	StoragePricePerGB   float64
	NetworkPricePerGB   float64
}

// NewCostAnalyzer creates a new cost analyzer with default AWS pricing
func NewCostAnalyzer() *CostAnalyzer {
	return &CostAnalyzer{
		CPUPricePerCore:   0.05,  // $0.05 per core per day
		MemoryPricePerGB:  0.01,  // $0.01 per GB per day
		StoragePricePerGB: 0.001, // $0.001 per GB per day
		NetworkPricePerGB: 0.01,  // $0.01 per GB
	}
}

// ResourceCost represents the cost breakdown for a resource
type ResourceCost struct {
	Name           string  `json:"name"`
	Namespace      string  `json:"namespace,omitempty"`
	Type           string  `json:"type"`
	CPUCores       float64 `json:"cpu_cores"`
	MemoryGB       float64 `json:"memory_gb"`
	StorageGB      float64 `json:"storage_gb"`
	DailyCost      float64 `json:"daily_cost"`
	MonthlyCost    float64 `json:"monthly_cost"`
	YearlyCost     float64 `json:"yearly_cost"`
	Efficiency     float64 `json:"efficiency"` // actual usage / requested
}

// AnalysisResult contains the full cost analysis
type AnalysisResult struct {
	Resources      []ResourceCost `json:"resources"`
	TotalDailyCost float64        `json:"total_daily_cost"`
	TotalMonthlyCost float64      `json:"total_monthly_cost"`
	TotalYearlyCost  float64      `json:"total_yearly_cost"`
	Recommendations []string      `json:"recommendations"`
}

// Analyze performs cost analysis on diagnostic data
func (a *CostAnalyzer) Analyze(ctx context.Context, data *types.DiagnosticData) (*AnalysisResult, error) {
	result := &AnalysisResult{
		Resources:       make([]ResourceCost, 0),
		Recommendations: make([]string, 0),
	}

	// Analyze system resources
	if data.SystemMetrics != nil {
		systemCost := a.calculateSystemCost(data)
		result.Resources = append(result.Resources, systemCost)
	}

	// Calculate totals
	for _, r := range result.Resources {
		result.TotalDailyCost += r.DailyCost
		result.TotalMonthlyCost += r.MonthlyCost
		result.TotalYearlyCost += r.YearlyCost
	}

	// Generate recommendations
	result.Recommendations = a.generateRecommendations(result)

	return result, nil
}

// calculateSystemCost calculates cost for system resources
func (a *CostAnalyzer) calculateSystemCost(data *types.DiagnosticData) ResourceCost {
	memGB := float64(data.SystemMetrics.MemTotal) / (1024 * 1024) // KB to GB

	// Estimate CPU cores from load average
	cpuCores := data.SystemMetrics.LoadAvg1Min
	if cpuCores < 1 {
		cpuCores = 1
	}

	dailyCost := (cpuCores * a.CPUPricePerCore) +
		(memGB * a.MemoryPricePerGB)

	return ResourceCost{
		Name:        data.NodeInfo.Hostname,
		Type:        "node",
		CPUCores:    cpuCores,
		MemoryGB:    memGB,
		DailyCost:   dailyCost,
		MonthlyCost: dailyCost * 30,
		YearlyCost:  dailyCost * 365,
	}
}

// generateRecommendations generates cost optimization recommendations
func (a *CostAnalyzer) generateRecommendations(result *AnalysisResult) []string {
	recommendations := make([]string, 0)

	// Check for high costs
	if result.TotalMonthlyCost > 1000 {
		recommendations = append(recommendations,
			fmt.Sprintf("月度成本 $%.2f 较高，考虑使用预留实例或 Spot 实例", result.TotalMonthlyCost))
	}

	// Check for low efficiency
	for _, r := range result.Resources {
		if r.Efficiency > 0 && r.Efficiency < 0.5 {
			recommendations = append(recommendations,
				fmt.Sprintf("资源 %s 利用率仅 %.1f%%，考虑缩减资源配置", r.Name, r.Efficiency*100))
		}
	}

	// Generic recommendations
	if len(recommendations) == 0 {
		recommendations = append(recommendations,
			"资源成本在合理范围内",
			"建议启用自动伸缩以优化成本",
			"考虑使用 Spot/Preemptible 实例节省成本")
	}

	return recommendations
}

// FormatResult formats the cost analysis result for display
func FormatResult(result *AnalysisResult) string {
	var sb strings.Builder

	sb.WriteString("\n💰 成本分析报告\n")
	sb.WriteString("=" + strings.Repeat("=", 50) + "\n\n")

	// Resource breakdown
	sb.WriteString("资源成本明细:\n")
	sb.WriteString(fmt.Sprintf("%-20s %-10s %-10s %-12s %-12s\n", "名称", "类型", "CPU核", "内存(GB)", "日成本"))
	sb.WriteString(strings.Repeat("-", 70) + "\n")

	for _, r := range result.Resources {
		sb.WriteString(fmt.Sprintf("%-20s %-10s %-10.1f %-12.1f $%-11.2f\n",
			truncate(r.Name, 20), r.Type, r.CPUCores, r.MemoryGB, r.DailyCost))
	}

	sb.WriteString("\n")

	// Totals
	sb.WriteString("总成本估算:\n")
	sb.WriteString(fmt.Sprintf("  日成本:   $%.2f\n", result.TotalDailyCost))
	sb.WriteString(fmt.Sprintf("  月成本:   $%.2f\n", result.TotalMonthlyCost))
	sb.WriteString(fmt.Sprintf("  年成本:   $%.2f\n", result.TotalYearlyCost))
	sb.WriteString("\n")

	// Recommendations
	sb.WriteString("优化建议:\n")
	for i, rec := range result.Recommendations {
		sb.WriteString(fmt.Sprintf("  %d. %s\n", i+1, rec))
	}

	return sb.String()
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
