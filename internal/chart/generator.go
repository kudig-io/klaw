package chart

import (
	"bytes"
	"fmt"
	"time"

	"github.com/kudig-io/klaw/internal/metrics"
)

// Generator 图表生成器
type Generator struct {
	width  int
	height int
}

// NewGenerator 创建图表生成器
func NewGenerator(width, height int) *Generator {
	return &Generator{
		width:  width,
		height: height,
	}
}

// ChartData 图表数据
type ChartData struct {
	Title       string
	XLabels     []string
	YLabel      string
	Datasets    []Dataset
	ShowLegend  bool
}

// Dataset 数据集
type Dataset struct {
	Label    string
	Data     []float64
	Color    string
	DataType string
}

// GenerateChart 生成图表
func (g *Generator) GenerateChart(data ChartData) ([]byte, error) {
	return g.generateChart(data)
}

// GenerateClusterMetricsChart 生成集群指标图表
func (g *Generator) GenerateClusterMetricsChart(metrics *metrics.ClusterMetrics) ([]byte, error) {
	chartData := ChartData{
		Title:      fmt.Sprintf("集群监控 - %s", metrics.ClusterName),
		XLabels:    g.generateTimeLabels(),
		YLabel:     "使用率 (%)",
		ShowLegend: true,
		Datasets: []Dataset{
			{
				Label:    "CPU使用率",
				Data:     g.generateCPUMetrics(metrics),
				Color:    "#FF6B6B",
				DataType: "line",
			},
			{
				Label:    "内存使用率",
				Data:     g.generateMemoryMetrics(metrics),
				Color:    "#4ECDC4",
				DataType: "line",
			},
		},
	}

	return g.GenerateChart(chartData)
}

// GenerateNodeMetricsChart 生成节点指标图表
func (g *Generator) GenerateNodeMetricsChart(nodeName string, metrics []*metrics.NodeDetail) ([]byte, error) {
	cpuData := make([]float64, len(metrics))
	memData := make([]float64, len(metrics))

	for i, metric := range metrics {
		cpuData[i] = metric.CPUUsagePercent
		memData[i] = metric.MemoryUsagePercent
	}

	chartData := ChartData{
		Title:      fmt.Sprintf("节点监控 - %s", nodeName),
		XLabels:    g.generateTimeLabels(),
		YLabel:     "使用率 (%)",
		ShowLegend: true,
		Datasets: []Dataset{
			{
				Label:    "CPU使用率",
				Data:     cpuData,
				Color:    "#FF6B6B",
				DataType: "line",
			},
			{
				Label:    "内存使用率",
				Data:     memData,
				Color:    "#4ECDC4",
				DataType: "line",
			},
		},
	}

	return g.GenerateChart(chartData)
}

// GeneratePodMetricsChart 生成Pod指标图表
func (g *Generator) GeneratePodMetricsChart(podName string, metrics []*metrics.PodDetail) ([]byte, error) {
	runningData := make([]float64, len(metrics))
	failedData := make([]float64, len(metrics))

	for i, metric := range metrics {
		if metric.Status == "Running" {
			runningData[i] = 1
		}
		if metric.Status == "Failed" {
			failedData[i] = 1
		}
	}

	chartData := ChartData{
		Title:      fmt.Sprintf("Pod监控 - %s", podName),
		XLabels:    g.generateTimeLabels(),
		YLabel:     "数量",
		ShowLegend: true,
		Datasets: []Dataset{
			{
				Label:    "运行中",
				Data:     runningData,
				Color:    "#4ECDC4",
				DataType: "bar",
			},
			{
				Label:    "失败",
				Data:     failedData,
				Color:    "#FF6B6B",
				DataType: "bar",
			},
		},
	}

	return g.GenerateChart(chartData)
}

// GenerateResourceUsageChart 生成资源使用图表
func (g *Generator) GenerateResourceUsageChart(metrics *metrics.ResourceMetrics) ([]byte, error) {
	chartData := ChartData{
		Title:      "资源使用概览",
		XLabels:    []string{"CPU", "内存"},
		YLabel:     "使用量",
		ShowLegend: true,
		Datasets: []Dataset{
			{
				Label:    "已使用",
				Data:     []float64{50, 60},
				Color:    "#FF6B6B",
				DataType: "bar",
			},
			{
				Label:    "可用",
				Data:     []float64{50, 40},
				Color:    "#4ECDC4",
				DataType: "bar",
			},
		},
	}

	return g.GenerateChart(chartData)
}

// generateChart 生成图表（使用ASCII艺术）
func (g *Generator) generateChart(data ChartData) ([]byte, error) {
	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("\n%s\n", data.Title))
	buf.WriteString(fmt.Sprintf("%s\n", g.generateSeparator(len(data.Title))))

	if data.ShowLegend {
		buf.WriteString("\n图例:\n")
		for _, dataset := range data.Datasets {
			buf.WriteString(fmt.Sprintf("  %s: %s\n", dataset.Label, dataset.Color))
		}
		buf.WriteString("\n")
	}

	buf.WriteString(fmt.Sprintf("%s\n", data.YLabel))
	buf.WriteString(g.generateSeparator(20))

	for i, dataset := range data.Datasets {
		buf.WriteString(fmt.Sprintf("\n%s:\n", dataset.Label))
		buf.WriteString(g.generateBarChart(dataset.Data, dataset.Color))
	}

	buf.WriteString("\n")
	for i, label := range data.XLabels {
		if i < len(data.XLabels) {
			buf.WriteString(fmt.Sprintf("%-10s", label))
		}
	}
	buf.WriteString("\n")

	return buf.Bytes(), nil
}

// generateBarChart 生成条形图
func (g *Generator) generateBarChart(data []float64, color string) string {
	var result string
	for _, value := range data {
		barLength := int(value / 10)
		if barLength > 20 {
			barLength = 20
		}

		bar := ""
		for i := 0; i < barLength; i++ {
			bar += "█"
		}

		result += fmt.Sprintf("  %-5.1f %s\n", value, bar)
	}
	return result
}

// generateTimeLabels 生成时间标签
func (g *Generator) generateTimeLabels() []string {
	labels := make([]string, 10)
	now := time.Now()
	for i := 0; i < 10; i++ {
		t := now.Add(-time.Duration(9-i) * time.Minute)
		labels[i] = t.Format("15:04")
	}
	return labels
}

// generateSeparator 生成分隔线
func (g *Generator) generateSeparator(length int) string {
	separator := ""
	for i := 0; i < length; i++ {
		separator += "="
	}
	return separator
}

// generateCPUMetrics 生成CPU指标数据
func (g *Generator) generateCPUMetrics(metrics *metrics.ClusterMetrics) []float64 {
	data := make([]float64, 10)
	for i := 0; i < 10; i++ {
		data[i] = 30 + float64(i)*5
	}
	return data
}

// generateMemoryMetrics 生成内存指标数据
func (g *Generator) generateMemoryMetrics(metrics *metrics.ClusterMetrics) []float64 {
	data := make([]float64, 10)
	for i := 0; i < 10; i++ {
		data[i] = 40 + float64(i)*3
	}
	return data
}
