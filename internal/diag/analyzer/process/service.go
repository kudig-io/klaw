// Package process provides process and service analyzers
package process

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/kudig-io/klaw/internal/diag/analyzer"
	"github.com/kudig-io/klaw/internal/diag/types"
)

// KubeletAnalyzer checks kubelet service status
type KubeletAnalyzer struct {
	*analyzer.BaseAnalyzer
}

// NewKubeletAnalyzer creates a new kubelet analyzer
func NewKubeletAnalyzer() *KubeletAnalyzer {
	return &KubeletAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer(
			"process.kubelet",
			"检查Kubelet服务状态",
			"process",
			[]types.DataMode{types.ModeOffline, types.ModeOnline},
		),
	}
}

// Analyze performs kubelet service analysis
func (a *KubeletAnalyzer) Analyze(ctx context.Context, data *types.DiagnosticData) ([]types.Issue, error) {
	var issues []types.Issue

	kubeletStatus, ok := data.GetRawFile("daemon_status/kubelet_status")
	if !ok {
		return issues, nil
	}

	content := string(kubeletStatus)
	status := parseServiceStatus(content)

	switch status {
	case "failed":
		issue := types.NewIssue(
			types.SeverityCritical,
			"Kubelet服务未运行",
			"KUBELET_SERVICE_DOWN",
			"kubelet.service状态为failed",
			"daemon_status/kubelet_status",
		).WithRemediation("检查kubelet日志: journalctl -u kubelet -n 100; systemctl restart kubelet")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)

	case "stopped", "inactive":
		issue := types.NewIssue(
			types.SeverityCritical,
			"Kubelet服务停止",
			"KUBELET_SERVICE_STOPPED",
			"kubelet.service未启动",
			"daemon_status/kubelet_status",
		).WithRemediation("启动kubelet: systemctl start kubelet; 检查日志: journalctl -u kubelet")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	}

	return issues, nil
}

// ContainerRuntimeAnalyzer checks container runtime (docker/containerd) status
type ContainerRuntimeAnalyzer struct {
	*analyzer.BaseAnalyzer
}

// NewContainerRuntimeAnalyzer creates a new container runtime analyzer
func NewContainerRuntimeAnalyzer() *ContainerRuntimeAnalyzer {
	return &ContainerRuntimeAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer(
			"process.container_runtime",
			"检查容器运行时服务状态",
			"process",
			[]types.DataMode{types.ModeOffline, types.ModeOnline},
		),
	}
}

// Analyze performs container runtime analysis
func (a *ContainerRuntimeAnalyzer) Analyze(ctx context.Context, data *types.DiagnosticData) ([]types.Issue, error) {
	var issues []types.Issue

	dockerStatus := "unknown"
	containerdStatus := "unknown"

	if content, ok := data.GetRawFile("daemon_status/docker_status"); ok {
		dockerStatus = parseServiceStatus(string(content))
	}

	if content, ok := data.GetRawFile("daemon_status/containerd_status"); ok {
		containerdStatus = parseServiceStatus(string(content))
	}

	// Both failed
	if dockerStatus == "failed" && containerdStatus == "failed" {
		issue := types.NewIssue(
			types.SeverityCritical,
			"容器运行时服务异常",
			"CONTAINER_RUNTIME_DOWN",
			"docker和containerd服务均为failed状态",
			"daemon_status/",
		).WithRemediation("检查容器运行时日志: journalctl -u containerd -n 100; systemctl restart containerd")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	} else if (dockerStatus == "stopped" || dockerStatus == "inactive") &&
		(containerdStatus == "stopped" || containerdStatus == "inactive") {
		issue := types.NewIssue(
			types.SeverityCritical,
			"容器运行时服务停止",
			"CONTAINER_RUNTIME_STOPPED",
			"docker和containerd服务均未启动",
			"daemon_status/",
		).WithRemediation("启动容器运行时: systemctl start containerd")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	}

	return issues, nil
}

// RuncAnalyzer checks for runc process hang
type RuncAnalyzer struct {
	*analyzer.BaseAnalyzer
}

// NewRuncAnalyzer creates a new runc analyzer
func NewRuncAnalyzer() *RuncAnalyzer {
	return &RuncAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer(
			"process.runc",
			"检查runc进程状态",
			"process",
			[]types.DataMode{types.ModeOffline},
		),
	}
}

// Analyze performs runc process analysis
func (a *RuncAnalyzer) Analyze(ctx context.Context, data *types.DiagnosticData) ([]types.Issue, error) {
	var issues []types.Issue

	systemStatus, ok := data.GetRawFile("system_status")
	if !ok {
		return issues, nil
	}

	content := string(systemStatus)

	if strings.Contains(content, "runc process") && strings.Contains(content, "maybe hang") {
		runcCount := strings.Count(content, "maybe hang")
		issue := types.NewIssue(
			types.SeverityWarning,
			"runc进程可能挂起",
			"RUNC_PROCESS_HANG",
			fmt.Sprintf("检测到 %d 个runc进程可能处于挂起状态", runcCount),
			"system_status",
		).WithRemediation("检查runc进程: ps aux | grep runc; 可能需要重启容器运行时")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	}

	return issues, nil
}

// FirewalldAnalyzer checks firewalld service status
type FirewalldAnalyzer struct {
	*analyzer.BaseAnalyzer
}

// NewFirewalldAnalyzer creates a new firewalld analyzer
func NewFirewalldAnalyzer() *FirewalldAnalyzer {
	return &FirewalldAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer(
			"process.firewalld",
			"检查Firewalld服务状态",
			"process",
			[]types.DataMode{types.ModeOffline},
		),
	}
}

// Analyze performs firewalld analysis
func (a *FirewalldAnalyzer) Analyze(ctx context.Context, data *types.DiagnosticData) ([]types.Issue, error) {
	var issues []types.Issue

	serviceStatus, ok := data.GetRawFile("service_status")
	if !ok {
		return issues, nil
	}

	content := string(serviceStatus)

	if strings.Contains(content, "firewalld") && strings.Contains(content, "running") {
		issue := types.NewIssue(
			types.SeverityWarning,
			"Firewalld服务运行中",
			"FIREWALLD_RUNNING",
			"Kubernetes节点建议关闭firewalld服务",
			"service_status",
		).WithRemediation("关闭firewalld: systemctl stop firewalld && systemctl disable firewalld")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	}

	return issues, nil
}

// PIDLeakAnalyzer checks for PID/thread leak
type PIDLeakAnalyzer struct {
	*analyzer.BaseAnalyzer
}

// NewPIDLeakAnalyzer creates a new PID leak analyzer
func NewPIDLeakAnalyzer() *PIDLeakAnalyzer {
	return &PIDLeakAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer(
			"process.pid_leak",
			"检查进程/线程数泄漏",
			"process",
			[]types.DataMode{types.ModeOffline},
		),
	}
}

// Analyze performs PID leak analysis
func (a *PIDLeakAnalyzer) Analyze(ctx context.Context, data *types.DiagnosticData) ([]types.Issue, error) {
	var issues []types.Issue

	systemStatus, ok := data.GetRawFile("system_status")
	if !ok {
		return issues, nil
	}

	content := string(systemStatus)

	// Look for PID leak detection section
	pidLeakRe := regexp.MustCompile(`(?m)^(\d+)\s+.*pid leak detect`)
	matches := pidLeakRe.FindAllStringSubmatch(content, -1)

	var maxThreads int
	for _, match := range matches {
		if len(match) > 1 {
			var threads int
			fmt.Sscanf(match[1], "%d", &threads)
			if threads > maxThreads {
				maxThreads = threads
			}
		}
	}

	if maxThreads > 10000 {
		issue := types.NewIssue(
			types.SeverityCritical,
			"进程/线程数异常",
			"PID_LEAK_DETECTED",
			fmt.Sprintf("某进程线程数达到 %d", maxThreads),
			"system_status",
		).WithRemediation("检查线程泄漏进程: ps -elT | awk '{print $3}' | sort | uniq -c | sort -rn | head")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	} else if maxThreads > 5000 {
		issue := types.NewIssue(
			types.SeverityWarning,
			"进程/线程数偏高",
			"HIGH_THREAD_COUNT",
			fmt.Sprintf("某进程线程数达到 %d", maxThreads),
			"system_status",
		).WithRemediation("关注线程数趋势，检查是否存在泄漏")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	}

	return issues, nil
}

// parseServiceStatus extracts service status from systemctl output
func parseServiceStatus(content string) string {
	content = strings.ToLower(content)

	if strings.Contains(content, "active: active (running)") ||
		strings.Contains(content, "active (running)") {
		return "running"
	}
	if strings.Contains(content, "active: failed") ||
		strings.Contains(content, "status=1/failure") {
		return "failed"
	}
	if strings.Contains(content, "active: inactive") ||
		strings.Contains(content, "inactive (dead)") {
		return "stopped"
	}
	return "unknown"
}

// init registers all process analyzers
func init() {
	analyzer.Register(NewKubeletAnalyzer())
	analyzer.Register(NewContainerRuntimeAnalyzer())
	analyzer.Register(NewRuncAnalyzer())
	analyzer.Register(NewFirewalldAnalyzer())
	analyzer.Register(NewPIDLeakAnalyzer())
}
