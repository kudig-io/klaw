package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/gorilla/mux"

	"github.com/kudig-io/klaw/internal/alerting"
	"github.com/kudig-io/klaw/internal/audit"
	"github.com/kudig-io/klaw/internal/automation"
	"github.com/kudig-io/klaw/internal/backup"
	"github.com/kudig-io/klaw/internal/storage"
	"github.com/kudig-io/klaw/internal/tenancy"
)

func newTestServer(t *testing.T) *Server {
	t.Helper()
	store, err := storage.NewStore(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	s := &Server{
		alertingManager:   alerting.NewManager(nil, store),
		backupManager:     backup.NewManager(store),
		tenancyManager:    tenancy.NewManager(nil, store),
		auditLogger:       audit.NewLogger(store),
		automationManager: automation.NewManager(store),
		router:            mux.NewRouter(),
	}
	s.setupAnalysisV1Routes()
	return s
}

func doRequest(t *testing.T, s *Server, method, path string, body interface{}) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	if path != "" {
		req = mux.SetURLVars(req, extractVars(path))
	}
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	return w
}

func extractVars(path string) map[string]string {
	return nil
}

func TestDiagAnalyzers(t *testing.T) {
	s := newTestServer(t)
	s.router.HandleFunc("/api/v1/diag/analyzers", s.handleDiagAnalyzers).Methods("GET")

	w := doRequest(t, s, "GET", "/api/v1/diag/analyzers", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if count, ok := resp["registeredAnalyzers"]; !ok {
		t.Error("missing registeredAnalyzers in response")
	} else {
		t.Logf("registeredAnalyzers: %v", count)
	}
}

func TestAutomationListScripts(t *testing.T) {
	s := newTestServer(t)
	w := doRequest(t, s, "GET", "/api/v1/automation/scripts", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	var scripts []interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &scripts); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(scripts) == 0 {
		t.Error("expected default scripts, got empty list")
	}
}

func TestAutomationCreateAndGetScript(t *testing.T) {
	s := newTestServer(t)

	w := doRequest(t, s, "POST", "/api/v1/automation/scripts", map[string]interface{}{
		"name":   "test-script",
		"type":   "custom",
		"script": "echo hello",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("create status = %d, body = %s", w.Code, w.Body.String())
	}
	var created automation.Script
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatalf("unmarshal created: %v", err)
	}
	if created.ID == "" {
		t.Fatal("created script has no ID")
	}

	w2 := doRequest(t, s, "GET", "/api/v1/automation/scripts/"+created.ID, nil)
	if w2.Code != http.StatusOK {
		t.Fatalf("get status = %d", w2.Code)
	}
}

func TestAutomationStatistics(t *testing.T) {
	s := newTestServer(t)
	w := doRequest(t, s, "GET", "/api/v1/automation/statistics", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
}

func TestAutomationHistory(t *testing.T) {
	s := newTestServer(t)
	w := doRequest(t, s, "GET", "/api/v1/automation/history", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
}

func TestAuditLogAndList(t *testing.T) {
	s := newTestServer(t)
	s.router.HandleFunc("/api/v1/audit/logs", func(w http.ResponseWriter, r *http.Request) {
		event := s.auditLogger.Log(audit.AuditEvent{
			EventType: "test.event",
			Category:  "test",
			Severity:  "info",
			Action:    "test",
		})
		writeJSON(w, http.StatusOK, event)
	}).Methods("POST")

	w := doRequest(t, s, "POST", "/api/v1/audit/logs", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("log status = %d", w.Code)
	}
}

func TestAlertingRules(t *testing.T) {
	s := newTestServer(t)
	s.router.HandleFunc("/api/v1/alerting/rules", func(w http.ResponseWriter, r *http.Request) {
		rules := s.alertingManager.GetRules("test-cluster")
		writeJSON(w, http.StatusOK, rules)
	}).Methods("GET")

	w := doRequest(t, s, "GET", "/api/v1/alerting/rules", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
}

func TestBackupListAndCreate(t *testing.T) {
	s := newTestServer(t)
	s.router.HandleFunc("/api/v1/backups", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, s.backupManager.List("test-cluster"))
	}).Methods("GET")

	w := doRequest(t, s, "GET", "/api/v1/backups", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("list status = %d", w.Code)
	}

	s.router.HandleFunc("/api/v1/backups", func(w http.ResponseWriter, r *http.Request) {
		b, err := s.backupManager.Create("test-cluster", backup.CreateBackupRequest{
			Name: "test-backup",
			StorageLocation: backup.StorageLocation{
				Bucket:            "test-bucket",
				Region:            "us-east-1",
				CredentialsSecret: "test-secret",
			},
		})
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusCreated, b)
	}).Methods("POST")

	w2 := doRequest(t, s, "POST", "/api/v1/backups", nil)
	if w2.Code != http.StatusCreated {
		t.Fatalf("create status = %d, body = %s", w2.Code, w2.Body.String())
	}
}

func TestWriteJSON(t *testing.T) {
	w := httptest.NewRecorder()
	writeJSON(w, http.StatusTeapot, map[string]string{"msg": "hi"})
	if w.Code != http.StatusTeapot {
		t.Fatalf("code = %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("Content-Type = %q", ct)
	}
}
