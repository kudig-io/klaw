package diag

import (
	"context"
	"fmt"

	"github.com/kudig-io/klaw/internal/diag/ai"
	"github.com/kudig-io/klaw/internal/diag/analyzer"
	"github.com/kudig-io/klaw/internal/diag/collector"
	"github.com/kudig-io/klaw/internal/diag/collector/online"
	"github.com/kudig-io/klaw/internal/diag/types"
)

type DiagnosisRequest struct {
	Kubeconfig       string   `json:"kubeconfig,omitempty"`
	Context          string   `json:"context,omitempty"`
	NodeName         string   `json:"nodeName,omitempty"`
	Namespace        string   `json:"namespace,omitempty"`
	Analyzers        []string `json:"analyzers,omitempty"`
	ExcludeAnalyzers []string `json:"excludeAnalyzers,omitempty"`
	// DisableAI 关闭 AI 摘要分析；默认在检测到 KUDIG_AI_API_KEY 时自动启用
	DisableAI bool `json:"disableAI,omitempty"`
}

type DiagnosisResult struct {
	Data    *types.DiagnosticData `json:"data"`
	Results []analyzer.Result     `json:"results"`
	Issues  []types.Issue         `json:"issues"`
	// AIAnalysis 为 LLM 对诊断结果的自然语言归纳，仅在配置了 AI provider 且分析成功时非空
	AIAnalysis *ai.AnalysisResult `json:"ai_analysis,omitempty"`
}

func RunOnlineDiagnostics(ctx context.Context, req DiagnosisRequest) (*DiagnosisResult, error) {
	c := online.NewCollector()
	cfg := &collector.Config{
		Kubeconfig: req.Kubeconfig,
		Context:    req.Context,
		NodeName:   req.NodeName,
		Namespace:  req.Namespace,
	}
	if err := c.Validate(cfg); err != nil {
		return nil, err
	}
	data, err := c.Collect(ctx, cfg)
	if err != nil {
		return nil, err
	}

	var results []analyzer.Result
	if len(req.Analyzers) > 0 {
		results, err = analyzer.DefaultRegistry.ExecuteByNames(ctx, req.Analyzers, data)
	} else {
		results, err = analyzer.DefaultRegistry.ExecuteAll(ctx, data)
	}
	if err != nil {
		return nil, err
	}

	issues := analyzer.CollectIssues(results)
	if len(req.ExcludeAnalyzers) > 0 {
		exclude := make(map[string]bool, len(req.ExcludeAnalyzers))
		for _, n := range req.ExcludeAnalyzers {
			exclude[n] = true
		}
		filtered := issues[:0]
		for _, iss := range issues {
			if !exclude[iss.AnalyzerName] {
				filtered = append(filtered, iss)
			}
		}
		issues = filtered
	}

	return &DiagnosisResult{
		Data:       data,
		Results:    results,
		Issues:     issues,
		AIAnalysis: runAIAnalysis(ctx, req.DisableAI, issues, data.NodeName),
	}, nil
}

// runAIAnalysis 尽力而为的 AI 摘要：未配置 provider（无 API Key）、被禁用或无问题时跳过，
// 调用失败仅告警不影响诊断主流程。
func runAIAnalysis(ctx context.Context, disabled bool, issues []types.Issue, hostname string) *ai.AnalysisResult {
	if disabled || len(issues) == 0 {
		return nil
	}
	assistant, err := ai.NewAssistant(nil)
	if err != nil || !assistant.IsAvailable() {
		return nil
	}
	analysis, err := assistant.AnalyzeWithAI(ctx, issues, hostname)
	if err != nil {
		fmt.Printf("⚠ AI 分析失败（诊断结果不受影响）: %v\n", err)
		return nil
	}
	return analysis
}

func RegisteredAnalyzerCount() int {
	return len(analyzer.DefaultRegistry.List())
}
