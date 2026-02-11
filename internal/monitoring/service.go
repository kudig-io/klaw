package monitoring

import (
	"fmt"
	"sync"
	"time"

	"github.com/kudig-io/klaw/internal/chart"
	"github.com/kudig-io/klaw/internal/kubernetes"
	"github.com/kudig-io/klaw/internal/metrics"
	"github.com/kudig-io/klaw/internal/messaging/dingtalk"
	"github.com/kudig-io/klaw/internal/messaging/feishu"
)

// Service 监控服务
type Service struct {
	k8sManager      *kubernetes.Manager
	dingtalkClient   *dingtalk.Client
	feishuClient     *feishu.Client
	metricsCollector  *metrics.Collector
	chartGenerator   *chart.Generator
	alerts          map[string]*Alert
	metricsHistory   map[string][]*metrics.ClusterMetrics
	historyMutex    sync.RWMutex
}

// NewService 创建监控服务
func NewService(k8sManager *kubernetes.Manager) *Service {
	return &Service{
		k8sManager:     k8sManager,
		metricsCollector: metrics.NewCollector(k8sManager),
		chartGenerator:  chart.NewGenerator(800, 600),
		alerts:         make(map[string]*Alert),
		metricsHistory: make(map[string][]*metrics.ClusterMetrics),
	}
}

// SetDingTalkClient 设置钉钉客户端
func (s *Service) SetDingTalkClient(client *dingtalk.Client) {
	s.dingtalkClient = client
}

// SetFeishuClient 设置飞书客户端
func (s *Service) SetFeishuClient(client *feishu.Client) {
	s.feishuClient = client
}

// Start 启动监控服务
func (s *Service) Start() {
	fmt.Println("Monitoring service started")

	// 启动指标收集循环
	go s.metricsCollectionLoop()

	// 启动图表发送循环
	go s.chartSendingLoop()

	// 启动告警检查循环
	go s.alertCheckingLoop()
}

// metricsCollectionLoop 指标收集循环
func (s *Service) metricsCollectionLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.collectMetrics()
		}
	}
}

// chartSendingLoop 图表发送循环
func (s *Service) chartSendingLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.sendCharts()
		}
	}
}

// alertCheckingLoop 告警检查循环
func (s *Service) alertCheckingLoop() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.checkAlerts()
		}
	}
}

// collectMetrics 收集指标
func (s *Service) collectMetrics() {
	clusters := s.k8sManager.GetClusters()

	for _, cluster := range clusters {
		metrics, err := s.metricsCollector.CollectClusterMetrics(cluster.Name)
		if err != nil {
			fmt.Printf("Failed to collect metrics for cluster %s: %v\n", cluster.Name, err)
			continue
		}

		// 保存指标历史
		s.historyMutex.Lock()
		if _, ok := s.metricsHistory[cluster.Name]; !ok {
			s.metricsHistory[cluster.Name] = make([]*metrics.ClusterMetrics, 0, 100)
		}
		s.metricsHistory[cluster.Name] = append(s.metricsHistory[cluster.Name], metrics)
		if len(s.metricsHistory[cluster.Name]) > 100 {
			s.metricsHistory[cluster.Name] = s.metricsHistory[cluster.Name][1:]
		}
		s.historyMutex.Unlock()

		fmt.Printf("Collected metrics for cluster %s: %d nodes, %d pods\n",
			cluster.Name, metrics.Nodes.Total, metrics.Pods.Total)
	}
}

// sendCharts 发送图表
func (s *Service) sendCharts() {
	s.historyMutex.RLock()
	defer s.historyMutex.RUnlock()

	for clusterName, history := range s.metricsHistory {
		if len(history) == 0 {
			continue
		}

		// 获取最新的指标
		latestMetrics := history[len(history)-1]

		// 生成集群监控图表
		chartData, err := s.chartGenerator.GenerateClusterMetricsChart(latestMetrics)
		if err != nil {
			fmt.Printf("Failed to generate chart for cluster %s: %v\n", clusterName, err)
			continue
		}

		// 发送到钉钉
		if s.dingtalkClient != nil {
			if err := s.dingtalkClient.SendChart(chartData, fmt.Sprintf("集群监控 - %s", clusterName)); err != nil {
				fmt.Printf("Failed to send chart to DingTalk: %v\n", err)
			}
		}

		// 发送到飞书
		if s.feishuClient != nil {
			if err := s.feishuClient.SendChart(chartData, fmt.Sprintf("集群监控 - %s", clusterName)); err != nil {
				fmt.Printf("Failed to send chart to Feishu: %v\n", err)
			}
		}

		fmt.Printf("Sent monitoring chart for cluster %s\n", clusterName)
	}
}

// checkAlerts 检查告警
func (s *Service) checkAlerts() {
	s.historyMutex.RLock()
	defer s.historyMutex.RUnlock()

	for clusterName, history := range s.metricsHistory {
		if len(history) == 0 {
			continue
		}

		latestMetrics := history[len(history)-1]

		// 检查节点状态
		if latestMetrics.Nodes.NotReady > 0 {
			s.createAlert(clusterName, "node", "warning",
				fmt.Sprintf("%d nodes are not ready", latestMetrics.Nodes.NotReady))
		}

		// 检查Pod状态
		if latestMetrics.Pods.Failed > 0 {
			s.createAlert(clusterName, "pod", "critical",
				fmt.Sprintf("%d pods have failed", latestMetrics.Pods.Failed))
		}

		if latestMetrics.Pods.Pending > 10 {
			s.createAlert(clusterName, "pod", "warning",
				fmt.Sprintf("%d pods are pending", latestMetrics.Pods.Pending))
		}
	}
}

// createAlert 创建告警
func (s *Service) createAlert(clusterName, alertType, level, message string) {
	alertID := fmt.Sprintf("%s-%s-%s", clusterName, alertType, time.Now().Format("20060102150405"))

	// 检查是否已经存在相同的告警
	if _, ok := s.alerts[alertID]; ok {
		return
	}

	alert := &Alert{
		ID:        alertID,
		Cluster:   clusterName,
		Type:      alertType,
		Level:     level,
		Message:   message,
		CreatedAt: time.Now(),
		Resolved:  false,
	}
	s.alerts[alertID] = alert

	// 发送告警通知
	s.sendAlert(alert)
}

// sendAlert 发送告警
func (s *Service) sendAlert(alert *Alert) {
	message := fmt.Sprintf("[Kubernetes Alert] %s - %s\nCluster: %s\nLevel: %s\nMessage: %s\nTime: %s",
		alert.Type, alert.ID, alert.Cluster, alert.Level, alert.Message, alert.CreatedAt.Format("2006-01-02 15:04:05"))

	// 发送到钉钉
	if s.dingtalkClient != nil {
		if err := s.dingtalkClient.SendMessage(message); err != nil {
			fmt.Printf("Failed to send alert to DingTalk: %v\n", err)
		}
	}

	// 发送到飞书
	if s.feishuClient != nil {
		if err := s.feishuClient.SendMessage(message); err != nil {
			fmt.Printf("Failed to send alert to Feishu: %v\n", err)
		}
	}
}

// GetMetricsHistory 获取指标历史
func (s *Service) GetMetricsHistory(clusterName string) []*metrics.ClusterMetrics {
	s.historyMutex.RLock()
	defer s.historyMutex.RUnlock()

	if history, ok := s.metricsHistory[clusterName]; ok {
		return history
	}
	return nil
}

// GetAlerts 获取所有告警
func (s *Service) GetAlerts() []*Alert {
	var alerts []*Alert
	for _, alert := range s.alerts {
		alerts = append(alerts, alert)
	}
	return alerts
}

// ResolveAlert 解决告警
func (s *Service) ResolveAlert(alertID string) error {
	alert, ok := s.alerts[alertID]
	if !ok {
		return fmt.Errorf("alert not found: %s", alertID)
	}

	alert.Resolved = true

	// 发送解决通知
	message := fmt.Sprintf("[Kubernetes Alert Resolved] %s - %s\nCluster: %s\nLevel: %s\nMessage: %s\nTime: %s",
		alert.Type, alert.ID, alert.Cluster, alert.Level, alert.Message, time.Now().Format("2006-01-02 15:04:05"))

	// 发送到钉钉
	if s.dingtalkClient != nil {
		if err := s.dingtalkClient.SendMessage(message); err != nil {
			fmt.Printf("Failed to send resolved alert to DingTalk: %v\n", err)
		}
	}

	// 发送到飞书
	if s.feishuClient != nil {
		if err := s.feishuClient.SendMessage(message); err != nil {
			fmt.Printf("Failed to send resolved alert to Feishu: %v\n", err)
		}
	}

	return nil
}
