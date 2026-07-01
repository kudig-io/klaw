package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"k8s.io/client-go/kubernetes"

	"github.com/kudig-io/klaw/internal/automation"
	"github.com/kudig-io/klaw/internal/networkanalysis"
	"github.com/kudig-io/klaw/internal/storageanalysis"
)

func (s *Server) setupAnalysisV1Routes() {
	s.router.HandleFunc("/api/v1/automation/scripts", s.handleListScripts).Methods("GET")
	s.router.HandleFunc("/api/v1/automation/scripts", s.handleCreateScript).Methods("POST")
	s.router.HandleFunc("/api/v1/automation/scripts/{id}", s.handleGetScript).Methods("GET")
	s.router.HandleFunc("/api/v1/automation/scripts/{id}", s.handleUpdateScript).Methods("PUT")
	s.router.HandleFunc("/api/v1/automation/scripts/{id}", s.handleDeleteScript).Methods("DELETE")
	s.router.HandleFunc("/api/v1/automation/scripts/{id}/execute", s.handleExecuteScript).Methods("POST")
	s.router.HandleFunc("/api/v1/automation/history", s.handleScriptHistory).Methods("GET")
	s.router.HandleFunc("/api/v1/automation/statistics", s.handleScriptStatistics).Methods("GET")

	s.router.HandleFunc("/api/v1/analysis/network", s.handleNetworkAnalysis).Methods("GET")
	s.router.HandleFunc("/api/v1/analysis/storage", s.handleStorageAnalysis).Methods("GET")
}

func (s *Server) handleListScripts(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.automationManager.List(automation.ScriptFilter{Limit: 100}))
}

func (s *Server) handleCreateScript(w http.ResponseWriter, r *http.Request) {
	var script automation.Script
	if err := json.NewDecoder(r.Body).Decode(&script); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	created, err := s.automationManager.Add(script)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) handleGetScript(w http.ResponseWriter, r *http.Request) {
	script, err := s.automationManager.Get(mux.Vars(r)["id"])
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, script)
}

func (s *Server) handleUpdateScript(w http.ResponseWriter, r *http.Request) {
	var updates automation.Script
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	updated, err := s.automationManager.Update(mux.Vars(r)["id"], updates)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (s *Server) handleDeleteScript(w http.ResponseWriter, r *http.Request) {
	if err := s.automationManager.Delete(mux.Vars(r)["id"]); err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (s *Server) handleExecuteScript(w http.ResponseWriter, r *http.Request) {
	var params map[string]interface{}
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&params)
	}
	exec, err := s.automationManager.Execute(context.Background(), mux.Vars(r)["id"], "api", params)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, exec)
}

func (s *Server) handleScriptHistory(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.automationManager.History(100))
}

func (s *Server) handleScriptStatistics(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.automationManager.Statistics())
}

func (s *Server) handleNetworkAnalysis(w http.ResponseWriter, r *http.Request) {
	cs, err := s.getK8sClientset()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	result, err := networkanalysis.NewAnalyzer(cs).AnalyzeNetwork(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleStorageAnalysis(w http.ResponseWriter, r *http.Request) {
	cs, err := s.getK8sClientset()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	result, err := storageanalysis.NewAnalyzer(cs).AnalyzeStorage(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) getK8sClientset() (kubernetes.Interface, error) {
	return s.k8sManager.GetClient("")
}
