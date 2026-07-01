// Package kubernetes provides Kubernetes component analyzers
package kubernetes

import (
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kudig-io/klaw/internal/diag/analyzer"
	"github.com/kudig-io/klaw/internal/diag/types"
)

// PLEGAnalyzer checks for PLEG (Pod Lifecycle Event Generator) issues
type PLEGAnalyzer struct {
	*analyzer.BaseAnalyzer
}

// NewPLEGAnalyzer creates a new PLEG analyzer
func NewPLEGAnalyzer() *PLEGAnalyzer {
	return &PLEGAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer(
			"kubernetes.pleg",
			"检查Kubelet PLEG状态",
			"kubernetes",
			[]types.DataMode{types.ModeOffline, types.ModeOnline},
		),
	}
}

// Analyze performs PLEG analysis
func (a *PLEGAnalyzer) Analyze(ctx context.Context, data *types.DiagnosticData) ([]types.Issue, error) {
	var issues []types.Issue

	kubeletLog, ok := data.GetRawFile("logs/kubelet.log")
	if !ok {
		// Try daemon_status/kubelet_status
		kubeletLog, ok = data.GetRawFile("daemon_status/kubelet_status")
		if !ok {
			return issues, nil
		}
	}

	content := string(kubeletLog)

	plegCount := strings.Count(content, "PLEG is not healthy")
	if plegCount > 0 {
		issue := types.NewIssue(
			types.SeverityCritical,
			"Kubelet PLEG不健康",
			"KUBELET_PLEG_UNHEALTHY",
			fmt.Sprintf("PLEG（Pod生命周期事件生成器）不健康，出现 %d 次", plegCount),
			"logs/kubelet.log",
		).WithRemediation("检查容器运行时状态; 重启容器运行时: systemctl restart containerd")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	}

	return issues, nil
}

// CNIAnalyzer checks for CNI plugin errors
type CNIAnalyzer struct {
	*analyzer.BaseAnalyzer
}

// NewCNIAnalyzer creates a new CNI analyzer
func NewCNIAnalyzer() *CNIAnalyzer {
	return &CNIAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer(
			"kubernetes.cni",
			"检查CNI网络插件状态",
			"kubernetes",
			[]types.DataMode{types.ModeOffline, types.ModeOnline},
		),
	}
}

// Analyze performs CNI analysis
func (a *CNIAnalyzer) Analyze(ctx context.Context, data *types.DiagnosticData) ([]types.Issue, error) {
	var issues []types.Issue

	kubeletLog, ok := data.GetRawFile("logs/kubelet.log")
	if !ok {
		kubeletLog, ok = data.GetRawFile("daemon_status/kubelet_status")
		if !ok {
			return issues, nil
		}
	}

	content := string(kubeletLog)

	// Check for CNI errors
	if strings.Contains(content, "Failed to create pod sandbox") && strings.Contains(content, "CNI") {
		cniErrorCount := strings.Count(content, "CNI") + strings.Count(content, "cni")
		issue := types.NewIssue(
			types.SeverityCritical,
			"CNI网络插件错误",
			"CNI_PLUGIN_ERROR",
			fmt.Sprintf("CNI网络插件失败 %d 次", cniErrorCount),
			"logs/kubelet.log",
		).WithRemediation("检查CNI插件状态: ls /etc/cni/net.d/; 重启网络插件Pod")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	}

	return issues, nil
}

// CertificateAnalyzer checks for certificate expiration
type CertificateAnalyzer struct {
	*analyzer.BaseAnalyzer
}

// NewCertificateAnalyzer creates a new certificate analyzer
func NewCertificateAnalyzer() *CertificateAnalyzer {
	return &CertificateAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer(
			"kubernetes.certificate",
			"检查证书状态",
			"kubernetes",
			[]types.DataMode{types.ModeOffline, types.ModeOnline},
		),
	}
}

// Analyze performs certificate analysis
func (a *CertificateAnalyzer) Analyze(ctx context.Context, data *types.DiagnosticData) ([]types.Issue, error) {
	var issues []types.Issue

	kubeletLog, ok := data.GetRawFile("logs/kubelet.log")
	if !ok {
		kubeletLog, ok = data.GetRawFile("daemon_status/kubelet_status")
		if !ok {
			return issues, nil
		}
	}

	content := string(kubeletLog)

	if strings.Contains(content, "certificate has expired") {
		issue := types.NewIssue(
			types.SeverityCritical,
			"证书已过期",
			"CERTIFICATE_EXPIRED",
			"Kubelet证书已过期",
			"logs/kubelet.log",
		).WithRemediation("更新证书: kubeadm certs renew all; systemctl restart kubelet")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	} else if strings.Contains(content, "certificate will expire") {
		issue := types.NewIssue(
			types.SeverityWarning,
			"证书即将过期",
			"CERTIFICATE_EXPIRING",
			"Kubelet证书即将过期",
			"logs/kubelet.log",
		).WithRemediation("尽快更新证书: kubeadm certs renew all")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	}

	return issues, nil
}

// APIServerAnalyzer checks for API server connection issues
type APIServerAnalyzer struct {
	*analyzer.BaseAnalyzer
}

// NewAPIServerAnalyzer creates a new API server analyzer
func NewAPIServerAnalyzer() *APIServerAnalyzer {
	return &APIServerAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer(
			"kubernetes.apiserver",
			"检查API Server连接状态",
			"kubernetes",
			[]types.DataMode{types.ModeOffline, types.ModeOnline},
		),
	}
}

// Analyze performs API server connection analysis
func (a *APIServerAnalyzer) Analyze(ctx context.Context, data *types.DiagnosticData) ([]types.Issue, error) {
	var issues []types.Issue

	kubeletLog, ok := data.GetRawFile("logs/kubelet.log")
	if !ok {
		kubeletLog, ok = data.GetRawFile("daemon_status/kubelet_status")
		if !ok {
			return issues, nil
		}
	}

	content := string(kubeletLog)

	// Check for connection failures
	connFailCount := strings.Count(content, "Unable to connect to the server")
	connFailCount += strings.Count(content, "connection refused")

	if connFailCount > 10 {
		issue := types.NewIssue(
			types.SeverityCritical,
			"API Server连接失败",
			"APISERVER_CONNECTION_FAILED",
			fmt.Sprintf("无法连接到API Server，出现 %d 次", connFailCount),
			"logs/kubelet.log",
		).WithRemediation("检查网络连接: curl -k https://<api-server>:6443/healthz; 检查防火墙规则")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	}

	// Check for authentication failures
	if strings.Contains(content, "Unauthorized") {
		authFailCount := strings.Count(content, "Unauthorized")
		issue := types.NewIssue(
			types.SeverityCritical,
			"Kubelet认证失败",
			"KUBELET_AUTH_FAILED",
			fmt.Sprintf("Kubelet认证失败 %d 次", authFailCount),
			"logs/kubelet.log",
		).WithRemediation("检查kubeconfig和证书; 重新生成bootstrap token")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	}

	return issues, nil
}

// NodeStatusAnalyzer checks for node status issues
type NodeStatusAnalyzer struct {
	*analyzer.BaseAnalyzer
}

// NewNodeStatusAnalyzer creates a new node status analyzer
func NewNodeStatusAnalyzer() *NodeStatusAnalyzer {
	return &NodeStatusAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer(
			"kubernetes.node_status",
			"检查节点状态",
			"kubernetes",
			[]types.DataMode{types.ModeOffline, types.ModeOnline},
		),
	}
}

// Analyze performs node status analysis
func (a *NodeStatusAnalyzer) Analyze(ctx context.Context, data *types.DiagnosticData) ([]types.Issue, error) {
	var issues []types.Issue

	// Online mode: check node conditions via K8s API
	if data.HasK8sClient() {
		onlineIssues, err := a.analyzeOnline(ctx, data)
		if err == nil {
			issues = append(issues, onlineIssues...)
		}
	}

	// Also check raw files (offline mode or supplementary data)
	kubeletLog, ok := data.GetRawFile("logs/kubelet.log")
	if !ok {
		kubeletLog, ok = data.GetRawFile("daemon_status/kubelet_status")
		if !ok {
			// Try k8s/node_conditions for online collected data
			kubeletLog, ok = data.GetRawFile("k8s/node_conditions")
		}
	}

	if ok {
		content := string(kubeletLog)
		issues = append(issues, a.analyzeFromLogs(content)...)
	}

	return issues, nil
}

// analyzeOnline checks node status via K8s API
func (a *NodeStatusAnalyzer) analyzeOnline(
	ctx context.Context,
	data *types.DiagnosticData,
) ([]types.Issue, error) {
	nodes, err := a.fetchNodes(ctx, data)
	if err != nil {
		return nil, err
	}

	var issues []types.Issue
	for _, node := range nodes {
		issues = append(issues, a.checkNodeConditions(node)...)
		issues = append(issues, a.checkNodeTaints(node)...)
	}

	return issues, nil
}

// fetchNodes retrieves nodes from K8s API
func (a *NodeStatusAnalyzer) fetchNodes(
	ctx context.Context,
	data *types.DiagnosticData,
) ([]corev1.Node, error) {
	if data.NodeName != "" {
		node, err := data.K8sClient.CoreV1().Nodes().Get(ctx, data.NodeName, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		return []corev1.Node{*node}, nil
	}

	nodeList, err := data.K8sClient.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return nodeList.Items, nil
}

// checkNodeConditions checks all conditions for a single node
func (a *NodeStatusAnalyzer) checkNodeConditions(node corev1.Node) []types.Issue {
	var issues []types.Issue

	for _, cond := range node.Status.Conditions {
		if issue := a.evalNodeCondition(node.Name, cond); issue != nil {
			issue.AnalyzerName = a.Name()
			issues = append(issues, *issue)
		}
	}

	return issues
}

// evalNodeCondition evaluates a single node condition and returns an issue if problematic
func (a *NodeStatusAnalyzer) evalNodeCondition(
	nodeName string,
	cond corev1.NodeCondition,
) *types.Issue {
	switch cond.Type {
	case corev1.NodeReady:
		return a.evalNodeReady(nodeName, cond)
	case corev1.NodeDiskPressure:
		return a.evalDiskPressure(nodeName, cond)
	case corev1.NodeMemoryPressure:
		return a.evalMemoryPressure(nodeName, cond)
	case corev1.NodePIDPressure:
		return a.evalPIDPressure(nodeName, cond)
	case corev1.NodeNetworkUnavailable:
		return a.evalNetworkUnavailable(nodeName, cond)
	}
	return nil
}

func (a *NodeStatusAnalyzer) evalNodeReady(nodeName string, cond corev1.NodeCondition) *types.Issue {
	if cond.Status == corev1.ConditionTrue {
		return nil
	}
	return types.NewIssue(
		types.SeverityCritical,
		fmt.Sprintf("节点 %s NotReady", nodeName),
		"NODE_NOT_READY",
		fmt.Sprintf("节点 %s 处于NotReady状态: %s", nodeName, cond.Message),
		"k8s/node_status",
	).WithRemediation("检查kubelet和容器运行时状态; kubectl describe node " + nodeName)
}

func (a *NodeStatusAnalyzer) evalDiskPressure(nodeName string, cond corev1.NodeCondition) *types.Issue {
	if cond.Status != corev1.ConditionTrue {
		return nil
	}
	return types.NewIssue(
		types.SeverityWarning,
		fmt.Sprintf("节点 %s 磁盘压力", nodeName),
		"DISK_PRESSURE",
		fmt.Sprintf("节点 %s 存在磁盘压力: %s", nodeName, cond.Message),
		"k8s/node_status",
	).WithRemediation("清理磁盘空间: df -h; 删除无用镜像: crictl rmi --prune")
}

func (a *NodeStatusAnalyzer) evalMemoryPressure(nodeName string, cond corev1.NodeCondition) *types.Issue {
	if cond.Status != corev1.ConditionTrue {
		return nil
	}
	return types.NewIssue(
		types.SeverityWarning,
		fmt.Sprintf("节点 %s 内存压力", nodeName),
		"MEMORY_PRESSURE",
		fmt.Sprintf("节点 %s 存在内存压力: %s", nodeName, cond.Message),
		"k8s/node_status",
	).WithRemediation("检查内存使用: free -h; 优化Pod资源配置")
}

func (a *NodeStatusAnalyzer) evalPIDPressure(nodeName string, cond corev1.NodeCondition) *types.Issue {
	if cond.Status != corev1.ConditionTrue {
		return nil
	}
	return types.NewIssue(
		types.SeverityWarning,
		fmt.Sprintf("节点 %s PID压力", nodeName),
		"PID_PRESSURE",
		fmt.Sprintf("节点 %s 存在PID压力: %s", nodeName, cond.Message),
		"k8s/node_status",
	).WithRemediation("检查进程数: ps aux | wc -l; 检查PID资源限制")
}

func (a *NodeStatusAnalyzer) evalNetworkUnavailable(
	nodeName string,
	cond corev1.NodeCondition,
) *types.Issue {
	if cond.Status != corev1.ConditionTrue {
		return nil
	}
	return types.NewIssue(
		types.SeverityCritical,
		fmt.Sprintf("节点 %s 网络不可用", nodeName),
		"NETWORK_UNAVAILABLE",
		fmt.Sprintf("节点 %s 网络不可用: %s", nodeName, cond.Message),
		"k8s/node_status",
	).WithRemediation("检查CNI插件状态; 重启网络插件")
}

// checkNodeTaints checks node taints for issues
func (a *NodeStatusAnalyzer) checkNodeTaints(node corev1.Node) []types.Issue {
	var issues []types.Issue

	for _, taint := range node.Spec.Taints {
		if issue := a.evalUnschedulableTaint(node.Name, taint); issue != nil {
			issue.AnalyzerName = a.Name()
			issues = append(issues, *issue)
		}
	}

	return issues
}

func (a *NodeStatusAnalyzer) evalUnschedulableTaint(
	nodeName string,
	taint corev1.Taint,
) *types.Issue {
	if taint.Effect != corev1.TaintEffectNoSchedule {
		return nil
	}
	if taint.Key != "node.kubernetes.io/unschedulable" {
		return nil
	}
	return types.NewIssue(
		types.SeverityInfo,
		fmt.Sprintf("节点 %s 不可调度", nodeName),
		"NODE_UNSCHEDULABLE",
		fmt.Sprintf("节点 %s 被标记为不可调度", nodeName),
		"k8s/node_status",
	).WithRemediation("如需恢复调度: kubectl uncordon " + nodeName)
}

// analyzeFromLogs checks node status from log content
func (a *NodeStatusAnalyzer) analyzeFromLogs(content string) []types.Issue {
	var issues []types.Issue

	// Check for NotReady status
	if strings.Contains(content, "Node") && strings.Contains(content, "NotReady") {
		issue := types.NewIssue(
			types.SeverityCritical,
			"节点NotReady状态",
			"NODE_NOT_READY",
			"节点处于NotReady状态",
			"logs/kubelet.log",
		).WithRemediation("检查kubelet和容器运行时状态; kubectl describe node <node>")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	}

	// Check for DiskPressure
	if strings.Contains(content, "DiskPressure") {
		issue := types.NewIssue(
			types.SeverityWarning,
			"磁盘压力",
			"DISK_PRESSURE",
			"节点存在磁盘压力",
			"logs/kubelet.log",
		).WithRemediation("清理磁盘空间: df -h; 删除无用镜像: crictl rmi --prune")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	}

	// Check for MemoryPressure
	if strings.Contains(content, "MemoryPressure") {
		issue := types.NewIssue(
			types.SeverityWarning,
			"内存压力",
			"MEMORY_PRESSURE",
			"节点存在内存压力",
			"logs/kubelet.log",
		).WithRemediation("检查内存使用: free -h; 优化Pod资源配置")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	}

	// Check for pod eviction
	if strings.Contains(content, "evicted pod") || strings.Contains(content, "Evicted") {
		evictedCount := strings.Count(content, "evicted") + strings.Count(content, "Evicted")
		issue := types.NewIssue(
			types.SeverityWarning,
			"Pod被驱逐",
			"POD_EVICTED",
			fmt.Sprintf("Pod被驱逐 %d 次，可能由于资源不足", evictedCount),
			"logs/kubelet.log",
		).WithRemediation("检查节点资源使用情况; 调整eviction阈值")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	}

	return issues
}

// ImagePullAnalyzer checks for image pull failures
type ImagePullAnalyzer struct {
	*analyzer.BaseAnalyzer
}

// NewImagePullAnalyzer creates a new image pull analyzer
func NewImagePullAnalyzer() *ImagePullAnalyzer {
	return &ImagePullAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer(
			"kubernetes.image_pull",
			"检查镜像拉取状态",
			"kubernetes",
			[]types.DataMode{types.ModeOffline, types.ModeOnline},
		),
	}
}

// Analyze performs image pull analysis
func (a *ImagePullAnalyzer) Analyze(ctx context.Context, data *types.DiagnosticData) ([]types.Issue, error) {
	var issues []types.Issue

	// Online mode: check pod events for image pull issues
	if data.HasK8sClient() {
		onlineIssues, err := a.analyzeOnline(ctx, data)
		if err == nil {
			issues = append(issues, onlineIssues...)
		}
	}

	// Offline mode: check logs
	kubeletLog, ok := data.GetRawFile("logs/kubelet.log")
	if !ok {
		return issues, nil
	}

	content := string(kubeletLog)

	pullFailCount := strings.Count(content, "Failed to pull image")
	pullFailCount += strings.Count(content, "ImagePullBackOff")

	if pullFailCount > 5 {
		issue := types.NewIssue(
			types.SeverityWarning,
			"镜像拉取失败",
			"IMAGE_PULL_FAILED",
			fmt.Sprintf("镜像拉取失败 %d 次", pullFailCount),
			"logs/kubelet.log",
		).WithRemediation("检查镜像仓库连接; 检查imagePullSecrets配置")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	}

	return issues, nil
}

// analyzeOnline checks for image pull issues via K8s API
func (a *ImagePullAnalyzer) analyzeOnline(ctx context.Context, data *types.DiagnosticData) ([]types.Issue, error) {
	var issues []types.Issue

	// Get pods with ImagePullBackOff or ErrImagePull
	pods, err := data.K8sClient.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	imagePullIssues := make(map[string]int) // image -> count
	for _, pod := range pods.Items {
		for _, cs := range pod.Status.ContainerStatuses {
			if cs.State.Waiting != nil {
				reason := cs.State.Waiting.Reason
				if reason == "ImagePullBackOff" || reason == "ErrImagePull" {
					imagePullIssues[cs.Image]++
				}
			}
		}
	}

	for image, count := range imagePullIssues {
		issue := types.NewIssue(
			types.SeverityWarning,
			"镜像拉取失败",
			"IMAGE_PULL_FAILED",
			fmt.Sprintf("镜像 %s 拉取失败，影响 %d 个容器", image, count),
			"k8s/pods",
		).WithRemediation("检查镜像地址是否正确; 检查imagePullSecrets配置; crictl pull " + image)
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	}

	return issues, nil
}

// PodStatusAnalyzer checks for problematic pods (online mode only)
type PodStatusAnalyzer struct {
	*analyzer.BaseAnalyzer
}

// NewPodStatusAnalyzer creates a new pod status analyzer
func NewPodStatusAnalyzer() *PodStatusAnalyzer {
	return &PodStatusAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer(
			"kubernetes.pod_status",
			"检查Pod状态",
			"kubernetes",
			[]types.DataMode{types.ModeOnline},
		),
	}
}

// Analyze performs pod status analysis
func (a *PodStatusAnalyzer) Analyze(ctx context.Context, data *types.DiagnosticData) ([]types.Issue, error) {
	var issues []types.Issue

	if !data.HasK8sClient() {
		return issues, nil
	}

	listOpts := metav1.ListOptions{}
	if data.NodeName != "" {
		listOpts.FieldSelector = fmt.Sprintf("spec.nodeName=%s", data.NodeName)
	}

	pods, err := data.K8sClient.CoreV1().Pods("").List(ctx, listOpts)
	if err != nil {
		return nil, err
	}

	crashLoopPods := 0
	pendingPods := 0
	failedPods := 0
	highRestartPods := 0

	for _, pod := range pods.Items {
		// Check phase
		switch pod.Status.Phase {
		case corev1.PodPending:
			pendingPods++
		case corev1.PodFailed:
			failedPods++
		}

		// Check container statuses
		for _, cs := range pod.Status.ContainerStatuses {
			if cs.State.Waiting != nil && cs.State.Waiting.Reason == "CrashLoopBackOff" {
				crashLoopPods++
			}
			if cs.RestartCount > 10 {
				highRestartPods++
			}
		}
	}

	if crashLoopPods > 0 {
		issue := types.NewIssue(
			types.SeverityWarning,
			"Pod CrashLoopBackOff",
			"POD_CRASHLOOP",
			fmt.Sprintf("%d 个Pod处于CrashLoopBackOff状态", crashLoopPods),
			"k8s/pods",
		).WithRemediation("检查Pod日志: kubectl logs <pod> --previous; 检查资源限制和健康检查配置")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	}

	if pendingPods > 5 {
		issue := types.NewIssue(
			types.SeverityWarning,
			"Pod Pending",
			"POD_PENDING",
			fmt.Sprintf("%d 个Pod处于Pending状态", pendingPods),
			"k8s/pods",
		).WithRemediation("检查资源是否充足; kubectl describe pod <pod> 查看事件")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	}

	if failedPods > 0 {
		issue := types.NewIssue(
			types.SeverityWarning,
			"Pod Failed",
			"POD_FAILED",
			fmt.Sprintf("%d 个Pod处于Failed状态", failedPods),
			"k8s/pods",
		).WithRemediation("检查Pod日志和事件; 可能需要重新创建Pod")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	}

	if highRestartPods > 0 {
		issue := types.NewIssue(
			types.SeverityInfo,
			"Pod重启次数过多",
			"POD_HIGH_RESTARTS",
			fmt.Sprintf("%d 个Pod重启次数超过10次", highRestartPods),
			"k8s/pods",
		).WithRemediation("检查Pod日志查找崩溃原因; 优化资源配置和健康检查")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	}

	return issues, nil
}

// EventAnalyzer checks for warning events (online mode only)
type EventAnalyzer struct {
	*analyzer.BaseAnalyzer
}

// NewEventAnalyzer creates a new event analyzer
func NewEventAnalyzer() *EventAnalyzer {
	return &EventAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer(
			"kubernetes.events",
			"检查集群事件",
			"kubernetes",
			[]types.DataMode{types.ModeOnline},
		),
	}
}

// Analyze performs event analysis
func (a *EventAnalyzer) Analyze(ctx context.Context, data *types.DiagnosticData) ([]types.Issue, error) {
	var issues []types.Issue

	if !data.HasK8sClient() {
		return issues, nil
	}

	events, err := data.K8sClient.CoreV1().Events("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	warningEvents := make(map[string]int) // reason -> count

	for _, event := range events.Items {
		if event.Type == "Warning" {
			warningEvents[event.Reason]++
		}
	}

	// Check for common warning patterns
	criticalReasons := map[string]string{
		"FailedScheduling":   "Pod调度失败，检查资源和节点选择器",
		"FailedMount":        "存储卷挂载失败，检查PV/PVC状态",
		"FailedAttachVolume": "存储卷附加失败，检查存储后端",
		"Unhealthy":          "容器健康检查失败",
		"BackOff":            "容器启动失败或频繁重启",
		"NodeNotReady":       "节点不可用",
		"NetworkNotReady":    "网络未就绪",
	}

	for reason, desc := range criticalReasons {
		if count, ok := warningEvents[reason]; ok && count > 3 {
			severity := types.SeverityWarning
			if reason == "NodeNotReady" || reason == "NetworkNotReady" {
				severity = types.SeverityCritical
			}

			issue := types.NewIssue(
				severity,
				fmt.Sprintf("集群事件: %s", reason),
				fmt.Sprintf("EVENT_%s", strings.ToUpper(reason)),
				fmt.Sprintf("检测到 %d 个 %s 事件", count, reason),
				"k8s/events",
			).WithRemediation(desc)
			issue.AnalyzerName = a.Name()
			issues = append(issues, *issue)
		}
	}

	return issues, nil
}

// init registers all kubernetes analyzers
func init() {
	analyzer.Register(NewPLEGAnalyzer())
	analyzer.Register(NewCNIAnalyzer())
	analyzer.Register(NewCertificateAnalyzer())
	analyzer.Register(NewAPIServerAnalyzer())
	analyzer.Register(NewNodeStatusAnalyzer())
	analyzer.Register(NewImagePullAnalyzer())
	analyzer.Register(NewPodStatusAnalyzer())
	analyzer.Register(NewEventAnalyzer())
}
