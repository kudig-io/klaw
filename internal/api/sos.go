package api

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/kudig-io/klaw/internal/sos"
)

// setupSOSRoutes 注册 SOS 语音应急快速对话路由（挂载在现有鉴权中间件链内）
func (s *Server) setupSOSRoutes() {
	s.router.HandleFunc("/api/v1/sos/status", s.handleSOSStatus).Methods("GET")
	s.router.HandleFunc("/api/v1/sos/session", s.handleSOSSession).Methods("GET")
}

func (s *Server) handleSOSStatus(w http.ResponseWriter, _ *http.Request) {
	if s.sosManager == nil {
		s.respondJSON(w, sos.StatusResponse{}, http.StatusOK)
		return
	}
	s.respondJSON(w, s.sosManager.Status(), http.StatusOK)
}

// handleSOSSession 浏览器 WebSocket 无法携带 Authorization 头，鉴权同时接受 ?token= 查询参数
func (s *Server) handleSOSSession(w http.ResponseWriter, r *http.Request) {
	if !s.checkToken(r) {
		s.respondError(w, "Unauthorized: missing or invalid token", http.StatusUnauthorized)
		return
	}
	if s.sosManager == nil {
		s.respondError(w, "sos not enabled", http.StatusServiceUnavailable)
		return
	}
	s.sosManager.HandleSessionWS(w, r)
}

// checkToken Bearer 头或 token 查询参数二选一（authEnabled=false 时放行）
func (s *Server) checkToken(r *http.Request) bool {
	if !s.authEnabled {
		return true
	}
	authz := r.Header.Get("Authorization")
	if strings.HasPrefix(authz, "Bearer ") {
		return subtle.ConstantTimeCompare([]byte(strings.TrimPrefix(authz, "Bearer ")), []byte(s.authToken)) == 1
	}
	q := r.URL.Query().Get("token")
	return q != "" && subtle.ConstantTimeCompare([]byte(q), []byte(s.authToken)) == 1
}
