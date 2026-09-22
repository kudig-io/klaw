package tenancy

import (
	"context"
	"reflect"
	"testing"

	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

// TestApplyTenantUserBindingCreatesRoleAndBinding 验证租户用户绑定流程：
// applyTenantRBAC 先在 fake 集群创建受管 Role（承载 rules），
// applyTenantUserBinding 再创建指向该 Role 的 RoleBinding（承载 subjects）。
func TestApplyTenantUserBindingCreatesRoleAndBinding(t *testing.T) {
	ctx := context.Background()
	cs := fake.NewSimpleClientset()

	tenant := Tenant{
		ID:         "tenant-1",
		Name:       "Team A",
		Namespaces: []string{"team-a"},
		RBAC:       RBACPolicy{Enabled: true, DefaultRole: "edit"},
	}
	user := TenantUser{
		ID:               "user-1",
		TenantID:         tenant.ID,
		Username:         "alice",
		Role:             "editor",
		SubjectKind:      "service-account",
		SubjectName:      "tenant-operator",
		SubjectNamespace: "team-a",
	}

	if err := applyTenantRBAC(cs, "team-a", tenant); err != nil {
		t.Fatalf("applyTenantRBAC() error = %v", err)
	}
	if err := applyTenantUserBinding(cs, "team-a", tenant, user); err != nil {
		t.Fatalf("applyTenantUserBinding() error = %v", err)
	}

	// Role：名称、rules 与受管标签
	role, err := cs.RbacV1().Roles("team-a").Get(ctx, "tenant-1-role", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("get role: %v", err)
	}
	wantRules := defaultPolicyRules("edit")
	if !reflect.DeepEqual(role.Rules, wantRules) {
		t.Fatalf("role.Rules = %+v, want %+v", role.Rules, wantRules)
	}
	if role.Labels["klaw.io/tenant-id"] != tenant.ID {
		t.Errorf("role label klaw.io/tenant-id = %q, want %q", role.Labels["klaw.io/tenant-id"], tenant.ID)
	}
	if role.Labels["app.kubernetes.io/managed-by"] != "klaw" {
		t.Errorf("role label managed-by = %q, want klaw", role.Labels["app.kubernetes.io/managed-by"])
	}

	// RoleBinding：受管命名、subjects、roleRef、用户标签
	binding, err := cs.RbacV1().RoleBindings("team-a").Get(ctx, "tenant-1-user-user-1", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("get role binding: %v", err)
	}
	wantSubjects := []rbacv1.Subject{{
		Kind:      rbacv1.ServiceAccountKind,
		APIGroup:  "",
		Name:      "tenant-operator",
		Namespace: "team-a",
	}}
	if !reflect.DeepEqual(binding.Subjects, wantSubjects) {
		t.Fatalf("binding.Subjects = %+v, want %+v", binding.Subjects, wantSubjects)
	}
	wantRef := rbacv1.RoleRef{APIGroup: rbacv1.GroupName, Kind: "Role", Name: "tenant-1-role"}
	if binding.RoleRef != wantRef {
		t.Fatalf("binding.RoleRef = %+v, want %+v", binding.RoleRef, wantRef)
	}
	if binding.Labels["klaw.io/tenant-user-id"] != user.ID {
		t.Errorf("binding label tenant-user-id = %q, want %q", binding.Labels["klaw.io/tenant-user-id"], user.ID)
	}
}

// TestTenantUserBindingSubjectNormalization 表驱动验证不同 subjectKind 的
// 归一化规则：User/Group 带 rbac APIGroup，ServiceAccount 无 APIGroup 且
// 缺省命名空间回退到绑定命名空间。
func TestTenantUserBindingSubjectNormalization(t *testing.T) {
	tests := []struct {
		desc             string
		subjectKind      string
		subjectNamespace string
		wantSubject      rbacv1.Subject
	}{
		{
			desc:        "user 类型携带 rbac APIGroup 且无命名空间",
			subjectKind: "user",
			wantSubject: rbacv1.Subject{Kind: rbacv1.UserKind, APIGroup: rbacv1.GroupName, Name: "alice"},
		},
		{
			desc:        "group 类型携带 rbac APIGroup 且无命名空间",
			subjectKind: "group",
			wantSubject: rbacv1.Subject{Kind: rbacv1.GroupKind, APIGroup: rbacv1.GroupName, Name: "alice"},
		},
		{
			desc:        "serviceaccount 简写类型缺省命名空间回退到绑定命名空间",
			subjectKind: "sa",
			wantSubject: rbacv1.Subject{Kind: rbacv1.ServiceAccountKind, APIGroup: "", Name: "alice", Namespace: "team-a"},
		},
		{
			desc:             "serviceaccount 显式命名空间原样保留",
			subjectKind:      "serviceaccount",
			subjectNamespace: "sa-ns",
			wantSubject:      rbacv1.Subject{Kind: rbacv1.ServiceAccountKind, APIGroup: "", Name: "alice", Namespace: "sa-ns"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			ctx := context.Background()
			cs := fake.NewSimpleClientset()
			tenant := Tenant{ID: "tenant-1", Namespaces: []string{"team-a"}, RBAC: RBACPolicy{Enabled: true}}
			user := TenantUser{
				ID:               "u1",
				TenantID:         tenant.ID,
				Username:         "alice",
				SubjectKind:      tc.subjectKind,
				SubjectName:      "alice",
				SubjectNamespace: tc.subjectNamespace,
			}

			if err := applyTenantUserBinding(cs, "team-a", tenant, user); err != nil {
				t.Fatalf("applyTenantUserBinding() error = %v", err)
			}
			binding, err := cs.RbacV1().RoleBindings("team-a").Get(ctx, "tenant-1-user-u1", metav1.GetOptions{})
			if err != nil {
				t.Fatalf("get role binding: %v", err)
			}
			if !reflect.DeepEqual(binding.Subjects, []rbacv1.Subject{tc.wantSubject}) {
				t.Fatalf("binding.Subjects = %+v, want [%+v]", binding.Subjects, tc.wantSubject)
			}
		})
	}
}

// TestApplyTenantUserBindingIsIdempotent 验证重复应用只更新既有 binding，
// 不产生重复对象，且 subjects 以最新一次为准。
func TestApplyTenantUserBindingIsIdempotent(t *testing.T) {
	ctx := context.Background()
	cs := fake.NewSimpleClientset()
	tenant := Tenant{ID: "tenant-1", Namespaces: []string{"team-a"}, RBAC: RBACPolicy{Enabled: true}}

	first := TenantUser{ID: "u1", TenantID: tenant.ID, Username: "alice", SubjectName: "alice"}
	if err := applyTenantUserBinding(cs, "team-a", tenant, first); err != nil {
		t.Fatalf("first apply error = %v", err)
	}
	second := TenantUser{ID: "u1", TenantID: tenant.ID, Username: "alice", SubjectName: "alice-v2"}
	if err := applyTenantUserBinding(cs, "team-a", tenant, second); err != nil {
		t.Fatalf("second apply error = %v", err)
	}

	list, err := cs.RbacV1().RoleBindings("team-a").List(ctx, metav1.ListOptions{})
	if err != nil {
		t.Fatalf("list role bindings: %v", err)
	}
	if len(list.Items) != 1 {
		t.Fatalf("len(roleBindings) = %d, want 1", len(list.Items))
	}
	if len(list.Items[0].Subjects) != 1 || list.Items[0].Subjects[0].Name != "alice-v2" {
		t.Fatalf("subjects after re-apply = %+v, want alice-v2", list.Items[0].Subjects)
	}
}

// TestDeleteTenantUserBindingRemovesBinding 验证删除绑定后资源从集群消失，
// 且重复删除幂等不报错。
func TestDeleteTenantUserBindingRemovesBinding(t *testing.T) {
	ctx := context.Background()
	cs := fake.NewSimpleClientset()
	tenant := Tenant{ID: "tenant-1", Namespaces: []string{"team-a"}, RBAC: RBACPolicy{Enabled: true}}
	user := TenantUser{ID: "u1", TenantID: tenant.ID, Username: "alice", SubjectName: "alice"}

	if err := applyTenantUserBinding(cs, "team-a", tenant, user); err != nil {
		t.Fatalf("applyTenantUserBinding() error = %v", err)
	}
	if err := deleteTenantUserBinding(cs, "team-a", tenant.ID, user.ID); err != nil {
		t.Fatalf("deleteTenantUserBinding() error = %v", err)
	}
	_, err := cs.RbacV1().RoleBindings("team-a").Get(ctx, "tenant-1-user-u1", metav1.GetOptions{})
	if !apierrors.IsNotFound(err) {
		t.Fatalf("binding should be gone after delete, got err = %v", err)
	}

	// 重复删除应视为成功（NotFound 吞掉）
	if err := deleteTenantUserBinding(cs, "team-a", tenant.ID, user.ID); err != nil {
		t.Fatalf("repeat deleteTenantUserBinding() error = %v, want nil", err)
	}
}
