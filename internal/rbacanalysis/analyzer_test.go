package rbacanalysis

import (
	"context"
	"testing"

	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestAnalyzeRBAC(t *testing.T) {
	client := fake.NewSimpleClientset(
		&rbacv1.Role{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "reader",
				Namespace: "default",
			},
		},
		&rbacv1.ClusterRole{
			ObjectMeta: metav1.ObjectMeta{Name: "cluster-admin"},
		},
		&rbacv1.RoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "reader-binding",
				Namespace: "default",
			},
			RoleRef: rbacv1.RoleRef{
				APIGroup: rbacv1.GroupName,
				Kind:     "Role",
				Name:     "reader",
			},
			Subjects: []rbacv1.Subject{{
				Kind:      "ServiceAccount",
				Name:      "default",
				Namespace: "default",
			}},
		},
		&rbacv1.ClusterRoleBinding{
			ObjectMeta: metav1.ObjectMeta{Name: "admins"},
			RoleRef: rbacv1.RoleRef{
				APIGroup: rbacv1.GroupName,
				Kind:     "ClusterRole",
				Name:     "cluster-admin",
			},
			Subjects: []rbacv1.Subject{{
				Kind: "User",
				Name: "admin@example.com",
			}},
		},
	)

	analysis, err := NewAnalyzer(client).AnalyzeRBAC(context.Background())
	if err != nil {
		t.Fatalf("AnalyzeRBAC() error = %v", err)
	}

	if analysis.TotalRoles != 1 {
		t.Fatalf("TotalRoles = %d, want 1", analysis.TotalRoles)
	}
	if analysis.TotalClusterRoles != 1 {
		t.Fatalf("TotalClusterRoles = %d, want 1", analysis.TotalClusterRoles)
	}
	if analysis.TotalBindings != 1 {
		t.Fatalf("TotalBindings = %d, want 1", analysis.TotalBindings)
	}
	if analysis.TotalClusterBindings != 1 {
		t.Fatalf("TotalClusterBindings = %d, want 1", analysis.TotalClusterBindings)
	}
	if len(analysis.BindingsBySubject) != 2 {
		t.Fatalf("BindingsBySubject = %d, want 2", len(analysis.BindingsBySubject))
	}
}
