package ops

import (
	"fmt"
	"strings"
)

// handler_monitoring.go 监控命令实现：监控状态、告警列表与监控图表。

// getMonitorStatus 获取监控状态
func (h *Handler) getMonitorStatus(clusterName string) (string, error) {
	if h.monitoringService == nil {
		return "", fmt.Errorf("monitoring service not initialized")
	}

	history := h.monitoringService.GetMetricsHistory(clusterName)
	if history == nil {
		return fmt.Sprintf("**Monitoring:** No data for cluster `%s`", clusterName), nil
	}

	return fmt.Sprintf("**Monitoring:** Cluster `%s` is Active (%d data points)", clusterName, len(history)), nil
}

// getMonitorAlerts 获取监控告警
func (h *Handler) getMonitorAlerts(clusterName string) (string, error) {
	if h.monitoringService == nil {
		return "", fmt.Errorf("monitoring service not initialized")
	}

	alerts := h.monitoringService.GetAlerts()
	clusterAlerts := []struct{ Level, Type, Message string }{}
	for _, alert := range alerts {
		if alert.Cluster == clusterName {
			clusterAlerts = append(clusterAlerts, struct{ Level, Type, Message string }{
				Level: alert.Level, Type: alert.Type, Message: alert.Message,
			})
		}
	}

	if len(clusterAlerts) == 0 {
		return fmt.Sprintf("**Alerts:** No active alerts for cluster `%s`", clusterName), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## Alerts for `%s`\n\n", clusterName))
	sb.WriteString("| Level | Type | Message |\n")
	sb.WriteString("|-------|------|---------|\n")
	for _, alert := range clusterAlerts {
		sb.WriteString(fmt.Sprintf("| %s | %s | %s |\n", alert.Level, alert.Type, alert.Message))
	}

	return sb.String(), nil
}

// sendMonitorChart 发送监控图表
func (h *Handler) sendMonitorChart(clusterName string) (string, error) {
	return h.sendClusterChart(clusterName)
}
