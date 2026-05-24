package kubernetes

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// Resources Kubernetes资源管理
type Resources struct {
	manager *Manager
}

// NewResources 创建资源管理器
func NewResources(manager *Manager) *Resources {
	return &Resources{manager: manager}
}

// ========== Pod 管理 ==========

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

// ========== Node 管理 ==========

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

// ========== Namespace 管理 ==========

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

// GetNamespace 获取命名空间详情
func (r *Resources) GetNamespace(clusterName, namespace string) (*corev1.Namespace, error) {
	client, err := r.manager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}

	ns, err := client.CoreV1().Namespaces().Get(context.Background(), namespace, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get namespace: %v", err)
	}

	return ns, nil
}

// ========== ConfigMap 管理 ==========

func (r *Resources) ListConfigMaps(clusterName, namespace string) ([]corev1.ConfigMap, error) {
	client, err := r.manager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}

	configMaps, err := client.CoreV1().ConfigMaps(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list configmaps: %v", err)
	}

	return configMaps.Items, nil
}

func (r *Resources) GetConfigMap(clusterName, namespace, name string) (*corev1.ConfigMap, error) {
	client, err := r.manager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}

	configMap, err := client.CoreV1().ConfigMaps(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get configmap: %v", err)
	}

	return configMap, nil
}

// ========== StatefulSet 管理 ==========

func (r *Resources) ListStatefulSets(clusterName, namespace string) ([]appsv1.StatefulSet, error) {
	client, err := r.manager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}

	statefulSets, err := client.AppsV1().StatefulSets(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list statefulsets: %v", err)
	}

	return statefulSets.Items, nil
}

func (r *Resources) GetStatefulSet(clusterName, namespace, name string) (*appsv1.StatefulSet, error) {
	client, err := r.manager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}

	statefulSet, err := client.AppsV1().StatefulSets(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get statefulset: %v", err)
	}

	return statefulSet, nil
}

// ========== Ingress 管理 ==========

func (r *Resources) ListIngresses(clusterName, namespace string) ([]networkingv1.Ingress, error) {
	client, err := r.manager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}

	ingresses, err := client.NetworkingV1().Ingresses(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list ingresses: %v", err)
	}

	return ingresses.Items, nil
}

func (r *Resources) GetIngress(clusterName, namespace, name string) (*networkingv1.Ingress, error) {
	client, err := r.manager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}

	ingress, err := client.NetworkingV1().Ingresses(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get ingress: %v", err)
	}

	return ingress, nil
}

// ========== Event 管理 ==========

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

// ========== Pod 日志 ==========

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

// ========== Node 指标 ==========

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

// ========== Deployment 管理 ==========

// ListDeployments 列出指定命名空间的所有 Deployment
func (r *Resources) ListDeployments(clusterName, namespace string) ([]appsv1.Deployment, error) {
	client, err := r.manager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}

	deployments, err := client.AppsV1().Deployments(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list deployments: %v", err)
	}

	return deployments.Items, nil
}

// GetDeployment 获取指定 Deployment 详情
func (r *Resources) GetDeployment(clusterName, namespace, deploymentName string) (*appsv1.Deployment, error) {
	client, err := r.manager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}

	deployment, err := client.AppsV1().Deployments(namespace).Get(context.Background(), deploymentName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment: %v", err)
	}

	return deployment, nil
}

// ScaleDeployment 扩缩容 Deployment
func (r *Resources) ScaleDeployment(clusterName, namespace, deploymentName string, replicas int32) error {
	client, err := r.manager.GetClient(clusterName)
	if err != nil {
		return err
	}

	// 使用 Patch 方式更新 replicas
	patchData := fmt.Sprintf(`{"spec":{"replicas":%d}}`, replicas)
	_, err = client.AppsV1().Deployments(namespace).Patch(
		context.Background(),
		deploymentName,
		types.StrategicMergePatchType,
		[]byte(patchData),
		metav1.PatchOptions{},
	)
	if err != nil {
		return fmt.Errorf("failed to scale deployment: %v", err)
	}

	return nil
}

// RestartDeployment 重启 Deployment（通过添加注解触发滚动更新）
func (r *Resources) RestartDeployment(clusterName, namespace, deploymentName string) error {
	client, err := r.manager.GetClient(clusterName)
	if err != nil {
		return err
	}

	// 通过更新注解触发滚动更新
	deployment, err := r.GetDeployment(clusterName, namespace, deploymentName)
	if err != nil {
		return err
	}

	// 如果注解不存在，创建一个
	if deployment.Spec.Template.Annotations == nil {
		deployment.Spec.Template.Annotations = make(map[string]string)
	}
	deployment.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = metav1.Now().Format("20060102-150405")

	_, err = client.AppsV1().Deployments(namespace).Update(context.Background(), deployment, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to restart deployment: %v", err)
	}

	return nil
}

// GetDeploymentPods 获取 Deployment 关联的 Pods
func (r *Resources) GetDeploymentPods(clusterName, namespace, deploymentName string) ([]corev1.Pod, error) {
	client, err := r.manager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}

	// 获取 Deployment
	deployment, err := r.GetDeployment(clusterName, namespace, deploymentName)
	if err != nil {
		return nil, err
	}

	// 使用 Deployment 的 selector 查找 Pod
	selector := deployment.Spec.Selector
	if selector == nil {
		return nil, fmt.Errorf("deployment has no selector")
	}

	listOptions := metav1.ListOptions{
		LabelSelector: metav1.FormatLabelSelector(selector),
	}

	pods, err := client.CoreV1().Pods(namespace).List(context.Background(), listOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to list deployment pods: %v", err)
	}

	return pods.Items, nil
}

// DeploymentStatus 表示 Deployment 的状态摘要
type DeploymentStatus struct {
	Name              string
	Namespace         string
	Replicas          int32
	AvailableReplicas int32
	ReadyReplicas     int32
	UpdatedReplicas   int32
	Conditions        []appsv1.DeploymentCondition
}

// GetDeploymentStatus 获取 Deployment 状态摘要
func (r *Resources) GetDeploymentStatus(clusterName, namespace, deploymentName string) (*DeploymentStatus, error) {
	deployment, err := r.GetDeployment(clusterName, namespace, deploymentName)
	if err != nil {
		return nil, err
	}

	status := deployment.Status
	return &DeploymentStatus{
		Name:              deployment.Name,
		Namespace:         deployment.Namespace,
		Replicas:          status.Replicas,
		AvailableReplicas: status.AvailableReplicas,
		ReadyReplicas:     status.ReadyReplicas,
		UpdatedReplicas:   status.UpdatedReplicas,
		Conditions:        status.Conditions,
	}, nil
}

// ========== Service 管理 ==========

// ListServices 列出指定命名空间的所有 Service
func (r *Resources) ListServices(clusterName, namespace string) ([]corev1.Service, error) {
	client, err := r.manager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}

	services, err := client.CoreV1().Services(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %v", err)
	}

	return services.Items, nil
}

// GetService 获取指定 Service 详情
func (r *Resources) GetService(clusterName, namespace, serviceName string) (*corev1.Service, error) {
	client, err := r.manager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}

	service, err := client.CoreV1().Services(namespace).Get(context.Background(), serviceName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get service: %v", err)
	}

	return service, nil
}

// GetServiceEndpoints 获取 Service 的 Endpoints
func (r *Resources) GetServiceEndpoints(clusterName, namespace, serviceName string) (*corev1.Endpoints, error) {
	client, err := r.manager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}

	endpoints, err := client.CoreV1().Endpoints(namespace).Get(context.Background(), serviceName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get endpoints: %v", err)
	}

	return endpoints, nil
}
