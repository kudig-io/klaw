package kubernetes

import (
	"context"
	"testing"

	"github.com/kudig-io/klaw/internal/diag/types"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/client-go/kubernetes/fake"
)

func TestGPUNodeAnalyzer_isGPUNode(t *testing.T) {
	analyzer := NewGPUNodeAnalyzer()

	tests := []struct {
		name     string
		node     *corev1.Node
		expected bool
	}{
		{
			name: "node with nvidia.com/gpu.present label",
			node: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "gpu-node-1",
					Labels: map[string]string{"nvidia.com/gpu.present": "true"},
				},
			},
			expected: true,
		},
		{
			name: "node with GPU capacity",
			node: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "gpu-node-2",
				},
				Status: corev1.NodeStatus{
					Capacity: corev1.ResourceList{
						NvidiaGPUResource: resource.MustParse("4"),
					},
				},
			},
			expected: true,
		},
		{
			name: "regular node without GPU",
			node: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "regular-node",
					Labels: map[string]string{},
				},
				Status: corev1.NodeStatus{
					Capacity: corev1.ResourceList{},
				},
			},
			expected: false,
		},
		{
			name: "node with AMD GPU",
			node: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "amd-gpu-node",
				},
				Status: corev1.NodeStatus{
					Capacity: corev1.ResourceList{
						AMDGPUResource: resource.MustParse("2"),
					},
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.isGPUNode(tt.node)
			if result != tt.expected {
				t.Errorf("isGPUNode() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGPUNodeAnalyzer_Analyze_NoPlugin(t *testing.T) {
	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "gpu-node",
			Labels: map[string]string{"nvidia.com/gpu.present": "true"},
		},
		Status: corev1.NodeStatus{
			Capacity: corev1.ResourceList{
				NvidiaGPUResource: resource.MustParse("4"),
			},
			Allocatable: corev1.ResourceList{
				NvidiaGPUResource: resource.MustParse("4"),
			},
		},
	}

	client := fake.NewSimpleClientset(node)
	data := &types.DiagnosticData{K8sClient: client}

	analyzer := NewGPUNodeAnalyzer()
	issues, err := analyzer.Analyze(context.Background(), data)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	hasNoPlugin := false
	for _, issue := range issues {
		if issue.ENName == "GPU_PLUGIN_NOT_RUNNING" {
			hasNoPlugin = true
			if issue.Severity != types.SeverityWarning {
				t.Errorf("expected warning severity, got %v", issue.Severity)
			}
		}
	}

	if !hasNoPlugin {
		t.Errorf("expected GPU_PLUGIN_NOT_RUNNING issue, got %v", issues)
	}
}

func TestGPUNodeAnalyzer_Analyze_NotAllocatable(t *testing.T) {
	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "gpu-node",
			Labels: map[string]string{"nvidia.com/gpu.present": "true"},
		},
		Status: corev1.NodeStatus{
			Capacity: corev1.ResourceList{
				NvidiaGPUResource: resource.MustParse("4"),
			},
			Allocatable: corev1.ResourceList{
				NvidiaGPUResource: resource.MustParse("0"),
			},
		},
	}

	client := fake.NewSimpleClientset(node)
	data := &types.DiagnosticData{K8sClient: client}

	analyzer := NewGPUNodeAnalyzer()
	issues, err := analyzer.Analyze(context.Background(), data)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	hasNotAllocatable := false
	for _, issue := range issues {
		if issue.ENName == "GPU_NOT_ALLOCATABLE" {
			hasNotAllocatable = true
			if issue.Severity != types.SeverityCritical {
				t.Errorf("expected critical severity, got %v", issue.Severity)
			}
		}
	}

	if !hasNotAllocatable {
		t.Errorf("expected GPU_NOT_ALLOCATABLE issue, got %v", issues)
	}
}

func TestGPUNodeAnalyzer_Analyze_WithPluginRunning(t *testing.T) {
	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "gpu-node",
			Labels: map[string]string{"nvidia.com/gpu.present": "true"},
		},
		Status: corev1.NodeStatus{
			Capacity: corev1.ResourceList{
				NvidiaGPUResource: resource.MustParse("4"),
			},
			Allocatable: corev1.ResourceList{
				NvidiaGPUResource: resource.MustParse("4"),
			},
			Conditions: []corev1.NodeCondition{
				{
					Type:   corev1.NodeReady,
					Status: corev1.ConditionTrue,
				},
			},
		},
	}

	pluginPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nvidia-device-plugin-daemonset-abc",
			Namespace: "kube-system",
		},
		Spec: corev1.PodSpec{
			NodeName: "gpu-node",
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
		},
	}

	client := fake.NewSimpleClientset(node, pluginPod)
	data := &types.DiagnosticData{K8sClient: client}

	analyzer := NewGPUNodeAnalyzer()
	issues, err := analyzer.Analyze(context.Background(), data)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// 不应该有 GPU_PLUGIN_NOT_RUNNING 问题
	for _, issue := range issues {
		if issue.ENName == "GPU_PLUGIN_NOT_RUNNING" {
			t.Errorf("expected no GPU_PLUGIN_NOT_RUNNING issue with running plugin, got %v", issues)
		}
	}
}

func TestGPUPodAnalyzer_Analyze_ResourceMismatch(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "gpu-pod",
			Namespace: "default",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name: "gpu-container",
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							NvidiaGPUResource: resource.MustParse("1"),
						},
						Limits: corev1.ResourceList{
							NvidiaGPUResource: resource.MustParse("2"), // Mismatch!
						},
					},
				},
			},
		},
	}

	client := fake.NewSimpleClientset(pod)
	data := &types.DiagnosticData{K8sClient: client}

	analyzer := NewGPUPodAnalyzer()
	issues, err := analyzer.Analyze(context.Background(), data)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	hasMismatch := false
	for _, issue := range issues {
		if issue.ENName == "GPU_RESOURCE_MISMATCH" {
			hasMismatch = true
			if issue.Severity != types.SeverityWarning {
				t.Errorf("expected warning severity, got %v", issue.Severity)
			}
		}
	}

	if !hasMismatch {
		t.Errorf("expected GPU_RESOURCE_MISMATCH issue, got %v", issues)
	}
}

func TestGPUPodAnalyzer_Analyze_FractionalGPU(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "gpu-pod",
			Namespace: "default",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name: "gpu-container",
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							NvidiaGPUResource: resource.MustParse("500m"), // Fractional
						},
						Limits: corev1.ResourceList{
							NvidiaGPUResource: resource.MustParse("500m"),
						},
					},
				},
			},
		},
	}

	client := fake.NewSimpleClientset(pod)
	data := &types.DiagnosticData{K8sClient: client}

	analyzer := NewGPUPodAnalyzer()
	issues, err := analyzer.Analyze(context.Background(), data)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	hasFractional := false
	for _, issue := range issues {
		if issue.ENName == "GPU_FRACTION_REQUESTED" {
			hasFractional = true
			if issue.Severity != types.SeverityInfo {
				t.Errorf("expected info severity, got %v", issue.Severity)
			}
		}
	}

	if !hasFractional {
		t.Errorf("expected GPU_FRACTION_REQUESTED issue, got %v", issues)
	}
}

func TestGPUPodAnalyzer_Analyze_SchedulingFailed(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "gpu-pod",
			Namespace: "default",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name: "gpu-container",
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							NvidiaGPUResource: resource.MustParse("8"),
						},
					},
				},
			},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodPending,
			Conditions: []corev1.PodCondition{
				{
					Type:    corev1.PodScheduled,
					Status:  corev1.ConditionFalse,
					Message: "0/3 nodes are available: 3 Insufficient nvidia.com/gpu",
				},
			},
		},
	}

	client := fake.NewSimpleClientset(pod)
	data := &types.DiagnosticData{K8sClient: client}

	analyzer := NewGPUPodAnalyzer()
	issues, err := analyzer.Analyze(context.Background(), data)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	hasSchedulingFailed := false
	for _, issue := range issues {
		if issue.ENName == "GPU_POD_SCHEDULING_FAILED" {
			hasSchedulingFailed = true
			if issue.Severity != types.SeverityWarning {
				t.Errorf("expected warning severity, got %v", issue.Severity)
			}
		}
	}

	if !hasSchedulingFailed {
		t.Errorf("expected GPU_POD_SCHEDULING_FAILED issue, got %v", issues)
	}
}

func TestGPUPodAnalyzer_hasGPURequest(t *testing.T) {
	analyzer := NewGPUPodAnalyzer()

	tests := []struct {
		name     string
		pod      *corev1.Pod
		expected bool
	}{
		{
			name: "pod with NVIDIA GPU request",
			pod: &corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									NvidiaGPUResource: resource.MustParse("1"),
								},
							},
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "pod with AMD GPU request",
			pod: &corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									AMDGPUResource: resource.MustParse("1"),
								},
							},
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "pod without GPU request",
			pod: &corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU: resource.MustParse("100m"),
								},
							},
						},
					},
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.hasGPURequest(tt.pod)
			if result != tt.expected {
				t.Errorf("hasGPURequest() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestNPUAnalyzer_Analyze_NoPlugin(t *testing.T) {
	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "npu-node",
			Labels: map[string]string{"huawei.com/ascend910": "true"},
		},
		Status: corev1.NodeStatus{
			Capacity: corev1.ResourceList{
				HUAWEINPUResource: resource.MustParse("8"),
			},
		},
	}

	client := fake.NewSimpleClientset(node)
	data := &types.DiagnosticData{K8sClient: client}

	analyzer := NewNPUAnalyzer()
	issues, err := analyzer.Analyze(context.Background(), data)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	hasNoPlugin := false
	for _, issue := range issues {
		if issue.ENName == "NPU_PLUGIN_NOT_RUNNING" {
			hasNoPlugin = true
			if issue.Severity != types.SeverityWarning {
				t.Errorf("expected warning severity, got %v", issue.Severity)
			}
		}
	}

	if !hasNoPlugin {
		t.Errorf("expected NPU_PLUGIN_NOT_RUNNING issue, got %v", issues)
	}
}

func TestNPUAnalyzer_Analyze_WithPlugin(t *testing.T) {
	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "npu-node",
			Labels: map[string]string{"huawei.com/ascend910": "true"},
		},
		Status: corev1.NodeStatus{
			Capacity: corev1.ResourceList{
				HUAWEINPUResource: resource.MustParse("8"),
			},
		},
	}

	pluginPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ascend-device-plugin-daemonset-abc",
			Namespace: "kube-system",
		},
		Spec: corev1.PodSpec{
			NodeName: "npu-node",
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
		},
	}

	client := fake.NewSimpleClientset(node, pluginPod)
	data := &types.DiagnosticData{K8sClient: client}

	analyzer := NewNPUAnalyzer()
	issues, err := analyzer.Analyze(context.Background(), data)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// 不应该有 NPU_PLUGIN_NOT_RUNNING 问题
	for _, issue := range issues {
		if issue.ENName == "NPU_PLUGIN_NOT_RUNNING" {
			t.Errorf("expected no NPU_PLUGIN_NOT_RUNNING issue with running plugin, got %v", issues)
		}
	}
}

func TestGPUAnalyzer_NoK8sClient(t *testing.T) {
	data := &types.DiagnosticData{K8sClient: nil}

	analyzer := NewGPUNodeAnalyzer()
	issues, err := analyzer.Analyze(context.Background(), data)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("expected no issues without k8s client, got %d", len(issues))
	}
}
