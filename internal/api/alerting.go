package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/kudig-io/klaw/internal/alerting"
)

func (s *Server) handleGetAlertRules(w http.ResponseWriter, r *http.Request) {
	cluster := mux.Vars(r)["cluster"]
	s.respondJSON(w, s.alertingManager.GetRules(cluster), http.StatusOK)
}

func (s *Server) handleCreateAlertRule(w http.ResponseWriter, r *http.Request) {
	cluster := mux.Vars(r)["cluster"]
	var rule alerting.Rule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		s.respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if rule.Cluster == "" {
		rule.Cluster = cluster
	}
	created, err := s.alertingManager.AddRule(rule)
	if err != nil {
		s.respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.respondJSON(w, created, http.StatusCreated)
}

func (s *Server) handleUpdateAlertRule(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cluster := vars["cluster"]
	ruleID := vars["id"]

	var rule alerting.Rule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		s.respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if rule.Cluster == "" {
		rule.Cluster = cluster
	}
	updated, err := s.alertingManager.UpdateRule(ruleID, rule)
	if err != nil {
		s.respondError(w, err.Error(), http.StatusNotFound)
		return
	}
	s.respondJSON(w, updated, http.StatusOK)
}

func (s *Server) handleDeleteAlertRule(w http.ResponseWriter, r *http.Request) {
	ruleID := mux.Vars(r)["id"]
	if err := s.alertingManager.DeleteRule(ruleID); err != nil {
		s.respondError(w, err.Error(), http.StatusNotFound)
		return
	}
	s.respondJSON(w, map[string]string{"message": "Alert rule deleted successfully"}, http.StatusOK)
}

func (s *Server) handleEvaluateAlerts(w http.ResponseWriter, r *http.Request) {
	cluster := mux.Vars(r)["cluster"]
	triggered, err := s.alertingManager.EvaluateCluster(cluster)
	if err != nil {
		s.respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.respondJSON(w, triggered, http.StatusOK)
}

func (s *Server) handleGetAlertHistory(w http.ResponseWriter, r *http.Request) {
	cluster := mux.Vars(r)["cluster"]
	limit := 100
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	s.respondJSON(w, s.alertingManager.GetHistory(cluster, limit), http.StatusOK)
}

func (s *Server) handleGetAlertStats(w http.ResponseWriter, r *http.Request) {
	cluster := mux.Vars(r)["cluster"]
	s.respondJSON(w, s.alertingManager.GetStats(cluster), http.StatusOK)
}

func (s *Server) handleAcknowledgeAlert(w http.ResponseWriter, r *http.Request) {
	alertID := mux.Vars(r)["id"]
	record, err := s.alertingManager.Acknowledge(alertID)
	if err != nil {
		s.respondError(w, err.Error(), http.StatusNotFound)
		return
	}
	s.respondJSON(w, record, http.StatusOK)
}

func (s *Server) handleResolveAlertRecord(w http.ResponseWriter, r *http.Request) {
	alertID := mux.Vars(r)["id"]
	record, err := s.alertingManager.Resolve(alertID)
	if err != nil {
		s.respondError(w, err.Error(), http.StatusNotFound)
		return
	}
	s.respondJSON(w, record, http.StatusOK)
}
