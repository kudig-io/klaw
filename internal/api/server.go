package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
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

func NewServer(k8sManager *kubernetes.Manager, monitoringService *monitoring.Service, serverCfg config.ServerConfig, sosCfg config.SOSConfig, autoCfg config.AutomationConfig) (*Server, error) {
	if serverCfg.Auth.Enabled && serverCfg.Auth.Token == "" {
		return nil, fmt.Errorf("server.auth.enabled is true but no API token configured (set server.auth.token or KLAW_API_TOKEN)")
	}
	resources := kubernetes.NewResources(k8sManager)
	store, err := storage.NewStore(filepath.Join("data", "klaw.db"))
	if err != nil {
		return nil, fmt.Errorf("init storage: %w", err)
	}
	autoMgr := automation.NewManager(store).WithGuardEnabled(autoCfg.Guard.IsEnabled())
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
	// 自动化脚本审计注入：危险命令防护拦截等安全事件落审计日志
	autoMgr.SetAuditLog(func(action, detail string) {
		if s.auditLogger == nil {
			return
		}
		s.auditLogger.Log(audit.AuditEvent{
			EventType: action,
			Category:  "automation",
			Severity:  "warning",
			Source:    "automation",
			Action:    action,
			Details:   map[string]interface{}{"detail": detail},
		})
	})
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
