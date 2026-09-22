// Package kubernetes provides Kubernetes component analyzers
package kubernetes

import (
	"context"
	"fmt"
	"strings"

	"github.com/kudig-io/klaw/internal/diag/analyzer"
	"github.com/kudig-io/klaw/internal/diag/types"
)

// CNIAnalyzer checks for CNI plugin errors
type CNIAnalyzer struct {
	*analyzer.BaseAnalyzer
}

// NewCNIAnalyzer creates a new CNI analyzer
func NewCNIAnalyzer() *CNIAnalyzer {
	return &CNIAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer(
			"kubernetes.cni",
			"检查CNI网络插件状态",
			"kubernetes",
			[]types.DataMode{types.ModeOffline, types.ModeOnline},
		),
	}
}

// Analyze performs CNI analysis
func (a *CNIAnalyzer) Analyze(ctx context.Context, data *types.DiagnosticData) ([]types.Issue, error) {
	var issues []types.Issue

	kubeletLog, ok := data.GetRawFile("logs/kubelet.log")
	if !ok {
		kubeletLog, ok = data.GetRawFile("daemon_status/kubelet_status")
		if !ok {
			return issues, nil
		}
	}

	content := string(kubeletLog)

	// Check for CNI errors
	if strings.Contains(content, "Failed to create pod sandbox") && strings.Contains(content, "CNI") {
		cniErrorCount := strings.Count(content, "CNI") + strings.Count(content, "cni")
		issue := types.NewIssue(
			types.SeverityCritical,
			"CNI网络插件错误",
			"CNI_PLUGIN_ERROR",
			fmt.Sprintf("CNI网络插件失败 %d 次", cniErrorCount),
			"logs/kubelet.log",
		).WithRemediation("检查CNI插件状态: ls /etc/cni/net.d/; 重启网络插件Pod")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	}

	return issues, nil
}

// APIServerAnalyzer checks for API server connection issues
type APIServerAnalyzer struct {
	*analyzer.BaseAnalyzer
}

// NewAPIServerAnalyzer creates a new API server analyzer
func NewAPIServerAnalyzer() *APIServerAnalyzer {
	return &APIServerAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer(
			"kubernetes.apiserver",
			"检查API Server连接状态",
			"kubernetes",
			[]types.DataMode{types.ModeOffline, types.ModeOnline},
		),
	}
}

// Analyze performs API server connection analysis
func (a *APIServerAnalyzer) Analyze(ctx context.Context, data *types.DiagnosticData) ([]types.Issue, error) {
	var issues []types.Issue

	kubeletLog, ok := data.GetRawFile("logs/kubelet.log")
	if !ok {
		kubeletLog, ok = data.GetRawFile("daemon_status/kubelet_status")
		if !ok {
			return issues, nil
		}
	}

	content := string(kubeletLog)

	// Check for connection failures
	connFailCount := strings.Count(content, "Unable to connect to the server")
	connFailCount += strings.Count(content, "connection refused")

	if connFailCount > 10 {
		issue := types.NewIssue(
			types.SeverityCritical,
			"API Server连接失败",
			"APISERVER_CONNECTION_FAILED",
			fmt.Sprintf("无法连接到API Server，出现 %d 次", connFailCount),
			"logs/kubelet.log",
		).WithRemediation("检查网络连接: curl -k https://<api-server>:6443/healthz; 检查防火墙规则")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	}

	// Check for authentication failures
	if strings.Contains(content, "Unauthorized") {
		authFailCount := strings.Count(content, "Unauthorized")
		issue := types.NewIssue(
			types.SeverityCritical,
			"Kubelet认证失败",
			"KUBELET_AUTH_FAILED",
			fmt.Sprintf("Kubelet认证失败 %d 次", authFailCount),
			"logs/kubelet.log",
		).WithRemediation("检查kubeconfig和证书; 重新生成bootstrap token")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	}

	return issues, nil
}
