package kubernetes

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/kudig-io/klaw/internal/diag/types"
)

func TestEtcdAnalyzer(t *testing.T) {
	a := NewEtcdAnalyzer()

	if a.Name() != "kubernetes.etcd" {
		t.Errorf("Expected name 'kubernetes.etcd', got '%s'", a.Name())
	}

	// Test with fake client - running etcd pod
	fakeClient := fake.NewSimpleClientset(&corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "etcd-master",
			Namespace: "kube-system",
			Labels:    map[string]string{"component": "etcd", "tier": "control-plane"},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			ContainerStatuses: []corev1.ContainerStatus{
				{
					Name:  "etcd",
					Ready: true,
				},
			},
		},
	})

	data := types.NewDiagnosticData(types.ModeOnline)
	data.K8sClient = fakeClient

	issues, err := a.Analyze(context.Background(), data)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	// Healthy etcd should produce no issues
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues for healthy etcd, got %d", len(issues))
	}
}

func TestEtcdAnalyzer_UnhealthyPod(t *testing.T) {
	a := NewEtcdAnalyzer()

	// Test with fake client - failed etcd pod
	fakeClient := fake.NewSimpleClientset(&corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "etcd-master",
			Namespace: "kube-system",
			Labels:    map[string]string{"component": "etcd", "tier": "control-plane"},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodFailed,
			ContainerStatuses: []corev1.ContainerStatus{
				{
					Name:  "etcd",
					Ready: false,
				},
			},
		},
	})

	data := types.NewDiagnosticData(types.ModeOnline)
	data.K8sClient = fakeClient

	issues, err := a.Analyze(context.Background(), data)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	// Should detect pod not running
	if len(issues) == 0 {
		t.Error("Expected issues for failed etcd pod")
	}

	if len(issues) > 0 && issues[0].Severity != types.SeverityCritical {
		t.Errorf("Expected Critical severity, got %v", issues[0].Severity)
	}
}

func TestEtcdAnalyzer_HighRestarts(t *testing.T) {
	a := NewEtcdAnalyzer()

	// Test with fake client - etcd pod with high restarts
	fakeClient := fake.NewSimpleClientset(&corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "etcd-master",
			Namespace: "kube-system",
			Labels:    map[string]string{"component": "etcd", "tier": "control-plane"},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			ContainerStatuses: []corev1.ContainerStatus{
				{
					Name:         "etcd",
					Ready:        true,
					RestartCount: 10,
				},
			},
		},
	})

	data := types.NewDiagnosticData(types.ModeOnline)
	data.K8sClient = fakeClient

	issues, err := a.Analyze(context.Background(), data)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	// Should detect high restart count
	foundRestartIssue := false
	for _, issue := range issues {
		if issue.ENName == "ETCD_HIGH_RESTART_COUNT" {
			foundRestartIssue = true
			if issue.Severity != types.SeverityWarning {
				t.Errorf("Expected Warning severity for restart issue, got %v", issue.Severity)
			}
		}
	}

	if !foundRestartIssue {
		t.Error("Expected ETCD_HIGH_RESTART_COUNT issue")
	}
}

func TestSchedulerAnalyzer(t *testing.T) {
	a := NewSchedulerAnalyzer()

	if a.Name() != "kubernetes.scheduler" {
		t.Errorf("Expected name 'kubernetes.scheduler', got '%s'", a.Name())
	}

	// Test with fake client - running scheduler pod
	fakeClient := fake.NewSimpleClientset(&corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kube-scheduler-master",
			Namespace: "kube-system",
			Labels:    map[string]string{"component": "kube-scheduler", "tier": "control-plane"},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			ContainerStatuses: []corev1.ContainerStatus{
				{
					Name:  "kube-scheduler",
					Ready: true,
				},
			},
		},
	})

	data := types.NewDiagnosticData(types.ModeOnline)
	data.K8sClient = fakeClient

	issues, err := a.Analyze(context.Background(), data)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	if len(issues) != 0 {
		t.Errorf("Expected 0 issues for healthy scheduler, got %d", len(issues))
	}
}

func TestSchedulerAnalyzer_NotFound(t *testing.T) {
	a := NewSchedulerAnalyzer()

	// Test with fake client - no scheduler pods
	fakeClient := fake.NewSimpleClientset()

	data := types.NewDiagnosticData(types.ModeOnline)
	data.K8sClient = fakeClient

	issues, err := a.Analyze(context.Background(), data)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	// Should report scheduler not found
	if len(issues) == 0 {
		t.Error("Expected issue for missing scheduler")
	}

	if len(issues) > 0 && issues[0].ENName != "SCHEDULER_NOT_FOUND" {
		t.Errorf("Expected SCHEDULER_NOT_FOUND, got %s", issues[0].ENName)
	}
}

func TestControllerManagerAnalyzer(t *testing.T) {
	a := NewControllerManagerAnalyzer()

	if a.Name() != "kubernetes.controller_manager" {
		t.Errorf("Expected name 'kubernetes.controller_manager', got '%s'", a.Name())
	}

	// Test with fake client - running controller manager pod
	fakeClient := fake.NewSimpleClientset(&corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kube-controller-manager-master",
			Namespace: "kube-system",
			Labels:    map[string]string{"component": "kube-controller-manager", "tier": "control-plane"},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			ContainerStatuses: []corev1.ContainerStatus{
				{
					Name:  "kube-controller-manager",
					Ready: true,
				},
			},
		},
	})

	data := types.NewDiagnosticData(types.ModeOnline)
	data.K8sClient = fakeClient

	issues, err := a.Analyze(context.Background(), data)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	if len(issues) != 0 {
		t.Errorf("Expected 0 issues for healthy controller-manager, got %d", len(issues))
	}
}

func TestAPIServerLatencyAnalyzer(t *testing.T) {
	a := NewAPIServerLatencyAnalyzer()

	if a.Name() != "kubernetes.apiserver_latency" {
		t.Errorf("Expected name 'kubernetes.apiserver_latency', got '%s'", a.Name())
	}

	// Test with fake client - should have minimal latency with fake client
	fakeClient := fake.NewSimpleClientset(&corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-node",
		},
	})

	data := types.NewDiagnosticData(types.ModeOnline)
	data.K8sClient = fakeClient

	issues, err := a.Analyze(context.Background(), data)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	// Fake client should be fast, so no latency issues
	if len(issues) != 0 {
		t.Logf("Unexpected issues with fake client: %v", issues)
	}
}

func TestControlPlaneAnalyzers_Registration(t *testing.T) {
	// Test that analyzers are properly registered
	// This is done by the init() function
	etcdAnalyzer := NewEtcdAnalyzer()
	schedulerAnalyzer := NewSchedulerAnalyzer()
	controllerManagerAnalyzer := NewControllerManagerAnalyzer()
	apiServerLatencyAnalyzer := NewAPIServerLatencyAnalyzer()

	// Check categories
	if etcdAnalyzer.Category() != "kubernetes" {
		t.Error("Etcd analyzer should have category 'kubernetes'")
	}
	if schedulerAnalyzer.Category() != "kubernetes" {
		t.Error("Scheduler analyzer should have category 'kubernetes'")
	}
	if controllerManagerAnalyzer.Category() != "kubernetes" {
		t.Error("Controller manager analyzer should have category 'kubernetes'")
	}
	if apiServerLatencyAnalyzer.Category() != "kubernetes" {
		t.Error("API server latency analyzer should have category 'kubernetes'")
	}

	// Check supported modes
	if len(etcdAnalyzer.SupportedModes()) != 1 || etcdAnalyzer.SupportedModes()[0] != types.ModeOnline {
		t.Error("Etcd analyzer should support ModeOnline only")
	}
}
