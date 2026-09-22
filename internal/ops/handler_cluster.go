package ops

import (
	"fmt"
	"strings"

	"github.com/kudig-io/klaw/internal/chart"
	"github.com/kudig-io/klaw/internal/metrics"
)

// handler_cluster.go 集群命令实现：集群状态、集群指标与资源图表。

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

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## Cluster: %s\n\n", clusterName))
	sb.WriteString("| Metric | Value |\n")
	sb.WriteString("|--------|-------|\n")
	sb.WriteString(fmt.Sprintf("| Nodes | %d |\n", len(nodes)))
	sb.WriteString(fmt.Sprintf("| Pods | %d |\n", len(pods)))

	return sb.String(), nil
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

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## Cluster Metrics: %s\n\n", clusterName))
	sb.WriteString("| Metric | Value |\n")
	sb.WriteString("|--------|-------|\n")
	sb.WriteString(fmt.Sprintf("| Timestamp | %s |\n", clusterMetrics.Timestamp.Format("2006-01-02 15:04:05")))
	sb.WriteString(fmt.Sprintf("| Nodes | %d (Ready: %d, NotReady: %d) |\n",
		clusterMetrics.Nodes.Total, clusterMetrics.Nodes.Ready, clusterMetrics.Nodes.NotReady))
	sb.WriteString(fmt.Sprintf("| Pods | %d (Running: %d, Pending: %d, Failed: %d) |\n",
		clusterMetrics.Pods.Total, clusterMetrics.Pods.Running, clusterMetrics.Pods.Pending, clusterMetrics.Pods.Failed))
	sb.WriteString(fmt.Sprintf("| Total CPU | %s |\n", clusterMetrics.Resources.TotalCPU))
	sb.WriteString(fmt.Sprintf("| Total Memory | %s |\n", clusterMetrics.Resources.TotalMemory))

	return sb.String(), nil
}

// sendClusterChart 生成集群资源图表（ASCII）直接返回到会话
func (h *Handler) sendClusterChart(clusterName string) (string, error) {
	if err := h.requireKubernetes(); err != nil {
		return "", err
	}

	collector := metrics.NewCollector(h.k8sManager)
	clusterMetrics, err := collector.CollectClusterMetrics(clusterName)
	if err != nil {
		return "", err
	}

	chartBytes, err := chart.NewGenerator(60, 15).GenerateClusterMetricsChart(clusterMetrics)
	if err != nil {
		return "", fmt.Errorf("生成图表失败: %w", err)
	}

	// 返回图表内容本身，由消息平台插件负责投递，
	// 不再谎报 "Sent"（此前只发提示语、从不生成图表）
	return fmt.Sprintf("## Resource Chart: `%s`\n\n```\n%s\n```", clusterName, chartBytes), nil
}
