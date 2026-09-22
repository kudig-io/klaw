package ops

import (
	"fmt"
	"strings"
)

// handler_node.go 节点命令实现：列表、详情与节点指标。

// listNodes 列出节点
func (h *Handler) listNodes(clusterName string) (string, error) {
	if err := h.requireKubernetes(); err != nil {
		return "", err
	}

	nodes, err := h.resources.ListNodes(clusterName)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## Nodes in `%s`\n\n", clusterName))
	sb.WriteString("| Name | Status | Roles | Version |\n")
	sb.WriteString("|------|--------|-------|---------|\n")
	for _, node := range nodes {
		status := "NotReady"
		for _, cond := range node.Status.Conditions {
			if cond.Type == "Ready" && cond.Status == "True" {
				status = "Ready"
				break
			}
		}
		roles := []string{}
		for label := range node.Labels {
			if strings.HasPrefix(label, "node-role.kubernetes.io/") {
				role := strings.TrimPrefix(label, "node-role.kubernetes.io/")
				if role != "" {
					roles = append(roles, role)
				}
			}
		}
		if len(roles) == 0 {
			roles = append(roles, "<none>")
		}
		version := node.Status.NodeInfo.KubeletVersion
		sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n", node.Name, status, strings.Join(roles, ","), version))
	}
	if len(nodes) == 0 {
		sb.WriteString("| *(none)* | | | |\n")
	}

	return sb.String(), nil
}

// describeNode 描述节点
func (h *Handler) describeNode(clusterName, nodeName string) (string, error) {
	if err := h.requireKubernetes(); err != nil {
		return "", err
	}

	node, err := h.resources.GetNode(clusterName, nodeName)
	if err != nil {
		return "", err
	}

	status := "NotReady"
	for _, cond := range node.Status.Conditions {
		if cond.Type == "Ready" && cond.Status == "True" {
			status = "Ready"
			break
		}
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## Node: %s\n\n", node.Name))
	sb.WriteString("| Property | Value |\n")
	sb.WriteString("|----------|-------|\n")
	sb.WriteString(fmt.Sprintf("| Status | %s |\n", status))
	sb.WriteString(fmt.Sprintf("| CPU | %s |\n", node.Status.Capacity.Cpu().String()))
	sb.WriteString(fmt.Sprintf("| Memory | %s |\n", node.Status.Capacity.Memory().String()))
	sb.WriteString(fmt.Sprintf("| OS | %s %s |\n", node.Status.NodeInfo.OperatingSystem, node.Status.NodeInfo.OSImage))
	sb.WriteString(fmt.Sprintf("| Kernel | %s |\n", node.Status.NodeInfo.KernelVersion))
	sb.WriteString(fmt.Sprintf("| Container Runtime | %s |\n", node.Status.NodeInfo.ContainerRuntimeVersion))

	return sb.String(), nil
}

// getNodeMetrics 获取节点指标
func (h *Handler) getNodeMetrics(clusterName string) (string, error) {
	if err := h.requireKubernetes(); err != nil {
		return "", err
	}

	nodeMetrics, err := h.resources.GetNodeMetrics(clusterName)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## Node Metrics: `%s`\n\n", clusterName))
	sb.WriteString("| Node | CPU | Memory |\n")
	sb.WriteString("|------|-----|--------|\n")
	for nodeName, nm := range nodeMetrics {
		sb.WriteString(fmt.Sprintf("| %s | %s | %s |\n", nodeName, nm.CPU, nm.Memory))
	}
	if len(nodeMetrics) == 0 {
		sb.WriteString("| *(none)* | | |\n")
	}

	return sb.String(), nil
}
