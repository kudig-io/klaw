// Package kubernetes provides Kubernetes component analyzers
package kubernetes

import (
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kudig-io/klaw/internal/diag/analyzer"
	"github.com/kudig-io/klaw/internal/diag/types"
)

// NodeStatusAnalyzer checks for node status issues
type NodeStatusAnalyzer struct {
	*analyzer.BaseAnalyzer
}

// NewNodeStatusAnalyzer creates a new node status analyzer
func NewNodeStatusAnalyzer() *NodeStatusAnalyzer {
	return &NodeStatusAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer(
			"kubernetes.node_status",
			"检查节点状态",
			"kubernetes",
			[]types.DataMode{types.ModeOffline, types.ModeOnline},
		),
	}
}

// Analyze performs node status analysis
func (a *NodeStatusAnalyzer) Analyze(ctx context.Context, data *types.DiagnosticData) ([]types.Issue, error) {
	var issues []types.Issue

	// Online mode: check node conditions via K8s API
	if data.HasK8sClient() {
		onlineIssues, err := a.analyzeOnline(ctx, data)
		if err == nil {
			issues = append(issues, onlineIssues...)
		}
	}

	// Also check raw files (offline mode or supplementary data)
	kubeletLog, ok := data.GetRawFile("logs/kubelet.log")
	if !ok {
		kubeletLog, ok = data.GetRawFile("daemon_status/kubelet_status")
		if !ok {
			// Try k8s/node_conditions for online collected data
			kubeletLog, ok = data.GetRawFile("k8s/node_conditions")
		}
	}

	if ok {
		content := string(kubeletLog)
		issues = append(issues, a.analyzeFromLogs(content)...)
	}

	return issues, nil
}

// analyzeOnline checks node status via K8s API
func (a *NodeStatusAnalyzer) analyzeOnline(
	ctx context.Context,
	data *types.DiagnosticData,
) ([]types.Issue, error) {
	nodes, err := a.fetchNodes(ctx, data)
	if err != nil {
		return nil, err
	}

	var issues []types.Issue
	for _, node := range nodes {
		issues = append(issues, a.checkNodeConditions(node)...)
		issues = append(issues, a.checkNodeTaints(node)...)
	}

	return issues, nil
}

// fetchNodes retrieves nodes from K8s API
func (a *NodeStatusAnalyzer) fetchNodes(
	ctx context.Context,
	data *types.DiagnosticData,
) ([]corev1.Node, error) {
	if data.NodeName != "" {
		node, err := data.K8sClient.CoreV1().Nodes().Get(ctx, data.NodeName, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		return []corev1.Node{*node}, nil
	}

	nodeList, err := data.K8sClient.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return nodeList.Items, nil
}

// checkNodeConditions checks all conditions for a single node
func (a *NodeStatusAnalyzer) checkNodeConditions(node corev1.Node) []types.Issue {
	var issues []types.Issue

	for _, cond := range node.Status.Conditions {
		if issue := a.evalNodeCondition(node.Name, cond); issue != nil {
			issue.AnalyzerName = a.Name()
			issues = append(issues, *issue)
		}
	}

	return issues
}

// evalNodeCondition evaluates a single node condition and returns an issue if problematic
func (a *NodeStatusAnalyzer) evalNodeCondition(
	nodeName string,
	cond corev1.NodeCondition,
) *types.Issue {
	switch cond.Type {
	case corev1.NodeReady:
		return a.evalNodeReady(nodeName, cond)
	case corev1.NodeDiskPressure:
		return a.evalDiskPressure(nodeName, cond)
	case corev1.NodeMemoryPressure:
		return a.evalMemoryPressure(nodeName, cond)
	case corev1.NodePIDPressure:
		return a.evalPIDPressure(nodeName, cond)
	case corev1.NodeNetworkUnavailable:
		return a.evalNetworkUnavailable(nodeName, cond)
	}
	return nil
}

func (a *NodeStatusAnalyzer) evalNodeReady(nodeName string, cond corev1.NodeCondition) *types.Issue {
	if cond.Status == corev1.ConditionTrue {
		return nil
	}
	return types.NewIssue(
		types.SeverityCritical,
		fmt.Sprintf("节点 %s NotReady", nodeName),
		"NODE_NOT_READY",
		fmt.Sprintf("节点 %s 处于NotReady状态: %s", nodeName, cond.Message),
		"k8s/node_status",
	).WithRemediation("检查kubelet和容器运行时状态; kubectl describe node " + nodeName)
}

func (a *NodeStatusAnalyzer) evalDiskPressure(nodeName string, cond corev1.NodeCondition) *types.Issue {
	if cond.Status != corev1.ConditionTrue {
		return nil
	}
	return types.NewIssue(
		types.SeverityWarning,
		fmt.Sprintf("节点 %s 磁盘压力", nodeName),
		"DISK_PRESSURE",
		fmt.Sprintf("节点 %s 存在磁盘压力: %s", nodeName, cond.Message),
		"k8s/node_status",
	).WithRemediation("清理磁盘空间: df -h; 删除无用镜像: crictl rmi --prune")
}

func (a *NodeStatusAnalyzer) evalMemoryPressure(nodeName string, cond corev1.NodeCondition) *types.Issue {
	if cond.Status != corev1.ConditionTrue {
		return nil
	}
	return types.NewIssue(
		types.SeverityWarning,
		fmt.Sprintf("节点 %s 内存压力", nodeName),
		"MEMORY_PRESSURE",
		fmt.Sprintf("节点 %s 存在内存压力: %s", nodeName, cond.Message),
		"k8s/node_status",
	).WithRemediation("检查内存使用: free -h; 优化Pod资源配置")
}

func (a *NodeStatusAnalyzer) evalPIDPressure(nodeName string, cond corev1.NodeCondition) *types.Issue {
	if cond.Status != corev1.ConditionTrue {
		return nil
	}
	return types.NewIssue(
		types.SeverityWarning,
		fmt.Sprintf("节点 %s PID压力", nodeName),
		"PID_PRESSURE",
		fmt.Sprintf("节点 %s 存在PID压力: %s", nodeName, cond.Message),
		"k8s/node_status",
	).WithRemediation("检查进程数: ps aux | wc -l; 检查PID资源限制")
}

func (a *NodeStatusAnalyzer) evalNetworkUnavailable(
	nodeName string,
	cond corev1.NodeCondition,
) *types.Issue {
	if cond.Status != corev1.ConditionTrue {
		return nil
	}
	return types.NewIssue(
		types.SeverityCritical,
		fmt.Sprintf("节点 %s 网络不可用", nodeName),
		"NETWORK_UNAVAILABLE",
		fmt.Sprintf("节点 %s 网络不可用: %s", nodeName, cond.Message),
		"k8s/node_status",
	).WithRemediation("检查CNI插件状态; 重启网络插件")
}

// checkNodeTaints checks node taints for issues
func (a *NodeStatusAnalyzer) checkNodeTaints(node corev1.Node) []types.Issue {
	var issues []types.Issue

	for _, taint := range node.Spec.Taints {
		if issue := a.evalUnschedulableTaint(node.Name, taint); issue != nil {
			issue.AnalyzerName = a.Name()
			issues = append(issues, *issue)
		}
	}

	return issues
}

func (a *NodeStatusAnalyzer) evalUnschedulableTaint(
	nodeName string,
	taint corev1.Taint,
) *types.Issue {
	if taint.Effect != corev1.TaintEffectNoSchedule {
		return nil
	}
	if taint.Key != "node.kubernetes.io/unschedulable" {
		return nil
	}
	return types.NewIssue(
		types.SeverityInfo,
		fmt.Sprintf("节点 %s 不可调度", nodeName),
		"NODE_UNSCHEDULABLE",
		fmt.Sprintf("节点 %s 被标记为不可调度", nodeName),
		"k8s/node_status",
	).WithRemediation("如需恢复调度: kubectl uncordon " + nodeName)
}

// analyzeFromLogs checks node status from log content
func (a *NodeStatusAnalyzer) analyzeFromLogs(content string) []types.Issue {
	var issues []types.Issue

	// Check for NotReady status
	if strings.Contains(content, "Node") && strings.Contains(content, "NotReady") {
		issue := types.NewIssue(
			types.SeverityCritical,
			"节点NotReady状态",
			"NODE_NOT_READY",
			"节点处于NotReady状态",
			"logs/kubelet.log",
		).WithRemediation("检查kubelet和容器运行时状态; kubectl describe node <node>")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	}

	// Check for DiskPressure
	if strings.Contains(content, "DiskPressure") {
		issue := types.NewIssue(
			types.SeverityWarning,
			"磁盘压力",
			"DISK_PRESSURE",
			"节点存在磁盘压力",
			"logs/kubelet.log",
		).WithRemediation("清理磁盘空间: df -h; 删除无用镜像: crictl rmi --prune")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	}

	// Check for MemoryPressure
	if strings.Contains(content, "MemoryPressure") {
		issue := types.NewIssue(
			types.SeverityWarning,
			"内存压力",
			"MEMORY_PRESSURE",
			"节点存在内存压力",
			"logs/kubelet.log",
		).WithRemediation("检查内存使用: free -h; 优化Pod资源配置")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	}

	// Check for pod eviction
	if strings.Contains(content, "evicted pod") || strings.Contains(content, "Evicted") {
		evictedCount := strings.Count(content, "evicted") + strings.Count(content, "Evicted")
		issue := types.NewIssue(
			types.SeverityWarning,
			"Pod被驱逐",
			"POD_EVICTED",
			fmt.Sprintf("Pod被驱逐 %d 次，可能由于资源不足", evictedCount),
			"logs/kubelet.log",
		).WithRemediation("检查节点资源使用情况; 调整eviction阈值")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	}

	return issues
}
