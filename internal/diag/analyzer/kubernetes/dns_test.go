package kubernetes

import (
	"context"
	"testing"

	"github.com/kudig-io/klaw/internal/diag/types"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestCoreDNSAnalyzer_Analyze_NoCoreDNS(t *testing.T) {
	client := fake.NewSimpleClientset()
	data := &types.DiagnosticData{K8sClient: client}

	analyzer := NewCoreDNSAnalyzer()
	issues, err := analyzer.Analyze(context.Background(), data)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(issues) != 1 {
		t.Errorf("expected 1 issue for missing CoreDNS, got %d", len(issues))
	}
	if len(issues) > 0 && issues[0].ENName != "COREDNS_NOT_FOUND" {
		t.Errorf("expected COREDNS_NOT_FOUND, got %s", issues[0].ENName)
	}
}

func TestCoreDNSAnalyzer_Analyze_RunningPods(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "coredns-abc123",
			Namespace: "kube-system",
			Labels:    map[string]string{"k8s-app": "kube-dns"},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			ContainerStatuses: []corev1.ContainerStatus{
				{
					Name:  "coredns",
					Ready: true,
				},
			},
		},
	}

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kube-dns",
			Namespace: "kube-system",
		},
		Spec: corev1.ServiceSpec{
			ClusterIP: "10.96.0.10",
			Selector:  map[string]string{"k8s-app": "kube-dns"},
		},
	}

	eps := &corev1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kube-dns",
			Namespace: "kube-system",
		},
		Subsets: []corev1.EndpointSubset{
			{
				Addresses: []corev1.EndpointAddress{
					{IP: "10.244.0.2"},
				},
			},
		},
	}

	client := fake.NewSimpleClientset(pod, svc, eps)
	data := &types.DiagnosticData{K8sClient: client}

	analyzer := NewCoreDNSAnalyzer()
	issues, err := analyzer.Analyze(context.Background(), data)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("expected no issues for healthy CoreDNS, got %d: %v", len(issues), issues)
	}
}

func TestCoreDNSAnalyzer_Analyze_CrashLoop(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "coredns-abc123",
			Namespace: "kube-system",
			Labels:    map[string]string{"k8s-app": "kube-dns"},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodPending,
			ContainerStatuses: []corev1.ContainerStatus{
				{
					Name:  "coredns",
					Ready: false,
					State: corev1.ContainerState{
						Waiting: &corev1.ContainerStateWaiting{
							Reason:  "CrashLoopBackOff",
							Message: "Back-off restarting failed container",
						},
					},
				},
			},
		},
	}

	client := fake.NewSimpleClientset(pod)
	data := &types.DiagnosticData{K8sClient: client}

	analyzer := NewCoreDNSAnalyzer()
	issues, err := analyzer.Analyze(context.Background(), data)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	hasCrashLoop := false
	for _, issue := range issues {
		if issue.ENName == "COREDNS_CRASH_LOOP" {
			hasCrashLoop = true
			if issue.Severity != types.SeverityCritical {
				t.Errorf("expected critical severity, got %v", issue.Severity)
			}
		}
	}

	if !hasCrashLoop {
		t.Errorf("expected COREDNS_CRASH_LOOP issue, got %v", issues)
	}
}

func TestCoreDNSAnalyzer_Analyze_HighRestarts(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "coredns-abc123",
			Namespace: "kube-system",
			Labels:    map[string]string{"k8s-app": "kube-dns"},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			ContainerStatuses: []corev1.ContainerStatus{
				{
					Name:         "coredns",
					Ready:        true,
					RestartCount: 10,
				},
			},
		},
	}

	client := fake.NewSimpleClientset(pod)
	data := &types.DiagnosticData{K8sClient: client}

	analyzer := NewCoreDNSAnalyzer()
	issues, err := analyzer.Analyze(context.Background(), data)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	hasHighRestarts := false
	for _, issue := range issues {
		if issue.ENName == "COREDNS_HIGH_RESTARTS" {
			hasHighRestarts = true
			if issue.Severity != types.SeverityWarning {
				t.Errorf("expected warning severity, got %v", issue.Severity)
			}
		}
	}

	if !hasHighRestarts {
		t.Errorf("expected COREDNS_HIGH_RESTARTS issue, got %v", issues)
	}
}

func TestCoreDNSAnalyzer_Analyze_NoEndpoints(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "coredns-abc123",
			Namespace: "kube-system",
			Labels:    map[string]string{"k8s-app": "kube-dns"},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			ContainerStatuses: []corev1.ContainerStatus{
				{
					Name:  "coredns",
					Ready: true,
				},
			},
		},
	}

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kube-dns",
			Namespace: "kube-system",
		},
		Spec: corev1.ServiceSpec{
			ClusterIP: "10.96.0.10",
			Selector:  map[string]string{"k8s-app": "kube-dns"},
		},
	}

	eps := &corev1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kube-dns",
			Namespace: "kube-system",
		},
		Subsets: []corev1.EndpointSubset{}, // Empty subsets
	}

	client := fake.NewSimpleClientset(pod, svc, eps)
	data := &types.DiagnosticData{K8sClient: client}

	analyzer := NewCoreDNSAnalyzer()
	issues, err := analyzer.Analyze(context.Background(), data)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	hasNoEndpoints := false
	for _, issue := range issues {
		if issue.ENName == "COREDNS_ENDPOINTS_EMPTY" {
			hasNoEndpoints = true
			if issue.Severity != types.SeverityCritical {
				t.Errorf("expected critical severity, got %v", issue.Severity)
			}
		}
	}

	if !hasNoEndpoints {
		t.Errorf("expected COREDNS_ENDPOINTS_EMPTY issue, got %v", issues)
	}
}

func TestDNSPodConfigAnalyzer_Analyze_DefaultDNS(t *testing.T) {
	// Create pods with Default DNS policy
	pods := &corev1.PodList{
		Items: []corev1.Pod{
			{
				ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "default"},
				Spec:       corev1.PodSpec{DNSPolicy: corev1.DNSDefault},
			},
			{
				ObjectMeta: metav1.ObjectMeta{Name: "pod2", Namespace: "default"},
				Spec:       corev1.PodSpec{DNSPolicy: corev1.DNSDefault},
			},
			{
				ObjectMeta: metav1.ObjectMeta{Name: "pod3", Namespace: "default"},
				Spec:       corev1.PodSpec{DNSPolicy: corev1.DNSClusterFirst},
			},
		},
	}

	client := fake.NewSimpleClientset(pods)
	data := &types.DiagnosticData{K8sClient: client}

	analyzer := NewDNSPodConfigAnalyzer()
	issues, err := analyzer.Analyze(context.Background(), data)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	hasDefaultDNSWarning := false
	for _, issue := range issues {
		if issue.ENName == "POD_DNS_DEFAULT_HIGH" {
			hasDefaultDNSWarning = true
			if issue.Severity != types.SeverityWarning {
				t.Errorf("expected warning severity, got %v", issue.Severity)
			}
		}
	}

	if !hasDefaultDNSWarning {
		t.Errorf("expected POD_DNS_DEFAULT_HIGH issue, got %v", issues)
	}
}

func TestCoreDNSAnalyzer_NoK8sClient(t *testing.T) {
	data := &types.DiagnosticData{K8sClient: nil}

	analyzer := NewCoreDNSAnalyzer()
	issues, err := analyzer.Analyze(context.Background(), data)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("expected no issues without k8s client, got %d", len(issues))
	}
}
