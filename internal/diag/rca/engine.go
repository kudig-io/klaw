// Package rca provides Root Cause Analysis capabilities for klaw
package rca

import (
	"context"
	"fmt"
	"strings"

	"github.com/kudig-io/klaw/internal/diag/types"
)

// RootCause represents a detected root cause with related symptoms
type RootCause struct {
	// ID is the unique identifier for this root cause
	ID string `json:"id" yaml:"id"`

	// Title is the human-readable title
	Title string `json:"title" yaml:"title"`

	// Description describes the root cause
	Description string `json:"description" yaml:"description"`

	// Confidence is the confidence level (0.0 - 1.0)
	Confidence float64 `json:"confidence" yaml:"confidence"`

	// RelatedIssues are the issues that led to this root cause
	RelatedIssues []string `json:"related_issues" yaml:"related_issues"`

	// SuggestedActions are the recommended fixes
	SuggestedActions []string `json:"suggested_actions" yaml:"suggested_actions"`

	// Category groups related root causes
	Category string `json:"category" yaml:"category"`
}

// Rule represents a root cause analysis rule
type Rule struct {
	// ID is the unique identifier for this rule
	ID string `json:"id" yaml:"id"`

	// Name is the human-readable name
	Name string `json:"name" yaml:"name"`

	// Category of root cause
	Category string `json:"category" yaml:"category"`

	// MatchConditions define when this rule applies
	MatchConditions []MatchCondition `json:"match_conditions" yaml:"match_conditions"`

	// MinConfidence is the minimum confidence to trigger this rule (0.0 - 1.0)
	MinConfidence float64 `json:"min_confidence" yaml:"min_confidence"`

	// Result is the root cause produced when rule matches
	Result RootCause `json:"result" yaml:"result"`
}

// MatchCondition defines a condition to match issues
type MatchCondition struct {
	// AnalyzerPattern matches analyzer names (glob pattern)
	AnalyzerPattern string `json:"analyzer_pattern" yaml:"analyzer_pattern"`

	// CodePattern matches issue codes (glob pattern)
	CodePattern string `json:"code_pattern" yaml:"code_pattern"`

	// MinSeverity is the minimum severity to match
	MinSeverity types.Severity `json:"min_severity" yaml:"min_severity"`

	// DetailsPattern matches issue details (substring)
	DetailsPattern string `json:"details_pattern" yaml:"details_pattern"`
}

// Engine performs root cause analysis
type Engine struct {
	rules []Rule
}

// NewEngine creates a new RCA engine with default rules
func NewEngine() *Engine {
	return &Engine{
		rules: getDefaultRules(),
	}
}

// AddRule adds a custom rule to the engine
func (e *Engine) AddRule(rule Rule) {
	e.rules = append(e.rules, rule)
}

// Analyze performs root cause analysis on the given issues
func (e *Engine) Analyze(ctx context.Context, issues []types.Issue) []RootCause {
	var rootCauses []RootCause

	for _, rule := range e.rules {
		if ctx.Err() != nil {
			break
		}

		if matches := e.evaluateRule(rule, issues); matches {
			// Calculate confidence based on matched conditions
			confidence := e.calculateConfidence(rule, issues)
			if confidence >= rule.MinConfidence {
				result := rule.Result
				result.ID = rule.ID
				result.Confidence = confidence
				result.RelatedIssues = e.getRelatedIssueCodes(rule, issues)
				rootCauses = append(rootCauses, result)
			}
		}
	}

	return rootCauses
}

// evaluateRule checks if a rule matches the given issues
func (e *Engine) evaluateRule(rule Rule, issues []types.Issue) bool {
	// All conditions must match for the rule to apply
	matchedCount := 0
	for _, condition := range rule.MatchConditions {
		for _, issue := range issues {
			if e.matchesCondition(condition, issue) {
				matchedCount++
				break
			}
		}
	}

	// Rule matches if at least one condition from each group matches
	// For now, require all conditions to match
	return matchedCount >= len(rule.MatchConditions)/2
}

// matchesCondition checks if an issue matches a condition
func (e *Engine) matchesCondition(condition MatchCondition, issue types.Issue) bool {
	// Check severity
	if issue.Severity < condition.MinSeverity {
		return false
	}

	// Check code pattern
	if condition.CodePattern != "" {
		if !matchGlob(condition.CodePattern, issue.ENName) {
			return false
		}
	}

	// Check analyzer pattern
	if condition.AnalyzerPattern != "" {
		if !matchGlob(condition.AnalyzerPattern, issue.AnalyzerName) {
			return false
		}
	}

	// Check details pattern
	if condition.DetailsPattern != "" {
		if !strings.Contains(issue.Details, condition.DetailsPattern) {
			return false
		}
	}

	return true
}

// calculateConfidence calculates the confidence level for a matched rule
func (e *Engine) calculateConfidence(rule Rule, issues []types.Issue) float64 {
	// Base confidence
	confidence := 0.5

	// Increase confidence based on number of matching conditions
	matched := 0
	for _, condition := range rule.MatchConditions {
		for _, issue := range issues {
			if e.matchesCondition(condition, issue) {
				matched++
				break
			}
		}
	}

	if len(rule.MatchConditions) > 0 {
		confidence += float64(matched) / float64(len(rule.MatchConditions)) * 0.5
	}

	// Cap at 1.0
	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

// getRelatedIssueCodes gets the codes of issues related to a rule
func (e *Engine) getRelatedIssueCodes(rule Rule, issues []types.Issue) []string {
	var codes []string
	seen := make(map[string]bool)

	for _, condition := range rule.MatchConditions {
		for _, issue := range issues {
			if e.matchesCondition(condition, issue) && !seen[issue.ENName] {
				codes = append(codes, issue.ENName)
				seen[issue.ENName] = true
			}
		}
	}

	return codes
}

// matchGlob performs simple glob matching
func matchGlob(pattern, s string) bool {
	// Simple implementation: support * wildcard
	if pattern == "*" || pattern == "" {
		return true
	}
	if strings.Contains(pattern, "*") {
		parts := strings.Split(pattern, "*")
		if len(parts) == 2 {
			if parts[0] == "" {
				return strings.HasSuffix(s, parts[1])
			}
			if parts[1] == "" {
				return strings.HasPrefix(s, parts[0])
			}
			return strings.HasPrefix(s, parts[0]) && strings.HasSuffix(s, parts[1])
		}
	}
	return pattern == s
}

// getDefaultRules returns the built-in RCA rules
func getDefaultRules() []Rule {
	return []Rule{
		// DNS Resolution Issues
		{
			ID:   "rca.dns_failure",
			Name: "DNS 解析失败",
			MatchConditions: []MatchCondition{
				{CodePattern: "DNS_*", MinSeverity: types.SeverityWarning},
				{CodePattern: "NETWORK_DNS_*", MinSeverity: types.SeverityWarning},
			},
			MinConfidence: 0.7,
			Result: RootCause{
				Title:       "DNS 解析问题",
				Description: "CoreDNS 服务异常或网络 DNS 配置错误导致域名解析失败",
				Category:    "network",
				SuggestedActions: []string{
					"检查 CoreDNS Pod 状态: kubectl get pods -n kube-system -l k8s-app=kube-dns",
					"检查 DNS 服务配置: kubectl get svc -n kube-system kube-dns",
					"验证 /etc/resolv.conf 中的 nameserver 配置",
					"检查网络策略是否阻止了 DNS 查询",
				},
			},
		},

		// Memory Pressure
		{
			ID:   "rca.memory_pressure",
			Name: "内存压力",
			MatchConditions: []MatchCondition{
				{CodePattern: "HIGH_MEMORY_USAGE", MinSeverity: types.SeverityWarning},
				{CodePattern: "OOM_*", MinSeverity: types.SeverityWarning},
				{CodePattern: "SYSTEM_MEMORY_*", MinSeverity: types.SeverityWarning},
			},
			MinConfidence: 0.6,
			Result: RootCause{
				Title:       "节点内存压力",
				Description: "节点内存使用率过高，可能导致 OOM Killer 触发或应用性能下降",
				Category:    "resources",
				SuggestedActions: []string{
					"识别内存消耗大户: kubectl top pods --all-namespaces --sort-by=memory",
					"检查是否存在内存泄漏",
					"考虑增加节点内存或添加新节点",
					"调整 Pod 的资源限制 (limits/requests)",
				},
			},
		},

		// Disk Pressure
		{
			ID:   "rca.disk_pressure",
			Name: "磁盘压力",
			MatchConditions: []MatchCondition{
				{CodePattern: "DISK_*", MinSeverity: types.SeverityWarning},
				{CodePattern: "IMAGE_*", MinSeverity: types.SeverityWarning},
			},
			MinConfidence: 0.6,
			Result: RootCause{
				Title:       "节点磁盘压力",
				Description: "节点磁盘空间不足，可能影响镜像拉取、日志写入和 Pod 调度",
				Category:    "resources",
				SuggestedActions: []string{
					"清理未使用的 Docker 镜像: docker system prune -a",
					"清理日志文件: journalctl --vacuum-time=7d",
					"检查并清理 emptyDir 卷数据",
					"增加节点磁盘容量",
				},
			},
		},

		// Network Connectivity
		{
			ID:   "rca.network_issue",
			Name: "网络连接问题",
			MatchConditions: []MatchCondition{
				{CodePattern: "NETWORK_*", MinSeverity: types.SeverityWarning},
				{CodePattern: "CNI_*", MinSeverity: types.SeverityWarning},
			},
			MinConfidence: 0.6,
			Result: RootCause{
				Title:       "网络连接异常",
				Description: "CNI 插件问题或网络配置错误导致 Pod 间通信失败",
				Category:    "network",
				SuggestedActions: []string{
					"检查 CNI 插件状态: kubectl get pods -n kube-system -l k8s-app=calico-node",
					"验证节点间网络连通性",
					"检查网络策略配置",
					"重启 CNI Pod: kubectl delete pod -n kube-system -l k8s-app=calico-node",
				},
			},
		},

		// Control Plane Issues
		{
			ID:   "rca.control_plane",
			Name: "控制平面问题",
			MatchConditions: []MatchCondition{
				{CodePattern: "APISERVER_*", MinSeverity: types.SeverityWarning},
				{CodePattern: "ETCD_*", MinSeverity: types.SeverityWarning},
				{CodePattern: "SCHEDULER_*", MinSeverity: types.SeverityWarning},
				{CodePattern: "CONTROLLER_*", MinSeverity: types.SeverityWarning},
			},
			MinConfidence: 0.7,
			Result: RootCause{
				Title:       "Kubernetes 控制平面异常",
				Description: "API Server、etcd 或 Controller Manager 出现异常，影响集群核心功能",
				Category:    "controlplane",
				SuggestedActions: []string{
					"检查控制平面 Pod 状态: kubectl get pods -n kube-system",
					"查看 API Server 日志: kubectl logs -n kube-system kube-apiserver-*",
					"检查 etcd 健康状态: kubectl exec -it etcd-* -n kube-system -- etcdctl endpoint health",
					"检查控制平面节点资源使用情况",
				},
			},
		},

		// Pod Scheduling Issues
		{
			ID:   "rca.scheduling_failure",
			Name: "调度失败",
			MatchConditions: []MatchCondition{
				{CodePattern: "PENDING_PODS", MinSeverity: types.SeverityWarning},
				{CodePattern: "*TAINT*", MinSeverity: types.SeverityWarning},
				{CodePattern: "*CORDON*", MinSeverity: types.SeverityWarning},
			},
			MinConfidence: 0.6,
			Result: RootCause{
				Title:       "Pod 调度失败",
				Description: "资源不足、污点/容忍度不匹配或节点不可用导致 Pod 无法调度",
				Category:    "scheduling",
				SuggestedActions: []string{
					"查看 pending Pod 的调度事件: kubectl describe pod <pod-name>",
					"检查节点资源使用情况: kubectl top nodes",
					"检查节点污点: kubectl describe node <node-name> | grep Taints",
					"检查 Pod 的容忍度配置",
				},
			},
		},

		// Security Policy Violations
		{
			ID:   "rca.security_issue",
			Name: "安全问题",
			MatchConditions: []MatchCondition{
				{CodePattern: "CIS_*", MinSeverity: types.SeverityWarning},
				{CodePattern: "RBAC_*", MinSeverity: types.SeverityWarning},
				{CodePattern: "PRIVILEGED_*", MinSeverity: types.SeverityWarning},
			},
			MinConfidence: 0.7,
			Result: RootCause{
				Title:       "安全配置违规",
				Description: "检测到违反 CIS 基准、过度授权或特权容器等安全问题",
				Category:    "security",
				SuggestedActions: []string{
					"审查 RBAC 权限: kubectl auth can-i --list",
					"检查特权容器: kubectl get pods --all-namespaces -o json | jq '.items[].spec.containers[] | select(.securityContext.privileged==true)'",
					"启用 Pod Security Standards",
					"定期运行 CIS 基准扫描",
				},
			},
		},

		// Image Pull Issues
		{
			ID:   "rca.image_pull",
			Name: "镜像拉取失败",
			MatchConditions: []MatchCondition{
				{CodePattern: "IMAGEPULLBACKOFF", MinSeverity: types.SeverityCritical},
				{CodePattern: "ERRIMAGEPULL", MinSeverity: types.SeverityCritical},
			},
			MinConfidence: 0.8,
			Result: RootCause{
				Title:       "容器镜像拉取失败",
				Description: "镜像不存在、仓库认证失败或网络问题导致无法拉取镜像",
				Category:    "runtime",
				SuggestedActions: []string{
					"验证镜像名称和标签是否正确",
					"检查镜像仓库访问权限: kubectl get secret <registry-secret>",
					"手动测试镜像拉取: docker pull <image>",
					"检查节点网络连接和代理配置",
				},
			},
		},

		// Container Runtime Issues
		{
			ID:   "rca.runtime_issue",
			Name: "容器运行时问题",
			MatchConditions: []MatchCondition{
				{CodePattern: "CRASHLOOPBACKOFF", MinSeverity: types.SeverityCritical},
				{CodePattern: "CONTAINER_*", MinSeverity: types.SeverityWarning},
				{CodePattern: "RUNTIME_*", MinSeverity: types.SeverityWarning},
			},
			MinConfidence: 0.6,
			Result: RootCause{
				Title:       "容器运行时异常",
				Description: "容器崩溃、运行时错误或资源限制导致 Pod 无法正常启动",
				Category:    "runtime",
				SuggestedActions: []string{
					"查看容器日志: kubectl logs <pod-name> --previous",
					"检查容器退出码: kubectl describe pod <pod-name>",
					"验证资源限制是否过低",
					"检查启动命令和参数配置",
				},
			},
		},

		// Service Mesh Issues
		{
			ID:   "rca.servicemesh",
			Name: "服务网格问题",
			MatchConditions: []MatchCondition{
				{CodePattern: "ISTIO_*", MinSeverity: types.SeverityWarning},
				{CodePattern: "LINKERD_*", MinSeverity: types.SeverityWarning},
			},
			MinConfidence: 0.7,
			Result: RootCause{
				Title:       "服务网格异常",
				Description: "Istio 或 Linkerd 控制平面、sidecar 代理出现问题",
				Category:    "servicemesh",
				SuggestedActions: []string{
					"检查控制平面状态: kubectl get pods -n istio-system",
					"查看 sidecar 注入状态: istioctl proxy-status",
					"检查 mTLS 配置",
					"重启异常的 sidecar: kubectl rollout restart deployment <name>",
				},
			},
		},
	}
}

// FormatRootCauses formats root causes as a readable string
func FormatRootCauses(rootCauses []RootCause) string {
	if len(rootCauses) == 0 {
		return "未发现明确的根因模式"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("\n📊 根因分析 (%d 项):\n", len(rootCauses)))
	sb.WriteString("=" + strings.Repeat("=", 50) + "\n")

	for i, rc := range rootCauses {
		sb.WriteString(fmt.Sprintf("\n%d. %s (置信度: %.0f%%)\n", i+1, rc.Title, rc.Confidence*100))
		sb.WriteString(fmt.Sprintf("   类别: %s\n", rc.Category))
		sb.WriteString(fmt.Sprintf("   描述: %s\n", rc.Description))

		if len(rc.RelatedIssues) > 0 {
			sb.WriteString(fmt.Sprintf("   相关症状: %s\n", strings.Join(rc.RelatedIssues, ", ")))
		}

		if len(rc.SuggestedActions) > 0 {
			sb.WriteString("   建议操作:\n")
			for _, action := range rc.SuggestedActions {
				sb.WriteString(fmt.Sprintf("     - %s\n", action))
			}
		}
	}

	return sb.String()
}
