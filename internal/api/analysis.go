package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/kudig-io/klaw/internal/loganalysis"
	"github.com/kudig-io/klaw/internal/rbacanalysis"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type AnalyzeLogsRequest struct {
	Logs string `json:"logs"`
}

func (s *Server) handleAnalyzePodLogs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clusterName := vars["cluster"]
	namespace := vars["namespace"]
	podName := vars["name"]

	tailLines := int64(300)
	if tailParam := r.URL.Query().Get("tailLines"); tailParam != "" {
		if val, err := strconv.ParseInt(tailParam, 10, 64); err == nil && val > 0 {
			tailLines = val
		}
	}

	logs, err := s.resources.GetPodLogs(clusterName, namespace, podName, tailLines)
	if err != nil {
		s.respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	analyzer := loganalysis.NewAnalyzer()
	s.respondJSON(w, analyzer.AnalyzeLogs(logs), http.StatusOK)
}

func (s *Server) handleAnalyzeRawLogs(w http.ResponseWriter, r *http.Request) {
	var req AnalyzeLogsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	analyzer := loganalysis.NewAnalyzer()
	s.respondJSON(w, analyzer.AnalyzeLogs(req.Logs), http.StatusOK)
}

func (s *Server) handleAnalyzeRBAC(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clusterName := vars["cluster"]

	client, err := s.k8sManager.GetClient(clusterName)
	if err != nil {
		s.respondError(w, err.Error(), http.StatusBadRequest)
		return
	}

	analyzer := rbacanalysis.NewAnalyzer(client)
	analysis, err := analyzer.AnalyzeRBAC(context.Background())
	if err != nil {
		s.respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.respondJSON(w, analysis, http.StatusOK)
}

func (s *Server) handleDeleteService(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clusterName := vars["cluster"]
	namespace := vars["namespace"]
	serviceName := vars["name"]

	client, err := s.k8sManager.GetClient(clusterName)
	if err != nil {
		s.respondError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := client.CoreV1().Services(namespace).Delete(r.Context(), serviceName, metav1.DeleteOptions{}); err != nil {
		s.respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.respondJSON(w, map[string]string{"message": "Service deleted successfully"}, http.StatusOK)
}
