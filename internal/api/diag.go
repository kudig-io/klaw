package api

import (
	"encoding/json"
	"net/http"

	"github.com/kudig-io/klaw/internal/diag"
)

func (s *Server) handleRunDiagnostics(w http.ResponseWriter, r *http.Request) {
	req := diag.DiagnosisRequest{
		Kubeconfig: r.URL.Query().Get("kubeconfig"),
		Context:    r.URL.Query().Get("context"),
		NodeName:   r.URL.Query().Get("node"),
		Namespace:  r.URL.Query().Get("namespace"),
		// ?ai=false 可显式关闭 AI 摘要，避免每次诊断都产生 LLM 调用开销
		DisableAI: r.URL.Query().Get("ai") == "false",
	}

	result, err := diag.RunOnlineDiagnostics(r.Context(), req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleDiagAnalyzers(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"registeredAnalyzers": diag.RegisteredAnalyzerCount(),
	})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
