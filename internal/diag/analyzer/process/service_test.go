package process

import (
	"context"
	"testing"

	"github.com/kudig-io/klaw/internal/diag/types"
)

func TestNewKubeletAnalyzer(t *testing.T) {
	a := NewKubeletAnalyzer()
	if a == nil {
		t.Fatal("NewKubeletAnalyzer() returned nil")
	}
	if a.Name() != "process.kubelet" {
		t.Errorf("Name() = %v, want %v", a.Name(), "process.kubelet")
	}
}

func TestKubeletAnalyzerAnalyze_NoFile(t *testing.T) {
	a := NewKubeletAnalyzer()
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

func TestKubeletAnalyzerAnalyze_Running(t *testing.T) {
	a := NewKubeletAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"daemon_status/kubelet_status": []byte("Active: active (running)"),
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues for running service, got %d", len(issues))
	}
}

func TestKubeletAnalyzerAnalyze_Failed(t *testing.T) {
	a := NewKubeletAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"daemon_status/kubelet_status": []byte("Active: failed (Result: exit-code)"),
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].ENName != "KUBELET_SERVICE_DOWN" {
		t.Errorf("ENName = %v, want %v", issues[0].ENName, "KUBELET_SERVICE_DOWN")
	}
	if issues[0].Severity != types.SeverityCritical {
		t.Errorf("Severity = %v, want %v", issues[0].Severity, types.SeverityCritical)
	}
}

func TestKubeletAnalyzerAnalyze_Stopped(t *testing.T) {
	a := NewKubeletAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"daemon_status/kubelet_status": []byte("Active: inactive (dead)"),
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].ENName != "KUBELET_SERVICE_STOPPED" {
		t.Errorf("ENName = %v, want %v", issues[0].ENName, "KUBELET_SERVICE_STOPPED")
	}
}

func TestNewContainerRuntimeAnalyzer(t *testing.T) {
	a := NewContainerRuntimeAnalyzer()
	if a == nil {
		t.Fatal("NewContainerRuntimeAnalyzer() returned nil")
	}
	if a.Name() != "process.container_runtime" {
		t.Errorf("Name() = %v, want %v", a.Name(), "process.container_runtime")
	}
}

func TestContainerRuntimeAnalyzerAnalyze_NoIssues(t *testing.T) {
	a := NewContainerRuntimeAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"daemon_status/docker_status":    []byte("Active: active (running)"),
		"daemon_status/containerd_status": []byte("Active: active (running)"),
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues, got %d", len(issues))
	}
}

func TestParseServiceStatus(t *testing.T) {
	tests := []struct {
		content  string
		expected string
	}{
		{"Active: active (running)", "running"},
		{"Active: failed (Result: exit-code)", "failed"},
		{"Active: inactive (dead)", "stopped"},
		{"Active: inactive", "stopped"},
		{"No status info", "unknown"},
	}

	for _, tt := range tests {
		result := parseServiceStatus(tt.content)
		if result != tt.expected {
			t.Errorf("parseServiceStatus(%q) = %v, want %v", tt.content, result, tt.expected)
		}
	}
}

func TestNewFirewalldAnalyzer(t *testing.T) {
	a := NewFirewalldAnalyzer()
	if a == nil {
		t.Fatal("NewFirewalldAnalyzer() returned nil")
	}
	if a.Name() != "process.firewalld" {
		t.Errorf("Name() = %v, want %v", a.Name(), "process.firewalld")
	}
}

func TestNewPIDLeakAnalyzer(t *testing.T) {
	a := NewPIDLeakAnalyzer()
	if a == nil {
		t.Fatal("NewPIDLeakAnalyzer() returned nil")
	}
	if a.Name() != "process.pid_leak" {
		t.Errorf("Name() = %v, want %v", a.Name(), "process.pid_leak")
	}
}

func TestContainerRuntimeAnalyzer_BothFailed(t *testing.T) {
	a := NewContainerRuntimeAnalyzer()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"daemon_status/docker_status":    []byte("Active: failed"),
		"daemon_status/containerd_status": []byte("Active: failed"),
	}
	issues, err := a.Analyze(context.Background(), data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].ENName != "CONTAINER_RUNTIME_DOWN" {
		t.Errorf("ENName = %v, want CONTAINER_RUNTIME_DOWN", issues[0].ENName)
	}
}

func TestContainerRuntimeAnalyzer_BothStopped(t *testing.T) {
	a := NewContainerRuntimeAnalyzer()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"daemon_status/docker_status":    []byte("Active: inactive (dead)"),
		"daemon_status/containerd_status": []byte("Active: inactive (dead)"),
	}
	issues, err := a.Analyze(context.Background(), data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].ENName != "CONTAINER_RUNTIME_STOPPED" {
		t.Errorf("ENName = %v, want CONTAINER_RUNTIME_STOPPED", issues[0].ENName)
	}
}

func TestContainerRuntimeAnalyzer_NoFiles(t *testing.T) {
	a := NewContainerRuntimeAnalyzer()
	data := types.NewDiagnosticData(types.ModeOffline)
	issues, err := a.Analyze(context.Background(), data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues, got %d", len(issues))
	}
}

func TestContainerRuntimeAnalyzer_OneRunning(t *testing.T) {
	a := NewContainerRuntimeAnalyzer()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"daemon_status/docker_status":    []byte("Active: active (running)"),
		"daemon_status/containerd_status": []byte("Active: failed"),
	}
	issues, err := a.Analyze(context.Background(), data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues when one runtime is running, got %d", len(issues))
	}
}

func TestNewRuncAnalyzer(t *testing.T) {
	a := NewRuncAnalyzer()
	if a == nil {
		t.Fatal("NewRuncAnalyzer() returned nil")
	}
	if a.Name() != "process.runc" {
		t.Errorf("Name() = %v, want process.runc", a.Name())
	}
}

func TestRuncAnalyzer_NoFile(t *testing.T) {
	a := NewRuncAnalyzer()
	data := types.NewDiagnosticData(types.ModeOffline)
	issues, err := a.Analyze(context.Background(), data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues, got %d", len(issues))
	}
}

func TestRuncAnalyzer_HangDetected(t *testing.T) {
	a := NewRuncAnalyzer()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"system_status": []byte("runc process [1234] maybe hang\nrunc process [5678] maybe hang"),
	}
	issues, err := a.Analyze(context.Background(), data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].ENName != "RUNC_PROCESS_HANG" {
		t.Errorf("ENName = %v, want RUNC_PROCESS_HANG", issues[0].ENName)
	}
}

func TestRuncAnalyzer_NoHang(t *testing.T) {
	a := NewRuncAnalyzer()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"system_status": []byte("runc process [1234] running normally"),
	}
	issues, err := a.Analyze(context.Background(), data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues, got %d", len(issues))
	}
}

func TestFirewalldAnalyzer_NoFile(t *testing.T) {
	a := NewFirewalldAnalyzer()
	data := types.NewDiagnosticData(types.ModeOffline)
	issues, err := a.Analyze(context.Background(), data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues, got %d", len(issues))
	}
}

func TestFirewalldAnalyzer_Running(t *testing.T) {
	a := NewFirewalldAnalyzer()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"service_status": []byte("firewalld -- running"),
	}
	issues, err := a.Analyze(context.Background(), data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].ENName != "FIREWALLD_RUNNING" {
		t.Errorf("ENName = %v, want FIREWALLD_RUNNING", issues[0].ENName)
	}
}

func TestFirewalldAnalyzer_Stopped(t *testing.T) {
	a := NewFirewalldAnalyzer()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"service_status": []byte("firewalld -- stopped"),
	}
	issues, err := a.Analyze(context.Background(), data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues when firewalld stopped, got %d", len(issues))
	}
}

func TestPIDLeakAnalyzer_NoFile(t *testing.T) {
	a := NewPIDLeakAnalyzer()
	data := types.NewDiagnosticData(types.ModeOffline)
	issues, err := a.Analyze(context.Background(), data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues, got %d", len(issues))
	}
}

func TestPIDLeakAnalyzer_CriticalLeak(t *testing.T) {
	a := NewPIDLeakAnalyzer()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"system_status": []byte("15000  process pid leak detect\n8000  other pid leak detect"),
	}
	issues, err := a.Analyze(context.Background(), data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].ENName != "PID_LEAK_DETECTED" {
		t.Errorf("ENName = %v, want PID_LEAK_DETECTED", issues[0].ENName)
	}
	if issues[0].Severity != types.SeverityCritical {
		t.Errorf("Severity = %v, want Critical", issues[0].Severity)
	}
}

func TestPIDLeakAnalyzer_WarningLeak(t *testing.T) {
	a := NewPIDLeakAnalyzer()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"system_status": []byte("6000  process pid leak detect"),
	}
	issues, err := a.Analyze(context.Background(), data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].ENName != "HIGH_THREAD_COUNT" {
		t.Errorf("ENName = %v, want HIGH_THREAD_COUNT", issues[0].ENName)
	}
}

func TestPIDLeakAnalyzer_Normal(t *testing.T) {
	a := NewPIDLeakAnalyzer()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"system_status": []byte("500  process pid leak detect"),
	}
	issues, err := a.Analyze(context.Background(), data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues for normal thread count, got %d", len(issues))
	}
}

func TestParseServiceStatus_EdgeCases(t *testing.T) {
	tests := []struct {
		content  string
		expected string
	}{
		{"active: active (running)", "running"},
		{"Status=1/Failure", "failed"},
		{"active: inactive", "stopped"},
	}
	for _, tt := range tests {
		result := parseServiceStatus(tt.content)
		if result != tt.expected {
			t.Errorf("parseServiceStatus(%q) = %v, want %v", tt.content, result, tt.expected)
		}
	}
}
