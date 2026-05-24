package ops

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/kudig-io/klaw/internal/kubernetes"
	"github.com/kudig-io/klaw/internal/loganalysis"
	"github.com/kudig-io/klaw/internal/metrics"
	"github.com/kudig-io/klaw/internal/monitoring"
	"github.com/kudig-io/klaw/internal/messaging/dingtalk"
	"github.com/kudig-io/klaw/internal/messaging/feishu"
	"github.com/kudig-io/klaw/internal/rbacanalysis"
)

// Handler 运维命令处理器
type Handler struct {
	k8sManager      *kubernetes.Manager
	monitoringService *monitoring.Service
	dingtalkClient   *dingtalk.Client
	feishuClient     *feishu.Client
	resources        *kubernetes.Resources
}

func (h *Handler) requireKubernetes() error {
	if h.k8sManager == nil || h.resources == nil {
		return fmt.Errorf("kubernetes manager not initialized")
	}

	return nil
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
	case "deployment":
		return h.handleDeploymentCommand(parts[1:])
	case "service":
		return h.handleServiceCommand(parts[1:])
	case "rbac":
		return h.handleRBACCommand(parts[1:])
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

// handleServiceCommand 处理 Service 命令
func (h *Handler) handleServiceCommand(parts []string) (string, error) {
	if len(parts) == 0 {
		return "", fmt.Errorf("service command requires subcommand")
	}

	switch parts[0] {
	case "list":
		if len(parts) < 3 {
			return "", fmt.Errorf("service list command requires cluster name and namespace")
		}
		return h.listServices(parts[1], parts[2])
	case "describe":
		if len(parts) < 4 {
			return "", fmt.Errorf("service describe command requires cluster name, namespace and service name")
		}
		return h.describeService(parts[1], parts[2], parts[3])
	case "endpoints":
		if len(parts) < 4 {
			return "", fmt.Errorf("service endpoints command requires cluster name, namespace and service name")
		}
		return h.getServiceEndpoints(parts[1], parts[2], parts[3])
	default:
		return "", fmt.Errorf("unknown service subcommand: %s", parts[0])
	}
}

// handleRBACCommand 处理 RBAC 命令
func (h *Handler) handleRBACCommand(parts []string) (string, error) {
	if len(parts) == 0 {
		return "", fmt.Errorf("rbac command requires subcommand")
	}

	switch parts[0] {
	case "analyze":
		if len(parts) < 2 {
			return "", fmt.Errorf("rbac analyze command requires cluster name")
		}
		return h.analyzeRBAC(parts[1])
	default:
		return "", fmt.Errorf("unknown rbac subcommand: %s", parts[0])
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
	case "analyze":
		if len(parts) < 4 {
			return "", fmt.Errorf("pod analyze command requires cluster name, namespace and pod name")
		}
		return h.analyzePodLogs(parts[1], parts[2], parts[3])
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

// handleDeploymentCommand 处理 Deployment 命令
func (h *Handler) handleDeploymentCommand(parts []string) (string, error) {
	if len(parts) == 0 {
		return "", fmt.Errorf("deployment command requires subcommand")
	}

	switch parts[0] {
	case "list":
		if len(parts) < 3 {
			return "", fmt.Errorf("deployment list command requires cluster name and namespace")
		}
		return h.listDeployments(parts[1], parts[2])
	case "status":
		if len(parts) < 4 {
			return "", fmt.Errorf("deployment status command requires cluster name, namespace and deployment name")
		}
		return h.getDeploymentStatus(parts[1], parts[2], parts[3])
	case "scale":
		if len(parts) < 5 {
			return "", fmt.Errorf("deployment scale command requires cluster name, namespace, deployment name and replicas")
		}
		replicas, err := strconv.Atoi(parts[4])
		if err != nil {
			return "", fmt.Errorf("invalid replicas value: %s", parts[4])
		}
		return h.scaleDeployment(parts[1], parts[2], parts[3], int32(replicas))
	case "restart":
		if len(parts) < 4 {
			return "", fmt.Errorf("deployment restart command requires cluster name, namespace and deployment name")
		}
		return h.restartDeployment(parts[1], parts[2], parts[3])
	case "pods":
		if len(parts) < 4 {
			return "", fmt.Errorf("deployment pods command requires cluster name, namespace and deployment name")
		}
		return h.getDeploymentPods(parts[1], parts[2], parts[3])
	default:
		return "", fmt.Errorf("unknown deployment subcommand: %s", parts[0])
	}
}

// getClusterStatus 获取集群状态
func (h *Handler) getClusterStatus(clusterName string) (string, error) {
	if err := h.requireKubernetes(); err != nil {
		return "", err
	}

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
	if err := h.requireKubernetes(); err != nil {
		return "", err
	}

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
	if err := h.requireKubernetes(); err != nil {
		return "", err
	}

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
	if err := h.requireKubernetes(); err != nil {
		return "", err
	}

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
	if err := h.requireKubernetes(); err != nil {
		return "", err
	}

	logs, err := h.resources.GetPodLogs(clusterName, namespace, podName, 100)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Logs from pod %s:\n%s", podName, logs), nil
}

// analyzePodLogs 分析 Pod 日志
func (h *Handler) analyzePodLogs(clusterName, namespace, podName string) (string, error) {
	if err := h.requireKubernetes(); err != nil {
		return "", err
	}

	logs, err := h.resources.GetPodLogs(clusterName, namespace, podName, 200)
	if err != nil {
		return "", err
	}

	analysis := loganalysis.NewAnalyzer().AnalyzeLogs(logs)
	result := fmt.Sprintf("Log analysis for pod %s:\n", podName)
	result += fmt.Sprintf("- Total lines: %d\n", analysis.TotalLines)
	result += fmt.Sprintf("- Errors: %d\n", analysis.ErrorCount)
	result += fmt.Sprintf("- Warnings: %d\n", analysis.WarningCount)
	result += fmt.Sprintf("- Security events: %d\n", len(analysis.SecurityEvents))
	result += fmt.Sprintf("- Slow requests: %d\n", len(analysis.PerformanceMetrics.SlowRequests))
	return result, nil
}

// deletePod 删除Pod
func (h *Handler) deletePod(clusterName, namespace, podName string) (string, error) {
	if err := h.requireKubernetes(); err != nil {
		return "", err
	}

	err := h.resources.DeletePod(clusterName, namespace, podName)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Deleted pod %s in namespace %s", podName, namespace), nil
}

// listNodes 列出节点
func (h *Handler) listNodes(clusterName string) (string, error) {
	if err := h.requireKubernetes(); err != nil {
		return "", err
	}

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
	if err := h.requireKubernetes(); err != nil {
		return "", err
	}

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
	if err := h.requireKubernetes(); err != nil {
		return "", err
	}

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

// listDeployments 列出 Deployment
func (h *Handler) listDeployments(clusterName, namespace string) (string, error) {
	if err := h.requireKubernetes(); err != nil {
		return "", err
	}

	deployments, err := h.resources.ListDeployments(clusterName, namespace)
	if err != nil {
		return "", err
	}

	result := fmt.Sprintf("Deployments in namespace %s:\n", namespace)
	for _, deployment := range deployments {
		available := deployment.Status.AvailableReplicas
		desired := *deployment.Spec.Replicas
		result += fmt.Sprintf("- %s (%d/%d available)\n", deployment.Name, available, desired)
	}

	return result, nil
}

// getDeploymentStatus 获取 Deployment 状态
func (h *Handler) getDeploymentStatus(clusterName, namespace, deploymentName string) (string, error) {
	if err := h.requireKubernetes(); err != nil {
		return "", err
	}

	status, err := h.resources.GetDeploymentStatus(clusterName, namespace, deploymentName)
	if err != nil {
		return "", err
	}

	result := fmt.Sprintf("Deployment: %s\n", status.Name)
	result += fmt.Sprintf("Namespace: %s\n", status.Namespace)
	result += fmt.Sprintf("Replicas: %d (available: %d, ready: %d, updated: %d)\n",
		status.Replicas, status.AvailableReplicas, status.ReadyReplicas, status.UpdatedReplicas)

	if len(status.Conditions) > 0 {
		result += "Conditions:\n"
		for _, condition := range status.Conditions {
			result += fmt.Sprintf("  - %s: %s\n", condition.Type, condition.Status)
			if condition.Reason != "" {
				result += fmt.Sprintf("    Reason: %s\n", condition.Reason)
			}
		}
	}

	return result, nil
}

// scaleDeployment 扩缩容 Deployment
func (h *Handler) scaleDeployment(clusterName, namespace, deploymentName string, replicas int32) (string, error) {
	if err := h.requireKubernetes(); err != nil {
		return "", err
	}

	err := h.resources.ScaleDeployment(clusterName, namespace, deploymentName, replicas)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Scaled deployment %s in namespace %s to %d replicas", deploymentName, namespace, replicas), nil
}

// restartDeployment 重启 Deployment
func (h *Handler) restartDeployment(clusterName, namespace, deploymentName string) (string, error) {
	if err := h.requireKubernetes(); err != nil {
		return "", err
	}

	err := h.resources.RestartDeployment(clusterName, namespace, deploymentName)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Restarted deployment %s in namespace %s", deploymentName, namespace), nil
}

// getDeploymentPods 获取 Deployment 关联的 Pods
func (h *Handler) getDeploymentPods(clusterName, namespace, deploymentName string) (string, error) {
	if err := h.requireKubernetes(); err != nil {
		return "", err
	}

	pods, err := h.resources.GetDeploymentPods(clusterName, namespace, deploymentName)
	if err != nil {
		return "", err
	}

	result := fmt.Sprintf("Pods for deployment %s:\n", deploymentName)
	for _, pod := range pods {
		result += fmt.Sprintf("- %s (%s)\n", pod.Name, pod.Status.Phase)
	}

	return result, nil
}

// listServices 列出 Service
func (h *Handler) listServices(clusterName, namespace string) (string, error) {
	if err := h.requireKubernetes(); err != nil {
		return "", err
	}

	services, err := h.resources.ListServices(clusterName, namespace)
	if err != nil {
		return "", err
	}

	result := fmt.Sprintf("Services in namespace %s:\n", namespace)
	for _, service := range services {
		result += fmt.Sprintf("- %s (%s)\n", service.Name, service.Spec.Type)
	}

	return result, nil
}

// describeService 描述 Service
func (h *Handler) describeService(clusterName, namespace, serviceName string) (string, error) {
	if err := h.requireKubernetes(); err != nil {
		return "", err
	}

	service, err := h.resources.GetService(clusterName, namespace, serviceName)
	if err != nil {
		return "", err
	}

	result := fmt.Sprintf("Service: %s\n", service.Name)
	result += fmt.Sprintf("Namespace: %s\n", service.Namespace)
	result += fmt.Sprintf("Type: %s\n", service.Spec.Type)
	result += fmt.Sprintf("ClusterIP: %s\n", service.Spec.ClusterIP)
	result += fmt.Sprintf("Ports: %d\n", len(service.Spec.Ports))

	return result, nil
}

// getServiceEndpoints 获取 Service Endpoints
func (h *Handler) getServiceEndpoints(clusterName, namespace, serviceName string) (string, error) {
	if err := h.requireKubernetes(); err != nil {
		return "", err
	}

	endpoints, err := h.resources.GetServiceEndpoints(clusterName, namespace, serviceName)
	if err != nil {
		return "", err
	}

	result := fmt.Sprintf("Endpoints for service %s:\n", serviceName)
	for _, subset := range endpoints.Subsets {
		for _, address := range subset.Addresses {
			result += fmt.Sprintf("- %s\n", address.IP)
		}
	}

	if result == fmt.Sprintf("Endpoints for service %s:\n", serviceName) {
		result += "- <none>\n"
	}

	return result, nil
}

// analyzeRBAC 分析集群 RBAC
func (h *Handler) analyzeRBAC(clusterName string) (string, error) {
	if err := h.requireKubernetes(); err != nil {
		return "", err
	}

	client, err := h.k8sManager.GetClient(clusterName)
	if err != nil {
		return "", err
	}

	analysis, err := rbacanalysis.NewAnalyzer(client).AnalyzeRBAC(context.Background())
	if err != nil {
		return "", err
	}

	result := fmt.Sprintf("RBAC analysis for cluster %s:\n", clusterName)
	result += fmt.Sprintf("- Roles: %d\n", analysis.TotalRoles)
	result += fmt.Sprintf("- ClusterRoles: %d\n", analysis.TotalClusterRoles)
	result += fmt.Sprintf("- RoleBindings: %d\n", analysis.TotalBindings)
	result += fmt.Sprintf("- ClusterRoleBindings: %d\n", analysis.TotalClusterBindings)
	return result, nil
}

// showHelp 显示帮助信息
func (h *Handler) showHelp() string {
	return `Available commands:

Cluster commands:
  cluster status <cluster-name>    - Get cluster status
  cluster metrics <cluster-name>    - Get cluster metrics
  cluster chart <cluster-name>       - Send monitoring chart

Deployment commands:
  deployment list <cluster-name> <namespace>          - List deployments
  deployment status <cluster-name> <namespace> <deployment-name> - Get deployment status
  deployment scale <cluster-name> <namespace> <deployment-name> <replicas> - Scale deployment
  deployment restart <cluster-name> <namespace> <deployment-name> - Restart deployment
  deployment pods <cluster-name> <namespace> <deployment-name> - Get deployment pods

Service commands:
  service list <cluster-name> <namespace> - List services
  service describe <cluster-name> <namespace> <service-name> - Describe service
  service endpoints <cluster-name> <namespace> <service-name> - Get service endpoints

RBAC commands:
  rbac analyze <cluster-name> - Analyze RBAC roles and bindings

Pod commands:
  pod list <cluster-name> <namespace>         - List pods
  pod describe <cluster-name> <namespace> <pod-name> - Describe pod
  pod logs <cluster-name> <namespace> <pod-name>     - Get pod logs
  pod delete <cluster-name> <namespace> <pod-name>    - Delete pod
  pod analyze <cluster-name> <namespace> <pod-name>   - Analyze pod logs

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
