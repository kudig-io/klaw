package monitoring

import (
	"context"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kudig-io/klaw/internal/kubernetes"
	"github.com/kudig-io/klaw/internal/messaging/dingtalk"
	"github.com/kudig-io/klaw/internal/messaging/feishu"
)

// Manager 监控管理器
type Manager struct {
	k8sManager   *kubernetes.Manager
	dingtalkClient *dingtalk.Client
	feishuClient   *feishu.Client
	alerts        map[string]*Alert
}

// Alert 告警信息
type Alert struct {
	ID        string    `json:"id"`
	Cluster   string    `json:"cluster"`
	Type      string    `json:"type"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
	Resolved  bool      `json:"resolved"`
}

// NewManager 创建监控管理器
func NewManager(k8sManager *kubernetes.Manager) *Manager {
	return &Manager{
		k8sManager: k8sManager,
		alerts:     make(map[string]*Alert),
	}
}

// SetDingTalkClient 设置钉钉客户端
func (m *Manager) SetDingTalkClient(client *dingtalk.Client) {
	m.dingtalkClient = client
}

// SetFeishuClient 设置飞书客户端
func (m *Manager) SetFeishuClient(client *feishu.Client) {
	m.feishuClient = client
}

// Start 启动监控管理器
func (m *Manager) Start() {
	fmt.Println("Monitoring manager started")

	// 启动监控循环
	go m.monitorLoop()
}

// monitorLoop 监控循环
func (m *Manager) monitorLoop() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.checkClusters()
		}
	}
}

// checkClusters 检查所有集群状态
func (m *Manager) checkClusters() {
	clusters := m.k8sManager.GetClusters()

	for _, cluster := range clusters {
		m.checkCluster(cluster.Name)
	}
}

// checkCluster 检查单个集群状态
func (m *Manager) checkCluster(clusterName string) {
	// 获取集群客户端
	client, err := m.k8sManager.GetClient(clusterName)
	if err != nil {
		// 创建告警
		alertID := fmt.Sprintf("%s-%s", clusterName, time.Now().Format("20060102150405"))
		alert := &Alert{
			ID:        alertID,
			Cluster:   clusterName,
			Type:      "connection",
			Level:     "critical",
			Message:   fmt.Sprintf("Failed to connect to cluster: %v", err),
			CreatedAt: time.Now(),
			Resolved:  false,
		}
		m.alerts[alertID] = alert
		m.sendAlert(alert)
		return
	}

	// 检查集群节点状态
	nodes, err := client.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		// 创建告警
		alertID := fmt.Sprintf("%s-%s", clusterName, time.Now().Format("20060102150405"))
		alert := &Alert{
			ID:        alertID,
			Cluster:   clusterName,
			Type:      "node",
			Level:     "critical",
			Message:   fmt.Sprintf("Failed to list nodes: %v", err),
			CreatedAt: time.Now(),
			Resolved:  false,
		}
		m.alerts[alertID] = alert
		m.sendAlert(alert)
		return
	}

	// 检查节点状态
	for _, node := range nodes.Items {
		for _, condition := range node.Status.Conditions {
			if condition.Type == "Ready" && condition.Status != "True" {
				// 创建告警
				alertID := fmt.Sprintf("%s-%s-%s", clusterName, node.Name, time.Now().Format("20060102150405"))
				alert := &Alert{
					ID:        alertID,
					Cluster:   clusterName,
					Type:      "node",
					Level:     "warning",
					Message:   fmt.Sprintf("Node %s is not ready: %s", node.Name, condition.Message),
					CreatedAt: time.Now(),
					Resolved:  false,
				}
				m.alerts[alertID] = alert
				m.sendAlert(alert)
			}
		}
	}
}

// sendAlert 发送告警通知
func (m *Manager) sendAlert(alert *Alert) {
	// 构建告警消息
	message := fmt.Sprintf("[Kubernetes Alert] %s - %s\nCluster: %s\nLevel: %s\nMessage: %s\nTime: %s",
		alert.Type, alert.ID, alert.Cluster, alert.Level, alert.Message, alert.CreatedAt.Format("2006-01-02 15:04:05"))

	// 发送到钉钉
	if m.dingtalkClient != nil {
		if err := m.dingtalkClient.SendMessage(message); err != nil {
			fmt.Printf("Failed to send alert to DingTalk: %v\n", err)
		}
	}

	// 发送到飞书
	if m.feishuClient != nil {
		if err := m.feishuClient.SendMessage(message); err != nil {
			fmt.Printf("Failed to send alert to Feishu: %v\n", err)
		}
	}
}

// GetAlerts 获取所有告警
func (m *Manager) GetAlerts() []*Alert {
	var alerts []*Alert
	for _, alert := range m.alerts {
		alerts = append(alerts, alert)
	}
	return alerts
}

// ResolveAlert 解决告警
func (m *Manager) ResolveAlert(alertID string) error {
	alert, ok := m.alerts[alertID]
	if !ok {
		return fmt.Errorf("alert not found: %s", alertID)
	}

	alert.Resolved = true

	// 发送解决通知
	message := fmt.Sprintf("[Kubernetes Alert Resolved] %s - %s\nCluster: %s\nLevel: %s\nMessage: %s\nTime: %s",
		alert.Type, alert.ID, alert.Cluster, alert.Level, alert.Message, time.Now().Format("2006-01-02 15:04:05"))

	// 发送到钉钉
	if m.dingtalkClient != nil {
		if err := m.dingtalkClient.SendMessage(message); err != nil {
			fmt.Printf("Failed to send resolved alert to DingTalk: %v\n", err)
		}
	}

	// 发送到飞书
	if m.feishuClient != nil {
		if err := m.feishuClient.SendMessage(message); err != nil {
			fmt.Printf("Failed to send resolved alert to Feishu: %v\n", err)
		}
	}

	return nil
}
