// Package kubernetes provides Kubernetes component analyzers
package kubernetes

import (
	"context"
	"fmt"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kudig-io/klaw/internal/diag/analyzer"
	"github.com/kudig-io/klaw/internal/diag/types"
)

// EtcdAnalyzer checks etcd cluster health
type EtcdAnalyzer struct {
	*analyzer.BaseAnalyzer
}

// NewEtcdAnalyzer creates a new etcd analyzer
func NewEtcdAnalyzer() *EtcdAnalyzer {
	return &EtcdAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer(
			"kubernetes.etcd",
			"检查 etcd 集群健康状态",
			"kubernetes",
			[]types.DataMode{types.ModeOnline},
		),
	}
}

// Analyze performs etcd health analysis
func (a *EtcdAnalyzer) Analyze(ctx context.Context, data *types.DiagnosticData) ([]types.Issue, error) {
	var issues []types.Issue

	if !data.HasK8sClient() {
		return issues, nil
	}

	// Check etcd pods in kube-system namespace
	pods, err := data.K8sClient.CoreV1().Pods("kube-system").List(ctx, metav1.ListOptions{
		LabelSelector: "component=etcd,tier=control-plane",
	})
	if err != nil {
		return issues, nil // Silently skip if we can't list pods
	}

	// Also try without the control-plane label for kubeadm clusters
	if len(pods.Items) == 0 {
		pods, _ = data.K8sClient.CoreV1().Pods("kube-system").List(ctx, metav1.ListOptions{
			LabelSelector: "app=etcd",
		})
	}

	for _, pod := range pods.Items {
		// Check pod phase
		if pod.Status.Phase != corev1.PodRunning {
			issue := types.NewIssue(
				types.SeverityCritical,
				fmt.Sprintf("etcd Pod %s 未运行", pod.Name),
				"ETCD_POD_NOT_RUNNING",
				fmt.Sprintf("etcd Pod %s 处于 %s 状态", pod.Name, pod.Status.Phase),
				"kube-system/etcd",
			).WithRemediation(fmt.Sprintf("检查 etcd Pod: kubectl describe pod %s -n kube-system", pod.Name))
			issue.AnalyzerName = a.Name()
			issues = append(issues, *issue)
			continue
		}

		// Check container statuses
		for _, containerStatus := range pod.Status.ContainerStatuses {
			if !containerStatus.Ready {
				issue := types.NewIssue(
					types.SeverityCritical,
					fmt.Sprintf("etcd 容器 %s 未就绪", containerStatus.Name),
					"ETCD_CONTAINER_NOT_READY",
					fmt.Sprintf("etcd 容器 %s 在 Pod %s 中未就绪", containerStatus.Name, pod.Name),
					"kube-system/etcd",
				).WithRemediation(fmt.Sprintf("检查容器日志: kubectl logs %s -n kube-system -c %s", pod.Name, containerStatus.Name))
				issue.AnalyzerName = a.Name()
				issues = append(issues, *issue)
			}
		}

		// Check restart count
		for _, containerStatus := range pod.Status.ContainerStatuses {
			if containerStatus.RestartCount > 5 {
				issue := types.NewIssue(
					types.SeverityWarning,
					fmt.Sprintf("etcd 容器重启次数过多"),
					"ETCD_HIGH_RESTART_COUNT",
					fmt.Sprintf("etcd 容器 %s 已重启 %d 次", containerStatus.Name, containerStatus.RestartCount),
					"kube-system/etcd",
				).WithRemediation("检查 etcd 健康状态: kubectl exec -it <etcd-pod> -n kube-system -- etcdctl endpoint health")
				issue.AnalyzerName = a.Name()
				issues = append(issues, *issue)
			}
		}
	}

	return issues, nil
}

// SchedulerAnalyzer checks kube-scheduler health
type SchedulerAnalyzer struct {
	*analyzer.BaseAnalyzer
}

// NewSchedulerAnalyzer creates a new scheduler analyzer
func NewSchedulerAnalyzer() *SchedulerAnalyzer {
	return &SchedulerAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer(
			"kubernetes.scheduler",
			"检查 kube-scheduler 健康状态",
			"kubernetes",
			[]types.DataMode{types.ModeOnline},
		),
	}
}

// Analyze performs scheduler health analysis
func (a *SchedulerAnalyzer) Analyze(ctx context.Context, data *types.DiagnosticData) ([]types.Issue, error) {
	var issues []types.Issue

	if !data.HasK8sClient() {
		return issues, nil
	}

	// Check scheduler pods
	pods, err := data.K8sClient.CoreV1().Pods("kube-system").List(ctx, metav1.ListOptions{
		LabelSelector: "component=kube-scheduler,tier=control-plane",
	})
	if err != nil || len(pods.Items) == 0 {
		// Try alternative labels
		pods, err = data.K8sClient.CoreV1().Pods("kube-system").List(ctx, metav1.ListOptions{
			LabelSelector: "app=kube-scheduler",
		})
		if err != nil || len(pods.Items) == 0 {
			// Try finding by name pattern
			allPods, _ := data.K8sClient.CoreV1().Pods("kube-system").List(ctx, metav1.ListOptions{})
			for _, pod := range allPods.Items {
				if strings.Contains(pod.Name, "scheduler") {
					pods.Items = append(pods.Items, pod)
				}
			}
		}
	}

	if len(pods.Items) == 0 {
		issue := types.NewIssue(
			types.SeverityWarning,
			"未找到 kube-scheduler Pod",
			"SCHEDULER_NOT_FOUND",
			"在 kube-system 命名空间中未找到 kube-scheduler Pod",
			"kube-system",
		).WithRemediation("检查 scheduler 是否部署: kubectl get pods -n kube-system -l component=kube-scheduler")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
		return issues, nil
	}

	for _, pod := range pods.Items {
		if pod.Status.Phase != corev1.PodRunning {
			issue := types.NewIssue(
				types.SeverityCritical,
				fmt.Sprintf("kube-scheduler %s 未运行", pod.Name),
				"SCHEDULER_NOT_RUNNING",
				fmt.Sprintf("kube-scheduler Pod %s 处于 %s 状态", pod.Name, pod.Status.Phase),
				"kube-system/scheduler",
			).WithRemediation(fmt.Sprintf("检查 scheduler: kubectl describe pod %s -n kube-system", pod.Name))
			issue.AnalyzerName = a.Name()
			issues = append(issues, *issue)
		}

		// Check for high restart count
		for _, cs := range pod.Status.ContainerStatuses {
			if cs.RestartCount > 3 {
				issue := types.NewIssue(
					types.SeverityWarning,
					"kube-scheduler 重启次数过多",
					"SCHEDULER_HIGH_RESTARTS",
					fmt.Sprintf("kube-scheduler 已重启 %d 次", cs.RestartCount),
					"kube-system/scheduler",
				).WithRemediation("检查 scheduler 日志: kubectl logs -n kube-system -l component=kube-scheduler")
				issue.AnalyzerName = a.Name()
				issues = append(issues, *issue)
			}
		}
	}

	return issues, nil
}

// ControllerManagerAnalyzer checks kube-controller-manager health
type ControllerManagerAnalyzer struct {
	*analyzer.BaseAnalyzer
}

// NewControllerManagerAnalyzer creates a new controller manager analyzer
func NewControllerManagerAnalyzer() *ControllerManagerAnalyzer {
	return &ControllerManagerAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer(
			"kubernetes.controller_manager",
			"检查 kube-controller-manager 健康状态",
			"kubernetes",
			[]types.DataMode{types.ModeOnline},
		),
	}
}

// Analyze performs controller manager health analysis
func (a *ControllerManagerAnalyzer) Analyze(ctx context.Context, data *types.DiagnosticData) ([]types.Issue, error) {
	var issues []types.Issue

	if !data.HasK8sClient() {
		return issues, nil
	}

	// Check controller manager pods
	pods, err := data.K8sClient.CoreV1().Pods("kube-system").List(ctx, metav1.ListOptions{
		LabelSelector: "component=kube-controller-manager,tier=control-plane",
	})
	if err != nil || len(pods.Items) == 0 {
		// Try alternative labels
		pods, err = data.K8sClient.CoreV1().Pods("kube-system").List(ctx, metav1.ListOptions{
			LabelSelector: "app=kube-controller-manager",
		})
		if err != nil || len(pods.Items) == 0 {
			// Try finding by name pattern
			allPods, _ := data.K8sClient.CoreV1().Pods("kube-system").List(ctx, metav1.ListOptions{})
			for _, pod := range allPods.Items {
				if strings.Contains(pod.Name, "controller-manager") {
					pods.Items = append(pods.Items, pod)
				}
			}
		}
	}

	if len(pods.Items) == 0 {
		issue := types.NewIssue(
			types.SeverityWarning,
			"未找到 kube-controller-manager Pod",
			"CONTROLLER_MANAGER_NOT_FOUND",
			"在 kube-system 命名空间中未找到 kube-controller-manager Pod",
			"kube-system",
		).WithRemediation("检查 controller-manager 是否部署: kubectl get pods -n kube-system -l component=kube-controller-manager")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
		return issues, nil
	}

	for _, pod := range pods.Items {
		if pod.Status.Phase != corev1.PodRunning {
			issue := types.NewIssue(
				types.SeverityCritical,
				fmt.Sprintf("kube-controller-manager %s 未运行", pod.Name),
				"CONTROLLER_MANAGER_NOT_RUNNING",
				fmt.Sprintf("kube-controller-manager Pod %s 处于 %s 状态", pod.Name, pod.Status.Phase),
				"kube-system/controller-manager",
			).WithRemediation(fmt.Sprintf("检查 controller-manager: kubectl describe pod %s -n kube-system", pod.Name))
			issue.AnalyzerName = a.Name()
			issues = append(issues, *issue)
		}

		// Check for high restart count
		for _, cs := range pod.Status.ContainerStatuses {
			if cs.RestartCount > 3 {
				issue := types.NewIssue(
					types.SeverityWarning,
					"kube-controller-manager 重启次数过多",
					"CONTROLLER_MANAGER_HIGH_RESTARTS",
					fmt.Sprintf("kube-controller-manager 已重启 %d 次", cs.RestartCount),
					"kube-system/controller-manager",
				).WithRemediation("检查 controller-manager 日志: kubectl logs -n kube-system -l component=kube-controller-manager")
				issue.AnalyzerName = a.Name()
				issues = append(issues, *issue)
			}
		}
	}

	return issues, nil
}

// APIServerLatencyAnalyzer checks API Server response latency
type APIServerLatencyAnalyzer struct {
	*analyzer.BaseAnalyzer
}

// NewAPIServerLatencyAnalyzer creates a new API server latency analyzer
func NewAPIServerLatencyAnalyzer() *APIServerLatencyAnalyzer {
	return &APIServerLatencyAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer(
			"kubernetes.apiserver_latency",
			"检查 API Server 响应延迟",
			"kubernetes",
			[]types.DataMode{types.ModeOnline},
		),
	}
}

// Analyze performs API server latency analysis
func (a *APIServerLatencyAnalyzer) Analyze(ctx context.Context, data *types.DiagnosticData) ([]types.Issue, error) {
	var issues []types.Issue

	if !data.HasK8sClient() {
		return issues, nil
	}

	// Measure API Server response time by listing nodes
	start := time.Now()
	_, err := data.K8sClient.CoreV1().Nodes().List(ctx, metav1.ListOptions{Limit: 1})
	duration := time.Since(start)

	if err != nil {
		issue := types.NewIssue(
			types.SeverityCritical,
			"API Server 无法访问",
			"APISERVER_UNREACHABLE",
			fmt.Sprintf("无法连接到 API Server: %v", err),
			"kubernetes/apiserver",
		).WithRemediation("检查 API Server 状态和认证配置")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
		return issues, nil
	}

	// Check latency thresholds
	if duration > 5*time.Second {
		issue := types.NewIssue(
			types.SeverityCritical,
			"API Server 响应延迟过高",
			"APISERVER_HIGH_LATENCY",
			fmt.Sprintf("API Server 响应时间 %.2f 秒，超过阈值 5 秒", duration.Seconds()),
			"kubernetes/apiserver",
		).WithRemediation("检查 API Server 负载和网络延迟; 考虑扩容 API Server 实例")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	} else if duration > 1*time.Second {
		issue := types.NewIssue(
			types.SeverityWarning,
			"API Server 响应延迟偏高",
			"APISERVER_ELEVATED_LATENCY",
			fmt.Sprintf("API Server 响应时间 %.2f 秒，建议关注", duration.Seconds()),
			"kubernetes/apiserver",
		).WithRemediation("监控 API Server 性能; 检查 etcd 健康状态")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	}

	return issues, nil
}

func init() {
	// Register control plane analyzers
	_ = analyzer.Register(NewEtcdAnalyzer())
	_ = analyzer.Register(NewSchedulerAnalyzer())
	_ = analyzer.Register(NewControllerManagerAnalyzer())
	_ = analyzer.Register(NewAPIServerLatencyAnalyzer())
}
