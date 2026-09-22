package api

import (
	"net/http"
)

// handleAuthVerify 前端启动时的认证探测端点。
// 认证开启时由 authMiddleware 统一校验 Bearer token，能到达这里即代表凭证有效；
// 响应中的 authEnabled 供前端决定是否展示「清除凭证」入口。
func (s *Server) handleAuthVerify(w http.ResponseWriter, _ *http.Request) {
	s.respondJSON(w, map[string]bool{
		"authenticated": true,
		"authEnabled":   s.authEnabled,
	}, http.StatusOK)
}
