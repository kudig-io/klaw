package ops

import (
	"fmt"
	"strings"
)

// handler_deployment.go Deployment 命令实现：列表、状态、扩缩容、重启与关联 Pods。

// listDeployments 列出 Deployment
func (h *Handler) listDeployments(clusterName, namespace string) (string, error) {
	if err := h.requireKubernetes(); err != nil {
		return "", err
	}

	deployments, err := h.resources.ListDeployments(clusterName, namespace)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## Deployments in `%s`\n\n", namespace))
	sb.WriteString("| Name | Ready | Up-to-date | Available | Age |\n")
	sb.WriteString("|------|-------|------------|-----------|-----|\n")
	for _, dep := range deployments {
		ready := dep.Status.ReadyReplicas
		var desired int32
		if dep.Spec.Replicas != nil {
			desired = *dep.Spec.Replicas
		}
		age := formatAge(dep.CreationTimestamp.Time)
		sb.WriteString(fmt.Sprintf("| %s | %d/%d | %d | %d | %s |\n",
			dep.Name, ready, desired, dep.Status.UpdatedReplicas, dep.Status.AvailableReplicas, age))
	}
	if len(deployments) == 0 {
		sb.WriteString("| *(none)* | | | | |\n")
	}

	return sb.String(), nil
}

// getDeploymentStatus 获取 Deployment 状态
func (h *Handler) getDeploymentStatus(clusterName, namespace, deploymentName string) (string, error) {
	if err := h.requireKubernetes(); err != nil {
		return "", err
	}

	status, err := h.resources.GetDeploymentStatus(clusterName, namespace, deploymentName)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## Deployment: %s\n\n", status.Name))
	sb.WriteString("| Property | Value |\n")
	sb.WriteString("|----------|-------|\n")
	sb.WriteString(fmt.Sprintf("| Namespace | %s |\n", status.Namespace))
	sb.WriteString(fmt.Sprintf("| Replicas | %d (available: %d, ready: %d, updated: %d) |\n",
		status.Replicas, status.AvailableReplicas, status.ReadyReplicas, status.UpdatedReplicas))

	if len(status.Conditions) > 0 {
		sb.WriteString("\n### Conditions\n\n")
		sb.WriteString("| Type | Status | Reason |\n")
		sb.WriteString("|------|--------|--------|\n")
		for _, condition := range status.Conditions {
			reason := condition.Reason
			if reason == "" {
				reason = "-"
			}
			sb.WriteString(fmt.Sprintf("| %s | %s | %s |\n", condition.Type, condition.Status, reason))
		}
	}

	return sb.String(), nil
}

// scaleDeployment 扩缩容 Deployment
func (h *Handler) scaleDeployment(clusterName, namespace, deploymentName string, replicas int32) (string, error) {
	if err := h.requireKubernetes(); err != nil {
		return "", err
	}

	err := h.resources.ScaleDeployment(clusterName, namespace, deploymentName, replicas)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("**Scaled** deployment `%s` in `%s` to **%d** replicas", deploymentName, namespace, replicas), nil
}

// restartDeployment 重启 Deployment
func (h *Handler) restartDeployment(clusterName, namespace, deploymentName string) (string, error) {
	if err := h.requireKubernetes(); err != nil {
		return "", err
	}

	err := h.resources.RestartDeployment(clusterName, namespace, deploymentName)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("**Restarted** deployment `%s` in namespace `%s`", deploymentName, namespace), nil
}

// getDeploymentPods 获取 Deployment 关联的 Pods
func (h *Handler) getDeploymentPods(clusterName, namespace, deploymentName string) (string, error) {
	if err := h.requireKubernetes(); err != nil {
		return "", err
	}

	pods, err := h.resources.GetDeploymentPods(clusterName, namespace, deploymentName)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## Pods for deployment `%s`\n\n", deploymentName))
	sb.WriteString("| Name | Status | Node |\n")
	sb.WriteString("|------|--------|------|\n")
	for _, pod := range pods {
		sb.WriteString(fmt.Sprintf("| %s | %s | %s |\n", pod.Name, pod.Status.Phase, pod.Spec.NodeName))
	}
	if len(pods) == 0 {
		sb.WriteString("| *(none)* | | |\n")
	}

	return sb.String(), nil
}
