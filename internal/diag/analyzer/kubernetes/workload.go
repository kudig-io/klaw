// Package kubernetes provides Kubernetes component analyzers
package kubernetes

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kudig-io/klaw/internal/diag/analyzer"
	"github.com/kudig-io/klaw/internal/diag/types"
)

// DeploymentAnalyzer checks Deployment health
type DeploymentAnalyzer struct {
	*analyzer.BaseAnalyzer
}

// NewDeploymentAnalyzer creates a new deployment analyzer
func NewDeploymentAnalyzer() *DeploymentAnalyzer {
	return &DeploymentAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer(
			"kubernetes.deployment",
			"检查 Deployment 健康状态",
			"kubernetes",
			[]types.DataMode{types.ModeOnline},
		),
	}
}

// Analyze performs deployment health analysis
func (a *DeploymentAnalyzer) Analyze(ctx context.Context, data *types.DiagnosticData) ([]types.Issue, error) {
	var issues []types.Issue

	if !data.HasK8sClient() {
		return issues, nil
	}

	// List deployments in the target namespace or all namespaces
	listOpts := metav1.ListOptions{}
	namespace := ""
	if data.Namespace != "" {
		namespace = data.Namespace
	}

	deployments, err := data.K8sClient.AppsV1().Deployments(namespace).List(ctx, listOpts)
	if err != nil {
		return issues, nil // Silently skip if we can't list deployments
	}

	for _, deploy := range deployments.Items {
		issues = append(issues, a.checkDeployment(&deploy)...)
	}

	return issues, nil
}

func (a *DeploymentAnalyzer) checkDeployment(deploy *appsv1.Deployment) []types.Issue {
	var issues []types.Issue

	// Check if replicas match ready replicas
	desired := *deploy.Spec.Replicas
	ready := deploy.Status.ReadyReplicas
	available := deploy.Status.AvailableReplicas

	if ready < desired {
		severity := types.SeverityWarning
		if ready == 0 && desired > 0 {
			severity = types.SeverityCritical
		}

		issue := types.NewIssue(
			severity,
			fmt.Sprintf("Deployment %s/%s 就绪副本不足", deploy.Namespace, deploy.Name),
			"DEPLOYMENT_INSUFFICIENT_REPLICAS",
			fmt.Sprintf("期望副本数: %d, 就绪: %d", desired, ready),
			fmt.Sprintf("%s/%s", deploy.Namespace, deploy.Name),
		).WithRemediation(fmt.Sprintf("检查 Pod 状态: kubectl get pods -n %s -l app=%s", deploy.Namespace, deploy.Name))
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	}

	// Check if update is stuck
	if deploy.Spec.Strategy.RollingUpdate != nil && deploy.Status.UpdatedReplicas < desired {
		// Check if progress deadline has been exceeded
		// This is a simplified check; in reality, we'd need to check conditions
		for _, cond := range deploy.Status.Conditions {
			if cond.Type == appsv1.DeploymentProgressing && cond.Status != "True" {
				issue := types.NewIssue(
					types.SeverityWarning,
					fmt.Sprintf("Deployment %s/%s 更新可能卡住", deploy.Namespace, deploy.Name),
					"DEPLOYMENT_UPDATE_STUCK",
					fmt.Sprintf("已更新副本: %d/%d", deploy.Status.UpdatedReplicas, desired),
					fmt.Sprintf("%s/%s", deploy.Namespace, deploy.Name),
				).WithRemediation(fmt.Sprintf("检查 Deployment 事件: kubectl describe deployment %s -n %s", deploy.Name, deploy.Namespace))
				issue.AnalyzerName = a.Name()
				issues = append(issues, *issue)
			}
		}
	}

	// Check if available replicas are less than desired
	if available < desired {
		issue := types.NewIssue(
			types.SeverityWarning,
			fmt.Sprintf("Deployment %s/%s 可用副本不足", deploy.Namespace, deploy.Name),
			"DEPLOYMENT_UNAVAILABLE_REPLICAS",
			fmt.Sprintf("期望副本数: %d, 可用: %d", desired, available),
			fmt.Sprintf("%s/%s", deploy.Namespace, deploy.Name),
		).WithRemediation(fmt.Sprintf("检查 Pod 健康状态: kubectl get pods -n %s -l app=%s", deploy.Namespace, deploy.Name))
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	}

	return issues
}

// StatefulSetAnalyzer checks StatefulSet health
type StatefulSetAnalyzer struct {
	*analyzer.BaseAnalyzer
}

// NewStatefulSetAnalyzer creates a new StatefulSet analyzer
func NewStatefulSetAnalyzer() *StatefulSetAnalyzer {
	return &StatefulSetAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer(
			"kubernetes.statefulset",
			"检查 StatefulSet 健康状态",
			"kubernetes",
			[]types.DataMode{types.ModeOnline},
		),
	}
}

// Analyze performs StatefulSet health analysis
func (a *StatefulSetAnalyzer) Analyze(ctx context.Context, data *types.DiagnosticData) ([]types.Issue, error) {
	var issues []types.Issue

	if !data.HasK8sClient() {
		return issues, nil
	}

	namespace := ""
	if data.Namespace != "" {
		namespace = data.Namespace
	}

	statefulsets, err := data.K8sClient.AppsV1().StatefulSets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return issues, nil
	}

	for _, sts := range statefulsets.Items {
		issues = append(issues, a.checkStatefulSet(&sts)...)
	}

	return issues, nil
}

func (a *StatefulSetAnalyzer) checkStatefulSet(sts *appsv1.StatefulSet) []types.Issue {
	var issues []types.Issue

	desired := *sts.Spec.Replicas
	ready := sts.Status.ReadyReplicas
	current := sts.Status.CurrentReplicas

	// Check ready replicas
	if ready < desired {
		severity := types.SeverityWarning
		if ready == 0 && desired > 0 {
			severity = types.SeverityCritical
		}

		issue := types.NewIssue(
			severity,
			fmt.Sprintf("StatefulSet %s/%s 就绪副本不足", sts.Namespace, sts.Name),
			"STATEFULSET_INSUFFICIENT_REPLICAS",
			fmt.Sprintf("期望副本数: %d, 就绪: %d", desired, ready),
			fmt.Sprintf("%s/%s", sts.Namespace, sts.Name),
		).WithRemediation(fmt.Sprintf("检查 Pod 状态: kubectl get pods -n %s -l app=%s", sts.Namespace, sts.Name))
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	}

	// Check update progress
	if sts.Status.UpdatedReplicas < desired {
		issue := types.NewIssue(
			types.SeverityInfo,
			fmt.Sprintf("StatefulSet %s/%s 正在更新", sts.Namespace, sts.Name),
			"STATEFULSET_UPDATING",
			fmt.Sprintf("已更新副本: %d/%d", sts.Status.UpdatedReplicas, desired),
			fmt.Sprintf("%s/%s", sts.Namespace, sts.Name),
		).WithRemediation("等待更新完成，或检查是否有 Pod 卡住")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	}

	// Check if there are pods stuck in pending (not enough current replicas)
	if current < desired {
		issue := types.NewIssue(
			types.SeverityWarning,
			fmt.Sprintf("StatefulSet %s/%s 部分 Pod 未创建", sts.Namespace, sts.Name),
			"STATEFULSET_PENDING_PODS",
			fmt.Sprintf("当前副本: %d/%d", current, desired),
			fmt.Sprintf("%s/%s", sts.Namespace, sts.Name),
		).WithRemediation(fmt.Sprintf("检查 PVC 和存储: kubectl get pvc -n %s", sts.Namespace))
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	}

	return issues
}

// DaemonSetAnalyzer checks DaemonSet health
type DaemonSetAnalyzer struct {
	*analyzer.BaseAnalyzer
}

// NewDaemonSetAnalyzer creates a new DaemonSet analyzer
func NewDaemonSetAnalyzer() *DaemonSetAnalyzer {
	return &DaemonSetAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer(
			"kubernetes.daemonset",
			"检查 DaemonSet 健康状态",
			"kubernetes",
			[]types.DataMode{types.ModeOnline},
		),
	}
}

// Analyze performs DaemonSet health analysis
func (a *DaemonSetAnalyzer) Analyze(ctx context.Context, data *types.DiagnosticData) ([]types.Issue, error) {
	var issues []types.Issue

	if !data.HasK8sClient() {
		return issues, nil
	}

	namespace := ""
	if data.Namespace != "" {
		namespace = data.Namespace
	}

	daemonsets, err := data.K8sClient.AppsV1().DaemonSets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return issues, nil
	}

	for _, ds := range daemonsets.Items {
		issues = append(issues, a.checkDaemonSet(&ds)...)
	}

	return issues, nil
}

func (a *DaemonSetAnalyzer) checkDaemonSet(ds *appsv1.DaemonSet) []types.Issue {
	var issues []types.Issue

	desired := ds.Status.DesiredNumberScheduled
	ready := ds.Status.NumberReady
	available := ds.Status.NumberAvailable

	// Check if all desired pods are scheduled
	if ds.Status.CurrentNumberScheduled < desired {
		issue := types.NewIssue(
			types.SeverityWarning,
			fmt.Sprintf("DaemonSet %s/%s 部分节点未调度", ds.Namespace, ds.Name),
			"DAEMONSET_UNSCHEDULED_PODS",
			fmt.Sprintf("期望调度: %d, 已调度: %d", desired, ds.Status.CurrentNumberScheduled),
			fmt.Sprintf("%s/%s", ds.Namespace, ds.Name),
		).WithRemediation(fmt.Sprintf("检查节点资源或污点: kubectl describe daemonset %s -n %s", ds.Name, ds.Namespace))
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	}

	// Check ready pods
	if ready < desired {
		severity := types.SeverityWarning
		if ready == 0 && desired > 0 {
			severity = types.SeverityCritical
		}

		issue := types.NewIssue(
			severity,
			fmt.Sprintf("DaemonSet %s/%s 就绪 Pod 不足", ds.Namespace, ds.Name),
			"DAEMONSET_INSUFFICIENT_READY",
			fmt.Sprintf("期望: %d, 就绪: %d", desired, ready),
			fmt.Sprintf("%s/%s", ds.Namespace, ds.Name),
		).WithRemediation(fmt.Sprintf("检查 Pod 状态: kubectl get pods -n %s -l app=%s", ds.Namespace, ds.Name))
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	}

	// Check available pods
	if available < desired {
		issue := types.NewIssue(
			types.SeverityWarning,
			fmt.Sprintf("DaemonSet %s/%s 可用 Pod 不足", ds.Namespace, ds.Name),
			"DAEMONSET_INSUFFICIENT_AVAILABLE",
			fmt.Sprintf("期望: %d, 可用: %d", desired, available),
			fmt.Sprintf("%s/%s", ds.Namespace, ds.Name),
		).WithRemediation(fmt.Sprintf("检查 Pod 是否通过健康检查: kubectl get pods -n %s -l app=%s", ds.Namespace, ds.Name))
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	}

	// Check for misscheduled pods
	if ds.Status.NumberMisscheduled > 0 {
		issue := types.NewIssue(
			types.SeverityWarning,
			fmt.Sprintf("DaemonSet %s/%s 存在错误调度 Pod", ds.Namespace, ds.Name),
			"DAEMONSET_MISCHEDULED",
			fmt.Sprintf("错误调度 Pod 数: %d", ds.Status.NumberMisscheduled),
			fmt.Sprintf("%s/%s", ds.Namespace, ds.Name),
		).WithRemediation(fmt.Sprintf("检查节点选择器: kubectl describe daemonset %s -n %s", ds.Name, ds.Namespace))
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	}

	return issues
}

// PodDisruptionBudgetAnalyzer checks if PDB is protecting workloads
type PodDisruptionBudgetAnalyzer struct {
	*analyzer.BaseAnalyzer
}

// NewPodDisruptionBudgetAnalyzer creates a new PDB analyzer
func NewPodDisruptionBudgetAnalyzer() *PodDisruptionBudgetAnalyzer {
	return &PodDisruptionBudgetAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer(
			"kubernetes.pdb",
			"检查 PodDisruptionBudget 配置",
			"kubernetes",
			[]types.DataMode{types.ModeOnline},
		),
	}
}

// Analyze performs PDB analysis
func (a *PodDisruptionBudgetAnalyzer) Analyze(ctx context.Context, data *types.DiagnosticData) ([]types.Issue, error) {
	var issues []types.Issue

	if !data.HasK8sClient() {
		return issues, nil
	}

	namespace := ""
	if data.Namespace != "" {
		namespace = data.Namespace
	}

	pdbs, err := data.K8sClient.PolicyV1().PodDisruptionBudgets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return issues, nil
	}

	for _, pdb := range pdbs.Items {
		// Check if disruptions are currently allowed
		if pdb.Status.DisruptionsAllowed == 0 {
			// Check if there are pods causing the block
			if pdb.Status.ExpectedPods > 0 && pdb.Status.CurrentHealthy <= pdb.Status.DesiredHealthy {
				issue := types.NewIssue(
					types.SeverityInfo,
					fmt.Sprintf("PDB %s/%s 当前不允许驱逐", pdb.Namespace, pdb.Name),
					"PDB_DISRUPTIONS_BLOCKED",
					fmt.Sprintf("期望健康 Pod: %d, 当前健康: %d", pdb.Status.DesiredHealthy, pdb.Status.CurrentHealthy),
					fmt.Sprintf("%s/%s", pdb.Namespace, pdb.Name),
				).WithRemediation("确保有足够健康 Pod 以满足 PDB 要求; 或考虑暂时移除 PDB 进行维护")
				issue.AnalyzerName = a.Name()
				issues = append(issues, *issue)
			}
		}
	}

	return issues, nil
}

func init() {
	// Register workload analyzers
	analyzer.MustRegister(NewDeploymentAnalyzer())
	analyzer.MustRegister(NewStatefulSetAnalyzer())
	analyzer.MustRegister(NewDaemonSetAnalyzer())
	analyzer.MustRegister(NewPodDisruptionBudgetAnalyzer())
}
