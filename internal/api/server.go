package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/kudig-io/klaw/internal/kubernetes"
	"github.com/kudig-io/klaw/internal/metrics"
	"github.com/kudig-io/klaw/internal/monitoring"
)

type Server struct {
	k8sManager       *kubernetes.Manager
	monitoringService *monitoring.Service
	resources        *kubernetes.Resources
	metricsCollector  *metrics.Collector
	router           *mux.Router
}

func NewServer(k8sManager *kubernetes.Manager, monitoringService *monitoring.Service) *Server {
	return &Server{
		k8sManager:       k8sManager,
		monitoringService: monitoringService,
		resources:        kubernetes.NewResources(k8sManager),
		metricsCollector:  metrics.NewCollector(k8sManager),
		router:           mux.NewRouter(),
	}
}

func (s *Server) SetupRoutes() {
	s.router.HandleFunc("/api/clusters", s.handleGetClusters).Methods("GET")
	s.router.HandleFunc("/api/clusters/{name}", s.handleGetCluster).Methods("GET")
	s.router.HandleFunc("/api/clusters/{name}/status", s.handleGetClusterStatus).Methods("GET")
	s.router.HandleFunc("/api/clusters/{name}/metrics", s.handleGetClusterMetrics).Methods("GET")
	s.router.HandleFunc("/api/clusters/{name}/namespaces", s.handleGetNamespaces).Methods("GET")

	s.router.HandleFunc("/api/clusters/{cluster}/pods", s.handleListPods).Methods("GET")
	s.router.HandleFunc("/api/clusters/{cluster}/namespaces/{namespace}/pods", s.handleListPods).Methods("GET")
	s.router.HandleFunc("/api/clusters/{cluster}/namespaces/{namespace}/pods/{name}", s.handleGetPod).Methods("GET")
	s.router.HandleFunc("/api/clusters/{cluster}/namespaces/{namespace}/pods/{name}/logs", s.handleGetPodLogs).Methods("GET")
	s.router.HandleFunc("/api/clusters/{cluster}/namespaces/{namespace}/pods/{name}", s.handleDeletePod).Methods("DELETE")

	s.router.HandleFunc("/api/clusters/{cluster}/nodes", s.handleListNodes).Methods("GET")
	s.router.HandleFunc("/api/clusters/{cluster}/nodes/{name}", s.handleGetNode).Methods("GET")
	s.router.HandleFunc("/api/clusters/{cluster}/nodes/metrics", s.handleGetNodeMetrics).Methods("GET")

	s.router.HandleFunc("/api/clusters/{cluster}/events", s.handleGetEvents).Methods("GET")
	s.router.HandleFunc("/api/clusters/{cluster}/namespaces/{namespace}/events", s.handleGetEvents).Methods("GET")

	s.router.HandleFunc("/api/monitoring/{cluster}/status", s.handleGetMonitorStatus).Methods("GET")
	s.router.HandleFunc("/api/monitoring/{cluster}/alerts", s.handleGetMonitorAlerts).Methods("GET")
	s.router.HandleFunc("/api/monitoring/{cluster}/history", s.handleGetMetricsHistory).Methods("GET")

	s.router.PathPrefix("/").Handler(http.FileServer(http.Dir("./web/dist"))).Methods("GET")
}

func (s *Server) Start(port int) error {
	s.SetupRoutes()
	addr := fmt.Sprintf(":%d", port)
	log.Printf("Starting server on %s", addr)
	return http.ListenAndServe(addr, s.router)
}

func (s *Server) respondJSON(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func (s *Server) respondError(w http.ResponseWriter, message string, statusCode int) {
	s.respondJSON(w, map[string]string{"error": message}, statusCode)
}

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

func (s *Server) handleListNodes(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clusterName := vars["cluster"]

	nodes, err := s.resources.ListNodes(clusterName)
	if err != nil {
		s.respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.respondJSON(w, nodes, http.StatusOK)
}

func (s *Server) handleGetNode(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clusterName := vars["cluster"]
	nodeName := vars["name"]

	node, err := s.resources.GetNode(clusterName, nodeName)
	if err != nil {
		s.respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.respondJSON(w, node, http.StatusOK)
}

func (s *Server) handleGetNodeMetrics(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clusterName := vars["cluster"]

	metrics, err := s.resources.GetNodeMetrics(clusterName)
	if err != nil {
		s.respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.respondJSON(w, metrics, http.StatusOK)
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

func (s *Server) handleGetMonitorStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clusterName := vars["cluster"]

	history := s.monitoringService.GetMetricsHistory(clusterName)
	status := map[string]interface{}{
		"cluster":  clusterName,
		"active":   len(history) > 0,
		"dataPoints": len(history),
	}

	s.respondJSON(w, status, http.StatusOK)
}

func (s *Server) handleGetMonitorAlerts(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clusterName := vars["cluster"]

	alerts := s.monitoringService.GetAlerts()
	var clusterAlerts []monitoring.Alert
	for _, alert := range alerts {
		if alert.Cluster == clusterName {
			clusterAlerts = append(clusterAlerts, alert)
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
