package audit

import (
	"path/filepath"
	"testing"

	"github.com/kudig-io/klaw/internal/storage"
)

func TestLoggerLifecycle(t *testing.T) {
	store, err := storage.NewStore(filepath.Join(t.TempDir(), "klaw.db"))
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}
	logger := NewLogger(store)

	event := logger.Log(AuditEvent{
		EventType: "tenant.create",
		Category:  "tenancy",
		Severity:  "info",
		User:      "tester",
		Action:    "tenant.create",
		Resource:  map[string]string{"tenantId": "tenant-1"},
	})

	if event.ID == "" {
		t.Fatal("Log() should assign an ID")
	}

	logs := logger.List(AuditFilter{Category: "tenancy"})
	if len(logs) != 1 {
		t.Fatalf("List() len = %d, want 1", len(logs))
	}

	stats := logger.Statistics()
	if stats.TotalLogs != 1 {
		t.Fatalf("Statistics().TotalLogs = %d, want 1", stats.TotalLogs)
	}
	if stats.ByCategory["tenancy"] != 1 {
		t.Fatalf("Statistics().ByCategory[tenancy] = %d, want 1", stats.ByCategory["tenancy"])
	}
}
