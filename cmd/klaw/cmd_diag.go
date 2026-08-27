package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/kudig-io/klaw/internal/diag"
	"github.com/kudig-io/klaw/internal/diag/ai"
	"github.com/kudig-io/klaw/internal/diag/types"
)

var (
	diagKubeconfig string
	diagContext    string
	diagNode       string
	diagNamespace  string
	diagJSON       bool
	diagNoAI       bool
	diagAnalyzers  []string
	diagExclude    []string
)

var diagCmd = &cobra.Command{
	Use:   "diag",
	Short: "对集群运行诊断分析 (70+ 分析器)",
	Long: `对 Kubernetes 集群运行深度诊断分析，涵盖内核、网络、存储、安全、服务网格等 9 大类 70+ 分析器。

示例:
  klaw diag                              # 诊断整个集群
  klaw diag --node worker-1              # 聚焦特定节点
  klaw diag --context prod-cluster       # 指定 kubeconfig context
  klaw diag --json                       # JSON 输出`,
	RunE: runDiag,
}

func init() {
	diagCmd.Flags().StringVar(&diagKubeconfig, "kubeconfig", "", "kubeconfig 路径 (默认 ~/.kube/config)")
	diagCmd.Flags().StringVar(&diagContext, "context", "", "kubeconfig context 名称")
	diagCmd.Flags().StringVar(&diagNode, "node", "", "聚焦特定节点")
	diagCmd.Flags().StringVar(&diagNamespace, "namespace", "", "聚焦特定命名空间")
	diagCmd.Flags().BoolVar(&diagJSON, "json", false, "JSON 格式输出")
	diagCmd.Flags().StringSliceVar(&diagAnalyzers, "analyzer", nil, "仅运行指定分析器 (逗号分隔)")
	diagCmd.Flags().StringSliceVar(&diagExclude, "exclude-analyzer", nil, "排除指定分析器 (逗号分隔)")
	diagCmd.Flags().BoolVar(&diagNoAI, "no-ai", false, "禁用 AI 摘要分析（配置了 KUDIG_AI_API_KEY 时默认启用）")
	rootCmd.AddCommand(diagCmd)
}

func runDiag(cmd *cobra.Command, _ []string) error {
	if diagKubeconfig == "" {
		if home, err := os.UserHomeDir(); err == nil {
			diagKubeconfig = home + "/.kube/config"
		}
	}

	fmt.Fprintf(cmd.ErrOrStderr(), "注册分析器数量: %d\n", diag.RegisteredAnalyzerCount())
	fmt.Fprintf(cmd.ErrOrStderr(), "开始诊断 (kubeconfig=%s, node=%s)...\n\n", diagKubeconfig, diagNode)

	result, err := diag.RunOnlineDiagnostics(context.Background(), diag.DiagnosisRequest{
		Kubeconfig:       diagKubeconfig,
		Context:          diagContext,
		NodeName:         diagNode,
		Namespace:        diagNamespace,
		Analyzers:        diagAnalyzers,
		ExcludeAnalyzers: diagExclude,
		DisableAI:        diagNoAI,
	})
	if err != nil {
		return fmt.Errorf("诊断失败: %w", err)
	}

	if diagJSON {
		printDiagJSON(cmd, result)
		return nil
	}
	printDiagText(cmd, result)
	return nil
}

func printDiagText(cmd *cobra.Command, result *diag.DiagnosisResult) {
	w := cmd.OutOrStdout()

	if result.Data != nil && result.Data.NodeName != "" {
		fmt.Fprintf(w, "节点: %s\n", result.Data.NodeName)
	}

	critical, warning, info := 0, 0, 0
	for _, issue := range result.Issues {
		switch issue.Severity {
		case types.SeverityCritical:
			critical++
		case types.SeverityWarning:
			warning++
		case types.SeverityInfo:
			info++
		}
	}

	fmt.Fprintf(w, "\n=== 诊断摘要 ===\n")
	fmt.Fprintf(w, "严重: %d  警告: %d  信息: %d  总计: %d\n", critical, warning, info, len(result.Issues))
	fmt.Fprintf(w, "分析器执行: %d\n\n", len(result.Results))

	if len(result.Issues) == 0 {
		fmt.Fprintf(w, "✅ 未发现问题\n")
		return
	}

	fmt.Fprintf(w, "=== 问题详情 ===\n")
	for _, issue := range result.Issues {
		fmt.Fprintf(w, "\n[%s] %s\n", issue.Severity.EnglishString(), issue.CNName)
		if issue.ENName != "" {
			fmt.Fprintf(w, "  标识: %s\n", issue.ENName)
		}
		if issue.Details != "" {
			fmt.Fprintf(w, "  描述: %s\n", issue.Details)
		}
		if issue.Location != "" {
			fmt.Fprintf(w, "  位置: %s\n", issue.Location)
		}
		if issue.Remediation != nil && issue.Remediation.Suggestion != "" {
			fmt.Fprintf(w, "  建议: %s\n", issue.Remediation.Suggestion)
		}
		if issue.AnalyzerName != "" {
			fmt.Fprintf(w, "  分析器: %s\n", issue.AnalyzerName)
		}
	}

	printAIAnalysis(cmd, result.AIAnalysis)
}

// printAIAnalysis 输出 AI 摘要段落（无 AI 结果时静默跳过）
func printAIAnalysis(cmd *cobra.Command, a *ai.AnalysisResult) {
	if a == nil {
		return
	}
	w := cmd.OutOrStdout()
	fmt.Fprintf(w, "\n=== AI 分析 ===\n")
	fmt.Fprintf(w, "摘要: %s\n", a.Summary)
	if a.RootCause != "" {
		fmt.Fprintf(w, "根因: %s\n", a.RootCause)
	}
	for _, s := range a.Suggestions {
		fmt.Fprintf(w, "建议: %s", s.Title)
		if s.Description != "" {
			fmt.Fprintf(w, " — %s", s.Description)
		}
		if s.Command != "" {
			fmt.Fprintf(w, "\n      命令: %s", s.Command)
		}
		fmt.Fprintln(w)
	}
	fmt.Fprintf(w, "置信度: %.0f%%\n", a.Confidence*100)
}

func printDiagJSON(cmd *cobra.Command, result *diag.DiagnosisResult) {
	w := cmd.OutOrStdout()
	fmt.Fprint(w, "{\n  \"issues\": [")
	for i, issue := range result.Issues {
		if i > 0 {
			fmt.Fprint(w, ",")
		}
		fmt.Fprintf(w, "\n    {\"severity\":\"%s\",\"name\":%q}", issue.Severity.EnglishString(), issue.CNName)
	}
	fmt.Fprintf(w, "\n  ],\n  \"totalAnalyzers\": %d,\n  \"totalIssues\": %d", len(result.Results), len(result.Issues))
	if result.AIAnalysis != nil {
		fmt.Fprintf(w, ",\n  \"aiSummary\": %q", result.AIAnalysis.Summary)
	}
	fmt.Fprint(w, "\n}\n")
}
