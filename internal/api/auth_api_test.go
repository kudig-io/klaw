package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// newAuthedTestHandler 构造开启认证配置的最小 Server，并返回与 Start() 中间件链
// 一致（authMiddleware 包裹路由）的 handler，用于验证认证行为本身。
func newAuthedTestHandler(t *testing.T, enabled bool, token string) http.Handler {
	t.Helper()
	s := newTestServer(t)
	s.authEnabled = enabled
	s.authToken = token
	s.router.HandleFunc("/api/v1/auth/verify", s.handleAuthVerify).Methods("GET")
	return s.authMiddleware(s.router)
}

func doAuthedRequest(handler http.Handler, authz string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/verify", nil)
	if authz != "" {
		req.Header.Set("Authorization", authz)
	}
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	return w
}

func TestAuthVerifyWithoutToken(t *testing.T) {
	handler := newAuthedTestHandler(t, true, "secret-token")

	w := doAuthedRequest(handler, "")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAuthVerifyWithInvalidToken(t *testing.T) {
	handler := newAuthedTestHandler(t, true, "secret-token")

	w := doAuthedRequest(handler, "Bearer wrong-token")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAuthVerifyWithValidToken(t *testing.T) {
	handler := newAuthedTestHandler(t, true, "secret-token")

	w := doAuthedRequest(handler, "Bearer secret-token")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	var resp map[string]bool
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !resp["authenticated"] || !resp["authEnabled"] {
		t.Fatalf("response = %v, want authenticated=true authEnabled=true", resp)
	}
}

func TestAuthVerifyWhenAuthDisabled(t *testing.T) {
	handler := newAuthedTestHandler(t, false, "")

	w := doAuthedRequest(handler, "")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	var resp map[string]bool
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp["authEnabled"] {
		t.Fatalf("authEnabled = true, want false")
	}
}
