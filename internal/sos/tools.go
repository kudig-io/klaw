package sos

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"

	"github.com/kudig-io/klaw/internal/diag"
)

// ClusterReader 抽象 SOS 工具所需的集群读操作；*kubernetes.Resources 隐式实现该接口
type ClusterReader interface {
	ListPods(clusterName, namespace string) ([]corev1.Pod, error)
	ListNodes(clusterName string) ([]corev1.Node, error)
	ListEvents(clusterName, namespace string) ([]corev1.Event, error)
	GetPodLogs(clusterName, namespace, podName string, tailLines int64) (string, error)
}

// DiagRunner 诊断流水线执行函数（测试可替换）
type DiagRunner func(ctx context.Context, req diag.DiagnosisRequest) (*diag.DiagnosisResult, error)

// RunDiagnosis 包级变量，便于测试注入
var RunDiagnosis DiagRunner = diag.RunOnlineDiagnostics

const (
	diagTimeout = 30 * time.Second
	maxLogChars = 8000
	defaultTail = int64(100)
	maxTail     = int64(500)
)

// ToolDefinition OpenAI Realtime 兼容的 function 工具定义
type ToolDefinition struct {
	Type        string         `json:"type"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

// ToolExecutor 执行集群查询/诊断工具（三层兜底第 2 层），全部只读或诊断类，无写操作
type ToolExecutor struct {
	reader  ClusterReader
	cluster string
}

func NewToolExecutor(reader ClusterReader, clusterName string) *ToolExecutor {
	return &ToolExecutor{reader: reader, cluster: clusterName}
}

func obj(props map[string]any, required []string) map[string]any {
	return map[string]any{"type": "object", "properties": props, "required": required}
}

func (e *ToolExecutor) Definitions() []ToolDefinition {
	return []ToolDefinition{
		{Type: "function", Name: "get_cluster_status",
			Description: "获取集群健康概览：节点与 Pod 统计、异常计数、最近 Warning 事件摘要",
			Parameters:  obj(map[string]any{}, nil)},
		{Type: "function", Name: "list_pods",
			Description: "列出 Pod。status 可选：abnormal（默认，仅异常 Pod）、running、all；namespace 为空表示全部命名空间",
			Parameters: obj(map[string]any{
				"namespace": map[string]any{"type": "string"},
				"status":    map[string]any{"type": "string", "enum": []string{"abnormal", "running", "all"}},
			}, nil)},
		{Type: "function", Name: "get_pod_logs",
			Description: "获取指定 Pod 的最近日志（默认 100 行，上限 500 行）",
			Parameters: obj(map[string]any{
				"namespace":  map[string]any{"type": "string"},
				"pod":        map[string]any{"type": "string"},
				"tail_lines": map[string]any{"type": "integer"},
			}, []string{"namespace", "pod"})},
		{Type: "function", Name: "list_events",
			Description: "列出最近 Warning/异常事件；namespace 为空表示全部命名空间",
			Parameters: obj(map[string]any{
				"namespace": map[string]any{"type": "string"},
			}, nil)},
		{Type: "function", Name: "run_diagnosis",
			Description: "对集群（可限定 namespace）执行 Klaw 诊断流水线，返回问题列表与修复建议；耗时可能较长（上限 30 秒）",
			Parameters: obj(map[string]any{
				"namespace": map[string]any{"type": "string"},
			}, nil)},
	}
}

// Execute 按工具名分发执行，返回 JSON 字符串结果
func (e *ToolExecutor) Execute(ctx context.Context, name string, args json.RawMessage) (string, error) {
	switch name {
	case "get_cluster_status":
		return e.getClusterStatus()
	case "list_pods":
		return e.listPods(args)
	case "get_pod_logs":
		return e.getPodLogs(args)
	case "list_events":
		return e.listEvents(args)
	case "run_diagnosis":
		return e.runDiagnosis(ctx, args)
	default:
		return "", fmt.Errorf("unknown tool: %s", name)
	}
}

func decodeArgs(args json.RawMessage, v any) error {
	if len(args) == 0 {
		return nil
	}
	return json.Unmarshal(args, v)
}

func marshal(v any) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// isPodAbnormal 判断 Pod 是否处于需要关注的异常状态
func isPodAbnormal(p corev1.Pod) bool {
	switch p.Status.Phase {
	case corev1.PodPending, corev1.PodFailed, corev1.PodUnknown:
		return true
	case corev1.PodSucceeded:
		return false
	}
	for _, cs := range append(p.Status.ContainerStatuses, p.Status.InitContainerStatuses...) {
		if w := cs.State.Waiting; w != nil {
			switch w.Reason {
			case "CrashLoopBackOff", "OOMKilled", "ImagePullBackOff", "ErrImagePull", "CreateContainerConfigError":
				return true
			}
		}
		if t := cs.State.Terminated; t != nil && t.ExitCode != 0 {
			return true
		}
	}
	return false
}

func (e *ToolExecutor) getClusterStatus() (string, error) {
	nodes, err := e.reader.ListNodes(e.cluster)
	if err != nil {
		return "", fmt.Errorf("list nodes: %w", err)
	}
	pods, err := e.reader.ListPods(e.cluster, "")
	if err != nil {
		return "", fmt.Errorf("list pods: %w", err)
	}
	ready, notReady := 0, 0
	for _, n := range nodes {
		if nodeReady(n) {
			ready++
		} else {
			notReady++
		}
	}
	phase := map[string]int{}
	abnormal := 0
	for _, p := range pods {
		phase[string(p.Status.Phase)]++
		if isPodAbnormal(p) {
			abnormal++
		}
	}
	events, _ := e.reader.ListEvents(e.cluster, "")
	return marshal(map[string]any{
		"nodes":           map[string]int{"total": len(nodes), "ready": ready, "not_ready": notReady},
		"pods":            map[string]any{"total": len(pods), "by_phase": phase, "abnormal": abnormal},
		"recent_warnings": summarizeEvents(events, 5),
	})
}

func nodeReady(n corev1.Node) bool {
	for _, c := range n.Status.Conditions {
		if c.Type == corev1.NodeReady {
			return c.Status == corev1.ConditionTrue
		}
	}
	return false
}

func summarizeEvents(events []corev1.Event, limit int) []map[string]any {
	out := []map[string]any{}
	for _, ev := range events {
		if ev.Type != corev1.EventTypeWarning {
			continue
		}
		out = append(out, map[string]any{
			"reason": ev.Reason, "message": truncate(ev.Message, 200),
			"object": ev.InvolvedObject.Namespace + "/" + ev.InvolvedObject.Name,
			"count":  ev.Count,
		})
		if len(out) >= limit {
			break
		}
	}
	return out
}

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "..."
}

func (e *ToolExecutor) listPods(args json.RawMessage) (string, error) {
	var a struct {
		Namespace string `json:"namespace"`
		Status    string `json:"status"`
	}
	if err := decodeArgs(args, &a); err != nil {
		return "", err
	}
	if a.Status == "" {
		a.Status = "abnormal"
	}
	pods, err := e.reader.ListPods(e.cluster, a.Namespace)
	if err != nil {
		return "", fmt.Errorf("list pods: %w", err)
	}
	abnormal := 0
	list := []map[string]any{}
	for _, p := range pods {
		bad := isPodAbnormal(p)
		if bad {
			abnormal++
		}
		switch a.Status {
		case "abnormal":
			if !bad {
				continue
			}
		case "running":
			if p.Status.Phase != corev1.PodRunning {
				continue
			}
		}
		list = append(list, map[string]any{
			"namespace": p.Namespace, "name": p.Name,
			"phase": string(p.Status.Phase), "abnormal": bad,
		})
		if len(list) >= 50 {
			break
		}
	}
	return marshal(map[string]any{"total": len(pods), "abnormal": abnormal, "pods": list})
}

func (e *ToolExecutor) getPodLogs(args json.RawMessage) (string, error) {
	var a struct {
		Namespace string `json:"namespace"`
		Pod       string `json:"pod"`
		TailLines int64  `json:"tail_lines"`
	}
	if err := decodeArgs(args, &a); err != nil {
		return "", err
	}
	if a.Namespace == "" || a.Pod == "" {
		return "", fmt.Errorf("namespace and pod are required")
	}
	if a.TailLines <= 0 {
		a.TailLines = defaultTail
	}
	if a.TailLines > maxTail {
		a.TailLines = maxTail
	}
	logs, err := e.reader.GetPodLogs(e.cluster, a.Namespace, a.Pod, a.TailLines)
	if err != nil {
		return "", fmt.Errorf("get pod logs: %w", err)
	}
	return marshal(map[string]any{"tail_lines": a.TailLines, "logs": truncate(logs, maxLogChars)})
}

func (e *ToolExecutor) listEvents(args json.RawMessage) (string, error) {
	var a struct {
		Namespace string `json:"namespace"`
	}
	if err := decodeArgs(args, &a); err != nil {
		return "", err
	}
	events, err := e.reader.ListEvents(e.cluster, a.Namespace)
	if err != nil {
		return "", fmt.Errorf("list events: %w", err)
	}
	warnings := summarizeEvents(events, 20)
	return marshal(map[string]any{"warnings": warnings, "warning_count": len(warnings)})
}

func (e *ToolExecutor) runDiagnosis(ctx context.Context, args json.RawMessage) (string, error) {
	var a struct {
		Namespace string `json:"namespace"`
	}
	if err := decodeArgs(args, &a); err != nil {
		return "", err
	}
	dctx, cancel := context.WithTimeout(ctx, diagTimeout)
	defer cancel()
	res, err := RunDiagnosis(dctx, diag.DiagnosisRequest{Namespace: a.Namespace, DisableAI: true})
	if err != nil {
		if dctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("diagnosis timed out after %s; suggest checking the diagnostics page", diagTimeout)
		}
		return "", fmt.Errorf("run diagnosis: %w", err)
	}
	issues := []map[string]any{}
	for _, iss := range res.Issues {
		issues = append(issues, map[string]any{
			"severity": iss.Severity.String(), "name": iss.CNName,
			"details": truncate(iss.Details, 300),
		})
		if len(issues) >= 20 {
			break
		}
	}
	return marshal(map[string]any{"issue_count": len(res.Issues), "issues": issues})
}
