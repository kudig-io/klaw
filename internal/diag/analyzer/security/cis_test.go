package security

import (
	"context"
	"testing"

	"github.com/kudig-io/klaw/internal/diag/types"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestCISAPIServerAnalyzer_Analyze_AnonymousAuth(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kube-apiserver-node1",
			Namespace: "kube-system",
			Labels:    map[string]string{"component": "kube-apiserver"},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "kube-apiserver",
					Command: []string{"kube-apiserver", "--anonymous-auth=true"},
				},
			},
		},
	}

	client := fake.NewSimpleClientset(pod)
	data := &types.DiagnosticData{K8sClient: client}

	analyzer := NewCISAPIServerAnalyzer()
	issues, err := analyzer.Analyze(context.Background(), data)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	hasAnonymousAuth := false
	for _, issue := range issues {
		if issue.ENName == "CIS_APISERVER_ANONYMOUS_AUTH" {
			hasAnonymousAuth = true
			if issue.Severity != types.SeverityCritical {
				t.Errorf("expected critical severity, got %v", issue.Severity)
			}
		}
	}

	if !hasAnonymousAuth {
		t.Errorf("expected CIS_APISERVER_ANONYMOUS_AUTH issue, got %v", issues)
	}
}

func TestCISAPIServerAnalyzer_Analyze_InsecurePort(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kube-apiserver-node1",
			Namespace: "kube-system",
			Labels:    map[string]string{"component": "kube-apiserver"},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "kube-apiserver",
					Command: []string{"kube-apiserver", "--insecure-port=8080"},
				},
			},
		},
	}

	client := fake.NewSimpleClientset(pod)
	data := &types.DiagnosticData{K8sClient: client}

	analyzer := NewCISAPIServerAnalyzer()
	issues, err := analyzer.Analyze(context.Background(), data)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	hasInsecurePort := false
	for _, issue := range issues {
		if issue.ENName == "CIS_APISERVER_INSECURE_PORT" {
			hasInsecurePort = true
		}
	}

	if !hasInsecurePort {
		t.Errorf("expected CIS_APISERVER_INSECURE_PORT issue, got %v", issues)
	}
}

func TestCISEtcdAnalyzer_Analyze_NoCertAuth(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "etcd-node1",
			Namespace: "kube-system",
			Labels:    map[string]string{"component": "etcd"},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "etcd",
					Command: []string{"etcd", "--listen-client-urls=https://0.0.0.0:2379"},
				},
			},
		},
	}

	client := fake.NewSimpleClientset(pod)
	data := &types.DiagnosticData{K8sClient: client}

	analyzer := NewCISEtcdAnalyzer()
	issues, err := analyzer.Analyze(context.Background(), data)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	hasNoCertAuth := false
	for _, issue := range issues {
		if issue.ENName == "CIS_ETCD_NO_CERT_AUTH" {
			hasNoCertAuth = true
			if issue.Severity != types.SeverityCritical {
				t.Errorf("expected critical severity, got %v", issue.Severity)
			}
		}
	}

	if !hasNoCertAuth {
		t.Errorf("expected CIS_ETCD_NO_CERT_AUTH issue, got %v", issues)
	}
}

func TestCISEtcdAnalyzer_Analyze_AutoTLS(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "etcd-node1",
			Namespace: "kube-system",
			Labels:    map[string]string{"component": "etcd"},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "etcd",
					Command: []string{"etcd", "--auto-tls=true", "--client-cert-auth=true"},
				},
			},
		},
	}

	client := fake.NewSimpleClientset(pod)
	data := &types.DiagnosticData{K8sClient: client}

	analyzer := NewCISEtcdAnalyzer()
	issues, err := analyzer.Analyze(context.Background(), data)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	hasAutoTLS := false
	for _, issue := range issues {
		if issue.ENName == "CIS_ETCD_AUTO_TLS" {
			hasAutoTLS = true
			if issue.Severity != types.SeverityWarning {
				t.Errorf("expected warning severity, got %v", issue.Severity)
			}
		}
	}

	if !hasAutoTLS {
		t.Errorf("expected CIS_ETCD_AUTO_TLS issue, got %v", issues)
	}
}

func TestCISPodSecurityAnalyzer_Analyze_Privileged(t *testing.T) {
	privileged := true
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "privileged-pod",
			Namespace: "default",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name: "privileged-container",
					SecurityContext: &corev1.SecurityContext{
						Privileged: &privileged,
					},
				},
			},
		},
	}

	client := fake.NewSimpleClientset(pod)
	data := &types.DiagnosticData{K8sClient: client}

	analyzer := NewCISPodSecurityAnalyzer()
	issues, err := analyzer.Analyze(context.Background(), data)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	hasPrivileged := false
	for _, issue := range issues {
		if issue.ENName == "CIS_PRIVILEGED_CONTAINERS" {
			hasPrivileged = true
		}
	}

	if !hasPrivileged {
		t.Errorf("expected CIS_PRIVILEGED_CONTAINERS issue, got %v", issues)
	}
}

func TestCISPodSecurityAnalyzer_Analyze_HostPID(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hostpid-pod",
			Namespace: "default",
		},
		Spec: corev1.PodSpec{
			HostPID: true,
			Containers: []corev1.Container{
				{
					Name: "container1",
				},
			},
		},
	}

	client := fake.NewSimpleClientset(pod)
	data := &types.DiagnosticData{K8sClient: client}

	analyzer := NewCISPodSecurityAnalyzer()
	issues, err := analyzer.Analyze(context.Background(), data)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	hasHostPID := false
	for _, issue := range issues {
		if issue.ENName == "CIS_HOST_PID" {
			hasHostPID = true
		}
	}

	if !hasHostPID {
		t.Errorf("expected CIS_HOST_PID issue, got %v", issues)
	}
}

func TestCISPodSecurityAnalyzer_Analyze_RunAsRoot(t *testing.T) {
	runAsRoot := int64(0)
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "root-pod",
			Namespace: "default",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name: "root-container",
					SecurityContext: &corev1.SecurityContext{
						RunAsUser: &runAsRoot,
					},
				},
			},
		},
	}

	client := fake.NewSimpleClientset(pod)
	data := &types.DiagnosticData{K8sClient: client}

	analyzer := NewCISPodSecurityAnalyzer()
	issues, err := analyzer.Analyze(context.Background(), data)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	hasRunAsRoot := false
	for _, issue := range issues {
		if issue.ENName == "CIS_RUN_AS_ROOT" {
			hasRunAsRoot = true
		}
	}

	if !hasRunAsRoot {
		t.Errorf("expected CIS_RUN_AS_ROOT issue, got %v", issues)
	}
}

func TestCISSecretAnalyzer_Analyze_SecretsInDefault(t *testing.T) {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-secret",
			Namespace: "default",
		},
	}

	client := fake.NewSimpleClientset(secret)
	data := &types.DiagnosticData{K8sClient: client}

	analyzer := NewCISSecretAnalyzer()
	issues, err := analyzer.Analyze(context.Background(), data)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	hasSecretInDefault := false
	for _, issue := range issues {
		if issue.ENName == "CIS_SECRET_IN_DEFAULT" {
			hasSecretInDefault = true
		}
	}

	if !hasSecretInDefault {
		t.Errorf("expected CIS_SECRET_IN_DEFAULT issue, got %v", issues)
	}
}

func TestCISAnalyzer_NoK8sClient(t *testing.T) {
	data := &types.DiagnosticData{K8sClient: nil}

	analyzer := NewCISAPIServerAnalyzer()
	issues, err := analyzer.Analyze(context.Background(), data)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("expected no issues without k8s client, got %d", len(issues))
	}
}
