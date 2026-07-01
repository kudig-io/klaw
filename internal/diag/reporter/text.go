// Package reporter provides report generation functionality
package reporter

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/kudig-io/klaw/internal/diag/types"
)

// TextReporter generates text format reports
type TextReporter struct {
	// UseColor controls whether to use ANSI colors
	UseColor bool
}

// NewTextReporter creates a new text reporter
func NewTextReporter(useColor bool) *TextReporter {
	return &TextReporter{UseColor: useColor}
}

// Format returns the output format name
func (r *TextReporter) Format() string {
	return "text"
}

// ANSI color codes
const (
	colorRed    = "\033[0;31m"
	colorYellow = "\033[1;33m"
	colorBlue   = "\033[0;34m"
	colorGreen  = "\033[0;32m"
	colorReset  = "\033[0m"
)

// Generate creates a text report from the issues
func (r *TextReporter) Generate(issues []types.Issue, metadata *ReportMetadata) ([]byte, error) {
	var buf bytes.Buffer

	// Header
	buf.WriteString("=== Kubernetes节点诊断异常报告 ===\n")
	buf.WriteString(fmt.Sprintf("诊断时间: %s\n", metadata.Timestamp.Format("2006-01-02 15:04:05")))
	buf.WriteString(fmt.Sprintf("节点信息: %s\n", metadata.Hostname))
	buf.WriteString(fmt.Sprintf("分析目录: %s\n", metadata.DiagnosePath))
	buf.WriteString(fmt.Sprintf("分析引擎: %s\n", metadata.Engine))
	buf.WriteString("\n")

	if len(issues) == 0 {
		if r.UseColor {
			buf.WriteString(fmt.Sprintf("%s✓ 未检测到异常%s\n", colorGreen, colorReset))
		} else {
			buf.WriteString("✓ 未检测到异常\n")
		}
		buf.WriteString("\n节点状态良好！\n")
		return buf.Bytes(), nil
	}

	// Group issues by severity
	critical := filterBySeverity(issues, types.SeverityCritical)
	warning := filterBySeverity(issues, types.SeverityWarning)
	info := filterBySeverity(issues, types.SeverityInfo)

	// Critical issues
	if len(critical) > 0 {
		buf.WriteString("-------------------------------------------\n")
		buf.WriteString("【严重级别】异常项\n")
		buf.WriteString("-------------------------------------------\n")
		for _, issue := range critical {
			r.writeIssue(&buf, issue, colorRed, "严重")
		}
	}

	// Warning issues
	if len(warning) > 0 {
		buf.WriteString("-------------------------------------------\n")
		buf.WriteString("【警告级别】异常项\n")
		buf.WriteString("-------------------------------------------\n")
		for _, issue := range warning {
			r.writeIssue(&buf, issue, colorYellow, "警告")
		}
	}

	// Info issues
	if len(info) > 0 {
		buf.WriteString("-------------------------------------------\n")
		buf.WriteString("【提示级别】异常项\n")
		buf.WriteString("-------------------------------------------\n")
		for _, issue := range info {
			r.writeIssue(&buf, issue, colorBlue, "提示")
		}
	}

	// Summary
	summary := types.CalculateSummary(issues)
	buf.WriteString("-------------------------------------------\n")
	buf.WriteString("异常统计\n")
	buf.WriteString("-------------------------------------------\n")
	buf.WriteString(fmt.Sprintf("严重: %d项\n", summary.Critical))
	buf.WriteString(fmt.Sprintf("警告: %d项\n", summary.Warning))
	buf.WriteString(fmt.Sprintf("提示: %d项\n", summary.Info))
	buf.WriteString(fmt.Sprintf("总计: %d项\n", summary.Total))

	return buf.Bytes(), nil
}

// writeIssue writes a single issue to the buffer
func (r *TextReporter) writeIssue(buf *bytes.Buffer, issue types.Issue, color, label string) {
	if r.UseColor {
		buf.WriteString(fmt.Sprintf("%s[%s]%s %s | %s\n", color, label, colorReset, issue.CNName, issue.ENName))
	} else {
		buf.WriteString(fmt.Sprintf("[%s] %s | %s\n", label, issue.CNName, issue.ENName))
	}
	buf.WriteString(fmt.Sprintf("  详情: %s\n", issue.Details))
	buf.WriteString(fmt.Sprintf("  位置: %s\n", issue.Location))

	if issue.Remediation != nil && issue.Remediation.Suggestion != "" {
		buf.WriteString(fmt.Sprintf("  建议: %s\n", issue.Remediation.Suggestion))
	}
	buf.WriteString("\n")
}

// filterBySeverity filters issues by severity level
func filterBySeverity(issues []types.Issue, severity types.Severity) []types.Issue {
	var result []types.Issue
	for _, issue := range issues {
		if issue.Severity == severity {
			result = append(result, issue)
		}
	}
	return result
}

// DeduplicateIssues removes duplicate issues based on ENName
func DeduplicateIssues(issues []types.Issue) []types.Issue {
	seen := make(map[string]bool)
	var result []types.Issue

	for _, issue := range issues {
		key := strings.ToUpper(issue.ENName)
		if !seen[key] {
			seen[key] = true
			result = append(result, issue)
		}
	}
	return result
}

// SortIssuesBySeverity sorts issues by severity (critical first)
func SortIssuesBySeverity(issues []types.Issue) []types.Issue {
	critical := filterBySeverity(issues, types.SeverityCritical)
	warning := filterBySeverity(issues, types.SeverityWarning)
	info := filterBySeverity(issues, types.SeverityInfo)

	result := make([]types.Issue, 0, len(issues))
	result = append(result, critical...)
	result = append(result, warning...)
	result = append(result, info...)
	return result
}

// init registers the text reporter
func init() {
	RegisterReporter(NewTextReporter(true))
}
