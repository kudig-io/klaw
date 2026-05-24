package api

import (
	"net/http"

	"github.com/kudig-io/klaw/internal/audit"
)

func (s *Server) logAudit(r *http.Request, category, action string, resource map[string]string, result string, details map[string]interface{}) {
	if s.auditLogger == nil {
		return
	}

	s.auditLogger.Log(audit.AuditEvent{
		EventType: action,
		Category:  category,
		Severity:  "info",
		Source:    "api",
		User:      r.Header.Get("X-User"),
		Action:    action,
		Resource:  resource,
		Result:    result,
		Details:   details,
		IPAddress: r.RemoteAddr,
		UserAgent: r.UserAgent(),
	})
}

func auditFilterFromRequest(r *http.Request, limit int) audit.AuditFilter {
	return audit.AuditFilter{
		EventType: r.URL.Query().Get("eventType"),
		Category:  r.URL.Query().Get("category"),
		Severity:  r.URL.Query().Get("severity"),
		User:      r.URL.Query().Get("user"),
		Limit:     limit,
	}
}
