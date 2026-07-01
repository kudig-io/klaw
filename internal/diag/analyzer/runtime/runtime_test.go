package runtime

import (
	"context"
	"testing"

	"github.com/kudig-io/klaw/internal/diag/types"
)

func TestNewDockerAnalyzer(t *testing.T) {
	a := NewDockerAnalyzer()
	if a == nil {
		t.Fatal("NewDockerAnalyzer() returned nil")
	}
	if a.Name() != "runtime.docker" {
		t.Errorf("Name() = %v, want %v", a.Name(), "runtime.docker")
	}
}

func TestDockerAnalyzerAnalyze_NoFile(t *testing.T) {
	a := NewDockerAnalyzer()
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

func TestDockerAnalyzerAnalyze_NoIssues(t *testing.T) {
	a := NewDockerAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"logs/docker.log": []byte("Docker daemon started successfully"),
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues, got %d", len(issues))
	}
}

func TestDockerAnalyzerAnalyze_StartFailed(t *testing.T) {
	a := NewDockerAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"logs/docker.log": []byte("Failed to start Docker daemon"),
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].ENName != "DOCKER_START_FAILED" {
		t.Errorf("ENName = %v, want %v", issues[0].ENName, "DOCKER_START_FAILED")
	}
	if issues[0].Severity != types.SeverityCritical {
		t.Errorf("Severity = %v, want %v", issues[0].Severity, types.SeverityCritical)
	}
}

func TestDockerAnalyzerAnalyze_StorageDriverError(t *testing.T) {
	a := NewDockerAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"logs/docker.log": []byte("Storage driver error occurred"),
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].ENName != "DOCKER_STORAGE_DRIVER_ERROR" {
		t.Errorf("ENName = %v, want %v", issues[0].ENName, "DOCKER_STORAGE_DRIVER_ERROR")
	}
}

func TestNewContainerdAnalyzer(t *testing.T) {
	a := NewContainerdAnalyzer()
	if a == nil {
		t.Fatal("NewContainerdAnalyzer() returned nil")
	}
	if a.Name() != "runtime.containerd" {
		t.Errorf("Name() = %v, want %v", a.Name(), "runtime.containerd")
	}
}

func TestContainerdAnalyzerAnalyze_NoFile(t *testing.T) {
	a := NewContainerdAnalyzer()
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

func TestContainerdAnalyzerAnalyze_CreateFailed(t *testing.T) {
	a := NewContainerdAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	// Need more than 10 occurrences to trigger
	content := ""
	for i := 0; i < 12; i++ {
		content += "failed to create container "
	}
	data.RawFiles = map[string][]byte{
		"logs/containerd.log": []byte(content),
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].ENName != "CONTAINER_CREATE_FAILED" {
		t.Errorf("ENName = %v, want %v", issues[0].ENName, "CONTAINER_CREATE_FAILED")
	}
}

func TestNewTimeSyncAnalyzer(t *testing.T) {
	a := NewTimeSyncAnalyzer()
	if a == nil {
		t.Fatal("NewTimeSyncAnalyzer() returned nil")
	}
	if a.Name() != "runtime.time_sync" {
		t.Errorf("Name() = %v, want %v", a.Name(), "runtime.time_sync")
	}
}

func TestNewConfigAnalyzer(t *testing.T) {
	a := NewConfigAnalyzer()
	if a == nil {
		t.Fatal("NewConfigAnalyzer() returned nil")
	}
	if a.Name() != "runtime.config" {
		t.Errorf("Name() = %v, want %v", a.Name(), "runtime.config")
	}
}

func TestTimeSyncAnalyzer_NoFile(t *testing.T) {
	a := NewTimeSyncAnalyzer()
	data := types.NewDiagnosticData(types.ModeOffline)
	issues, err := a.Analyze(context.Background(), data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues, got %d", len(issues))
	}
}

func TestTimeSyncAnalyzer_NotRunning(t *testing.T) {
	a := NewTimeSyncAnalyzer()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"service_status": []byte("sshd -- running\nfirewalld -- stopped"),
	}
	issues, err := a.Analyze(context.Background(), data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].ENName != "TIME_SYNC_SERVICE_DOWN" {
		t.Errorf("ENName = %v, want TIME_SYNC_SERVICE_DOWN", issues[0].ENName)
	}
}

func TestTimeSyncAnalyzer_NtpdRunning(t *testing.T) {
	a := NewTimeSyncAnalyzer()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"service_status": []byte("ntpd -- running"),
	}
	issues, err := a.Analyze(context.Background(), data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues when ntpd running, got %d", len(issues))
	}
}

func TestTimeSyncAnalyzer_ChronydRunning(t *testing.T) {
	a := NewTimeSyncAnalyzer()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"service_status": []byte("chronyd -- running"),
	}
	issues, err := a.Analyze(context.Background(), data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues when chronyd running, got %d", len(issues))
	}
}

func TestConfigAnalyzer_NoFile(t *testing.T) {
	a := NewConfigAnalyzer()
	data := types.NewDiagnosticData(types.ModeOffline)
	issues, err := a.Analyze(context.Background(), data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues, got %d", len(issues))
	}
}

func TestConfigAnalyzer_IPForwardDisabled(t *testing.T) {
	a := NewConfigAnalyzer()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"system_info": []byte("net.ipv4.ip_forward = 0"),
	}
	issues, err := a.Analyze(context.Background(), data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	found := false
	for _, issue := range issues {
		if issue.ENName == "IP_FORWARD_DISABLED" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected IP_FORWARD_DISABLED issue")
	}
}

func TestConfigAnalyzer_BridgeNfCallDisabled(t *testing.T) {
	a := NewConfigAnalyzer()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"system_info": []byte("net.bridge.bridge-nf-call-iptables = 0"),
	}
	issues, err := a.Analyze(context.Background(), data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	found := false
	for _, issue := range issues {
		if issue.ENName == "BRIDGE_NF_CALL_IPTABLES_DISABLED" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected BRIDGE_NF_CALL_IPTABLES_DISABLED issue")
	}
}

func TestConfigAnalyzer_LowUlimit(t *testing.T) {
	a := NewConfigAnalyzer()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"system_info": []byte("open files 1024"),
	}
	issues, err := a.Analyze(context.Background(), data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	found := false
	for _, issue := range issues {
		if issue.ENName == "LOW_ULIMIT_NOFILE" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected LOW_ULIMIT_NOFILE issue")
	}
}

func TestConfigAnalyzer_SelinuxEnforcing(t *testing.T) {
	a := NewConfigAnalyzer()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"system_info": []byte("SELinux status: enforcing"),
	}
	issues, err := a.Analyze(context.Background(), data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	found := false
	for _, issue := range issues {
		if issue.ENName == "SELINUX_ENFORCING" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected SELINUX_ENFORCING issue")
	}
}

func TestConfigAnalyzer_AllOK(t *testing.T) {
	a := NewConfigAnalyzer()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"system_info": []byte("net.ipv4.ip_forward = 1\nnet.bridge.bridge-nf-call-iptables = 1\nopen files 65536"),
	}
	issues, err := a.Analyze(context.Background(), data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues for good config, got %d", len(issues))
	}
}

func TestContainerdAnalyzer_FewFailures(t *testing.T) {
	a := NewContainerdAnalyzer()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"logs/containerd.log": []byte("failed to create container failed to create container"),
	}
	issues, err := a.Analyze(context.Background(), data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues for few failures, got %d", len(issues))
	}
}

func TestDockerAnalyzer_BothIssues(t *testing.T) {
	a := NewDockerAnalyzer()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"logs/docker.log": []byte("Failed to start Docker\nStorage driver error found"),
	}
	issues, err := a.Analyze(context.Background(), data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 2 {
		t.Fatalf("Expected 2 issues, got %d", len(issues))
	}
}
