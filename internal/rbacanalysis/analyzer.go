package rbacanalysis

import (
	"context"
	"fmt"
	"strings"
	"time"

	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Analyzer performs RBAC analysis against a Kubernetes cluster.
type Analyzer struct {
	clientset kubernetes.Interface
}

// NewAnalyzer creates a new RBAC Analyzer.
func NewAnalyzer(clientset kubernetes.Interface) *Analyzer {
	return &Analyzer{clientset: clientset}
}

type SubjectBinding struct {
	Type      string `json:"type"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Role      string `json:"role"`
	RoleKind  string `json:"roleKind"`
}

type RoleBindingInfo struct {
	Type      string   `json:"type"`
	Name      string   `json:"name"`
	Namespace string   `json:"namespace"`
	Subjects  []string `json:"subjects"`
}

type RBACAnalysis struct {
	TotalRoles           int                             `json:"totalRoles"`
	TotalClusterRoles    int                             `json:"totalClusterRoles"`
	TotalBindings        int                             `json:"totalBindings"`
	TotalClusterBindings int                             `json:"totalClusterBindings"`
	RolesByNamespace     map[string][]string             `json:"rolesByNamespace"`
	BindingsBySubject    map[string][]SubjectBinding     `json:"bindingsBySubject"`
	BindingsByRole       map[string][]RoleBindingInfo    `json:"bindingsByRole"`
	Timestamp            time.Time                       `json:"timestamp"`
}

func (a *Analyzer) AnalyzeRBAC(ctx context.Context) (*RBACAnalysis, error) {
	roles, err := a.clientset.RbacV1().Roles("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	clusterRoles, err := a.clientset.RbacV1().ClusterRoles().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	roleBindings, err := a.clientset.RbacV1().RoleBindings("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	clusterRoleBindings, err := a.clientset.RbacV1().ClusterRoleBindings().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	analysis := &RBACAnalysis{
		TotalRoles:           len(roles.Items),
		TotalClusterRoles:    len(clusterRoles.Items),
		TotalBindings:        len(roleBindings.Items),
		TotalClusterBindings: len(clusterRoleBindings.Items),
		RolesByNamespace:     make(map[string][]string),
		BindingsBySubject:    make(map[string][]SubjectBinding),
		BindingsByRole:       make(map[string][]RoleBindingInfo),
		Timestamp:            time.Now(),
	}

	for _, r := range roles.Items {
		analysis.RolesByNamespace[r.Namespace] = append(analysis.RolesByNamespace[r.Namespace], r.Name)
	}

	for _, rb := range roleBindings.Items {
		roleName := rb.RoleRef.Name
		roleKind := rb.RoleRef.Kind
		bindingKey := fmt.Sprintf("%s/%s", rb.Namespace, roleName)

		info := RoleBindingInfo{
			Type:      "RoleBinding",
			Name:      rb.Name,
			Namespace: rb.Namespace,
			Subjects:  make([]string, 0, len(rb.Subjects)),
		}

		for _, s := range rb.Subjects {
			subjectKey := formatSubjectKey(s)
			info.Subjects = append(info.Subjects, subjectKey)
			analysis.BindingsBySubject[subjectKey] = append(analysis.BindingsBySubject[subjectKey], SubjectBinding{
				Type:      "RoleBinding",
				Name:      rb.Name,
				Namespace: rb.Namespace,
				Role:      roleName,
				RoleKind:  roleKind,
			})
		}

		analysis.BindingsByRole[bindingKey] = append(analysis.BindingsByRole[bindingKey], info)
	}

	for _, crb := range clusterRoleBindings.Items {
		roleName := crb.RoleRef.Name
		roleKind := crb.RoleRef.Kind
		bindingKey := fmt.Sprintf("(cluster)/%s", roleName)

		info := RoleBindingInfo{
			Type:      "ClusterRoleBinding",
			Name:      crb.Name,
			Namespace: "",
			Subjects:  make([]string, 0, len(crb.Subjects)),
		}

		for _, s := range crb.Subjects {
			subjectKey := formatSubjectKey(s)
			info.Subjects = append(info.Subjects, subjectKey)
			analysis.BindingsBySubject[subjectKey] = append(analysis.BindingsBySubject[subjectKey], SubjectBinding{
				Type:      "ClusterRoleBinding",
				Name:      crb.Name,
				Namespace: "",
				Role:      roleName,
				RoleKind:  roleKind,
			})
		}

		analysis.BindingsByRole[bindingKey] = append(analysis.BindingsByRole[bindingKey], info)
	}

	return analysis, nil
}

func formatSubjectKey(subject rbacv1.Subject) string {
	return fmt.Sprintf("%s:%s:%s", subject.Kind, subject.Namespace, subject.Name)
}

func (a *Analyzer) ListRoles(ctx context.Context, namespace string) ([]rbacv1.Role, error) {
	items, err := a.clientset.RbacV1().Roles(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return items.Items, nil
}

func (a *Analyzer) ListClusterRoles(ctx context.Context) ([]rbacv1.ClusterRole, error) {
	items, err := a.clientset.RbacV1().ClusterRoles().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return items.Items, nil
}

func (a *Analyzer) ClusterRolesByPrefix(ctx context.Context, prefix string) ([]rbacv1.ClusterRole, error) {
	all, err := a.ListClusterRoles(ctx)
	if err != nil {
		return nil, err
	}
	var matched []rbacv1.ClusterRole
	for _, cr := range all {
		if strings.HasPrefix(cr.Name, prefix) {
			matched = append(matched, cr)
		}
	}
	return matched, nil
}
