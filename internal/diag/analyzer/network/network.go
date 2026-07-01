// Package network provides network analyzers
package network

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/kudig-io/klaw/internal/diag/analyzer"
	"github.com/kudig-io/klaw/internal/diag/types"
)

// InterfaceAnalyzer checks network interface status.
type InterfaceAnalyzer struct {
	*analyzer.BaseAnalyzer
}

// NewInterfaceAnalyzer creates a new network interface analyzer.
func NewInterfaceAnalyzer() *InterfaceAnalyzer {
	return &InterfaceAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer(
			"network.interface",
			"检查网卡接口状态",
			"network",
			[]types.DataMode{types.ModeOffline, types.ModeOnline},
		),
	}
}

// Analyze performs network interface analysis.
func (a *InterfaceAnalyzer) Analyze(_ context.Context, data *types.DiagnosticData) ([]types.Issue, error) {
	var issues []types.Issue

	networkInfo, ok := data.GetRawFile("network_info")
	if !ok {
		return issues, nil
	}

	content := string(networkInfo)

	// Find interfaces in DOWN state (excluding lo and veth*)
	downRe := regexp.MustCompile(`(\w+):\s+<[^>]*>\s+.*state DOWN`)
	matches := downRe.FindAllStringSubmatch(content, -1)

	var downInterfaces []string
	for _, match := range matches {
		if len(match) > 1 {
			iface := match[1]
			// Exclude loopback and veth interfaces
			if iface != "lo" && !strings.HasPrefix(iface, "veth") {
				downInterfaces = append(downInterfaces, iface)
			}
		}
	}

	if len(downInterfaces) > 0 {
		issue := types.NewIssue(
			types.SeverityWarning,
			"网卡接口down",
			"NETWORK_INTERFACE_DOWN",
			fmt.Sprintf("以下网卡处于down状态: %s", strings.Join(downInterfaces, ", ")),
			"network_info",
		).WithRemediation("检查网卡配置: ip link show; ip addr show")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	}

	return issues, nil
}

// RouteAnalyzer checks routing configuration
type RouteAnalyzer struct {
	*analyzer.BaseAnalyzer
}

// NewRouteAnalyzer creates a new route analyzer
func NewRouteAnalyzer() *RouteAnalyzer {
	return &RouteAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer(
			"network.route",
			"检查路由配置",
			"network",
			[]types.DataMode{types.ModeOffline, types.ModeOnline},
		),
	}
}

// Analyze performs route analysis
func (a *RouteAnalyzer) Analyze(_ context.Context, data *types.DiagnosticData) ([]types.Issue, error) {
	var issues []types.Issue

	networkInfo, ok := data.GetRawFile("network_info")
	if !ok {
		return issues, nil
	}

	content := string(networkInfo)

	if !strings.Contains(content, "default via") {
		issue := types.NewIssue(
			types.SeverityWarning,
			"缺少默认路由",
			"NO_DEFAULT_ROUTE",
			"未检测到默认路由配置",
			"network_info",
		).WithRemediation("检查路由表: ip route show; 添加默认路由: ip route add default via <gateway>")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	}

	return issues, nil
}

// PortAnalyzer checks critical port listening status
type PortAnalyzer struct {
	*analyzer.BaseAnalyzer
}

// NewPortAnalyzer creates a new port analyzer
func NewPortAnalyzer() *PortAnalyzer {
	return &PortAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer(
			"network.port",
			"检查关键端口监听状态",
			"network",
			[]types.DataMode{types.ModeOffline, types.ModeOnline},
		),
	}
}

// Analyze performs port analysis
func (a *PortAnalyzer) Analyze(_ context.Context, data *types.DiagnosticData) ([]types.Issue, error) {
	var issues []types.Issue

	systemStatus, ok := data.GetRawFile("system_status")
	if !ok {
		return issues, nil
	}

	content := string(systemStatus)

	// Check kubelet port 10250
	if !strings.Contains(content, ":10250") || !strings.Contains(content, "LISTEN") {
		// More flexible check
		kubeletListening := regexp.MustCompile(`:10250\s+.*LISTEN`).MatchString(content)
		if !kubeletListening {
			issue := types.NewIssue(
				types.SeverityCritical,
				"Kubelet端口未监听",
				"KUBELET_PORT_NOT_LISTENING",
				"10250端口未处于监听状态",
				"system_status",
			).WithRemediation("检查kubelet服务: systemctl status kubelet; journalctl -u kubelet")
			issue.AnalyzerName = a.Name()
			issues = append(issues, *issue)
		}
	}

	return issues, nil
}

// IptablesAnalyzer checks iptables rules count
type IptablesAnalyzer struct {
	*analyzer.BaseAnalyzer
}

// NewIptablesAnalyzer creates a new iptables analyzer
func NewIptablesAnalyzer() *IptablesAnalyzer {
	return &IptablesAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer(
			"network.iptables",
			"检查iptables规则数量",
			"network",
			[]types.DataMode{types.ModeOffline},
		),
	}
}

// Analyze performs iptables analysis
func (a *IptablesAnalyzer) Analyze(_ context.Context, data *types.DiagnosticData) ([]types.Issue, error) {
	var issues []types.Issue

	networkInfo, ok := data.GetRawFile("network_info")
	if !ok {
		return issues, nil
	}

	content := string(networkInfo)

	// Count iptables rules (lines starting with -A)
	ruleCount := strings.Count(content, "\n-A ")

	if ruleCount > 50000 {
		issue := types.NewIssue(
			types.SeverityWarning,
			"iptables规则过多",
			"TOO_MANY_IPTABLES_RULES",
			fmt.Sprintf("iptables规则数量: %d，可能影响性能", ruleCount),
			"network_info",
		).WithRemediation("检查iptables规则: iptables -L -n | wc -l; 考虑使用IPVS模式")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	}

	return issues, nil
}

// InodeAnalyzer checks inode usage (moved from system to network for organization)
type InodeAnalyzer struct {
	*analyzer.BaseAnalyzer
}

// NewInodeAnalyzer creates a new inode analyzer
func NewInodeAnalyzer() *InodeAnalyzer {
	return &InodeAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer(
			"network.inode",
			"检查Inode使用状态",
			"network",
			[]types.DataMode{types.ModeOffline},
		),
	}
}

// Analyze performs inode analysis
func (a *InodeAnalyzer) Analyze(_ context.Context, data *types.DiagnosticData) ([]types.Issue, error) {
	var issues []types.Issue

	systemStatus, ok := data.GetRawFile("system_status")
	if !ok {
		return issues, nil
	}

	content := string(systemStatus)

	// Parse df -i output for inode usage
	inodeRe := regexp.MustCompile(`(/\S+)\s+\d+\s+\d+\s+\d+\s+(\d+)%\s+(/\S*)`)
	matches := inodeRe.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) > 2 {
			usageStr := match[2]
			mountPoint := match[3]
			if mountPoint == "" {
				mountPoint = match[1]
			}

			usage, err := strconv.Atoi(usageStr)
			if err != nil {
				continue
			}

			if usage >= 90 {
				issue := types.NewIssue(
					types.SeverityWarning,
					"Inode使用率过高",
					"HIGH_INODE_USAGE",
					fmt.Sprintf("挂载点 %s 的inode使用率 %d%%", mountPoint, usage),
					"system_status",
				).WithRemediation(fmt.Sprintf("检查小文件数量: find %s -type f | wc -l", mountPoint))
				issue.AnalyzerName = a.Name()
				issues = append(issues, *issue)
			}
		}
	}

	return issues, nil
}

// init registers all network analyzers
func init() {
	_ = analyzer.Register(NewInterfaceAnalyzer())
	_ = analyzer.Register(NewRouteAnalyzer())
	_ = analyzer.Register(NewPortAnalyzer())
	_ = analyzer.Register(NewIptablesAnalyzer())
	_ = analyzer.Register(NewInodeAnalyzer())
}
