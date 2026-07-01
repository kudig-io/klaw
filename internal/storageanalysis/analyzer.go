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
	TotalPVs            int                 `json:"totalPVs"`
	TotalPVCs           int                 `json:"totalPVCs"`
	TotalStorageClasses int                 `json:"totalStorageClasses"`
	PVByStatus          map[string][]string `json:"pvByStatus"`
	PVCByStatus         map[string][]string `json:"pvcByStatus"`
	PVByStorageClass    map[string][]string `json:"pvByStorageClass"`
	StorageCapacity     CapacityInfo        `json:"storageCapacity"`
	SCByProvisioner     map[string][]string `json:"scByProvisioner"`
	Timestamp           time.Time           `json:"timestamp"`
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
	pvcs, err := a.clientset.CoreV1().PersistentVolumeClaims("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	scs, err := a.ListStorageClasses(ctx)
	if err != nil {
		return nil, err
	}

	analysis := &StorageAnalysis{
		TotalPVs:            len(pvs),
		TotalPVCs:           len(pvcs.Items),
		TotalStorageClasses: len(scs),
		PVByStatus:          make(map[string][]string),
		PVCByStatus:         make(map[string][]string),
		PVByStorageClass:    make(map[string][]string),
		SCByProvisioner:     make(map[string][]string),
		Timestamp:           time.Now(),
	}

	for _, pv := range pvs {
		status := string(pv.Status.Phase)
		analysis.PVByStatus[status] = append(analysis.PVByStatus[status], pv.Name)

		sc := pv.Spec.StorageClassName
		if sc == "" {
			sc = "<none>"
		}
		analysis.PVByStorageClass[sc] = append(analysis.PVByStorageClass[sc], pv.Name)

		if cap, ok := pv.Spec.Capacity[corev1.ResourceStorage]; ok {
			analysis.StorageCapacity.TotalBytes += cap.Value()
		}
	}

	for _, pvc := range pvcs.Items {
		status := string(pvc.Status.Phase)
		analysis.PVCByStatus[status] = append(analysis.PVCByStatus[status], pvc.Name)
	}

	for _, sc := range scs {
		analysis.SCByProvisioner[sc.Provisioner] = append(analysis.SCByProvisioner[sc.Provisioner], sc.Name)
	}

	analysis.StorageCapacity.AvailableBytes = analysis.StorageCapacity.TotalBytes - analysis.StorageCapacity.UsedBytes
	if analysis.StorageCapacity.AvailableBytes < 0 {
		analysis.StorageCapacity.AvailableBytes = 0
	}

	return analysis, nil
}
