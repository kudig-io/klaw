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
