package storage

import (
	"os"
	"testing"
	"time"
)

func TestSchema_InitAndMigrate(t *testing.T) {
	dbPath := "/tmp/klaw_schema_test.db"
	_ = os.Remove(dbPath)
	defer os.Remove(dbPath)

	store, err := NewStore(dbPath)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	defer store.Close()

	// Seed legacy document data.
	rules := []AlertRuleRow{
		{ID: "rule-1", Cluster: "c1", Name: "High CPU", Enabled: true, Severity: "critical", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}
	if err := store.PutJSON("alerting", "rules", rules); err != nil {
		t.Fatalf("seed rules: %v", err)
	}

	history := []AlertRecordRow{
		{ID: "rec-1", Cluster: "c1", RuleID: "rule-1", Severity: "critical", CreatedAt: time.Now()},
	}
	if err := store.PutJSON("alerting", "history", history); err != nil {
		t.Fatalf("seed history: %v", err)
	}

	tenants := []TenantRow{
		{ID: "t1", Name: "Default", Namespaces: []string{"default"}, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}
	if err := store.PutJSON("tenancy", "tenants", tenants); err != nil {
		t.Fatalf("seed tenants: %v", err)
	}

	users := []TenantUserRow{
		{ID: "u1", TenantID: "t1", Username: "admin", Role: "admin", CreatedAt: time.Now()},
	}
	if err := store.PutJSON("tenancy", "users", users); err != nil {
		t.Fatalf("seed users: %v", err)
	}

	logs := []AuditEventRow{
		{ID: "a1", Timestamp: time.Now(), EventType: "login", User: "admin", Result: "success"},
	}
	if err := store.PutJSON("audit", "logs", logs); err != nil {
		t.Fatalf("seed audit: %v", err)
	}

	backups := map[string][]BackupRow{
		"c1": {{ID: "b1", Name: "daily", Phase: "Completed", CreatedAt: time.Now()}},
	}
	if err := store.PutJSON("backup", "items", backups); err != nil {
		t.Fatalf("seed backups: %v", err)
	}

	// Run migration.
	if err := store.MigrateFromDocuments(); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	// Verify tables have data.
	var count int
	if err := store.db.QueryRow("SELECT COUNT(*) FROM alert_rules").Scan(&count); err != nil {
		t.Fatalf("count alert_rules: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 alert rule, got %d", count)
	}

	if err := store.db.QueryRow("SELECT COUNT(*) FROM alert_history").Scan(&count); err != nil {
		t.Fatalf("count alert_history: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 alert record, got %d", count)
	}

	if err := store.db.QueryRow("SELECT COUNT(*) FROM tenants").Scan(&count); err != nil {
		t.Fatalf("count tenants: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 tenant, got %d", count)
	}

	if err := store.db.QueryRow("SELECT COUNT(*) FROM tenant_users").Scan(&count); err != nil {
		t.Fatalf("count tenant_users: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 tenant user, got %d", count)
	}

	if err := store.db.QueryRow("SELECT COUNT(*) FROM audit_logs").Scan(&count); err != nil {
		t.Fatalf("count audit_logs: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 audit log, got %d", count)
	}

	if err := store.db.QueryRow("SELECT COUNT(*) FROM backups").Scan(&count); err != nil {
		t.Fatalf("count backups: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 backup, got %d", count)
	}
}
