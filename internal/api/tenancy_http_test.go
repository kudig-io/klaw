package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kudig-io/klaw/internal/tenancy"
)

// registerTenancyRoutes 注册与 SetupRoutes 等价的租户/租户用户 v1 路由。
func registerTenancyRoutes(s *Server) {
	s.router.HandleFunc("/api/v1/tenants", s.handleListTenants).Methods("GET")
	s.router.HandleFunc("/api/v1/tenants", s.handleCreateTenant).Methods("POST")
	s.router.HandleFunc("/api/v1/tenants/stats", s.handleTenantStatistics).Methods("GET")
	s.router.HandleFunc("/api/v1/tenants/{id}", s.handleGetTenant).Methods("GET")
	s.router.HandleFunc("/api/v1/tenants/{id}", s.handleUpdateTenant).Methods("PUT")
	s.router.HandleFunc("/api/v1/tenants/{id}", s.handleDeleteTenant).Methods("DELETE")
	s.router.HandleFunc("/api/v1/tenant-users", s.handleListTenantUsers).Methods("GET")
	s.router.HandleFunc("/api/v1/tenant-users", s.handleCreateTenantUser).Methods("POST")
	s.router.HandleFunc("/api/v1/tenant-users/{id}", s.handleDeleteTenantUser).Methods("DELETE")
}

// TestTenantHTTPCRUDLifecycle 走 HTTP 覆盖租户 CRUD 全生命周期，
// 并断言创建/更新/删除在 fake 集群上产生的受管资源变化。
func TestTenantHTTPCRUDLifecycle(t *testing.T) {
	ctx := context.Background()
	s, cs := newK8sTestServer(t)
	registerTenancyRoutes(s)

	// 1. 创建前：名字为空应被 400 拒绝
	w := doRequest(t, s, "POST", "/api/v1/tenants", tenancy.Tenant{Name: " "})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("empty-name status = %d, want %d, body = %s", w.Code, http.StatusBadRequest, w.Body.String())
	}

	// 2. 创建租户：默认值应被填充
	w = doRequest(t, s, "POST", "/api/v1/tenants", tenancy.Tenant{
		Cluster:    testClusterName,
		Name:       "Team A",
		Namespaces: []string{"team-a"},
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("create status = %d, body = %s", w.Code, w.Body.String())
	}
	var created tenancy.Tenant
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatalf("unmarshal created: %v", err)
	}
	if !strings.HasPrefix(created.ID, "tenant-") {
		t.Fatalf("created.ID = %q, want prefix tenant-", created.ID)
	}
	if created.Name != "Team A" || created.Cluster != testClusterName {
		t.Fatalf("created = %+v, want name Team A cluster %s", created, testClusterName)
	}
	if len(created.Namespaces) != 1 || created.Namespaces[0] != "team-a" {
		t.Fatalf("created.Namespaces = %v, want [team-a]", created.Namespaces)
	}
	if !created.RBAC.Enabled || created.RBAC.DefaultRole != "view" {
		t.Fatalf("created.RBAC = %+v, want enabled with default role view", created.RBAC)
	}
	if created.ResourceQuotas.CPU != "10" || created.ResourceQuotas.Memory != "20Gi" || created.ResourceQuotas.Pods != "100" {
		t.Fatalf("created.ResourceQuotas = %+v, want default quotas", created.ResourceQuotas)
	}

	// 3. 集群侧：Namespace/ResourceQuota/Role/RoleBinding 应随创建落地
	nsObj, err := cs.CoreV1().Namespaces().Get(ctx, "team-a", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("namespace team-a not created: %v", err)
	}
	if nsObj.Labels["klaw.io/tenant-id"] != created.ID || nsObj.Labels["app.kubernetes.io/managed-by"] != "klaw" {
		t.Fatalf("namespace labels = %v, want managed-by klaw + tenant-id", nsObj.Labels)
	}
	quota, err := cs.CoreV1().ResourceQuotas("team-a").Get(ctx, created.ID+"-quota", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("quota not created: %v", err)
	}
	if cpu := quota.Spec.Hard[corev1.ResourceLimitsCPU]; cpu.String() != "10" {
		t.Fatalf("quota cpu hard = %q, want 10", cpu.String())
	}
	if _, err := cs.RbacV1().Roles("team-a").Get(ctx, created.ID+"-role", metav1.GetOptions{}); err != nil {
		t.Fatalf("role not created: %v", err)
	}
	if _, err := cs.RbacV1().RoleBindings("team-a").Get(ctx, created.ID+"-binding", metav1.GetOptions{}); err != nil {
		t.Fatalf("role binding not created: %v", err)
	}

	// 4. 列表：按 cluster 过滤
	w = doRequest(t, s, "GET", "/api/v1/tenants?cluster="+testClusterName, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("list status = %d", w.Code)
	}
	var listed []tenancy.Tenant
	if err := json.Unmarshal(w.Body.Bytes(), &listed); err != nil {
		t.Fatalf("unmarshal list: %v", err)
	}
	if len(listed) != 1 || listed[0].ID != created.ID {
		t.Fatalf("filtered list = %+v, want exactly the created tenant", listed)
	}

	// 5. 详情：不存在返回 404
	w = doRequest(t, s, "GET", "/api/v1/tenants/tenant-nope", nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("get missing status = %d, want %d", w.Code, http.StatusNotFound)
	}

	// 6. 更新：改名但保留 ID/Cluster/Namespaces
	w = doRequest(t, s, "PUT", "/api/v1/tenants/"+created.ID, tenancy.Tenant{
		Name:        "Team A Renamed",
		Description: "updated via api",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("update status = %d, body = %s", w.Code, w.Body.String())
	}
	var updated tenancy.Tenant
	if err := json.Unmarshal(w.Body.Bytes(), &updated); err != nil {
		t.Fatalf("unmarshal updated: %v", err)
	}
	if updated.ID != created.ID || updated.Name != "Team A Renamed" {
		t.Fatalf("updated = %+v, want same id with new name", updated)
	}
	if updated.Cluster != testClusterName {
		t.Fatalf("updated.Cluster = %q, want %q", updated.Cluster, testClusterName)
	}
	if len(updated.Namespaces) != 1 || updated.Namespaces[0] != "team-a" {
		t.Fatalf("updated.Namespaces = %v, want preserved [team-a]", updated.Namespaces)
	}
	if updated.UpdatedAt.Before(updated.CreatedAt) {
		t.Fatalf("UpdatedAt %v should not be before CreatedAt %v", updated.UpdatedAt, updated.CreatedAt)
	}

	// 7. 删除：租户消失，受管集群资源被清理，Namespace 本体保留
	w = doRequest(t, s, "DELETE", "/api/v1/tenants/"+created.ID, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("delete status = %d, body = %s", w.Code, w.Body.String())
	}
	w = doRequest(t, s, "GET", "/api/v1/tenants/"+created.ID, nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("get after delete status = %d, want %d", w.Code, http.StatusNotFound)
	}
	if _, err := cs.CoreV1().ResourceQuotas("team-a").Get(ctx, created.ID+"-quota", metav1.GetOptions{}); !apierrors.IsNotFound(err) {
		t.Fatalf("quota should be deleted after tenant delete, got err = %v", err)
	}
	if _, err := cs.RbacV1().Roles("team-a").Get(ctx, created.ID+"-role", metav1.GetOptions{}); !apierrors.IsNotFound(err) {
		t.Fatalf("role should be deleted after tenant delete, got err = %v", err)
	}
	if _, err := cs.RbacV1().RoleBindings("team-a").Get(ctx, created.ID+"-binding", metav1.GetOptions{}); !apierrors.IsNotFound(err) {
		t.Fatalf("role binding should be deleted after tenant delete, got err = %v", err)
	}
	if _, err := cs.CoreV1().Namespaces().Get(ctx, "team-a", metav1.GetOptions{}); err != nil {
		t.Fatalf("namespace itself should be preserved, got err = %v", err)
	}
}

// TestTenantUserHTTPCRUDLifecycle 走 HTTP 覆盖租户用户 CRUD，
// 并断言用户在 fake 集群上的 RoleBinding 应用与清理。
func TestTenantUserHTTPCRUDLifecycle(t *testing.T) {
	ctx := context.Background()
	s, cs := newK8sTestServer(t)
	registerTenancyRoutes(s)

	// 准备租户
	w := doRequest(t, s, "POST", "/api/v1/tenants", tenancy.Tenant{
		Cluster:    testClusterName,
		Name:       "Team B",
		Namespaces: []string{"team-b"},
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("create tenant status = %d, body = %s", w.Code, w.Body.String())
	}
	var tenant tenancy.Tenant
	if err := json.Unmarshal(w.Body.Bytes(), &tenant); err != nil {
		t.Fatalf("unmarshal tenant: %v", err)
	}

	// 1. 未知租户的用户应被 400 拒绝
	w = doRequest(t, s, "POST", "/api/v1/tenant-users", tenancy.TenantUser{TenantID: "tenant-nope", Username: "bob"})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("unknown tenant status = %d, want %d, body = %s", w.Code, http.StatusBadRequest, w.Body.String())
	}

	// 2. 创建用户：namespaces 缺省回退到租户 namespaces
	w = doRequest(t, s, "POST", "/api/v1/tenant-users", tenancy.TenantUser{
		TenantID: tenant.ID,
		Username: "alice",
		Role:     "editor",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("create user status = %d, body = %s", w.Code, w.Body.String())
	}
	var user tenancy.TenantUser
	if err := json.Unmarshal(w.Body.Bytes(), &user); err != nil {
		t.Fatalf("unmarshal user: %v", err)
	}
	if !strings.HasPrefix(user.ID, "user-") {
		t.Fatalf("user.ID = %q, want prefix user-", user.ID)
	}
	if user.Role != "editor" {
		t.Fatalf("user.Role = %q, want editor", user.Role)
	}
	if len(user.Namespaces) != 1 || user.Namespaces[0] != "team-b" {
		t.Fatalf("user.Namespaces = %v, want fallback [team-b]", user.Namespaces)
	}
	if user.SubjectKind != "User" || user.SubjectName != "alice" {
		t.Fatalf("user subject = %s/%s, want User/alice", user.SubjectKind, user.SubjectName)
	}

	// 3. 集群侧：用户 RoleBinding 指向租户受管 Role
	binding, err := cs.RbacV1().RoleBindings("team-b").Get(ctx, tenant.ID+"-user-"+user.ID, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("user role binding not created: %v", err)
	}
	if len(binding.Subjects) != 1 || binding.Subjects[0].Kind != "User" || binding.Subjects[0].Name != "alice" {
		t.Fatalf("binding subjects = %+v, want single User alice", binding.Subjects)
	}
	if binding.RoleRef.Name != tenant.ID+"-role" || binding.RoleRef.Kind != "Role" {
		t.Fatalf("binding roleRef = %+v, want Role %s-role", binding.RoleRef, tenant.ID)
	}

	// 4. 列表 + 过滤
	w = doRequest(t, s, "GET", "/api/v1/tenant-users?tenantId="+tenant.ID, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("list users status = %d", w.Code)
	}
	var users []tenancy.TenantUser
	if err := json.Unmarshal(w.Body.Bytes(), &users); err != nil {
		t.Fatalf("unmarshal users: %v", err)
	}
	if len(users) != 1 || users[0].ID != user.ID {
		t.Fatalf("users = %+v, want exactly alice", users)
	}
	w = doRequest(t, s, "GET", "/api/v1/tenant-users?role=viewer", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("list users by role status = %d", w.Code)
	}
	users = nil
	if err := json.Unmarshal(w.Body.Bytes(), &users); err != nil {
		t.Fatalf("unmarshal users: %v", err)
	}
	if len(users) != 0 {
		t.Fatalf("viewer filter should return empty, got %+v", users)
	}

	// 5. 删除用户：binding 从集群消失
	w = doRequest(t, s, "DELETE", "/api/v1/tenant-users/"+user.ID, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("delete user status = %d, body = %s", w.Code, w.Body.String())
	}
	w = doRequest(t, s, "GET", "/api/v1/tenant-users?tenantId="+tenant.ID, nil)
	users = nil
	_ = json.Unmarshal(w.Body.Bytes(), &users)
	if len(users) != 0 {
		t.Fatalf("users after delete = %+v, want empty", users)
	}
	if _, err := cs.RbacV1().RoleBindings("team-b").Get(ctx, tenant.ID+"-user-"+user.ID, metav1.GetOptions{}); !apierrors.IsNotFound(err) {
		t.Fatalf("user binding should be deleted, got err = %v", err)
	}
}

// TestTenantStatisticsEndpoint 验证 /api/v1/tenants/stats 的统计口径。
func TestTenantStatisticsEndpoint(t *testing.T) {
	s, _ := newK8sTestServer(t)
	registerTenancyRoutes(s)

	// 默认租户占 3 个 namespace（default/kube-system/kube-public）
	w := doRequest(t, s, "POST", "/api/v1/tenants", tenancy.Tenant{
		Cluster:    testClusterName,
		Name:       "Stats Team",
		Namespaces: []string{"stats-a", "stats-b"},
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("create tenant status = %d, body = %s", w.Code, w.Body.String())
	}
	var tenant tenancy.Tenant
	if err := json.Unmarshal(w.Body.Bytes(), &tenant); err != nil {
		t.Fatalf("unmarshal tenant: %v", err)
	}

	// 两个用户，分别落在不同角色
	for _, username := range []string{"alice", "bob"} {
		w = doRequest(t, s, "POST", "/api/v1/tenant-users", tenancy.TenantUser{
			TenantID: tenant.ID,
			Username: username,
			Role:     "editor",
		})
		if w.Code != http.StatusCreated {
			t.Fatalf("create user %s status = %d, body = %s", username, w.Code, w.Body.String())
		}
	}

	w = doRequest(t, s, "GET", "/api/v1/tenants/stats", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("stats status = %d, body = %s", w.Code, w.Body.String())
	}
	var got tenancy.Statistics
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal stats: %v", err)
	}
	if got.TotalTenants != 2 {
		t.Errorf("totalTenants = %d, want 2（default + Stats Team）", got.TotalTenants)
	}
	if got.TotalNamespaces != 5 {
		t.Errorf("totalNamespaces = %d, want 5（default 租户 3 + 新租户 2）", got.TotalNamespaces)
	}
	if got.TotalUsers != 2 {
		t.Errorf("totalUsers = %d, want 2", got.TotalUsers)
	}
	if got.UsersByRole["editor"] != 2 {
		t.Errorf("usersByRole[editor] = %d, want 2", got.UsersByRole["editor"])
	}
}
