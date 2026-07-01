package analyzer

import (
	"context"
	"fmt"
	"time"

	"github.com/kudig-io/klaw/internal/diag/analyzer"
	"github.com/kudig-io/klaw/internal/diag/ebpf/probe"
	"github.com/kudig-io/klaw/internal/diag/types"
)

// DNSAnalyzer analyzes DNS queries using eBPF
type DNSAnalyzer struct {
	*analyzer.BaseAnalyzer
	config *probe.Config
}

// NewDNSAnalyzer creates a new DNS analyzer
func NewDNSAnalyzer() *DNSAnalyzer {
	return &DNSAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer(
			"ebpf.dns",
			"使用 eBPF 分析 DNS 查询延迟",
			"ebpf",
			[]types.DataMode{types.ModeOnline},
		),
		config: probe.DefaultConfig(),
	}
}

// Analyze performs DNS analysis using eBPF
func (a *DNSAnalyzer) Analyze(ctx context.Context, data *types.DiagnosticData) ([]types.Issue, error) {
	var issues []types.Issue

	if !a.hasPermission() {
		return issues, nil
	}

	config := &probe.Config{
		EnableTCP:        false,
		EnableDNS:        true,
		EnableSyscall:    false,
		EnableFileIO:     false,
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

	stats, err := probeMgr.GetDNSStats()
	if err != nil {
		return issues, nil
	}

	// Check for failed DNS queries
	if stats.TotalQueries > 0 {
		failureRate := float64(stats.FailedQueries) / float64(stats.TotalQueries) * 100
		if failureRate > 10 {
			issue := types.NewIssue(
				types.SeverityCritical,
				"DNS 查询失败率高",
				"EBPF_DNS_HIGH_FAILURE",
				fmt.Sprintf("DNS 失败率 %.1f%% (%d/%d)", failureRate, stats.FailedQueries, stats.TotalQueries),
				"network/dns",
			).WithRemediation("检查 CoreDNS 状态，验证上游 DNS 服务器配置")
			issue.AnalyzerName = a.Name()
			issues = append(issues, *issue)
		} else if failureRate > 1 {
			issue := types.NewIssue(
				types.SeverityWarning,
				"DNS 查询有失败",
				"EBPF_DNS_FAILURES",
				fmt.Sprintf("DNS 失败率 %.1f%% (%d/%d)", failureRate, stats.FailedQueries, stats.TotalQueries),
				"network/dns",
			).WithRemediation("监控 DNS 服务状态")
			issue.AnalyzerName = a.Name()
			issues = append(issues, *issue)
		}
	}

	// Check for high DNS latency
	if stats.AvgLatencyMicrosec > 100000 { // 100ms
		issue := types.NewIssue(
			types.SeverityWarning,
			"DNS 查询延迟高",
			"EBPF_DNS_HIGH_LATENCY",
			fmt.Sprintf("DNS 平均延迟 %.2f ms", float64(stats.AvgLatencyMicrosec)/1000),
			"network/dns",
		).WithRemediation("检查 CoreDNS 缓存配置，考虑启用 NodeLocal DNSCache")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	}

	return issues, nil
}

func (a *DNSAnalyzer) hasPermission() bool {
	return true
}
