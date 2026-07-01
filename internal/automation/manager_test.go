package automation

import (
	"path/filepath"
	"testing"

	"github.com/kudig-io/klaw/internal/storage"
)

func newTestManager(t *testing.T) *Manager {
	t.Helper()
	store, err := storage.NewStore(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	return NewManager(store)
}

func TestDefaultScripts(t *testing.T) {
	m := newTestManager(t)
	scripts := m.List(ScriptFilter{})
	if len(scripts) == 0 {
		t.Fatal("expected default scripts, got 0")
	}
	found := false
	for _, s := range scripts {
		if s.ID == "cleanup-evicted-pods" {
			found = true
			if !s.Enabled {
				t.Error("cleanup-evicted-pods should be enabled by default")
			}
		}
	}
	if !found {
		t.Error("cleanup-evicted-pods not in defaults")
	}
}

func TestScriptCRUD(t *testing.T) {
	m := newTestManager(t)

	created, err := m.Add(Script{
		Name:        "test-script",
		Description: "a test",
		Type:        ScriptTypeCustom,
		Script:      "echo hello",
	})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	if created.ID == "" {
		t.Fatal("Add should assign ID")
	}

	got, err := m.Get(created.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Name != "test-script" {
		t.Errorf("Name = %q, want test-script", got.Name)
	}

	updated, err := m.Update(created.ID, Script{Name: "renamed", Enabled: true})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.Name != "renamed" {
		t.Errorf("Update Name = %q, want renamed", updated.Name)
	}

	if err := m.Delete(created.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := m.Get(created.ID); err == nil {
		t.Fatal("Get after Delete should fail")
	}
}

func TestScriptFilter(t *testing.T) {
	m := newTestManager(t)
	enabled := true
	result := m.List(ScriptFilter{Enabled: &enabled})
	for _, s := range result {
		if !s.Enabled {
			t.Error("filter Enabled=true returned disabled script")
		}
	}
}

func TestStatistics(t *testing.T) {
	m := newTestManager(t)
	stats := m.Statistics()
	if stats.Total != 0 {
		t.Errorf("initial Statistics.Total = %d, want 0", stats.Total)
	}
}
