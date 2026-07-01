package servicemesh

import (
	"context"
	"testing"

	"github.com/kudig-io/klaw/internal/diag/types"
)

func TestIstioAnalyzer_Name(t *testing.T) {
	a := NewIstioAnalyzer()
	if a.Name() != "servicemesh.istio" {
		t.Errorf("Expected name 'servicemesh.istio', got '%s'", a.Name())
	}
}

func TestIstioAnalyzer_Category(t *testing.T) {
	a := NewIstioAnalyzer()
	if a.Category() != "servicemesh" {
		t.Errorf("Expected category 'servicemesh', got '%s'", a.Category())
	}
}

func TestIstioAnalyzer_NoIstio(t *testing.T) {
	a := NewIstioAnalyzer()
	data := &types.DiagnosticData{}
	issues, err := a.Analyze(context.Background(), data)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues when Istio not installed, got %d", len(issues))
	}
}

func TestIstioAnalyzer_IstiodMissing(t *testing.T) {
	a := NewIstioAnalyzer()
	data := &types.DiagnosticData{
		IstioInfo: &types.IstioInfo{
			IstiodPods: 0,
		},
	}
	issues, err := a.Analyze(context.Background(), data)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	found := false
	for _, issue := range issues {
		if issue.ENName == "ISTIO_ISTIOD_MISSING" {
			found = true
			if issue.Severity != types.SeverityCritical {
				t.Errorf("Expected critical severity, got %v", issue.Severity)
			}
		}
	}
	if !found {
		t.Error("Expected ISTIO_ISTIOD_MISSING issue")
	}
}

func TestIstioAnalyzer_IstiodNotReady(t *testing.T) {
	a := NewIstioAnalyzer()
	data := &types.DiagnosticData{
		IstioInfo: &types.IstioInfo{
			IstiodPods:  3,
			IstiodReady: 1,
		},
	}
	issues, err := a.Analyze(context.Background(), data)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	found := false
	for _, issue := range issues {
		if issue.ENName == "ISTIO_ISTIOD_NOT_READY" {
			found = true
		}
	}
	if !found {
		t.Error("Expected ISTIO_ISTIOD_NOT_READY issue")
	}
}

func TestIstioAnalyzer_ProxyNotSynced(t *testing.T) {
	a := NewIstioAnalyzer()
	data := &types.DiagnosticData{
		IstioInfo: &types.IstioInfo{
			IstiodPods:  1,
			IstiodReady: 1,
			IngressPods: 1,
			IngressReady: 1,
			ProxyStatus: []types.IstioProxyStatus{
				{PodName: "app-1", Namespace: "default", Synced: false, CDS: "SYNCED", LDS: "SYNCED", RDS: "SYNCED", EDS: "SYNCED"},
			},
		},
	}
	issues, err := a.Analyze(context.Background(), data)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	found := false
	for _, issue := range issues {
		if issue.ENName == "ISTIO_PROXY_NOT_SYNCED" {
			found = true
		}
	}
	if !found {
		t.Error("Expected ISTIO_PROXY_NOT_SYNCED issue")
	}
}

func TestIstioAnalyzer_VersionMismatch(t *testing.T) {
	a := NewIstioAnalyzer()
	data := &types.DiagnosticData{
		IstioInfo: &types.IstioInfo{
			IstiodPods:    1,
			IstiodReady:   1,
			IngressPods:   1,
			IngressReady:  1,
			ProxyVersions: []string{"1.18.0", "1.19.0"},
		},
	}
	issues, err := a.Analyze(context.Background(), data)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	found := false
	for _, issue := range issues {
		if issue.ENName == "ISTIO_VERSION_MISMATCH" {
			found = true
		}
	}
	if !found {
		t.Error("Expected ISTIO_VERSION_MISMATCH issue")
	}
}

func TestLinkerdAnalyzer_Name(t *testing.T) {
	a := NewLinkerdAnalyzer()
	if a.Name() != "servicemesh.linkerd" {
		t.Errorf("Expected name 'servicemesh.linkerd', got '%s'", a.Name())
	}
}

func TestLinkerdAnalyzer_NoLinkerd(t *testing.T) {
	a := NewLinkerdAnalyzer()
	data := &types.DiagnosticData{}
	issues, err := a.Analyze(context.Background(), data)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues when Linkerd not installed, got %d", len(issues))
	}
}

func TestLinkerdAnalyzer_ControlPlaneMissing(t *testing.T) {
	a := NewLinkerdAnalyzer()
	data := &types.DiagnosticData{
		LinkerdInfo: &types.LinkerdInfo{
			ControlPlanePods: 0,
		},
	}
	issues, err := a.Analyze(context.Background(), data)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	found := false
	for _, issue := range issues {
		if issue.ENName == "LINKERD_CONTROL_PLANE_MISSING" {
			found = true
			if issue.Severity != types.SeverityCritical {
				t.Errorf("Expected critical severity, got %v", issue.Severity)
			}
		}
	}
	if !found {
		t.Error("Expected LINKERD_CONTROL_PLANE_MISSING issue")
	}
}

func TestLinkerdAnalyzer_ProxyUnhealthy(t *testing.T) {
	a := NewLinkerdAnalyzer()
	data := &types.DiagnosticData{
		LinkerdInfo: &types.LinkerdInfo{
			ControlPlanePods:  1,
			ControlPlaneReady: 1,
			ProxyStatus: []types.LinkerdProxyStatus{
				{PodName: "app-1", Namespace: "default", Status: "unhealthy"},
			},
		},
	}
	issues, err := a.Analyze(context.Background(), data)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	found := false
	for _, issue := range issues {
		if issue.ENName == "LINKERD_PROXY_UNHEALTHY" {
			found = true
		}
	}
	if !found {
		t.Error("Expected LINKERD_PROXY_UNHEALTHY issue")
	}
}

func TestLinkerdAnalyzer_HighLatency(t *testing.T) {
	a := NewLinkerdAnalyzer()
	data := &types.DiagnosticData{
		LinkerdInfo: &types.LinkerdInfo{
			ControlPlanePods:  1,
			ControlPlaneReady: 1,
			ServiceLatencies: map[string]float64{
				"my-svc": 2500.0,
			},
		},
	}
	issues, err := a.Analyze(context.Background(), data)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	found := false
	for _, issue := range issues {
		if issue.ENName == "LINKERD_HIGH_LATENCY" {
			found = true
		}
	}
	if !found {
		t.Error("Expected LINKERD_HIGH_LATENCY issue")
	}
}

func TestLinkerdAnalyzer_HighErrorRate(t *testing.T) {
	a := NewLinkerdAnalyzer()
	data := &types.DiagnosticData{
		LinkerdInfo: &types.LinkerdInfo{
			ControlPlanePods:  1,
			ControlPlaneReady: 1,
			ServiceErrorRates: map[string]float64{
				"my-svc": 0.05,
			},
		},
	}
	issues, err := a.Analyze(context.Background(), data)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	found := false
	for _, issue := range issues {
		if issue.ENName == "LINKERD_HIGH_ERROR_RATE" {
			found = true
		}
	}
	if !found {
		t.Error("Expected LINKERD_HIGH_ERROR_RATE issue")
	}
}

func TestLinkerdAnalyzer_Healthy(t *testing.T) {
	a := NewLinkerdAnalyzer()
	data := &types.DiagnosticData{
		LinkerdInfo: &types.LinkerdInfo{
			ControlPlanePods:  3,
			ControlPlaneReady: 3,
		},
	}
	issues, err := a.Analyze(context.Background(), data)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues for healthy Linkerd, got %d", len(issues))
	}
}
