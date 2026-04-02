package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
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

	// Pod APIs - 支持所有命名空间或特定命名空间
	s.router.HandleFunc("/api/clusters/{cluster}/pods", s.handleListAllPods).Methods("GET")
	s.router.HandleFunc("/api/clusters/{cluster}/namespaces/{namespace}/pods", s.handleListPods).Methods("GET")
	s.router.HandleFunc("/api/clusters/{cluster}/namespaces/{namespace}/pods/{name}", s.handleGetPod).Methods("GET")
	s.router.HandleFunc("/api/clusters/{cluster}/namespaces/{namespace}/pods/{name}/logs", s.handleGetPodLogs).Methods("GET")
	s.router.HandleFunc("/api/clusters/{cluster}/namespaces/{namespace}/pods/{name}", s.handleDeletePod).Methods("DELETE")

	s.router.HandleFunc("/api/clusters/{cluster}/nodes", s.handleListNodes).Methods("GET")
	s.router.HandleFunc("/api/clusters/{cluster}/nodes/{name}", s.handleGetNode).Methods("GET")
	s.router.HandleFunc("/api/clusters/{cluster}/nodes/metrics", s.handleGetNodeMetrics).Methods("GET")

	s.router.HandleFunc("/api/clusters/{cluster}/events", s.handleGetEvents).Methods("GET")
	s.router.HandleFunc("/api/clusters/{cluster}/namespaces/{namespace}/events", s.handleGetEvents).Methods("GET")

	// Deployment 管理 API - 支持所有命名空间或特定命名空间
	s.router.HandleFunc("/api/clusters/{cluster}/deployments", s.handleListAllDeployments).Methods("GET")
	s.router.HandleFunc("/api/clusters/{cluster}/namespaces/{namespace}/deployments", s.handleListDeployments).Methods("GET")
	s.router.HandleFunc("/api/clusters/{cluster}/namespaces/{namespace}/deployments/{name}", s.handleGetDeployment).Methods("GET")
	s.router.HandleFunc("/api/clusters/{cluster}/namespaces/{namespace}/deployments/{name}/scale", s.handleScaleDeployment).Methods("POST")
	s.router.HandleFunc("/api/clusters/{cluster}/namespaces/{namespace}/deployments/{name}/restart", s.handleRestartDeployment).Methods("POST")
	s.router.HandleFunc("/api/clusters/{cluster}/namespaces/{namespace}/deployments/{name}/pods", s.handleGetDeploymentPods).Methods("GET")
	s.router.HandleFunc("/api/clusters/{cluster}/namespaces/{namespace}/deployments/{name}/status", s.handleGetDeploymentStatus).Methods("GET")

	// Service 管理 API - 支持所有命名空间或特定命名空间
	s.router.HandleFunc("/api/clusters/{cluster}/services", s.handleListAllServices).Methods("GET")
	s.router.HandleFunc("/api/clusters/{cluster}/namespaces/{namespace}/services", s.handleListServices).Methods("GET")
	s.router.HandleFunc("/api/clusters/{cluster}/namespaces/{namespace}/services/{name}", s.handleGetService).Methods("GET")
	s.router.HandleFunc("/api/clusters/{cluster}/namespaces/{namespace}/services/{name}/endpoints", s.handleGetServiceEndpoints).Methods("GET")

	s.router.HandleFunc("/api/monitoring/{cluster}/status", s.handleGetMonitorStatus).Methods("GET")
	s.router.HandleFunc("/api/monitoring/{cluster}/alerts", s.handleGetMonitorAlerts).Methods("GET")
	s.router.HandleFunc("/api/monitoring/{cluster}/history", s.handleGetMetricsHistory).Methods("GET")

	// SPA 路由支持 - 所有非 API 请求返回 index.html
	s.router.PathPrefix("/").HandlerFunc(s.serveSPA).Methods("GET")
}

// serveSPA 为单页应用提供支持
func (s *Server) serveSPA(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	
	// 检查是否是 API 请求
	if len(path) >= 4 && path[:4] == "/api" {
		s.respondError(w, "Not Found", http.StatusNotFound)
		return
	}
	
	// 构建文件路径
	filePath := "./web/dist" + path
	
	// 检查文件是否存在
	if _, err := os.Stat(filePath); err == nil {
		// 文件存在，直接服务
		http.ServeFile(w, r, filePath)
		return
	}
	
	// 文件不存在，返回 index.html (让前端路由处理)
	http.ServeFile(w, r, "./web/dist/index.html")
}

// enableCORS 添加跨域支持
func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

func (s *Server) Start(port int) error {
	s.SetupRoutes()
	addr := fmt.Sprintf(":%d", port)
	log.Printf("Starting server on %s", addr)
	
	// 包装 router 添加 CORS 支持
	handler := enableCORS(s.router)
	return http.ListenAndServe(addr, handler)
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
	var clusterAlerts []monitoring.Alert
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

// ========== Service Handlers ==========

func (s *Server) handleListAllServices(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clusterName := vars["cluster"]

	services, err := s.resources.ListServices(clusterName, "")
	if err != nil {
		s.respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.respondJSON(w, services, http.StatusOK)
}

func (s *Server) handleListServices(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clusterName := vars["cluster"]
	namespace := vars["namespace"]

	services, err := s.resources.ListServices(clusterName, namespace)
	if err != nil {
		s.respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.respondJSON(w, services, http.StatusOK)
}

func (s *Server) handleGetService(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clusterName := vars["cluster"]
	namespace := vars["namespace"]
	serviceName := vars["name"]

	service, err := s.resources.GetService(clusterName, namespace, serviceName)
	if err != nil {
		s.respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.respondJSON(w, service, http.StatusOK)
}

func (s *Server) handleGetServiceEndpoints(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clusterName := vars["cluster"]
	namespace := vars["namespace"]
	serviceName := vars["name"]

	endpoints, err := s.resources.GetServiceEndpoints(clusterName, namespace, serviceName)
	if err != nil {
		s.respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.respondJSON(w, endpoints, http.StatusOK)
}
