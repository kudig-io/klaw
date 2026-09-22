// Package kubernetes provides Kubernetes component analyzers
package kubernetes

import (
	"context"
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kudig-io/klaw/internal/diag/analyzer"
	"github.com/kudig-io/klaw/internal/diag/types"
)

// EventAnalyzer checks for warning events (online mode only)
type EventAnalyzer struct {
	*analyzer.BaseAnalyzer
}

// NewEventAnalyzer creates a new event analyzer
func NewEventAnalyzer() *EventAnalyzer {
	return &EventAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer(
			"kubernetes.events",
			"检查集群事件",
			"kubernetes",
			[]types.DataMode{types.ModeOnline},
		),
	}
}

// Analyze performs event analysis
func (a *EventAnalyzer) Analyze(ctx context.Context, data *types.DiagnosticData) ([]types.Issue, error) {
	var issues []types.Issue

	if !data.HasK8sClient() {
		return issues, nil
	}

	events, err := data.K8sClient.CoreV1().Events("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	warningEvents := make(map[string]int) // reason -> count

	for _, event := range events.Items {
		if event.Type == "Warning" {
			warningEvents[event.Reason]++
		}
	}

	// Check for common warning patterns
	criticalReasons := map[string]string{
		"FailedScheduling":   "Pod调度失败，检查资源和节点选择器",
		"FailedMount":        "存储卷挂载失败，检查PV/PVC状态",
		"FailedAttachVolume": "存储卷附加失败，检查存储后端",
		"Unhealthy":          "容器健康检查失败",
		"BackOff":            "容器启动失败或频繁重启",
		"NodeNotReady":       "节点不可用",
		"NetworkNotReady":    "网络未就绪",
	}

	for reason, desc := range criticalReasons {
		if count, ok := warningEvents[reason]; ok && count > 3 {
			severity := types.SeverityWarning
			if reason == "NodeNotReady" || reason == "NetworkNotReady" {
				severity = types.SeverityCritical
			}

			issue := types.NewIssue(
				severity,
				fmt.Sprintf("集群事件: %s", reason),
				fmt.Sprintf("EVENT_%s", strings.ToUpper(reason)),
				fmt.Sprintf("检测到 %d 个 %s 事件", count, reason),
				"k8s/events",
			).WithRemediation(desc)
			issue.AnalyzerName = a.Name()
			issues = append(issues, *issue)
		}
	}

	return issues, nil
}
