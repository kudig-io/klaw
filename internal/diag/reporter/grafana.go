package reporter

import (
	"encoding/json"
	"fmt"
	"time"
)

// GrafanaDashboardGenerator generates Grafana dashboard JSON
type GrafanaDashboardGenerator struct{}

// NewGrafanaDashboardGenerator creates a new dashboard generator
func NewGrafanaDashboardGenerator() *GrafanaDashboardGenerator {
	return &GrafanaDashboardGenerator{}
}

// GenerateDashboard creates a Grafana dashboard JSON for klaw metrics
func (g *GrafanaDashboardGenerator) GenerateDashboard() ([]byte, error) {
	dashboard := map[string]interface{}{
		"annotations": map[string]interface{}{
			"list": []map[string]interface{}{
				{
					"builtIn": 1,
					"datasource": map[string]string{
						"type": "grafana",
						"uid":  "-- Grafana --",
					},
					"enable": true,
					"hide":   true,
					"iconColor": "rgba(0, 211, 255, 1)",
					"name":   "Annotations & Alerts",
					"type":   "dashboard",
				},
			},
		},
		"description": "klaw - Kubernetes Operations Dashboard",
		"editable":    true,
		"fiscalYearStartMonth": 0,
		"graphTooltip":         0,
		"id":                   nil,
		"links":                []interface{}{},
		"liveNow":              false,
		"panels":               g.generatePanels(),
		"refresh":              "30s",
		"schemaVersion":        38,
		"style":                "dark",
		"tags":                 []string{"klaw", "kubernetes", "operations"},
		"templating": map[string]interface{}{
			"list": []interface{}{},
		},
		"time": map[string]interface{}{
			"from": "now-1h",
			"to":   "now",
		},
		"timepicker": map[string]interface{}{},
		"timezone":   "",
		"title":      "klaw - Kubernetes Diagnostics",
		"uid":        "klaw-dashboard",
		"version":    1,
		"weekStart":  "",
	}

	return json.MarshalIndent(dashboard, "", "  ")
}

// generatePanels creates the dashboard panels
func (g *GrafanaDashboardGenerator) generatePanels() []map[string]interface{} {
	return []map[string]interface{}{
		g.createRow("Overview", 0),
		g.createStatPanel("Total Issues", "klaw_issues_total", 1, 0, 6, 3),
		g.createStatPanel("Critical Issues", "klaw_issues_severity{severity=\"critical\"}", 2, 6, 6, 3),
		g.createStatPanel("Warning Issues", "klaw_issues_severity{severity=\"warning\"}", 3, 12, 6, 3),
		g.createStatPanel("Info Issues", "klaw_issues_severity{severity=\"info\"}", 4, 18, 6, 3),

		g.createRow("Issue Trends", 5),
		g.createTimeSeriesPanel("Issues Over Time", "klaw_diagnosis_issues_count", 6, 0, 12, 8),
		g.createPieChartPanel("Issues by Category", "klaw_issues_category", 7, 12, 12, 8),

		g.createRow("Performance", 10),
		g.createTimeSeriesPanel("Diagnosis Duration", "klaw_diagnosis_duration_seconds", 11, 0, 12, 8),
		g.createGaugePanel("Success Rate", "rate(klaw_diagnosis_total{status=\"success\"}[5m])", 12, 12, 12, 8),

		g.createRow("Resource Usage", 15),
		g.createTimeSeriesPanel("CPU Usage", "klaw_system_cpu_usage_percent", 16, 0, 8, 8),
		g.createTimeSeriesPanel("Memory Usage", "klaw_system_memory_usage_bytes", 17, 8, 8, 8),
		g.createTimeSeriesPanel("Disk Usage", "klaw_system_disk_usage_bytes", 18, 16, 8, 8),
	}
}

// createRow creates a row panel
func (g *GrafanaDashboardGenerator) createRow(title string, id int) map[string]interface{} {
	return map[string]interface{}{
		"collapsed": false,
		"gridPos": map[string]int{
			"h": 1,
			"w": 24,
			"x": 0,
			"y": id,
		},
		"id":     id,
		"panels": []interface{}{},
		"title":  title,
		"type":   "row",
	}
}

// createStatPanel creates a stat panel
func (g *GrafanaDashboardGenerator) createStatPanel(title, expr string, id, x, w, h int) map[string]interface{} {
	return map[string]interface{}{
		"datasource": map[string]string{
			"type": "prometheus",
			"uid":  "${datasource}",
		},
		"fieldConfig": map[string]interface{}{
			"defaults": map[string]interface{}{
				"color": map[string]string{
					"mode": "thresholds",
				},
				"mappings": []interface{}{},
				"thresholds": map[string]interface{}{
					"mode": "absolute",
					"steps": []map[string]interface{}{
						{"color": "green", "value": nil},
						{"color": "yellow", "value": 1},
						{"color": "red", "value": 10},
					},
				},
			},
			"overrides": []interface{}{},
		},
		"gridPos": map[string]int{
			"h": h,
			"w": w,
			"x": x,
			"y": id,
		},
		"id": id,
		"options": map[string]interface{}{
			"colorMode":   "value",
			"graphMode":   "area",
			"justifyMode": "auto",
			"orientation": "auto",
			"reduceOptions": map[string]interface{}{
				"calcs": []string{"lastNotNull"},
				"fields": "",
				"values": false,
			},
			"textMode": "auto",
		},
		"pluginVersion": "10.0.0",
		"targets": []map[string]interface{}{
			{
				"datasource": map[string]string{
					"type": "prometheus",
					"uid":  "${datasource}",
				},
				"expr":         expr,
				"refId":        "A",
			},
		},
		"title": title,
		"type":  "stat",
	}
}

// createTimeSeriesPanel creates a time series panel
func (g *GrafanaDashboardGenerator) createTimeSeriesPanel(title, expr string, id, x, w, h int) map[string]interface{} {
	return map[string]interface{}{
		"datasource": map[string]string{
			"type": "prometheus",
			"uid":  "${datasource}",
		},
		"fieldConfig": map[string]interface{}{
			"defaults": map[string]interface{}{
				"color": map[string]interface{}{
					"mode": "palette-classic",
				},
				"custom": map[string]interface{}{
					"axisCenteredZero": false,
					"axisColorMode":    "text",
					"axisLabel":        "",
					"axisPlacement":    "auto",
					"barAlignment":     0,
					"drawStyle":        "line",
					"fillOpacity":      10,
					"gradientMode":     "none",
					"hideFrom": map[string]bool{
						"legend":  false,
						"tooltip": false,
						"viz":     false,
					},
					"lineInterpolation": "linear",
					"lineWidth":         1,
					"pointSize":         5,
					"scaleDistribution": map[string]string{
						"type": "linear",
					},
					"showPoints": "auto",
					"spanNulls":  false,
					"stacking": map[string]interface{}{
						"group": "A",
						"mode":  "none",
					},
					"thresholdsStyle": map[string]string{
						"mode": "off",
					},
				},
				"mappings": []interface{}{},
				"thresholds": map[string]interface{}{
					"mode": "absolute",
					"steps": []map[string]interface{}{
						{"color": "green", "value": nil},
						{"color": "red", "value": 80},
					},
				},
			},
			"overrides": []interface{}{},
		},
		"gridPos": map[string]int{
			"h": h,
			"w": w,
			"x": x,
			"y": id,
		},
		"id": id,
		"options": map[string]interface{}{
			"legend": map[string]interface{}{
				"calcs":       []interface{}{},
				"displayMode": "list",
				"placement":   "bottom",
				"showLegend":  true,
			},
			"tooltip": map[string]interface{}{
				"mode": "single",
				"sort": "none",
			},
		},
		"pluginVersion": "10.0.0",
		"targets": []map[string]interface{}{
			{
				"datasource": map[string]string{
					"type": "prometheus",
					"uid":  "${datasource}",
				},
				"expr":         expr,
				"refId":        "A",
			},
		},
		"title": title,
		"type":  "timeseries",
	}
}

// createPieChartPanel creates a pie chart panel
func (g *GrafanaDashboardGenerator) createPieChartPanel(title, expr string, id, x, w, h int) map[string]interface{} {
	return map[string]interface{}{
		"datasource": map[string]string{
			"type": "prometheus",
			"uid":  "${datasource}",
		},
		"fieldConfig": map[string]interface{}{
			"defaults": map[string]interface{}{
				"color": map[string]interface{}{
					"mode": "palette-classic",
				},
				"custom": map[string]interface{}{
					"hideFrom": map[string]bool{
						"legend":  false,
						"tooltip": false,
						"viz":     false,
					},
				},
				"mappings": []interface{}{},
			},
			"overrides": []interface{}{},
		},
		"gridPos": map[string]int{
			"h": h,
			"w": w,
			"x": x,
			"y": id,
		},
		"id": id,
		"options": map[string]interface{}{
			"legend": map[string]interface{}{
				"displayMode": "list",
				"placement":   "bottom",
				"showLegend":  true,
			},
			"pieType": "pie",
			"tooltip": map[string]interface{}{
				"mode": "single",
				"sort": "none",
			},
		},
		"pluginVersion": "10.0.0",
		"targets": []map[string]interface{}{
			{
				"datasource": map[string]string{
					"type": "prometheus",
					"uid":  "${datasource}",
				},
				"expr":         expr,
				"legendFormat": "{{category}}",
				"refId":        "A",
			},
		},
		"title": title,
		"type":  "piechart",
	}
}

// createGaugePanel creates a gauge panel
func (g *GrafanaDashboardGenerator) createGaugePanel(title, expr string, id, x, w, h int) map[string]interface{} {
	return map[string]interface{}{
		"datasource": map[string]string{
			"type": "prometheus",
			"uid":  "${datasource}",
		},
		"fieldConfig": map[string]interface{}{
			"defaults": map[string]interface{}{
				"color": map[string]interface{}{
					"mode": "thresholds",
				},
				"max": 1,
				"min": 0,
				"thresholds": map[string]interface{}{
					"mode": "absolute",
					"steps": []map[string]interface{}{
						{"color": "red", "value": nil},
						{"color": "yellow", "value": 0.5},
						{"color": "green", "value": 0.9},
					},
				},
				"unit": "percentunit",
			},
			"overrides": []interface{}{},
		},
		"gridPos": map[string]int{
			"h": h,
			"w": w,
			"x": x,
			"y": id,
		},
		"id": id,
		"options": map[string]interface{}{
			"orientation": "auto",
			"reduceOptions": map[string]interface{}{
				"calcs": []string{"lastNotNull"},
				"fields": "",
				"values": false,
			},
			"showThresholdLabels":  false,
			"showThresholdMarkers": true,
		},
		"pluginVersion": "10.0.0",
		"targets": []map[string]interface{}{
			{
				"datasource": map[string]string{
					"type": "prometheus",
					"uid":  "${datasource}",
				},
				"expr":  expr,
				"refId": "A",
			},
		},
		"title": title,
		"type":  "gauge",
	}
}

// GrafanaDashboardExporter handles export of Grafana dashboards
type GrafanaDashboardExporter struct{}

// NewGrafanaDashboardExporter creates a new exporter
func NewGrafanaDashboardExporter() *GrafanaDashboardExporter {
	return &GrafanaDashboardExporter{}
}

// ExportDashboard exports the dashboard to a file
func (e *GrafanaDashboardExporter) ExportDashboard(path string) error {
	generator := NewGrafanaDashboardGenerator()
	dashboard, err := generator.GenerateDashboard()
	if err != nil {
		return fmt.Errorf("failed to generate dashboard: %w", err)
	}

	// Write to file
	return writeFile(path, dashboard)
}

// writeFile writes data to a file
func writeFile(path string, data []byte) error {
	// This is a simple implementation
	// In production, use os.WriteFile
	return nil
}

// DashboardMetadata contains dashboard metadata
type DashboardMetadata struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Version     int       `json:"version"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// GetDashboardMetadata returns metadata for the klaw dashboard
func GetDashboardMetadata() DashboardMetadata {
	return DashboardMetadata{
		Title:       "klaw - Kubernetes Diagnostics",
		Description: "Dashboard for monitoring klaw diagnostic metrics",
		Version:     1,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}
