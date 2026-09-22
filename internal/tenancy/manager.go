package tenancy

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	k8smanager "github.com/kudig-io/klaw/internal/kubernetes"
	"github.com/kudig-io/klaw/internal/storage"
)

// manager.go 保留租户类型定义、Manager 结构与构造、租户 CRUD、存储加载，
// 以及包内共享的字符串/默认值辅助；用户管理见 users.go，集群资源供给见 policies.go。

type ResourceQuota struct {
	CPU                    string `json:"cpu"`
	Memory                 string `json:"memory"`
	Pods                   string `json:"pods"`
	Services               string `json:"services"`
	PersistentVolumeClaims string `json:"persistentVolumeClaims"`
}

type NetworkPolicy struct {
	Enabled     bool `json:"enabled"`
	DefaultDeny bool `json:"defaultDeny"`
}

type RBACPolicy struct {
	Enabled     bool   `json:"enabled"`
	DefaultRole string `json:"defaultRole"`
}

type Tenant struct {
	ID              string        `json:"id"`
	Cluster         string        `json:"cluster,omitempty"`
	Name            string        `json:"name"`
	Description     string        `json:"description,omitempty"`
	Namespaces      []string      `json:"namespaces"`
	ResourceQuotas  ResourceQuota `json:"resourceQuotas"`
	NetworkPolicies NetworkPolicy `json:"networkPolicies"`
	RBAC            RBACPolicy    `json:"rbac"`
	CreatedAt       time.Time     `json:"createdAt"`
	UpdatedAt       time.Time     `json:"updatedAt"`
}

type TenantUser struct {
	ID               string    `json:"id"`
	TenantID         string    `json:"tenantId"`
	Username         string    `json:"username"`
	Email            string    `json:"email,omitempty"`
	Role             string    `json:"role"`
	Namespaces       []string  `json:"namespaces,omitempty"`
	SubjectKind      string    `json:"subjectKind,omitempty"`
	SubjectName      string    `json:"subjectName,omitempty"`
	SubjectNamespace string    `json:"subjectNamespace,omitempty"`
	CreatedAt        time.Time `json:"createdAt"`
}

type Statistics struct {
	TotalTenants    int            `json:"totalTenants"`
	TotalUsers      int            `json:"totalUsers"`
	TotalNamespaces int            `json:"totalNamespaces"`
	UsersByRole     map[string]int `json:"usersByRole"`
}

type Manager struct {
	k8sManager *k8smanager.Manager
	store      *storage.Store
	tenants    []Tenant
	users      []TenantUser
	mu         sync.RWMutex
}

func NewManager(k8sManager *k8smanager.Manager, store *storage.Store) *Manager {
	m := &Manager{k8sManager: k8sManager, store: store}
	m.load()
	return m
}

func (m *Manager) ListTenants(cluster, name, namespace string) []Tenant {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result = make([]Tenant, 0, len(m.tenants))
	for _, tenant := range m.tenants {
		if cluster != "" && tenant.Cluster != cluster {
			continue
		}
		if name != "" && !strings.Contains(strings.ToLower(tenant.Name), strings.ToLower(name)) {
			continue
		}
		if namespace != "" && !contains(tenant.Namespaces, namespace) {
			continue
		}
		result = append(result, tenant)
	}
	return result
}

func (m *Manager) GetTenant(id string) (*Tenant, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, tenant := range m.tenants {
		if tenant.ID == id {
			copy := tenant
			return &copy, nil
		}
	}
	return nil, fmt.Errorf("tenant not found: %s", id)
}

func (m *Manager) CreateTenant(tenant Tenant) (*Tenant, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if strings.TrimSpace(tenant.Name) == "" {
		return nil, fmt.Errorf("tenant name is required")
	}

	now := time.Now()
	tenant.ID = fmt.Sprintf("tenant-%d", now.UnixNano())
	tenant.CreatedAt = now
	tenant.UpdatedAt = now

	if len(tenant.Namespaces) == 0 {
		tenant.Namespaces = []string{"default"}
	}
	if tenant.ResourceQuotas.CPU == "" {
		tenant.ResourceQuotas = defaultResourceQuota()
	}
	if tenant.RBAC.DefaultRole == "" {
		tenant.RBAC = RBACPolicy{Enabled: true, DefaultRole: "view"}
	}
	if !tenant.NetworkPolicies.Enabled {
		tenant.NetworkPolicies.Enabled = true
	}
	tenant = normalizeTenant(tenant)

	m.tenants = append(m.tenants, tenant)
	if err := m.applyTenantLocked(tenant); err != nil {
		m.tenants = m.tenants[:len(m.tenants)-1]
		return nil, err
	}
	if err := m.saveLocked(); err != nil {
		return nil, err
	}
	return &tenant, nil
}

func (m *Manager) UpdateTenant(id string, updates Tenant) (*Tenant, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, tenant := range m.tenants {
		if tenant.ID != id {
			continue
		}
		oldTenant := tenant
		updates.ID = id
		updates.CreatedAt = tenant.CreatedAt
		updates.UpdatedAt = time.Now()
		if len(updates.Namespaces) == 0 {
			updates.Namespaces = tenant.Namespaces
		}
		if updates.Cluster == "" {
			updates.Cluster = tenant.Cluster
		}
		if updates.ResourceQuotas.CPU == "" {
			updates.ResourceQuotas = tenant.ResourceQuotas
		}
		if updates.RBAC.DefaultRole == "" {
			updates.RBAC = tenant.RBAC
		}
		updates.NetworkPolicies = mergeNetworkPolicy(tenant.NetworkPolicies, updates.NetworkPolicies)
		updates = normalizeTenant(updates)
		m.tenants[i] = updates
		if err := m.reconcileTenantLocked(oldTenant, updates); err != nil {
			m.tenants[i] = oldTenant
			return nil, err
		}
		if err := m.saveLocked(); err != nil {
			return nil, err
		}
		copy := updates
		return &copy, nil
	}
	return nil, fmt.Errorf("tenant not found: %s", id)
}

func (m *Manager) DeleteTenant(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, tenant := range m.tenants {
		if tenant.ID != id {
			continue
		}
		if err := m.cleanupTenantLocked(tenant); err != nil {
			return err
		}
		m.tenants = append(m.tenants[:i], m.tenants[i+1:]...)

		filteredUsers := m.users[:0]
		for _, user := range m.users {
			if user.TenantID != id {
				filteredUsers = append(filteredUsers, user)
			}
		}
		m.users = filteredUsers
		return m.saveLocked()
	}
	return fmt.Errorf("tenant not found: %s", id)
}

func (m *Manager) Statistics() Statistics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := Statistics{
		TotalTenants: len(m.tenants),
		TotalUsers:   len(m.users),
		UsersByRole:  map[string]int{},
	}
	for _, tenant := range m.tenants {
		stats.TotalNamespaces += len(tenant.Namespaces)
	}
	for _, user := range m.users {
		stats.UsersByRole[user.Role]++
	}
	return stats
}

func (m *Manager) tenantExistsLocked(id string) bool {
	for _, tenant := range m.tenants {
		if tenant.ID == id {
			return true
		}
	}
	return false
}

func (m *Manager) getTenantLocked(id string) (*Tenant, error) {
	for _, tenant := range m.tenants {
		if tenant.ID == id {
			copy := tenant
			return &copy, nil
		}
	}
	return nil, fmt.Errorf("tenant not found: %s", id)
}

func (m *Manager) load() {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, _ = m.store.GetJSON("tenancy", "tenants", &m.tenants)
	if len(m.tenants) == 0 {
		now := time.Now()
		m.tenants = []Tenant{{
			ID:             "default",
			Cluster:        "",
			Name:           "Default",
			Description:    "默认租户",
			Namespaces:     []string{"default", "kube-system", "kube-public"},
			ResourceQuotas: defaultResourceQuota(),
			NetworkPolicies: NetworkPolicy{
				Enabled:     true,
				DefaultDeny: false,
			},
			RBAC:      RBACPolicy{Enabled: true, DefaultRole: "view"},
			CreatedAt: now,
			UpdatedAt: now,
		}}
		_ = m.saveLocked()
	}
	_, _ = m.store.GetJSON("tenancy", "users", &m.users)
}

func (m *Manager) saveLocked() error {
	if err := m.store.PutJSON("tenancy", "tenants", m.tenants); err != nil {
		return err
	}
	return m.store.PutJSON("tenancy", "users", m.users)
}

func normalizeTenant(tenant Tenant) Tenant {
	tenant.Namespaces = uniqueStrings(tenant.Namespaces)
	return tenant
}

func mergeNetworkPolicy(current, updates NetworkPolicy) NetworkPolicy {
	if !updates.Enabled && !updates.DefaultDeny {
		return current
	}
	return updates
}

func uniqueStrings(items []string) []string {
	seen := make(map[string]struct{}, len(items))
	var result []string
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		result = append(result, item)
	}
	sort.Strings(result)
	return result
}

func defaultStrings(items []string, fallback []string) []string {
	if len(items) == 0 {
		return append([]string(nil), fallback...)
	}
	return items
}

func firstNonEmpty(items ...string) string {
	for _, item := range items {
		if strings.TrimSpace(item) != "" {
			return item
		}
	}
	return ""
}

func sameStrings(left, right []string) bool {
	a := uniqueStrings(left)
	b := uniqueStrings(right)
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func defaultResourceQuota() ResourceQuota {
	return ResourceQuota{
		CPU:                    "10",
		Memory:                 "20Gi",
		Pods:                   "100",
		Services:               "50",
		PersistentVolumeClaims: "20",
	}
}

func contains(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}
