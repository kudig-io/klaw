package tenancy

import (
	"context"
	"fmt"
	"strings"
	"time"

	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// users.go 租户用户管理：用户 CRUD、用户 RBAC 绑定在集群上的应用与清理、Subject 归一化。

func (m *Manager) ListUsers(tenantID, role string) []TenantUser {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result = make([]TenantUser, 0, len(m.users))
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

func applyTenantUserBinding(client kubernetes.Interface, namespace string, tenant Tenant, user TenantUser) error {
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

func deleteTenantUserBinding(client kubernetes.Interface, namespace, tenantID, userID string) error {
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
