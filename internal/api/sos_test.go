package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

func TestSOSStatusDisabled(t *testing.T) {
	s := &Server{router: mux.NewRouter(), authEnabled: false}
	s.setupSOSRoutes()
	rec := httptest.NewRecorder()
	s.router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/sos/status", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var got map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &got)
	if got["enabled"] != false {
		t.Fatalf("expected enabled=false, got %v", got)
	}
}

func TestSOSSessionDisabled(t *testing.T) {
	s := &Server{router: mux.NewRouter(), authEnabled: false}
	s.setupSOSRoutes()
	rec := httptest.NewRecorder()
	s.router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/sos/session", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 when sos not enabled, got %d", rec.Code)
	}
}

// fullHandler 组装与 Server.Start 一致的完整中间件链：指标 -> CORS -> 认证 -> 弃用提示 -> 路由
func (s *Server) fullHandler() http.Handler {
	return s.metrics.middleware(
		corsMiddleware(s.corsCfg,
			s.authMiddleware(
				deprecationMiddleware(s.router))))
}

func newAuthTestServer() *Server {
	s := &Server{router: mux.NewRouter(), authEnabled: true, authToken: "secret", metrics: newHTTPMetrics()}
	s.setupSOSRoutes()
	return s
}

// TestSOSSessionAuthChain 启用鉴权时经完整中间件链访问会话端点：
// 浏览器 WS 无法携带 Authorization 头，认证中间件须放行该路径，由 handler 的 checkToken 完成鉴权
func TestSOSSessionAuthChain(t *testing.T) {
	s := newAuthTestServer()
	h := s.fullHandler()

	// 无 token → 401
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/sos/session", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("no token: expected 401, got %d", rec.Code)
	}
	// 错误 token → 401
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/sos/session?token=wrong", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("wrong token: expected 401, got %d", rec.Code)
	}
	// 正确 token → 不再 401（sos 未配置，预期 503）
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/sos/session?token=secret", nil))
	if rec.Code == http.StatusUnauthorized {
		t.Fatalf("valid token: still 401, auth middleware did not bypass sos session path")
	}
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("valid token: expected 503 (sos not configured), got %d", rec.Code)
	}
}

func TestCheckTokenQueryAndHeader(t *testing.T) {
	s := &Server{router: mux.NewRouter(), authEnabled: true, authToken: "secret"}
	// 无凭证
	if s.checkToken(httptest.NewRequest(http.MethodGet, "/x", nil)) {
		t.Fatal("expected deny without credentials")
	}
	// Bearer 头
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Authorization", "Bearer secret")
	if !s.checkToken(req) {
		t.Fatal("expected allow with bearer header")
	}
	// token 查询参数
	if !s.checkToken(httptest.NewRequest(http.MethodGet, "/x?token=secret", nil)) {
		t.Fatal("expected allow with token query param")
	}
	// 错误 token
	if s.checkToken(httptest.NewRequest(http.MethodGet, "/x?token=wrong", nil)) {
		t.Fatal("expected deny with wrong token")
	}
	// authEnabled=false 时放行
	s2 := &Server{router: mux.NewRouter(), authEnabled: false}
	if !s2.checkToken(httptest.NewRequest(http.MethodGet, "/x", nil)) {
		t.Fatal("expected allow when auth disabled")
	}
}
