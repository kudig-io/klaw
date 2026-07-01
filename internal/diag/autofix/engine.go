// Package autofix provides automatic remediation capabilities
package autofix

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/kudig-io/klaw/internal/diag/types"
)

// FixAction represents an automatic fix action
type FixAction struct {
	// IssueCode is the issue code this fix applies to
	IssueCode string

	// Description describes what the fix does
	Description string

	// Command is the shell command to execute
	Command string

	// Risk level: low, medium, high
	Risk string

	// ConfirmationRequired if true, requires user confirmation
	ConfirmationRequired bool

	// CheckCommand verifies if fix is needed (optional)
	CheckCommand string
}

// FixResult represents the result of a fix operation
type FixResult struct {
	IssueCode   string
	Success     bool
	Message     string
	BeforeState string
	AfterState  string
}

// allowedCommands defines the set of safe commands permitted for auto-fix.
// Any command not matching this prefix will be rejected.
var allowedCommands = []string{
	"swapoff",
	"sed ",
	"docker system prune",
	"journalctl --vacuum",
	"kubectl delete pod",
	"kubectl rollout restart",
}

// isCommandAllowed checks if a command matches one of the allowed prefixes.
func isCommandAllowed(cmd string) bool {
	for _, allowed := range allowedCommands {
		if strings.HasPrefix(cmd, allowed) {
			return true
		}
	}
	return false
}

// Engine manages automatic fixes
type Engine struct {
	actions map[string][]FixAction
	dryRun  bool
}

// NewEngine creates a new auto-fix engine
func NewEngine(dryRun bool) *Engine {
	return &Engine{
		actions: getDefaultFixActions(),
		dryRun:  dryRun,
	}
}

// CanFix checks if an issue can be automatically fixed
func (e *Engine) CanFix(issue types.Issue) (FixAction, bool) {
	actions, exists := e.actions[issue.ENName]
	if !exists || len(actions) == 0 {
		return FixAction{}, false
	}

	// Return the first (safest) action
	return actions[0], true
}

// Fix attempts to fix an issue
func (e *Engine) Fix(ctx context.Context, issue types.Issue) FixResult {
	action, canFix := e.CanFix(issue)
	if !canFix {
		return FixResult{
			IssueCode: issue.ENName,
			Success:   false,
			Message:   "No automatic fix available for this issue",
		}
	}

	if action.ConfirmationRequired && !e.dryRun {
		return FixResult{
			IssueCode: issue.ENName,
			Success:   false,
			Message:   fmt.Sprintf("Fix requires confirmation: %s (use --confirm to proceed)", action.Description),
		}
	}

	if e.dryRun {
		return FixResult{
			IssueCode: issue.ENName,
			Success:   true,
			Message:   fmt.Sprintf("[DRY-RUN] Would execute: %s", action.Command),
		}
	}

	// Validate command against allowlist to prevent injection
	if !isCommandAllowed(action.Command) {
		return FixResult{
			IssueCode: issue.ENName,
			Success:   false,
			Message:   fmt.Sprintf("Command not in allowlist: %s", action.Command),
		}
	}

	// Execute the fix command
	cmd := exec.CommandContext(ctx, "sh", "-c", action.Command)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return FixResult{
			IssueCode: issue.ENName,
			Success:   false,
			Message:   fmt.Sprintf("Fix failed: %v - %s", err, string(output)),
		}
	}

	return FixResult{
		IssueCode: issue.ENName,
		Success:   true,
		Message:   fmt.Sprintf("Fix applied successfully: %s", strings.TrimSpace(string(output))),
	}
}

// FixAll attempts to fix all applicable issues
func (e *Engine) FixAll(ctx context.Context, issues []types.Issue) []FixResult {
	var results []FixResult

	for _, issue := range issues {
		if _, canFix := e.CanFix(issue); canFix {
			result := e.Fix(ctx, issue)
			results = append(results, result)
		}
	}

	return results
}

// GetFixableIssues returns issues that can be auto-fixed
func (e *Engine) GetFixableIssues(issues []types.Issue) []types.Issue {
	var fixable []types.Issue
	for _, issue := range issues {
		if _, canFix := e.CanFix(issue); canFix {
			fixable = append(fixable, issue)
		}
	}
	return fixable
}

// getDefaultFixActions returns the built-in fix actions
func getDefaultFixActions() map[string][]FixAction {
	return map[string][]FixAction{
		"SWAP_NOT_DISABLED": {
			{
				IssueCode:            "SWAP_NOT_DISABLED",
				Description:          "Disable swap on the node",
				Command:              "swapoff -a && sed -i '/swap/d' /etc/fstab",
				Risk:                 "low",
				ConfirmationRequired: true,
			},
		},
		"DOCKER_IMAGE_CLEANUP": {
			{
				IssueCode:            "DOCKER_IMAGE_CLEANUP",
				Description:          "Clean up unused Docker images",
				Command:              "docker system prune -af --volumes",
				Risk:                 "medium",
				ConfirmationRequired: true,
			},
		},
		"JOURNAL_LOGS_CLEANUP": {
			{
				IssueCode:            "JOURNAL_LOGS_CLEANUP",
				Description:          "Clean up old journal logs",
				Command:              "journalctl --vacuum-time=7d",
				Risk:                 "low",
				ConfirmationRequired: false,
			},
		},
		"CNI_RESTART": {
			{
				IssueCode:            "CNI_RESTART",
				Description:          "Restart CNI pods",
				Command:              "kubectl delete pod -n kube-system -l k8s-app=calico-node",
				Risk:                 "medium",
				ConfirmationRequired: true,
			},
		},
		"COREDNS_RESTART": {
			{
				IssueCode:            "COREDNS_RESTART",
				Description:          "Restart CoreDNS pods",
				Command:              "kubectl rollout restart deployment coredns -n kube-system",
				Risk:                 "low",
				ConfirmationRequired: false,
			},
		},
	}
}

// FormatResults formats fix results for display
func FormatResults(results []FixResult) string {
	if len(results) == 0 {
		return "没有可自动修复的问题"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("\n🔧 自动修复结果 (%d 项):\n", len(results)))
	sb.WriteString("=" + strings.Repeat("=", 50) + "\n")

	successCount := 0
	for _, result := range results {
		status := "❌ 失败"
		if result.Success {
			status = "✅ 成功"
			successCount++
		}
		sb.WriteString(fmt.Sprintf("\n%s %s\n", status, result.IssueCode))
		sb.WriteString(fmt.Sprintf("   %s\n", result.Message))
	}

	sb.WriteString(fmt.Sprintf("\n总计: %d 成功 / %d 尝试\n", successCount, len(results)))
	return sb.String()
}
