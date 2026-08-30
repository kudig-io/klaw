// Package kernel provides kernel-related analyzers
package kernel

import (
	"context"
	"fmt"
	"strings"

	"github.com/kudig-io/klaw/internal/diag/analyzer"
	"github.com/kudig-io/klaw/internal/diag/types"
)

// PanicAnalyzer checks for kernel panic events.
type PanicAnalyzer struct {
	*analyzer.BaseAnalyzer
}

// NewPanicAnalyzer creates a new kernel panic analyzer.
func NewPanicAnalyzer() *PanicAnalyzer {
	return &PanicAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer(
			"kernel.panic",
			"检查内核Panic事件",
			"kernel",
			[]types.DataMode{types.ModeOffline},
		),
	}
}

// Analyze performs kernel panic analysis.
func (a *PanicAnalyzer) Analyze(_ context.Context, data *types.DiagnosticData) ([]types.Issue, error) {
	var issues []types.Issue

	// Check dmesg log
	dmesgLog, ok := data.GetRawFile("logs/dmesg.log")
	if !ok {
		// Also check varlogmessage.log
		dmesgLog, ok = data.GetRawFile("varlogmessage.log")
		if !ok {
			return issues, nil
		}
	}

	content := string(dmesgLog)

	if strings.Contains(content, "Kernel panic") {
		issue := types.NewIssue(
			types.SeverityCritical,
			"内核Panic",
			"KERNEL_PANIC",
			"内核发生panic事件",
			"logs/dmesg.log",
		).WithRemediation("检查dmesg日志定位问题原因; 可能需要联系厂商支持")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	}

	return issues, nil
}

// OOMAnalyzer checks for OOM killer events
type OOMAnalyzer struct {
	*analyzer.BaseAnalyzer
}

// NewOOMAnalyzer creates a new OOM analyzer
func NewOOMAnalyzer() *OOMAnalyzer {
	return &OOMAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer(
			"kernel.oom",
			"检查OOM Killer事件",
			"kernel",
			[]types.DataMode{types.ModeOffline},
		),
	}
}

// Analyze performs OOM killer analysis
func (a *OOMAnalyzer) Analyze(_ context.Context, data *types.DiagnosticData) ([]types.Issue, error) {
	var issues []types.Issue

	// Check multiple log sources
	logFiles := []string{"logs/dmesg.log", "logs/messages", "varlogmessage.log"}

	for _, logFile := range logFiles {
		logContent, ok := data.GetRawFile(logFile)
		if !ok {
			continue
		}

		content := string(logContent)
		oomCount := strings.Count(content, "Out of memory: Kill process")
		oomCount += strings.Count(content, "Out of memory:")

		if oomCount > 0 {
			issue := types.NewIssue(
				types.SeverityCritical,
				"内核触发OOM杀进程",
				"KERNEL_OOM_KILLER",
				fmt.Sprintf("内核OOM Killer被触发 %d 次", oomCount),
				logFile,
			).WithRemediation("检查被kill的进程: dmesg | grep -i oom; 考虑增加内存或限制Pod资源")
			issue.AnalyzerName = a.Name()
			issues = append(issues, *issue)

			break // Only report once
		}
	}

	return issues, nil
}

// FilesystemAnalyzer checks for filesystem errors
type FilesystemAnalyzer struct {
	*analyzer.BaseAnalyzer
}

// NewFilesystemAnalyzer creates a new filesystem analyzer
func NewFilesystemAnalyzer() *FilesystemAnalyzer {
	return &FilesystemAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer(
			"kernel.filesystem",
			"检查文件系统状态",
			"kernel",
			[]types.DataMode{types.ModeOffline},
		),
	}
}

// Analyze performs filesystem analysis
func (a *FilesystemAnalyzer) Analyze(_ context.Context, data *types.DiagnosticData) ([]types.Issue, error) {
	var issues []types.Issue

	logFiles := []string{"logs/dmesg.log", "varlogmessage.log"}

	for _, logFile := range logFiles {
		logContent, ok := data.GetRawFile(logFile)
		if !ok {
			continue
		}

		content := string(logContent)

		// Check for read-only filesystem
		if strings.Contains(content, "Read-only file system") ||
			strings.Contains(content, "Remounting filesystem read-only") {
			issue := types.NewIssue(
				types.SeverityCritical,
				"文件系统只读",
				"FILESYSTEM_READONLY",
				"文件系统被重新挂载为只读模式",
				logFile,
			).WithRemediation("检查磁盘健康: dmesg | grep -i error; fsck修复文件系统")
			issue.AnalyzerName = a.Name()
			issues = append(issues, *issue)
		}

		// Check for I/O errors
		ioErrorCount := strings.Count(content, "I/O error")
		if ioErrorCount > 10 {
			issue := types.NewIssue(
				types.SeverityCritical,
				"磁盘IO错误",
				"DISK_IO_ERROR",
				fmt.Sprintf("检测到 %d 次IO错误", ioErrorCount),
				logFile,
			).WithRemediation("检查磁盘健康: smartctl -a /dev/sdX; 考虑更换磁盘")
			issue.AnalyzerName = a.Name()
			issues = append(issues, *issue)
		}

		if len(issues) > 0 {
			break
		}
	}

	return issues, nil
}

// ModuleAnalyzer checks for kernel module loading failures
type ModuleAnalyzer struct {
	*analyzer.BaseAnalyzer
}

// NewModuleAnalyzer creates a new kernel module analyzer
func NewModuleAnalyzer() *ModuleAnalyzer {
	return &ModuleAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer(
			"kernel.module",
			"检查内核模块加载状态",
			"kernel",
			[]types.DataMode{types.ModeOffline},
		),
	}
}

// Analyze performs kernel module analysis
func (a *ModuleAnalyzer) Analyze(ctx context.Context, data *types.DiagnosticData) ([]types.Issue, error) {
	var issues []types.Issue

	dmesgLog, ok := data.GetRawFile("logs/dmesg.log")
	if !ok {
		dmesgLog, ok = data.GetRawFile("varlogmessage.log")
		if !ok {
			return issues, nil
		}
	}

	content := string(dmesgLog)

	if strings.Contains(content, "module") && strings.Contains(content, "failed") {
		issue := types.NewIssue(
			types.SeverityWarning,
			"内核模块加载失败",
			"KERNEL_MODULE_LOAD_FAILED",
			"存在内核模块加载失败",
			"logs/dmesg.log",
		).WithRemediation("检查模块依赖: lsmod; modprobe <module_name>")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	}

	return issues, nil
}

// NMIWatchdogAnalyzer checks for NMI watchdog events
type NMIWatchdogAnalyzer struct {
	*analyzer.BaseAnalyzer
}

// NewNMIWatchdogAnalyzer creates a new NMI watchdog analyzer
func NewNMIWatchdogAnalyzer() *NMIWatchdogAnalyzer {
	return &NMIWatchdogAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer(
			"kernel.nmi_watchdog",
			"检查NMI Watchdog事件",
			"kernel",
			[]types.DataMode{types.ModeOffline},
		),
	}
}

// Analyze performs NMI watchdog analysis
func (a *NMIWatchdogAnalyzer) Analyze(ctx context.Context, data *types.DiagnosticData) ([]types.Issue, error) {
	var issues []types.Issue

	dmesgLog, ok := data.GetRawFile("logs/dmesg.log")
	if !ok {
		return issues, nil
	}

	content := string(dmesgLog)

	if strings.Contains(content, "NMI watchdog") && strings.Contains(content, "hard LOCKUP") {
		issue := types.NewIssue(
			types.SeverityWarning,
			"NMI Watchdog触发",
			"NMI_WATCHDOG_TRIGGERED",
			"硬件看门狗被触发，可能存在CPU死锁",
			"logs/dmesg.log",
		).WithRemediation("检查系统负载和中断; 可能需要重启节点")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	}

	return issues, nil
}

// init registers all kernel analyzers.
func init() {
	analyzer.MustRegister(NewPanicAnalyzer())
	analyzer.MustRegister(NewOOMAnalyzer())
	analyzer.MustRegister(NewFilesystemAnalyzer())
	analyzer.MustRegister(NewModuleAnalyzer())
	analyzer.MustRegister(NewNMIWatchdogAnalyzer())
}
