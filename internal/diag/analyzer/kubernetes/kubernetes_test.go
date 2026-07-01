package kubernetes

import (
	"context"
	"fmt"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/kudig-io/klaw/internal/diag/types"
)

// ============ PLEG Analyzer Tests ============

func TestNewPLEGAnalyzer(t *testing.T) {
	a := NewPLEGAnalyzer()
	if a == nil {
		t.Fatal("NewPLEGAnalyzer() returned nil")
	}
	if a.Name() != "kubernetes.pleg" {
		t.Errorf("Name() = %v, want %v", a.Name(), "kubernetes.pleg")
	}
}

func TestPLEGAnalyzerAnalyze_NoFile(t *testing.T) {
	a := NewPLEGAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues, got %d", len(issues))
	}
}

func TestPLEGAnalyzerAnalyze_NoIssue(t *testing.T) {
	a := NewPLEGAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"logs/kubelet.log": []byte("normal log content"),
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues, got %d", len(issues))
	}
}

func TestPLEGAnalyzerAnalyze_WithIssue(t *testing.T) {
	a := NewPLEGAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"logs/kubelet.log": []byte("PLEG is not healthy PLEG is not healthy"),
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].ENName != "KUBELET_PLEG_UNHEALTHY" {
		t.Errorf("ENName = %v, want %v", issues[0].ENName, "KUBELET_PLEG_UNHEALTHY")
	}
	if issues[0].Severity != types.SeverityCritical {
		t.Errorf("Severity = %v, want %v", issues[0].Severity, types.SeverityCritical)
	}
}

func TestPLEGAnalyzerAnalyze_DaemonStatus(t *testing.T) {
	a := NewPLEGAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"daemon_status/kubelet_status": []byte("PLEG is not healthy"),
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Errorf("Expected 1 issue from daemon_status, got %d", len(issues))
	}
}

// ============ CNI Analyzer Tests ============

func TestNewCNIAnalyzer(t *testing.T) {
	a := NewCNIAnalyzer()
	if a == nil {
		t.Fatal("NewCNIAnalyzer() returned nil")
	}
	if a.Name() != "kubernetes.cni" {
		t.Errorf("Name() = %v, want %v", a.Name(), "kubernetes.cni")
	}
}

func TestCNIAnalyzerAnalyze_NoIssue(t *testing.T) {
	a := NewCNIAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"logs/kubelet.log": []byte("normal log content"),
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues, got %d", len(issues))
	}
}

func TestCNIAnalyzerAnalyze_WithIssue(t *testing.T) {
	a := NewCNIAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"logs/kubelet.log": []byte("Failed to create pod sandbox CNI error"),
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].ENName != "CNI_PLUGIN_ERROR" {
		t.Errorf("ENName = %v, want %v", issues[0].ENName, "CNI_PLUGIN_ERROR")
	}
}

func TestCNIAnalyzerAnalyze_DaemonStatus(t *testing.T) {
	a := NewCNIAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"daemon_status/kubelet_status": []byte("Failed to create pod sandbox CNI failed"),
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Errorf("Expected 1 issue from daemon_status, got %d", len(issues))
	}
}

// ============ Certificate Analyzer Tests ============

func TestNewCertificateAnalyzer(t *testing.T) {
	a := NewCertificateAnalyzer()
	if a == nil {
		t.Fatal("NewCertificateAnalyzer() returned nil")
	}
	if a.Name() != "kubernetes.certificate" {
		t.Errorf("Name() = %v, want %v", a.Name(), "kubernetes.certificate")
	}
}

func TestCertificateAnalyzerAnalyze_NoIssue(t *testing.T) {
	a := NewCertificateAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"logs/kubelet.log": []byte("normal log content"),
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues, got %d", len(issues))
	}
}

func TestCertificateAnalyzerAnalyze_Expired(t *testing.T) {
	a := NewCertificateAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"logs/kubelet.log": []byte("certificate has expired"),
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].ENName != "CERTIFICATE_EXPIRED" {
		t.Errorf("ENName = %v, want %v", issues[0].ENName, "CERTIFICATE_EXPIRED")
	}
	if issues[0].Severity != types.SeverityCritical {
		t.Errorf("Severity = %v, want %v", issues[0].Severity, types.SeverityCritical)
	}
}

func TestCertificateAnalyzerAnalyze_Expiring(t *testing.T) {
	a := NewCertificateAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"logs/kubelet.log": []byte("certificate will expire soon"),
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].ENName != "CERTIFICATE_EXPIRING" {
		t.Errorf("ENName = %v, want %v", issues[0].ENName, "CERTIFICATE_EXPIRING")
	}
	if issues[0].Severity != types.SeverityWarning {
		t.Errorf("Severity = %v, want %v", issues[0].Severity, types.SeverityWarning)
	}
}

// ============ API Server Analyzer Tests ============

func TestNewAPIServerAnalyzer(t *testing.T) {
	a := NewAPIServerAnalyzer()
	if a == nil {
		t.Fatal("NewAPIServerAnalyzer() returned nil")
	}
	if a.Name() != "kubernetes.apiserver" {
		t.Errorf("Name() = %v, want %v", a.Name(), "kubernetes.apiserver")
	}
}

func TestAPIServerAnalyzerAnalyze_NoIssue(t *testing.T) {
	a := NewAPIServerAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"logs/kubelet.log": []byte("normal log content"),
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues, got %d", len(issues))
	}
}

func TestAPIServerAnalyzerAnalyze_ConnectionFailed(t *testing.T) {
	a := NewAPIServerAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	// Create log with >10 connection failures
	logContent := ""
	for i := 0; i < 15; i++ {
		logContent += "Unable to connect to the server\n"
	}
	data.RawFiles = map[string][]byte{
		"logs/kubelet.log": []byte(logContent),
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].ENName != "APISERVER_CONNECTION_FAILED" {
		t.Errorf("ENName = %v, want %v", issues[0].ENName, "APISERVER_CONNECTION_FAILED")
	}
}

func TestAPIServerAnalyzerAnalyze_ConnectionRefused(t *testing.T) {
	a := NewAPIServerAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	// Create log with >10 connection refused
	logContent := ""
	for i := 0; i < 15; i++ {
		logContent += "connection refused\n"
	}
	data.RawFiles = map[string][]byte{
		"logs/kubelet.log": []byte(logContent),
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
}

func TestAPIServerAnalyzerAnalyze_Unauthorized(t *testing.T) {
	a := NewAPIServerAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"logs/kubelet.log": []byte("Unauthorized access denied Unauthorized"),
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].ENName != "KUBELET_AUTH_FAILED" {
		t.Errorf("ENName = %v, want %v", issues[0].ENName, "KUBELET_AUTH_FAILED")
	}
}

// ============ Node Status Analyzer Tests ============

func TestNewNodeStatusAnalyzer(t *testing.T) {
	a := NewNodeStatusAnalyzer()
	if a == nil {
		t.Fatal("NewNodeStatusAnalyzer() returned nil")
	}
	if a.Name() != "kubernetes.node_status" {
		t.Errorf("Name() = %v, want %v", a.Name(), "kubernetes.node_status")
	}
}

func TestNodeStatusAnalyzerAnalyze_NoData(t *testing.T) {
	a := NewNodeStatusAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues, got %d", len(issues))
	}
}

func TestNodeStatusAnalyzerAnalyze_Offline_NotReady(t *testing.T) {
	a := NewNodeStatusAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"logs/kubelet.log": []byte("Node worker-1 NotReady"),
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].ENName != "NODE_NOT_READY" {
		t.Errorf("ENName = %v, want %v", issues[0].ENName, "NODE_NOT_READY")
	}
}

func TestNodeStatusAnalyzerAnalyze_Offline_DiskPressure(t *testing.T) {
	a := NewNodeStatusAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"logs/kubelet.log": []byte("Node has DiskPressure condition"),
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].ENName != "DISK_PRESSURE" {
		t.Errorf("ENName = %v, want %v", issues[0].ENName, "DISK_PRESSURE")
	}
}

func TestNodeStatusAnalyzerAnalyze_Offline_MemoryPressure(t *testing.T) {
	a := NewNodeStatusAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"logs/kubelet.log": []byte("Node has MemoryPressure condition"),
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].ENName != "MEMORY_PRESSURE" {
		t.Errorf("ENName = %v, want %v", issues[0].ENName, "MEMORY_PRESSURE")
	}
}

func TestNodeStatusAnalyzerAnalyze_Offline_PodEvicted(t *testing.T) {
	a := NewNodeStatusAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"logs/kubelet.log": []byte("evicted pod test-pod Evicted due to resource pressure"),
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].ENName != "POD_EVICTED" {
		t.Errorf("ENName = %v, want %v", issues[0].ENName, "POD_EVICTED")
	}
}

func TestNodeStatusAnalyzerAnalyze_Online_NodeNotReady(t *testing.T) {
	a := NewNodeStatusAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOnline)
	data.NodeName = "test-node"
	
	// Create fake client with NotReady node
	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-node",
		},
		Status: corev1.NodeStatus{
			Conditions: []corev1.NodeCondition{
				{
					Type:    corev1.NodeReady,
					Status:  corev1.ConditionFalse,
					Message: "Kubelet not ready",
				},
			},
		},
	}
	data.K8sClient = fake.NewSimpleClientset(node)

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].ENName != "NODE_NOT_READY" {
		t.Errorf("ENName = %v, want %v", issues[0].ENName, "NODE_NOT_READY")
	}
}

func TestNodeStatusAnalyzerAnalyze_Online_DiskPressure(t *testing.T) {
	a := NewNodeStatusAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOnline)
	data.NodeName = "test-node"
	
	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-node",
		},
		Status: corev1.NodeStatus{
			Conditions: []corev1.NodeCondition{
				{
					Type:    corev1.NodeReady,
					Status:  corev1.ConditionTrue,
				},
				{
					Type:    corev1.NodeDiskPressure,
					Status:  corev1.ConditionTrue,
					Message: "Disk pressure detected",
				},
			},
		},
	}
	data.K8sClient = fake.NewSimpleClientset(node)

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].ENName != "DISK_PRESSURE" {
		t.Errorf("ENName = %v, want %v", issues[0].ENName, "DISK_PRESSURE")
	}
}

func TestNodeStatusAnalyzerAnalyze_Online_MemoryPressure(t *testing.T) {
	a := NewNodeStatusAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOnline)
	data.NodeName = "test-node"
	
	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-node",
		},
		Status: corev1.NodeStatus{
			Conditions: []corev1.NodeCondition{
				{
					Type:    corev1.NodeReady,
					Status:  corev1.ConditionTrue,
				},
				{
					Type:    corev1.NodeMemoryPressure,
					Status:  corev1.ConditionTrue,
					Message: "Memory pressure detected",
				},
			},
		},
	}
	data.K8sClient = fake.NewSimpleClientset(node)

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].ENName != "MEMORY_PRESSURE" {
		t.Errorf("ENName = %v, want %v", issues[0].ENName, "MEMORY_PRESSURE")
	}
}

func TestNodeStatusAnalyzerAnalyze_Online_PIDPressure(t *testing.T) {
	a := NewNodeStatusAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOnline)
	data.NodeName = "test-node"
	
	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-node",
		},
		Status: corev1.NodeStatus{
			Conditions: []corev1.NodeCondition{
				{
					Type:    corev1.NodeReady,
					Status:  corev1.ConditionTrue,
				},
				{
					Type:    corev1.NodePIDPressure,
					Status:  corev1.ConditionTrue,
					Message: "PID pressure detected",
				},
			},
		},
	}
	data.K8sClient = fake.NewSimpleClientset(node)

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].ENName != "PID_PRESSURE" {
		t.Errorf("ENName = %v, want %v", issues[0].ENName, "PID_PRESSURE")
	}
}

func TestNodeStatusAnalyzerAnalyze_Online_NetworkUnavailable(t *testing.T) {
	a := NewNodeStatusAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOnline)
	data.NodeName = "test-node"
	
	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-node",
		},
		Status: corev1.NodeStatus{
			Conditions: []corev1.NodeCondition{
				{
					Type:    corev1.NodeReady,
					Status:  corev1.ConditionTrue,
				},
				{
					Type:    corev1.NodeNetworkUnavailable,
					Status:  corev1.ConditionTrue,
					Message: "Network not ready",
				},
			},
		},
	}
	data.K8sClient = fake.NewSimpleClientset(node)

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].ENName != "NETWORK_UNAVAILABLE" {
		t.Errorf("ENName = %v, want %v", issues[0].ENName, "NETWORK_UNAVAILABLE")
	}
	if issues[0].Severity != types.SeverityCritical {
		t.Errorf("Severity = %v, want %v", issues[0].Severity, types.SeverityCritical)
	}
}

func TestNodeStatusAnalyzerAnalyze_Online_Unschedulable(t *testing.T) {
	a := NewNodeStatusAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOnline)
	data.NodeName = "test-node"
	
	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-node",
		},
		Spec: corev1.NodeSpec{
			Taints: []corev1.Taint{
				{
					Key:    "node.kubernetes.io/unschedulable",
					Effect: corev1.TaintEffectNoSchedule,
				},
			},
		},
		Status: corev1.NodeStatus{
			Conditions: []corev1.NodeCondition{
				{
					Type:   corev1.NodeReady,
					Status: corev1.ConditionTrue,
				},
			},
		},
	}
	data.K8sClient = fake.NewSimpleClientset(node)

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].ENName != "NODE_UNSCHEDULABLE" {
		t.Errorf("ENName = %v, want %v", issues[0].ENName, "NODE_UNSCHEDULABLE")
	}
	if issues[0].Severity != types.SeverityInfo {
		t.Errorf("Severity = %v, want %v", issues[0].Severity, types.SeverityInfo)
	}
}

func TestNodeStatusAnalyzerAnalyze_Online_ListAllNodes(t *testing.T) {
	a := NewNodeStatusAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOnline)
	// No specific node name, should list all nodes
	
	node1 := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "node-1",
		},
		Status: corev1.NodeStatus{
			Conditions: []corev1.NodeCondition{
				{
					Type:   corev1.NodeReady,
					Status: corev1.ConditionFalse,
				},
			},
		},
	}
	node2 := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "node-2",
		},
		Status: corev1.NodeStatus{
			Conditions: []corev1.NodeCondition{
				{
					Type:   corev1.NodeReady,
					Status: corev1.ConditionTrue,
				},
				{
					Type:   corev1.NodeDiskPressure,
					Status: corev1.ConditionTrue,
				},
			},
		},
	}
	data.K8sClient = fake.NewSimpleClientset(node1, node2)

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 2 {
		t.Fatalf("Expected 2 issues, got %d", len(issues))
	}
}

func TestNodeStatusAnalyzerAnalyze_Online_K8sClientError(t *testing.T) {
	a := NewNodeStatusAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOnline)
	data.NodeName = "nonexistent-node"
	// Use empty fake client - Get will return error
	data.K8sClient = fake.NewSimpleClientset()

	// Should fallback to empty issues (no error)
	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() should not error on K8s client error, got: %v", err)
	}
	// Should return empty issues when online fails and no offline data
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues when K8s fails, got %d", len(issues))
	}
}

func TestNodeStatusAnalyzerAnalyze_K8sNodeConditions(t *testing.T) {
	a := NewNodeStatusAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"k8s/node_conditions": []byte("Node Ready DiskPressure MemoryPressure"),
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	// Should detect DiskPressure and MemoryPressure
	if len(issues) != 2 {
		t.Fatalf("Expected 2 issues, got %d", len(issues))
	}
}

// ============ Image Pull Analyzer Tests ============

func TestNewImagePullAnalyzer(t *testing.T) {
	a := NewImagePullAnalyzer()
	if a == nil {
		t.Fatal("NewImagePullAnalyzer() returned nil")
	}
	if a.Name() != "kubernetes.image_pull" {
		t.Errorf("Name() = %v, want %v", a.Name(), "kubernetes.image_pull")
	}
}

func TestImagePullAnalyzerAnalyze_Offline_NoIssue(t *testing.T) {
	a := NewImagePullAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"logs/kubelet.log": []byte("normal log content"),
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues, got %d", len(issues))
	}
}

func TestImagePullAnalyzerAnalyze_Offline_WithIssue(t *testing.T) {
	a := NewImagePullAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	// Create log with >5 failures
	logContent := ""
	for i := 0; i < 10; i++ {
		logContent += "Failed to pull image nginx:latest\n"
	}
	data.RawFiles = map[string][]byte{
		"logs/kubelet.log": []byte(logContent),
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].ENName != "IMAGE_PULL_FAILED" {
		t.Errorf("ENName = %v, want %v", issues[0].ENName, "IMAGE_PULL_FAILED")
	}
}

func TestImagePullAnalyzerAnalyze_Offline_ImagePullBackOff(t *testing.T) {
	a := NewImagePullAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	logContent := ""
	for i := 0; i < 10; i++ {
		logContent += "ImagePullBackOff for image redis:latest\n"
	}
	data.RawFiles = map[string][]byte{
		"logs/kubelet.log": []byte(logContent),
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
}

func TestImagePullAnalyzerAnalyze_Online_ImagePullBackOff(t *testing.T) {
	a := NewImagePullAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOnline)
	
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
		},
		Status: corev1.PodStatus{
			ContainerStatuses: []corev1.ContainerStatus{
				{
					Name:  "app",
					Image: "nginx:latest",
					State: corev1.ContainerState{
						Waiting: &corev1.ContainerStateWaiting{
							Reason: "ImagePullBackOff",
						},
					},
				},
			},
		},
	}
	data.K8sClient = fake.NewSimpleClientset(pod)

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].ENName != "IMAGE_PULL_FAILED" {
		t.Errorf("ENName = %v, want %v", issues[0].ENName, "IMAGE_PULL_FAILED")
	}
}

func TestImagePullAnalyzerAnalyze_Online_ErrImagePull(t *testing.T) {
	a := NewImagePullAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOnline)
	
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
		},
		Status: corev1.PodStatus{
			ContainerStatuses: []corev1.ContainerStatus{
				{
					Name:  "app",
					Image: "private.registry.com/app:v1",
					State: corev1.ContainerState{
						Waiting: &corev1.ContainerStateWaiting{
							Reason: "ErrImagePull",
						},
					},
				},
			},
		},
	}
	data.K8sClient = fake.NewSimpleClientset(pod)

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
}

func TestImagePullAnalyzerAnalyze_Online_MultipleImages(t *testing.T) {
	a := NewImagePullAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOnline)
	
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
		},
		Status: corev1.PodStatus{
			ContainerStatuses: []corev1.ContainerStatus{
				{
					Name:  "app1",
					Image: "nginx:latest",
					State: corev1.ContainerState{
						Waiting: &corev1.ContainerStateWaiting{
							Reason: "ImagePullBackOff",
						},
					},
				},
				{
					Name:  "app2",
					Image: "redis:latest",
					State: corev1.ContainerState{
						Waiting: &corev1.ContainerStateWaiting{
							Reason: "ImagePullBackOff",
						},
					},
				},
			},
		},
	}
	data.K8sClient = fake.NewSimpleClientset(pod)

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 2 {
		t.Fatalf("Expected 2 issues, got %d", len(issues))
	}
}

func TestImagePullAnalyzerAnalyze_Online_K8sClientError(t *testing.T) {
	a := NewImagePullAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOnline)
	data.K8sClient = fake.NewSimpleClientset()
	// No pods, should return empty but not error

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() should not error, got: %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues, got %d", len(issues))
	}
}

// ============ Pod Status Analyzer Tests ============

func TestNewPodStatusAnalyzer(t *testing.T) {
	a := NewPodStatusAnalyzer()
	if a == nil {
		t.Fatal("NewPodStatusAnalyzer() returned nil")
	}
	if a.Name() != "kubernetes.pod_status" {
		t.Errorf("Name() = %v, want %v", a.Name(), "kubernetes.pod_status")
	}
}

func TestPodStatusAnalyzerAnalyze_NoK8sClient(t *testing.T) {
	a := NewPodStatusAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues without K8s client, got %d", len(issues))
	}
}

func TestPodStatusAnalyzerAnalyze_CrashLoopBackOff(t *testing.T) {
	a := NewPodStatusAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOnline)
	
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "crash-pod",
			Namespace: "default",
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			ContainerStatuses: []corev1.ContainerStatus{
				{
					Name: "app",
					State: corev1.ContainerState{
						Waiting: &corev1.ContainerStateWaiting{
							Reason: "CrashLoopBackOff",
						},
					},
				},
			},
		},
	}
	data.K8sClient = fake.NewSimpleClientset(pod)

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].ENName != "POD_CRASHLOOP" {
		t.Errorf("ENName = %v, want %v", issues[0].ENName, "POD_CRASHLOOP")
	}
}

func TestPodStatusAnalyzerAnalyze_MultiplePendingPods(t *testing.T) {
	a := NewPodStatusAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOnline)
	
	// Create 6 pending pods
	var pods []corev1.Pod
	for i := 0; i < 6; i++ {
		pods = append(pods, corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("pending-pod-%d", i),
				Namespace: "default",
			},
			Status: corev1.PodStatus{
				Phase: corev1.PodPending,
			},
		})
	}
	
	client := fake.NewSimpleClientset()
	for i := range pods {
		_, err := client.CoreV1().Pods("default").Create(ctx, &pods[i], metav1.CreateOptions{})
		if err != nil {
			t.Fatalf("Failed to create pod: %v", err)
		}
	}
	data.K8sClient = client

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].ENName != "POD_PENDING" {
		t.Errorf("ENName = %v, want %v", issues[0].ENName, "POD_PENDING")
	}
}

func TestPodStatusAnalyzerAnalyze_FailedPods(t *testing.T) {
	a := NewPodStatusAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOnline)
	
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "failed-pod",
			Namespace: "default",
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodFailed,
		},
	}
	data.K8sClient = fake.NewSimpleClientset(pod)

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].ENName != "POD_FAILED" {
		t.Errorf("ENName = %v, want %v", issues[0].ENName, "POD_FAILED")
	}
}

func TestPodStatusAnalyzerAnalyze_HighRestartCount(t *testing.T) {
	a := NewPodStatusAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOnline)
	
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "restarting-pod",
			Namespace: "default",
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			ContainerStatuses: []corev1.ContainerStatus{
				{
					Name:         "app",
					RestartCount: 15,
					State:        corev1.ContainerState{Running: &corev1.ContainerStateRunning{}},
				},
			},
		},
	}
	data.K8sClient = fake.NewSimpleClientset(pod)

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].ENName != "POD_HIGH_RESTARTS" {
		t.Errorf("ENName = %v, want %v", issues[0].ENName, "POD_HIGH_RESTARTS")
	}
	if issues[0].Severity != types.SeverityInfo {
		t.Errorf("Severity = %v, want %v", issues[0].Severity, types.SeverityInfo)
	}
}

func TestPodStatusAnalyzerAnalyze_FilterByNode(t *testing.T) {
	a := NewPodStatusAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOnline)
	data.NodeName = "worker-1"
	
	// Create pods on different nodes
	pod1 := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod-on-worker-1",
			Namespace: "default",
		},
		Spec: corev1.PodSpec{},
		Status: corev1.PodStatus{
			Phase: corev1.PodFailed,
		},
	}
	data.K8sClient = fake.NewSimpleClientset(pod1)

	// With fake client, FieldSelector doesn't work, so we get all pods
	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	// Should detect the failed pod
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
}

// ============ Event Analyzer Tests ============

func TestNewEventAnalyzer(t *testing.T) {
	a := NewEventAnalyzer()
	if a == nil {
		t.Fatal("NewEventAnalyzer() returned nil")
	}
	if a.Name() != "kubernetes.events" {
		t.Errorf("Name() = %v, want %v", a.Name(), "kubernetes.events")
	}
}

func TestEventAnalyzerAnalyze_NoK8sClient(t *testing.T) {
	a := NewEventAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues without K8s client, got %d", len(issues))
	}
}

func TestEventAnalyzerAnalyze_FailedScheduling(t *testing.T) {
	a := NewEventAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOnline)
	
	// Create multiple events with same reason to exceed threshold
	var events []corev1.Event
	for i := 0; i < 5; i++ {
		events = append(events, corev1.Event{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("scheduling-failed-%d", i),
				Namespace: "default",
			},
			Type:   "Warning",
			Reason: "FailedScheduling",
		})
	}
	
	client := fake.NewSimpleClientset()
	for i := range events {
		_, err := client.CoreV1().Events("default").Create(ctx, &events[i], metav1.CreateOptions{})
		if err != nil {
			t.Fatalf("Failed to create event: %v", err)
		}
	}
	data.K8sClient = client

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].ENName != "EVENT_FAILEDSCHEDULING" {
		t.Errorf("ENName = %v, want %v", issues[0].ENName, "EVENT_FAILEDSCHEDULING")
	}
}

func TestEventAnalyzerAnalyze_FailedMount(t *testing.T) {
	a := NewEventAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOnline)
	
	var events []corev1.Event
	for i := 0; i < 5; i++ {
		events = append(events, corev1.Event{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("mount-failed-%d", i),
				Namespace: "default",
			},
			Type:   "Warning",
			Reason: "FailedMount",
		})
	}
	
	client := fake.NewSimpleClientset()
	for i := range events {
		_, err := client.CoreV1().Events("default").Create(ctx, &events[i], metav1.CreateOptions{})
		if err != nil {
			t.Fatalf("Failed to create event: %v", err)
		}
	}
	data.K8sClient = client

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].ENName != "EVENT_FAILEDMOUNT" {
		t.Errorf("ENName = %v, want %v", issues[0].ENName, "EVENT_FAILEDMOUNT")
	}
}

func TestEventAnalyzerAnalyze_NodeNotReady(t *testing.T) {
	a := NewEventAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOnline)
	
	var events []corev1.Event
	for i := 0; i < 5; i++ {
		events = append(events, corev1.Event{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("node-not-ready-%d", i),
				Namespace: "default",
			},
			Type:   "Warning",
			Reason: "NodeNotReady",
		})
	}
	
	client := fake.NewSimpleClientset()
	for i := range events {
		_, err := client.CoreV1().Events("default").Create(ctx, &events[i], metav1.CreateOptions{})
		if err != nil {
			t.Fatalf("Failed to create event: %v", err)
		}
	}
	data.K8sClient = client

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].ENName != "EVENT_NODENOTREADY" {
		t.Errorf("ENName = %v, want %v", issues[0].ENName, "EVENT_NODENOTREADY")
	}
	if issues[0].Severity != types.SeverityCritical {
		t.Errorf("Severity = %v, want %v", issues[0].Severity, types.SeverityCritical)
	}
}

func TestEventAnalyzerAnalyze_NetworkNotReady(t *testing.T) {
	a := NewEventAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOnline)
	
	var events []corev1.Event
	for i := 0; i < 5; i++ {
		events = append(events, corev1.Event{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("network-not-ready-%d", i),
				Namespace: "default",
			},
			Type:   "Warning",
			Reason: "NetworkNotReady",
		})
	}
	
	client := fake.NewSimpleClientset()
	for i := range events {
		_, err := client.CoreV1().Events("default").Create(ctx, &events[i], metav1.CreateOptions{})
		if err != nil {
			t.Fatalf("Failed to create event: %v", err)
		}
	}
	data.K8sClient = client

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].Severity != types.SeverityCritical {
		t.Errorf("Severity = %v, want %v", issues[0].Severity, types.SeverityCritical)
	}
}

func TestEventAnalyzerAnalyze_BelowThreshold(t *testing.T) {
	a := NewEventAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOnline)
	
	// Only 2 events, below threshold of 3
	var events []corev1.Event
	for i := 0; i < 2; i++ {
		events = append(events, corev1.Event{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("scheduling-failed-%d", i),
				Namespace: "default",
			},
			Type:   "Warning",
			Reason: "FailedScheduling",
		})
	}
	
	client := fake.NewSimpleClientset()
	for i := range events {
		_, err := client.CoreV1().Events("default").Create(ctx, &events[i], metav1.CreateOptions{})
		if err != nil {
			t.Fatalf("Failed to create event: %v", err)
		}
	}
	data.K8sClient = client

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues below threshold, got %d", len(issues))
	}
}

func TestEventAnalyzerAnalyze_NormalEvent(t *testing.T) {
	a := NewEventAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOnline)
	
	// Normal event, not Warning type (create multiple to exceed threshold)
	var events []corev1.Event
	for i := 0; i < 10; i++ {
		events = append(events, corev1.Event{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("normal-event-%d", i),
				Namespace: "default",
			},
			Type:   "Normal",
			Reason: "Scheduled",
		})
	}
	
	client := fake.NewSimpleClientset()
	for i := range events {
		_, err := client.CoreV1().Events("default").Create(ctx, &events[i], metav1.CreateOptions{})
		if err != nil {
			t.Fatalf("Failed to create event: %v", err)
		}
	}
	data.K8sClient = client

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues for Normal events, got %d", len(issues))
	}
}

func TestEventAnalyzerAnalyze_MultipleIssues(t *testing.T) {
	a := NewEventAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOnline)
	
	// Create multiple events for each reason to exceed threshold
	client := fake.NewSimpleClientset()
	
	// 5 FailedScheduling events
	for i := 0; i < 5; i++ {
		event := &corev1.Event{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("e1-%d", i),
				Namespace: "default",
			},
			Type:   "Warning",
			Reason: "FailedScheduling",
		}
		_, err := client.CoreV1().Events("default").Create(ctx, event, metav1.CreateOptions{})
		if err != nil {
			t.Fatalf("Failed to create event: %v", err)
		}
	}
	
	// 4 FailedMount events
	for i := 0; i < 4; i++ {
		event := &corev1.Event{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("e2-%d", i),
				Namespace: "default",
			},
			Type:   "Warning",
			Reason: "FailedMount",
		}
		_, err := client.CoreV1().Events("default").Create(ctx, event, metav1.CreateOptions{})
		if err != nil {
			t.Fatalf("Failed to create event: %v", err)
		}
	}
	
	// 6 Unhealthy events
	for i := 0; i < 6; i++ {
		event := &corev1.Event{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("e3-%d", i),
				Namespace: "default",
			},
			Type:   "Warning",
			Reason: "Unhealthy",
		}
		_, err := client.CoreV1().Events("default").Create(ctx, event, metav1.CreateOptions{})
		if err != nil {
			t.Fatalf("Failed to create event: %v", err)
		}
	}
	
	data.K8sClient = client

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 3 {
		t.Fatalf("Expected 3 issues, got %d", len(issues))
	}
}

// ============ Integration Tests ============

func TestAllAnalyzers_Registration(t *testing.T) {
	// Test that all analyzers are registered
	tests := []struct {
		name     string
		analyzer interface{ Name() string }
	}{
		{"PLEG", NewPLEGAnalyzer()},
		{"CNI", NewCNIAnalyzer()},
		{"Certificate", NewCertificateAnalyzer()},
		{"APIServer", NewAPIServerAnalyzer()},
		{"NodeStatus", NewNodeStatusAnalyzer()},
		{"ImagePull", NewImagePullAnalyzer()},
		{"PodStatus", NewPodStatusAnalyzer()},
		{"Event", NewEventAnalyzer()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.analyzer == nil {
				t.Errorf("%s analyzer is nil", tt.name)
			}
		})
	}
}
