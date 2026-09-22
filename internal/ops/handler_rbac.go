package ops

import (
	"context"
	"fmt"
	"strings"

	"github.com/kudig-io/klaw/internal/rbacanalysis"
)

// handler_rbac.go RBAC 分析命令实现。

// analyzeRBAC 分析集群 RBAC
func (h *Handler) analyzeRBAC(clusterName string) (string, error) {
	if err := h.requireKubernetes(); err != nil {
		return "", err
	}

	client, err := h.k8sManager.GetClient(clusterName)
	if err != nil {
		return "", err
	}

	analysis, err := rbacanalysis.NewAnalyzer(client).AnalyzeRBAC(context.Background())
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## RBAC Analysis: `%s`\n\n", clusterName))
	sb.WriteString("| Resource | Count |\n")
	sb.WriteString("|----------|-------|\n")
	sb.WriteString(fmt.Sprintf("| Roles | %d |\n", analysis.TotalRoles))
	sb.WriteString(fmt.Sprintf("| ClusterRoles | %d |\n", analysis.TotalClusterRoles))
	sb.WriteString(fmt.Sprintf("| RoleBindings | %d |\n", analysis.TotalBindings))
	sb.WriteString(fmt.Sprintf("| ClusterRoleBindings | %d |\n", analysis.TotalClusterBindings))

	return sb.String(), nil
}
