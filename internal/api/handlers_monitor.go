// 监控相关 handlers（监控状态/告警/历史指标）
package api

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/kudig-io/klaw/internal/monitoring"
)

// ========== Monitoring Handlers ==========

func (s *Server) handleGetMonitorStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clusterName := vars["cluster"]

	history := s.monitoringService.GetMetricsHistory(clusterName)
	status := map[string]interface{}{
		"cluster":    clusterName,
		"active":     len(history) > 0,
		"dataPoints": len(history),
	}

	s.respondJSON(w, status, http.StatusOK)
}

func (s *Server) handleGetMonitorAlerts(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clusterName := vars["cluster"]

	alerts := s.monitoringService.GetAlerts()
	clusterAlerts := make([]monitoring.Alert, 0, len(alerts))
	for _, alert := range alerts {
		if alert.Cluster == clusterName {
			clusterAlerts = append(clusterAlerts, *alert)
		}
	}

	s.respondJSON(w, clusterAlerts, http.StatusOK)
}

func (s *Server) handleGetMetricsHistory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clusterName := vars["cluster"]

	history := s.monitoringService.GetMetricsHistory(clusterName)
	s.respondJSON(w, history, http.StatusOK)
}
