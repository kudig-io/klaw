package sos

import (
	"context"
	"encoding/json"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kudig-io/klaw/internal/diag"
	"github.com/kudig-io/klaw/internal/diag/types"
)

type fakeReader struct {
	pods   []corev1.Pod
	nodes  []corev1.Node
	events []corev1.Event
	logs   string
}

func (f *fakeReader) ListPods(_, _ string) ([]corev1.Pod, error) { return f.pods, nil }
func (f *fakeReader) ListNodes(string) ([]corev1.Node, error)    { return f.nodes, nil }
func (f *fakeReader) ListEvents(_, _ string) ([]corev1.Event, error) {
	return f.events, nil
}
func (f *fakeReader) GetPodLogs(_, _, _ string, _ int64) (string, error) {
	return f.logs, nil
}

func abnormalPod(name string, reason string) corev1.Pod {
	return corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "default"},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			ContainerStatuses: []corev1.ContainerStatus{{
				State: corev1.ContainerState{Waiting: &corev1.ContainerStateWaiting{Reason: reason}},
			}},
		},
	}
}

func TestToolDefinitions(t *testing.T) {
	e := NewToolExecutor(&fakeReader{}, "")
	defs := e.Definitions()
	want := []string{"get_cluster_status", "list_pods", "get_pod_logs", "list_events", "run_diagnosis"}
	if len(defs) != len(want) {
		t.Fatalf("expected %d tools, got %d", len(want), len(defs))
	}
	for i, d := range defs {
		if d.Name != want[i] || d.Type != "function" {
			t.Fatalf("def[%d] = %+v, want %s", i, d, want[i])
		}
	}
}

func TestExecuteListPodsAbnormalOnly(t *testing.T) {
	fr := &fakeReader{pods: []corev1.Pod{
		abnormalPod("bad", "CrashLoopBackOff"),
		{ObjectMeta: metav1.ObjectMeta{Name: "ok", Namespace: "default"}, Status: corev1.PodStatus{Phase: corev1.PodRunning}},
	}}
	out, err := NewToolExecutor(fr, "").Execute(context.Background(), "list_pods", json.RawMessage(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	var got struct {
		Total    int `json:"total"`
		Abnormal int `json:"abnormal"`
		Pods     []struct {
			Name string `json:"name"`
		} `json:"pods"`
	}
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatal(err)
	}
	if got.Abnormal != 1 || len(got.Pods) != 1 || got.Pods[0].Name != "bad" {
		t.Fatalf("unexpected result: %s", out)
	}
}

func TestExecuteGetPodLogsClamp(t *testing.T) {
	fr := &fakeReader{logs: "x"}
	out, err := NewToolExecutor(fr, "").Execute(context.Background(), "get_pod_logs",
		json.RawMessage(`{"namespace":"default","pod":"p","tail_lines":9999}`))
	if err != nil {
		t.Fatal(err)
	}
	var got struct {
		TailLines int64 `json:"tail_lines"`
	}
	_ = json.Unmarshal([]byte(out), &got)
	if got.TailLines != maxTail {
		t.Fatalf("expected tail clamped to %d, got %d", maxTail, got.TailLines)
	}
}

func TestExecuteUnknownTool(t *testing.T) {
	if _, err := NewToolExecutor(&fakeReader{}, "").Execute(context.Background(), "nope", json.RawMessage(`{}`)); err == nil {
		t.Fatal("expected error for unknown tool")
	}
}

func TestExecuteRunDiagnosis(t *testing.T) {
	orig := RunDiagnosis
	defer func() { RunDiagnosis = orig }()
	RunDiagnosis = func(ctx context.Context, req diag.DiagnosisRequest) (*diag.DiagnosisResult, error) {
		if !req.DisableAI {
			t.Fatal("expected DisableAI=true to avoid extra LLM cost")
		}
		return &diag.DiagnosisResult{Issues: []types.Issue{{Severity: types.SeverityWarning, CNName: "测试问题", Details: "细节"}}}, nil
	}
	out, err := NewToolExecutor(&fakeReader{}, "").Execute(context.Background(), "run_diagnosis", json.RawMessage(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	var got struct {
		IssueCount int `json:"issue_count"`
		Issues     []struct {
			Name string `json:"name"`
		} `json:"issues"`
	}
	_ = json.Unmarshal([]byte(out), &got)
	if got.IssueCount != 1 || got.Issues[0].Name != "测试问题" {
		t.Fatalf("unexpected: %s", out)
	}
}
