package tenancy

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// policies.go 租户在集群上的资源供给：Namespace、ResourceQuota、NetworkPolicy 与 RBAC
// 的应用、清理与调和，以及受管资源命名/标签辅助。

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
