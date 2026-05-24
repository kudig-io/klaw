package loganalysis

import "time"

// LogAnalysis holds the results of analyzing a block of log text.
type LogAnalysis struct {
	TotalLines         int             `json:"totalLines"`
	ErrorCount         int             `json:"errorCount"`
	WarningCount       int             `json:"warningCount"`
	InfoCount          int             `json:"infoCount"`
	DebugCount         int             `json:"debugCount"`
	Errors             []LogEntry      `json:"errors,omitempty"`
	Warnings           []LogEntry      `json:"warnings,omitempty"`
	StackTraces        []string        `json:"stackTraces,omitempty"`
	PerformanceMetrics PerformanceData `json:"performanceMetrics"`
	SecurityEvents     []SecurityEvent `json:"securityEvents,omitempty"`
	LogLevels          map[string]int  `json:"logLevels"`
	PatternStats       map[string]int  `json:"patternStats"`
}

// LogEntry represents a single classified log line.
type LogEntry struct {
	Line      int       `json:"line"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
}

// PerformanceData holds performance-related metrics extracted from logs.
type PerformanceData struct {
	SlowRequests []SlowRequest `json:"slowRequests,omitempty"`
}

// SlowRequest represents a request that exceeded a latency threshold.
type SlowRequest struct {
	URL          string `json:"url"`
	ResponseTime string `json:"responseTime"`
}

// SecurityEvent represents a security-relevant log entry.
type SecurityEvent struct {
	Type     string `json:"type"`
	Message  string `json:"message"`
	Severity string `json:"severity"`
}
