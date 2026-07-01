package storageanalysis

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestAnalyzeStorage(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&corev1.PersistentVolume{
			ObjectMeta: metav1.ObjectMeta{Name: "pv-1"},
			Spec: corev1.PersistentVolumeSpec{
				StorageClassName: "fast-ssd",
				Capacity: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse("10Gi"),
				},
			},
			Status: corev1.PersistentVolumeStatus{Phase: corev1.VolumeBound},
		},
		&corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{Name: "pvc-1", Namespace: "default"},
			Status:     corev1.PersistentVolumeClaimStatus{Phase: corev1.ClaimBound},
		},
		&storagev1.StorageClass{
			ObjectMeta: metav1.ObjectMeta{Name: "fast-ssd"},
			Provisioner: "kubernetes.io/aws-ebs",
		},
	)

	result, err := NewAnalyzer(clientset).AnalyzeStorage(context.Background())
	if err != nil {
		t.Fatalf("AnalyzeStorage: %v", err)
	}
	if result.TotalPVs != 1 {
		t.Errorf("TotalPVs = %d, want 1", result.TotalPVs)
	}
	if result.TotalPVCs != 1 {
		t.Errorf("TotalPVCs = %d, want 1", result.TotalPVCs)
	}
	if result.TotalStorageClasses != 1 {
		t.Errorf("TotalStorageClasses = %d, want 1", result.TotalStorageClasses)
	}
	if len(result.PVByStorageClass["fast-ssd"]) != 1 {
		t.Error("expected PV indexed by storage class fast-ssd")
	}
	if result.StorageCapacity.TotalBytes <= 0 {
		t.Error("expected positive total capacity from 10Gi PV")
	}
}
