// Package system provides system resource analyzers
package system

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/kudig-io/klaw/internal/diag/analyzer"
	"github.com/kudig-io/klaw/internal/diag/types"
)

// CPUAnalyzer checks CPU load conditions.
type CPUAnalyzer struct {
	*analyzer.BaseAnalyzer
}

// NewCPUAnalyzer creates a new CPU analyzer.
func NewCPUAnalyzer() *CPUAnalyzer {
	return &CPUAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer(
			"system.cpu",
			"检查CPU负载状态",
			"system",
			[]types.DataMode{types.ModeOffline, types.ModeOnline},
		),
	}
}

// Analyze performs CPU load analysis.
func (a *CPUAnalyzer) Analyze(_ context.Context, data *types.DiagnosticData) ([]types.Issue, error) {
	var issues []types.Issue

	if data.SystemMetrics == nil {
		return issues, nil
	}

	cpuCores := data.SystemMetrics.CPUCores
	if cpuCores == 0 {
		cpuCores = 4 // Default
	}

	load15Min := data.SystemMetrics.LoadAvg15Min

	// Critical: load > 4 * cores
	criticalThreshold := float64(cpuCores) * 4
	warningThreshold := float64(cpuCores) * 2

	if load15Min > criticalThreshold {
		issue := types.NewIssue(
			types.SeverityCritical,
			"系统负载过高",
			"HIGH_SYSTEM_LOAD",
			fmt.Sprintf("15分钟平均负载 %.2f，超过CPU核心数(%d)的4倍", load15Min, cpuCores),
			"system_status",
		).WithRemediation("检查高CPU进程: top -c 或 ps aux --sort=-%cpu | head")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	} else if load15Min > warningThreshold {
		issue := types.NewIssue(
			types.SeverityWarning,
			"系统负载偏高",
			"ELEVATED_SYSTEM_LOAD",
			fmt.Sprintf("15分钟平均负载 %.2f，超过CPU核心数(%d)的2倍", load15Min, cpuCores),
			"system_status",
		).WithRemediation("检查高CPU进程并评估是否正常")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	}

	return issues, nil
}

// MemoryAnalyzer checks memory usage conditions.
type MemoryAnalyzer struct {
	*analyzer.BaseAnalyzer
}

// NewMemoryAnalyzer creates a new memory analyzer.
func NewMemoryAnalyzer() *MemoryAnalyzer {
	return &MemoryAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer(
			"system.memory",
			"检查内存使用状态",
			"system",
			[]types.DataMode{types.ModeOffline, types.ModeOnline},
		),
	}
}

// Analyze performs memory usage analysis.
func (a *MemoryAnalyzer) Analyze(_ context.Context, data *types.DiagnosticData) ([]types.Issue, error) {
	var issues []types.Issue

	if data.SystemMetrics == nil {
		return issues, nil
	}

	memTotal := data.SystemMetrics.MemTotal
	memAvailable := data.SystemMetrics.MemAvailable

	if memTotal == 0 || memAvailable == 0 {
		return issues, nil
	}

	usagePercent := float64(memTotal-memAvailable) / float64(memTotal) * 100

	if usagePercent >= 95 {
		issue := types.NewIssue(
			types.SeverityCritical,
			"内存使用率过高",
			"HIGH_MEMORY_USAGE",
			fmt.Sprintf("内存使用率 %.0f%%，可能导致OOM", usagePercent),
			"memory_info",
		).WithRemediation("检查占用内存高的进程: ps aux --sort=-rss | head; 考虑扩容或优化应用")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	} else if usagePercent >= 85 {
		issue := types.NewIssue(
			types.SeverityWarning,
			"内存使用率偏高",
			"ELEVATED_MEMORY_USAGE",
			fmt.Sprintf("内存使用率 %.0f%%", usagePercent),
			"memory_info",
		).WithRemediation("关注内存趋势，考虑清理缓存: sync; echo 3 > /proc/sys/vm/drop_caches")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	}

	return issues, nil
}

// DiskAnalyzer checks disk space conditions.
type DiskAnalyzer struct {
	*analyzer.BaseAnalyzer
}

// NewDiskAnalyzer creates a new disk analyzer.
func NewDiskAnalyzer() *DiskAnalyzer {
	return &DiskAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer(
			"system.disk",
			"检查磁盘空间状态",
			"system",
			[]types.DataMode{types.ModeOffline, types.ModeOnline},
		),
	}
}

// Analyze performs disk space analysis.
func (a *DiskAnalyzer) Analyze(_ context.Context, data *types.DiagnosticData) ([]types.Issue, error) {
	var issues []types.Issue

	if data.SystemMetrics == nil {
		return issues, nil
	}

	for _, disk := range data.SystemMetrics.DiskUsage {
		if disk.UsedPercent >= 95 {
			issue := types.NewIssue(
				types.SeverityCritical,
				"磁盘空间严重不足",
				"DISK_SPACE_CRITICAL",
				fmt.Sprintf("挂载点 %s 使用率 %.0f%%", disk.MountPoint, disk.UsedPercent),
				"system_status",
			).WithRemediation(fmt.Sprintf("清理磁盘: du -sh %s/* | sort -rh | head; 删除无用文件或扩容", disk.MountPoint))
			issue.AnalyzerName = a.Name()
			issues = append(issues, *issue)
		} else if disk.UsedPercent >= 90 {
			issue := types.NewIssue(
				types.SeverityWarning,
				"磁盘空间不足",
				"DISK_SPACE_LOW",
				fmt.Sprintf("挂载点 %s 使用率 %.0f%%", disk.MountPoint, disk.UsedPercent),
				"system_status",
			).WithRemediation(fmt.Sprintf("检查占用空间大的目录: du -sh %s/* | sort -rh | head", disk.MountPoint))
			issue.AnalyzerName = a.Name()
			issues = append(issues, *issue)
		}
	}

	return issues, nil
}

// SwapAnalyzer checks swap configuration
type SwapAnalyzer struct {
	*analyzer.BaseAnalyzer
}

// NewSwapAnalyzer creates a new swap analyzer
func NewSwapAnalyzer() *SwapAnalyzer {
	return &SwapAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer(
			"system.swap",
			"检查Swap配置",
			"system",
			[]types.DataMode{types.ModeOffline, types.ModeOnline},
		),
	}
}

// Analyze performs swap configuration analysis
func (a *SwapAnalyzer) Analyze(ctx context.Context, data *types.DiagnosticData) ([]types.Issue, error) {
	var issues []types.Issue

	if data.SystemMetrics == nil {
		return issues, nil
	}

	if data.SystemMetrics.SwapTotal > 0 {
		issue := types.NewIssue(
			types.SeverityInfo,
			"Swap未禁用",
			"SWAP_NOT_DISABLED",
			fmt.Sprintf("Kubernetes节点建议禁用swap，当前 %dKB", data.SystemMetrics.SwapTotal),
			"system_info",
		).WithRemediation("禁用swap: swapoff -a && sed -i '/swap/d' /etc/fstab")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	}

	return issues, nil
}

// ConntrackAnalyzer checks connection tracking table
type ConntrackAnalyzer struct {
	*analyzer.BaseAnalyzer
}

// NewConntrackAnalyzer creates a new conntrack analyzer
func NewConntrackAnalyzer() *ConntrackAnalyzer {
	return &ConntrackAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer(
			"system.conntrack",
			"检查连接跟踪表状态",
			"system",
			[]types.DataMode{types.ModeOffline, types.ModeOnline},
		),
	}
}

// Analyze performs conntrack table analysis
func (a *ConntrackAnalyzer) Analyze(ctx context.Context, data *types.DiagnosticData) ([]types.Issue, error) {
	var issues []types.Issue

	if data.SystemMetrics == nil {
		return issues, nil
	}

	current := data.SystemMetrics.ConntrackCurrent
	max := data.SystemMetrics.ConntrackMax

	if max == 0 {
		max = 65536 // Default
	}

	usagePercent := float64(current) / float64(max) * 100

	if usagePercent >= 95 {
		issue := types.NewIssue(
			types.SeverityCritical,
			"连接跟踪表满",
			"CONNTRACK_TABLE_FULL",
			fmt.Sprintf("当前连接数 %d/%d (%.0f%%)，接近上限", current, max, usagePercent),
			"network_info",
		).WithRemediation("增大conntrack表: sysctl -w net.netfilter.nf_conntrack_max=262144")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	} else if usagePercent >= 80 {
		issue := types.NewIssue(
			types.SeverityWarning,
			"连接跟踪表使用率高",
			"CONNTRACK_TABLE_HIGH_USAGE",
			fmt.Sprintf("当前连接数 %d/%d (%.0f%%)", current, max, usagePercent),
			"network_info",
		).WithRemediation("考虑增大conntrack表大小")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	}

	return issues, nil
}

// FileHandleAnalyzer checks file handle usage
type FileHandleAnalyzer struct {
	*analyzer.BaseAnalyzer
}

// NewFileHandleAnalyzer creates a new file handle analyzer
func NewFileHandleAnalyzer() *FileHandleAnalyzer {
	return &FileHandleAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer(
			"system.filehandle",
			"检查文件句柄使用状态",
			"system",
			[]types.DataMode{types.ModeOffline},
		),
	}
}

// Analyze performs file handle analysis
func (a *FileHandleAnalyzer) Analyze(ctx context.Context, data *types.DiagnosticData) ([]types.Issue, error) {
	var issues []types.Issue

	systemStatus, ok := data.GetRawFile("system_status")
	if !ok {
		return issues, nil
	}

	content := string(systemStatus)

	// Look for file handle count pattern
	fdRe := regexp.MustCompile(`(\d+)\s+fds\s+\(PID`)
	matches := fdRe.FindStringSubmatch(content)

	if len(matches) > 1 {
		maxFds, _ := strconv.Atoi(matches[1])
		if maxFds > 50000 {
			issue := types.NewIssue(
				types.SeverityWarning,
				"文件句柄使用量过高",
				"HIGH_FILE_HANDLES",
				fmt.Sprintf("进程最大文件句柄数: %d", maxFds),
				"system_status",
			).WithRemediation("检查文件句柄泄漏: lsof | wc -l")
			issue.AnalyzerName = a.Name()
			issues = append(issues, *issue)
		}
	}

	return issues, nil
}

// ProcessStateAnalyzer checks for D-state processes and ps command hang
type ProcessStateAnalyzer struct {
	*analyzer.BaseAnalyzer
}

// NewProcessStateAnalyzer creates a new process state analyzer
func NewProcessStateAnalyzer() *ProcessStateAnalyzer {
	return &ProcessStateAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer(
			"system.process_state",
			"检查进程状态异常",
			"system",
			[]types.DataMode{types.ModeOffline},
		),
	}
}

// Analyze performs process state analysis
func (a *ProcessStateAnalyzer) Analyze(ctx context.Context, data *types.DiagnosticData) ([]types.Issue, error) {
	var issues []types.Issue

	psStatus, ok := data.GetRawFile("ps_command_status")
	if !ok {
		return issues, nil
	}

	content := string(psStatus)

	// Check if ps command is hung
	if strings.Contains(content, "ps -ef command is hung") {
		issue := types.NewIssue(
			types.SeverityCritical,
			"ps命令挂起",
			"PS_COMMAND_HUNG",
			"ps -ef命令挂起，系统可能存在D状态进程",
			"ps_command_status",
		).WithRemediation("检查D状态进程: ps aux | grep ' D '; 可能需要重启节点")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	}

	// Check for D-state processes
	if strings.Contains(content, "is in State D") {
		dProcCount := strings.Count(content, "is in State D")
		issue := types.NewIssue(
			types.SeverityCritical,
			"存在D状态进程",
			"PROCESS_IN_D_STATE",
			fmt.Sprintf("检测到 %d 个不可中断睡眠状态的进程", dProcCount),
			"ps_command_status",
		).WithRemediation("D状态进程通常由IO阻塞引起，检查磁盘/NFS状态; 可能需要重启节点")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	}

	return issues, nil
}

// init registers all system analyzers.
func init() {
	_ = analyzer.Register(NewCPUAnalyzer())
	_ = analyzer.Register(NewMemoryAnalyzer())
	_ = analyzer.Register(NewDiskAnalyzer())
	_ = analyzer.Register(NewSwapAnalyzer())
	_ = analyzer.Register(NewConntrackAnalyzer())
	_ = analyzer.Register(NewFileHandleAnalyzer())
	_ = analyzer.Register(NewProcessStateAnalyzer())
}
