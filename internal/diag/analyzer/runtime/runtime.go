// Package runtime provides container runtime analyzers
package runtime

import (
	"context"
	"fmt"
	"strings"

	"github.com/kudig-io/klaw/internal/diag/analyzer"
	"github.com/kudig-io/klaw/internal/diag/types"
)

// DockerAnalyzer checks for Docker-specific issues
type DockerAnalyzer struct {
	*analyzer.BaseAnalyzer
}

// NewDockerAnalyzer creates a new Docker analyzer
func NewDockerAnalyzer() *DockerAnalyzer {
	return &DockerAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer(
			"runtime.docker",
			"检查Docker运行时状态",
			"runtime",
			[]types.DataMode{types.ModeOffline},
		),
	}
}

// Analyze performs Docker analysis
func (a *DockerAnalyzer) Analyze(ctx context.Context, data *types.DiagnosticData) ([]types.Issue, error) {
	var issues []types.Issue

	dockerLog, ok := data.GetRawFile("logs/docker.log")
	if !ok {
		return issues, nil
	}

	content := string(dockerLog)

	// Check for Docker start failure
	if strings.Contains(content, "Failed to start") {
		issue := types.NewIssue(
			types.SeverityCritical,
			"Docker启动失败",
			"DOCKER_START_FAILED",
			"Docker服务启动失败",
			"logs/docker.log",
		).WithRemediation("检查Docker日志: journalctl -u docker; 检查存储驱动配置")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	}

	// Check for storage driver errors
	if strings.Contains(strings.ToLower(content), "storage driver") &&
		strings.Contains(strings.ToLower(content), "error") {
		issue := types.NewIssue(
			types.SeverityCritical,
			"Docker存储驱动错误",
			"DOCKER_STORAGE_DRIVER_ERROR",
			"Docker存储驱动出现错误",
			"logs/docker.log",
		).WithRemediation("检查存储驱动: docker info | grep Storage; 检查磁盘空间")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	}

	return issues, nil
}

// ContainerdAnalyzer checks for containerd-specific issues
type ContainerdAnalyzer struct {
	*analyzer.BaseAnalyzer
}

// NewContainerdAnalyzer creates a new containerd analyzer
func NewContainerdAnalyzer() *ContainerdAnalyzer {
	return &ContainerdAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer(
			"runtime.containerd",
			"检查Containerd运行时状态",
			"runtime",
			[]types.DataMode{types.ModeOffline},
		),
	}
}

// Analyze performs containerd analysis
func (a *ContainerdAnalyzer) Analyze(ctx context.Context, data *types.DiagnosticData) ([]types.Issue, error) {
	var issues []types.Issue

	containerdLog, ok := data.GetRawFile("logs/containerd.log")
	if !ok {
		return issues, nil
	}

	content := string(containerdLog)

	// Check for container creation failures
	createFailCount := strings.Count(content, "failed to create")
	createFailCount += strings.Count(content, "Failed to create")

	if createFailCount > 10 {
		issue := types.NewIssue(
			types.SeverityWarning,
			"容器创建失败率高",
			"CONTAINER_CREATE_FAILED",
			fmt.Sprintf("容器创建失败 %d 次", createFailCount),
			"logs/containerd.log",
		).WithRemediation("检查containerd日志: journalctl -u containerd; 检查资源限制")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	}

	return issues, nil
}

// TimeSyncAnalyzer checks time synchronization service status
type TimeSyncAnalyzer struct {
	*analyzer.BaseAnalyzer
}

// NewTimeSyncAnalyzer creates a new time sync analyzer
func NewTimeSyncAnalyzer() *TimeSyncAnalyzer {
	return &TimeSyncAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer(
			"runtime.time_sync",
			"检查时间同步服务状态",
			"runtime",
			[]types.DataMode{types.ModeOffline},
		),
	}
}

// Analyze performs time sync analysis
func (a *TimeSyncAnalyzer) Analyze(ctx context.Context, data *types.DiagnosticData) ([]types.Issue, error) {
	var issues []types.Issue

	serviceStatus, ok := data.GetRawFile("service_status")
	if !ok {
		return issues, nil
	}

	content := strings.ToLower(string(serviceStatus))

	ntpdRunning := strings.Contains(content, "ntpd") && strings.Contains(content, "running")
	chronydRunning := strings.Contains(content, "chronyd") && strings.Contains(content, "running")

	if !ntpdRunning && !chronydRunning {
		issue := types.NewIssue(
			types.SeverityInfo,
			"时间同步服务未运行",
			"TIME_SYNC_SERVICE_DOWN",
			"ntpd和chronyd服务均未运行",
			"service_status",
		).WithRemediation("启用时间同步: systemctl enable --now chronyd 或 systemctl enable --now ntpd")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	}

	return issues, nil
}

// ConfigAnalyzer checks system configuration for K8s compatibility
type ConfigAnalyzer struct {
	*analyzer.BaseAnalyzer
}

// NewConfigAnalyzer creates a new config analyzer
func NewConfigAnalyzer() *ConfigAnalyzer {
	return &ConfigAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer(
			"runtime.config",
			"检查系统配置",
			"runtime",
			[]types.DataMode{types.ModeOffline},
		),
	}
}

// Analyze performs system configuration analysis
func (a *ConfigAnalyzer) Analyze(ctx context.Context, data *types.DiagnosticData) ([]types.Issue, error) {
	var issues []types.Issue

	systemInfo, ok := data.GetRawFile("system_info")
	if !ok {
		return issues, nil
	}

	content := string(systemInfo)

	// Check IP forwarding
	if strings.Contains(content, "net.ipv4.ip_forward = 0") {
		issue := types.NewIssue(
			types.SeverityWarning,
			"IP转发未启用",
			"IP_FORWARD_DISABLED",
			"net.ipv4.ip_forward = 0，Kubernetes需要启用",
			"system_info",
		).WithRemediation("启用IP转发: sysctl -w net.ipv4.ip_forward=1; 写入/etc/sysctl.conf永久生效")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	}

	// Check bridge-nf-call-iptables
	if strings.Contains(content, "net.bridge.bridge-nf-call-iptables = 0") {
		issue := types.NewIssue(
			types.SeverityWarning,
			"bridge-nf-call-iptables未启用",
			"BRIDGE_NF_CALL_IPTABLES_DISABLED",
			"net.bridge.bridge-nf-call-iptables = 0，Kubernetes需要启用",
			"system_info",
		).WithRemediation("启用: sysctl -w net.bridge.bridge-nf-call-iptables=1")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	}

	// Check ulimit
	if strings.Contains(content, "open files") && strings.Contains(content, "1024") {
		issue := types.NewIssue(
			types.SeverityInfo,
			"文件句柄限制过低",
			"LOW_ULIMIT_NOFILE",
			"open files限制为1024，建议设置为65536或更高",
			"system_info",
		).WithRemediation("增加限制: ulimit -n 65536; 修改/etc/security/limits.conf永久生效")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	}

	// Check SELinux
	if strings.Contains(strings.ToLower(content), "selinux") &&
		strings.Contains(strings.ToLower(content), "enforcing") {
		issue := types.NewIssue(
			types.SeverityInfo,
			"SELinux处于Enforcing模式",
			"SELINUX_ENFORCING",
			"SELinux处于Enforcing模式，可能影响Kubernetes运行",
			"system_info",
		).WithRemediation("设置为permissive: setenforce 0; 或修改/etc/selinux/config")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	}

	return issues, nil
}

// init registers all runtime analyzers
func init() {
	analyzer.Register(NewDockerAnalyzer())
	analyzer.Register(NewContainerdAnalyzer())
	analyzer.Register(NewTimeSyncAnalyzer())
	analyzer.Register(NewConfigAnalyzer())
}
