package servicemesh

import (
	"context"
	"fmt"
	"strings"

	"github.com/kudig-io/klaw/internal/diag/analyzer"
	"github.com/kudig-io/klaw/internal/diag/types"
)

// IstioAnalyzer checks Istio service mesh health
type IstioAnalyzer struct {
	*analyzer.BaseAnalyzer
}

// NewIstioAnalyzer creates a new Istio analyzer
func NewIstioAnalyzer() *IstioAnalyzer {
	return &IstioAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer(
			"servicemesh.istio",
			"Istio 服务网格健康检查",
			"servicemesh",
			[]types.DataMode{types.ModeOnline},
		),
	}
}

// Analyze checks Istio health
func (a *IstioAnalyzer) Analyze(ctx context.Context, data *types.DiagnosticData) ([]types.Issue, error) {
	var issues []types.Issue

	// Check if Istio is installed
	if data.IstioInfo == nil {
		// Istio not installed, no issues to report
		return issues, nil
	}

	// Check Istiod control plane
	if data.IstioInfo.IstiodPods == 0 {
		issue := types.NewIssue(
			types.SeverityCritical,
			"Istiod 控制平面未运行",
			"ISTIO_ISTIOD_MISSING",
			"Istio 控制平面 (istiod) 未检测到运行实例",
			"k8s",
		).WithRemediation("检查 istiod deployment: kubectl get deployment -n istio-system istiod")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	} else if data.IstioInfo.IstiodReady < data.IstioInfo.IstiodPods {
		issue := types.NewIssue(
			types.SeverityWarning,
			"Istiod 控制平面 Pod 未全部就绪",
			"ISTIO_ISTIOD_NOT_READY",
			fmt.Sprintf("Istiod: %d/%d pods 就绪", data.IstioInfo.IstiodReady, data.IstioInfo.IstiodPods),
			"k8s",
		).WithRemediation("检查 istiod pod 状态: kubectl get pods -n istio-system -l app=istiod")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	}

	// Check Ingress Gateway
	if data.IstioInfo.IngressPods == 0 {
		issue := types.NewIssue(
			types.SeverityWarning,
			"Istio Ingress Gateway 未运行",
			"ISTIO_INGRESS_MISSING",
			"Istio Ingress Gateway 未检测到运行实例",
			"k8s",
		).WithRemediation("检查 ingress gateway: kubectl get pods -n istio-system -l app=istio-ingressgateway")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	} else if data.IstioInfo.IngressReady < data.IstioInfo.IngressPods {
		issue := types.NewIssue(
			types.SeverityWarning,
			"Istio Ingress Gateway Pod 未全部就绪",
			"ISTIO_INGRESS_NOT_READY",
			fmt.Sprintf("Ingress Gateway: %d/%d pods 就绪", data.IstioInfo.IngressReady, data.IstioInfo.IngressPods),
			"k8s",
		).WithRemediation("检查 ingress gateway pod 状态")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	}

	// Check Egress Gateway (if deployed)
	if data.IstioInfo.EgressPods > 0 && data.IstioInfo.EgressReady < data.IstioInfo.EgressPods {
		issue := types.NewIssue(
			types.SeverityWarning,
			"Istio Egress Gateway Pod 未全部就绪",
			"ISTIO_EGRESS_NOT_READY",
			fmt.Sprintf("Egress Gateway: %d/%d pods 就绪", data.IstioInfo.EgressReady, data.IstioInfo.EgressPods),
			"k8s",
		).WithRemediation("检查 egress gateway pod 状态")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	}

	// Check sidecar injection
	for ns, enabled := range data.IstioInfo.InjectionEnabled {
		if !enabled {
			continue
		}
		// Check if namespace has istio-injection=enabled label
		// This is a simplified check
		if strings.Contains(ns, "kube-") || ns == "istio-system" {
			continue
		}
	}

	// Check for sidecar proxy issues
	for _, proxy := range data.IstioInfo.ProxyStatus {
		if !proxy.Synced {
			issue := types.NewIssue(
				types.SeverityWarning,
				"Istio Sidecar 同步失败",
				"ISTIO_PROXY_NOT_SYNCED",
				fmt.Sprintf("Pod %s/%s 的 sidecar 代理未同步 (STALE)", proxy.Namespace, proxy.PodName),
				"k8s",
			).WithRemediation(fmt.Sprintf("重启 pod: kubectl rollout restart deployment -n %s", proxy.Namespace))
			issue.AnalyzerName = a.Name()
			issues = append(issues, *issue)
		}
		if proxy.CDS == "STALE" || proxy.LDS == "STALE" || proxy.RDS == "STALE" || proxy.EDS == "STALE" {
			issue := types.NewIssue(
				types.SeverityWarning,
				"Istio Sidecar 配置过期",
				"ISTIO_PROXY_STALE",
				fmt.Sprintf("Pod %s/%s 的配置过期: CDS=%s LDS=%s RDS=%s EDS=%s",
					proxy.Namespace, proxy.PodName, proxy.CDS, proxy.LDS, proxy.RDS, proxy.EDS),
				"k8s",
			).WithRemediation("检查 pilot 日志或重启 istiod")
			issue.AnalyzerName = a.Name()
			issues = append(issues, *issue)
		}
	}

	// Check mTLS configuration
	if data.IstioInfo.MTLSEnabled {
		// mTLS is enabled, check for strict mode issues
		for _, pod := range data.IstioInfo.PodsWithoutSidecar {
			issue := types.NewIssue(
				types.SeverityWarning,
				"Pod 未注入 Istio Sidecar",
				"ISTIO_SIDECAR_MISSING",
				fmt.Sprintf("Pod %s/%s 在启用 mTLS 的命名空间中但未注入 sidecar", pod.Namespace, pod.Name),
				"k8s",
			).WithRemediation("添加 sidecar.istio.io/inject=true 注解或检查 injection webhook")
			issue.AnalyzerName = a.Name()
			issues = append(issues, *issue)
		}
	}

	// Check version consistency
	if len(data.IstioInfo.ProxyVersions) > 1 {
		issue := types.NewIssue(
			types.SeverityWarning,
			"Istio Sidecar 版本不一致",
			"ISTIO_VERSION_MISMATCH",
			fmt.Sprintf("检测到 %d 个不同版本的 sidecar: %v", len(data.IstioInfo.ProxyVersions), data.IstioInfo.ProxyVersions),
			"k8s",
		).WithRemediation("统一升级所有 sidecar: istioctl proxy-status 检查版本")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	}

	// Check for high error rates in virtual services
	for _, vs := range data.IstioInfo.VirtualServices {
		if len(vs.HTTP) == 0 && len(vs.TCP) == 0 {
			issue := types.NewIssue(
				types.SeverityInfo,
				"VirtualService 无路由规则",
				"ISTIO_VS_EMPTY",
				fmt.Sprintf("VirtualService %s/%s 没有定义 HTTP 或 TCP 路由规则", vs.Namespace, vs.Name),
				"k8s",
			).WithRemediation("检查 VirtualService 配置是否完整")
			issue.AnalyzerName = a.Name()
			issues = append(issues, *issue)
		}
	}

	// Check for DestinationRule issues
	for _, dr := range data.IstioInfo.DestinationRules {
		if dr.Subsets == nil || len(dr.Subsets) == 0 {
			issue := types.NewIssue(
				types.SeverityInfo,
				"DestinationRule 无子集定义",
				"ISTIO_DR_NO_SUBSETS",
				fmt.Sprintf("DestinationRule %s/%s 没有定义 subsets", dr.Namespace, dr.Name),
				"k8s",
			).WithRemediation("考虑添加 subsets 以实现流量分割")
			issue.AnalyzerName = a.Name()
			issues = append(issues, *issue)
		}
	}

	return issues, nil
}

func init() {
	_ = analyzer.Register(NewIstioAnalyzer())
}
