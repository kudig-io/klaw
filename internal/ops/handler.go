package ops

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/kudig-io/klaw/internal/kubernetes"
	"github.com/kudig-io/klaw/internal/messaging/dingtalk"
	"github.com/kudig-io/klaw/internal/messaging/feishu"
	"github.com/kudig-io/klaw/internal/monitoring"
)

// handler.go 保留 Handler 结构、构造函数、命令路由/分发与共享辅助；
// 各命令域实现拆分在 handler_cluster.go / handler_pod.go / handler_node.go /
// handler_deployment.go / handler_service.go / handler_monitoring.go / handler_rbac.go。

// Handler 运维命令处理器
type Handler struct {
	k8sManager        *kubernetes.Manager
	monitoringService *monitoring.Service
	dingtalkClient    *dingtalk.Client
	feishuClient      *feishu.Client
	resources         *kubernetes.Resources
	allowDestructive  bool // 是否允许破坏性命令（pod delete 等），由配置注入
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
		k8sManager:        k8sManager,
		monitoringService: monitoringService,
		resources:         kubernetes.NewResources(k8sManager),
	}
}

// SetAllowDestructive 设置是否允许破坏性命令（默认关闭）
func (h *Handler) SetAllowDestructive(allow bool) {
	h.allowDestructive = allow
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

// formatAge returns a human-readable duration string for a given creation time.
func formatAge(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	}
}

// renderHelpMarkdown 生成统一 Markdown 表格格式的帮助信息
func renderHelpMarkdown() string {
	return `## Klaw ChatOps Commands

### Cluster

| Command | Arguments | Description |
|---------|-----------|-------------|
| cluster status | <cluster> | Get cluster status |
| cluster metrics | <cluster> | Get cluster metrics |
| cluster chart | <cluster> | Send monitoring chart |

### Resources

| Command | Arguments | Description |
|---------|-----------|-------------|
| pod list | <cluster> <namespace> | List pods |
| pod describe | <cluster> <namespace> <pod> | Describe pod |
| pod logs | <cluster> <namespace> <pod> | Get pod logs |
| pod delete | <cluster> <namespace> <pod> | Delete pod |
| pod analyze | <cluster> <namespace> <pod> | Analyze pod logs |
| deployment list | <cluster> <namespace> | List deployments |
| deployment status | <cluster> <namespace> <deploy> | Get deployment status |
| deployment scale | <cluster> <namespace> <deploy> <replicas> | Scale deployment |
| deployment restart | <cluster> <namespace> <deploy> | Restart deployment |
| service list | <cluster> <namespace> | List services |
| service describe | <cluster> <namespace> <service> | Describe service |
| service endpoints | <cluster> <namespace> <service> | Get service endpoints |
| node list | <cluster> | List nodes |
| node describe | <cluster> <node> | Describe node |
| node metrics | <cluster> | Get node metrics |

### Analysis

| Command | Arguments | Description |
|---------|-----------|-------------|
| rbac analyze | <cluster> | Analyze RBAC configuration |
| monitor status | <cluster> | Get monitoring status |
| monitor alerts | <cluster> | Get active alerts |
| monitor chart | <cluster> | Send monitoring chart |

**Shortcuts:** c=cluster, p=pod, n=node, d=deployment, s=service, r=rbac, m=monitor, ls=list, desc=describe, log=logs, del=delete, rm=delete
`
}

// showHelp 显示帮助信息（统一为 Markdown 表格格式，与 klaw 对齐）
func (h *Handler) showHelp() string {
	return renderHelpMarkdown()
}
