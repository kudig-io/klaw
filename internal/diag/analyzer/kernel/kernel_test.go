package kernel

import (
	"context"
	"testing"

	"github.com/kudig-io/klaw/internal/diag/types"
)

func TestNewPanicAnalyzer(t *testing.T) {
	a := NewPanicAnalyzer()
	if a == nil {
		t.Fatal("NewPanicAnalyzer() returned nil")
	}
	if a.Name() != "kernel.panic" {
		t.Errorf("Name() = %v, want %v", a.Name(), "kernel.panic")
	}
}

func TestPanicAnalyzerAnalyze_NoFile(t *testing.T) {
	a := NewPanicAnalyzer()
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

func TestPanicAnalyzerAnalyze_NoIssue(t *testing.T) {
	a := NewPanicAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"logs/dmesg.log": []byte("normal kernel operations"),
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues, got %d", len(issues))
	}
}

func TestPanicAnalyzerAnalyze_WithPanic(t *testing.T) {
	a := NewPanicAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"logs/dmesg.log": []byte("Kernel panic - not syncing: VFS: Unable to mount root fs"),
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].ENName != "KERNEL_PANIC" {
		t.Errorf("ENName = %v, want %v", issues[0].ENName, "KERNEL_PANIC")
	}
	if issues[0].Severity != types.SeverityCritical {
		t.Errorf("Severity = %v, want %v", issues[0].Severity, types.SeverityCritical)
	}
}

func TestPanicAnalyzerAnalyze_VarLogMessage(t *testing.T) {
	a := NewPanicAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"varlogmessage.log": []byte("Kernel panic occurred"),
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
}

func TestNewOOMAnalyzer(t *testing.T) {
	a := NewOOMAnalyzer()
	if a == nil {
		t.Fatal("NewOOMAnalyzer() returned nil")
	}
	if a.Name() != "kernel.oom" {
		t.Errorf("Name() = %v, want %v", a.Name(), "kernel.oom")
	}
}

func TestOOMAnalyzerAnalyze_NoIssue(t *testing.T) {
	a := NewOOMAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"logs/dmesg.log": []byte("normal operations"),
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues, got %d", len(issues))
	}
}

func TestOOMAnalyzerAnalyze_WithOOM(t *testing.T) {
	a := NewOOMAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"logs/dmesg.log": []byte("Out of memory: Kill process 1234 (java) score 999"),
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].ENName != "KERNEL_OOM_KILLER" {
		t.Errorf("ENName = %v, want %v", issues[0].ENName, "KERNEL_OOM_KILLER")
	}
}

func TestOOMAnalyzerAnalyze_MultipleSources(t *testing.T) {
	a := NewOOMAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	// OOM in messages file
	data.RawFiles = map[string][]byte{
		"logs/messages": []byte("Out of memory:"),
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
}

func TestNewFilesystemAnalyzer(t *testing.T) {
	a := NewFilesystemAnalyzer()
	if a == nil {
		t.Fatal("NewFilesystemAnalyzer() returned nil")
	}
	if a.Name() != "kernel.filesystem" {
		t.Errorf("Name() = %v, want %v", a.Name(), "kernel.filesystem")
	}
}

func TestFilesystemAnalyzerAnalyze_NoIssue(t *testing.T) {
	a := NewFilesystemAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"logs/dmesg.log": []byte("normal filesystem operations"),
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues, got %d", len(issues))
	}
}

func TestFilesystemAnalyzerAnalyze_ReadOnly(t *testing.T) {
	a := NewFilesystemAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"logs/dmesg.log": []byte("Remounting filesystem read-only due to error"),
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].ENName != "FILESYSTEM_READONLY" {
		t.Errorf("ENName = %v, want %v", issues[0].ENName, "FILESYSTEM_READONLY")
	}
}

func TestFilesystemAnalyzerAnalyze_IOError(t *testing.T) {
	a := NewFilesystemAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	// Create content with more than 10 I/O errors
	content := ""
	for i := 0; i < 15; i++ {
		content += "I/O error on device sda1\n"
	}
	data.RawFiles = map[string][]byte{
		"logs/dmesg.log": []byte(content),
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].ENName != "DISK_IO_ERROR" {
		t.Errorf("ENName = %v, want %v", issues[0].ENName, "DISK_IO_ERROR")
	}
}

func TestNewModuleAnalyzer(t *testing.T) {
	a := NewModuleAnalyzer()
	if a == nil {
		t.Fatal("NewModuleAnalyzer() returned nil")
	}
	if a.Name() != "kernel.module" {
		t.Errorf("Name() = %v, want %v", a.Name(), "kernel.module")
	}
}
