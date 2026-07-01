package rules

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/kudig-io/klaw/internal/diag/types"
)

// Engine evaluates rules against diagnostic data
type Engine struct {
	loader *Loader
}

// NewEngine creates a new rule engine
func NewEngine(loader *Loader) *Engine {
	return &Engine{
		loader: loader,
	}
}

// Evaluate runs all rules against the diagnostic data
func (e *Engine) Evaluate(ctx context.Context, data *types.DiagnosticData) ([]types.Issue, error) {
	var issues []types.Issue

	rules := e.loader.GetAllRules()
	for _, rule := range rules {
		select {
		case <-ctx.Done():
			return issues, ctx.Err()
		default:
		}

		matched, details, location, err := e.evaluateRule(rule, data)
		if err != nil {
			// Log error but continue
			continue
		}

		if matched {
			issue := rule.ToIssue(details, location)
			issues = append(issues, *issue)
		}
	}

	return issues, nil
}

// EvaluateByCategory runs rules of a specific category
func (e *Engine) EvaluateByCategory(
	ctx context.Context,
	category string,
	data *types.DiagnosticData,
) ([]types.Issue, error) {
	var issues []types.Issue

	rules := e.loader.GetRulesByCategory(category)
	for _, rule := range rules {
		select {
		case <-ctx.Done():
			return issues, ctx.Err()
		default:
		}

		matched, details, location, err := e.evaluateRule(rule, data)
		if err != nil {
			continue
		}

		if matched {
			issue := rule.ToIssue(details, location)
			issues = append(issues, *issue)
		}
	}

	return issues, nil
}

// evaluateRule evaluates a single rule
func (e *Engine) evaluateRule(
	rule Rule,
	data *types.DiagnosticData,
) (matched bool, details string, location string, err error) {
	return e.evaluateCondition(rule.Condition, data)
}

// evaluateCondition evaluates a condition recursively
func (e *Engine) evaluateCondition(
	cond Condition,
	data *types.DiagnosticData,
) (matched bool, details string, location string, err error) {
	switch cond.Type {
	case "file_contains":
		return e.evalFileContains(cond, data)
	case "regex_match":
		return e.evalRegexMatch(cond, data)
	case "metric_threshold":
		return e.evalMetricThreshold(cond, data)
	case "and":
		return e.evalAnd(cond, data)
	case "or":
		return e.evalOr(cond, data)
	default:
		return false, "", "", fmt.Errorf("unknown condition type: %s", cond.Type)
	}
}

// evalFileContains checks if a file contains a pattern
func (e *Engine) evalFileContains(cond Condition, data *types.DiagnosticData) (bool, string, string, error) {
	content, ok := data.GetRawFile(cond.File)
	if !ok {
		return false, "", cond.File, nil
	}

	// Try regex match
	re, err := regexp.Compile("(?i)" + cond.Pattern)
	if err != nil {
		// Fall back to simple contains
		if strings.Contains(string(content), cond.Pattern) {
			count := strings.Count(string(content), cond.Pattern)
			if cond.Count > 0 && count < cond.Count {
				return false, "", cond.File, nil
			}
			result := !cond.Negate
			details := fmt.Sprintf("在 %s 中发现模式 '%s' 出现 %d 次", cond.File, cond.Pattern, count)
			return result, details, cond.File, nil
		}
		return cond.Negate, "", cond.File, nil
	}

	matches := re.FindAllString(string(content), -1)
	count := len(matches)

	if count > 0 {
		if cond.Count > 0 && count < cond.Count {
			return cond.Negate, "", cond.File, nil
		}
		result := !cond.Negate
		details := fmt.Sprintf("在 %s 中发现模式 '%s' 匹配 %d 次", cond.File, cond.Pattern, count)
		return result, details, cond.File, nil
	}

	return cond.Negate, "", cond.File, nil
}

// evalRegexMatch performs regex matching
func (e *Engine) evalRegexMatch(cond Condition, data *types.DiagnosticData) (bool, string, string, error) {
	content, ok := data.GetRawFile(cond.File)
	if !ok {
		return false, "", cond.File, nil
	}

	re, err := regexp.Compile(cond.Pattern)
	if err != nil {
		return false, "", cond.File, err
	}

	match := re.FindString(string(content))
	if match != "" {
		result := !cond.Negate
		details := fmt.Sprintf("正则表达式 '%s' 匹配到: %s", cond.Pattern, truncateString(match, 100))
		return result, details, cond.File, nil
	}

	return cond.Negate, "", cond.File, nil
}

// evalMetricThreshold checks if a metric exceeds threshold
func (e *Engine) evalMetricThreshold(cond Condition, data *types.DiagnosticData) (bool, string, string, error) {
	if data.SystemMetrics == nil {
		return false, "", "system_metrics", nil
	}

	var value float64
	var found bool

	switch cond.Metric {
	case "load_avg_1min":
		value = data.SystemMetrics.LoadAvg1Min
		// Normalize by CPU cores
		if data.SystemMetrics.CPUCores > 0 {
			value = value / float64(data.SystemMetrics.CPUCores)
		}
		found = true
	case "load_avg_5min":
		value = data.SystemMetrics.LoadAvg5Min
		if data.SystemMetrics.CPUCores > 0 {
			value = value / float64(data.SystemMetrics.CPUCores)
		}
		found = true
	case "load_avg_15min":
		value = data.SystemMetrics.LoadAvg15Min
		if data.SystemMetrics.CPUCores > 0 {
			value = value / float64(data.SystemMetrics.CPUCores)
		}
		found = true
	case "mem_used_percent":
		if data.SystemMetrics.MemTotal > 0 {
			used := data.SystemMetrics.MemTotal - data.SystemMetrics.MemAvailable
			value = float64(used) / float64(data.SystemMetrics.MemTotal) * 100
			found = true
		}
	case "mem_available_percent":
		if data.SystemMetrics.MemTotal > 0 {
			value = float64(data.SystemMetrics.MemAvailable) / float64(data.SystemMetrics.MemTotal) * 100
			found = true
		}
	case "swap_used_percent":
		if data.SystemMetrics.SwapTotal > 0 {
			used := data.SystemMetrics.SwapTotal - data.SystemMetrics.SwapFree
			value = float64(used) / float64(data.SystemMetrics.SwapTotal) * 100
			found = true
		}
	case "disk_used_percent":
		// Check all disk mounts, use max
		for _, disk := range data.SystemMetrics.DiskUsage {
			if disk.UsedPercent > value {
				value = disk.UsedPercent
			}
			found = true
		}
	case "conntrack_percent":
		if data.SystemMetrics.ConntrackMax > 0 {
			value = float64(data.SystemMetrics.ConntrackCurrent) / float64(data.SystemMetrics.ConntrackMax) * 100
			found = true
		}
	}

	if !found {
		return false, "", "system_metrics", nil
	}

	var matched bool
	switch cond.Operator {
	case "gt", ">":
		matched = value > cond.Threshold
	case "gte", ">=":
		matched = value >= cond.Threshold
	case "lt", "<":
		matched = value < cond.Threshold
	case "lte", "<=":
		matched = value <= cond.Threshold
	case "eq", "==":
		matched = value == cond.Threshold
	case "ne", "!=":
		matched = value != cond.Threshold
	default:
		return false, "", "system_metrics", fmt.Errorf("unknown operator: %s", cond.Operator)
	}

	if cond.Negate {
		matched = !matched
	}

	details := fmt.Sprintf("%s = %.2f (threshold: %s %.2f)", cond.Metric, value, cond.Operator, cond.Threshold)
	return matched, details, "system_metrics", nil
}

// evalAnd evaluates AND condition
func (e *Engine) evalAnd(cond Condition, data *types.DiagnosticData) (bool, string, string, error) {
	var details []string
	var location string

	for _, c := range cond.And {
		matched, d, loc, err := e.evaluateCondition(c, data)
		if err != nil {
			return false, "", "", err
		}
		if !matched {
			return false, "", "", nil
		}
		if d != "" {
			details = append(details, d)
		}
		if loc != "" {
			location = loc
		}
	}

	return true, strings.Join(details, "; "), location, nil
}

// evalOr evaluates OR condition
func (e *Engine) evalOr(cond Condition, data *types.DiagnosticData) (bool, string, string, error) {
	for _, c := range cond.Or {
		matched, details, location, err := e.evaluateCondition(c, data)
		if err != nil {
			continue
		}
		if matched {
			return true, details, location, nil
		}
	}

	return false, "", "", nil
}

// truncateString truncates a string to max length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
