// Package kubernetes provides Kubernetes component analyzers
package kubernetes

import (
	"context"
	"fmt"
	"strings"

	"github.com/kudig-io/klaw/internal/diag/analyzer"
	"github.com/kudig-io/klaw/internal/diag/types"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	CoreDNSNamespace     = "kube-system"
	CoreDNSLabelSelector = "k8s-app=kube-dns"
)

// CoreDNSAnalyzer 检查 CoreDNS 组件健康状态
type CoreDNSAnalyzer struct {
	*analyzer.BaseAnalyzer
}

// NewCoreDNSAnalyzer 创建 CoreDNS 分析器
func NewCoreDNSAnalyzer() *CoreDNSAnalyzer {
	return &CoreDNSAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer(
			"kubernetes.coredns",
			"检查 CoreDNS 健康状态",
			"kubernetes",
			[]types.DataMode{types.ModeOnline},
		),
	}
}

// Analyze 执行 CoreDNS 健康检查
func (a *CoreDNSAnalyzer) Analyze(ctx context.Context, data *types.DiagnosticData) ([]types.Issue, error) {
	var issues []types.Issue

	if !data.HasK8sClient() {
		return issues, nil
	}

	client := data.K8sClient

	// 检查 CoreDNS Pod 状态
	pods, err := client.CoreV1().Pods(CoreDNSNamespace).List(ctx, metav1.ListOptions{
		LabelSelector: CoreDNSLabelSelector,
	})
	if err != nil {
		return issues, nil // 静默跳过
	}

	if len(pods.Items) == 0 {
		issue := types.NewIssue(
			types.SeverityCritical,
			"CoreDNS 未部署",
			"COREDNS_NOT_FOUND",
			"未在 kube-system 命名空间找到 CoreDNS Pod，集群 DNS 服务不可用",
			"kube-system/coredns",
		).WithRemediation("检查 kube-system 命名空间的 coredns Deployment: kubectl get deployment -n kube-system")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
		return issues, nil
	}

	// 检查每个 Pod 的状态
	for _, pod := range pods.Items {
		podIssues := a.checkPodStatus(&pod)
		for i := range podIssues {
			podIssues[i].AnalyzerName = a.Name()
			issues = append(issues, podIssues[i])
		}
	}

	// 检查 CoreDNS Service
	svcIssues := a.checkService(ctx, client)
	for i := range svcIssues {
		svcIssues[i].AnalyzerName = a.Name()
		issues = append(issues, svcIssues[i])
	}

	// 检查 ConfigMap 配置
	configIssues := a.checkConfigMap(ctx, client)
	for i := range configIssues {
		configIssues[i].AnalyzerName = a.Name()
		issues = append(issues, configIssues[i])
	}

	return issues, nil
}

func (a *CoreDNSAnalyzer) checkPodStatus(pod *corev1.Pod) []types.Issue {
	var issues []types.Issue
	podName := pod.Name

	// 先检查容器状态（即使 Pod 不是 Running，也可能有容器问题）
	for _, container := range pod.Status.ContainerStatuses {
		// 检查是否在 CrashLoopBackOff（最高优先级）
		if container.State.Waiting != nil {
			waiting := container.State.Waiting
			if waiting.Reason == "CrashLoopBackOff" {
				issue := types.NewIssue(
					types.SeverityCritical,
					fmt.Sprintf("CoreDNS Pod %s 处于 CrashLoopBackOff", podName),
					"COREDNS_CRASH_LOOP",
					fmt.Sprintf("容器 %s 反复崩溃，原因: %s", container.Name, waiting.Message),
					fmt.Sprintf("kube-system/%s", podName),
				).WithRemediation(fmt.Sprintf("检查 CoreDNS 日志: kubectl logs %s -n kube-system", podName))
				issues = append(issues, *issue)
				return issues
			} else if waiting.Reason == "ImagePullBackOff" {
				issue := types.NewIssue(
					types.SeverityCritical,
					fmt.Sprintf("CoreDNS Pod %s 镜像拉取失败", podName),
					"COREDNS_IMAGE_PULL_FAILED",
					fmt.Sprintf("容器 %s 无法拉取镜像: %s", container.Name, waiting.Message),
					fmt.Sprintf("kube-system/%s", podName),
				).WithRemediation("检查镜像仓库访问和网络连接")
				issues = append(issues, *issue)
				return issues
			}
		}
	}

	// 检查 Pod 阶段
	if pod.Status.Phase != corev1.PodRunning {
		issue := types.NewIssue(
			types.SeverityWarning,
			fmt.Sprintf("CoreDNS Pod %s 状态异常", podName),
			"COREDNS_POD_NOT_RUNNING",
			fmt.Sprintf("Pod 当前状态为 %s，期望状态为 Running", pod.Status.Phase),
			fmt.Sprintf("kube-system/%s", podName),
		).WithRemediation(fmt.Sprintf("检查 Pod 事件和日志: kubectl describe pod %s -n kube-system", podName))
		issues = append(issues, *issue)
		return issues
	}

	// 检查容器状态（Running Pod 的容器检查）
	for _, container := range pod.Status.ContainerStatuses {
		// 检查重启次数
		if container.RestartCount > 5 {
			issue := types.NewIssue(
				types.SeverityWarning,
				fmt.Sprintf("CoreDNS Pod %s 重启次数过多", podName),
				"COREDNS_HIGH_RESTARTS",
				fmt.Sprintf("容器 %s 已重启 %d 次", container.Name, container.RestartCount),
				fmt.Sprintf("kube-system/%s", podName),
			).WithRemediation(fmt.Sprintf("检查容器日志: kubectl logs %s -n kube-system --previous", podName))
			issues = append(issues, *issue)
		}
	}

	return issues
}

func (a *CoreDNSAnalyzer) checkService(ctx context.Context, client kubernetes.Interface) []types.Issue {
	var issues []types.Issue

	svc, err := client.CoreV1().Services(CoreDNSNamespace).Get(ctx, "kube-dns", metav1.GetOptions{})
	if err != nil {
		// 可能是自定义 DNS 解决方案
		return nil
	}

	// 检查 Service 是否有 Endpoints
	eps, err := client.CoreV1().Endpoints(CoreDNSNamespace).Get(ctx, "kube-dns", metav1.GetOptions{})
	if err != nil {
		issue := types.NewIssue(
			types.SeverityCritical,
			"kube-dns Service 无 Endpoints",
			"COREDNS_NO_ENDPOINTS",
			"kube-dns Service 没有关联的 Endpoints，DNS 请求无法路由",
			"kube-system/kube-dns",
		).WithRemediation("检查 CoreDNS Pod 标签是否与 Service selector 匹配: kubectl get pods -n kube-system -l k8s-app=kube-dns")
		issues = append(issues, *issue)
		return issues
	}

	hasReadyAddresses := false
	for _, subset := range eps.Subsets {
		if len(subset.Addresses) > 0 {
			hasReadyAddresses = true
			break
		}
	}

	if !hasReadyAddresses {
		issue := types.NewIssue(
			types.SeverityCritical,
			"kube-dns Endpoints 无可用地址",
			"COREDNS_ENDPOINTS_EMPTY",
			"kube-dns Service 的 Endpoints 中没有可用的 Pod 地址",
			"kube-system/kube-dns",
		).WithRemediation("检查 CoreDNS Pod 是否正常运行: kubectl get pods -n kube-system -l k8s-app=kube-dns")
		issues = append(issues, *issue)
	}

	// 检查 Service IP 是否为空
	if svc.Spec.ClusterIP == "None" {
		// Headless Service，需要特殊处理
		_ = svc
	}

	return issues
}

func (a *CoreDNSAnalyzer) checkConfigMap(ctx context.Context, client kubernetes.Interface) []types.Issue {
	var issues []types.Issue

	cm, err := client.CoreV1().ConfigMaps(CoreDNSNamespace).Get(ctx, "coredns", metav1.GetOptions{})
	if err != nil {
		return nil // 可能使用其他配置方式
	}

	corefile, ok := cm.Data["Corefile"]
	if !ok {
		return nil
	}

	// 检查是否有上游 DNS 配置
	if !strings.Contains(corefile, "forward") && !strings.Contains(corefile, "proxy") {
		issue := types.NewIssue(
			types.SeverityWarning,
			"CoreDNS 缺少上游 DNS 配置",
			"COREDNS_NO_UPSTREAM",
			"Corefile 中没有 forward 或 proxy 插件配置",
			"kube-system/coredns",
		).WithRemediation("确保 CoreDNS ConfigMap 包含上游 DNS 服务器地址")
		issues = append(issues, *issue)
	}

	return issues
}

// DNSPodConfigAnalyzer 检查 Pod DNS 配置
type DNSPodConfigAnalyzer struct {
	*analyzer.BaseAnalyzer
}

// NewDNSPodConfigAnalyzer 创建 Pod DNS 配置分析器
func NewDNSPodConfigAnalyzer() *DNSPodConfigAnalyzer {
	return &DNSPodConfigAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer(
			"kubernetes.dns.pod_config",
			"检查 Pod DNS 配置",
			"kubernetes",
			[]types.DataMode{types.ModeOnline},
		),
	}
}

// Analyze 检查 Pod 的 DNS 配置
func (a *DNSPodConfigAnalyzer) Analyze(ctx context.Context, data *types.DiagnosticData) ([]types.Issue, error) {
	var issues []types.Issue

	if !data.HasK8sClient() {
		return issues, nil
	}

	client := data.K8sClient

	// 获取所有 Pod
	pods, err := client.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return issues, nil
	}

	// 统计使用不同 DNS 策略的 Pod
	dnsPolicyCount := make(map[corev1.DNSPolicy]int)
	for _, pod := range pods.Items {
		dnsPolicyCount[pod.Spec.DNSPolicy]++

		// 检查自定义 DNS 配置
		if pod.Spec.DNSConfig != nil {
			// 检查 ndots 配置
			if pod.Spec.DNSConfig.Options != nil {
				for _, opt := range pod.Spec.DNSConfig.Options {
					if opt.Name == "ndots" && opt.Value != nil {
						if *opt.Value == "1" {
							// ndots:1 可能导致意外的 DNS 查询
							issue := types.NewIssue(
								types.SeverityInfo,
								fmt.Sprintf("Pod %s/%s 使用 ndots:1", pod.Namespace, pod.Name),
								"POD_DNS_NDOTS_LOW",
								"ndots:1 可能导致大量外部 DNS 查询",
								fmt.Sprintf("%s/%s", pod.Namespace, pod.Name),
							).WithRemediation("考虑使用 ndots:5 以获得更好的集群内服务解析性能")
							issue.AnalyzerName = a.Name()
							issues = append(issues, *issue)
						}
					}
				}
			}
		}
	}

	// 检查是否有过多的 Pod 使用 Default DNS 策略
	// Default 策略使用节点的 /etc/resolv.conf，可能导致 DNS 解析不一致
	if pods.Items != nil && len(pods.Items) > 0 {
		defaultRatio := float64(dnsPolicyCount[corev1.DNSDefault]) / float64(len(pods.Items))
		if defaultRatio > 0.5 {
			issue := types.NewIssue(
				types.SeverityWarning,
				"大量 Pod 使用 Default DNS 策略",
				"POD_DNS_DEFAULT_HIGH",
				fmt.Sprintf("%.1f%% 的 Pod 使用 DNSDefault 策略，可能导致 DNS 解析不一致", defaultRatio*100),
				"cluster/dns",
			).WithRemediation("建议使用 ClusterFirst 策略以获得一致的集群 DNS 解析")
			issue.AnalyzerName = a.Name()
			issues = append(issues, *issue)
		}
	}

	return issues, nil
}

func init() {
	// 注册所有 DNS 分析器
	analyzer.MustRegister(NewCoreDNSAnalyzer())
	analyzer.MustRegister(NewDNSPodConfigAnalyzer())
}
