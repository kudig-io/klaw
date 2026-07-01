package loganalysis

import (
	"context"
	"fmt"
	"io"
	"strconv"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Engine struct {
	clientset kubernetes.Interface
	analyzer  *Analyzer
}

func NewEngine(clientset kubernetes.Interface) *Engine {
	return &Engine{
		clientset: clientset,
		analyzer:  NewAnalyzer(),
	}
}

func (e *Engine) GetPodLogs(ctx context.Context, namespace, name, container string, tailLines int64) (string, error) {
	opts := &corev1.PodLogOptions{
		Container: container,
		TailLines: &tailLines,
	}
	req := e.clientset.CoreV1().Pods(namespace).GetLogs(name, opts)
	stream, err := req.Stream(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to open log stream for %s/%s: %w", namespace, name, err)
	}
	defer stream.Close()

	data, err := io.ReadAll(stream)
	if err != nil {
		return "", fmt.Errorf("failed to read log stream for %s/%s: %w", namespace, name, err)
	}
	return string(data), nil
}

func (e *Engine) AnalyzePodLogs(ctx context.Context, namespace, name, container string) (*LogAnalysis, error) {
	const defaultTail int64 = 1000

	pod, err := e.clientset.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get pod %s/%s: %w", namespace, name, err)
	}
	if container == "" && len(pod.Spec.Containers) > 0 {
		container = pod.Spec.Containers[0].Name
	}

	tail := defaultTail
	if v, ok := pod.Annotations["klaw.io/log-tail-lines"]; ok {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil && n > 0 {
			tail = n
		}
	}

	logs, err := e.GetPodLogs(ctx, namespace, name, container, tail)
	if err != nil {
		return nil, err
	}
	return e.analyzer.AnalyzeLogs(logs), nil
}
