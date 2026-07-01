// Package analyzer provides eBPF-based analyzers
package analyzer

import (
	"context"
	"fmt"
	"time"

	"github.com/kudig-io/klaw/internal/diag/analyzer"
	"github.com/kudig-io/klaw/internal/diag/ebpf/probe"
	"github.com/kudig-io/klaw/internal/diag/types"
)

// TCPAnalyzer analyzes TCP connections using eBPF
type TCPAnalyzer struct {
	*analyzer.BaseAnalyzer
	probeMgr *probe.ProbeManager
	config   *probe.Config
}

// NewTCPAnalyzer creates a new TCP analyzer
func NewTCPAnalyzer() *TCPAnalyzer {
	return &TCPAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer(
			"ebpf.tcp",
			"使用 eBPF 分析 TCP 连接",
			"ebpf",
			[]types.DataMode{types.ModeOnline},
		),
		config: probe.DefaultConfig(),
	}
}

// Analyze performs TCP analysis using eBPF
func (a *TCPAnalyzer) Analyze(ctx context.Context, data *types.DiagnosticData) ([]types.Issue, error) {
	var issues []types.Issue

	// Check if we have permission to run eBPF
	if !a.hasPermission() {
		return issues, nil
	}

	// Create probe manager
	config := &probe.Config{
		EnableTCP:        true,
		EnableDNS:        false,
		EnableSyscall:    false,
		EnableFileIO:     false,
		EnablePacketDrop: false,
		Duration:         10 * time.Second,
	}

	probeMgr, err := probe.NewProbeManager(config)
	if err != nil {
		return issues, nil // Silently skip if eBPF is not available
	}

	// Start probing
	probeCtx, cancel := context.WithTimeout(ctx, config.Duration)
	defer cancel()

	if err := probeMgr.Start(probeCtx); err != nil {
		return issues, nil
	}
	defer probeMgr.Stop()

	// Wait for collection
	select {
	case <-probeCtx.Done():
	case <-ctx.Done():
		return issues, ctx.Err()
	}

	// Get TCP statistics
	stats, err := probeMgr.GetTCPStats()
	if err != nil {
		return issues, nil
	}

	// Check for high retransmit rate
	if stats.Retransmits > 100 {
		issue := types.NewIssue(
			types.SeverityWarning,
			"TCP 重传率高",
			"EBPF_TCP_HIGH_RETRANSMIT",
			fmt.Sprintf("检测到 %d 次 TCP 重传，可能存在网络不稳定或拥塞", stats.Retransmits),
			"network/tcp",
		).WithRemediation("检查网络连接质量，调整 TCP 缓冲区大小或拥塞控制算法")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	}

	// Check for high latency
	if stats.AvgLatencyMicrosec > 1000000 { // 1 second
		issue := types.NewIssue(
			types.SeverityCritical,
			"TCP 连接延迟过高",
			"EBPF_TCP_HIGH_LATENCY",
			fmt.Sprintf("TCP 平均延迟 %.2f ms，超过 1 秒阈值", float64(stats.AvgLatencyMicrosec)/1000),
			"network/tcp",
		).WithRemediation("检查网络延迟，优化应用连接池配置，考虑使用连接池预热")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	} else if stats.AvgLatencyMicrosec > 100000 { // 100ms
		issue := types.NewIssue(
			types.SeverityWarning,
			"TCP 连接延迟偏高",
			"EBPF_TCP_ELEVATED_LATENCY",
			fmt.Sprintf("TCP 平均延迟 %.2f ms，建议关注", float64(stats.AvgLatencyMicrosec)/1000),
			"network/tcp",
		).WithRemediation("监控网络延迟趋势，检查是否有网络抖动")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	}

	return issues, nil
}

func (a *TCPAnalyzer) hasPermission() bool {
	// Check if running with CAP_BPF or as root
	// This is a simplified check
	return true // Assume permission for now
}
