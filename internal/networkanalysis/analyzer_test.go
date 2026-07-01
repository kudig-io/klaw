package networkanalysis

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestAnalyzeNetwork(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&corev1.Service{
			ObjectMeta: metav1.ObjectMeta{Name: "svc-lb", Namespace: "default"},
			Spec: corev1.ServiceSpec{
				Type: corev1.ServiceTypeLoadBalancer,
				Ports: []corev1.ServicePort{{Port: 80}},
			},
		},
		&corev1.Service{
			ObjectMeta: metav1.ObjectMeta{Name: "svc-internal", Namespace: "default"},
			Spec:       corev1.ServiceSpec{Type: corev1.ServiceTypeClusterIP},
		},
		&networkingv1.NetworkPolicy{
			ObjectMeta: metav1.ObjectMeta{Name: "deny-all", Namespace: "prod"},
		},
		&networkingv1.Ingress{
			ObjectMeta: metav1.ObjectMeta{Name: "ing-1", Namespace: "default"},
			Spec: networkingv1.IngressSpec{
				Rules: []networkingv1.IngressRule{{Host: "example.com"}},
			},
		},
	)

	result, err := NewAnalyzer(clientset).AnalyzeNetwork(context.Background())
	if err != nil {
		t.Fatalf("AnalyzeNetwork: %v", err)
	}
	if result.TotalServices != 2 {
		t.Errorf("TotalServices = %d, want 2", result.TotalServices)
	}
	if len(result.ExposedServices) != 1 {
		t.Errorf("ExposedServices = %d, want 1 (LoadBalancer only)", len(result.ExposedServices))
	}
	if result.ExposedServices[0].Name != "svc-lb" {
		t.Errorf("ExposedServices[0].Name = %q, want svc-lb", result.ExposedServices[0].Name)
	}
	if len(result.IngressesByHost["example.com"]) != 1 {
		t.Error("expected ingress indexed by host example.com")
	}
}
