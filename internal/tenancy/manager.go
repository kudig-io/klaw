package tenancy

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	k8smanager "github.com/kudig-io/klaw/internal/kubernetes"
	"github.com/kudig-io/klaw/internal/storage"
)

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
	ID             string         `json:"id"`
	Cluster        string         `json:"cluster,omitempty"`
	Name           string         `json:"name"`
	Description    string         `json:"description,omitempty"`
	Namespaces     []string       `json:"namespaces"`
	ResourceQuotas ResourceQuota  `json:"resourceQuotas"`
	NetworkPolicies NetworkPolicy `json:"networkPolicies"`
	RBAC           RBACPolicy     `json:"rbac"`
	CreatedAt      time.Time      `json:"createdAt"`
	UpdatedAt      time.Time      `json:"updatedAt"`
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
	tenants []Tenant
	users   []TenantUser
	mu      sync.RWMutex
}

func NewManager(k8sManager *k8smanager.Manager, store *storage.Store) *Manager {
	m := &Manager{k8sManager: k8sManager, store: store}
	m.load()
	return m
}

func (m *Manager) ListTenants(cluster, name, namespace string) []Tenant {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []Tenant
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

func (m *Manager) ListUsers(tenantID, role string) []TenantUser {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []TenantUser
	for _, user := range m.users {
		if tenantID != "" && user.TenantID != tenantID {
			continue
		}
		if role != "" && user.Role != role {
			continue
		}
		result = append(result, user)
	}
	return result
}

func (m *Manager) AddUser(user TenantUser) (*TenantUser, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if user.TenantID == "" || strings.TrimSpace(user.Username) == "" {
		return nil, fmt.Errorf("tenantId and username are required")
	}
	tenant, err := m.getTenantLocked(user.TenantID)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	user.ID = fmt.Sprintf("user-%d", now.UnixNano())
	user.CreatedAt = now
	if user.Role == "" {
		user.Role = "viewer"
	}
	user = normalizeTenantUser(user, *tenant)
	m.users = append(m.users, user)
	if err := m.applyTenantUserLocked(*tenant, user); err != nil {
		m.users = m.users[:len(m.users)-1]
		return nil, err
	}
	if err := m.saveLocked(); err != nil {
		return nil, err
	}
	return &user, nil
}

func (m *Manager) DeleteUser(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, user := range m.users {
		if user.ID != id {
			continue
		}
		tenant, err := m.getTenantLocked(user.TenantID)
		if err == nil {
			if err := m.cleanupTenantUserLocked(*tenant, user); err != nil {
				return err
			}
		}
		m.users = append(m.users[:i], m.users[i+1:]...)
		return m.saveLocked()
	}
	return fmt.Errorf("user not found: %s", id)
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

func (m *Manager) applyTenantLocked(tenant Tenant) error {
	if m.k8sManager == nil || tenant.Cluster == "" {
		return nil
	}

	client, err := m.k8sManager.GetClient(tenant.Cluster)
	if err != nil {
		return err
	}

	for _, namespace := range tenant.Namespaces {
		if err := ensureNamespace(client, namespace, tenant.ID); err != nil {
			return err
		}
		if err := applyResourceQuota(client, namespace, tenant); err != nil {
			return err
		}
		if tenant.NetworkPolicies.Enabled && tenant.NetworkPolicies.DefaultDeny {
			if err := applyDefaultDenyPolicy(client, namespace, tenant); err != nil {
				return err
			}
		}
		if tenant.RBAC.Enabled {
			if err := applyTenantRBAC(client, namespace, tenant); err != nil {
				return err
			}
		}
	}

	return nil
}

func (m *Manager) cleanupTenantLocked(tenant Tenant) error {
	if m.k8sManager == nil || tenant.Cluster == "" {
		return nil
	}

	client, err := m.k8sManager.GetClient(tenant.Cluster)
	if err != nil {
		return err
	}

	for _, namespace := range tenant.Namespaces {
		if err := deleteTenantManagedResources(client, namespace, tenant.ID); err != nil {
			return err
		}
	}

	for _, user := range m.users {
		if user.TenantID != tenant.ID {
			continue
		}
		if err := m.cleanupTenantUserLocked(tenant, user); err != nil {
			return err
		}
	}

	return nil
}

func (m *Manager) reconcileTenantLocked(oldTenant, newTenant Tenant) error {
	if oldTenant.Cluster != "" && (oldTenant.Cluster != newTenant.Cluster || !sameStrings(oldTenant.Namespaces, newTenant.Namespaces)) {
		if err := m.cleanupTenantLocked(oldTenant); err != nil {
			return err
		}
	}
	if err := m.applyTenantLocked(newTenant); err != nil {
		return err
	}
	for i, user := range m.users {
		if user.TenantID != newTenant.ID {
			continue
		}
		normalized := normalizeTenantUser(user, newTenant)
		m.users[i] = normalized
		if err := m.applyTenantUserLocked(newTenant, normalized); err != nil {
			return err
		}
	}
	return nil
}

func normalizeTenant(tenant Tenant) Tenant {
	tenant.Namespaces = uniqueStrings(tenant.Namespaces)
	return tenant
}

func normalizeTenantUser(user TenantUser, tenant Tenant) TenantUser {
	if user.Role == "" {
		user.Role = "viewer"
	}
	user.Namespaces = uniqueStrings(defaultStrings(user.Namespaces, tenant.Namespaces))
	user.SubjectKind = normalizeSubjectKind(user.SubjectKind)
	if user.SubjectName == "" {
		user.SubjectName = user.Username
	}
	if user.SubjectKind == rbacv1.ServiceAccountKind {
		if user.SubjectNamespace == "" {
			user.SubjectNamespace = firstNonEmpty(user.Namespaces...)
		}
		if user.SubjectNamespace == "" {
			user.SubjectNamespace = "default"
		}
	} else {
		user.SubjectNamespace = ""
	}
	return user
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

func ensureNamespace(client *kubernetes.Clientset, namespace, tenantID string) error {
	ctx := context.Background()
	labels := managedLabels(tenantID)

	existing, err := client.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = client.CoreV1().Namespaces().Create(ctx, &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:   namespace,
				Labels: labels,
			},
		}, metav1.CreateOptions{})
		return err
	}
	if err != nil {
		return err
	}

	if existing.Labels == nil {
		existing.Labels = map[string]string{}
	}
	changed := false
	for key, value := range labels {
		if existing.Labels[key] != value {
			existing.Labels[key] = value
			changed = true
		}
	}
	if !changed {
		return nil
	}
	_, err = client.CoreV1().Namespaces().Update(ctx, existing, metav1.UpdateOptions{})
	return err
}

func applyResourceQuota(client *kubernetes.Clientset, namespace string, tenant Tenant) error {
	ctx := context.Background()
	name := managedName(tenant.ID, "quota")
	quota := &corev1.ResourceQuota{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    managedLabels(tenant.ID),
		},
		Spec: corev1.ResourceQuotaSpec{
			Hard: corev1.ResourceList{
				corev1.ResourceLimitsCPU:              resource.MustParse(tenant.ResourceQuotas.CPU),
				corev1.ResourceLimitsMemory:           resource.MustParse(tenant.ResourceQuotas.Memory),
				corev1.ResourcePods:                   resource.MustParse(tenant.ResourceQuotas.Pods),
				corev1.ResourceServices:               resource.MustParse(tenant.ResourceQuotas.Services),
				corev1.ResourcePersistentVolumeClaims: resource.MustParse(tenant.ResourceQuotas.PersistentVolumeClaims),
			},
		},
	}

	existing, err := client.CoreV1().ResourceQuotas(namespace).Get(ctx, name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = client.CoreV1().ResourceQuotas(namespace).Create(ctx, quota, metav1.CreateOptions{})
		return err
	}
	if err != nil {
		return err
	}
	existing.Labels = quota.Labels
	existing.Spec = quota.Spec
	_, err = client.CoreV1().ResourceQuotas(namespace).Update(ctx, existing, metav1.UpdateOptions{})
	return err
}

func applyDefaultDenyPolicy(client *kubernetes.Clientset, namespace string, tenant Tenant) error {
	ctx := context.Background()
	name := managedName(tenant.ID, "default-deny")
	policy := &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    managedLabels(tenant.ID),
		},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{},
			PolicyTypes: []networkingv1.PolicyType{
				networkingv1.PolicyTypeIngress,
				networkingv1.PolicyTypeEgress,
			},
		},
	}

	existing, err := client.NetworkingV1().NetworkPolicies(namespace).Get(ctx, name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = client.NetworkingV1().NetworkPolicies(namespace).Create(ctx, policy, metav1.CreateOptions{})
		return err
	}
	if err != nil {
		return err
	}
	existing.Labels = policy.Labels
	existing.Spec = policy.Spec
	_, err = client.NetworkingV1().NetworkPolicies(namespace).Update(ctx, existing, metav1.UpdateOptions{})
	return err
}

func applyTenantRBAC(client *kubernetes.Clientset, namespace string, tenant Tenant) error {
	ctx := context.Background()
	roleName := managedName(tenant.ID, "role")
	bindingName := managedName(tenant.ID, "binding")
	rules := defaultPolicyRules(tenant.RBAC.DefaultRole)

	role := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      roleName,
			Namespace: namespace,
			Labels:    managedLabels(tenant.ID),
		},
		Rules: rules,
	}
	roleExisting, err := client.RbacV1().Roles(namespace).Get(ctx, roleName, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		if _, err = client.RbacV1().Roles(namespace).Create(ctx, role, metav1.CreateOptions{}); err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else {
		roleExisting.Labels = role.Labels
		roleExisting.Rules = role.Rules
		if _, err = client.RbacV1().Roles(namespace).Update(ctx, roleExisting, metav1.UpdateOptions{}); err != nil {
			return err
		}
	}

	binding := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      bindingName,
			Namespace: namespace,
			Labels:    managedLabels(tenant.ID),
		},
		Subjects: []rbacv1.Subject{{
			Kind:      rbacv1.ServiceAccountKind,
			Name:      "default",
			Namespace: namespace,
		}},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "Role",
			Name:     roleName,
		},
	}
	bindingExisting, err := client.RbacV1().RoleBindings(namespace).Get(ctx, bindingName, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = client.RbacV1().RoleBindings(namespace).Create(ctx, binding, metav1.CreateOptions{})
		return err
	}
	if err != nil {
		return err
	}
	bindingExisting.Labels = binding.Labels
	bindingExisting.Subjects = binding.Subjects
	bindingExisting.RoleRef = binding.RoleRef
	_, err = client.RbacV1().RoleBindings(namespace).Update(ctx, bindingExisting, metav1.UpdateOptions{})
	return err
}

func (m *Manager) applyTenantUserLocked(tenant Tenant, user TenantUser) error {
	if m.k8sManager == nil || tenant.Cluster == "" || !tenant.RBAC.Enabled {
		return nil
	}

	client, err := m.k8sManager.GetClient(tenant.Cluster)
	if err != nil {
		return err
	}

	for _, namespace := range user.Namespaces {
		if !contains(tenant.Namespaces, namespace) {
			continue
		}
		if err := applyTenantUserBinding(client, namespace, tenant, user); err != nil {
			return err
		}
	}
	return nil
}

func (m *Manager) cleanupTenantUserLocked(tenant Tenant, user TenantUser) error {
	if m.k8sManager == nil || tenant.Cluster == "" {
		return nil
	}

	client, err := m.k8sManager.GetClient(tenant.Cluster)
	if err != nil {
		return err
	}

	for _, namespace := range user.Namespaces {
		if err := deleteTenantUserBinding(client, namespace, tenant.ID, user.ID); err != nil {
			return err
		}
	}
	return nil
}

func applyTenantUserBinding(client *kubernetes.Clientset, namespace string, tenant Tenant, user TenantUser) error {
	ctx := context.Background()
	bindingName := managedName(tenant.ID, "user-"+user.ID)
	roleName := managedName(tenant.ID, "role")
	subject := buildSubject(user, namespace)

	binding := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      bindingName,
			Namespace: namespace,
			Labels: mergeLabels(managedLabels(tenant.ID), map[string]string{
				"klaw.io/tenant-user-id": user.ID,
			}),
		},
		Subjects: []rbacv1.Subject{subject},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "Role",
			Name:     roleName,
		},
	}

	existing, err := client.RbacV1().RoleBindings(namespace).Get(ctx, bindingName, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = client.RbacV1().RoleBindings(namespace).Create(ctx, binding, metav1.CreateOptions{})
		return err
	}
	if err != nil {
		return err
	}

	existing.Labels = binding.Labels
	existing.Subjects = binding.Subjects
	existing.RoleRef = binding.RoleRef
	_, err = client.RbacV1().RoleBindings(namespace).Update(ctx, existing, metav1.UpdateOptions{})
	return err
}

func deleteTenantUserBinding(client *kubernetes.Clientset, namespace, tenantID, userID string) error {
	ctx := context.Background()
	err := client.RbacV1().RoleBindings(namespace).Delete(ctx, managedName(tenantID, "user-"+userID), metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		return nil
	}
	return err
}

func buildSubject(user TenantUser, namespace string) rbacv1.Subject {
	subject := rbacv1.Subject{
		Kind:      normalizeSubjectKind(user.SubjectKind),
		APIGroup:  rbacAPIGroup(user.SubjectKind),
		Name:      user.SubjectName,
		Namespace: user.SubjectNamespace,
	}
	if subject.Kind == rbacv1.ServiceAccountKind {
		if subject.Namespace == "" {
			subject.Namespace = namespace
		}
		subject.APIGroup = ""
	} else {
		subject.Namespace = ""
	}
	return subject
}

func normalizeSubjectKind(kind string) string {
	switch strings.ToLower(kind) {
	case "group":
		return rbacv1.GroupKind
	case "serviceaccount", "service-account", "sa":
		return rbacv1.ServiceAccountKind
	default:
		return rbacv1.UserKind
	}
}

func rbacAPIGroup(kind string) string {
	if normalizeSubjectKind(kind) == rbacv1.ServiceAccountKind {
		return ""
	}
	return rbacv1.GroupName
}

func deleteTenantManagedResources(client *kubernetes.Clientset, namespace, tenantID string) error {
	ctx := context.Background()
	resources := []func() error{
		func() error {
			err := client.RbacV1().RoleBindings(namespace).Delete(ctx, managedName(tenantID, "binding"), metav1.DeleteOptions{})
			if apierrors.IsNotFound(err) {
				return nil
			}
			return err
		},
		func() error {
			err := client.RbacV1().Roles(namespace).Delete(ctx, managedName(tenantID, "role"), metav1.DeleteOptions{})
			if apierrors.IsNotFound(err) {
				return nil
			}
			return err
		},
		func() error {
			err := client.NetworkingV1().NetworkPolicies(namespace).Delete(ctx, managedName(tenantID, "default-deny"), metav1.DeleteOptions{})
			if apierrors.IsNotFound(err) {
				return nil
			}
			return err
		},
		func() error {
			err := client.CoreV1().ResourceQuotas(namespace).Delete(ctx, managedName(tenantID, "quota"), metav1.DeleteOptions{})
			if apierrors.IsNotFound(err) {
				return nil
			}
			return err
		},
	}

	for _, deleteFn := range resources {
		if err := deleteFn(); err != nil {
			return err
		}
	}
	return nil
}

func defaultPolicyRules(role string) []rbacv1.PolicyRule {
	switch role {
	case "admin":
		return []rbacv1.PolicyRule{{
			APIGroups: []string{"", "apps", "batch", "networking.k8s.io"},
			Resources: []string{"*"},
			Verbs:     []string{"*"},
		}}
	case "edit", "editor":
		return []rbacv1.PolicyRule{{
			APIGroups: []string{"", "apps", "batch", "networking.k8s.io"},
			Resources: []string{"pods", "services", "configmaps", "secrets", "deployments", "statefulsets", "jobs", "cronjobs", "ingresses"},
			Verbs:     []string{"get", "list", "watch", "create", "update", "patch", "delete"},
		}}
	default:
		return []rbacv1.PolicyRule{{
			APIGroups: []string{"", "apps", "batch", "networking.k8s.io"},
			Resources: []string{"pods", "services", "configmaps", "deployments", "statefulsets", "jobs", "cronjobs", "ingresses"},
			Verbs:     []string{"get", "list", "watch"},
		}}
	}
}

func managedLabels(tenantID string) map[string]string {
	return map[string]string{
		"app.kubernetes.io/managed-by": "klaw",
		"klaw.io/tenant-id":            tenantID,
	}
}

func mergeLabels(base map[string]string, extra map[string]string) map[string]string {
	merged := make(map[string]string, len(base)+len(extra))
	for key, value := range base {
		merged[key] = value
	}
	for key, value := range extra {
		merged[key] = value
	}
	return merged
}

func managedName(tenantID, suffix string) string {
	return fmt.Sprintf("%s-%s", tenantID, suffix)
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
