// Package kubernetes provides Kubernetes component analyzers
package kubernetes

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/kudig-io/klaw/internal/diag/analyzer"
	"github.com/kudig-io/klaw/internal/diag/types"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PVCAnalyzer 检查 PVC 状态
type PVCAnalyzer struct {
	*analyzer.BaseAnalyzer
}

// NewPVCAnalyzer 创建 PVC 分析器
func NewPVCAnalyzer() *PVCAnalyzer {
	return &PVCAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer(
			"kubernetes.pvc",
			"检查 PVC 状态",
			"kubernetes",
			[]types.DataMode{types.ModeOnline},
		),
	}
}

// Analyze 执行 PVC 健康检查
func (a *PVCAnalyzer) Analyze(ctx context.Context, data *types.DiagnosticData) ([]types.Issue, error) {
	var issues []types.Issue

	if !data.HasK8sClient() {
		return issues, nil
	}

	client := data.K8sClient

	pvcs, err := client.CoreV1().PersistentVolumeClaims("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return issues, nil
	}

	for _, pvc := range pvcs.Items {
		pvcIssues := a.checkPVC(&pvc)
		for i := range pvcIssues {
			pvcIssues[i].AnalyzerName = a.Name()
			issues = append(issues, pvcIssues[i])
		}
	}

	return issues, nil
}

func (a *PVCAnalyzer) checkPVC(pvc *corev1.PersistentVolumeClaim) []types.Issue {
	var issues []types.Issue
	name := pvc.Name
	ns := pvc.Namespace

	switch pvc.Status.Phase {
	case corev1.ClaimPending:
		// 检查是否长时间处于 Pending
		age := time.Since(pvc.CreationTimestamp.Time)
		if age > 5*time.Minute {
			issue := types.NewIssue(
				types.SeverityWarning,
				fmt.Sprintf("PVC %s/%s 长时间 Pending", ns, name),
				"PVC_PENDING_LONG",
				fmt.Sprintf("PVC 创建于 %v 前，但仍处于 Pending 状态", age.Round(time.Minute)),
				fmt.Sprintf("%s/%s", ns, name),
			).WithRemediation("检查 StorageClass 是否存在且可用，或 provisioner 是否正常运行")
			issues = append(issues, *issue)
		}
	case corev1.ClaimLost:
		issue := types.NewIssue(
			types.SeverityCritical,
			fmt.Sprintf("PVC %s/%s 已丢失", ns, name),
			"PVC_LOST",
			"PVC 绑定的 PV 已丢失或不可用",
			fmt.Sprintf("%s/%s", ns, name),
		).WithRemediation("检查 PV 状态，可能需要手动恢复或重新创建数据")
		issues = append(issues, *issue)
	}

	// 检查容量使用情况（如果有 annotation）
	if used, ok := pvc.Annotations["volume.kubernetes.io/selected-node"]; ok && used == "" {
		issue := types.NewIssue(
			types.SeverityWarning,
			fmt.Sprintf("PVC %s/%s 未调度到节点", ns, name),
			"PVC_NOT_SCHEDULED",
			"PVC 的 selected-node annotation 为空，可能无法调度到任何节点",
			fmt.Sprintf("%s/%s", ns, name),
		).WithRemediation("检查 storage class 的 volumeBindingMode 或节点可用性")
		issues = append(issues, *issue)
	}

	return issues
}

// PVAnalyzer 检查 PV 状态
type PVAnalyzer struct {
	*analyzer.BaseAnalyzer
}

// NewPVAnalyzer 创建 PV 分析器
func NewPVAnalyzer() *PVAnalyzer {
	return &PVAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer(
			"kubernetes.pv",
			"检查 PV 状态",
			"kubernetes",
			[]types.DataMode{types.ModeOnline},
		),
	}
}

// Analyze 执行 PV 健康检查
func (a *PVAnalyzer) Analyze(ctx context.Context, data *types.DiagnosticData) ([]types.Issue, error) {
	var issues []types.Issue

	if !data.HasK8sClient() {
		return issues, nil
	}

	client := data.K8sClient

	pvs, err := client.CoreV1().PersistentVolumes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return issues, nil
	}

	for _, pv := range pvs.Items {
		pvIssues := a.checkPV(&pv)
		for i := range pvIssues {
			pvIssues[i].AnalyzerName = a.Name()
			issues = append(issues, pvIssues[i])
		}
	}

	return issues, nil
}

func (a *PVAnalyzer) checkPV(pv *corev1.PersistentVolume) []types.Issue {
	var issues []types.Issue
	name := pv.Name

	switch pv.Status.Phase {
	case corev1.VolumeFailed:
		issue := types.NewIssue(
			types.SeverityCritical,
			fmt.Sprintf("PV %s 处于 Failed 状态", name),
			"PV_FAILED",
			fmt.Sprintf("PV 回收/删除失败: %s", pv.Status.Reason),
			name,
		).WithRemediation("检查 PV 的后端存储状态，可能需要手动清理")
		issues = append(issues, *issue)
	case corev1.VolumeReleased:
		// 检查是否长时间处于 Released 状态
		age := time.Since(pv.CreationTimestamp.Time)
		if age > 24*time.Hour && pv.Spec.PersistentVolumeReclaimPolicy == corev1.PersistentVolumeReclaimDelete {
			// Delete 策略下应该很快被删除
			issue := types.NewIssue(
				types.SeverityWarning,
				fmt.Sprintf("PV %s 长时间处于 Released 状态", name),
				"PV_RELEASED_LONG",
				"PV 已被释放但未删除，可能存在回收问题",
				name,
			).WithRemediation("检查 volume provisioner 或手动删除 PV")
			issues = append(issues, *issue)
		}
	}

	// 检查容量
	capacity := pv.Spec.Capacity.Storage()
	if capacity != nil && capacity.Value() == 0 {
		issue := types.NewIssue(
			types.SeverityWarning,
			fmt.Sprintf("PV %s 容量为 0", name),
			"PV_ZERO_CAPACITY",
			"PV 的存储容量为 0，可能配置错误",
			name,
		).WithRemediation("检查 PV 的 capacity 配置")
		issues = append(issues, *issue)
	}

	return issues
}

// StorageClassAnalyzer 检查 StorageClass 配置
type StorageClassAnalyzer struct {
	*analyzer.BaseAnalyzer
}

// NewStorageClassAnalyzer 创建 StorageClass 分析器
func NewStorageClassAnalyzer() *StorageClassAnalyzer {
	return &StorageClassAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer(
			"kubernetes.storageclass",
			"检查 StorageClass 配置",
			"kubernetes",
			[]types.DataMode{types.ModeOnline},
		),
	}
}

// Analyze 执行 StorageClass 健康检查
func (a *StorageClassAnalyzer) Analyze(ctx context.Context, data *types.DiagnosticData) ([]types.Issue, error) {
	var issues []types.Issue

	if !data.HasK8sClient() {
		return issues, nil
	}

	client := data.K8sClient

	scs, err := client.StorageV1().StorageClasses().List(ctx, metav1.ListOptions{})
	if err != nil {
		return issues, nil
	}

	// 检查是否有默认 StorageClass
	var defaultSC *storagev1.StorageClass
	for i := range scs.Items {
		sc := &scs.Items[i]
		if sc.Annotations["storageclass.kubernetes.io/is-default-class"] == "true" {
			defaultSC = sc
			break
		}
	}

	if defaultSC == nil {
		issue := types.NewIssue(
			types.SeverityWarning,
			"未设置默认 StorageClass",
			"NO_DEFAULT_STORAGECLASS",
			"集群中没有标记为默认的 StorageClass，PVC 必须显式指定 storageClassName",
			"cluster/storage",
		).WithRemediation("使用 kubectl patch 设置默认 StorageClass")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	}

	// 检查每个 StorageClass
	for _, sc := range scs.Items {
		scIssues := a.checkStorageClass(&sc)
		for i := range scIssues {
			scIssues[i].AnalyzerName = a.Name()
			issues = append(issues, scIssues[i])
		}
	}

	return issues, nil
}

func (a *StorageClassAnalyzer) checkStorageClass(sc *storagev1.StorageClass) []types.Issue {
	var issues []types.Issue

	// 检查回收策略
	if sc.ReclaimPolicy == nil {
		issue := types.NewIssue(
			types.SeverityInfo,
			fmt.Sprintf("StorageClass %s 未设置 reclaimPolicy", sc.Name),
			"SC_NO_RECLAIM_POLICY",
			"未设置 reclaimPolicy 时默认使用 Delete",
			sc.Name,
		).WithRemediation("根据数据保留需求，显式设置 reclaimPolicy 为 Delete 或 Retain")
		issues = append(issues, *issue)
	}

	// 检查卷绑定模式
	if sc.VolumeBindingMode != nil && *sc.VolumeBindingMode == storagev1.VolumeBindingImmediate {
		// Immediate 模式在非多可用区环境可能有性能影响
		// 但在单可用区是正常的，所以只作为信息提示
		if strings.Contains(sc.Provisioner, "aws") || strings.Contains(sc.Provisioner, "gce") {
			// 云提供商建议使用 WaitForFirstConsumer
			issue := types.NewIssue(
				types.SeverityInfo,
				fmt.Sprintf("StorageClass %s 使用 Immediate 绑定模式", sc.Name),
				"SC_IMMEDIATE_BINDING",
				"云存储建议使用 WaitForFirstConsumer 模式以支持拓扑感知调度",
				sc.Name,
			).WithRemediation("考虑修改 volumeBindingMode 为 WaitForFirstConsumer")
			issues = append(issues, *issue)
		}
	}

	return issues
}

// VolumeAttachmentAnalyzer 检查 VolumeAttachment 状态
type VolumeAttachmentAnalyzer struct {
	*analyzer.BaseAnalyzer
}

// NewVolumeAttachmentAnalyzer 创建 VolumeAttachment 分析器
func NewVolumeAttachmentAnalyzer() *VolumeAttachmentAnalyzer {
	return &VolumeAttachmentAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer(
			"kubernetes.volumeattachment",
			"检查 VolumeAttachment 状态",
			"kubernetes",
			[]types.DataMode{types.ModeOnline},
		),
	}
}

// Analyze 执行 VolumeAttachment 健康检查
func (a *VolumeAttachmentAnalyzer) Analyze(ctx context.Context, data *types.DiagnosticData) ([]types.Issue, error) {
	var issues []types.Issue

	if !data.HasK8sClient() {
		return issues, nil
	}

	client := data.K8sClient

	vas, err := client.StorageV1().VolumeAttachments().List(ctx, metav1.ListOptions{})
	if err != nil {
		return issues, nil
	}

	for _, va := range vas.Items {
		if !va.Status.Attached {
			pvName := "unknown"
			if va.Spec.Source.PersistentVolumeName != nil {
				pvName = *va.Spec.Source.PersistentVolumeName
			}
			issue := types.NewIssue(
				types.SeverityWarning,
				fmt.Sprintf("Volume %s 未成功附加", pvName),
				"VOLUME_NOT_ATTACHED",
				fmt.Sprintf("VolumeAttachment %s 状态为未附加", va.Name),
				va.Name,
			).WithRemediation("检查 CSI driver 状态和节点存储插件")
			issue.AnalyzerName = a.Name()
			issues = append(issues, *issue)
		}
	}

	return issues, nil
}

// CSIDriverAnalyzer 检查 CSI Driver 状态
type CSIDriverAnalyzer struct {
	*analyzer.BaseAnalyzer
}

// NewCSIDriverAnalyzer 创建 CSI Driver 分析器
func NewCSIDriverAnalyzer() *CSIDriverAnalyzer {
	return &CSIDriverAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer(
			"kubernetes.csi",
			"检查 CSI Driver 状态",
			"kubernetes",
			[]types.DataMode{types.ModeOnline},
		),
	}
}

// Analyze 执行 CSI Driver 健康检查
func (a *CSIDriverAnalyzer) Analyze(ctx context.Context, data *types.DiagnosticData) ([]types.Issue, error) {
	var issues []types.Issue

	if !data.HasK8sClient() {
		return issues, nil
	}

	client := data.K8sClient

	drivers, err := client.StorageV1().CSIDrivers().List(ctx, metav1.ListOptions{})
	if err != nil {
		return issues, nil
	}

	for _, driver := range drivers.Items {
		// 检查 CSI Driver 的 Pod 是否运行
		selector := fmt.Sprintf("app=%s", driver.Name)
		pods, err := client.CoreV1().Pods("").List(ctx, metav1.ListOptions{
			LabelSelector: selector,
		})
		if err != nil {
			continue
		}

		if len(pods.Items) == 0 {
			// 尝试其他常见标签
			selector = fmt.Sprintf("app.kubernetes.io/name=%s", driver.Name)
			pods, _ = client.CoreV1().Pods("").List(ctx, metav1.ListOptions{
				LabelSelector: selector,
			})
		}

		if len(pods.Items) == 0 {
			issue := types.NewIssue(
				types.SeverityWarning,
				fmt.Sprintf("CSI Driver %s 的 Pod 未找到", driver.Name),
				"CSI_DRIVER_POD_MISSING",
				fmt.Sprintf("未找到运行 CSI Driver %s 的 Pod", driver.Name),
				fmt.Sprintf("csi/%s", driver.Name),
			).WithRemediation("检查 CSI Driver 的 Deployment/DaemonSet 是否正常部署")
			issue.AnalyzerName = a.Name()
			issues = append(issues, *issue)
		}
	}

	return issues, nil
}

// StoragePodAnalyzer 检查挂载存储的 Pod
type StoragePodAnalyzer struct {
	*analyzer.BaseAnalyzer
}

// NewStoragePodAnalyzer 创建存储 Pod 分析器
func NewStoragePodAnalyzer() *StoragePodAnalyzer {
	return &StoragePodAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer(
			"kubernetes.storage.pod",
			"检查 Pod 存储挂载状态",
			"kubernetes",
			[]types.DataMode{types.ModeOnline},
		),
	}
}

// Analyze 检查 Pod 的存储挂载问题
func (a *StoragePodAnalyzer) Analyze(ctx context.Context, data *types.DiagnosticData) ([]types.Issue, error) {
	var issues []types.Issue

	if !data.HasK8sClient() {
		return issues, nil
	}

	client := data.K8sClient

	pods, err := client.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return issues, nil
	}

	for _, pod := range pods.Items {
		// 检查 Pod 事件中的挂载错误
		if pod.Status.Phase == corev1.PodPending {
			for _, cond := range pod.Status.Conditions {
				if cond.Type == corev1.PodScheduled && cond.Status == corev1.ConditionFalse {
					if strings.Contains(cond.Message, "volume") || strings.Contains(cond.Message, "mount") {
						issue := types.NewIssue(
							types.SeverityWarning,
							fmt.Sprintf("Pod %s/%s 调度失败（存储相关）", pod.Namespace, pod.Name),
							"POD_SCHEDULE_VOLUME_FAILED",
							cond.Message,
							fmt.Sprintf("%s/%s", pod.Namespace, pod.Name),
						).WithRemediation("检查 PVC 状态、PV 可用性和节点存储资源")
						issue.AnalyzerName = a.Name()
						issues = append(issues, *issue)
					}
				}
			}
		}
	}

	return issues, nil
}

func init() {
	// 注册所有存储分析器
	_ = analyzer.Register(NewPVCAnalyzer())
	_ = analyzer.Register(NewPVAnalyzer())
	_ = analyzer.Register(NewStorageClassAnalyzer())
	_ = analyzer.Register(NewVolumeAttachmentAnalyzer())
	_ = analyzer.Register(NewCSIDriverAnalyzer())
	_ = analyzer.Register(NewStoragePodAnalyzer())
}
