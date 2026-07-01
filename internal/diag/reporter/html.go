// Package reporter provides HTML report generation with chart visualization
package reporter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"time"

	"github.com/kudig-io/klaw/internal/diag/types"
)

// HTMLReporter generates HTML reports with embedded charts
type HTMLReporter struct{}

// NewHTMLReporter creates a new HTML reporter
func NewHTMLReporter() *HTMLReporter {
	return &HTMLReporter{}
}

// Format returns the output format name
func (r *HTMLReporter) Format() string {
	return "html"
}

// Generate creates an HTML report from the issues
func (r *HTMLReporter) Generate(issues []types.Issue, metadata *ReportMetadata) ([]byte, error) {
	// Prepare data for template
	data := struct {
		Metadata       *ReportMetadata
		Issues         []types.Issue
		CriticalIssues []types.Issue
		WarningIssues  []types.Issue
		InfoIssues     []types.Issue
		SeverityData   template.JS
		CategoryData   template.JS
		Timestamp      string
	}{
		Metadata:       metadata,
		Issues:         issues,
		CriticalIssues: filterIssuesBySeverity(issues, types.SeverityCritical),
		WarningIssues:  filterIssuesBySeverity(issues, types.SeverityWarning),
		InfoIssues:     filterIssuesBySeverity(issues, types.SeverityInfo),
		Timestamp:      time.Now().Format("2006-01-02 15:04:05"),
	}

	// Generate chart data
	severityCounts := r.calculateSeverityCounts(issues)
	severityJSON, _ := json.Marshal(severityCounts)
	data.SeverityData = template.JS(severityJSON)

	categoryCounts := r.calculateCategoryCounts(issues)
	categoryJSON, _ := json.Marshal(categoryCounts)
	data.CategoryData = template.JS(categoryJSON)

	// Parse and execute template
	tmpl, err := template.New("html-report").Parse(htmlTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("failed to execute HTML template: %w", err)
	}

	return buf.Bytes(), nil
}

func filterIssuesBySeverity(issues []types.Issue, severity types.Severity) []types.Issue {
	var result []types.Issue
	for _, issue := range issues {
		if issue.Severity == severity {
			result = append(result, issue)
		}
	}
	return result
}

func (r *HTMLReporter) calculateSeverityCounts(issues []types.Issue) map[string]int {
	counts := map[string]int{
		"Critical": 0,
		"Warning":  0,
		"Info":     0,
	}
	for _, issue := range issues {
		switch issue.Severity {
		case types.SeverityCritical:
			counts["Critical"]++
		case types.SeverityWarning:
			counts["Warning"]++
		case types.SeverityInfo:
			counts["Info"]++
		}
	}
	return counts
}

func (r *HTMLReporter) calculateCategoryCounts(issues []types.Issue) map[string]int {
	counts := make(map[string]int)
	for _, issue := range issues {
		category := issueCategory(issue)
		counts[category]++
	}
	return counts
}

func issueCategory(issue types.Issue) string {
	if issue.AnalyzerName != "" {
		// Extract category from analyzer name (e.g., "system.cpu" -> "system")
		for _, cat := range []string{"system", "network", "process", "kernel", "kubernetes", "runtime", "log"} {
			if len(issue.AnalyzerName) > len(cat) && issue.AnalyzerName[:len(cat)] == cat {
				return cat
			}
		}
	}
	return "other"
}

// htmlTemplate is the HTML report template with embedded Chart.js
const htmlTemplate = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Kudig Diagnostic Report</title>
    <script src="https://cdn.jsdelivr.net/npm/chart.js@4.4.1/dist/chart.umd.min.js"></script>
    <style>
        :root {
            --bg-color: #f5f5f5;
            --card-bg: #ffffff;
            --text-color: #333333;
            --text-secondary: #666666;
            --border-color: #e0e0e0;
            --critical-color: #dc3545;
            --warning-color: #ffc107;
            --info-color: #17a2b8;
            --success-color: #28a745;
        }
        
        @media (prefers-color-scheme: dark) {
            :root {
                --bg-color: #1a1a2e;
                --card-bg: #16213e;
                --text-color: #eaeaea;
                --text-secondary: #a0a0a0;
                --border-color: #0f3460;
            }
        }
        
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
            background-color: var(--bg-color);
            color: var(--text-color);
            line-height: 1.6;
            padding: 20px;
        }
        
        .container {
            max-width: 1200px;
            margin: 0 auto;
        }
        
        header {
            background-color: var(--card-bg);
            padding: 30px;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
            margin-bottom: 20px;
        }
        
        h1 {
            font-size: 28px;
            margin-bottom: 10px;
        }
        
        .meta-info {
            color: var(--text-secondary);
            font-size: 14px;
        }
        
        .meta-info span {
            margin-right: 20px;
        }
        
        .summary-cards {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 15px;
            margin-bottom: 20px;
        }
        
        .summary-card {
            background-color: var(--card-bg);
            padding: 20px;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
            text-align: center;
        }
        
        .summary-card.critical {
            border-top: 4px solid var(--critical-color);
        }
        
        .summary-card.warning {
            border-top: 4px solid var(--warning-color);
        }
        
        .summary-card.info {
            border-top: 4px solid var(--info-color);
        }
        
        .summary-card.total {
            border-top: 4px solid var(--success-color);
        }
        
        .summary-card h3 {
            font-size: 14px;
            color: var(--text-secondary);
            margin-bottom: 10px;
        }
        
        .summary-card .count {
            font-size: 36px;
            font-weight: bold;
        }
        
        .summary-card.critical .count {
            color: var(--critical-color);
        }
        
        .summary-card.warning .count {
            color: var(--warning-color);
        }
        
        .summary-card.info .count {
            color: var(--info-color);
        }
        
        .charts-section {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(400px, 1fr));
            gap: 20px;
            margin-bottom: 20px;
        }
        
        .chart-card {
            background-color: var(--card-bg);
            padding: 20px;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        
        .chart-card h3 {
            margin-bottom: 15px;
            font-size: 16px;
        }
        
        .issues-section {
            background-color: var(--card-bg);
            padding: 20px;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
            margin-bottom: 20px;
        }
        
        .issues-section h2 {
            margin-bottom: 15px;
            padding-bottom: 10px;
            border-bottom: 2px solid var(--border-color);
        }
        
        .issue-item {
            padding: 15px;
            margin-bottom: 10px;
            border-radius: 6px;
            border-left: 4px solid;
        }
        
        .issue-item.critical {
            background-color: rgba(220, 53, 69, 0.1);
            border-left-color: var(--critical-color);
        }
        
        .issue-item.warning {
            background-color: rgba(255, 193, 7, 0.1);
            border-left-color: var(--warning-color);
        }
        
        .issue-item.info {
            background-color: rgba(23, 162, 184, 0.1);
            border-left-color: var(--info-color);
        }
        
        .issue-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 8px;
        }
        
        .issue-title {
            font-weight: bold;
            font-size: 16px;
        }
        
        .issue-severity {
            padding: 2px 10px;
            border-radius: 12px;
            font-size: 12px;
            font-weight: bold;
            text-transform: uppercase;
        }
        
        .issue-severity.critical {
            background-color: var(--critical-color);
            color: white;
        }
        
        .issue-severity.warning {
            background-color: var(--warning-color);
            color: #333;
        }
        
        .issue-severity.info {
            background-color: var(--info-color);
            color: white;
        }
        
        .issue-details {
            color: var(--text-secondary);
            margin-bottom: 8px;
        }
        
        .issue-meta {
            font-size: 12px;
            color: var(--text-secondary);
        }
        
        .issue-remediation {
            margin-top: 10px;
            padding: 10px;
            background-color: rgba(40, 167, 69, 0.1);
            border-radius: 4px;
            font-size: 14px;
        }
        
        .issue-remediation strong {
            color: var(--success-color);
        }
        
        .empty-state {
            text-align: center;
            padding: 40px;
            color: var(--text-secondary);
        }
        
        .empty-state .icon {
            font-size: 48px;
            margin-bottom: 15px;
        }
        
        footer {
            text-align: center;
            padding: 20px;
            color: var(--text-secondary);
            font-size: 12px;
        }
    </style>
</head>
<body>
    <div class="container">
        <header>
            <h1>🔍 Kudig Diagnostic Report</h1>
            <div class="meta-info">
                <span>📅 {{.Timestamp}}</span>
                {{if .Metadata.Hostname}}<span>🖥️ Host: {{.Metadata.Hostname}}</span>{{end}}
                {{if .Metadata.Mode}}<span>📊 Mode: {{.Metadata.Mode}}</span>{{end}}
                {{if .Metadata.DiagnosePath}}<span>📁 Path: {{.Metadata.DiagnosePath}}</span>{{end}}
            </div>
        </header>
        
        <div class="summary-cards">
            <div class="summary-card critical">
                <h3>Critical Issues</h3>
                <div class="count">{{.Metadata.Summary.Critical}}</div>
            </div>
            <div class="summary-card warning">
                <h3>Warning Issues</h3>
                <div class="count">{{.Metadata.Summary.Warning}}</div>
            </div>
            <div class="summary-card info">
                <h3>Info Issues</h3>
                <div class="count">{{.Metadata.Summary.Info}}</div>
            </div>
            <div class="summary-card total">
                <h3>Total Issues</h3>
                <div class="count">{{.Metadata.Summary.Total}}</div>
            </div>
        </div>
        
        <div class="charts-section">
            <div class="chart-card">
                <h3>Severity Distribution</h3>
                <canvas id="severityChart"></canvas>
            </div>
            <div class="chart-card">
                <h3>Issues by Category</h3>
                <canvas id="categoryChart"></canvas>
            </div>
        </div>
        
        {{if .CriticalIssues}}
        <div class="issues-section">
            <h2>🚨 Critical Issues</h2>
            {{range .CriticalIssues}}
            <div class="issue-item critical">
                <div class="issue-header">
                    <span class="issue-title">{{.CNName}}</span>
                    <span class="issue-severity critical">Critical</span>
                </div>
                <div class="issue-details">{{.Details}}</div>
                <div class="issue-meta">Code: {{.ENName}} | Location: {{.Location}}</div>
                {{if .Remediation}}
                <div class="issue-remediation">
                    <strong>💡 Suggestion:</strong> {{.Remediation.Suggestion}}
                </div>
                {{end}}
            </div>
            {{end}}
        </div>
        {{end}}
        
        {{if .WarningIssues}}
        <div class="issues-section">
            <h2>⚠️ Warning Issues</h2>
            {{range .WarningIssues}}
            <div class="issue-item warning">
                <div class="issue-header">
                    <span class="issue-title">{{.CNName}}</span>
                    <span class="issue-severity warning">Warning</span>
                </div>
                <div class="issue-details">{{.Details}}</div>
                <div class="issue-meta">Code: {{.ENName}} | Location: {{.Location}}</div>
                {{if .Remediation}}
                <div class="issue-remediation">
                    <strong>💡 Suggestion:</strong> {{.Remediation.Suggestion}}
                </div>
                {{end}}
            </div>
            {{end}}
        </div>
        {{end}}
        
        {{if .InfoIssues}}
        <div class="issues-section">
            <h2>ℹ️ Info Issues</h2>
            {{range .InfoIssues}}
            <div class="issue-item info">
                <div class="issue-header">
                    <span class="issue-title">{{.CNName}}</span>
                    <span class="issue-severity info">Info</span>
                </div>
                <div class="issue-details">{{.Details}}</div>
                <div class="issue-meta">Code: {{.ENName}} | Location: {{.Location}}</div>
                {{if .Remediation}}
                <div class="issue-remediation">
                    <strong>💡 Suggestion:</strong> {{.Remediation.Suggestion}}
                </div>
                {{end}}
            </div>
            {{end}}
        </div>
        {{end}}
        
        {{if eq .Metadata.Summary.Total 0}}
        <div class="issues-section">
            <div class="empty-state">
                <div class="icon">✅</div>
                <h3>No Issues Found</h3>
                <p>Your Kubernetes cluster appears to be healthy!</p>
            </div>
        </div>
        {{end}}
        
        <footer>
            Generated by Kudig v2.0 | Kubernetes Diagnostic Toolkit
        </footer>
    </div>
    
    <script>
        // Severity Chart
        const severityCtx = document.getElementById('severityChart').getContext('2d');
        const severityData = {{.SeverityData}};
        new Chart(severityCtx, {
            type: 'doughnut',
            data: {
                labels: Object.keys(severityData),
                datasets: [{
                    data: Object.values(severityData),
                    backgroundColor: [
                        '#dc3545', // Critical - red
                        '#ffc107', // Warning - yellow
                        '#17a2b8'  // Info - blue
                    ],
                    borderWidth: 0
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: true,
                plugins: {
                    legend: {
                        position: 'bottom'
                    }
                }
            }
        });
        
        // Category Chart
        const categoryCtx = document.getElementById('categoryChart').getContext('2d');
        const categoryData = {{.CategoryData}};
        new Chart(categoryCtx, {
            type: 'bar',
            data: {
                labels: Object.keys(categoryData),
                datasets: [{
                    label: 'Issues',
                    data: Object.values(categoryData),
                    backgroundColor: '#0d6efd',
                    borderRadius: 4
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: true,
                plugins: {
                    legend: {
                        display: false
                    }
                },
                scales: {
                    y: {
                        beginAtZero: true,
                        ticks: {
                            stepSize: 1
                        }
                    }
                }
            }
        });
    </script>
</body>
</html>
`

// init registers the HTML reporter
func init() {
	RegisterReporter(NewHTMLReporter())
}
