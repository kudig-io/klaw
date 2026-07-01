package servicemesh

import (
	"context"
	"fmt"

	"github.com/kudig-io/klaw/internal/diag/analyzer"
	"github.com/kudig-io/klaw/internal/diag/types"
)

// LinkerdAnalyzer checks Linkerd service mesh health
type LinkerdAnalyzer struct {
	*analyzer.BaseAnalyzer
}

// NewLinkerdAnalyzer creates a new Linkerd analyzer
func NewLinkerdAnalyzer() *LinkerdAnalyzer {
	return &LinkerdAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer(
			"servicemesh.linkerd",
			"Linkerd 服务网格健康检查",
			"servicemesh",
			[]types.DataMode{types.ModeOnline},
		),
	}
}

// Analyze checks Linkerd health
func (a *LinkerdAnalyzer) Analyze(ctx context.Context, data *types.DiagnosticData) ([]types.Issue, error) {
	var issues []types.Issue

	// Check if Linkerd is installed
	if data.LinkerdInfo == nil {
		// Linkerd not installed, no issues to report
		return issues, nil
	}

	// Check control plane pods
	if data.LinkerdInfo.ControlPlanePods == 0 {
		issue := types.NewIssue(
			types.SeverityCritical,
			"Linkerd 控制平面未运行",
			"LINKERD_CONTROL_PLANE_MISSING",
			"Linkerd 控制平面未检测到运行实例",
			"k8s",
		).WithRemediation("检查 linkerd control plane: kubectl get pods -n linkerd")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	} else if data.LinkerdInfo.ControlPlaneReady < data.LinkerdInfo.ControlPlanePods {
		issue := types.NewIssue(
			types.SeverityWarning,
			"Linkerd 控制平面 Pod 未全部就绪",
			"LINKERD_CONTROL_PLANE_NOT_READY",
			fmt.Sprintf("控制平面: %d/%d pods 就绪", data.LinkerdInfo.ControlPlaneReady, data.LinkerdInfo.ControlPlanePods),
			"k8s",
		).WithRemediation("检查 control plane pod 状态: kubectl get pods -n linkerd")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	}

	// Check proxy status
	for _, proxy := range data.LinkerdInfo.ProxyStatus {
		if proxy.Status != "healthy" {
			issue := types.NewIssue(
				types.SeverityWarning,
				"Linkerd Proxy 不健康",
				"LINKERD_PROXY_UNHEALTHY",
				fmt.Sprintf("Pod %s/%s 的 Linkerd proxy 状态: %s", proxy.Namespace, proxy.PodName, proxy.Status),
				"k8s",
			).WithRemediation(fmt.Sprintf("检查 proxy 日志: kubectl logs -n %s %s -c linkerd-proxy", proxy.Namespace, proxy.PodName))
			issue.AnalyzerName = a.Name()
			issues = append(issues, *issue)
		}
	}

	// Check for meshed pods without proxy
	for _, pod := range data.LinkerdInfo.PodsWithoutProxy {
		issue := types.NewIssue(
			types.SeverityInfo,
			"Pod 未注入 Linkerd Proxy",
			"LINKERD_PROXY_MISSING",
			fmt.Sprintf("Pod %s/%s 未注入 Linkerd proxy", pod.Namespace, pod.Name),
			"k8s",
		).WithRemediation("添加 linkerd.io/inject: enabled 注解")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	}

	// Check version consistency
	if len(data.LinkerdInfo.ProxyVersions) > 1 {
		issue := types.NewIssue(
			types.SeverityWarning,
			"Linkerd Proxy 版本不一致",
			"LINKERD_VERSION_MISMATCH",
			fmt.Sprintf("检测到 %d 个不同版本的 proxy: %v", len(data.LinkerdInfo.ProxyVersions), data.LinkerdInfo.ProxyVersions),
			"k8s",
		).WithRemediation("使用 linkerd upgrade 统一升级")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	}

	// Check for high latency services
	for svc, latency := range data.LinkerdInfo.ServiceLatencies {
		if latency > 1000 {
			issue := types.NewIssue(
				types.SeverityWarning,
				"Linkerd 服务延迟过高",
				"LINKERD_HIGH_LATENCY",
				fmt.Sprintf("服务 %s P99 延迟: %.2fms", svc, latency),
				"metrics",
			).WithRemediation("检查服务性能和资源使用情况")
			issue.AnalyzerName = a.Name()
			issues = append(issues, *issue)
		}
	}

	// Check for high error rate services
	for svc, errorRate := range data.LinkerdInfo.ServiceErrorRates {
		if errorRate > 0.01 {
			issue := types.NewIssue(
				types.SeverityCritical,
				"Linkerd 服务错误率过高",
				"LINKERD_HIGH_ERROR_RATE",
				fmt.Sprintf("服务 %s 错误率: %.2f%%", svc, errorRate*100),
				"metrics",
			).WithRemediation("检查服务日志和健康状况")
			issue.AnalyzerName = a.Name()
			issues = append(issues, *issue)
		}
	}

	return issues, nil
}

func init() {
	_ = analyzer.Register(NewLinkerdAnalyzer())
}
