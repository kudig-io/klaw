package ops

import (
	"fmt"
	"strings"

	"github.com/kudig-io/klaw/internal/loganalysis"
)

// handler_pod.go Pod 命令实现：列表、详情、日志、日志分析与删除。

// listPods 列出Pod
func (h *Handler) listPods(clusterName, namespace string) (string, error) {
	if err := h.requireKubernetes(); err != nil {
		return "", err
	}

	pods, err := h.resources.ListPods(clusterName, namespace)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## Pods in `%s`\n\n", namespace))
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

// describePod 描述Pod
func (h *Handler) describePod(clusterName, namespace, podName string) (string, error) {
	if err := h.requireKubernetes(); err != nil {
		return "", err
	}

	pod, err := h.resources.GetPod(clusterName, namespace, podName)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## Pod: %s\n\n", pod.Name))
	sb.WriteString("| Property | Value |\n")
	sb.WriteString("|----------|-------|\n")
	sb.WriteString(fmt.Sprintf("| Namespace | %s |\n", pod.Namespace))
	sb.WriteString(fmt.Sprintf("| Status | %s |\n", pod.Status.Phase))
	sb.WriteString(fmt.Sprintf("| Node | %s |\n", pod.Spec.NodeName))
	sb.WriteString(fmt.Sprintf("| Created | %s |\n", pod.CreationTimestamp.Format("2006-01-02 15:04:05")))

	return sb.String(), nil
}

// getPodLogs 获取Pod日志
func (h *Handler) getPodLogs(clusterName, namespace, podName string) (string, error) {
	if err := h.requireKubernetes(); err != nil {
		return "", err
	}

	logs, err := h.resources.GetPodLogs(clusterName, namespace, podName, 100)
	if err != nil {
		return "", err
	}

	// 四反引号围栏：日志正文含三反引号时不会破坏 Markdown 结构
	return fmt.Sprintf("## Logs from `%s`\n\n````text\n%s\n````", podName, logs), nil
}

// analyzePodLogs 分析 Pod 日志
func (h *Handler) analyzePodLogs(clusterName, namespace, podName string) (string, error) {
	if err := h.requireKubernetes(); err != nil {
		return "", err
	}

	logs, err := h.resources.GetPodLogs(clusterName, namespace, podName, 200)
	if err != nil {
		return "", err
	}

	analysis := loganalysis.NewAnalyzer().AnalyzeLogs(logs)
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## Log Analysis: `%s`\n\n", podName))
	sb.WriteString("| Metric | Value |\n")
	sb.WriteString("|--------|-------|\n")
	sb.WriteString(fmt.Sprintf("| Total lines | %d |\n", analysis.TotalLines))
	sb.WriteString(fmt.Sprintf("| Errors | %d |\n", analysis.ErrorCount))
	sb.WriteString(fmt.Sprintf("| Warnings | %d |\n", analysis.WarningCount))
	sb.WriteString(fmt.Sprintf("| Security events | %d |\n", len(analysis.SecurityEvents)))
	sb.WriteString(fmt.Sprintf("| Slow requests | %d |\n", len(analysis.PerformanceMetrics.SlowRequests)))

	return sb.String(), nil
}

// deletePod 删除Pod（破坏性操作，需 server.ops.allow_destructive 开启）
func (h *Handler) deletePod(clusterName, namespace, podName string) (string, error) {
	if !h.allowDestructive {
		return "", fmt.Errorf("破坏性操作已禁用：删除 Pod 需在配置 server.ops.allow_destructive: true 后重试")
	}
	if err := h.requireKubernetes(); err != nil {
		return "", err
	}

	err := h.resources.DeletePod(clusterName, namespace, podName)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("**Deleted** pod `%s` in namespace `%s`", podName, namespace), nil
}
