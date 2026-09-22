// 集群概览与命名空间/事件 handlers
package api

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func (s *Server) handleGetClusters(w http.ResponseWriter, r *http.Request) {
	clusters := s.k8sManager.GetClusters()
	s.respondJSON(w, clusters, http.StatusOK)
}

func (s *Server) handleGetCluster(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	clusters := s.k8sManager.GetClusters()
	for _, cluster := range clusters {
		if cluster.Name == name {
			s.respondJSON(w, cluster, http.StatusOK)
			return
		}
	}

	s.respondError(w, "Cluster not found", http.StatusNotFound)
}

func (s *Server) handleGetClusterStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clusterName := vars["name"]

	nodes, err := s.resources.ListNodes(clusterName)
	if err != nil {
		s.respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pods, err := s.resources.ListPods(clusterName, "")
	if err != nil {
		s.respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	readyNodes := 0
	for _, node := range nodes {
		for _, condition := range node.Status.Conditions {
			if condition.Type == "Ready" && condition.Status == "True" {
				readyNodes++
			}
		}
	}

	runningPods := 0
	pendingPods := 0
	failedPods := 0
	for _, pod := range pods {
		switch pod.Status.Phase {
		case "Running":
			runningPods++
		case "Pending":
			pendingPods++
		case "Failed":
			failedPods++
		}
	}

	status := map[string]interface{}{
		"cluster": clusterName,
		"nodes": map[string]int{
			"total":    len(nodes),
			"ready":    readyNodes,
			"notReady": len(nodes) - readyNodes,
		},
		"pods": map[string]int{
			"total":   len(pods),
			"running": runningPods,
			"pending": pendingPods,
			"failed":  failedPods,
		},
		"timestamp": time.Now().Format(time.RFC3339),
	}

	s.respondJSON(w, status, http.StatusOK)
}

func (s *Server) handleGetClusterMetrics(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clusterName := vars["name"]

	clusterMetrics, err := s.metricsCollector.CollectClusterMetrics(clusterName)
	if err != nil {
		s.respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.respondJSON(w, clusterMetrics, http.StatusOK)
}

func (s *Server) handleGetNamespaces(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clusterName := vars["name"]

	namespaces, err := s.resources.ListNamespaces(clusterName)
	if err != nil {
		s.respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.respondJSON(w, namespaces, http.StatusOK)
}

func (s *Server) handleGetEvents(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clusterName := vars["cluster"]
	namespace := vars["namespace"]

	events, err := s.resources.ListEvents(clusterName, namespace)
	if err != nil {
		s.respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.respondJSON(w, events, http.StatusOK)
}
