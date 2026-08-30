// Package security provides security compliance analyzers
package security

import (
	"context"
	"fmt"

	"github.com/kudig-io/klaw/internal/diag/analyzer"
	"github.com/kudig-io/klaw/internal/diag/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RBACAdminAnalyzer checks for cluster-admin role bindings
type RBACAdminAnalyzer struct {
	*analyzer.BaseAnalyzer
}

// NewRBACAdminAnalyzer creates a new RBAC admin analyzer
func NewRBACAdminAnalyzer() *RBACAdminAnalyzer {
	return &RBACAdminAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer(
			"security.rbac.admin",
			"检查 RBAC 管理员权限",
			"security",
			[]types.DataMode{types.ModeOnline},
		),
	}
}

// Analyze performs RBAC admin checks
func (a *RBACAdminAnalyzer) Analyze(ctx context.Context, data *types.DiagnosticData) ([]types.Issue, error) {
	var issues []types.Issue

	if !data.HasK8sClient() {
		return issues, nil
	}

	client := data.K8sClient

	// Check ClusterRoleBindings for cluster-admin
	crbs, err := client.RbacV1().ClusterRoleBindings().List(ctx, metav1.ListOptions{})
	if err != nil {
		return issues, nil
	}

	adminBindings := []string{}
	for _, crb := range crbs.Items {
		if crb.RoleRef.Name == "cluster-admin" {
			for _, subject := range crb.Subjects {
				if subject.Kind == "User" || subject.Kind == "Group" {
					adminBindings = append(adminBindings, fmt.Sprintf("%s (%s)", subject.Name, subject.Kind))
				}
			}
		}
	}

	if len(adminBindings) > 3 {
		issue := types.NewIssue(
			types.SeverityWarning,
			fmt.Sprintf("发现 %d 个 cluster-admin 绑定", len(adminBindings)),
			"RBAC_TOO_MANY_ADMINS",
			fmt.Sprintf("大量用户/组绑定了 cluster-admin 角色: %v", adminBindings),
			"cluster/rbac",
		).WithRemediation("审查 cluster-admin 权限，遵循最小权限原则")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	}

	return issues, nil
}

// RBACServiceAccountAnalyzer checks ServiceAccount security
type RBACServiceAccountAnalyzer struct {
	*analyzer.BaseAnalyzer
}

// NewRBACServiceAccountAnalyzer creates a new RBAC ServiceAccount analyzer
func NewRBACServiceAccountAnalyzer() *RBACServiceAccountAnalyzer {
	return &RBACServiceAccountAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer(
			"security.rbac.serviceaccount",
			"检查 RBAC ServiceAccount 安全",
			"security",
			[]types.DataMode{types.ModeOnline},
		),
	}
}

// Analyze performs RBAC ServiceAccount checks
func (a *RBACServiceAccountAnalyzer) Analyze(ctx context.Context, data *types.DiagnosticData) ([]types.Issue, error) {
	var issues []types.Issue

	if !data.HasK8sClient() {
		return issues, nil
	}

	client := data.K8sClient

	// Check default ServiceAccounts
	namespaces, err := client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return issues, nil
	}

	for _, ns := range namespaces.Items {
		sa, err := client.CoreV1().ServiceAccounts(ns.Name).Get(ctx, "default", metav1.GetOptions{})
		if err != nil {
			continue
		}

		// Check if default SA has automountServiceAccountToken enabled
		if sa.AutomountServiceAccountToken == nil || *sa.AutomountServiceAccountToken {
			issue := types.NewIssue(
				types.SeverityInfo,
				fmt.Sprintf("命名空间 %s 的 default ServiceAccount 自动挂载 Token", ns.Name),
				"RBAC_DEFAULT_SA_AUTOMOUNT",
				fmt.Sprintf("命名空间 %s 的 default ServiceAccount 启用了自动挂载，可能导致安全风险", ns.Name),
				fmt.Sprintf("%s/default", ns.Name),
			).WithRemediation("设置 automountServiceAccountToken: false")
			issue.AnalyzerName = a.Name()
			issues = append(issues, *issue)
		}
	}

	return issues, nil
}

// RBACDangerousPermissionsAnalyzer checks for dangerous RBAC permissions
type RBACDangerousPermissionsAnalyzer struct {
	*analyzer.BaseAnalyzer
}

// NewRBACDangerousPermissionsAnalyzer creates a new RBAC dangerous permissions analyzer
func NewRBACDangerousPermissionsAnalyzer() *RBACDangerousPermissionsAnalyzer {
	return &RBACDangerousPermissionsAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer(
			"security.rbac.dangerous",
			"检查危险 RBAC 权限",
			"security",
			[]types.DataMode{types.ModeOnline},
		),
	}
}

// Analyze performs RBAC dangerous permissions checks
func (a *RBACDangerousPermissionsAnalyzer) Analyze(ctx context.Context, data *types.DiagnosticData) ([]types.Issue, error) {
	var issues []types.Issue

	if !data.HasK8sClient() {
		return issues, nil
	}

	client := data.K8sClient

	// Check ClusterRoles for dangerous permissions
	clusterRoles, err := client.RbacV1().ClusterRoles().List(ctx, metav1.ListOptions{})
	if err != nil {
		return issues, nil
	}

	dangerousRoles := []string{}
	for _, cr := range clusterRoles.Items {
		hasEscalate := false
		hasPodsCreate := false
		hasSecretsAccess := false
		hasNodesProxy := false

		for _, rule := range cr.Rules {
			// Check for escalate verb
			for _, verb := range rule.Verbs {
				if verb == "escalate" || verb == "*" {
					hasEscalate = true
				}
				if verb == "create" || verb == "*" {
					for _, resource := range rule.Resources {
						if resource == "pods" || resource == "*" {
							hasPodsCreate = true
						}
						if resource == "secrets" || resource == "*" {
							hasSecretsAccess = true
						}
					}
				}
				if verb == "proxy" || verb == "*" {
					for _, resource := range rule.Resources {
						if resource == "nodes" || resource == "*" {
							hasNodesProxy = true
						}
					}
				}
			}
		}

		if hasEscalate || (hasPodsCreate && cr.Name != "system:aggregate-to-edit") || hasSecretsAccess || hasNodesProxy {
			dangerousRoles = append(dangerousRoles, cr.Name)
		}
	}

	if len(dangerousRoles) > 0 {
		issue := types.NewIssue(
			types.SeverityWarning,
			fmt.Sprintf("发现 %d 个具有危险权限的 ClusterRole", len(dangerousRoles)),
			"RBAC_DANGEROUS_PERMISSIONS",
			fmt.Sprintf("ClusterRoles 具有危险权限 (escalate/create pods/secrets/nodes proxy): %v", dangerousRoles),
			"cluster/rbac",
		).WithRemediation("审查这些 ClusterRole，移除不必要的危险权限")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	}

	return issues, nil
}

// RBACUnusedRolesAnalyzer checks for unused Roles and ClusterRoles
type RBACUnusedRolesAnalyzer struct {
	*analyzer.BaseAnalyzer
}

// NewRBACUnusedRolesAnalyzer creates a new RBAC unused roles analyzer
func NewRBACUnusedRolesAnalyzer() *RBACUnusedRolesAnalyzer {
	return &RBACUnusedRolesAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer(
			"security.rbac.unused",
			"检查未使用的 RBAC 角色",
			"security",
			[]types.DataMode{types.ModeOnline},
		),
	}
}

// Analyze performs RBAC unused roles checks
func (a *RBACUnusedRolesAnalyzer) Analyze(ctx context.Context, data *types.DiagnosticData) ([]types.Issue, error) {
	var issues []types.Issue

	if !data.HasK8sClient() {
		return issues, nil
	}

	client := data.K8sClient

	// Get all ClusterRoles
	clusterRoles, err := client.RbacV1().ClusterRoles().List(ctx, metav1.ListOptions{})
	if err != nil {
		return issues, nil
	}

	// Get all ClusterRoleBindings
	crbs, err := client.RbacV1().ClusterRoleBindings().List(ctx, metav1.ListOptions{})
	if err != nil {
		return issues, nil
	}

	// Find used ClusterRoles
	usedClusterRoles := make(map[string]bool)
	for _, crb := range crbs.Items {
		usedClusterRoles[crb.RoleRef.Name] = true
	}

	// Check for unused ClusterRoles (excluding system roles)
	unusedRoles := []string{}
	systemPrefixes := []string{"system:", "kubeadm:", "eks:", "gke:", "aks:"}
	for _, cr := range clusterRoles.Items {
		isSystem := false
		for _, prefix := range systemPrefixes {
			if len(cr.Name) >= len(prefix) && cr.Name[:len(prefix)] == prefix {
				isSystem = true
				break
			}
		}
		if !isSystem && !usedClusterRoles[cr.Name] {
			unusedRoles = append(unusedRoles, cr.Name)
		}
	}

	if len(unusedRoles) > 10 {
		issue := types.NewIssue(
			types.SeverityInfo,
			fmt.Sprintf("发现 %d 个未使用的 ClusterRole", len(unusedRoles)),
			"RBAC_UNUSED_CLUSTER_ROLES",
			"大量 ClusterRole 未被绑定，建议清理",
			"cluster/rbac",
		).WithRemediation("删除未使用的 ClusterRole 以保持集群整洁")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	}

	return issues, nil
}

// RBACTokenAnalyzer checks for ServiceAccount tokens
type RBACTokenAnalyzer struct {
	*analyzer.BaseAnalyzer
}

// NewRBACTokenAnalyzer creates a new RBAC token analyzer
func NewRBACTokenAnalyzer() *RBACTokenAnalyzer {
	return &RBACTokenAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer(
			"security.rbac.token",
			"检查 RBAC Token 安全",
			"security",
			[]types.DataMode{types.ModeOnline},
		),
	}
}

// Analyze performs RBAC token checks
func (a *RBACTokenAnalyzer) Analyze(ctx context.Context, data *types.DiagnosticData) ([]types.Issue, error) {
	var issues []types.Issue

	if !data.HasK8sClient() {
		return issues, nil
	}

	client := data.K8sClient

	// Get all secrets
	secrets, err := client.CoreV1().Secrets("").List(ctx, metav1.ListOptions{
		FieldSelector: "type=kubernetes.io/service-account-token",
	})
	if err != nil {
		return issues, nil
	}

	// Count tokens per ServiceAccount
	saTokenCount := make(map[string]int)
	for _, secret := range secrets.Items {
		if secret.Annotations != nil {
			if saName, ok := secret.Annotations["kubernetes.io/service-account.name"]; ok {
				key := fmt.Sprintf("%s/%s", secret.Namespace, saName)
				saTokenCount[key]++
			}
		}
	}

	// Check for multiple tokens per SA
	multiTokenSAs := []string{}
	for sa, count := range saTokenCount {
		if count > 3 {
			multiTokenSAs = append(multiTokenSAs, fmt.Sprintf("%s (%d tokens)", sa, count))
		}
	}

	if len(multiTokenSAs) > 0 {
		issue := types.NewIssue(
			types.SeverityInfo,
			fmt.Sprintf("发现 %d 个 ServiceAccount 有过期 Token", len(multiTokenSAs)),
			"RBAC_MULTIPLE_TOKENS",
			fmt.Sprintf("ServiceAccounts 拥有多个 Token，可能有旧 Token 未清理: %v", multiTokenSAs),
			"cluster/secrets",
		).WithRemediation("清理过期的 ServiceAccount Token")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	}

	return issues, nil
}

func init() {
	analyzer.MustRegister(NewRBACAdminAnalyzer())
	analyzer.MustRegister(NewRBACServiceAccountAnalyzer())
	analyzer.MustRegister(NewRBACDangerousPermissionsAnalyzer())
	analyzer.MustRegister(NewRBACUnusedRolesAnalyzer())
	analyzer.MustRegister(NewRBACTokenAnalyzer())
}
