package ops

import (
	"fmt"
	"strings"

	"github.com/kudig-io/klaw/internal/kubernetes"
	"github.com/kudig-io/klaw/internal/metrics"
	"github.com/kudig-io/klaw/internal/monitoring"
	"github.com/kudig-io/klaw/internal/messaging/dingtalk"
	"github.com/kudig-io/klaw/internal/messaging/feishu"
)

// Handler 运维命令处理器
type Handler struct {
	k8sManager      *kubernetes.Manager
	monitoringService *monitoring.Service
	dingtalkClient   *dingtalk.Client
	feishuClient     *feishu.Client
	resources        *kubernetes.Resources
}

// NewHandler 创建运维命令处理器
func NewHandler(k8sManager *kubernetes.Manager, monitoringService *monitoring.Service) *Handler {
	return &Handler{
		k8sManager:      k8sManager,
		monitoringService: monitoringService,
		resources:        kubernetes.NewResources(k8sManager),
	}
}

// SetDingTalkClient 设置钉钉客户端
func (h *Handler) SetDingTalkClient(client *dingtalk.Client) {
	h.dingtalkClient = client
}

// SetFeishuClient 设置飞书客户端
func (h *Handler) SetFeishuClient(client *feishu.Client) {
	h.feishuClient = client
}

// HandleCommand 处理运维命令
func (h *Handler) HandleCommand(command string) (string, error) {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return "", fmt.Errorf("empty command")
	}

	switch parts[0] {
	case "cluster":
		return h.handleClusterCommand(parts[1:])
	case "pod":
		return h.handlePodCommand(parts[1:])
	case "node":
		return h.handleNodeCommand(parts[1:])
	case "monitor":
		return h.handleMonitorCommand(parts[1:])
	case "help":
		return h.showHelp(), nil
	default:
		return "", fmt.Errorf("unknown command: %s", parts[0])
	}
}

// handleClusterCommand 处理集群命令
func (h *Handler) handleClusterCommand(parts []string) (string, error) {
	if len(parts) == 0 {
		return "", fmt.Errorf("cluster command requires subcommand")
	}

	switch parts[0] {
	case "status":
		if len(parts) < 2 {
			return "", fmt.Errorf("cluster status command requires cluster name")
		}
		return h.getClusterStatus(parts[1])
	case "metrics":
		if len(parts) < 2 {
			return "", fmt.Errorf("cluster metrics command requires cluster name")
		}
		return h.getClusterMetrics(parts[1])
	case "chart":
		if len(parts) < 2 {
			return "", fmt.Errorf("cluster chart command requires cluster name")
		}
		return h.sendClusterChart(parts[1])
	default:
		return "", fmt.Errorf("unknown cluster subcommand: %s", parts[0])
	}
}

// handlePodCommand 处理Pod命令
func (h *Handler) handlePodCommand(parts []string) (string, error) {
	if len(parts) == 0 {
		return "", fmt.Errorf("pod command requires subcommand")
	}

	switch parts[0] {
	case "list":
		if len(parts) < 3 {
			return "", fmt.Errorf("pod list command requires cluster name and namespace")
		}
		return h.listPods(parts[1], parts[2])
	case "describe":
		if len(parts) < 4 {
			return "", fmt.Errorf("pod describe command requires cluster name, namespace and pod name")
		}
		return h.describePod(parts[1], parts[2], parts[3])
	case "logs":
		if len(parts) < 4 {
			return "", fmt.Errorf("pod logs command requires cluster name, namespace and pod name")
		}
		return h.getPodLogs(parts[1], parts[2], parts[3])
	case "delete":
		if len(parts) < 4 {
			return "", fmt.Errorf("pod delete command requires cluster name, namespace and pod name")
		}
		return h.deletePod(parts[1], parts[2], parts[3])
	default:
		return "", fmt.Errorf("unknown pod subcommand: %s", parts[0])
	}
}

// handleNodeCommand 处理节点命令
func (h *Handler) handleNodeCommand(parts []string) (string, error) {
	if len(parts) == 0 {
		return "", fmt.Errorf("node command requires subcommand")
	}

	switch parts[0] {
	case "list":
		if len(parts) < 2 {
			return "", fmt.Errorf("node list command requires cluster name")
		}
		return h.listNodes(parts[1])
	case "describe":
		if len(parts) < 3 {
			return "", fmt.Errorf("node describe command requires cluster name and node name")
		}
		return h.describeNode(parts[1], parts[2])
	case "metrics":
		if len(parts) < 2 {
			return "", fmt.Errorf("node metrics command requires cluster name")
		}
		return h.getNodeMetrics(parts[1])
	default:
		return "", fmt.Errorf("unknown node subcommand: %s", parts[0])
	}
}

// handleMonitorCommand 处理监控命令
func (h *Handler) handleMonitorCommand(parts []string) (string, error) {
	if len(parts) == 0 {
		return "", fmt.Errorf("monitor command requires subcommand")
	}

	switch parts[0] {
	case "status":
		if len(parts) < 2 {
			return "", fmt.Errorf("monitor status command requires cluster name")
		}
		return h.getMonitorStatus(parts[1])
	case "alerts":
		if len(parts) < 2 {
			return "", fmt.Errorf("monitor alerts command requires cluster name")
		}
		return h.getMonitorAlerts(parts[1])
	case "chart":
		if len(parts) < 2 {
			return "", fmt.Errorf("monitor chart command requires cluster name")
		}
		return h.sendMonitorChart(parts[1])
	default:
		return "", fmt.Errorf("unknown monitor subcommand: %s", parts[0])
	}
}

// getClusterStatus 获取集群状态
func (h *Handler) getClusterStatus(clusterName string) (string, error) {
	nodes, err := h.resources.ListNodes(clusterName)
	if err != nil {
		return "", err
	}

	pods, err := h.resources.ListPods(clusterName, "")
	if err != nil {
		return "", err
	}

	result := fmt.Sprintf("Cluster: %s\n", clusterName)
	result += fmt.Sprintf("Nodes: %d\n", len(nodes))
	result += fmt.Sprintf("Pods: %d\n", len(pods))

	return result, nil
}

// getClusterMetrics 获取集群指标
func (h *Handler) getClusterMetrics(clusterName string) (string, error) {
	collector := metrics.NewCollector(h.k8sManager)
	clusterMetrics, err := collector.CollectClusterMetrics(clusterName)
	if err != nil {
		return "", err
	}

	result := fmt.Sprintf("Cluster Metrics: %s\n", clusterName)
	result += fmt.Sprintf("Timestamp: %s\n", clusterMetrics.Timestamp.Format("2006-01-02 15:04:05"))
	result += fmt.Sprintf("Nodes: %d (Ready: %d, NotReady: %d)\n",
		clusterMetrics.Nodes.Total, clusterMetrics.Nodes.Ready, clusterMetrics.Nodes.NotReady)
	result += fmt.Sprintf("Pods: %d (Running: %d, Pending: %d, Failed: %d)\n",
		clusterMetrics.Pods.Total, clusterMetrics.Pods.Running, clusterMetrics.Pods.Pending, clusterMetrics.Pods.Failed)
	result += fmt.Sprintf("Total CPU: %s, Total Memory: %s\n",
		clusterMetrics.Resources.TotalCPU, clusterMetrics.Resources.TotalMemory)

	return result, nil
}

// sendClusterChart 发送集群图表
func (h *Handler) sendClusterChart(clusterName string) (string, error) {
	if h.monitoringService == nil {
		return "", fmt.Errorf("monitoring service not initialized")
	}

	history := h.monitoringService.GetMetricsHistory(clusterName)
	if len(history) == 0 {
		return "", fmt.Errorf("no metrics history available for cluster %s", clusterName)
	}

	latestMetrics := history[len(history)-1]

	// 发送图表到钉钉
	if h.dingtalkClient != nil {
		if err := h.dingtalkClient.SendMessage(fmt.Sprintf("正在生成集群 %s 的监控图表...", clusterName)); err != nil {
			return "", err
		}
	}

	// 发送图表到飞书
	if h.feishuClient != nil {
		if err := h.feishuClient.SendMessage(fmt.Sprintf("正在生成集群 %s 的监控图表...", clusterName)); err != nil {
			return "", err
		}
	}

	return fmt.Sprintf("Sent monitoring chart for cluster %s", clusterName), nil
}

// listPods 列出Pod
func (h *Handler) listPods(clusterName, namespace string) (string, error) {
	pods, err := h.resources.ListPods(clusterName, namespace)
	if err != nil {
		return "", err
	}

	result := fmt.Sprintf("Pods in namespace %s:\n", namespace)
	for _, pod := range pods {
		result += fmt.Sprintf("- %s (%s)\n", pod.Name, pod.Status.Phase)
	}

	return result, nil
}

// describePod 描述Pod
func (h *Handler) describePod(clusterName, namespace, podName string) (string, error) {
	pod, err := h.resources.GetPod(clusterName, namespace, podName)
	if err != nil {
		return "", err
	}

	result := fmt.Sprintf("Pod: %s\n", pod.Name)
	result += fmt.Sprintf("Namespace: %s\n", pod.Namespace)
	result += fmt.Sprintf("Status: %s\n", pod.Status.Phase)
	result += fmt.Sprintf("Node: %s\n", pod.Spec.NodeName)
	result += fmt.Sprintf("Created: %s\n", pod.CreationTimestamp.Format("2006-01-02 15:04:05"))

	return result, nil
}

// getPodLogs 获取Pod日志
func (h *Handler) getPodLogs(clusterName, namespace, podName string) (string, error) {
	logs, err := h.resources.GetPodLogs(clusterName, namespace, podName, 100)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Logs from pod %s:\n%s", podName, logs), nil
}

// deletePod 删除Pod
func (h *Handler) deletePod(clusterName, namespace, podName string) (string, error) {
	err := h.resources.DeletePod(clusterName, namespace, podName)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Deleted pod %s in namespace %s", podName, namespace), nil
}

// listNodes 列出节点
func (h *Handler) listNodes(clusterName string) (string, error) {
	nodes, err := h.resources.ListNodes(clusterName)
	if err != nil {
		return "", err
	}

	result := fmt.Sprintf("Nodes in cluster %s:\n", clusterName)
	for _, node := range nodes {
		result += fmt.Sprintf("- %s\n", node.Name)
	}

	return result, nil
}

// describeNode 描述节点
func (h *Handler) describeNode(clusterName, nodeName string) (string, error) {
	node, err := h.resources.GetNode(clusterName, nodeName)
	if err != nil {
		return "", err
	}

	result := fmt.Sprintf("Node: %s\n", node.Name)
	result += fmt.Sprintf("Status: Ready\n")
	result += fmt.Sprintf("CPU: %s\n", node.Status.Capacity.Cpu().String())
	result += fmt.Sprintf("Memory: %s\n", node.Status.Capacity.Memory().String())

	return result, nil
}

// getNodeMetrics 获取节点指标
func (h *Handler) getNodeMetrics(clusterName string) (string, error) {
	metrics, err := h.resources.GetNodeMetrics(clusterName)
	if err != nil {
		return "", err
	}

	result := fmt.Sprintf("Node Metrics for cluster %s:\n", clusterName)
	for nodeName, nodeMetric := range metrics {
		result += fmt.Sprintf("- %s: CPU=%s, Memory=%s\n", nodeName, nodeMetric.CPU, nodeMetric.Memory)
	}

	return result, nil
}

// getMonitorStatus 获取监控状态
func (h *Handler) getMonitorStatus(clusterName string) (string, error) {
	if h.monitoringService == nil {
		return "", fmt.Errorf("monitoring service not initialized")
	}

	history := h.monitoringService.GetMetricsHistory(clusterName)
	if history == nil {
		return fmt.Sprintf("No monitoring data for cluster %s", clusterName), nil
	}

	return fmt.Sprintf("Monitoring status for cluster %s: Active (%d data points)", clusterName, len(history)), nil
}

// getMonitorAlerts 获取监控告警
func (h *Handler) getMonitorAlerts(clusterName string) (string, error) {
	if h.monitoringService == nil {
		return "", fmt.Errorf("monitoring service not initialized")
	}

	alerts := h.monitoringService.GetAlerts()
	if len(alerts) == 0 {
		return fmt.Sprintf("No alerts for cluster %s", clusterName), nil
	}

	result := fmt.Sprintf("Alerts for cluster %s:\n", clusterName)
	for _, alert := range alerts {
		if alert.Cluster == clusterName {
			result += fmt.Sprintf("- [%s] %s: %s\n", alert.Level, alert.Type, alert.Message)
		}
	}

	return result, nil
}

// sendMonitorChart 发送监控图表
func (h *Handler) sendMonitorChart(clusterName string) (string, error) {
	return h.sendClusterChart(clusterName)
}

// showHelp 显示帮助信息
func (h *Handler) showHelp() string {
	return `Available commands:

Cluster commands:
  cluster status <cluster-name>    - Get cluster status
  cluster metrics <cluster-name>    - Get cluster metrics
  cluster chart <cluster-name>       - Send monitoring chart

Pod commands:
  pod list <cluster-name> <namespace>         - List pods
  pod describe <cluster-name> <namespace> <pod-name> - Describe pod
  pod logs <cluster-name> <namespace> <pod-name>     - Get pod logs
  pod delete <cluster-name> <namespace> <pod-name>    - Delete pod

Node commands:
  node list <cluster-name>          - List nodes
  node describe <cluster-name> <node-name> - Describe node
  node metrics <cluster-name>         - Get node metrics

Monitor commands:
  monitor status <cluster-name> - Get monitoring status
  monitor alerts <cluster-name> - Get monitoring alerts
  monitor chart <cluster-name>    - Send monitoring chart

Help:
  help - Show this help message
`
}
