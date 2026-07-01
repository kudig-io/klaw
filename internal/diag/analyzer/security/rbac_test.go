package security

import (
	"context"
	"testing"

	"github.com/kudig-io/klaw/internal/diag/types"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestRBACAdminAnalyzer_Analyze_TooManyAdmins(t *testing.T) {
	crb1 := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{Name: "admin-user-1"},
		RoleRef:    rbacv1.RoleRef{Name: "cluster-admin"},
		Subjects:   []rbacv1.Subject{{Kind: "User", Name: "admin1"}},
	}
	crb2 := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{Name: "admin-user-2"},
		RoleRef:    rbacv1.RoleRef{Name: "cluster-admin"},
		Subjects:   []rbacv1.Subject{{Kind: "User", Name: "admin2"}},
	}
	crb3 := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{Name: "admin-user-3"},
		RoleRef:    rbacv1.RoleRef{Name: "cluster-admin"},
		Subjects:   []rbacv1.Subject{{Kind: "User", Name: "admin3"}},
	}
	crb4 := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{Name: "admin-user-4"},
		RoleRef:    rbacv1.RoleRef{Name: "cluster-admin"},
		Subjects:   []rbacv1.Subject{{Kind: "Group", Name: "admins"}},
	}

	client := fake.NewSimpleClientset(crb1, crb2, crb3, crb4)
	data := &types.DiagnosticData{K8sClient: client}

	analyzer := NewRBACAdminAnalyzer()
	issues, err := analyzer.Analyze(context.Background(), data)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	hasTooManyAdmins := false
	for _, issue := range issues {
		if issue.ENName == "RBAC_TOO_MANY_ADMINS" {
			hasTooManyAdmins = true
		}
	}

	if !hasTooManyAdmins {
		t.Errorf("expected RBAC_TOO_MANY_ADMINS issue, got %v", issues)
	}
}

func TestRBACServiceAccountAnalyzer_Analyze_AutomountEnabled(t *testing.T) {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: "test-ns"},
	}
	automount := true
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default",
			Namespace: "test-ns",
		},
		AutomountServiceAccountToken: &automount,
	}

	client := fake.NewSimpleClientset(ns, sa)
	data := &types.DiagnosticData{K8sClient: client}

	analyzer := NewRBACServiceAccountAnalyzer()
	issues, err := analyzer.Analyze(context.Background(), data)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	hasAutomount := false
	for _, issue := range issues {
		if issue.ENName == "RBAC_DEFAULT_SA_AUTOMOUNT" {
			hasAutomount = true
		}
	}

	if !hasAutomount {
		t.Errorf("expected RBAC_DEFAULT_SA_AUTOMOUNT issue, got %v", issues)
	}
}

func TestRBACDangerousPermissionsAnalyzer_Analyze_Escalate(t *testing.T) {
	cr := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{Name: "dangerous-role"},
		Rules: []rbacv1.PolicyRule{
			{
				Verbs:     []string{"*"},
				Resources: []string{"*"},
			},
		},
	}

	client := fake.NewSimpleClientset(cr)
	data := &types.DiagnosticData{K8sClient: client}

	analyzer := NewRBACDangerousPermissionsAnalyzer()
	issues, err := analyzer.Analyze(context.Background(), data)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	hasDangerous := false
	for _, issue := range issues {
		if issue.ENName == "RBAC_DANGEROUS_PERMISSIONS" {
			hasDangerous = true
		}
	}

	if !hasDangerous {
		t.Errorf("expected RBAC_DANGEROUS_PERMISSIONS issue, got %v", issues)
	}
}

func TestRBACUnusedRolesAnalyzer_Analyze_UnusedRoles(t *testing.T) {
	// Create many unused ClusterRoles - use explicit list instead of slice
	cr0 := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{Name: "unused-role-0"},
		Rules:      []rbacv1.PolicyRule{{Verbs: []string{"get"}, Resources: []string{"pods"}}},
	}
	cr1 := &rbacv1.ClusterRole{ObjectMeta: metav1.ObjectMeta{Name: "unused-role-1"}, Rules: []rbacv1.PolicyRule{{Verbs: []string{"get"}, Resources: []string{"pods"}}}}
	cr2 := &rbacv1.ClusterRole{ObjectMeta: metav1.ObjectMeta{Name: "unused-role-2"}, Rules: []rbacv1.PolicyRule{{Verbs: []string{"get"}, Resources: []string{"pods"}}}}
	cr3 := &rbacv1.ClusterRole{ObjectMeta: metav1.ObjectMeta{Name: "unused-role-3"}, Rules: []rbacv1.PolicyRule{{Verbs: []string{"get"}, Resources: []string{"pods"}}}}
	cr4 := &rbacv1.ClusterRole{ObjectMeta: metav1.ObjectMeta{Name: "unused-role-4"}, Rules: []rbacv1.PolicyRule{{Verbs: []string{"get"}, Resources: []string{"pods"}}}}
	cr5 := &rbacv1.ClusterRole{ObjectMeta: metav1.ObjectMeta{Name: "unused-role-5"}, Rules: []rbacv1.PolicyRule{{Verbs: []string{"get"}, Resources: []string{"pods"}}}}
	cr6 := &rbacv1.ClusterRole{ObjectMeta: metav1.ObjectMeta{Name: "unused-role-6"}, Rules: []rbacv1.PolicyRule{{Verbs: []string{"get"}, Resources: []string{"pods"}}}}
	cr7 := &rbacv1.ClusterRole{ObjectMeta: metav1.ObjectMeta{Name: "unused-role-7"}, Rules: []rbacv1.PolicyRule{{Verbs: []string{"get"}, Resources: []string{"pods"}}}}
	cr8 := &rbacv1.ClusterRole{ObjectMeta: metav1.ObjectMeta{Name: "unused-role-8"}, Rules: []rbacv1.PolicyRule{{Verbs: []string{"get"}, Resources: []string{"pods"}}}}
	cr9 := &rbacv1.ClusterRole{ObjectMeta: metav1.ObjectMeta{Name: "unused-role-9"}, Rules: []rbacv1.PolicyRule{{Verbs: []string{"get"}, Resources: []string{"pods"}}}}
	cr10 := &rbacv1.ClusterRole{ObjectMeta: metav1.ObjectMeta{Name: "unused-role-10"}, Rules: []rbacv1.PolicyRule{{Verbs: []string{"get"}, Resources: []string{"pods"}}}}
	cr11 := &rbacv1.ClusterRole{ObjectMeta: metav1.ObjectMeta{Name: "unused-role-11"}, Rules: []rbacv1.PolicyRule{{Verbs: []string{"get"}, Resources: []string{"pods"}}}}
	cr12 := &rbacv1.ClusterRole{ObjectMeta: metav1.ObjectMeta{Name: "unused-role-12"}, Rules: []rbacv1.PolicyRule{{Verbs: []string{"get"}, Resources: []string{"pods"}}}}
	cr13 := &rbacv1.ClusterRole{ObjectMeta: metav1.ObjectMeta{Name: "unused-role-13"}, Rules: []rbacv1.PolicyRule{{Verbs: []string{"get"}, Resources: []string{"pods"}}}}
	cr14 := &rbacv1.ClusterRole{ObjectMeta: metav1.ObjectMeta{Name: "unused-role-14"}, Rules: []rbacv1.PolicyRule{{Verbs: []string{"get"}, Resources: []string{"pods"}}}}

	// Create only one binding
	crb := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{Name: "used-binding"},
		RoleRef:    rbacv1.RoleRef{Name: "unused-role-0"},
		Subjects:   []rbacv1.Subject{{Kind: "User", Name: "user1"}},
	}

	client := fake.NewSimpleClientset(cr0, cr1, cr2, cr3, cr4, cr5, cr6, cr7, cr8, cr9, cr10, cr11, cr12, cr13, cr14, crb)
	data := &types.DiagnosticData{K8sClient: client}

	analyzer := NewRBACUnusedRolesAnalyzer()
	issues, err := analyzer.Analyze(context.Background(), data)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	hasUnused := false
	for _, issue := range issues {
		if issue.ENName == "RBAC_UNUSED_CLUSTER_ROLES" {
			hasUnused = true
		}
	}

	if !hasUnused {
		t.Errorf("expected RBAC_UNUSED_CLUSTER_ROLES issue, got %v", issues)
	}
}

func TestRBACTokenAnalyzer_Analyze_MultipleTokens(t *testing.T) {
	// Create a ServiceAccount
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-sa",
			Namespace: "default",
		},
	}

	// Create multiple tokens for the same SA
	s0 := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-sa-token-0",
			Namespace: "default",
			Annotations: map[string]string{
				"kubernetes.io/service-account.name": "test-sa",
			},
		},
		Type: corev1.SecretTypeServiceAccountToken,
	}
	s1 := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "test-sa-token-1", Namespace: "default", Annotations: map[string]string{"kubernetes.io/service-account.name": "test-sa"}}, Type: corev1.SecretTypeServiceAccountToken}
	s2 := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "test-sa-token-2", Namespace: "default", Annotations: map[string]string{"kubernetes.io/service-account.name": "test-sa"}}, Type: corev1.SecretTypeServiceAccountToken}
	s3 := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "test-sa-token-3", Namespace: "default", Annotations: map[string]string{"kubernetes.io/service-account.name": "test-sa"}}, Type: corev1.SecretTypeServiceAccountToken}
	s4 := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "test-sa-token-4", Namespace: "default", Annotations: map[string]string{"kubernetes.io/service-account.name": "test-sa"}}, Type: corev1.SecretTypeServiceAccountToken}

	client := fake.NewSimpleClientset(sa, s0, s1, s2, s3, s4)
	data := &types.DiagnosticData{K8sClient: client}

	analyzer := NewRBACTokenAnalyzer()
	issues, err := analyzer.Analyze(context.Background(), data)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	hasMultipleTokens := false
	for _, issue := range issues {
		if issue.ENName == "RBAC_MULTIPLE_TOKENS" {
			hasMultipleTokens = true
		}
	}

	if !hasMultipleTokens {
		t.Errorf("expected RBAC_MULTIPLE_TOKENS issue, got %v", issues)
	}
}

func TestRBACAnalyzer_NoK8sClient(t *testing.T) {
	data := &types.DiagnosticData{K8sClient: nil}

	analyzer := NewRBACAdminAnalyzer()
	issues, err := analyzer.Analyze(context.Background(), data)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("expected no issues without k8s client, got %d", len(issues))
	}
}
