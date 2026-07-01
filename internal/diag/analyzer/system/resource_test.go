package system

import (
	"context"
	"strings"
	"testing"

	"github.com/kudig-io/klaw/internal/diag/types"
)

// ============ CPU Analyzer Tests ============

func TestNewCPUAnalyzer(t *testing.T) {
	a := NewCPUAnalyzer()
	if a == nil {
		t.Fatal("NewCPUAnalyzer() returned nil")
	}
	if a.Name() != "system.cpu" {
		t.Errorf("Name() = %v, want %v", a.Name(), "system.cpu")
	}
	if a.Category() != "system" {
		t.Errorf("Category() = %v, want %v", a.Category(), "system")
	}
}

func TestCPUAnalyzerAnalyze_NoData(t *testing.T) {
	a := NewCPUAnalyzer()
	ctx := context.Background()
	data := &types.DiagnosticData{
		SystemMetrics: nil,
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues, got %d", len(issues))
	}
}

func TestCPUAnalyzerAnalyze_NormalLoad(t *testing.T) {
	a := NewCPUAnalyzer()
	ctx := context.Background()
	data := &types.DiagnosticData{
		SystemMetrics: &types.SystemMetrics{
			CPUCores:     4,
			LoadAvg15Min: 2.0, // Below 2*cores (8) threshold
		},
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues for normal load, got %d", len(issues))
	}
}

func TestCPUAnalyzerAnalyze_WarningLoad(t *testing.T) {
	a := NewCPUAnalyzer()
	ctx := context.Background()
	data := &types.DiagnosticData{
		SystemMetrics: &types.SystemMetrics{
			CPUCores:     4,
			LoadAvg15Min: 10.0, // Between 2*cores (8) and 4*cores (16)
		},
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].Severity != types.SeverityWarning {
		t.Errorf("Expected Warning severity, got %v", issues[0].Severity)
	}
	if issues[0].ENName != "ELEVATED_SYSTEM_LOAD" {
		t.Errorf("Expected ELEVATED_SYSTEM_LOAD, got %v", issues[0].ENName)
	}
}

func TestCPUAnalyzerAnalyze_CriticalLoad(t *testing.T) {
	a := NewCPUAnalyzer()
	ctx := context.Background()
	data := &types.DiagnosticData{
		SystemMetrics: &types.SystemMetrics{
			CPUCores:     4,
			LoadAvg15Min: 20.0, // Above 4*cores (16)
		},
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].Severity != types.SeverityCritical {
		t.Errorf("Expected Critical severity, got %v", issues[0].Severity)
	}
	if issues[0].ENName != "HIGH_SYSTEM_LOAD" {
		t.Errorf("Expected HIGH_SYSTEM_LOAD, got %v", issues[0].ENName)
	}
}

func TestCPUAnalyzerAnalyze_DefaultCores(t *testing.T) {
	a := NewCPUAnalyzer()
	ctx := context.Background()
	data := &types.DiagnosticData{
		SystemMetrics: &types.SystemMetrics{
			CPUCores:     0,    // Will default to 4
			LoadAvg15Min: 10.0, // Should trigger warning with default cores
		},
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	// With default 4 cores, 10.0 load should be warning (between 8 and 16)
	if issues[0].Severity != types.SeverityWarning {
		t.Errorf("Expected Warning severity, got %v", issues[0].Severity)
	}
}

func TestCPUAnalyzerAnalyze_ExactWarningThreshold(t *testing.T) {
	a := NewCPUAnalyzer()
	ctx := context.Background()
	// Exactly at 2*cores threshold (8.0)
	data := &types.DiagnosticData{
		SystemMetrics: &types.SystemMetrics{
			CPUCores:     4,
			LoadAvg15Min: 8.0,
		},
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	// Exactly at threshold should not trigger (needs to be >)
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues at exact threshold, got %d", len(issues))
	}
}

func TestCPUAnalyzerAnalyze_JustAboveWarning(t *testing.T) {
	a := NewCPUAnalyzer()
	ctx := context.Background()
	data := &types.DiagnosticData{
		SystemMetrics: &types.SystemMetrics{
			CPUCores:     4,
			LoadAvg15Min: 8.1, // Just above warning threshold
		},
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].Severity != types.SeverityWarning {
		t.Errorf("Expected Warning severity, got %v", issues[0].Severity)
	}
}

func TestCPUAnalyzerAnalyze_JustAboveCriticalThreshold(t *testing.T) {
	a := NewCPUAnalyzer()
	ctx := context.Background()
	// Just above 4*cores threshold (16.0)
	data := &types.DiagnosticData{
		SystemMetrics: &types.SystemMetrics{
			CPUCores:     4,
			LoadAvg15Min: 16.1,
		},
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	// Above critical threshold should trigger critical
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue above critical threshold, got %d", len(issues))
	}
	if issues[0].Severity != types.SeverityCritical {
		t.Errorf("Expected Critical severity, got %v", issues[0].Severity)
	}
}

func TestCPUAnalyzerAnalyze_HighCoreCount(t *testing.T) {
	a := NewCPUAnalyzer()
	ctx := context.Background()
	// 32-core system with moderate load
	data := &types.DiagnosticData{
		SystemMetrics: &types.SystemMetrics{
			CPUCores:     32,
			LoadAvg15Min: 70.0, // Between 2*32=64 and 4*32=128
		},
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].Severity != types.SeverityWarning {
		t.Errorf("Expected Warning severity, got %v", issues[0].Severity)
	}
}

// ============ Memory Analyzer Tests ============

func TestNewMemoryAnalyzer(t *testing.T) {
	a := NewMemoryAnalyzer()
	if a == nil {
		t.Fatal("NewMemoryAnalyzer() returned nil")
	}
	if a.Name() != "system.memory" {
		t.Errorf("Name() = %v, want %v", a.Name(), "system.memory")
	}
}

func TestMemoryAnalyzerAnalyze_NoData(t *testing.T) {
	a := NewMemoryAnalyzer()
	ctx := context.Background()
	data := &types.DiagnosticData{
		SystemMetrics: nil,
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues, got %d", len(issues))
	}
}

func TestMemoryAnalyzerAnalyze_EmptyMetrics(t *testing.T) {
	a := NewMemoryAnalyzer()
	ctx := context.Background()
	data := &types.DiagnosticData{
		SystemMetrics: &types.SystemMetrics{
			MemTotal:     0,
			MemAvailable: 0,
		},
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues with empty metrics, got %d", len(issues))
	}
}

func TestMemoryAnalyzerAnalyze_Normal(t *testing.T) {
	a := NewMemoryAnalyzer()
	ctx := context.Background()
	data := &types.DiagnosticData{
		SystemMetrics: &types.SystemMetrics{
			MemTotal:     8000000,
			MemAvailable: 4000000, // 50% used
		},
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues, got %d", len(issues))
	}
}

func TestMemoryAnalyzerAnalyze_Exactly85Percent(t *testing.T) {
	a := NewMemoryAnalyzer()
	ctx := context.Background()
	// Exactly 85% usage
	data := &types.DiagnosticData{
		SystemMetrics: &types.SystemMetrics{
			MemTotal:     10000,
			MemAvailable: 1500, // 85% used (8500/10000)
		},
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue at exactly 85%%, got %d", len(issues))
	}
	if issues[0].Severity != types.SeverityWarning {
		t.Errorf("Expected Warning severity, got %v", issues[0].Severity)
	}
	if issues[0].ENName != "ELEVATED_MEMORY_USAGE" {
		t.Errorf("Expected ELEVATED_MEMORY_USAGE, got %v", issues[0].ENName)
	}
}

func TestMemoryAnalyzerAnalyze_JustBelow85Percent(t *testing.T) {
	a := NewMemoryAnalyzer()
	ctx := context.Background()
	// Just below 85% usage
	data := &types.DiagnosticData{
		SystemMetrics: &types.SystemMetrics{
			MemTotal:     10000,
			MemAvailable: 1600, // 84% used
		},
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	// 84% should not trigger warning (>= 85%)
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues just below 85%%, got %d", len(issues))
	}
}

func TestMemoryAnalyzerAnalyze_Exactly95Percent(t *testing.T) {
	a := NewMemoryAnalyzer()
	ctx := context.Background()
	// Exactly 95% usage
	data := &types.DiagnosticData{
		SystemMetrics: &types.SystemMetrics{
			MemTotal:     10000,
			MemAvailable: 500, // 95% used
		},
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue at exactly 95%%, got %d", len(issues))
	}
	if issues[0].Severity != types.SeverityCritical {
		t.Errorf("Expected Critical severity, got %v", issues[0].Severity)
	}
	if issues[0].ENName != "HIGH_MEMORY_USAGE" {
		t.Errorf("Expected HIGH_MEMORY_USAGE, got %v", issues[0].ENName)
	}
}

func TestMemoryAnalyzerAnalyze_CriticalMemoryUsage(t *testing.T) {
	a := NewMemoryAnalyzer()
	ctx := context.Background()
	// 96% usage - critical
	data := &types.DiagnosticData{
		SystemMetrics: &types.SystemMetrics{
			MemTotal:     1000000,
			MemAvailable: 40000, // 96% used
		},
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].Severity != types.SeverityCritical {
		t.Errorf("Expected Critical severity, got %v", issues[0].Severity)
	}
}

func TestMemoryAnalyzerAnalyze_WarningMemoryUsage(t *testing.T) {
	a := NewMemoryAnalyzer()
	ctx := context.Background()
	// 90% usage - warning
	data := &types.DiagnosticData{
		SystemMetrics: &types.SystemMetrics{
			MemTotal:     1000000,
			MemAvailable: 100000, // 90% used
		},
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].Severity != types.SeverityWarning {
		t.Errorf("Expected Warning severity, got %v", issues[0].Severity)
	}
}

// ============ Disk Analyzer Tests ============

func TestNewDiskAnalyzer(t *testing.T) {
	a := NewDiskAnalyzer()
	if a == nil {
		t.Fatal("NewDiskAnalyzer() returned nil")
	}
	if a.Name() != "system.disk" {
		t.Errorf("Name() = %v, want %v", a.Name(), "system.disk")
	}
}

func TestDiskAnalyzerAnalyze_NoData(t *testing.T) {
	a := NewDiskAnalyzer()
	ctx := context.Background()
	data := &types.DiagnosticData{
		SystemMetrics: nil,
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues, got %d", len(issues))
	}
}

func TestDiskAnalyzerAnalyze_NoDisks(t *testing.T) {
	a := NewDiskAnalyzer()
	ctx := context.Background()
	data := &types.DiagnosticData{
		SystemMetrics: &types.SystemMetrics{
			DiskUsage: []types.DiskUsage{},
		},
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues with no disks, got %d", len(issues))
	}
}

func TestDiskAnalyzerAnalyze_NormalUsage(t *testing.T) {
	a := NewDiskAnalyzer()
	ctx := context.Background()
	data := &types.DiagnosticData{
		SystemMetrics: &types.SystemMetrics{
			DiskUsage: []types.DiskUsage{
				{MountPoint: "/", UsedPercent: 50.0},
				{MountPoint: "/var", UsedPercent: 75.0},
			},
		},
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues for normal disk usage, got %d", len(issues))
	}
}

func TestDiskAnalyzerAnalyze_Exactly90Percent(t *testing.T) {
	a := NewDiskAnalyzer()
	ctx := context.Background()
	data := &types.DiagnosticData{
		SystemMetrics: &types.SystemMetrics{
			DiskUsage: []types.DiskUsage{
				{MountPoint: "/", UsedPercent: 90.0},
			},
		},
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue at exactly 90%%, got %d", len(issues))
	}
	if issues[0].Severity != types.SeverityWarning {
		t.Errorf("Expected Warning severity, got %v", issues[0].Severity)
	}
	if issues[0].ENName != "DISK_SPACE_LOW" {
		t.Errorf("Expected DISK_SPACE_LOW, got %v", issues[0].ENName)
	}
}

func TestDiskAnalyzerAnalyze_JustBelow90Percent(t *testing.T) {
	a := NewDiskAnalyzer()
	ctx := context.Background()
	data := &types.DiagnosticData{
		SystemMetrics: &types.SystemMetrics{
			DiskUsage: []types.DiskUsage{
				{MountPoint: "/", UsedPercent: 89.9},
			},
		},
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues just below 90%%, got %d", len(issues))
	}
}

func TestDiskAnalyzerAnalyze_Exactly95Percent(t *testing.T) {
	a := NewDiskAnalyzer()
	ctx := context.Background()
	data := &types.DiagnosticData{
		SystemMetrics: &types.SystemMetrics{
			DiskUsage: []types.DiskUsage{
				{MountPoint: "/", UsedPercent: 95.0},
			},
		},
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue at exactly 95%%, got %d", len(issues))
	}
	if issues[0].Severity != types.SeverityCritical {
		t.Errorf("Expected Critical severity, got %v", issues[0].Severity)
	}
	if issues[0].ENName != "DISK_SPACE_CRITICAL" {
		t.Errorf("Expected DISK_SPACE_CRITICAL, got %v", issues[0].ENName)
	}
}

func TestDiskAnalyzerAnalyze_CriticalDiskUsage(t *testing.T) {
	a := NewDiskAnalyzer()
	ctx := context.Background()
	data := &types.DiagnosticData{
		SystemMetrics: &types.SystemMetrics{
			DiskUsage: []types.DiskUsage{
				{MountPoint: "/", UsedPercent: 96.0},
				{MountPoint: "/var", UsedPercent: 98.0},
			},
		},
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 2 {
		t.Fatalf("Expected 2 issues, got %d", len(issues))
	}
	for _, issue := range issues {
		if issue.Severity != types.SeverityCritical {
			t.Errorf("Expected Critical severity, got %v", issue.Severity)
		}
	}
}

func TestDiskAnalyzerAnalyze_MixedUsage(t *testing.T) {
	a := NewDiskAnalyzer()
	ctx := context.Background()
	data := &types.DiagnosticData{
		SystemMetrics: &types.SystemMetrics{
			DiskUsage: []types.DiskUsage{
				{MountPoint: "/", UsedPercent: 50.0},    // Normal
				{MountPoint: "/var", UsedPercent: 92.0}, // Warning
				{MountPoint: "/tmp", UsedPercent: 97.0}, // Critical
			},
		},
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 2 {
		t.Fatalf("Expected 2 issues, got %d", len(issues))
	}
	
	// Check that we have one warning and one critical
	var warningCount, criticalCount int
	for _, issue := range issues {
		switch issue.Severity {
		case types.SeverityWarning:
			warningCount++
		case types.SeverityCritical:
			criticalCount++
		}
	}
	if warningCount != 1 {
		t.Errorf("Expected 1 warning issue, got %d", warningCount)
	}
	if criticalCount != 1 {
		t.Errorf("Expected 1 critical issue, got %d", criticalCount)
	}
}

// ============ Swap Analyzer Tests ============

func TestNewSwapAnalyzer(t *testing.T) {
	a := NewSwapAnalyzer()
	if a == nil {
		t.Fatal("NewSwapAnalyzer() returned nil")
	}
	if a.Name() != "system.swap" {
		t.Errorf("Name() = %v, want %v", a.Name(), "system.swap")
	}
}

func TestSwapAnalyzerAnalyze_NoData(t *testing.T) {
	a := NewSwapAnalyzer()
	ctx := context.Background()
	data := &types.DiagnosticData{
		SystemMetrics: nil,
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues, got %d", len(issues))
	}
}

func TestSwapAnalyzerAnalyze_NoSwap(t *testing.T) {
	a := NewSwapAnalyzer()
	ctx := context.Background()
	data := &types.DiagnosticData{
		SystemMetrics: &types.SystemMetrics{
			SwapTotal: 0,
		},
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues with no swap, got %d", len(issues))
	}
}

func TestSwapAnalyzerAnalyze_WithSwap(t *testing.T) {
	a := NewSwapAnalyzer()
	ctx := context.Background()
	data := &types.DiagnosticData{
		SystemMetrics: &types.SystemMetrics{
			SwapTotal: 2097152, // 2GB in KB
		},
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].Severity != types.SeverityInfo {
		t.Errorf("Expected Info severity, got %v", issues[0].Severity)
	}
	if issues[0].ENName != "SWAP_NOT_DISABLED" {
		t.Errorf("Expected SWAP_NOT_DISABLED, got %v", issues[0].ENName)
	}
}

// ============ Conntrack Analyzer Tests ============

func TestNewConntrackAnalyzer(t *testing.T) {
	a := NewConntrackAnalyzer()
	if a == nil {
		t.Fatal("NewConntrackAnalyzer() returned nil")
	}
	if a.Name() != "system.conntrack" {
		t.Errorf("Name() = %v, want %v", a.Name(), "system.conntrack")
	}
}

func TestConntrackAnalyzerAnalyze_NoData(t *testing.T) {
	a := NewConntrackAnalyzer()
	ctx := context.Background()
	data := &types.DiagnosticData{
		SystemMetrics: nil,
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues, got %d", len(issues))
	}
}

func TestConntrackAnalyzerAnalyze_Normal(t *testing.T) {
	a := NewConntrackAnalyzer()
	ctx := context.Background()
	data := &types.DiagnosticData{
		SystemMetrics: &types.SystemMetrics{
			ConntrackCurrent: 10000,
			ConntrackMax:     65536,
		},
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	// 10000/65536 = 15.26% - below warning threshold
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues for normal conntrack usage, got %d", len(issues))
	}
}

func TestConntrackAnalyzerAnalyze_Exactly80Percent(t *testing.T) {
	a := NewConntrackAnalyzer()
	ctx := context.Background()
	data := &types.DiagnosticData{
		SystemMetrics: &types.SystemMetrics{
			ConntrackCurrent: 52429, // ~80% of 65536
			ConntrackMax:     65536,
		},
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue at exactly 80%%, got %d", len(issues))
	}
	if issues[0].Severity != types.SeverityWarning {
		t.Errorf("Expected Warning severity, got %v", issues[0].Severity)
	}
	if issues[0].ENName != "CONNTRACK_TABLE_HIGH_USAGE" {
		t.Errorf("Expected CONNTRACK_TABLE_HIGH_USAGE, got %v", issues[0].ENName)
	}
}

func TestConntrackAnalyzerAnalyze_Exactly95Percent(t *testing.T) {
	a := NewConntrackAnalyzer()
	ctx := context.Background()
	data := &types.DiagnosticData{
		SystemMetrics: &types.SystemMetrics{
			ConntrackCurrent: 62260, // >= 95% of 65536 (actually 95.001%)
			ConntrackMax:     65536,
		},
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue at exactly 95%%, got %d", len(issues))
	}
	if issues[0].Severity != types.SeverityCritical {
		t.Errorf("Expected Critical severity, got %v", issues[0].Severity)
	}
	if issues[0].ENName != "CONNTRACK_TABLE_FULL" {
		t.Errorf("Expected CONNTRACK_TABLE_FULL, got %v", issues[0].ENName)
	}
}

func TestConntrackAnalyzerAnalyze_DefaultMax(t *testing.T) {
	a := NewConntrackAnalyzer()
	ctx := context.Background()
	// Max is 0, should default to 65536
	data := &types.DiagnosticData{
		SystemMetrics: &types.SystemMetrics{
			ConntrackCurrent: 60000,
			ConntrackMax:     0, // Will default to 65536
		},
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	// 60000/65536 = 91.55% - above warning threshold
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].Severity != types.SeverityWarning {
		t.Errorf("Expected Warning severity, got %v", issues[0].Severity)
	}
}

func TestConntrackAnalyzerAnalyze_WarningLevel(t *testing.T) {
	a := NewConntrackAnalyzer()
	ctx := context.Background()
	data := &types.DiagnosticData{
		SystemMetrics: &types.SystemMetrics{
			ConntrackCurrent: 55000, // ~84%
			ConntrackMax:     65536,
		},
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].Severity != types.SeverityWarning {
		t.Errorf("Expected Warning severity, got %v", issues[0].Severity)
	}
}

func TestConntrackAnalyzerAnalyze_CriticalLevel(t *testing.T) {
	a := NewConntrackAnalyzer()
	ctx := context.Background()
	data := &types.DiagnosticData{
		SystemMetrics: &types.SystemMetrics{
			ConntrackCurrent: 64000, // ~97.6%
			ConntrackMax:     65536,
		},
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].Severity != types.SeverityCritical {
		t.Errorf("Expected Critical severity, got %v", issues[0].Severity)
	}
}

// ============ File Handle Analyzer Tests ============

func TestNewFileHandleAnalyzer(t *testing.T) {
	a := NewFileHandleAnalyzer()
	if a == nil {
		t.Fatal("NewFileHandleAnalyzer() returned nil")
	}
	if a.Name() != "system.filehandle" {
		t.Errorf("Name() = %v, want %v", a.Name(), "system.filehandle")
	}
}

func TestFileHandleAnalyzerAnalyze_NoFile(t *testing.T) {
	a := NewFileHandleAnalyzer()
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

func TestFileHandleAnalyzerAnalyze_NoMatch(t *testing.T) {
	a := NewFileHandleAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"system_status": []byte("normal status output without file handle info"),
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues, got %d", len(issues))
	}
}

func TestFileHandleAnalyzerAnalyze_BelowThreshold(t *testing.T) {
	a := NewFileHandleAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"system_status": []byte("10000 fds (PID 1234)"),
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	// 10000 < 50000 threshold
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues below threshold, got %d", len(issues))
	}
}

func TestFileHandleAnalyzerAnalyze_AboveThreshold(t *testing.T) {
	a := NewFileHandleAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"system_status": []byte("55000 fds (PID 1234)"),
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].Severity != types.SeverityWarning {
		t.Errorf("Expected Warning severity, got %v", issues[0].Severity)
	}
	if issues[0].ENName != "HIGH_FILE_HANDLES" {
		t.Errorf("Expected HIGH_FILE_HANDLES, got %v", issues[0].ENName)
	}
}

func TestFileHandleAnalyzerAnalyze_ExactlyThreshold(t *testing.T) {
	a := NewFileHandleAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"system_status": []byte("50001 fds (PID 1234)"),
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue just above threshold, got %d", len(issues))
	}
}

// ============ Process State Analyzer Tests ============

func TestNewProcessStateAnalyzer(t *testing.T) {
	a := NewProcessStateAnalyzer()
	if a == nil {
		t.Fatal("NewProcessStateAnalyzer() returned nil")
	}
	if a.Name() != "system.process_state" {
		t.Errorf("Name() = %v, want %v", a.Name(), "system.process_state")
	}
}

func TestProcessStateAnalyzerAnalyze_NoFile(t *testing.T) {
	a := NewProcessStateAnalyzer()
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

func TestProcessStateAnalyzerAnalyze_NoIssue(t *testing.T) {
	a := NewProcessStateAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"ps_command_status": []byte("normal process status"),
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues, got %d", len(issues))
	}
}

func TestProcessStateAnalyzerAnalyze_PSHung(t *testing.T) {
	a := NewProcessStateAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"ps_command_status": []byte("ps -ef command is hung for 30 seconds"),
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].Severity != types.SeverityCritical {
		t.Errorf("Expected Critical severity, got %v", issues[0].Severity)
	}
	if issues[0].ENName != "PS_COMMAND_HUNG" {
		t.Errorf("Expected PS_COMMAND_HUNG, got %v", issues[0].ENName)
	}
}

func TestProcessStateAnalyzerAnalyze_DStateProcess(t *testing.T) {
	a := NewProcessStateAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"ps_command_status": []byte("process nginx is in State D for 60 seconds"),
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].Severity != types.SeverityCritical {
		t.Errorf("Expected Critical severity, got %v", issues[0].Severity)
	}
	if issues[0].ENName != "PROCESS_IN_D_STATE" {
		t.Errorf("Expected PROCESS_IN_D_STATE, got %v", issues[0].ENName)
	}
}

func TestProcessStateAnalyzerAnalyze_MultipleDStateProcesses(t *testing.T) {
	a := NewProcessStateAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"ps_command_status": []byte(
			"process nginx is in State D\n" +
			"process php-fpm is in State D\n" +
			"process mysql is in State D",
		),
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	// Check that count is reflected in description
	if !strings.Contains(issues[0].Details, "3") {
		t.Errorf("Expected details to contain '3', got: %s", issues[0].Details)
	}
}

func TestProcessStateAnalyzerAnalyze_BothIssues(t *testing.T) {
	a := NewProcessStateAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"ps_command_status": []byte(
			"ps -ef command is hung\n" +
			"process nginx is in State D",
		),
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 2 {
		t.Fatalf("Expected 2 issues, got %d", len(issues))
	}
	
	var hasHung, hasDState bool
	for _, issue := range issues {
		switch issue.ENName {
		case "PS_COMMAND_HUNG":
			hasHung = true
		case "PROCESS_IN_D_STATE":
			hasDState = true
		}
	}
	if !hasHung {
		t.Error("Expected PS_COMMAND_HUNG issue")
	}
	if !hasDState {
		t.Error("Expected PROCESS_IN_D_STATE issue")
	}
}

// ============ Integration Tests ============

func TestAllSystemAnalyzers_Registration(t *testing.T) {
	tests := []struct {
		name     string
		analyzer interface{ Name() string }
	}{
		{"CPU", NewCPUAnalyzer()},
		{"Memory", NewMemoryAnalyzer()},
		{"Disk", NewDiskAnalyzer()},
		{"Swap", NewSwapAnalyzer()},
		{"Conntrack", NewConntrackAnalyzer()},
		{"FileHandle", NewFileHandleAnalyzer()},
		{"ProcessState", NewProcessStateAnalyzer()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.analyzer == nil {
				t.Errorf("%s analyzer is nil", tt.name)
			}
		})
	}
}
