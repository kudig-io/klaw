package kubernetes

import (
	"context"
	"testing"
	"time"

	"github.com/kudig-io/klaw/internal/diag/types"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/client-go/kubernetes/fake"
)

func TestPVCAnalyzer_Analyze_PendingLongTime(t *testing.T) {
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-pvc",
			Namespace:         "default",
			CreationTimestamp: metav1.Time{Time: time.Now().Add(-10 * time.Minute)},
		},
		Status: corev1.PersistentVolumeClaimStatus{
			Phase: corev1.ClaimPending,
		},
	}

	client := fake.NewSimpleClientset(pvc)
	data := &types.DiagnosticData{K8sClient: client}

	analyzer := NewPVCAnalyzer()
	issues, err := analyzer.Analyze(context.Background(), data)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(issues) != 1 {
		t.Errorf("expected 1 issue for pending PVC, got %d", len(issues))
	}
	if len(issues) > 0 && issues[0].ENName != "PVC_PENDING_LONG" {
		t.Errorf("expected PVC_PENDING_LONG, got %s", issues[0].ENName)
	}
}

func TestPVCAnalyzer_Analyze_Lost(t *testing.T) {
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pvc",
			Namespace: "default",
		},
		Status: corev1.PersistentVolumeClaimStatus{
			Phase: corev1.ClaimLost,
		},
	}

	client := fake.NewSimpleClientset(pvc)
	data := &types.DiagnosticData{K8sClient: client}

	analyzer := NewPVCAnalyzer()
	issues, err := analyzer.Analyze(context.Background(), data)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(issues) != 1 {
		t.Errorf("expected 1 issue for lost PVC, got %d", len(issues))
	}
	if len(issues) > 0 && issues[0].ENName != "PVC_LOST" {
		t.Errorf("expected PVC_LOST, got %s", issues[0].ENName)
	}
	if len(issues) > 0 && issues[0].Severity != types.SeverityCritical {
		t.Errorf("expected critical severity, got %v", issues[0].Severity)
	}
}

func TestPVCAnalyzer_Analyze_Bound(t *testing.T) {
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pvc",
			Namespace: "default",
		},
		Status: corev1.PersistentVolumeClaimStatus{
			Phase: corev1.ClaimBound,
		},
	}

	client := fake.NewSimpleClientset(pvc)
	data := &types.DiagnosticData{K8sClient: client}

	analyzer := NewPVCAnalyzer()
	issues, err := analyzer.Analyze(context.Background(), data)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("expected no issues for bound PVC, got %d", len(issues))
	}
}

func TestPVAnalyzer_Analyze_Failed(t *testing.T) {
	storageQuantity := resource.MustParse("10Gi")
	pv := &corev1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-pv",
		},
		Spec: corev1.PersistentVolumeSpec{
			Capacity: corev1.ResourceList{
				corev1.ResourceStorage: storageQuantity,
			},
		},
		Status: corev1.PersistentVolumeStatus{
			Phase:  corev1.VolumeFailed,
			Reason: "RecyclingFailed",
		},
	}

	client := fake.NewSimpleClientset(pv)
	data := &types.DiagnosticData{K8sClient: client}

	analyzer := NewPVAnalyzer()
	issues, err := analyzer.Analyze(context.Background(), data)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	
	hasFailed := false
	for _, issue := range issues {
		if issue.ENName == "PV_FAILED" {
			hasFailed = true
			break
		}
	}
	if !hasFailed {
		t.Errorf("expected PV_FAILED issue, got %v", issues)
	}
}

func TestPVAnalyzer_Analyze_ReleasedLongTime(t *testing.T) {
	storageQuantity := resource.MustParse("10Gi")
	deletePolicy := corev1.PersistentVolumeReclaimDelete
	pv := &corev1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-pv",
			CreationTimestamp: metav1.Time{Time: time.Now().Add(-48 * time.Hour)},
		},
		Spec: corev1.PersistentVolumeSpec{
			PersistentVolumeReclaimPolicy: deletePolicy,
			Capacity: corev1.ResourceList{
				corev1.ResourceStorage: storageQuantity,
			},
		},
		Status: corev1.PersistentVolumeStatus{
			Phase: corev1.VolumeReleased,
		},
	}

	client := fake.NewSimpleClientset(pv)
	data := &types.DiagnosticData{K8sClient: client}

	analyzer := NewPVAnalyzer()
	issues, err := analyzer.Analyze(context.Background(), data)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	
	hasReleasedLong := false
	for _, issue := range issues {
		if issue.ENName == "PV_RELEASED_LONG" {
			hasReleasedLong = true
			break
		}
	}
	if !hasReleasedLong {
		t.Errorf("expected PV_RELEASED_LONG issue, got %v", issues)
	}
}

func TestStorageClassAnalyzer_Analyze_NoDefault(t *testing.T) {
	sc := &storagev1.StorageClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: "standard",
		},
		Provisioner: "kubernetes.io/gce-pd",
		ReclaimPolicy: nil, // No reclaim policy set
	}

	client := fake.NewSimpleClientset(sc)
	data := &types.DiagnosticData{K8sClient: client}

	analyzer := NewStorageClassAnalyzer()
	issues, err := analyzer.Analyze(context.Background(), data)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	
	// 应该至少有 NO_DEFAULT_STORAGECLASS 问题
	hasNoDefault := false
	for _, issue := range issues {
		if issue.ENName == "NO_DEFAULT_STORAGECLASS" {
			hasNoDefault = true
			break
		}
	}
	if !hasNoDefault {
		t.Errorf("expected NO_DEFAULT_STORAGECLASS issue, got %v", issues)
	}
}

func TestStorageClassAnalyzer_Analyze_WithDefault(t *testing.T) {
	deletePolicy := corev1.PersistentVolumeReclaimDelete
	sc := &storagev1.StorageClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: "standard",
			Annotations: map[string]string{
				"storageclass.kubernetes.io/is-default-class": "true",
			},
		},
		Provisioner:   "kubernetes.io/gce-pd",
		ReclaimPolicy: &deletePolicy,
	}

	client := fake.NewSimpleClientset(sc)
	data := &types.DiagnosticData{K8sClient: client}

	analyzer := NewStorageClassAnalyzer()
	issues, err := analyzer.Analyze(context.Background(), data)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	
	// 检查没有 NO_DEFAULT_STORAGECLASS 问题
	hasNoDefault := false
	for _, issue := range issues {
		if issue.ENName == "NO_DEFAULT_STORAGECLASS" {
			hasNoDefault = true
			break
		}
	}
	if hasNoDefault {
		t.Errorf("expected no NO_DEFAULT_STORAGECLASS issue, got %v", issues)
	}
}

func TestStorageClassAnalyzer_Analyze_NoReclaimPolicy(t *testing.T) {
	sc := &storagev1.StorageClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: "standard",
			Annotations: map[string]string{
				"storageclass.kubernetes.io/is-default-class": "true",
			},
		},
		Provisioner: "kubernetes.io/gce-pd",
		// No ReclaimPolicy set
	}

	client := fake.NewSimpleClientset(sc)
	data := &types.DiagnosticData{K8sClient: client}

	analyzer := NewStorageClassAnalyzer()
	issues, err := analyzer.Analyze(context.Background(), data)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	hasNoReclaimPolicy := false
	for _, issue := range issues {
		if issue.ENName == "SC_NO_RECLAIM_POLICY" {
			hasNoReclaimPolicy = true
		}
	}

	if !hasNoReclaimPolicy {
		t.Errorf("expected SC_NO_RECLAIM_POLICY issue, got %v", issues)
	}
}

func TestVolumeAttachmentAnalyzer_Analyze_NotAttached(t *testing.T) {
	pvName := "test-pv"
	va := &storagev1.VolumeAttachment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-va",
		},
		Spec: storagev1.VolumeAttachmentSpec{
			Source: storagev1.VolumeAttachmentSource{
				PersistentVolumeName: &pvName,
			},
		},
		Status: storagev1.VolumeAttachmentStatus{
			Attached: false,
		},
	}

	client := fake.NewSimpleClientset(va)
	data := &types.DiagnosticData{K8sClient: client}

	analyzer := NewVolumeAttachmentAnalyzer()
	issues, err := analyzer.Analyze(context.Background(), data)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(issues) != 1 {
		t.Errorf("expected 1 issue for unattached volume, got %d", len(issues))
	}
	if len(issues) > 0 && issues[0].ENName != "VOLUME_NOT_ATTACHED" {
		t.Errorf("expected VOLUME_NOT_ATTACHED, got %s", issues[0].ENName)
	}
}

func TestCSIDriverAnalyzer_Analyze_MissingPod(t *testing.T) {
	driver := &storagev1.CSIDriver{
		ObjectMeta: metav1.ObjectMeta{
			Name: "ebs.csi.aws.com",
		},
	}

	client := fake.NewSimpleClientset(driver)
	data := &types.DiagnosticData{K8sClient: client}

	analyzer := NewCSIDriverAnalyzer()
	issues, err := analyzer.Analyze(context.Background(), data)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(issues) != 1 {
		t.Errorf("expected 1 issue for missing CSI driver pod, got %d", len(issues))
	}
	if len(issues) > 0 && issues[0].ENName != "CSI_DRIVER_POD_MISSING" {
		t.Errorf("expected CSI_DRIVER_POD_MISSING, got %s", issues[0].ENName)
	}
}

func TestStoragePodAnalyzer_Analyze_VolumeSchedulingFailed(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodPending,
			Conditions: []corev1.PodCondition{
				{
					Type:    corev1.PodScheduled,
					Status:  corev1.ConditionFalse,
					Message: "0/3 nodes are available: 3 node(s) had volume node affinity conflict",
				},
			},
		},
	}

	client := fake.NewSimpleClientset(pod)
	data := &types.DiagnosticData{K8sClient: client}

	analyzer := NewStoragePodAnalyzer()
	issues, err := analyzer.Analyze(context.Background(), data)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(issues) != 1 {
		t.Errorf("expected 1 issue for volume scheduling failure, got %d", len(issues))
	}
	if len(issues) > 0 && issues[0].ENName != "POD_SCHEDULE_VOLUME_FAILED" {
		t.Errorf("expected POD_SCHEDULE_VOLUME_FAILED, got %s", issues[0].ENName)
	}
}

func TestPVAnalyzer_Analyze_ZeroCapacity(t *testing.T) {
	pv := &corev1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-pv",
		},
		Spec: corev1.PersistentVolumeSpec{
			Capacity: corev1.ResourceList{
				corev1.ResourceStorage: resource.MustParse("0"),
			},
		},
		Status: corev1.PersistentVolumeStatus{
			Phase: corev1.VolumeAvailable,
		},
	}

	client := fake.NewSimpleClientset(pv)
	data := &types.DiagnosticData{K8sClient: client}

	analyzer := NewPVAnalyzer()
	issues, err := analyzer.Analyze(context.Background(), data)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(issues) != 1 {
		t.Errorf("expected 1 issue for zero capacity PV, got %d", len(issues))
	}
	if len(issues) > 0 && issues[0].ENName != "PV_ZERO_CAPACITY" {
		t.Errorf("expected PV_ZERO_CAPACITY, got %s", issues[0].ENName)
	}
}
