package automation

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (m *Manager) executeBuiltin(ctx context.Context, name string, params map[string]interface{}) (string, error) {
	if m.clientset == nil {
		return "", fmt.Errorf("kubernetes client not configured")
	}

	switch name {
	case "cleanup-evicted-pods":
		return m.builtinCleanupEvictedPods(ctx, params)
	case "restart-crashing-pods":
		return m.builtinRestartCrashingPods(ctx, params)
	case "scale-deployment":
		return m.builtinScaleDeployment(ctx, params)
	case "backup-configmaps":
		return m.builtinBackupConfigMaps(ctx, params)
	case "check-node-health":
		return m.builtinCheckNodeHealth(ctx, params)
	case "update-image-tags":
		return m.builtinUpdateImageTags(ctx, params)
	case "cleanup-old-images":
		return "", fmt.Errorf("cleanup-old-images requires node-level access (not implemented via API)")
	case "rotate-logs":
		return "", fmt.Errorf("rotate-logs requires node-level access (not implemented via API)")
	default:
		return "", fmt.Errorf("unknown builtin script: %s", name)
	}
}

func (m *Manager) builtinCleanupEvictedPods(ctx context.Context, params map[string]interface{}) (string, error) {
	dryRun := params["dryRun"] == true
	namespaces := namespacesParam(params)

	list, err := m.clientset.CoreV1().Pods(namespaces).List(ctx, metav1.ListOptions{})
	if err != nil {
		return "", fmt.Errorf("list pods: %w", err)
	}

	var cleaned int
	var details []string
	for _, pod := range list.Items {
		if pod.Status.Phase != corev1.PodFailed {
			continue
		}
		if pod.Status.Reason != "Evicted" {
			continue
		}
		if dryRun {
			details = append(details, fmt.Sprintf("[dry-run] would delete %s/%s", pod.Namespace, pod.Name))
			cleaned++
			continue
		}
		if err := m.clientset.CoreV1().Pods(pod.Namespace).Delete(ctx, pod.Name, metav1.DeleteOptions{}); err != nil {
			details = append(details, fmt.Sprintf("ERROR deleting %s/%s: %v", pod.Namespace, pod.Name, err))
			continue
		}
		details = append(details, fmt.Sprintf("deleted %s/%s", pod.Namespace, pod.Name))
		cleaned++
	}

	return fmt.Sprintf("Cleaned up %d evicted pods (dryRun=%v)\n%s", cleaned, dryRun, strings.Join(details, "\n")), nil
}

func (m *Manager) builtinRestartCrashingPods(ctx context.Context, params map[string]interface{}) (string, error) {
	threshold := intParam(params, "restartThreshold", 5)
	namespaces := namespacesParam(params)

	list, err := m.clientset.CoreV1().Pods(namespaces).List(ctx, metav1.ListOptions{})
	if err != nil {
		return "", fmt.Errorf("list pods: %w", err)
	}

	var restarted int
	var details []string
	for _, pod := range list.Items {
		for _, cs := range pod.Status.ContainerStatuses {
			if int(cs.RestartCount) >= threshold {
				if err := m.clientset.CoreV1().Pods(pod.Namespace).Delete(ctx, pod.Name, metav1.DeleteOptions{}); err != nil {
					details = append(details, fmt.Sprintf("ERROR deleting %s/%s: %v", pod.Namespace, pod.Name, err))
					continue
				}
				details = append(details, fmt.Sprintf("restarted %s/%s (restarts=%d)", pod.Namespace, pod.Name, cs.RestartCount))
				restarted++
				break
			}
		}
	}

	return fmt.Sprintf("Restarted %d crashing pods (threshold=%d)\n%s", restarted, threshold, strings.Join(details, "\n")), nil
}

func (m *Manager) builtinScaleDeployment(ctx context.Context, params map[string]interface{}) (string, error) {
	cpuThreshold := intParam(params, "cpuThreshold", 80)
	maxReplicas := int32(intParam(params, "maxReplicas", 10))
	namespaces := namespacesParam(params)

	deploys, err := m.clientset.AppsV1().Deployments(namespaces).List(ctx, metav1.ListOptions{})
	if err != nil {
		return "", fmt.Errorf("list deployments: %w", err)
	}

	var scaled int
	var details []string
	for i := range deploys.Items {
		d := &deploys.Items[i]
		if d.Spec.Replicas == nil || *d.Spec.Replicas >= maxReplicas {
			continue
		}
		requested := 0
		for _, c := range d.Spec.Template.Spec.Containers {
			if req, ok := c.Resources.Requests[corev1.ResourceCPU]; ok {
				requested += int(req.MilliValue())
			}
		}
		if requested > cpuThreshold*1000 {
			newReplicas := *d.Spec.Replicas + 1
			d.Spec.Replicas = &newReplicas
			if _, err := m.clientset.AppsV1().Deployments(d.Namespace).Update(ctx, d, metav1.UpdateOptions{}); err != nil {
				details = append(details, fmt.Sprintf("ERROR scaling %s/%s: %v", d.Namespace, d.Name, err))
				continue
			}
			details = append(details, fmt.Sprintf("scaled %s/%s %d→%d", d.Namespace, d.Name, newReplicas-1, newReplicas))
			scaled++
		}
	}

	return fmt.Sprintf("Scaled %d deployments (cpuThreshold=%d%%, max=%d)\n%s", scaled, cpuThreshold, maxReplicas, strings.Join(details, "\n")), nil
}

func (m *Manager) builtinBackupConfigMaps(ctx context.Context, params map[string]interface{}) (string, error) {
	backupPath := stringParam(params, "backupPath", "/tmp/klaw-backup/configmaps")
	namespaces := namespacesParam(params)

	list, err := m.clientset.CoreV1().ConfigMaps(namespaces).List(ctx, metav1.ListOptions{})
	if err != nil {
		return "", fmt.Errorf("list configmaps: %w", err)
	}

	if err := os.MkdirAll(backupPath, 0o755); err != nil {
		return "", fmt.Errorf("create backup dir: %w", err)
	}

	var backed int
	for _, cm := range list.Items {
		data, _ := json.MarshalIndent(cm, "", "  ")
		filename := filepath.Join(backupPath, fmt.Sprintf("%s_%s.json", cm.Namespace, cm.Name))
		if err := os.WriteFile(filename, data, 0o644); err != nil {
			continue
		}
		backed++
	}

	return fmt.Sprintf("Backed up %d ConfigMaps to %s", backed, backupPath), nil
}

func (m *Manager) builtinCheckNodeHealth(ctx context.Context, params map[string]interface{}) (string, error) {
	alertNotReady := params["alertOnNotReady"] != false
	alertPressure := params["alertOnPressure"] != false

	nodes, err := m.clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return "", fmt.Errorf("list nodes: %w", err)
	}

	var healthy, notReady int
	var issues []string
	for _, node := range nodes.Items {
		ready := false
		for _, cond := range node.Status.Conditions {
			if cond.Type == corev1.NodeReady {
				ready = cond.Status == corev1.ConditionTrue
			}
			if alertPressure && strings.HasSuffix(string(cond.Type), "Pressure") && cond.Status == corev1.ConditionTrue {
				issues = append(issues, fmt.Sprintf("node %s has %s", node.Name, cond.Type))
			}
		}
		if ready {
			healthy++
		} else {
			notReady++
			if alertNotReady {
				issues = append(issues, fmt.Sprintf("node %s is NOT READY", node.Name))
			}
		}
	}

	summary := fmt.Sprintf("Nodes: %d healthy, %d not ready, %d total", healthy, notReady, len(nodes.Items))
	if len(issues) > 0 {
		summary += "\nIssues:\n" + strings.Join(issues, "\n")
	}
	return summary, nil
}

func (m *Manager) builtinUpdateImageTags(ctx context.Context, params map[string]interface{}) (string, error) {
	newTag := stringParam(params, "newTag", "latest")
	namespaces := namespacesParam(params)

	deploys, err := m.clientset.AppsV1().Deployments(namespaces).List(ctx, metav1.ListOptions{})
	if err != nil {
		return "", fmt.Errorf("list deployments: %w", err)
	}

	var updated int
	var details []string
	for _, d := range deploys.Items {
		changed := false
		for i, c := range d.Spec.Template.Spec.Containers {
			parts := strings.SplitN(c.Image, ":", 2)
			if len(parts) < 2 {
				continue
			}
			newImage := parts[0] + ":" + newTag
			if c.Image != newImage {
				d.Spec.Template.Spec.Containers[i].Image = newImage
				changed = true
				details = append(details, fmt.Sprintf("%s/%s %s: %s → %s", d.Namespace, d.Name, c.Name, c.Image, newImage))
			}
		}
		if changed {
			if _, err := m.clientset.AppsV1().Deployments(d.Namespace).Update(ctx, &d, metav1.UpdateOptions{}); err != nil {
				details = append(details, fmt.Sprintf("ERROR updating %s/%s: %v", d.Namespace, d.Name, err))
				continue
			}
			updated++
		}
	}

	return fmt.Sprintf("Updated %d deployments to tag '%s'\n%s", updated, newTag, strings.Join(details, "\n")), nil
}

func namespacesParam(params map[string]interface{}) string {
	if ns, ok := params["namespaces"]; ok {
		if arr, ok := ns.([]interface{}); ok && len(arr) > 0 {
			return arr[0].(string)
		}
	}
	return ""
}

func intParam(params map[string]interface{}, key string, def int) int {
	if v, ok := params[key]; ok {
		switch n := v.(type) {
		case float64:
			return int(n)
		case int:
			return n
		}
	}
	return def
}

func stringParam(params map[string]interface{}, key, def string) string {
	if v, ok := params[key].(string); ok && v != "" {
		return v
	}
	return def
}
