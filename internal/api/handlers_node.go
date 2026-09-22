// 节点相关 handlers
package api

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (s *Server) handleListNodes(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clusterName := vars["cluster"]

	nodes, err := s.resources.ListNodes(clusterName)
	if err != nil {
		s.respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.respondJSON(w, nodes, http.StatusOK)
}

func (s *Server) handleGetNode(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clusterName := vars["cluster"]
	nodeName := vars["name"]

	node, err := s.resources.GetNode(clusterName, nodeName)
	if err != nil {
		s.respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.respondJSON(w, node, http.StatusOK)
}

func (s *Server) handleGetNodeMetrics(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clusterName := vars["cluster"]

	metrics, err := s.resources.GetNodeMetrics(clusterName)
	if err != nil {
		s.respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.respondJSON(w, metrics, http.StatusOK)
}
