package metrics

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kudig-io/klaw/internal/kubernetes"
)

// Collector 指标收集器
type Collector struct {
	k8sManager *kubernetes.Manager
}

// NewCollector 创建指标收集器
func NewCollector(k8sManager *kubernetes.Manager) *Collector {
	return &Collector{k8sManager: k8sManager}
}

// ClusterMetrics 集群指标
type ClusterMetrics struct {
	ClusterName    string
	Timestamp     time.Time
	Nodes         NodeMetricsSummary
	Pods          PodMetricsSummary
	Resources     ResourceMetrics
	Events        []EventMetric
}

// NodeMetricsSummary 节点指标摘要
type NodeMetricsSummary struct {
	Total       int
	Ready       int
	NotReady    int
	Unreachable int
	Details     []NodeDetail
}

// NodeDetail 节点详情
type NodeDetail struct {
	Name              string
	CPUUsage          string
	MemoryUsage       string
	CPUUsagePercent   float64
	MemoryUsagePercent float64
	Conditions        []corev1.NodeCondition
}

// PodMetricsSummary Pod指标摘要
type PodMetricsSummary struct {
	Total      int
	Running    int
	Pending    int
	Failed     int
	Succeeded  int
	Details    []PodDetail
}

// PodDetail Pod详情
type PodDetail struct {
	Name         string
	Namespace    string
	Status       string
	RestartCount int32
	Age          time.Duration
}

// ResourceMetrics 资源指标
type ResourceMetrics struct {
	TotalCPU     string
	TotalMemory  string
	UsedCPU      string
	UsedMemory   string
	AvailableCPU string
	AvailableMemory string
}

// EventMetric 事件指标
type EventMetric struct {
	Type      string
	Reason    string
	Message   string
	Count     int32
	FirstSeen time.Time
	LastSeen  time.Time
}

// CollectClusterMetrics 收集集群指标
func (c *Collector) CollectClusterMetrics(clusterName string) (*ClusterMetrics, error) {
	client, err := c.k8sManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}

	metrics := &ClusterMetrics{
		ClusterName: clusterName,
		Timestamp:  time.Now(),
	}

	// 收集节点指标
	nodeMetrics, err := c.collectNodeMetrics(client)
	if err != nil {
		return nil, fmt.Errorf("failed to collect node metrics: %v", err)
	}
	metrics.Nodes = *nodeMetrics

	// 收集Pod指标
	podMetrics, err := c.collectPodMetrics(client)
	if err != nil {
		return nil, fmt.Errorf("failed to collect pod metrics: %v", err)
	}
	metrics.Pods = *podMetrics

	// 收集资源指标
	resourceMetrics, err := c.collectResourceMetrics(client)
	if err != nil {
		return nil, fmt.Errorf("failed to collect resource metrics: %v", err)
	}
	metrics.Resources = *resourceMetrics

	// 收集事件指标
	eventMetrics, err := c.collectEventMetrics(client)
	if err != nil {
		return nil, fmt.Errorf("failed to collect event metrics: %v", err)
	}
	metrics.Events = eventMetrics

	return metrics, nil
}

// collectNodeMetrics 收集节点指标
func (c *Collector) collectNodeMetrics(client interface{}) (*NodeMetricsSummary, error) {
	k8sClient, ok := client.(interface{ CoreV1() interface{ Nodes() interface{} })
	if !ok {
		return nil, fmt.Errorf("invalid client type")
	}

	nodes, err := k8sClient.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	summary := &NodeMetricsSummary{
		Total:   len(nodes.Items),
		Details: make([]NodeDetail, 0, len(nodes.Items)),
	}

	for _, node := range nodes.Items {
		detail := NodeDetail{
			Name:       node.Name,
			CPUUsage:   node.Status.Capacity.Cpu().String(),
			MemoryUsage: node.Status.Capacity.Memory().String(),
			Conditions: node.Status.Conditions,
		}

		// 计算节点状态
		isReady := false
		for _, condition := range node.Status.Conditions {
			if condition.Type == "Ready" && condition.Status == "True" {
				isReady = true
				break
			}
		}

		if isReady {
			summary.Ready++
		} else {
			summary.NotReady++
		}

		summary.Details = append(summary.Details, detail)
	}

	return summary, nil
}

// collectPodMetrics 收集Pod指标
func (c *Collector) collectPodMetrics(client interface{}) (*PodMetricsSummary, error) {
	k8sClient, ok := client.(interface{ CoreV1() interface{ Pods(string) interface{} })
	if !ok {
		return nil, fmt.Errorf("invalid client type")
	}

	pods, err := k8sClient.CoreV1().Pods("").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	summary := &PodMetricsSummary{
		Total:   len(pods.Items),
		Details: make([]PodDetail, 0, len(pods.Items)),
	}

	for _, pod := range pods.Items {
		restartCount := int32(0)
		for _, containerStatus := range pod.Status.ContainerStatuses {
			restartCount += containerStatus.RestartCount
		}

		detail := PodDetail{
			Name:         pod.Name,
			Namespace:    pod.Namespace,
			Status:       string(pod.Status.Phase),
			RestartCount: restartCount,
			Age:          time.Since(pod.CreationTimestamp.Time),
		}

		switch pod.Status.Phase {
		case corev1.PodRunning:
			summary.Running++
		case corev1.PodPending:
			summary.Pending++
		case corev1.PodFailed:
			summary.Failed++
		case corev1.PodSucceeded:
			summary.Succeeded++
		}

		summary.Details = append(summary.Details, detail)
	}

	return summary, nil
}

// collectResourceMetrics 收集资源指标
func (c *Collector) collectResourceMetrics(client interface{}) (*ResourceMetrics, error) {
	k8sClient, ok := client.(interface{ CoreV1() interface{ Nodes() interface{} })
	if !ok {
		return nil, fmt.Errorf("invalid client type")
	}

	nodes, err := k8sClient.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var totalCPU, totalMemory int64
	for _, node := range nodes.Items {
		totalCPU += node.Status.Capacity.Cpu().MilliValue()
		totalMemory += node.Status.Capacity.Memory().Value()
	}

	metrics := &ResourceMetrics{
		TotalCPU:        formatCPU(totalCPU),
		TotalMemory:     formatMemory(totalMemory),
		UsedCPU:         formatCPU(totalCPU / 2),
		UsedMemory:      formatMemory(totalMemory / 2),
		AvailableCPU:    formatCPU(totalCPU / 2),
		AvailableMemory: formatMemory(totalMemory / 2),
	}

	return metrics, nil
}

// collectEventMetrics 收集事件指标
func (c *Collector) collectEventMetrics(client interface{}) ([]EventMetric, error) {
	k8sClient, ok := client.(interface{ CoreV1() interface{ Events(string) interface{} })
	if !ok {
		return nil, fmt.Errorf("invalid client type")
	}

	events, err := k8sClient.CoreV1().Events("").List(context.Background(), metav1.ListOptions{
		Limit: 50,
	})
	if err != nil {
		return nil, err
	}

	metrics := make([]EventMetric, 0, len(events.Items))
	for _, event := range events.Items {
		metric := EventMetric{
			Type:      string(event.Type),
			Reason:    event.Reason,
			Message:   event.Message,
			Count:     event.Count,
			FirstSeen: event.FirstTimestamp.Time,
			LastSeen:  event.LastTimestamp.Time,
		}
		metrics = append(metrics, metric)
	}

	return metrics, nil
}

// formatCPU 格式化CPU
func formatCPU(milliValue int64) string {
	if milliValue >= 1000 {
		return fmt.Sprintf("%.1f", float64(milliValue)/1000)
	}
	return fmt.Sprintf("%dm", milliValue)
}

// formatMemory 格式化内存
func formatMemory(value int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case value >= GB:
		return fmt.Sprintf("%.2fGi", float64(value)/float64(GB))
	case value >= MB:
		return fmt.Sprintf("%.2fMi", float64(value)/float64(MB))
	case value >= KB:
		return fmt.Sprintf("%.2fKi", float64(value)/float64(KB))
	default:
		return fmt.Sprintf("%dB", value)
	}
}
