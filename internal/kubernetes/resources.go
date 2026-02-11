package kubernetes

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Resources Kubernetes资源管理
type Resources struct {
	manager *Manager
}

// NewResources 创建资源管理器
func NewResources(manager *Manager) *Resources {
	return &Resources{manager: manager}
}

// ListPods 列出Pod
func (r *Resources) ListPods(clusterName, namespace string) ([]corev1.Pod, error) {
	client, err := r.manager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}

	pods, err := client.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %v", err)
	}

	return pods.Items, nil
}

// GetPod 获取Pod详情
func (r *Resources) GetPod(clusterName, namespace, podName string) (*corev1.Pod, error) {
	client, err := r.manager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}

	pod, err := client.CoreV1().Pods(namespace).Get(context.Background(), podName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get pod: %v", err)
	}

	return pod, nil
}

// DeletePod 删除Pod
func (r *Resources) DeletePod(clusterName, namespace, podName string) error {
	client, err := r.manager.GetClient(clusterName)
	if err != nil {
		return err
	}

	err = client.CoreV1().Pods(namespace).Delete(context.Background(), podName, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete pod: %v", err)
	}

	return nil
}

// ListNodes 列出节点
func (r *Resources) ListNodes(clusterName string) ([]corev1.Node, error) {
	client, err := r.manager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}

	nodes, err := client.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %v", err)
	}

	return nodes.Items, nil
}

// GetNode 获取节点详情
func (r *Resources) GetNode(clusterName, nodeName string) (*corev1.Node, error) {
	client, err := r.manager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}

	node, err := client.CoreV1().Nodes().Get(context.Background(), nodeName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get node: %v", err)
	}

	return node, nil
}

// ListNamespaces 列出命名空间
func (r *Resources) ListNamespaces(clusterName string) ([]corev1.Namespace, error) {
	client, err := r.manager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}

	namespaces, err := client.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %v", err)
	}

	return namespaces.Items, nil
}

// ListEvents 列出事件
func (r *Resources) ListEvents(clusterName, namespace string) ([]corev1.Event, error) {
	client, err := r.manager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}

	events, err := client.CoreV1().Events(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list events: %v", err)
	}

	return events.Items, nil
}

// GetPodLogs 获取Pod日志
func (r *Resources) GetPodLogs(clusterName, namespace, podName string, tailLines int64) (string, error) {
	client, err := r.manager.GetClient(clusterName)
	if err != nil {
		return "", err
	}

	req := client.CoreV1().Pods(namespace).GetLogs(podName, &corev1.PodLogOptions{
		TailLines: &tailLines,
	})

	logs, err := req.Stream(context.Background())
	if err != nil {
		return "", fmt.Errorf("failed to get pod logs: %v", err)
	}
	defer logs.Close()

	buf := make([]byte, 1024)
	var result string
	for {
		n, err := logs.Read(buf)
		if err != nil {
			break
		}
		result += string(buf[:n])
	}

	return result, nil
}

// GetNodeMetrics 获取节点指标
func (r *Resources) GetNodeMetrics(clusterName string) (map[string]NodeMetrics, error) {
	client, err := r.manager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}

	nodes, err := client.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	metrics := make(map[string]NodeMetrics)
	for _, node := range nodes.Items {
		metrics[node.Name] = NodeMetrics{
			Name:       node.Name,
			CPU:        node.Status.Capacity.Cpu().String(),
			Memory:     node.Status.Capacity.Memory().String(),
			Conditions: node.Status.Conditions,
		}
	}

	return metrics, nil
}

// NodeMetrics 节点指标
type NodeMetrics struct {
	Name       string
	CPU        string
	Memory     string
	Conditions []corev1.NodeCondition
}
