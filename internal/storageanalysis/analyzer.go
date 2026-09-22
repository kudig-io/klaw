package storageanalysis

import (
	"context"
	"time"

	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Analyzer struct {
	clientset kubernetes.Interface
}

func NewAnalyzer(clientset kubernetes.Interface) *Analyzer {
	return &Analyzer{clientset: clientset}
}

type CapacityInfo struct {
	TotalBytes     int64 `json:"totalBytes"`
	UsedBytes      int64 `json:"usedBytes"`
	AvailableBytes int64 `json:"availableBytes"`
}

type StorageAnalysis struct {
	TotalPVs            int            `json:"totalPVs"`
	TotalPVCs           int            `json:"totalPVCs"`
	TotalStorageClasses int            `json:"totalStorageClasses"`
	PVByStatus          map[string]int `json:"pvByStatus"`
	PVCByStatus         map[string]int `json:"pvcByStatus"`
	PVByStorageClass    map[string]int `json:"pvByStorageClass"`
	StorageCapacity     CapacityInfo   `json:"storageCapacity"`
	SCByProvisioner     map[string]int `json:"scByProvisioner"`
	Timestamp           time.Time      `json:"timestamp"`
}

// ListPVCs 列出 PVC；ns 为空时跨全部命名空间
func (a *Analyzer) ListPVCs(ctx context.Context, ns string) ([]corev1.PersistentVolumeClaim, error) {
	list, err := a.clientset.CoreV1().PersistentVolumeClaims(ns).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

func (a *Analyzer) ListPersistentVolumes(ctx context.Context) ([]corev1.PersistentVolume, error) {
	list, err := a.clientset.CoreV1().PersistentVolumes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

func (a *Analyzer) ListStorageClasses(ctx context.Context) ([]storagev1.StorageClass, error) {
	list, err := a.clientset.StorageV1().StorageClasses().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

func (a *Analyzer) AnalyzeStorage(ctx context.Context) (*StorageAnalysis, error) {
	pvs, err := a.ListPersistentVolumes(ctx)
	if err != nil {
		return nil, err
	}
	pvcs, err := a.ListPVCs(ctx, "")
	if err != nil {
		return nil, err
	}
	scs, err := a.ListStorageClasses(ctx)
	if err != nil {
		return nil, err
	}

	analysis := &StorageAnalysis{
		TotalPVs:            len(pvs),
		TotalPVCs:           len(pvcs),
		TotalStorageClasses: len(scs),
		PVByStatus:          make(map[string]int),
		PVCByStatus:         make(map[string]int),
		PVByStorageClass:    make(map[string]int),
		SCByProvisioner:     make(map[string]int),
		Timestamp:           time.Now(),
	}

	for _, pv := range pvs {
		status := string(pv.Status.Phase)
		analysis.PVByStatus[status]++

		sc := pv.Spec.StorageClassName
		if sc == "" {
			sc = "<none>"
		}
		analysis.PVByStorageClass[sc]++

		if cap, ok := pv.Spec.Capacity[corev1.ResourceStorage]; ok {
			analysis.StorageCapacity.TotalBytes += cap.Value()
		}
	}

	for _, pvc := range pvcs {
		status := string(pvc.Status.Phase)
		analysis.PVCByStatus[status]++

		// 以 PVC 请求量近似已用容量（与前端 mock 同口径），供页面渲染容量条
		if q, ok := pvc.Spec.Resources.Requests[corev1.ResourceStorage]; ok {
			analysis.StorageCapacity.UsedBytes += q.Value()
		}
	}

	for _, sc := range scs {
		analysis.SCByProvisioner[sc.Provisioner]++
	}

	analysis.StorageCapacity.AvailableBytes = analysis.StorageCapacity.TotalBytes - analysis.StorageCapacity.UsedBytes
	if analysis.StorageCapacity.AvailableBytes < 0 {
		analysis.StorageCapacity.AvailableBytes = 0
	}

	return analysis, nil
}
