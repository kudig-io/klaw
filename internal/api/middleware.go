package api

import (
	"crypto/subtle"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/kudig-io/klaw/internal/config"
)

// authMiddleware Bearer token 认证：保护所有 /api/* 路由（健康检查与指标端点除外）
func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !s.authEnabled || !strings.HasPrefix(r.URL.Path, "/api") {
			next.ServeHTTP(w, r)
			return
		}
		// SOS 会话端点放行：浏览器 WebSocket 无法携带自定义 Authorization 头，
		// 由 handleSOSSession 的 checkToken（Bearer 头或 ?token= 查询参数）完成完整鉴权
		if r.URL.Path == "/api/v1/sos/session" {
			next.ServeHTTP(w, r)
			return
		}
		authz := r.Header.Get("Authorization")
		const prefix = "Bearer "
		if !strings.HasPrefix(authz, prefix) {
			s.respondError(w, "Unauthorized: missing bearer token", http.StatusUnauthorized)
			return
		}
		token := strings.TrimPrefix(authz, prefix)
		if subtle.ConstantTimeCompare([]byte(token), []byte(s.authToken)) != 1 {
			s.respondError(w, "Unauthorized: invalid token", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// corsMiddleware 基于白名单的跨域控制；未配置白名单时不下发跨域头（仅允许同源）
func corsMiddleware(cfg config.CORSConfig, next http.Handler) http.Handler {
	allowed := make(map[string]bool, len(cfg.AllowedOrigins))
	for _, o := range cfg.AllowedOrigins {
		allowed[o] = true
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && (allowed[origin] || allowed["*"]) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// httpMetrics 进程内 HTTP 请求计数（Prometheus 文本格式暴露，无外部依赖）
type httpMetrics struct {
	mu       sync.Mutex
	requests map[string]int64 // key: method|codeClass
	started  time.Time
}

func newHTTPMetrics() *httpMetrics {
	return &httpMetrics{requests: make(map[string]int64), started: time.Now()}
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func (m *httpMetrics) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		key := fmt.Sprintf("%s|%dxx", r.Method, rec.status/100)
		m.mu.Lock()
		m.requests[key]++
		m.mu.Unlock()
	})
}

// handleMetrics 输出 Prometheus 文本格式指标
func (s *Server) handleMetrics(w http.ResponseWriter, _ *http.Request) {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	fmt.Fprintf(w, "# HELP klaw_uptime_seconds Time since the server started.\n")
	fmt.Fprintf(w, "# TYPE klaw_uptime_seconds gauge\n")
	fmt.Fprintf(w, "klaw_uptime_seconds %.0f\n", time.Since(s.metrics.started).Seconds())

	fmt.Fprintf(w, "# HELP klaw_go_goroutines Number of goroutines.\n")
	fmt.Fprintf(w, "# TYPE klaw_go_goroutines gauge\n")
	fmt.Fprintf(w, "klaw_go_goroutines %d\n", runtime.NumGoroutine())

	fmt.Fprintf(w, "# HELP klaw_go_mem_alloc_bytes Bytes of allocated heap objects.\n")
	fmt.Fprintf(w, "# TYPE klaw_go_mem_alloc_bytes gauge\n")
	fmt.Fprintf(w, "klaw_go_mem_alloc_bytes %d\n", mem.Alloc)

	fmt.Fprintf(w, "# HELP klaw_http_requests_total Total HTTP requests by method and status class.\n")
	fmt.Fprintf(w, "# TYPE klaw_http_requests_total counter\n")
	s.metrics.mu.Lock()
	for key, count := range s.metrics.requests {
		parts := strings.SplitN(key, "|", 2)
		fmt.Fprintf(w, "klaw_http_requests_total{method=%q,code=%q} %d\n", parts[0], parts[1], count)
	}
	s.metrics.mu.Unlock()
}

// handleHealthz 存活探针：进程存活即返回 200
func (s *Server) handleHealthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

// handleReadyz 就绪探针：校验 Kubernetes 客户端可用
func (s *Server) handleReadyz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if _, err := s.k8sManager.GetClient(""); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`{"status":"unavailable","reason":"kubernetes client not ready"}`))
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ready"}`))
}
