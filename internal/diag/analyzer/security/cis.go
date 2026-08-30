// Package security provides security compliance analyzers
package security

import (
	"context"
	"fmt"
	"strings"

	"github.com/kudig-io/klaw/internal/diag/analyzer"
	"github.com/kudig-io/klaw/internal/diag/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CISAPIServerAnalyzer checks API Server security configuration per CIS benchmark
type CISAPIServerAnalyzer struct {
	*analyzer.BaseAnalyzer
}

// NewCISAPIServerAnalyzer creates a new CIS API Server analyzer
func NewCISAPIServerAnalyzer() *CISAPIServerAnalyzer {
	return &CISAPIServerAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer(
			"security.cis.apiserver",
			"检查 API Server CIS 安全配置",
			"security",
			[]types.DataMode{types.ModeOnline},
		),
	}
}

// Analyze performs CIS API Server security checks
func (a *CISAPIServerAnalyzer) Analyze(ctx context.Context, data *types.DiagnosticData) ([]types.Issue, error) {
	var issues []types.Issue

	if !data.HasK8sClient() {
		return issues, nil
	}

	client := data.K8sClient

	// Get API Server pods
	pods, err := client.CoreV1().Pods("kube-system").List(ctx, metav1.ListOptions{
		LabelSelector: "component=kube-apiserver",
	})
	if err != nil || len(pods.Items) == 0 {
		return issues, nil
	}

	for _, pod := range pods.Items {
		// Check for anonymous authentication
		for _, container := range pod.Spec.Containers {
			if container.Name == "kube-apiserver" {
				for _, arg := range container.Command {
					if strings.Contains(arg, "--anonymous-auth=true") {
						issue := types.NewIssue(
							types.SeverityCritical,
							"API Server 启用匿名认证",
							"CIS_APISERVER_ANONYMOUS_AUTH",
							"API Server 配置启用了匿名认证，违反 CIS Benchmark 1.2.1",
							fmt.Sprintf("kube-system/%s", pod.Name),
						).WithRemediation("设置 --anonymous-auth=false")
						issue.AnalyzerName = a.Name()
						issues = append(issues, *issue)
					}
					if strings.Contains(arg, "--insecure-port") && !strings.Contains(arg, "--insecure-port=0") {
						issue := types.NewIssue(
							types.SeverityCritical,
							"API Server 启用不安全端口",
							"CIS_APISERVER_INSECURE_PORT",
							"API Server 配置启用了不安全端口，违反 CIS Benchmark 1.2.19",
							fmt.Sprintf("kube-system/%s", pod.Name),
						).WithRemediation("设置 --insecure-port=0")
						issue.AnalyzerName = a.Name()
						issues = append(issues, *issue)
					}
					if strings.Contains(arg, "--insecure-bind-address") {
						issue := types.NewIssue(
							types.SeverityCritical,
							"API Server 绑定到不安全地址",
							"CIS_APISERVER_INSECURE_BIND",
							"API Server 配置绑定了不安全地址，违反 CIS Benchmark 1.2.20",
							fmt.Sprintf("kube-system/%s", pod.Name),
						).WithRemediation("移除 --insecure-bind-address 参数")
						issue.AnalyzerName = a.Name()
						issues = append(issues, *issue)
					}
				}
			}
		}
	}

	return issues, nil
}

// CISEtcdAnalyzer checks etcd security configuration per CIS benchmark
type CISEtcdAnalyzer struct {
	*analyzer.BaseAnalyzer
}

// NewCISEtcdAnalyzer creates a new CIS etcd analyzer
func NewCISEtcdAnalyzer() *CISEtcdAnalyzer {
	return &CISEtcdAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer(
			"security.cis.etcd",
			"检查 etcd CIS 安全配置",
			"security",
			[]types.DataMode{types.ModeOnline},
		),
	}
}

// Analyze performs CIS etcd security checks
func (a *CISEtcdAnalyzer) Analyze(ctx context.Context, data *types.DiagnosticData) ([]types.Issue, error) {
	var issues []types.Issue

	if !data.HasK8sClient() {
		return issues, nil
	}

	client := data.K8sClient

	// Get etcd pods
	pods, err := client.CoreV1().Pods("kube-system").List(ctx, metav1.ListOptions{
		LabelSelector: "component=etcd",
	})
	if err != nil || len(pods.Items) == 0 {
		return issues, nil
	}

	for _, pod := range pods.Items {
		for _, container := range pod.Spec.Containers {
			if container.Name == "etcd" {
				hasCertAuth := false
				hasAutoTLS := false
				for _, arg := range container.Command {
					if strings.Contains(arg, "--client-cert-auth=true") {
						hasCertAuth = true
					}
					if strings.Contains(arg, "--auto-tls=true") {
						hasAutoTLS = true
					}
				}
				if !hasCertAuth {
					issue := types.NewIssue(
						types.SeverityCritical,
						"etcd 未启用客户端证书认证",
						"CIS_ETCD_NO_CERT_AUTH",
						"etcd 未启用客户端证书认证，违反 CIS Benchmark 2.1",
						fmt.Sprintf("kube-system/%s", pod.Name),
					).WithRemediation("设置 --client-cert-auth=true")
					issue.AnalyzerName = a.Name()
					issues = append(issues, *issue)
				}
				if hasAutoTLS {
					issue := types.NewIssue(
						types.SeverityWarning,
						"etcd 使用自动 TLS",
						"CIS_ETCD_AUTO_TLS",
						"etcd 使用自动 TLS 而非手动配置的证书，违反 CIS Benchmark 2.2",
						fmt.Sprintf("kube-system/%s", pod.Name),
					).WithRemediation("禁用 --auto-tls 并使用手动配置的 TLS 证书")
					issue.AnalyzerName = a.Name()
					issues = append(issues, *issue)
				}
			}
		}
	}

	return issues, nil
}

// CISKubeletAnalyzer checks Kubelet security configuration per CIS benchmark
type CISKubeletAnalyzer struct {
	*analyzer.BaseAnalyzer
}

// NewCISKubeletAnalyzer creates a new CIS Kubelet analyzer
func NewCISKubeletAnalyzer() *CISKubeletAnalyzer {
	return &CISKubeletAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer(
			"security.cis.kubelet",
			"检查 Kubelet CIS 安全配置",
			"security",
			[]types.DataMode{types.ModeOnline},
		),
	}
}

// Analyze performs CIS Kubelet security checks
func (a *CISKubeletAnalyzer) Analyze(ctx context.Context, data *types.DiagnosticData) ([]types.Issue, error) {
	var issues []types.Issue

	if !data.HasK8sClient() {
		return issues, nil
	}

	client := data.K8sClient

	nodes, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return issues, nil
	}

	for _, node := range nodes.Items {
		// Check Kubelet configuration via node annotations
		if kubeletConfig, ok := node.Annotations["node.alpha.kubernetes.io/ttl"]; ok {
			_ = kubeletConfig
			// Note: Detailed Kubelet config checking requires access to Kubelet configz endpoint
			// which is not available via API server. We'll check what we can from node annotations.
		}
	}

	return issues, nil
}

// CISPodSecurityAnalyzer checks Pod Security compliance per CIS benchmark
type CISPodSecurityAnalyzer struct {
	*analyzer.BaseAnalyzer
}

// NewCISPodSecurityAnalyzer creates a new CIS Pod Security analyzer
func NewCISPodSecurityAnalyzer() *CISPodSecurityAnalyzer {
	return &CISPodSecurityAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer(
			"security.cis.pod",
			"检查 Pod CIS 安全配置",
			"security",
			[]types.DataMode{types.ModeOnline},
		),
	}
}

// Analyze performs CIS Pod Security checks
func (a *CISPodSecurityAnalyzer) Analyze(ctx context.Context, data *types.DiagnosticData) ([]types.Issue, error) {
	var issues []types.Issue

	if !data.HasK8sClient() {
		return issues, nil
	}

	client := data.K8sClient

	// Check for privileged pods
	pods, err := client.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return issues, nil
	}

	privilegedCount := 0
	hostPIDCount := 0
	hostNetworkCount := 0
	runAsRootCount := 0
	var privilegedPods []string

	for _, pod := range pods.Items {
		for _, container := range pod.Spec.Containers {
			if container.SecurityContext != nil {
				if container.SecurityContext.Privileged != nil && *container.SecurityContext.Privileged {
					privilegedCount++
					privilegedPods = append(privilegedPods, fmt.Sprintf("%s/%s", pod.Namespace, pod.Name))
				}
				if container.SecurityContext.RunAsUser != nil && *container.SecurityContext.RunAsUser == 0 {
					runAsRootCount++
				}
			}
		}
		if pod.Spec.HostPID {
			hostPIDCount++
		}
		if pod.Spec.HostNetwork {
			hostNetworkCount++
		}
	}

	if privilegedCount > 0 {
		issue := types.NewIssue(
			types.SeverityWarning,
			fmt.Sprintf("发现 %d 个特权容器", privilegedCount),
			"CIS_PRIVILEGED_CONTAINERS",
			fmt.Sprintf("发现特权容器运行: %v，违反 CIS Benchmark 5.2.1", privilegedPods),
			"cluster/pods",
		).WithRemediation("避免使用 privileged: true，使用更细粒度的 capabilities")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	}

	if hostPIDCount > 0 {
		issue := types.NewIssue(
			types.SeverityWarning,
			fmt.Sprintf("发现 %d 个使用 HostPID 的 Pod", hostPIDCount),
			"CIS_HOST_PID",
			"Pod 使用 HostPID 共享主机进程命名空间，违反 CIS Benchmark 5.2.3",
			"cluster/pods",
		).WithRemediation("避免使用 hostPID: true")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	}

	if hostNetworkCount > 0 {
		issue := types.NewIssue(
			types.SeverityInfo,
			fmt.Sprintf("发现 %d 个使用 HostNetwork 的 Pod", hostNetworkCount),
			"CIS_HOST_NETWORK",
			"Pod 使用 HostNetwork 共享主机网络命名空间",
			"cluster/pods",
		).WithRemediation("仅在必要时使用 hostNetwork: true")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	}

	if runAsRootCount > 0 {
		issue := types.NewIssue(
			types.SeverityWarning,
			fmt.Sprintf("发现 %d 个以 root 运行的容器", runAsRootCount),
			"CIS_RUN_AS_ROOT",
			"容器以 root 用户运行，违反 CIS Benchmark 5.2.6",
			"cluster/pods",
		).WithRemediation("设置 securityContext.runAsNonRoot: true")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	}

	return issues, nil
}

// CISNetworkPolicyAnalyzer checks Network Policy compliance per CIS benchmark
type CISNetworkPolicyAnalyzer struct {
	*analyzer.BaseAnalyzer
}

// NewCISNetworkPolicyAnalyzer creates a new CIS Network Policy analyzer
func NewCISNetworkPolicyAnalyzer() *CISNetworkPolicyAnalyzer {
	return &CISNetworkPolicyAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer(
			"security.cis.network",
			"检查 Network Policy CIS 合规",
			"security",
			[]types.DataMode{types.ModeOnline},
		),
	}
}

// Analyze performs CIS Network Policy checks
func (a *CISNetworkPolicyAnalyzer) Analyze(ctx context.Context, data *types.DiagnosticData) ([]types.Issue, error) {
	var issues []types.Issue

	if !data.HasK8sClient() {
		return issues, nil
	}

	client := data.K8sClient

	// Get all namespaces
	namespaces, err := client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return issues, nil
	}

	// Get all network policies
	networkPolicies, err := client.NetworkingV1().NetworkPolicies("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return issues, nil
	}

	// Create a map of namespace -> has network policy
	nsWithPolicy := make(map[string]bool)
	for _, np := range networkPolicies.Items {
		nsWithPolicy[np.Namespace] = true
	}

	// Check for namespaces without network policies (excluding system namespaces)
	systemNamespaces := map[string]bool{
		"kube-system":     true,
		"kube-public":     true,
		"kube-node-lease": true,
	}

	for _, ns := range namespaces.Items {
		if systemNamespaces[ns.Name] {
			continue
		}
		if !nsWithPolicy[ns.Name] {
			// Count pods in this namespace
			pods, _ := client.CoreV1().Pods(ns.Name).List(ctx, metav1.ListOptions{})
			if len(pods.Items) > 0 {
				issue := types.NewIssue(
					types.SeverityInfo,
					fmt.Sprintf("命名空间 %s 缺少 Network Policy", ns.Name),
					"CIS_NO_NETWORK_POLICY",
					fmt.Sprintf("命名空间 %s 有 %d 个 Pod 但没有定义 Network Policy，违反 CIS Benchmark 5.3.2", ns.Name, len(pods.Items)),
					ns.Name,
				).WithRemediation("为命名空间创建默认拒绝的 Network Policy")
				issue.AnalyzerName = a.Name()
				issues = append(issues, *issue)
			}
		}
	}

	return issues, nil
}

// CISSecretAnalyzer checks Secret management compliance per CIS benchmark
type CISSecretAnalyzer struct {
	*analyzer.BaseAnalyzer
}

// NewCISSecretAnalyzer creates a new CIS Secret analyzer
func NewCISSecretAnalyzer() *CISSecretAnalyzer {
	return &CISSecretAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer(
			"security.cis.secret",
			"检查 Secret CIS 管理合规",
			"security",
			[]types.DataMode{types.ModeOnline},
		),
	}
}

// Analyze performs CIS Secret management checks
func (a *CISSecretAnalyzer) Analyze(ctx context.Context, data *types.DiagnosticData) ([]types.Issue, error) {
	var issues []types.Issue

	if !data.HasK8sClient() {
		return issues, nil
	}

	client := data.K8sClient

	// Check for secrets in default namespace
	secrets, err := client.CoreV1().Secrets("default").List(ctx, metav1.ListOptions{})
	if err != nil {
		return issues, nil
	}

	if len(secrets.Items) > 0 {
		issue := types.NewIssue(
			types.SeverityWarning,
			fmt.Sprintf("default 命名空间有 %d 个 Secret", len(secrets.Items)),
			"CIS_SECRET_IN_DEFAULT",
			"default 命名空间不应存放 Secret，违反 CIS Benchmark 5.4.1",
			"default",
		).WithRemediation("将 Secret 迁移到专用命名空间")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	}

	return issues, nil
}

func init() {
	analyzer.MustRegister(NewCISAPIServerAnalyzer())
	analyzer.MustRegister(NewCISEtcdAnalyzer())
	analyzer.MustRegister(NewCISKubeletAnalyzer())
	analyzer.MustRegister(NewCISPodSecurityAnalyzer())
	analyzer.MustRegister(NewCISNetworkPolicyAnalyzer())
	analyzer.MustRegister(NewCISSecretAnalyzer())
}
