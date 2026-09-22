// Service 相关 handlers（列表/详情/Endpoints）
package api

import (
	"net/http"

	"github.com/gorilla/mux"
)

// ========== Service Handlers ==========

func (s *Server) handleListAllServices(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clusterName := vars["cluster"]

	services, err := s.resources.ListServices(clusterName, "")
	if err != nil {
		s.respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.respondJSON(w, services, http.StatusOK)
}

func (s *Server) handleListServices(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clusterName := vars["cluster"]
	namespace := vars["namespace"]

	services, err := s.resources.ListServices(clusterName, namespace)
	if err != nil {
		s.respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.respondJSON(w, services, http.StatusOK)
}

func (s *Server) handleGetService(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clusterName := vars["cluster"]
	namespace := vars["namespace"]
	serviceName := vars["name"]

	service, err := s.resources.GetService(clusterName, namespace, serviceName)
	if err != nil {
		s.respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.respondJSON(w, service, http.StatusOK)
}

func (s *Server) handleGetServiceEndpoints(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clusterName := vars["cluster"]
	namespace := vars["namespace"]
	serviceName := vars["name"]

	endpoints, err := s.resources.GetServiceEndpoints(clusterName, namespace, serviceName)
	if err != nil {
		s.respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.respondJSON(w, endpoints, http.StatusOK)
}
