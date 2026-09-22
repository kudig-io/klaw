package api

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (s *Server) handleListIngresses(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	items, err := s.resources.ListIngresses(vars["cluster"], vars["namespace"])
	if err != nil {
		s.respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.respondJSON(w, items, http.StatusOK)
}

func (s *Server) handleListNetworkPolicies(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	items, err := s.resources.ListNetworkPolicies(vars["cluster"], vars["namespace"])
	if err != nil {
		s.respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.respondJSON(w, items, http.StatusOK)
}

func (s *Server) handleListPVCs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	items, err := s.resources.ListPVCs(vars["cluster"], vars["namespace"])
	if err != nil {
		s.respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.respondJSON(w, items, http.StatusOK)
}

func (s *Server) handleListPersistentVolumes(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	items, err := s.resources.ListPersistentVolumes(vars["cluster"])
	if err != nil {
		s.respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.respondJSON(w, items, http.StatusOK)
}

func (s *Server) handleListStorageClasses(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	items, err := s.resources.ListStorageClasses(vars["cluster"])
	if err != nil {
		s.respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.respondJSON(w, items, http.StatusOK)
}
