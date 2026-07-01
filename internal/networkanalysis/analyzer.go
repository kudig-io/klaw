package networkanalysis

import (
	"context"
	"time"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Analyzer struct {
	clientset kubernetes.Interface
}

func NewAnalyzer(clientset kubernetes.Interface) *Analyzer {
	return &Analyzer{clientset: clientset}
}

type ExposedService struct {
	Name      string               `json:"name"`
	Namespace string               `json:"namespace"`
	Type      string               `json:"type"`
	Ports     []corev1.ServicePort `json:"ports"`
}

type NetworkAnalysis struct {
	TotalNetworkPolicies int                 `json:"totalNetworkPolicies"`
	TotalServices        int                 `json:"totalServices"`
	TotalIngresses       int                 `json:"totalIngresses"`
	PoliciesByNamespace  map[string][]string `json:"policiesByNamespace"`
	ServicesByType       map[string][]string `json:"servicesByType"`
	IngressesByHost      map[string][]string `json:"ingressesByHost"`
	ExposedServices      []ExposedService    `json:"exposedServices"`
	Timestamp            time.Time           `json:"timestamp"`
}

func (a *Analyzer) ListNetworkPolicies(ctx context.Context, ns string) ([]networkingv1.NetworkPolicy, error) {
	list, err := a.clientset.NetworkingV1().NetworkPolicies(ns).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

func (a *Analyzer) ListIngressClasses(ctx context.Context) ([]networkingv1.IngressClass, error) {
	list, err := a.clientset.NetworkingV1().IngressClasses().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

func (a *Analyzer) AnalyzeNetwork(ctx context.Context) (*NetworkAnalysis, error) {
	policies, err := a.ListNetworkPolicies(ctx, "")
	if err != nil {
		return nil, err
	}
	services, err := a.clientset.CoreV1().Services("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	ingresses, err := a.clientset.NetworkingV1().Ingresses("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	analysis := &NetworkAnalysis{
		TotalNetworkPolicies: len(policies),
		TotalServices:        len(services.Items),
		TotalIngresses:       len(ingresses.Items),
		PoliciesByNamespace:  make(map[string][]string),
		ServicesByType:       make(map[string][]string),
		IngressesByHost:      make(map[string][]string),
		ExposedServices:      make([]ExposedService, 0),
		Timestamp:            time.Now(),
	}

	for _, p := range policies {
		analysis.PoliciesByNamespace[p.Namespace] = append(analysis.PoliciesByNamespace[p.Namespace], p.Name)
	}

	for _, svc := range services.Items {
		typeStr := string(svc.Spec.Type)
		analysis.ServicesByType[typeStr] = append(analysis.ServicesByType[typeStr], svc.Name)
		if isExposed(svc.Spec.Type) {
			analysis.ExposedServices = append(analysis.ExposedServices, ExposedService{
				Name:      svc.Name,
				Namespace: svc.Namespace,
				Type:      typeStr,
				Ports:     svc.Spec.Ports,
			})
		}
	}

	for _, ing := range ingresses.Items {
		for _, rule := range ing.Spec.Rules {
			host := rule.Host
			if host == "" {
				host = "*"
			}
			analysis.IngressesByHost[host] = append(analysis.IngressesByHost[host], ing.Name)
		}
	}

	return analysis, nil
}

func isExposed(t corev1.ServiceType) bool {
	return t == corev1.ServiceTypeLoadBalancer || t == corev1.ServiceTypeNodePort
}
