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
