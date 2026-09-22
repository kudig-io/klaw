// Pod 相关 handlers（列表/详情/日志/删除）
package api

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// handleListPods 列出特定命名空间的 Pods
func (s *Server) handleListPods(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clusterName := vars["cluster"]
	namespace := vars["namespace"]

	pods, err := s.resources.ListPods(clusterName, namespace)
	if err != nil {
		s.respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.respondJSON(w, pods, http.StatusOK)
}

// handleListAllPods 列出所有命名空间的 Pods
func (s *Server) handleListAllPods(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clusterName := vars["cluster"]

	pods, err := s.resources.ListPods(clusterName, "")
	if err != nil {
		s.respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.respondJSON(w, pods, http.StatusOK)
}

func (s *Server) handleGetPod(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clusterName := vars["cluster"]
	namespace := vars["namespace"]
	podName := vars["name"]

	pod, err := s.resources.GetPod(clusterName, namespace, podName)
	if err != nil {
		s.respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.respondJSON(w, pod, http.StatusOK)
}

func (s *Server) handleGetPodLogs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clusterName := vars["cluster"]
	namespace := vars["namespace"]
	podName := vars["name"]

	tailLines := int64(100)
	if tailParam := r.URL.Query().Get("tailLines"); tailParam != "" {
		if val, err := strconv.ParseInt(tailParam, 10, 64); err == nil {
			tailLines = val
		}
	}

	logs, err := s.resources.GetPodLogs(clusterName, namespace, podName, tailLines)
	if err != nil {
		s.respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.respondJSON(w, map[string]string{"logs": logs}, http.StatusOK)
}

func (s *Server) handleDeletePod(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clusterName := vars["cluster"]
	namespace := vars["namespace"]
	podName := vars["name"]

	err := s.resources.DeletePod(clusterName, namespace, podName)
	if err != nil {
		s.respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.respondJSON(w, map[string]string{"message": "Pod deleted successfully"}, http.StatusOK)
}
