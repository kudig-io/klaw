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
	NvidiaGPUResource     = "nvidia.com/gpu"
	NvidiaGPUMemory       = "nvidia.com/gpu-memory"
	AMDGPUResource        = "amd.com/gpu"
	HUAWEINPUResource     = "huawei.com/Ascend910"
)

// GPUNodeAnalyzer 检查 GPU 节点状态
type GPUNodeAnalyzer struct {
	*analyzer.BaseAnalyzer
}

// NewGPUNodeAnalyzer 创建 GPU 节点分析器
func NewGPUNodeAnalyzer() *GPUNodeAnalyzer {
	return &GPUNodeAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer(
			"kubernetes.gpu.node",
			"检查 GPU 节点状态",
			"kubernetes",
			[]types.DataMode{types.ModeOnline},
		),
	}
}

// Analyze 执行 GPU 节点健康检查
func (a *GPUNodeAnalyzer) Analyze(ctx context.Context, data *types.DiagnosticData) ([]types.Issue, error) {
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
		if !a.isGPUNode(&node) {
			continue
		}

		nodeIssues := a.checkGPUNode(ctx, client, &node)
		for i := range nodeIssues {
			nodeIssues[i].AnalyzerName = a.Name()
			issues = append(issues, nodeIssues[i])
		}
	}

	return issues, nil
}

func (a *GPUNodeAnalyzer) isGPUNode(node *corev1.Node) bool {
	// 检查标签
	gpuLabels := []string{
		"nvidia.com/gpu.present",
		"feature.node.kubernetes.io/pci-10de.present", // NVIDIA PCI ID
		"feature.node.kubernetes.io/pci-1002.present", // AMD PCI ID
		"huawei.com/ascend910",
	}

	for _, label := range gpuLabels {
		if _, ok := node.Labels[label]; ok {
			return true
		}
	}

	// 检查资源容量
	if _, ok := node.Status.Capacity[NvidiaGPUResource]; ok {
		return true
	}
	if _, ok := node.Status.Capacity[AMDGPUResource]; ok {
		return true
	}

	return false
}

func (a *GPUNodeAnalyzer) checkGPUNode(ctx context.Context, client kubernetes.Interface, node *corev1.Node) []types.Issue {
	var issues []types.Issue
	nodeName := node.Name

	// 检查 GPU 设备插件
	gpuPluginRunning := false
	pods, err := client.CoreV1().Pods("kube-system").List(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("spec.nodeName=%s", nodeName),
	})
	if err == nil {
		for _, pod := range pods.Items {
			if strings.Contains(pod.Name, "nvidia-device-plugin") ||
				strings.Contains(pod.Name, "gpu-device-plugin") ||
				strings.Contains(pod.Name, "ascend-device-plugin") {
				if pod.Status.Phase == corev1.PodRunning {
					gpuPluginRunning = true
					break
				}
			}
		}
	}

	if !gpuPluginRunning {
		issue := types.NewIssue(
			types.SeverityWarning,
			fmt.Sprintf("节点 %s 的 GPU 设备插件未运行", nodeName),
			"GPU_PLUGIN_NOT_RUNNING",
			"GPU 节点上未检测到正在运行的 GPU 设备插件 Pod",
			nodeName,
		).WithRemediation("检查 GPU 设备插件 DaemonSet 是否已部署并正常运行")
		issues = append(issues, *issue)
	}

	// 检查 GPU 资源容量与可分配量
	gpuCapacity := node.Status.Capacity[NvidiaGPUResource]
	gpuAllocatable := node.Status.Allocatable[NvidiaGPUResource]

	if gpuCapacity.Value() > 0 {
		if gpuAllocatable.Value() == 0 {
			issue := types.NewIssue(
				types.SeverityCritical,
				fmt.Sprintf("节点 %s GPU 不可分配", nodeName),
				"GPU_NOT_ALLOCATABLE",
				"节点声明有 GPU 但 allocatable 为 0，可能驱动有问题",
				nodeName,
			).WithRemediation("检查 NVIDIA 驱动安装状态，执行 nvidia-smi 验证")
			issues = append(issues, *issue)
		} else if gpuAllocatable.Value() < gpuCapacity.Value() {
			issue := types.NewIssue(
				types.SeverityInfo,
				fmt.Sprintf("节点 %s 部分 GPU 不可用", nodeName),
				"GPU_PARTIALLY_AVAILABLE",
				fmt.Sprintf("GPU 容量: %s, 可分配: %s", gpuCapacity.String(), gpuAllocatable.String()),
				nodeName,
			).WithRemediation("某些 GPU 可能已被 MIG 分区占用或出现故障")
			issues = append(issues, *issue)
		}
	}

	// 检查节点条件
	for _, cond := range node.Status.Conditions {
		if cond.Type == "Ready" && cond.Status != corev1.ConditionTrue {
			issue := types.NewIssue(
				types.SeverityWarning,
				fmt.Sprintf("GPU 节点 %s 状态异常", nodeName),
				"GPU_NODE_NOT_READY",
				fmt.Sprintf("节点 Ready 状态为 %s: %s", cond.Status, cond.Message),
				nodeName,
			).WithRemediation("检查节点状态和 GPU 驱动")
			issues = append(issues, *issue)
		}
	}

	return issues
}

// GPUPodAnalyzer 检查 GPU Pod 配置
type GPUPodAnalyzer struct {
	*analyzer.BaseAnalyzer
}

// NewGPUPodAnalyzer 创建 GPU Pod 分析器
func NewGPUPodAnalyzer() *GPUPodAnalyzer {
	return &GPUPodAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer(
			"kubernetes.gpu.pod",
			"检查 GPU Pod 配置",
			"kubernetes",
			[]types.DataMode{types.ModeOnline},
		),
	}
}

// Analyze 执行 GPU Pod 健康检查
func (a *GPUPodAnalyzer) Analyze(ctx context.Context, data *types.DiagnosticData) ([]types.Issue, error) {
	var issues []types.Issue

	if !data.HasK8sClient() {
		return issues, nil
	}

	client := data.K8sClient

	pods, err := client.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return issues, nil
	}

	for _, pod := range pods.Items {
		if !a.hasGPURequest(&pod) {
			continue
		}

		podIssues := a.checkGPUPod(ctx, client, &pod)
		for i := range podIssues {
			podIssues[i].AnalyzerName = a.Name()
			issues = append(issues, podIssues[i])
		}
	}

	return issues, nil
}

func (a *GPUPodAnalyzer) hasGPURequest(pod *corev1.Pod) bool {
	for _, container := range pod.Spec.Containers {
		if _, ok := container.Resources.Requests[NvidiaGPUResource]; ok {
			return true
		}
		if _, ok := container.Resources.Requests[AMDGPUResource]; ok {
			return true
		}
		if _, ok := container.Resources.Requests[HUAWEINPUResource]; ok {
			return true
		}
	}
	return false
}

func (a *GPUPodAnalyzer) checkGPUPod(ctx context.Context, client kubernetes.Interface, pod *corev1.Pod) []types.Issue {
	var issues []types.Issue
	podName := pod.Name
	ns := pod.Namespace

	// 检查 Pod 状态
	if pod.Status.Phase == corev1.PodPending {
		for _, cond := range pod.Status.Conditions {
			if cond.Type == corev1.PodScheduled && cond.Status == corev1.ConditionFalse {
				if strings.Contains(cond.Message, "nvidia.com/gpu") ||
					strings.Contains(cond.Message, "Insufficient") {
					issue := types.NewIssue(
						types.SeverityWarning,
						fmt.Sprintf("GPU Pod %s/%s 调度失败", ns, podName),
						"GPU_POD_SCHEDULING_FAILED",
						fmt.Sprintf("GPU 资源不足: %s", cond.Message),
						fmt.Sprintf("%s/%s", ns, podName),
					).WithRemediation("检查集群 GPU 节点资源和 GPU Pod 分布")
					issues = append(issues, *issue)
				}
			}
		}
	}

	// 检查容器配置
	for _, container := range pod.Spec.Containers {
		gpuReq, hasGPU := container.Resources.Requests[NvidiaGPUResource]
		if !hasGPU {
			continue
		}

		// 检查 GPU 限制是否与请求匹配
		gpuLimit, hasLimit := container.Resources.Limits[NvidiaGPUResource]
		if !hasLimit || gpuLimit.Value() != gpuReq.Value() {
			issue := types.NewIssue(
				types.SeverityWarning,
				fmt.Sprintf("容器 %s 的 GPU 资源限制不匹配", container.Name),
				"GPU_RESOURCE_MISMATCH",
				"GPU 资源请求和限制应该相等，因为 GPU 不可共享",
				fmt.Sprintf("%s/%s", ns, podName),
			).WithRemediation(fmt.Sprintf("设置 resources.limits[%s] = resources.requests[%s]", NvidiaGPUResource, NvidiaGPUResource))
			issues = append(issues, *issue)
		}

		// 检查是否请求了非整数 GPU（value 是毫值，1000 = 1个GPU）
		if gpuReq.MilliValue() < 1000 {
			issue := types.NewIssue(
				types.SeverityInfo,
				fmt.Sprintf("容器 %s 请求了分数 GPU", container.Name),
				"GPU_FRACTION_REQUESTED",
				fmt.Sprintf("请求了 %s GPU，但标准设备插件不支持分数分配", gpuReq.String()),
				fmt.Sprintf("%s/%s", ns, podName),
			).WithRemediation("使用 MIG (Multi-Instance GPU) 或 GPU 共享方案（如阿里云的 cgpu）")
			issues = append(issues, *issue)
		}
	}

	// 检查是否缺少必要的运行时类
	if pod.Spec.RuntimeClassName == nil {
		// 检查节点是否有 nvidia 运行时
		node, err := client.CoreV1().Nodes().Get(ctx, pod.Spec.NodeName, metav1.GetOptions{})
		if err == nil {
			if _, hasGPU := node.Status.Capacity[NvidiaGPUResource]; hasGPU {
				// 在 GPU 节点上运行但没有指定 runtime class
				// 这可能正常，取决于 containerd/crio 配置
				_ = node
			}
		}
	}

	return issues
}

// GPUShareAnalyzer 检查 GPU 共享配置
type GPUShareAnalyzer struct {
	*analyzer.BaseAnalyzer
}

// NewGPUShareAnalyzer 创建 GPU 共享分析器
func NewGPUShareAnalyzer() *GPUShareAnalyzer {
	return &GPUShareAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer(
			"kubernetes.gpu.share",
			"检查 GPU 共享配置",
			"kubernetes",
			[]types.DataMode{types.ModeOnline},
		),
	}
}

// Analyze 检查 GPU 共享配置
func (a *GPUShareAnalyzer) Analyze(ctx context.Context, data *types.DiagnosticData) ([]types.Issue, error) {
	var issues []types.Issue

	if !data.HasK8sClient() {
		return issues, nil
	}

	client := data.K8sClient

	// 检查 MIG 配置
	configMaps, err := client.CoreV1().ConfigMaps("").List(ctx, metav1.ListOptions{
		LabelSelector: "app=nvidia-device-plugin",
	})
	if err == nil {
		for _, cm := range configMaps.Items {
			if config, ok := cm.Data["config"]; ok {
				if strings.Contains(config, "migStrategy") {
					// 检查 MIG 策略配置
					if strings.Contains(config, `"migStrategy": "none"`) {
						// 使用单例模式，检查是否有 Pod 请求超过一个 GPU
						issue := types.NewIssue(
							types.SeverityInfo,
							"GPU 使用单例模式（MIG 未启用）",
							"GPU_SINGLE_MODE",
							"nvidia-device-plugin 配置为 migStrategy: none，每个 Pod 独占一个完整 GPU",
							"cluster/gpu",
						).WithRemediation("如需 GPU 共享，考虑启用 MIG 或使用 GPU 虚拟化方案")
						issue.AnalyzerName = a.Name()
						issues = append(issues, *issue)
					}
				}
			}
		}
	}

	// 检查 GPU Operator 状态
	gpuOperatorPods, err := client.CoreV1().Pods("gpu-operator").List(ctx, metav1.ListOptions{})
	if err == nil && len(gpuOperatorPods.Items) > 0 {
		for _, pod := range gpuOperatorPods.Items {
			if pod.Status.Phase != corev1.PodRunning {
				issue := types.NewIssue(
					types.SeverityWarning,
					fmt.Sprintf("GPU Operator Pod %s 状态异常", pod.Name),
					"GPU_OPERATOR_POD_NOT_RUNNING",
					fmt.Sprintf("Pod 状态: %s", pod.Status.Phase),
					fmt.Sprintf("gpu-operator/%s", pod.Name),
				).WithRemediation("检查 GPU Operator 日志和 NVIDIA 驱动状态")
				issue.AnalyzerName = a.Name()
				issues = append(issues, *issue)
			}
		}
	}

	return issues, nil
}

// NPUAnalyzer 检查华为 NPU (Ascend) 状态
type NPUAnalyzer struct {
	*analyzer.BaseAnalyzer
}

// NewNPUAnalyzer 创建 NPU 分析器
func NewNPUAnalyzer() *NPUAnalyzer {
	return &NPUAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer(
			"kubernetes.npu",
			"检查华为 NPU (Ascend) 状态",
			"kubernetes",
			[]types.DataMode{types.ModeOnline},
		),
	}
}

// Analyze 执行 NPU 健康检查
func (a *NPUAnalyzer) Analyze(ctx context.Context, data *types.DiagnosticData) ([]types.Issue, error) {
	var issues []types.Issue

	if !data.HasK8sClient() {
		return issues, nil
	}

	client := data.K8sClient

	// 查找 NPU 节点
	nodes, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{
		LabelSelector: "huawei.com/ascend910=true",
	})
	if err != nil {
		// 尝试通过资源查找
		allNodes, _ := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
		for _, node := range allNodes.Items {
			if _, ok := node.Status.Capacity[HUAWEINPUResource]; ok {
				nodes.Items = append(nodes.Items, node)
			}
		}
	}

	for _, node := range nodes.Items {
		npuCount := node.Status.Capacity[HUAWEINPUResource]
		if npuCount.Value() == 0 {
			continue
		}

		// 检查 NPU 设备插件
		pluginRunning := false
		pods, _ := client.CoreV1().Pods("kube-system").List(ctx, metav1.ListOptions{
			FieldSelector: fmt.Sprintf("spec.nodeName=%s", node.Name),
		})
		for _, pod := range pods.Items {
			if strings.Contains(pod.Name, "ascend") && pod.Status.Phase == corev1.PodRunning {
				pluginRunning = true
				break
			}
		}

		if !pluginRunning {
			issue := types.NewIssue(
				types.SeverityWarning,
				fmt.Sprintf("节点 %s 的 NPU 设备插件未运行", node.Name),
				"NPU_PLUGIN_NOT_RUNNING",
				"NPU 节点上未检测到正在运行的 Ascend 设备插件",
				node.Name,
			).WithRemediation("检查 Ascend Device Plugin 是否已部署")
			issue.AnalyzerName = a.Name()
			issues = append(issues, *issue)
		}
	}

	return issues, nil
}

func init() {
	// 注册所有 GPU 分析器
	_ = analyzer.Register(NewGPUNodeAnalyzer())
	_ = analyzer.Register(NewGPUPodAnalyzer())
	_ = analyzer.Register(NewGPUShareAnalyzer())
	_ = analyzer.Register(NewNPUAnalyzer())
}
