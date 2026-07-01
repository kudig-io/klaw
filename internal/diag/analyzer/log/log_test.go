package log

import (
	"context"
	"strings"
	"testing"

	"github.com/kudig-io/klaw/internal/diag/types"
)

// ============ Syslog Analyzer Tests ============

func TestNewSyslogAnalyzer(t *testing.T) {
	a := NewSyslogAnalyzer()
	if a == nil {
		t.Fatal("NewSyslogAnalyzer() returned nil")
	}
	if a.Name() != "log.syslog" {
		t.Errorf("Name() = %v, want %v", a.Name(), "log.syslog")
	}
}

func TestSyslogAnalyzerAnalyze_NoData(t *testing.T) {
	a := NewSyslogAnalyzer()
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

func TestSyslogAnalyzerAnalyze_Segfault(t *testing.T) {
	a := NewSyslogAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"logs/syslog": []byte("kernel: segfault at 0 ip 00007f... error 4 in libc"),
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].ENName != "SYSLOG_SEGFAULT" {
		t.Errorf("ENName = %v, want %v", issues[0].ENName, "SYSLOG_SEGFAULT")
	}
	if issues[0].Severity != types.SeverityCritical {
		t.Errorf("Severity = %v, want %v", issues[0].Severity, types.SeverityCritical)
	}
}

func TestSyslogAnalyzerAnalyze_KernelPanic(t *testing.T) {
	a := NewSyslogAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"logs/messages": []byte("kernel panic: VFS: Unable to mount root fs"),
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].ENName != "SYSLOG_KERNEL_PANIC" {
		t.Errorf("ENName = %v, want %v", issues[0].ENName, "SYSLOG_KERNEL_PANIC")
	}
}

func TestSyslogAnalyzerAnalyze_OOMKill(t *testing.T) {
	a := NewSyslogAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"logs/syslog": []byte("Out of memory: Killed process 1234 (java)"),
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].ENName != "SYSLOG_OOM_KILL" {
		t.Errorf("ENName = %v, want %v", issues[0].ENName, "SYSLOG_OOM_KILL")
	}
}

func TestSyslogAnalyzerAnalyze_IOError(t *testing.T) {
	a := NewSyslogAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"logs/syslog": []byte("I/O error, dev sda, sector 12345678"),
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].ENName != "SYSLOG_IO_ERROR" {
		t.Errorf("ENName = %v, want %v", issues[0].ENName, "SYSLOG_IO_ERROR")
	}
}

func TestSyslogAnalyzerAnalyze_NFSErrors(t *testing.T) {
	a := NewSyslogAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	// Create content with 3 NFS errors (threshold)
	content := strings.Repeat("NFS server not responding, still trying\n", 3)
	data.RawFiles = map[string][]byte{
		"logs/syslog": []byte(content),
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].ENName != "SYSLOG_NFS_ERROR" {
		t.Errorf("ENName = %v, want %v", issues[0].ENName, "SYSLOG_NFS_ERROR")
	}
}

func TestSyslogAnalyzerAnalyze_NFSErrors_BelowThreshold(t *testing.T) {
	a := NewSyslogAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	// Create content with only 2 NFS errors (below threshold)
	content := strings.Repeat("NFS server not responding\n", 2)
	data.RawFiles = map[string][]byte{
		"logs/syslog": []byte(content),
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues below threshold, got %d", len(issues))
	}
}

func TestSyslogAnalyzerAnalyze_ConnectionErrors(t *testing.T) {
	a := NewSyslogAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	// Create content with 5 connection errors (threshold)
	content := strings.Repeat("Connection refused\n", 5)
	data.RawFiles = map[string][]byte{
		"logs/syslog": []byte(content),
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].ENName != "SYSLOG_CONNECTION_ERROR" {
		t.Errorf("ENName = %v, want %v", issues[0].ENName, "SYSLOG_CONNECTION_ERROR")
	}
}

func TestSyslogAnalyzerAnalyze_AuthFailures(t *testing.T) {
	a := NewSyslogAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	// Create content with 5 auth failures (threshold)
	content := strings.Repeat("authentication failure\n", 5)
	data.RawFiles = map[string][]byte{
		"logs/syslog": []byte(content),
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].ENName != "SYSLOG_AUTH_FAILURE" {
		t.Errorf("ENName = %v, want %v", issues[0].ENName, "SYSLOG_AUTH_FAILURE")
	}
}

func TestSyslogAnalyzerAnalyze_ServiceFailed(t *testing.T) {
	a := NewSyslogAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	// Create content with 3 service failures (threshold)
	content := strings.Repeat("service failed to start\n", 3)
	data.RawFiles = map[string][]byte{
		"logs/syslog": []byte(content),
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].ENName != "SYSLOG_SERVICE_FAILED" {
		t.Errorf("ENName = %v, want %v", issues[0].ENName, "SYSLOG_SERVICE_FAILED")
	}
}

func TestSyslogAnalyzerAnalyze_MultipleIssues(t *testing.T) {
	a := NewSyslogAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"logs/syslog": []byte(
			"segfault at 0\n" +
			"Out of memory: Killed process 1234\n" +
			"I/O error, dev sda\n",
		),
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 3 {
		t.Fatalf("Expected 3 issues, got %d", len(issues))
	}
}

func TestSyslogAnalyzerAnalyze_VarLogMessages(t *testing.T) {
	a := NewSyslogAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"varlogmessages.log": []byte("kernel: segfault at 0"),
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
}

// ============ JournalCtl Analyzer Tests ============

func TestNewJournalCtlAnalyzer(t *testing.T) {
	a := NewJournalCtlAnalyzer()
	if a == nil {
		t.Fatal("NewJournalCtlAnalyzer() returned nil")
	}
	if a.Name() != "log.journalctl" {
		t.Errorf("Name() = %v, want %v", a.Name(), "log.journalctl")
	}
}

func TestJournalCtlAnalyzerAnalyze_NoData(t *testing.T) {
	a := NewJournalCtlAnalyzer()
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

func TestJournalCtlAnalyzerAnalyze_ServiceFailed(t *testing.T) {
	a := NewJournalCtlAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	content := strings.Repeat("Failed to start nginx.service\n", 3)
	data.RawFiles = map[string][]byte{
		"logs/journalctl.log": []byte(content),
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].ENName != "JOURNAL_SERVICE_FAILED" {
		t.Errorf("ENName = %v, want %v", issues[0].ENName, "JOURNAL_SERVICE_FAILED")
	}
}

func TestJournalCtlAnalyzerAnalyze_DependencyFailed(t *testing.T) {
	a := NewJournalCtlAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	content := strings.Repeat("Dependency failed for mysql.service\n", 3)
	data.RawFiles = map[string][]byte{
		"logs/journalctl.log": []byte(content),
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].ENName != "JOURNAL_DEPENDENCY_FAILED" {
		t.Errorf("ENName = %v, want %v", issues[0].ENName, "JOURNAL_DEPENDENCY_FAILED")
	}
}

func TestJournalCtlAnalyzerAnalyze_WatchdogTimeout(t *testing.T) {
	a := NewJournalCtlAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"logs/journalctl.log": []byte("watchdog timeout, service didn't stop"),
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].ENName != "JOURNAL_WATCHDOG_TIMEOUT" {
		t.Errorf("ENName = %v, want %v", issues[0].ENName, "JOURNAL_WATCHDOG_TIMEOUT")
	}
	if issues[0].Severity != types.SeverityCritical {
		t.Errorf("Severity = %v, want %v", issues[0].Severity, types.SeverityCritical)
	}
}

func TestJournalCtlAnalyzerAnalyze_Coredump(t *testing.T) {
	a := NewJournalCtlAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"logs/journalctl.log": []byte("Process 1234 dumped core"),
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].ENName != "JOURNAL_COREDUMP" {
		t.Errorf("ENName = %v, want %v", issues[0].ENName, "JOURNAL_COREDUMP")
	}
}

func TestJournalCtlAnalyzerAnalyze_FatalError(t *testing.T) {
	a := NewJournalCtlAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"logs/journalctl.log": []byte("FATAL: could not connect to database"),
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].ENName != "JOURNAL_FATAL_ERROR" {
		t.Errorf("ENName = %v, want %v", issues[0].ENName, "JOURNAL_FATAL_ERROR")
	}
}

func TestJournalCtlAnalyzerAnalyze_ResourceExhausted(t *testing.T) {
	a := NewJournalCtlAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	content := strings.Repeat("Resource temporarily unavailable\n", 3)
	data.RawFiles = map[string][]byte{
		"logs/journalctl.log": []byte(content),
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].ENName != "JOURNAL_RESOURCE_EXHAUSTED" {
		t.Errorf("ENName = %v, want %v", issues[0].ENName, "JOURNAL_RESOURCE_EXHAUSTED")
	}
}

// ============ Kubelet Log Analyzer Tests ============

func TestNewKubeletLogAnalyzer(t *testing.T) {
	a := NewKubeletLogAnalyzer()
	if a == nil {
		t.Fatal("NewKubeletLogAnalyzer() returned nil")
	}
	if a.Name() != "log.kubelet" {
		t.Errorf("Name() = %v, want %v", a.Name(), "log.kubelet")
	}
}

func TestKubeletLogAnalyzerAnalyze_NoData(t *testing.T) {
	a := NewKubeletLogAnalyzer()
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

func TestKubeletLogAnalyzerAnalyze_PLEGUnhealthy(t *testing.T) {
	a := NewKubeletLogAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"logs/kubelet.log": []byte("PLEG is not healthy"),
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
}

func TestKubeletLogAnalyzerAnalyze_RuntimeDown(t *testing.T) {
	a := NewKubeletLogAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"logs/kubelet.log": []byte("Container runtime is down"),
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].ENName != "KUBELET_RUNTIME_DOWN" {
		t.Errorf("ENName = %v, want %v", issues[0].ENName, "KUBELET_RUNTIME_DOWN")
	}
}

func TestKubeletLogAnalyzerAnalyze_SandboxFailed(t *testing.T) {
	a := NewKubeletLogAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	content := strings.Repeat("Failed to create pod sandbox\n", 3)
	data.RawFiles = map[string][]byte{
		"logs/kubelet.log": []byte(content),
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].ENName != "KUBELET_SANDBOX_FAILED" {
		t.Errorf("ENName = %v, want %v", issues[0].ENName, "KUBELET_SANDBOX_FAILED")
	}
}

func TestKubeletLogAnalyzerAnalyze_VolumeMountFailed(t *testing.T) {
	a := NewKubeletLogAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	content := strings.Repeat("Volume failed to mount\n", 3)
	data.RawFiles = map[string][]byte{
		"logs/kubelet.log": []byte(content),
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].ENName != "KUBELET_VOLUME_MOUNT_FAILED" {
		t.Errorf("ENName = %v, want %v", issues[0].ENName, "KUBELET_VOLUME_MOUNT_FAILED")
	}
}

func TestKubeletLogAnalyzerAnalyze_EvictionTriggered(t *testing.T) {
	a := NewKubeletLogAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"logs/kubelet.log": []byte("Eviction manager: attempting to reclaim memory"),
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].ENName != "KUBELET_EVICTION_TRIGGERED" {
		t.Errorf("ENName = %v, want %v", issues[0].ENName, "KUBELET_EVICTION_TRIGGERED")
	}
}

func TestKubeletLogAnalyzerAnalyze_ImagePullFailed(t *testing.T) {
	a := NewKubeletLogAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	content := strings.Repeat("Failed to pull image nginx:latest\n", 5)
	data.RawFiles = map[string][]byte{
		"logs/kubelet.log": []byte(content),
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].ENName != "KUBELET_IMAGE_PULL_FAILED" {
		t.Errorf("ENName = %v, want %v", issues[0].ENName, "KUBELET_IMAGE_PULL_FAILED")
	}
}

func TestKubeletLogAnalyzerAnalyze_DaemonStatus(t *testing.T) {
	a := NewKubeletLogAnalyzer()
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
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
}

// ============ Integration Tests ============

func TestAllLogAnalyzers_Registration(t *testing.T) {
	tests := []struct {
		name     string
		analyzer interface{ Name() string }
	}{
		{"Syslog", NewSyslogAnalyzer()},
		{"JournalCtl", NewJournalCtlAnalyzer()},
		{"KubeletLog", NewKubeletLogAnalyzer()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.analyzer == nil {
				t.Errorf("%s analyzer is nil", tt.name)
			}
		})
	}
}
