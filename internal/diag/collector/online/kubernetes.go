// Package online implements the online K8s data collector
package online

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"

	"github.com/kudig-io/klaw/internal/diag/collector"
	"github.com/kudig-io/klaw/internal/diag/types"
)

// Collector collects diagnostic data from a live Kubernetes cluster
type Collector struct {
	clientMu sync.Mutex
	client   kubernetes.Interface
}

// NewCollector creates a new online collector
func NewCollector() *Collector {
	return &Collector{}
}

// Name returns the collector name
func (c *Collector) Name() string {
	return "online"
}

// Mode returns the data collection mode
func (c *Collector) Mode() types.DataMode {
	return types.ModeOnline
}

// Validate checks if the collector can operate with given config
func (c *Collector) Validate(config *collector.Config) error {
	// Try to build kubernetes client
	_, err := c.buildClient(config)
	if err != nil {
		return fmt.Errorf("failed to validate kubernetes config: %w", err)
	}
	return nil
}

// buildClient creates a kubernetes client from config
func (c *Collector) buildClient(config *collector.Config) (kubernetes.Interface, error) {
	var restConfig *rest.Config
	var err error

	// Try in-cluster config first if no kubeconfig specified
	if config.Kubeconfig == "" {
		restConfig, err = rest.InClusterConfig()
		if err != nil {
			// Fall back to default kubeconfig location
			kubeconfig := filepath.Join(homeDir(), ".kube", "config")
			if _, statErr := os.Stat(kubeconfig); statErr == nil {
				config.Kubeconfig = kubeconfig
			}
		}
	}

	// Build from kubeconfig if available
	if restConfig == nil && config.Kubeconfig != "" {
		loadingRules := &clientcmd.ClientConfigLoadingRules{ExplicitPath: config.Kubeconfig}
		configOverrides := &clientcmd.ConfigOverrides{}

		if config.Context != "" {
			configOverrides.CurrentContext = config.Context
		}

		clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
		restConfig, err = clientConfig.ClientConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to build client config: %w", err)
		}
	}

	if restConfig == nil {
		return nil, fmt.Errorf("no kubernetes configuration found (tried in-cluster and kubeconfig)")
	}

	// Set reasonable timeouts
	restConfig.Timeout = time.Duration(config.TimeoutSeconds) * time.Second

	// Create client
	client, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	return client, nil
}

func (c *Collector) getClient(config *collector.Config) (kubernetes.Interface, error) {
	c.clientMu.Lock()
	defer c.clientMu.Unlock()
	if c.client != nil {
		return c.client, nil
	}
	client, err := c.buildClient(config)
	if err != nil {
		return nil, err
	}
	c.client = client
	return client, nil
}

// Collect gathers diagnostic data from the kubernetes cluster
func (c *Collector) Collect(ctx context.Context, config *collector.Config) (*types.DiagnosticData, error) {
	client, err := c.getClient(config)
	if err != nil {
		return nil, err
	}

	data := types.NewDiagnosticData(types.ModeOnline)
	data.K8sClient = client
	data.Namespace = config.Namespace
	data.NodeName = config.NodeName

	// Collect node information
	if config.NodeName != "" {
		if err := c.collectNodeData(ctx, config.NodeName, data); err != nil {
			return nil, fmt.Errorf("failed to collect node data: %w", err)
		}
	} else if config.AllNodes {
		if err := c.collectAllNodesData(ctx, data); err != nil {
			return nil, fmt.Errorf("failed to collect all nodes data: %w", err)
		}
	} else {
		// Get current node (if running in-cluster)
		nodeName := os.Getenv("NODE_NAME")
		if nodeName != "" {
			if err := c.collectNodeData(ctx, nodeName, data); err != nil {
				return nil, fmt.Errorf("failed to collect current node data: %w", err)
			}
		}
	}

	// Collect cluster-wide data
	if err := c.collectClusterData(ctx, config, data); err != nil {
		klog.ErrorS(err, "Failed to collect cluster data")
		return nil, fmt.Errorf("failed to collect cluster data: %w", err)
	}

	return data, nil
}

// NodeResult contains the collected data and any error for a single node
type NodeResult struct {
	NodeName string
	Data     *types.DiagnosticData
	Error    error
}

// CollectAllNodesConcurrent collects diagnostic data from all nodes concurrently.
// It returns a map of node names to their collected data.
func (c *Collector) CollectAllNodesConcurrent(
	ctx context.Context,
	config *collector.Config,
	progressFn func(current, total int, nodeName string),
) ([]NodeResult, error) {
	client, err := c.getClient(config)
	if err != nil {
		return nil, err
	}

	// Get all nodes
	nodes, err := c.client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}

	totalNodes := len(nodes.Items)
	if totalNodes == 0 {
		return nil, fmt.Errorf("no nodes found in cluster")
	}

	// Results channel
	results := make([]NodeResult, 0, totalNodes)
	var mu sync.Mutex

	// Use errgroup for concurrent collection with limited concurrency
	g, ctx := errgroup.WithContext(ctx)
	semaphore := make(chan struct{}, 5) // Limit to 5 concurrent nodes

	for i, node := range nodes.Items {
		nodeName := node.Name
		nodeIndex := i

		g.Go(func() error {
			semaphore <- struct{}{}        // Acquire
			defer func() { <-semaphore }() // Release

			// Report progress start
			if progressFn != nil {
				progressFn(nodeIndex+1, totalNodes, nodeName)
			}

			// Collect data for this node
			nodeData := types.NewDiagnosticData(types.ModeOnline)
			nodeData.K8sClient = client
			nodeData.NodeName = nodeName

			if err := c.collectNodeData(ctx, nodeName, nodeData); err != nil {
				mu.Lock()
				results = append(results, NodeResult{
					NodeName: nodeName,
					Data:     nil,
					Error:    err,
				})
				mu.Unlock()
				return nil // Don't fail the whole operation for one node
			}

			mu.Lock()
			results = append(results, NodeResult{
				NodeName: nodeName,
				Data:     nodeData,
				Error:    nil,
			})
			mu.Unlock()

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return results, nil
}

// collectNodeData collects data for a specific node
func (c *Collector) collectNodeData(ctx context.Context, nodeName string, data *types.DiagnosticData) error {
	var node *corev1.Node
	err := retryWithBackoff(ctx, 2, func() error {
		var err error
		node, err = c.client.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
		return err
	})
	if err != nil {
		return fmt.Errorf("failed to get node %s: %w", nodeName, err)
	}

	// Populate node info
	data.NodeInfo = types.NodeInfo{
		Hostname:         node.Name,
		KernelVersion:    node.Status.NodeInfo.KernelVersion,
		OSImage:          node.Status.NodeInfo.OSImage,
		ContainerRuntime: node.Status.NodeInfo.ContainerRuntimeVersion,
		KubeletVersion:   node.Status.NodeInfo.KubeletVersion,
	}
	data.NodeName = nodeName

	// Store node status as raw data for analyzers
	nodeStatus := c.formatNodeStatus(node)
	data.SetRawFile("k8s/node_status", []byte(nodeStatus))

	// Store node conditions
	conditions := c.formatNodeConditions(node)
	data.SetRawFile("k8s/node_conditions", []byte(conditions))

	// Get node events
	events, err := c.getNodeEvents(ctx, nodeName)
	if err != nil {
		klog.V(2).InfoS("Failed to get node events", "node", nodeName, "err", err)
	} else if len(events) > 0 {
		data.SetRawFile("k8s/node_events", []byte(strings.Join(events, "\n")))
	}

	// Get pods on node
	pods, err := c.getPodsOnNode(ctx, nodeName)
	if err != nil {
		klog.V(2).InfoS("Failed to get pods on node", "node", nodeName, "err", err)
	} else {
		data.SetRawFile("k8s/node_pods", []byte(pods))
	}

	return nil
}

// collectAllNodesData collects data for all nodes
func (c *Collector) collectAllNodesData(ctx context.Context, data *types.DiagnosticData) error {
	nodes, err := c.client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list nodes: %w", err)
	}

	var nodesList strings.Builder
	for _, node := range nodes.Items {
		status := "Ready"
		for _, cond := range node.Status.Conditions {
			if cond.Type == corev1.NodeReady {
				if cond.Status != corev1.ConditionTrue {
					status = "NotReady"
				}
				break
			}
		}
		nodesList.WriteString(fmt.Sprintf("%s\t%s\t%s\t%s\n",
			node.Name,
			status,
			node.Status.NodeInfo.KubeletVersion,
			node.Status.NodeInfo.ContainerRuntimeVersion))
	}

	data.SetRawFile("k8s/nodes_list", []byte(nodesList.String()))

	// If only one node, collect its detailed info
	if len(nodes.Items) == 1 {
		return c.collectNodeData(ctx, nodes.Items[0].Name, data)
	}

	return nil
}

// collectClusterData collects cluster-wide information
func (c *Collector) collectClusterData(
	ctx context.Context,
	config *collector.Config,
	data *types.DiagnosticData,
) error {
	// ComponentStatuses API was removed in Kubernetes 1.21+
	// Skip collection on modern clusters

	// Get events from kube-system namespace
	kubeSystemEvents, err := c.getNamespaceEvents(ctx, "kube-system")
	if err != nil {
		klog.V(2).InfoS("Failed to get kube-system events", "err", err)
	} else if len(kubeSystemEvents) > 0 {
		data.SetRawFile("k8s/kube_system_events", []byte(strings.Join(kubeSystemEvents, "\n")))
	}

	// Get target namespace events if specified
	if config.Namespace != "" && config.Namespace != "kube-system" {
		nsEvents, err := c.getNamespaceEvents(ctx, config.Namespace)
		if err != nil {
			klog.V(2).InfoS("Failed to get namespace events", "namespace", config.Namespace, "err", err)
		} else if len(nsEvents) > 0 {
			data.SetRawFile("k8s/namespace_events", []byte(strings.Join(nsEvents, "\n")))
		}
	}

	// Get system pods status
	systemPods, err := c.getSystemPodsStatus(ctx)
	if err != nil {
		klog.V(2).InfoS("Failed to get system pods status", "err", err)
	} else {
		data.SetRawFile("k8s/system_pods", []byte(systemPods))
	}

	// Get daemonset status
	daemonsets, err := c.getDaemonSetStatus(ctx)
	if err != nil {
		klog.V(2).InfoS("Failed to get daemonset status", "err", err)
	} else {
		data.SetRawFile("k8s/daemonsets", []byte(daemonsets))
	}

	// Collect service mesh information
	c.collectServiceMeshInfo(ctx, data)

	return nil
}

// formatNodeStatus formats node status for analysis
func (c *Collector) formatNodeStatus(node *corev1.Node) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Node: %s\n", node.Name))
	sb.WriteString(fmt.Sprintf("Created: %s\n", node.CreationTimestamp.Format(time.RFC3339)))

	// Labels
	sb.WriteString("Labels:\n")
	for k, v := range node.Labels {
		sb.WriteString(fmt.Sprintf("  %s=%s\n", k, v))
	}

	// Taints
	sb.WriteString("Taints:\n")
	for _, taint := range node.Spec.Taints {
		sb.WriteString(fmt.Sprintf("  %s=%s:%s\n", taint.Key, taint.Value, taint.Effect))
	}

	// Capacity
	sb.WriteString("Capacity:\n")
	for name, qty := range node.Status.Capacity {
		sb.WriteString(fmt.Sprintf("  %s: %s\n", name, qty.String()))
	}

	// Allocatable
	sb.WriteString("Allocatable:\n")
	for name, qty := range node.Status.Allocatable {
		sb.WriteString(fmt.Sprintf("  %s: %s\n", name, qty.String()))
	}

	return sb.String()
}

// formatNodeConditions formats node conditions
func (c *Collector) formatNodeConditions(node *corev1.Node) string {
	var sb strings.Builder

	sb.WriteString("Conditions:\n")
	for _, cond := range node.Status.Conditions {
		sb.WriteString(fmt.Sprintf("  %s: %s (Reason: %s, Message: %s)\n",
			cond.Type, cond.Status, cond.Reason, cond.Message))
	}

	return sb.String()
}

// getNodeEvents gets recent events for a node
func (c *Collector) getNodeEvents(ctx context.Context, nodeName string) ([]string, error) {
	events, err := c.client.CoreV1().Events("").List(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("involvedObject.name=%s,involvedObject.kind=Node", nodeName),
	})
	if err != nil {
		return nil, err
	}

	var result []string
	for _, event := range events.Items {
		result = append(result, fmt.Sprintf("[%s] %s: %s - %s",
			event.LastTimestamp.Format(time.RFC3339),
			event.Type,
			event.Reason,
			event.Message))
	}

	return result, nil
}

// getPodsOnNode gets pods running on a specific node
func (c *Collector) getPodsOnNode(ctx context.Context, nodeName string) (string, error) {
	pods, err := c.client.CoreV1().Pods("").List(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("spec.nodeName=%s", nodeName),
		Limit:         500,
	})
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	for _, pod := range pods.Items {
		status := string(pod.Status.Phase)
		restarts := int32(0)
		for _, cs := range pod.Status.ContainerStatuses {
			restarts += cs.RestartCount
		}
		sb.WriteString(fmt.Sprintf("%s/%s\t%s\t%d restarts\n",
			pod.Namespace, pod.Name, status, restarts))
	}

	return sb.String(), nil
}

// getNamespaceEvents gets events from a namespace
func (c *Collector) getNamespaceEvents(ctx context.Context, namespace string) ([]string, error) {
	events, err := c.client.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var result []string
	for _, event := range events.Items {
		// Only include warning events from last hour
		if event.Type == "Warning" {
			if time.Since(event.LastTimestamp.Time) < time.Hour {
				result = append(result, fmt.Sprintf("[%s] %s/%s: %s - %s",
					event.LastTimestamp.Format(time.RFC3339),
					event.InvolvedObject.Kind,
					event.InvolvedObject.Name,
					event.Reason,
					event.Message))
			}
		}
	}

	return result, nil
}

// getSystemPodsStatus gets status of system pods
func (c *Collector) getSystemPodsStatus(ctx context.Context) (string, error) {
	pods, err := c.client.CoreV1().Pods("kube-system").List(ctx, metav1.ListOptions{
		Limit: 500,
	})
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	for _, pod := range pods.Items {
		status := string(pod.Status.Phase)
		ready := "0/0"

		totalContainers := len(pod.Spec.Containers)
		readyContainers := 0
		for _, cs := range pod.Status.ContainerStatuses {
			if cs.Ready {
				readyContainers++
			}
		}
		ready = fmt.Sprintf("%d/%d", readyContainers, totalContainers)

		sb.WriteString(fmt.Sprintf("%s\t%s\t%s\n", pod.Name, ready, status))
	}

	return sb.String(), nil
}

// getDaemonSetStatus gets daemonset status
func (c *Collector) getDaemonSetStatus(ctx context.Context) (string, error) {
	daemonsets, err := c.client.AppsV1().DaemonSets("kube-system").List(ctx, metav1.ListOptions{})
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	for _, ds := range daemonsets.Items {
		sb.WriteString(fmt.Sprintf("%s\tDesired:%d Ready:%d Updated:%d Available:%d\n",
			ds.Name,
			ds.Status.DesiredNumberScheduled,
			ds.Status.NumberReady,
			ds.Status.UpdatedNumberScheduled,
			ds.Status.NumberAvailable))
	}

	return sb.String(), nil
}

// collectServiceMeshInfo collects Istio and Linkerd information
func (c *Collector) collectServiceMeshInfo(ctx context.Context, data *types.DiagnosticData) {
	// Check for Istio
	c.collectIstioInfo(ctx, data)

	// Check for Linkerd
	c.collectLinkerdInfo(ctx, data)
}

// collectIstioInfo collects Istio service mesh information
func (c *Collector) collectIstioInfo(ctx context.Context, data *types.DiagnosticData) {
	istioInfo := &types.IstioInfo{
		InjectionEnabled: make(map[string]bool),
	}

	// Check if istio-system namespace exists
	_, err := c.client.CoreV1().Namespaces().Get(ctx, "istio-system", metav1.GetOptions{})
	if err != nil {
		// Istio not installed
		return
	}

	// Count istiod pods
	istiodPods, err := c.client.CoreV1().Pods("istio-system").List(ctx, metav1.ListOptions{
		LabelSelector: "app=istiod",
	})
	if err == nil {
		istioInfo.IstiodPods = len(istiodPods.Items)
		for _, pod := range istiodPods.Items {
			if pod.Status.Phase == corev1.PodRunning {
				for _, cond := range pod.Status.Conditions {
					if cond.Type == corev1.PodReady && cond.Status == corev1.ConditionTrue {
						istioInfo.IstiodReady++
						break
					}
				}
			}
		}
	}

	// Count ingress gateway pods
	ingressPods, err := c.client.CoreV1().Pods("istio-system").List(ctx, metav1.ListOptions{
		LabelSelector: "app=istio-ingressgateway",
	})
	if err == nil {
		istioInfo.IngressPods = len(ingressPods.Items)
		for _, pod := range ingressPods.Items {
			if pod.Status.Phase == corev1.PodRunning {
				for _, cond := range pod.Status.Conditions {
					if cond.Type == corev1.PodReady && cond.Status == corev1.ConditionTrue {
						istioInfo.IngressReady++
						break
					}
				}
			}
		}
	}

	// Count egress gateway pods
	egressPods, err := c.client.CoreV1().Pods("istio-system").List(ctx, metav1.ListOptions{
		LabelSelector: "app=istio-egressgateway",
	})
	if err == nil {
		istioInfo.EgressPods = len(egressPods.Items)
		for _, pod := range egressPods.Items {
			if pod.Status.Phase == corev1.PodRunning {
				for _, cond := range pod.Status.Conditions {
					if cond.Type == corev1.PodReady && cond.Status == corev1.ConditionTrue {
						istioInfo.EgressReady++
						break
					}
				}
			}
		}
	}

	data.IstioInfo = istioInfo
}

// collectLinkerdInfo collects Linkerd service mesh information
func (c *Collector) collectLinkerdInfo(ctx context.Context, data *types.DiagnosticData) {
	linkerdInfo := &types.LinkerdInfo{}

	// Check if linkerd namespace exists
	_, err := c.client.CoreV1().Namespaces().Get(ctx, "linkerd", metav1.GetOptions{})
	if err != nil {
		// Linkerd not installed
		return
	}

	// Count control plane pods
	controlPlanePods, err := c.client.CoreV1().Pods("linkerd").List(ctx, metav1.ListOptions{
		LabelSelector: "linkerd.io/control-plane-component",
	})
	if err == nil {
		linkerdInfo.ControlPlanePods = len(controlPlanePods.Items)
		for _, pod := range controlPlanePods.Items {
			if pod.Status.Phase == corev1.PodRunning {
				for _, cond := range pod.Status.Conditions {
					if cond.Type == corev1.PodReady && cond.Status == corev1.ConditionTrue {
						linkerdInfo.ControlPlaneReady++
						break
					}
				}
			}
		}
	}

	data.LinkerdInfo = linkerdInfo
}

// homeDir returns the user's home directory
func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // Windows
}

// retryWithBackoff retries a function with exponential backoff
func retryWithBackoff(ctx context.Context, maxRetries int, fn func() error) error {
	var lastErr error
	for i := 0; i <= maxRetries; i++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if lastErr = fn(); lastErr == nil {
			return nil
		}
		if i < maxRetries {
			backoff := time.Duration(1<<uint(i)) * 500 * time.Millisecond
			klog.V(2).InfoS("Retrying after error", "attempt", i+1, "backoff", backoff, "err", lastErr)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
		}
	}
	return lastErr
}

// init registers the collector with the default factory
func init() {
	collector.RegisterCollector(NewCollector())
}
