package loganalysis

import (
	"regexp"
	"strings"
	"time"
)

// Analyzer performs static analysis on raw log text.
type Analyzer struct{}

// NewAnalyzer returns a new Analyzer instance.
func NewAnalyzer() *Analyzer {
	return &Analyzer{}
}

// AnalyzeLogs parses the supplied log text and returns a structured analysis.
func (a *Analyzer) AnalyzeLogs(logs string) *LogAnalysis {
	analysis := &LogAnalysis{
		LogLevels:    make(map[string]int),
		PatternStats: make(map[string]int),
	}

	lines := strings.Split(logs, "\n")
	analysis.TotalLines = len(lines)

	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		lineNum := i + 1
		lower := strings.ToLower(line)
		level := classifyLine(lower)
		analysis.LogLevels[level]++

		entry := LogEntry{
			Line:    lineNum,
			Content: line,
			Level:   level,
		}

		if ts, ok := extractTimestamp(line); ok {
			entry.Timestamp = ts
		}

		switch level {
		case "error":
			analysis.ErrorCount++
			analysis.Errors = append(analysis.Errors, entry)
		case "warning":
			analysis.WarningCount++
			analysis.Warnings = append(analysis.Warnings, entry)
		case "info":
			analysis.InfoCount++
		case "debug":
			analysis.DebugCount++
		}

		if isStackTrace(line) {
			analysis.StackTraces = append(analysis.StackTraces, line)
		}

		if evt, ok := detectSecurityEvent(lower, line); ok {
			analysis.SecurityEvents = append(analysis.SecurityEvents, evt)
		}

		countPatterns(lower, analysis.PatternStats)
	}

	analysis.PerformanceMetrics = extractPerformance(logs)
	return analysis
}

func classifyLine(lower string) string {
	switch {
	case strings.Contains(lower, "error") ||
		strings.Contains(lower, "fatal") ||
		strings.Contains(lower, "panic"):
		return "error"
	case strings.Contains(lower, "warn"):
		return "warning"
	case strings.Contains(lower, "debug"):
		return "debug"
	default:
		return "info"
	}
}

func isStackTrace(line string) bool {
	patterns := []string{
		"goroutine",
		"panic:",
		"Traceback (most recent call last)",
		"\tat ",
	}
	for _, p := range patterns {
		if strings.Contains(line, p) {
			return true
		}
	}
	return false
}

func detectSecurityEvent(lower, original string) (SecurityEvent, bool) {
	keywords := []struct {
		keyword  string
		evtType  string
		severity string
	}{
		{"unauthorized", "Unauthorized Access", "high"},
		{"forbidden", "Forbidden Access", "high"},
		{"auth failed", "Authentication Failure", "medium"},
		{"authentication failure", "Authentication Failure", "medium"},
		{"access denied", "Access Denied", "high"},
		{"invalid token", "Invalid Token", "medium"},
		{"permission denied", "Permission Denied", "medium"},
	}
	for _, kw := range keywords {
		if strings.Contains(lower, kw.keyword) {
			return SecurityEvent{
				Type:     kw.evtType,
				Message:  original,
				Severity: kw.severity,
			}, true
		}
	}
	return SecurityEvent{}, false
}

func countPatterns(lower string, stats map[string]int) {
	patterns := []string{"timeout", "retry", "oom", "connection refused", "rate limit"}
	for _, p := range patterns {
		if strings.Contains(lower, p) {
			stats[p]++
		}
	}
}

func extractPerformance(logs string) PerformanceData {
	var pd PerformanceData
	re := regexp.MustCompile(`(?i)(https?://\S+|/\S+)\s+.*?(\d+(?:\.\d+)?(?:ms|s))`)
	matches := re.FindAllStringSubmatch(logs, -1)
	for _, m := range matches {
		if len(m) >= 3 {
			pd.SlowRequests = append(pd.SlowRequests, SlowRequest{
				URL:          m[1],
				ResponseTime: m[2],
			})
		}
	}
	return pd
}

func extractTimestamp(line string) (time.Time, bool) {
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05.000Z07:00",
		"2006-01-02 15:04:05",
	}
	for _, layout := range formats {
		if len(line) >= len(layout) {
			if ts, err := time.Parse(layout, line[:len(layout)]); err == nil {
				return ts, true
			}
		}
	}
	return time.Time{}, false
}
