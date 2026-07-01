package kubernetes

import (
	"context"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/kudig-io/klaw/internal/diag/types"
)

func TestDeploymentAnalyzer(t *testing.T) {
	a := NewDeploymentAnalyzer()

	if a.Name() != "kubernetes.deployment" {
		t.Errorf("Expected name 'kubernetes.deployment', got '%s'", a.Name())
	}

	replicas := int32(3)

	// Test with fake client - healthy deployment
	fakeClient := fake.NewSimpleClientset(&appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nginx-deployment",
			Namespace: "default",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
		},
		Status: appsv1.DeploymentStatus{
			ReadyReplicas:     3,
			AvailableReplicas: 3,
			UpdatedReplicas:   3,
		},
	})

	data := types.NewDiagnosticData(types.ModeOnline)
	data.K8sClient = fakeClient

	issues, err := a.Analyze(context.Background(), data)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	// Healthy deployment should produce no issues
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues for healthy deployment, got %d", len(issues))
	}
}

func TestDeploymentAnalyzer_InsufficientReplicas(t *testing.T) {
	a := NewDeploymentAnalyzer()

	replicas := int32(5)

	// Test with fake client - deployment with insufficient ready replicas
	fakeClient := fake.NewSimpleClientset(&appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nginx-deployment",
			Namespace: "default",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
		},
		Status: appsv1.DeploymentStatus{
			ReadyReplicas:     2,
			AvailableReplicas: 2,
			UpdatedReplicas:   5,
		},
	})

	data := types.NewDiagnosticData(types.ModeOnline)
	data.K8sClient = fakeClient

	issues, err := a.Analyze(context.Background(), data)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	// Should detect insufficient replicas
	if len(issues) == 0 {
		t.Error("Expected issues for deployment with insufficient replicas")
	}

	found := false
	for _, issue := range issues {
		if issue.ENName == "DEPLOYMENT_INSUFFICIENT_REPLICAS" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected DEPLOYMENT_INSUFFICIENT_REPLICAS issue")
	}
}

func TestDeploymentAnalyzer_ZeroReady(t *testing.T) {
	a := NewDeploymentAnalyzer()

	replicas := int32(3)

	// Test with fake client - deployment with zero ready replicas
	fakeClient := fake.NewSimpleClientset(&appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nginx-deployment",
			Namespace: "default",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
		},
		Status: appsv1.DeploymentStatus{
			ReadyReplicas:     0,
			AvailableReplicas: 0,
			UpdatedReplicas:   0,
		},
	})

	data := types.NewDiagnosticData(types.ModeOnline)
	data.K8sClient = fakeClient

	issues, err := a.Analyze(context.Background(), data)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	// Should detect critical issue
	if len(issues) == 0 {
		t.Fatal("Expected issues for deployment with zero ready replicas")
	}

	foundCritical := false
	for _, issue := range issues {
		if issue.ENName == "DEPLOYMENT_INSUFFICIENT_REPLICAS" && issue.Severity == types.SeverityCritical {
			foundCritical = true
			break
		}
	}
	if !foundCritical {
		t.Error("Expected Critical severity for deployment with zero ready replicas")
	}
}

func TestStatefulSetAnalyzer(t *testing.T) {
	a := NewStatefulSetAnalyzer()

	if a.Name() != "kubernetes.statefulset" {
		t.Errorf("Expected name 'kubernetes.statefulset', got '%s'", a.Name())
	}

	replicas := int32(3)

	// Test with fake client - healthy statefulset
	fakeClient := fake.NewSimpleClientset(&appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "web",
			Namespace: "default",
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: &replicas,
		},
		Status: appsv1.StatefulSetStatus{
			ReadyReplicas:    3,
			CurrentReplicas:  3,
			UpdatedReplicas:  3,
		},
	})

	data := types.NewDiagnosticData(types.ModeOnline)
	data.K8sClient = fakeClient

	issues, err := a.Analyze(context.Background(), data)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	// Healthy statefulset should produce no issues (or just info about update)
	// Since UpdatedReplicas == desired, no updating issue
	for _, issue := range issues {
		if issue.Severity == types.SeverityCritical || issue.Severity == types.SeverityWarning {
			t.Errorf("Expected no critical/warning issues for healthy statefulset, got: %v", issue)
		}
	}
}

func TestStatefulSetAnalyzer_InsufficientReplicas(t *testing.T) {
	a := NewStatefulSetAnalyzer()

	replicas := int32(3)

	// Test with fake client - statefulset with insufficient replicas
	fakeClient := fake.NewSimpleClientset(&appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "web",
			Namespace: "default",
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: &replicas,
		},
		Status: appsv1.StatefulSetStatus{
			ReadyReplicas:   1,
			CurrentReplicas: 1,
			UpdatedReplicas: 1,
		},
	})

	data := types.NewDiagnosticData(types.ModeOnline)
	data.K8sClient = fakeClient

	issues, err := a.Analyze(context.Background(), data)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	// Should detect insufficient replicas
	if len(issues) == 0 {
		t.Error("Expected issues for statefulset with insufficient replicas")
	}

	found := false
	for _, issue := range issues {
		if issue.ENName == "STATEFULSET_INSUFFICIENT_REPLICAS" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected STATEFULSET_INSUFFICIENT_REPLICAS issue")
	}
}

func TestDaemonSetAnalyzer(t *testing.T) {
	a := NewDaemonSetAnalyzer()

	if a.Name() != "kubernetes.daemonset" {
		t.Errorf("Expected name 'kubernetes.daemonset', got '%s'", a.Name())
	}

	// Test with fake client - healthy daemonset
	fakeClient := fake.NewSimpleClientset(&appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "fluentd",
			Namespace: "kube-system",
		},
		Status: appsv1.DaemonSetStatus{
			DesiredNumberScheduled: 3,
			CurrentNumberScheduled: 3,
			NumberReady:            3,
			NumberAvailable:        3,
			NumberMisscheduled:     0,
		},
	})

	data := types.NewDiagnosticData(types.ModeOnline)
	data.K8sClient = fakeClient

	issues, err := a.Analyze(context.Background(), data)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	// Healthy daemonset should produce no issues
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues for healthy daemonset, got %d", len(issues))
	}
}

func TestDaemonSetAnalyzer_Misscheduled(t *testing.T) {
	a := NewDaemonSetAnalyzer()

	// Test with fake client - daemonset with misscheduled pods
	fakeClient := fake.NewSimpleClientset(&appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "fluentd",
			Namespace: "kube-system",
		},
		Status: appsv1.DaemonSetStatus{
			DesiredNumberScheduled: 3,
			CurrentNumberScheduled: 3,
			NumberReady:            3,
			NumberAvailable:        3,
			NumberMisscheduled:     2,
		},
	})

	data := types.NewDiagnosticData(types.ModeOnline)
	data.K8sClient = fakeClient

	issues, err := a.Analyze(context.Background(), data)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	// Should detect misscheduled pods
	found := false
	for _, issue := range issues {
		if issue.ENName == "DAEMONSET_MISCHEDULED" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected DAEMONSET_MISCHEDULED issue")
	}
}

func TestDaemonSetAnalyzer_InsufficientReady(t *testing.T) {
	a := NewDaemonSetAnalyzer()

	// Test with fake client - daemonset with insufficient ready pods
	fakeClient := fake.NewSimpleClientset(&appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "fluentd",
			Namespace: "kube-system",
		},
		Status: appsv1.DaemonSetStatus{
			DesiredNumberScheduled: 3,
			CurrentNumberScheduled: 3,
			NumberReady:            1,
			NumberAvailable:        1,
			NumberMisscheduled:     0,
		},
	})

	data := types.NewDiagnosticData(types.ModeOnline)
	data.K8sClient = fakeClient

	issues, err := a.Analyze(context.Background(), data)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	// Should detect insufficient ready
	found := false
	for _, issue := range issues {
		if issue.ENName == "DAEMONSET_INSUFFICIENT_READY" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected DAEMONSET_INSUFFICIENT_READY issue")
	}
}

func TestPodDisruptionBudgetAnalyzer(t *testing.T) {
	a := NewPodDisruptionBudgetAnalyzer()

	if a.Name() != "kubernetes.pdb" {
		t.Errorf("Expected name 'kubernetes.pdb', got '%s'", a.Name())
	}

	minAvailable := intstr.FromInt(2)

	// Test with fake client - PDB blocking disruptions
	fakeClient := fake.NewSimpleClientset(&policyv1.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nginx-pdb",
			Namespace: "default",
		},
		Spec: policyv1.PodDisruptionBudgetSpec{
			MinAvailable: &minAvailable,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "nginx",
				},
			},
		},
		Status: policyv1.PodDisruptionBudgetStatus{
			DisruptionsAllowed: 0,
			ExpectedPods:       2,
			CurrentHealthy:     1,
			DesiredHealthy:     2,
		},
	})

	data := types.NewDiagnosticData(types.ModeOnline)
	data.K8sClient = fakeClient

	issues, err := a.Analyze(context.Background(), data)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	// Should detect blocked disruptions
	if len(issues) == 0 {
		t.Error("Expected issues for PDB blocking disruptions")
	}

	found := false
	for _, issue := range issues {
		if issue.ENName == "PDB_DISRUPTIONS_BLOCKED" {
			found = true
			if issue.Severity != types.SeverityInfo {
				t.Errorf("Expected Info severity, got %v", issue.Severity)
			}
			break
		}
	}
	if !found {
		t.Error("Expected PDB_DISRUPTIONS_BLOCKED issue")
	}
}

func TestPodDisruptionBudgetAnalyzer_Allowed(t *testing.T) {
	a := NewPodDisruptionBudgetAnalyzer()

	minAvailable := intstr.FromInt(1)

	// Test with fake client - PDB allowing disruptions
	fakeClient := fake.NewSimpleClientset(&policyv1.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nginx-pdb",
			Namespace: "default",
		},
		Spec: policyv1.PodDisruptionBudgetSpec{
			MinAvailable: &minAvailable,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "nginx",
				},
			},
		},
		Status: policyv1.PodDisruptionBudgetStatus{
			DisruptionsAllowed: 1,
			ExpectedPods:       2,
			CurrentHealthy:     2,
			DesiredHealthy:     1,
		},
	})

	data := types.NewDiagnosticData(types.ModeOnline)
	data.K8sClient = fakeClient

	issues, err := a.Analyze(context.Background(), data)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	// Healthy PDB should produce no issues
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues for healthy PDB, got %d", len(issues))
	}
}

func TestWorkloadAnalyzers_Registration(t *testing.T) {
	// Test that analyzers are properly registered
	deploymentAnalyzer := NewDeploymentAnalyzer()
	statefulSetAnalyzer := NewStatefulSetAnalyzer()
	daemonSetAnalyzer := NewDaemonSetAnalyzer()
	pdbAnalyzer := NewPodDisruptionBudgetAnalyzer()

	// Check categories
	analyzers := []struct {
		name     string
		category string
	}{
		{deploymentAnalyzer.Name(), deploymentAnalyzer.Category()},
		{statefulSetAnalyzer.Name(), statefulSetAnalyzer.Category()},
		{daemonSetAnalyzer.Name(), daemonSetAnalyzer.Category()},
		{pdbAnalyzer.Name(), pdbAnalyzer.Category()},
	}

	for _, a := range analyzers {
		if a.category != "kubernetes" {
			t.Errorf("Analyzer %s should have category 'kubernetes', got '%s'", a.name, a.category)
		}
	}

	// Check supported modes
	if len(deploymentAnalyzer.SupportedModes()) != 1 || deploymentAnalyzer.SupportedModes()[0] != types.ModeOnline {
		t.Error("Deployment analyzer should support ModeOnline only")
	}
}
