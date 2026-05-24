package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/kudig-io/klaw/internal/tenancy"
)

func (s *Server) handleListTenants(w http.ResponseWriter, r *http.Request) {
	cluster := r.URL.Query().Get("cluster")
	name := r.URL.Query().Get("name")
	namespace := r.URL.Query().Get("namespace")
	s.respondJSON(w, s.tenancyManager.ListTenants(cluster, name, namespace), http.StatusOK)
}

func (s *Server) handleGetTenant(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	tenant, err := s.tenancyManager.GetTenant(id)
	if err != nil {
		s.respondError(w, err.Error(), http.StatusNotFound)
		return
	}
	s.respondJSON(w, tenant, http.StatusOK)
}

func (s *Server) handleCreateTenant(w http.ResponseWriter, r *http.Request) {
	var tenant tenancy.Tenant
	if err := json.NewDecoder(r.Body).Decode(&tenant); err != nil {
		s.respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	item, err := s.tenancyManager.CreateTenant(tenant)
	if err != nil {
		s.respondError(w, err.Error(), http.StatusBadRequest)
		return
	}
	s.logAudit(r, "tenancy", "tenant.create", map[string]string{"tenantId": item.ID, "tenantName": item.Name, "cluster": item.Cluster}, "success", map[string]interface{}{"namespaces": item.Namespaces})
	s.respondJSON(w, item, http.StatusCreated)
}

func (s *Server) handleUpdateTenant(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var tenant tenancy.Tenant
	if err := json.NewDecoder(r.Body).Decode(&tenant); err != nil {
		s.respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	item, err := s.tenancyManager.UpdateTenant(id, tenant)
	if err != nil {
		s.respondError(w, err.Error(), http.StatusNotFound)
		return
	}
	s.logAudit(r, "tenancy", "tenant.update", map[string]string{"tenantId": item.ID, "tenantName": item.Name, "cluster": item.Cluster}, "success", map[string]interface{}{"namespaces": item.Namespaces})
	s.respondJSON(w, item, http.StatusOK)
}

func (s *Server) handleDeleteTenant(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if err := s.tenancyManager.DeleteTenant(id); err != nil {
		s.respondError(w, err.Error(), http.StatusNotFound)
		return
	}
	s.logAudit(r, "tenancy", "tenant.delete", map[string]string{"tenantId": id}, "success", map[string]interface{}{"runtimeCleanup": true})
	s.respondJSON(w, map[string]string{"message": "Tenant deleted successfully"}, http.StatusOK)
}

func (s *Server) handleListTenantUsers(w http.ResponseWriter, r *http.Request) {
	tenantID := r.URL.Query().Get("tenantId")
	role := r.URL.Query().Get("role")
	s.respondJSON(w, s.tenancyManager.ListUsers(tenantID, role), http.StatusOK)
}

func (s *Server) handleCreateTenantUser(w http.ResponseWriter, r *http.Request) {
	var user tenancy.TenantUser
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		s.respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	item, err := s.tenancyManager.AddUser(user)
	if err != nil {
		s.respondError(w, err.Error(), http.StatusBadRequest)
		return
	}
	s.logAudit(r, "tenancy", "tenant-user.create", map[string]string{
		"userId":         item.ID,
		"tenantId":       item.TenantID,
		"username":       item.Username,
		"subjectKind":    item.SubjectKind,
		"subjectName":    item.SubjectName,
		"subjectNamespace": item.SubjectNamespace,
	}, "success", map[string]interface{}{"namespaces": item.Namespaces})
	s.respondJSON(w, item, http.StatusCreated)
}

func (s *Server) handleDeleteTenantUser(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if err := s.tenancyManager.DeleteUser(id); err != nil {
		s.respondError(w, err.Error(), http.StatusNotFound)
		return
	}
	s.logAudit(r, "tenancy", "tenant-user.delete", map[string]string{"userId": id}, "success", nil)
	s.respondJSON(w, map[string]string{"message": "Tenant user deleted successfully"}, http.StatusOK)
}

func (s *Server) handleTenantStatistics(w http.ResponseWriter, r *http.Request) {
	s.respondJSON(w, s.tenancyManager.Statistics(), http.StatusOK)
}

func (s *Server) handleAuditLogs(w http.ResponseWriter, r *http.Request) {
	limit := 100
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	s.respondJSON(w, s.auditLogger.List(auditFilterFromRequest(r, limit)), http.StatusOK)
}

func (s *Server) handleAuditStats(w http.ResponseWriter, r *http.Request) {
	s.respondJSON(w, s.auditLogger.Statistics(), http.StatusOK)
}
