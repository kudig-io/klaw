package analyzer

import (
	"context"
	"fmt"
	"time"

	"github.com/kudig-io/klaw/internal/diag/analyzer"
	"github.com/kudig-io/klaw/internal/diag/ebpf/probe"
	"github.com/kudig-io/klaw/internal/diag/types"
)

// FileIOAnalyzer analyzes file I/O using eBPF
type FileIOAnalyzer struct {
	*analyzer.BaseAnalyzer
	config *probe.Config
}

// NewFileIOAnalyzer creates a new file I/O analyzer
func NewFileIOAnalyzer() *FileIOAnalyzer {
	return &FileIOAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer(
			"ebpf.fileio",
			"使用 eBPF 分析文件 I/O 性能",
			"ebpf",
			[]types.DataMode{types.ModeOnline},
		),
		config: probe.DefaultConfig(),
	}
}

// Analyze performs file I/O analysis using eBPF
func (a *FileIOAnalyzer) Analyze(ctx context.Context, data *types.DiagnosticData) ([]types.Issue, error) {
	var issues []types.Issue

	if !a.hasPermission() {
		return issues, nil
	}

	config := &probe.Config{
		EnableTCP:        false,
		EnableDNS:        false,
		EnableSyscall:    false,
		EnableFileIO:     true,
		EnablePacketDrop: false,
		Duration:         10 * time.Second,
	}

	probeMgr, err := probe.NewProbeManager(config)
	if err != nil {
		return issues, nil
	}

	probeCtx, cancel := context.WithTimeout(ctx, config.Duration)
	defer cancel()

	if err := probeMgr.Start(probeCtx); err != nil {
		return issues, nil
	}
	defer probeMgr.Stop()

	select {
	case <-probeCtx.Done():
	case <-ctx.Done():
		return issues, ctx.Err()
	}

	stats, err := probeMgr.GetFileIOStats()
	if err != nil {
		return issues, nil
	}

	// Check for high file I/O latency
	if stats.AvgLatencyMicrosec > 10000 { // 10ms
		issue := types.NewIssue(
			types.SeverityWarning,
			"文件 I/O 延迟高",
			"EBPF_FILEIO_HIGH_LATENCY",
			fmt.Sprintf("文件 I/O 平均延迟 %.2f ms", float64(stats.AvgLatencyMicrosec)/1000),
			"storage/io",
		).WithRemediation("检查磁盘性能，考虑使用 SSD 或优化文件访问模式")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	}

	return issues, nil
}

func (a *FileIOAnalyzer) hasPermission() bool {
	return true
}
