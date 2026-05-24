package alerting

import (
	"path/filepath"
	"testing"

	"github.com/kudig-io/klaw/internal/storage"
)

func TestManagerRuleLifecycle(t *testing.T) {
	store, err := storage.NewStore(filepath.Join(t.TempDir(), "klaw.db"))
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}
	manager := NewManager(nil, store)

	rule, err := manager.AddRule(Rule{
		Cluster:   "demo",
		Name:      "Test Rule",
		Enabled:   true,
		Severity:  "warning",
		Condition: Condition{Type: "event", Field: "type", Operator: "==", Threshold: "Warning"},
	})
	if err != nil {
		t.Fatalf("AddRule() error = %v", err)
	}

	if len(manager.GetRules("demo")) == 0 {
		t.Fatal("GetRules() should return the created rule")
	}

	updated, err := manager.UpdateRule(rule.ID, Rule{
		Cluster:     "demo",
		Name:        "Updated Rule",
		Description: "updated",
		Enabled:     true,
		Severity:    "critical",
		Condition:   Condition{Type: "event", Field: "reason", Operator: "contains", Threshold: "BackOff"},
	})
	if err != nil {
		t.Fatalf("UpdateRule() error = %v", err)
	}

	if updated.Name != "Updated Rule" {
		t.Fatalf("updated.Name = %q, want Updated Rule", updated.Name)
	}

	if err := manager.DeleteRule(rule.ID); err != nil {
		t.Fatalf("DeleteRule() error = %v", err)
	}
}

func TestManagerAlertLifecycle(t *testing.T) {
	store, err := storage.NewStore(filepath.Join(t.TempDir(), "klaw.db"))
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}
	manager := NewManager(nil, store)

	record, created, err := manager.storeRecord(Record{
		ID:           "record-1",
		Cluster:      "demo",
		RuleID:       "rule-1",
		RuleName:     "Rule 1",
		ResourceKind: "Pod",
		ResourceName: "pod-1",
		Severity:     "warning",
		Value:        "Warning",
	})
	if err != nil {
		t.Fatalf("storeRecord() error = %v", err)
	}
	if !created {
		t.Fatal("storeRecord() should create a record")
	}

	ack, err := manager.Acknowledge(record.ID)
	if err != nil {
		t.Fatalf("Acknowledge() error = %v", err)
	}
	if !ack.Acknowledged {
		t.Fatal("Acknowledge() should mark the record acknowledged")
	}

	resolved, err := manager.Resolve(record.ID)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if !resolved.Resolved {
		t.Fatal("Resolve() should mark the record resolved")
	}
}
