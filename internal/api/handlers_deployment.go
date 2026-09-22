// Deployment 相关 handlers（列表/详情/伸缩/重启/Pods/状态）
package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

// ========== Deployment Handlers ==========

// handleListDeployments 列出特定命名空间的 Deployments
func (s *Server) handleListDeployments(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clusterName := vars["cluster"]
	namespace := vars["namespace"]

	deployments, err := s.resources.ListDeployments(clusterName, namespace)
	if err != nil {
		s.respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.respondJSON(w, deployments, http.StatusOK)
}

// handleListAllDeployments 列出所有命名空间的 Deployments
func (s *Server) handleListAllDeployments(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clusterName := vars["cluster"]

	deployments, err := s.resources.ListDeployments(clusterName, "")
	if err != nil {
		s.respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.respondJSON(w, deployments, http.StatusOK)
}

func (s *Server) handleGetDeployment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clusterName := vars["cluster"]
	namespace := vars["namespace"]
	deploymentName := vars["name"]

	deployment, err := s.resources.GetDeployment(clusterName, namespace, deploymentName)
	if err != nil {
		s.respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.respondJSON(w, deployment, http.StatusOK)
}

type ScaleDeploymentRequest struct {
	Replicas int32 `json:"replicas"`
}

func (s *Server) handleScaleDeployment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clusterName := vars["cluster"]
	namespace := vars["namespace"]
	deploymentName := vars["name"]

	var req ScaleDeploymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := s.resources.ScaleDeployment(clusterName, namespace, deploymentName, req.Replicas); err != nil {
		s.respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.respondJSON(w, map[string]interface{}{
		"message":  "Deployment scaled successfully",
		"replicas": req.Replicas,
	}, http.StatusOK)
}

func (s *Server) handleRestartDeployment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clusterName := vars["cluster"]
	namespace := vars["namespace"]
	deploymentName := vars["name"]

	if err := s.resources.RestartDeployment(clusterName, namespace, deploymentName); err != nil {
		s.respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.respondJSON(w, map[string]string{
		"message": "Deployment restarted successfully",
	}, http.StatusOK)
}

func (s *Server) handleGetDeploymentPods(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clusterName := vars["cluster"]
	namespace := vars["namespace"]
	deploymentName := vars["name"]

	pods, err := s.resources.GetDeploymentPods(clusterName, namespace, deploymentName)
	if err != nil {
		s.respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.respondJSON(w, pods, http.StatusOK)
}

func (s *Server) handleGetDeploymentStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clusterName := vars["cluster"]
	namespace := vars["namespace"]
	deploymentName := vars["name"]

	status, err := s.resources.GetDeploymentStatus(clusterName, namespace, deploymentName)
	if err != nil {
		s.respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.respondJSON(w, status, http.StatusOK)
}
