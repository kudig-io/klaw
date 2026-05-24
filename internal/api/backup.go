package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/kudig-io/klaw/internal/backup"
)

func (s *Server) handleListBackups(w http.ResponseWriter, r *http.Request) {
	cluster := mux.Vars(r)["cluster"]
	s.respondJSON(w, s.backupManager.List(cluster), http.StatusOK)
}

func (s *Server) handleGetBackup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cluster := vars["cluster"]
	name := vars["name"]

	item, err := s.backupManager.Get(cluster, name)
	if err != nil {
		s.respondError(w, err.Error(), http.StatusNotFound)
		return
	}
	s.respondJSON(w, item, http.StatusOK)
}

func (s *Server) handleCreateBackup(w http.ResponseWriter, r *http.Request) {
	cluster := mux.Vars(r)["cluster"]
	var req backup.CreateBackupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	item, err := s.backupManager.Create(cluster, req)
	if err != nil {
		s.respondError(w, err.Error(), http.StatusBadRequest)
		return
	}
	s.logAudit(r, "backup", "backup.create", map[string]string{"cluster": cluster, "backupName": item.Name}, "success", map[string]interface{}{"mode": item.Spec.BackupMode})
	s.respondJSON(w, item, http.StatusCreated)
}

func (s *Server) handleDeleteBackup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cluster := vars["cluster"]
	name := vars["name"]

	if err := s.backupManager.Delete(cluster, name); err != nil {
		s.respondError(w, err.Error(), http.StatusNotFound)
		return
	}
	s.logAudit(r, "backup", "backup.delete", map[string]string{"cluster": cluster, "backupName": name}, "success", nil)
	s.respondJSON(w, map[string]string{"message": "Backup deleted successfully"}, http.StatusOK)
}

func (s *Server) handleBackupSummary(w http.ResponseWriter, r *http.Request) {
	cluster := mux.Vars(r)["cluster"]
	s.respondJSON(w, s.backupManager.Summary(cluster), http.StatusOK)
}
