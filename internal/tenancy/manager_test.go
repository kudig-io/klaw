package tenancy

import (
	"path/filepath"
	"testing"

	"github.com/kudig-io/klaw/internal/storage"
)

func TestTenancyManagerLifecycle(t *testing.T) {
	store, err := storage.NewStore(filepath.Join(t.TempDir(), "klaw.db"))
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}
	manager := NewManager(nil, store)

	tenant, err := manager.CreateTenant(Tenant{
		Cluster:     "demo",
		Name:        "Team A",
		Description: "demo tenant",
		Namespaces:  []string{"team-a"},
		ResourceQuotas: ResourceQuota{
			CPU:                    "2",
			Memory:                 "4Gi",
			Pods:                   "20",
			Services:               "10",
			PersistentVolumeClaims: "5",
		},
		NetworkPolicies: NetworkPolicy{Enabled: true, DefaultDeny: true},
		RBAC:            RBACPolicy{Enabled: true, DefaultRole: "edit"},
	})
	if err != nil {
		t.Fatalf("CreateTenant() error = %v", err)
	}

	if len(manager.ListTenants("", "", "")) < 2 {
		t.Fatal("ListTenants() should include default tenant and created tenant")
	}
	if len(manager.ListTenants("demo", "", "")) != 1 {
		t.Fatal("ListTenants() should filter by cluster")
	}

	user, err := manager.AddUser(TenantUser{
		TenantID:     tenant.ID,
		Username:     "alice",
		Role:         "admin",
		SubjectKind:  "service-account",
		SubjectName:  "tenant-operator",
		Namespaces:   []string{"team-a"},
	})
	if err != nil {
		t.Fatalf("AddUser() error = %v", err)
	}
	if user.SubjectKind != "ServiceAccount" {
		t.Fatalf("expected subject kind to normalize, got %q", user.SubjectKind)
	}
	if user.SubjectNamespace != "team-a" {
		t.Fatalf("expected service account namespace to default from user namespaces, got %q", user.SubjectNamespace)
	}
	if len(user.Namespaces) != 1 || user.Namespaces[0] != "team-a" {
		t.Fatalf("expected user namespaces to be preserved, got %+v", user.Namespaces)
	}

	if len(manager.ListUsers(tenant.ID, "")) != 1 {
		t.Fatal("ListUsers() should return created user")
	}

	stats := manager.Statistics()
	if stats.TotalTenants < 2 || stats.TotalUsers != 1 {
		t.Fatalf("unexpected stats: %+v", stats)
	}

	if err := manager.DeleteUser(user.ID); err != nil {
		t.Fatalf("DeleteUser() error = %v", err)
	}
	if err := manager.DeleteTenant(tenant.ID); err != nil {
		t.Fatalf("DeleteTenant() error = %v", err)
	}
}
