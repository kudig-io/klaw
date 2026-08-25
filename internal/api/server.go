package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"time"

	"github.com/gorilla/mux"
	"github.com/kudig-io/klaw/internal/alerting"
	"github.com/kudig-io/klaw/internal/audit"
	"github.com/kudig-io/klaw/internal/automation"
	"github.com/kudig-io/klaw/internal/backup"
	"github.com/kudig-io/klaw/internal/config"
	"github.com/kudig-io/klaw/internal/kubernetes"
	"github.com/kudig-io/klaw/internal/metrics"
	"github.com/kudig-io/klaw/internal/monitoring"
	"github.com/kudig-io/klaw/internal/sos"
	"github.com/kudig-io/klaw/internal/storage"
	"github.com/kudig-io/klaw/internal/tenancy"
)

type Server struct {
	k8sManager        *kubernetes.Manager
	monitoringService *monitoring.Service
	alertingManager   *alerting.Manager
	backupManager     *backup.Manager
	tenancyManager    *tenancy.Manager
	auditLogger       *audit.Logger
	automationManager *automation.Manager
	resources         *kubernetes.Resources
	metricsCollector  *metrics.Collector
	sosManager        *sos.Manager
	router            *mux.Router
	authEnabled       bool
	authToken         string
	corsCfg           config.CORSConfig
	metrics           *httpMetrics
	httpServer        *http.Server
}

func NewServer(k8sManager *kubernetes.Manager, monitoringService *monitoring.Service, serverCfg config.ServerConfig, sosCfg config.SOSConfig) (*Server, error) {
	if serverCfg.Auth.Enabled && serverCfg.Auth.Token == "" {
		return nil, fmt.Errorf("server.auth.enabled is true but no API token configured (set server.auth.token or KLAW_API_TOKEN)")
	}
	resources := kubernetes.NewResources(k8sManager)
	store, err := storage.NewStore(filepath.Join("data", "klaw.db"))
	if err != nil {
		return nil, fmt.Errorf("init storage: %w", err)
	}
	autoMgr := automation.NewManager(store)
	if client, err := k8sManager.GetClient(""); err == nil {
		autoMgr.WithClientset(client)
	}
	var sosMgr *sos.Manager
	if sosCfg.Enabled {
		sosMgr, err = sos.NewManager(sosCfg, resources, "", serverCfg.CORS.AllowedOrigins)
		if err != nil {
			return nil, fmt.Errorf("init sos manager: %w", err)
		}
	}
	s := &Server{
		k8sManager:        k8sManager,
		monitoringService: monitoringService,
		alertingManager:   alerting.NewManager(resources, store),
		backupManager:     backup.NewManager(store),
		tenancyManager:    tenancy.NewManager(k8sManager, store),
		auditLogger:       audit.NewLogger(store),
		automationManager: autoMgr,
		resources:         resources,
		metricsCollector:  metrics.NewCollector(k8sManager),
		sosManager:        sosMgr,
		router:            mux.NewRouter(),
		authEnabled:       serverCfg.Auth.Enabled,
		authToken:         serverCfg.Auth.Token,
		corsCfg:           serverCfg.CORS,
		metrics:           newHTTPMetrics(),
	}
	// SOS 会话审计注入：仅记录会话开始/结束/工具调用元数据，不含音频与转写内容
	if sosMgr != nil {
		sosMgr.SetAuditLog(func(action, detail string) {
			if s.auditLogger == nil {
				return
			}
			s.auditLogger.Log(audit.AuditEvent{
				EventType: action,
				Category:  "sos",
				Severity:  "info",
				Source:    "sos",
				Action:    action,
				Details:   map[string]interface{}{"detail": detail},
			})
		})
	}
	s.SetupRoutes()
	return s, nil
}

func (s *Server) SetupRoutes() {
	// 运维端点：健康检查与指标（不经过认证）
	s.router.HandleFunc("/healthz", s.handleHealthz).Methods("GET")
	s.router.HandleFunc("/readyz", s.handleReadyz).Methods("GET")
	s.router.HandleFunc("/metrics", s.handleMetrics).Methods("GET")

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
	s.router.HandleFunc("/api/clusters/{cluster}/namespaces/{namespace}/services/{name}", s.handleDeleteService).Methods("DELETE")

	s.router.HandleFunc("/api/clusters/{cluster}/namespaces/{namespace}/pods/{name}/logs/analysis", s.handleAnalyzePodLogs).Methods("GET")
	s.router.HandleFunc("/api/analysis/logs", s.handleAnalyzeRawLogs).Methods("POST")
	s.router.HandleFunc("/api/clusters/{cluster}/rbac/analysis", s.handleAnalyzeRBAC).Methods("GET")

	s.router.HandleFunc("/api/monitoring/{cluster}/status", s.handleGetMonitorStatus).Methods("GET")
	s.router.HandleFunc("/api/monitoring/{cluster}/alerts", s.handleGetMonitorAlerts).Methods("GET")
	s.router.HandleFunc("/api/monitoring/{cluster}/history", s.handleGetMetricsHistory).Methods("GET")
	s.router.HandleFunc("/api/clusters/{cluster}/alerts/rules", s.handleGetAlertRules).Methods("GET")
	s.router.HandleFunc("/api/clusters/{cluster}/alerts/rules", s.handleCreateAlertRule).Methods("POST")
	s.router.HandleFunc("/api/clusters/{cluster}/alerts/rules/{id}", s.handleUpdateAlertRule).Methods("PUT")
	s.router.HandleFunc("/api/clusters/{cluster}/alerts/rules/{id}", s.handleDeleteAlertRule).Methods("DELETE")
	s.router.HandleFunc("/api/clusters/{cluster}/alerts/evaluate", s.handleEvaluateAlerts).Methods("POST")
	s.router.HandleFunc("/api/clusters/{cluster}/alerts/history", s.handleGetAlertHistory).Methods("GET")
	s.router.HandleFunc("/api/clusters/{cluster}/alerts/stats", s.handleGetAlertStats).Methods("GET")
	s.router.HandleFunc("/api/clusters/{cluster}/alerts/{id}/acknowledge", s.handleAcknowledgeAlert).Methods("POST")
	s.router.HandleFunc("/api/clusters/{cluster}/alerts/{id}/resolve", s.handleResolveAlertRecord).Methods("POST")
	s.router.HandleFunc("/api/clusters/{cluster}/backups", s.handleListBackups).Methods("GET")
	s.router.HandleFunc("/api/clusters/{cluster}/backups", s.handleCreateBackup).Methods("POST")
	s.router.HandleFunc("/api/clusters/{cluster}/backups/summary", s.handleBackupSummary).Methods("GET")
	s.router.HandleFunc("/api/clusters/{cluster}/backups/{name}", s.handleGetBackup).Methods("GET")
	s.router.HandleFunc("/api/clusters/{cluster}/backups/{name}", s.handleDeleteBackup).Methods("DELETE")
	s.router.HandleFunc("/api/tenants", s.handleListTenants).Methods("GET")
	s.router.HandleFunc("/api/tenants", s.handleCreateTenant).Methods("POST")
	s.router.HandleFunc("/api/tenants/stats", s.handleTenantStatistics).Methods("GET")
	s.router.HandleFunc("/api/tenants/{id}", s.handleGetTenant).Methods("GET")
	s.router.HandleFunc("/api/tenants/{id}", s.handleUpdateTenant).Methods("PUT")
	s.router.HandleFunc("/api/tenants/{id}", s.handleDeleteTenant).Methods("DELETE")
	s.router.HandleFunc("/api/tenant-users", s.handleListTenantUsers).Methods("GET")
	s.router.HandleFunc("/api/tenant-users", s.handleCreateTenantUser).Methods("POST")
	s.router.HandleFunc("/api/tenant-users/{id}", s.handleDeleteTenantUser).Methods("DELETE")
	s.router.HandleFunc("/api/audit/logs", s.handleAuditLogs).Methods("GET")
	s.router.HandleFunc("/api/audit/stats", s.handleAuditStats).Methods("GET")

	s.setupUnifiedV1Routes()

	s.router.HandleFunc("/api/v1/diag/run", s.handleRunDiagnostics).Methods("GET")
	s.router.HandleFunc("/api/v1/diag/analyzers", s.handleDiagAnalyzers).Methods("GET")
	s.setupAnalysisV1Routes()
	s.setupSOSRoutes()

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

func (s *Server) Start(port int) error {
	addr := fmt.Sprintf(":%d", port)
	log.Printf("Starting server on %s (auth=%v)", addr, s.authEnabled)

	// 中间件链：指标 -> CORS白名单 -> 认证 -> 弃用提示 -> 路由
	handler := s.metrics.middleware(
		corsMiddleware(s.corsCfg,
			s.authMiddleware(
				deprecationMiddleware(s.router))))

	s.httpServer = &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// Shutdown 优雅停机：等待存量请求完成后关闭服务
func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpServer == nil {
		return nil
	}
	return s.httpServer.Shutdown(ctx)
}

func deprecationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") && !strings.HasPrefix(r.URL.Path, "/api/v1/") {
			w.Header().Set("Deprecation", "true")
			w.Header().Set("Sunset", "2026-12-31")
			w.Header().Set("Link", `</api/v1/>; rel="successor-version"`)
		}
		next.ServeHTTP(w, r)
	})
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
