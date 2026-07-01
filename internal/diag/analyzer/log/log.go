// Package log provides log analysis analyzers for journalctl and syslog
package log

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/kudig-io/klaw/internal/diag/analyzer"
	"github.com/kudig-io/klaw/internal/diag/types"
)

// SyslogAnalyzer analyzes syslog for critical patterns
 type SyslogAnalyzer struct {
	*analyzer.BaseAnalyzer
}

// NewSyslogAnalyzer creates a new syslog analyzer
func NewSyslogAnalyzer() *SyslogAnalyzer {
	return &SyslogAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer(
			"log.syslog",
			"分析系统日志关键错误",
			"log",
			[]types.DataMode{types.ModeOffline, types.ModeOnline},
		),
	}
}

// Analyze performs syslog analysis
func (a *SyslogAnalyzer) Analyze(_ context.Context, data *types.DiagnosticData) ([]types.Issue, error) {
	var issues []types.Issue

	// Check various log files
	logFiles := []string{
		"logs/syslog",
		"logs/messages",
		"varlogsyslog.log",
		"varlogmessages.log",
	}

	for _, logFile := range logFiles {
		content, ok := data.GetRawFile(logFile)
		if !ok {
			continue
		}

		issues = append(issues, a.analyzeLogContent(string(content), logFile)...)
	}

	return issues, nil
}

func (a *SyslogAnalyzer) analyzeLogContent(content, source string) []types.Issue {
	var issues []types.Issue
	lines := strings.Split(content, "\n")

	// Define error patterns with their severity and codes
	patterns := []struct {
		pattern   *regexp.Regexp
		name      string
		code      string
		severity  types.Severity
		message   string
		threshold int // minimum occurrences to trigger
	}{
		{
			regexp.MustCompile(`(?i)segfault|segmentation fault`),
			"段错误",
			"SYSLOG_SEGFAULT",
			types.SeverityCritical,
			"检测到程序段错误（内存访问违规）",
			1,
		},
		{
			regexp.MustCompile(`(?i)kernel panic|panic:`),
			"内核恐慌",
			"SYSLOG_KERNEL_PANIC",
			types.SeverityCritical,
			"检测到内核恐慌",
			1,
		},
		{
			regexp.MustCompile(`(?i)oom.kill|out of memory|killed process`),
			"OOM杀进程",
			"SYSLOG_OOM_KILL",
			types.SeverityCritical,
			"系统因内存不足杀死进程",
			1,
		},
		{
			regexp.MustCompile(`(?i)i/o error|io error|disk error|read error|write error`),
			"磁盘I/O错误",
			"SYSLOG_IO_ERROR",
			types.SeverityCritical,
			"检测到磁盘I/O错误",
			1,
		},
		{
			regexp.MustCompile(`(?i)nfs.*error|nfs.*timeout|nfs.*not responding`),
			"NFS错误",
			"SYSLOG_NFS_ERROR",
			types.SeverityWarning,
			"NFS文件系统错误",
			3,
		},
		{
			regexp.MustCompile(`(?i)connection refused|connection timed out`),
			"连接错误",
			"SYSLOG_CONNECTION_ERROR",
			types.SeverityWarning,
			"网络连接错误",
			5,
		},
		{
			regexp.MustCompile(`(?i)authentication failure|failed password|invalid user`),
			"认证失败",
			"SYSLOG_AUTH_FAILURE",
			types.SeverityWarning,
			"多次认证失败",
			5,
		},
		{
			regexp.MustCompile(`(?i)service failed|failed to start|exit code`),
			"服务启动失败",
			"SYSLOG_SERVICE_FAILED",
			types.SeverityWarning,
			"系统服务启动失败",
			3,
		},
	}

	// Count occurrences for each pattern
	for _, p := range patterns {
		count := 0
		for _, line := range lines {
			if p.pattern.MatchString(line) {
				count++
			}
		}

		if count >= p.threshold {
			issue := types.NewIssue(
				p.severity,
				p.name,
				p.code,
				fmt.Sprintf("%s: 在 %s 中发现 %d 次", p.message, source, count),
				source,
			)
			issue.AnalyzerName = a.Name()
			issues = append(issues, *issue)
		}
	}

	return issues
}

// JournalCtlAnalyzer analyzes journalctl logs
 type JournalCtlAnalyzer struct {
	*analyzer.BaseAnalyzer
}

// NewJournalCtlAnalyzer creates a new journalctl analyzer
func NewJournalCtlAnalyzer() *JournalCtlAnalyzer {
	return &JournalCtlAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer(
			"log.journalctl",
			"分析journalctl日志",
			"log",
			[]types.DataMode{types.ModeOffline, types.ModeOnline},
		),
	}
}

// Analyze performs journalctl analysis
func (a *JournalCtlAnalyzer) Analyze(_ context.Context, data *types.DiagnosticData) ([]types.Issue, error) {
	var issues []types.Issue

	// Check journalctl output files
	journalFiles := []string{
		"logs/journalctl.log",
		"logs/journalctl.txt",
		"varlogjournalctl.log",
	}

	for _, journalFile := range journalFiles {
		content, ok := data.GetRawFile(journalFile)
		if !ok {
			continue
		}

		issues = append(issues, a.analyzeJournalContent(string(content), journalFile)...)
	}

	return issues, nil
}

func (a *JournalCtlAnalyzer) analyzeJournalContent(content, source string) []types.Issue {
	var issues []types.Issue
	lines := strings.Split(content, "\n")

	// Define systemd/journald specific patterns
	patterns := []struct {
		pattern   *regexp.Regexp
		name      string
		code      string
		severity  types.Severity
		message   string
		threshold int
	}{
		{
			regexp.MustCompile(`(?i)failed to start|failed to restart|start request repeated too quickly`),
			"Systemd服务启动失败",
			"JOURNAL_SERVICE_FAILED",
			types.SeverityWarning,
			"Systemd服务启动失败",
			3,
		},
		{
			regexp.MustCompile(`(?i)dependency failed|required.*not available`),
			"Systemd依赖失败",
			"JOURNAL_DEPENDENCY_FAILED",
			types.SeverityWarning,
			"服务依赖未满足",
			3,
		},
		{
			regexp.MustCompile(`(?i)watchdog timeout|watchdog.*didn't stop`),
			"Watchdog超时",
			"JOURNAL_WATCHDOG_TIMEOUT",
			types.SeverityCritical,
			"服务Watchdog超时",
			1,
		},
		{
			regexp.MustCompile(`(?i)coredump|core dump|dumped core`),
			"程序崩溃",
			"JOURNAL_COREDUMP",
			types.SeverityCritical,
			"检测到程序崩溃生成core dump",
			1,
		},
		{
			regexp.MustCompile(`(?i)fatal|fatal error|fatal signal`),
			"致命错误",
			"JOURNAL_FATAL_ERROR",
			types.SeverityCritical,
			"检测到致命错误",
			1,
		},
		{
			regexp.MustCompile(`(?i)resource.*temporarily unavailable|too many open files`),
			"资源耗尽",
			"JOURNAL_RESOURCE_EXHAUSTED",
			types.SeverityWarning,
			"系统资源耗尽",
			3,
		},
	}

	for _, p := range patterns {
		count := 0
		for _, line := range lines {
			if p.pattern.MatchString(line) {
				count++
			}
		}

		if count >= p.threshold {
			issue := types.NewIssue(
				p.severity,
				p.name,
				p.code,
				fmt.Sprintf("%s: 在 %s 中发现 %d 次", p.message, source, count),
				source,
			)
			issue.AnalyzerName = a.Name()
			issues = append(issues, *issue)
		}
	}

	return issues
}

// KubeletLogAnalyzer analyzes kubelet logs for critical patterns
 type KubeletLogAnalyzer struct {
	*analyzer.BaseAnalyzer
}

// NewKubeletLogAnalyzer creates a new kubelet log analyzer
func NewKubeletLogAnalyzer() *KubeletLogAnalyzer {
	return &KubeletLogAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer(
			"log.kubelet",
			"分析Kubelet日志",
			"log",
			[]types.DataMode{types.ModeOffline, types.ModeOnline},
		),
	}
}

// Analyze performs kubelet log analysis
func (a *KubeletLogAnalyzer) Analyze(_ context.Context, data *types.DiagnosticData) ([]types.Issue, error) {
	var issues []types.Issue

	logFiles := []string{
		"logs/kubelet.log",
		"daemon_status/kubelet_status",
	}

	for _, logFile := range logFiles {
		content, ok := data.GetRawFile(logFile)
		if !ok {
			continue
		}

		issues = append(issues, a.analyzeKubeletContent(string(content), logFile)...)
	}

	return issues, nil
}

func (a *KubeletLogAnalyzer) analyzeKubeletContent(content, source string) []types.Issue {
	var issues []types.Issue
	lines := strings.Split(content, "\n")

	patterns := []struct {
		pattern   *regexp.Regexp
		name      string
		code      string
		severity  types.Severity
		message   string
		threshold int
	}{
		{
			regexp.MustCompile(`(?i)pleg is not healthy|pleg.*unhealthy`),
			"PLEG不健康",
			"KUBELET_PLEG_UNHEALTHY",
			types.SeverityCritical,
			"Kubelet PLEG（Pod生命周期事件生成器）不健康",
			1,
		},
		{
			regexp.MustCompile(`(?i)container runtime is down|runtime.*not ready`),
			"容器运行时异常",
			"KUBELET_RUNTIME_DOWN",
			types.SeverityCritical,
			"容器运行时不可用",
			1,
		},
		{
			regexp.MustCompile(`(?i)failed to create pod sandbox`),
			"Pod沙盒创建失败",
			"KUBELET_SANDBOX_FAILED",
			types.SeverityWarning,
			"Pod沙盒创建失败",
			3,
		},
		{
			regexp.MustCompile(`(?i)volume.*failed to mount|mount.*failed`),
			"卷挂载失败",
			"KUBELET_VOLUME_MOUNT_FAILED",
			types.SeverityWarning,
			"存储卷挂载失败",
			3,
		},
		{
			regexp.MustCompile(`(?i)eviction manager|attempting to reclaim|hard eviction`),
			"资源驱逐触发",
			"KUBELET_EVICTION_TRIGGERED",
			types.SeverityWarning,
			"Kubelet触发资源驱逐",
			1,
		},
		{
			regexp.MustCompile(`(?i)image pull failed|failed to pull image`),
			"镜像拉取失败",
			"KUBELET_IMAGE_PULL_FAILED",
			types.SeverityWarning,
			"容器镜像拉取失败",
			5,
		},
	}

	for _, p := range patterns {
		count := 0
		for _, line := range lines {
			if p.pattern.MatchString(line) {
				count++
			}
		}

		if count >= p.threshold {
			issue := types.NewIssue(
				p.severity,
				p.name,
				p.code,
				fmt.Sprintf("%s: 在 %s 中发现 %d 次", p.message, source, count),
				source,
			)
			issue.AnalyzerName = a.Name()
			issues = append(issues, *issue)
		}
	}

	return issues
}

// init registers all log analyzers
func init() {
	analyzer.Register(NewSyslogAnalyzer())
	analyzer.Register(NewJournalCtlAnalyzer())
	analyzer.Register(NewKubeletLogAnalyzer())
}
