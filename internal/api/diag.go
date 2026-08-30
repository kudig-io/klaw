package api

import (
	"encoding/json"
	"net/http"

	"github.com/kudig-io/klaw/internal/diag"
)

// diagRunSem 限制同时运行的诊断数量：全量诊断耗时长且对目标集群有采集压力，
// 不允许多个请求并发触发
var diagRunSem = make(chan struct{}, 1)

func (s *Server) handleRunDiagnostics(w http.ResponseWriter, r *http.Request) {
	select {
	case diagRunSem <- struct{}{}:
		defer func() { <-diagRunSem }()
	default:
		writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "another diagnosis is already running, retry later"})
		return
	}

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
