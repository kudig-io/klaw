package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// InitSchema creates domain-specific tables alongside the legacy document store.
func (s *Store) InitSchema() error {
	tables := []string{
		`CREATE TABLE IF NOT EXISTS alert_rules (
			id TEXT PRIMARY KEY,
			cluster TEXT,
			name TEXT NOT NULL,
			description TEXT,
			enabled INTEGER NOT NULL DEFAULT 1,
			severity TEXT NOT NULL,
			condition_type TEXT,
			condition_field TEXT,
			condition_operator TEXT,
			condition_threshold TEXT,
			condition_time_window TEXT,
			actions TEXT,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS alert_history (
			id TEXT PRIMARY KEY,
			cluster TEXT,
			rule_id TEXT,
			rule_name TEXT,
			rule_type TEXT,
			resource_kind TEXT,
			resource_name TEXT,
			namespace TEXT,
			severity TEXT,
			value TEXT,
			threshold TEXT,
			operator TEXT,
			message TEXT,
			acknowledged INTEGER DEFAULT 0,
			resolved INTEGER DEFAULT 0,
			created_at DATETIME NOT NULL,
			acknowledged_at DATETIME,
			resolved_at DATETIME
		)`,
		`CREATE INDEX IF NOT EXISTS idx_alert_history_cluster ON alert_history(cluster)`,
		`CREATE INDEX IF NOT EXISTS idx_alert_history_created ON alert_history(created_at)`,
		`CREATE TABLE IF NOT EXISTS tenants (
			id TEXT PRIMARY KEY,
			cluster TEXT,
			name TEXT NOT NULL,
			description TEXT,
			namespaces TEXT,
			resource_quotas TEXT,
			network_policies TEXT,
			rbac TEXT,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS tenant_users (
			id TEXT PRIMARY KEY,
			tenant_id TEXT NOT NULL,
			username TEXT NOT NULL,
			email TEXT,
			role TEXT,
			namespaces TEXT,
			subject_kind TEXT,
			subject_name TEXT,
			subject_namespace TEXT,
			created_at DATETIME NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_tenant_users_tenant ON tenant_users(tenant_id)`,
		`CREATE TABLE IF NOT EXISTS audit_logs (
			id TEXT PRIMARY KEY,
			timestamp DATETIME NOT NULL,
			event_type TEXT,
			category TEXT,
			severity TEXT,
			source TEXT,
			user TEXT,
			action TEXT,
			resource_kind TEXT,
			resource_name TEXT,
			resource_namespace TEXT,
			result TEXT,
			details TEXT,
			ip_address TEXT,
			user_agent TEXT
		)`,
		`CREATE INDEX IF NOT EXISTS idx_audit_timestamp ON audit_logs(timestamp)`,
		`CREATE INDEX IF NOT EXISTS idx_audit_user ON audit_logs(user)`,
		`CREATE TABLE IF NOT EXISTS backups (
			id TEXT PRIMARY KEY,
			cluster TEXT,
			name TEXT NOT NULL,
			phase TEXT,
			backup_mode TEXT,
			etcd_endpoints TEXT,
			storage_provider TEXT,
			storage_bucket TEXT,
			storage_prefix TEXT,
			storage_region TEXT,
			storage_endpoint TEXT,
			storage_credentials_secret TEXT,
			validation_enabled INTEGER,
			consistency_check INTEGER,
			retention_max_backups INTEGER,
			retention_max_age TEXT,
			velero_integration INTEGER,
			snapshot_size INTEGER,
			snapshot_location TEXT,
			etcd_revision INTEGER,
			validation_valid INTEGER,
			validation_hash TEXT,
			validation_message TEXT,
			start_time DATETIME,
			completion_time DATETIME,
			message TEXT,
			created_at DATETIME NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_backups_cluster ON backups(cluster)`,
	}
	for _, ddl := range tables {
		if _, err := s.db.Exec(ddl); err != nil {
			return fmt.Errorf("schema init: %w", err)
		}
	}
	return nil
}

// MigrateFromDocuments reads legacy JSON documents and migrates them into domain tables.
func (s *Store) MigrateFromDocuments() error {
	if err := s.migrateAlertRules(); err != nil {
		return fmt.Errorf("migrate alert rules: %w", err)
	}
	if err := s.migrateAlertHistory(); err != nil {
		return fmt.Errorf("migrate alert history: %w", err)
	}
	if err := s.migrateTenants(); err != nil {
		return fmt.Errorf("migrate tenants: %w", err)
	}
	if err := s.migrateTenantUsers(); err != nil {
		return fmt.Errorf("migrate tenant users: %w", err)
	}
	if err := s.migrateAuditLogs(); err != nil {
		return fmt.Errorf("migrate audit logs: %w", err)
	}
	if err := s.migrateBackups(); err != nil {
		return fmt.Errorf("migrate backups: %w", err)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Alert Rules
// ---------------------------------------------------------------------------

type AlertRuleRow struct {
	ID          string    `json:"id"`
	Cluster     string    `json:"cluster"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Enabled     bool      `json:"enabled"`
	Severity    string    `json:"severity"`
	Condition   struct {
		Type       string      `json:"type"`
		Field      string      `json:"field"`
		Operator   string      `json:"operator"`
		Threshold  interface{} `json:"threshold"`
		TimeWindow string      `json:"timeWindow"`
	} `json:"condition"`
	Actions   []string  `json:"actions"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (s *Store) migrateAlertRules() error {
	var rules []AlertRuleRow
	if _, err := s.GetJSON("alerting", "rules", &rules); err != nil || len(rules) == 0 {
		return nil
	}
	for _, r := range rules {
		threshold, _ := json.Marshal(r.Condition.Threshold)
		actions, _ := json.Marshal(r.Actions)
		_, err := s.db.Exec(`
			INSERT OR REPLACE INTO alert_rules
			(id, cluster, name, description, enabled, severity,
			 condition_type, condition_field, condition_operator, condition_threshold, condition_time_window,
			 actions, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			r.ID, r.Cluster, r.Name, r.Description, boolInt(r.Enabled), r.Severity,
			r.Condition.Type, r.Condition.Field, r.Condition.Operator, string(threshold), r.Condition.TimeWindow,
			string(actions), r.CreatedAt, r.UpdatedAt)
		if err != nil {
			return err
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// Alert History
// ---------------------------------------------------------------------------

type AlertRecordRow struct {
	ID             string     `json:"id"`
	Cluster        string     `json:"cluster"`
	RuleID         string     `json:"ruleId"`
	RuleName       string     `json:"ruleName"`
	RuleType       string     `json:"ruleType"`
	ResourceKind   string     `json:"resourceKind"`
	ResourceName   string     `json:"resourceName"`
	Namespace      string     `json:"namespace"`
	Severity       string     `json:"severity"`
	Value          interface{} `json:"value"`
	Threshold      interface{} `json:"threshold"`
	Operator       string     `json:"operator"`
	Message        string     `json:"message"`
	Acknowledged   bool       `json:"acknowledged"`
	Resolved       bool       `json:"resolved"`
	CreatedAt      time.Time  `json:"createdAt"`
	AcknowledgedAt *time.Time `json:"acknowledgedAt"`
	ResolvedAt     *time.Time `json:"resolvedAt"`
}

func (s *Store) migrateAlertHistory() error {
	var records []AlertRecordRow
	if _, err := s.GetJSON("alerting", "history", &records); err != nil || len(records) == 0 {
		return nil
	}
	for _, r := range records {
		val, _ := json.Marshal(r.Value)
		thr, _ := json.Marshal(r.Threshold)
		var ackAt, resAt interface{}
		if r.AcknowledgedAt != nil {
			ackAt = *r.AcknowledgedAt
		}
		if r.ResolvedAt != nil {
			resAt = *r.ResolvedAt
		}
		_, err := s.db.Exec(`
			INSERT OR REPLACE INTO alert_history
			(id, cluster, rule_id, rule_name, rule_type, resource_kind, resource_name, namespace,
			 severity, value, threshold, operator, message, acknowledged, resolved,
			 created_at, acknowledged_at, resolved_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			r.ID, r.Cluster, r.RuleID, r.RuleName, r.RuleType, r.ResourceKind, r.ResourceName, r.Namespace,
			r.Severity, string(val), string(thr), r.Operator, r.Message, boolInt(r.Acknowledged), boolInt(r.Resolved),
			r.CreatedAt, ackAt, resAt)
		if err != nil {
			return err
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// Tenants
// ---------------------------------------------------------------------------

type TenantRow struct {
	ID              string                 `json:"id"`
	Cluster         string                 `json:"cluster"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	Namespaces      []string               `json:"namespaces"`
	ResourceQuotas  map[string]interface{} `json:"resourceQuotas"`
	NetworkPolicies map[string]interface{} `json:"networkPolicies"`
	RBAC            map[string]interface{} `json:"rbac"`
	CreatedAt       time.Time              `json:"createdAt"`
	UpdatedAt       time.Time              `json:"updatedAt"`
}

func (s *Store) migrateTenants() error {
	var tenants []TenantRow
	if _, err := s.GetJSON("tenancy", "tenants", &tenants); err != nil || len(tenants) == 0 {
		return nil
	}
	for _, t := range tenants {
		ns, _ := json.Marshal(t.Namespaces)
		rq, _ := json.Marshal(t.ResourceQuotas)
		np, _ := json.Marshal(t.NetworkPolicies)
		rbac, _ := json.Marshal(t.RBAC)
		_, err := s.db.Exec(`
			INSERT OR REPLACE INTO tenants
			(id, cluster, name, description, namespaces, resource_quotas, network_policies, rbac, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			t.ID, t.Cluster, t.Name, t.Description, string(ns), string(rq), string(np), string(rbac), t.CreatedAt, t.UpdatedAt)
		if err != nil {
			return err
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// Tenant Users
// ---------------------------------------------------------------------------

type TenantUserRow struct {
	ID               string    `json:"id"`
	TenantID         string    `json:"tenantId"`
	Username         string    `json:"username"`
	Email            string    `json:"email"`
	Role             string    `json:"role"`
	Namespaces       []string  `json:"namespaces"`
	SubjectKind      string    `json:"subjectKind"`
	SubjectName      string    `json:"subjectName"`
	SubjectNamespace string    `json:"subjectNamespace"`
	CreatedAt        time.Time `json:"createdAt"`
}

func (s *Store) migrateTenantUsers() error {
	var users []TenantUserRow
	if _, err := s.GetJSON("tenancy", "users", &users); err != nil || len(users) == 0 {
		return nil
	}
	for _, u := range users {
		ns, _ := json.Marshal(u.Namespaces)
		_, err := s.db.Exec(`
			INSERT OR REPLACE INTO tenant_users
			(id, tenant_id, username, email, role, namespaces, subject_kind, subject_name, subject_namespace, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			u.ID, u.TenantID, u.Username, u.Email, u.Role, string(ns), u.SubjectKind, u.SubjectName, u.SubjectNamespace, u.CreatedAt)
		if err != nil {
			return err
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// Audit Logs
// ---------------------------------------------------------------------------

type AuditEventRow struct {
	ID        string                 `json:"id"`
	Timestamp time.Time              `json:"timestamp"`
	EventType string                 `json:"eventType"`
	Category  string                 `json:"category"`
	Severity  string                 `json:"severity"`
	Source    string                 `json:"source"`
	User      string                 `json:"user"`
	Action    string                 `json:"action"`
	Resource  map[string]string      `json:"resource"`
	Result    string                 `json:"result"`
	Details   map[string]interface{} `json:"details"`
	IPAddress string                 `json:"ipAddress"`
	UserAgent string                 `json:"userAgent"`
}

func (s *Store) migrateAuditLogs() error {
	var logs []AuditEventRow
	if _, err := s.GetJSON("audit", "logs", &logs); err != nil || len(logs) == 0 {
		return nil
	}
	for _, l := range logs {
		det, _ := json.Marshal(l.Details)
		_, err := s.db.Exec(`
			INSERT OR REPLACE INTO audit_logs
			(id, timestamp, event_type, category, severity, source, user, action,
			 resource_kind, resource_name, resource_namespace, result, details, ip_address, user_agent)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			l.ID, l.Timestamp, l.EventType, l.Category, l.Severity, l.Source, l.User, l.Action,
			l.Resource["kind"], l.Resource["name"], l.Resource["namespace"], l.Result, string(det), l.IPAddress, l.UserAgent)
		if err != nil {
			return err
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// Backups
// ---------------------------------------------------------------------------

type BackupRow struct {
	ID               string                 `json:"id"`
	Cluster          string                 `json:"cluster"`
	Name             string                 `json:"name"`
	Phase            string                 `json:"phase"`
	Spec             map[string]interface{} `json:"spec"`
	SnapshotSize     int64                  `json:"snapshotSize"`
	SnapshotLocation string                 `json:"snapshotLocation"`
	EtcdRevision     int64                  `json:"etcdRevision"`
	ValidationResult map[string]interface{} `json:"validationResult"`
	StartTime        time.Time              `json:"startTime"`
	CompletionTime   *time.Time             `json:"completionTime"`
	Message          string                 `json:"message"`
	CreatedAt        time.Time              `json:"createdAt"`
}

func (s *Store) migrateBackups() error {
	var byCluster map[string][]BackupRow
	if _, err := s.GetJSON("backup", "items", &byCluster); err != nil || len(byCluster) == 0 {
		return nil
	}
	for cluster, backups := range byCluster {
		for _, b := range backups {
			var vrValid sql.NullInt64
			var vrHash, vrMsg sql.NullString
			if vr, ok := b.ValidationResult["valid"].(bool); ok {
				vrValid = sql.NullInt64{Int64: int64(boolInt(vr)), Valid: true}
			}
			if h, ok := b.ValidationResult["hash"].(string); ok {
				vrHash = sql.NullString{String: h, Valid: true}
			}
			if m, ok := b.ValidationResult["message"].(string); ok {
				vrMsg = sql.NullString{String: m, Valid: true}
			}
			var compAt interface{}
			if b.CompletionTime != nil {
				compAt = *b.CompletionTime
			}
			_, err := s.db.Exec(`
				INSERT OR REPLACE INTO backups
				(id, cluster, name, phase, backup_mode, etcd_endpoints,
				 storage_provider, storage_bucket, storage_prefix, storage_region, storage_endpoint, storage_credentials_secret,
				 validation_enabled, consistency_check, retention_max_backups, retention_max_age, velero_integration,
				 snapshot_size, snapshot_location, etcd_revision,
				 validation_valid, validation_hash, validation_message,
				 start_time, completion_time, message, created_at)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				b.ID, cluster, b.Name, b.Phase,
				getStr(b.Spec, "backupMode"), getStrSlice(b.Spec, "etcdEndpoints"),
				getStr(getMap(b.Spec, "storageLocation"), "provider"),
				getStr(getMap(b.Spec, "storageLocation"), "bucket"),
				getStr(getMap(b.Spec, "storageLocation"), "prefix"),
				getStr(getMap(b.Spec, "storageLocation"), "region"),
				getStr(getMap(b.Spec, "storageLocation"), "endpoint"),
				getStr(getMap(b.Spec, "storageLocation"), "credentialsSecret"),
				getBool(getMap(b.Spec, "validation"), "enabled"),
				getBool(getMap(b.Spec, "validation"), "consistencyCheck"),
				getInt(getMap(b.Spec, "retentionPolicy"), "maxBackups"),
				getStr(getMap(b.Spec, "retentionPolicy"), "maxAge"),
				getBool(b.Spec, "veleroIntegration"),
				b.SnapshotSize, b.SnapshotLocation, b.EtcdRevision,
				vrValid, vrHash, vrMsg,
				b.StartTime, compAt, b.Message, b.CreatedAt)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func getStr(m map[string]interface{}, key string) string {
	if m == nil {
		return ""
	}
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getStrSlice(m map[string]interface{}, key string) string {
	if m == nil {
		return ""
	}
	if v, ok := m[key].([]interface{}); ok {
		b, _ := json.Marshal(v)
		return string(b)
	}
	return ""
}

func getInt(m map[string]interface{}, key string) int {
	if m == nil {
		return 0
	}
	switch v := m[key].(type) {
	case int:
		return v
	case float64:
		return int(v)
	case int64:
		return int(v)
	}
	return 0
}

func getBool(m map[string]interface{}, key string) int {
	if m == nil {
		return 0
	}
	if v, ok := m[key].(bool); ok {
		return boolInt(v)
	}
	return 0
}

func getMap(m map[string]interface{}, key string) map[string]interface{} {
	if m == nil {
		return nil
	}
	if v, ok := m[key].(map[string]interface{}); ok {
		return v
	}
	return nil
}
