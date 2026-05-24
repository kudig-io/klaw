package api

import (
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNormalizeResourceKind(t *testing.T) {
	tests := map[string]string{
		"pod":         "pods",
		"pods":        "pods",
		"deployment":  "deployments",
		"service":     "services",
		"namespace":   "namespaces",
		"node":        "nodes",
		"event":       "events",
		"custom-kind": "custom-kind",
	}

	for input, want := range tests {
		if got := normalizeResourceKind(input); got != want {
			t.Fatalf("normalizeResourceKind(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestToUnifiedPod(t *testing.T) {
	now := time.Now()
	pod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "demo-pod",
			Namespace:         "default",
			CreationTimestamp: metav1.NewTime(now),
			Labels:            map[string]string{"app": "demo"},
		},
		Status: corev1.PodStatus{Phase: corev1.PodRunning},
	}

	info := toUnifiedPod(pod)
	if info.Kind != "Pod" {
		t.Fatalf("Kind = %q, want Pod", info.Kind)
	}
	if info.Status != "Running" {
		t.Fatalf("Status = %q, want Running", info.Status)
	}
	if info.Namespace != "default" {
		t.Fatalf("Namespace = %q, want default", info.Namespace)
	}
}

func TestToUnifiedDeployment(t *testing.T) {
	replicas := int32(3)
	deployment := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo-deployment",
			Namespace: "default",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
		},
		Status: appsv1.DeploymentStatus{
			ReadyReplicas:     2,
			AvailableReplicas: 2,
		},
	}

	info := toUnifiedDeployment(deployment)
	if info.Kind != "Deployment" {
		t.Fatalf("Kind = %q, want Deployment", info.Kind)
	}
	if info.Status == "" {
		t.Fatal("Status should not be empty")
	}
}
