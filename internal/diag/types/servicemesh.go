package types

// IstioInfo contains Istio service mesh information
type IstioInfo struct {
	// Istiod control plane
	IstiodPods  int `json:"istiod_pods" yaml:"istiod_pods"`
	IstiodReady int `json:"istiod_ready" yaml:"istiod_ready"`

	// Ingress Gateway
	IngressPods  int `json:"ingress_pods" yaml:"ingress_pods"`
	IngressReady int `json:"ingress_ready" yaml:"ingress_ready"`

	// Egress Gateway
	EgressPods  int `json:"egress_pods" yaml:"egress_pods"`
	EgressReady int `json:"egress_ready" yaml:"egress_ready"`

	// Sidecar injection configuration per namespace
	InjectionEnabled map[string]bool `json:"injection_enabled,omitempty" yaml:"injection_enabled,omitempty"`

	// Proxy status for each sidecar
	ProxyStatus []IstioProxyStatus `json:"proxy_status,omitempty" yaml:"proxy_status,omitempty"`

	// Pods without sidecar in mesh-enabled namespaces
	PodsWithoutSidecar []PodInfo `json:"pods_without_sidecar,omitempty" yaml:"pods_without_sidecar,omitempty"`

	// Proxy versions found
	ProxyVersions []string `json:"proxy_versions,omitempty" yaml:"proxy_versions,omitempty"`

	// mTLS enabled
	MTLSEnabled bool `json:"mtls_enabled" yaml:"mtls_enabled"`

	// VirtualServices
	VirtualServices []VirtualServiceInfo `json:"virtual_services,omitempty" yaml:"virtual_services,omitempty"`

	// DestinationRules
	DestinationRules []DestinationRuleInfo `json:"destination_rules,omitempty" yaml:"destination_rules,omitempty"`
}

// IstioProxyStatus represents the status of an Istio sidecar proxy
type IstioProxyStatus struct {
	PodName   string `json:"pod_name" yaml:"pod_name"`
	Namespace string `json:"namespace" yaml:"namespace"`
	Synced    bool   `json:"synced" yaml:"synced"`
	CDS       string `json:"cds" yaml:"cds"`
	LDS       string `json:"lds" yaml:"lds"`
	RDS       string `json:"rds" yaml:"rds"`
	EDS       string `json:"eds" yaml:"eds"`
	Version   string `json:"version,omitempty" yaml:"version,omitempty"`
}

// PodInfo represents basic pod information
type PodInfo struct {
	Name      string `json:"name" yaml:"name"`
	Namespace string `json:"namespace" yaml:"namespace"`
}

// VirtualServiceInfo represents Istio VirtualService information
type VirtualServiceInfo struct {
	Name      string `json:"name" yaml:"name"`
	Namespace string `json:"namespace" yaml:"namespace"`
	Hosts     []string `json:"hosts,omitempty" yaml:"hosts,omitempty"`
	Gateways  []string `json:"gateways,omitempty" yaml:"gateways,omitempty"`
	HTTP      []HTTPRouteInfo `json:"http,omitempty" yaml:"http,omitempty"`
	TCP       []TCPRouteInfo `json:"tcp,omitempty" yaml:"tcp,omitempty"`
}

// HTTPRouteInfo represents HTTP route information
type HTTPRouteInfo struct {
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	Match []string `json:"match,omitempty" yaml:"match,omitempty"`
	Route []DestinationInfo `json:"route,omitempty" yaml:"route,omitempty"`
}

// TCPRouteInfo represents TCP route information
type TCPRouteInfo struct {
	Match []string `json:"match,omitempty" yaml:"match,omitempty"`
	Route []DestinationInfo `json:"route,omitempty" yaml:"route,omitempty"`
}

// DestinationInfo represents destination information
type DestinationInfo struct {
	Host   string `json:"host" yaml:"host"`
	Subset string `json:"subset,omitempty" yaml:"subset,omitempty"`
	Port   int    `json:"port,omitempty" yaml:"port,omitempty"`
}

// DestinationRuleInfo represents Istio DestinationRule information
type DestinationRuleInfo struct {
	Name      string `json:"name" yaml:"name"`
	Namespace string `json:"namespace" yaml:"namespace"`
	Host      string `json:"host" yaml:"host"`
	Subsets   []SubsetInfo `json:"subsets,omitempty" yaml:"subsets,omitempty"`
	TrafficPolicy *TrafficPolicyInfo `json:"traffic_policy,omitempty" yaml:"traffic_policy,omitempty"`
}

// SubsetInfo represents subset information
type SubsetInfo struct {
	Name   string `json:"name" yaml:"name"`
	Labels map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
}

// TrafficPolicyInfo represents traffic policy information
type TrafficPolicyInfo struct {
	LoadBalancer string `json:"load_balancer,omitempty" yaml:"load_balancer,omitempty"`
	OutlierDetection *OutlierDetectionInfo `json:"outlier_detection,omitempty" yaml:"outlier_detection,omitempty"`
}

// OutlierDetectionInfo represents outlier detection information
type OutlierDetectionInfo struct {
	ConsecutiveErrors  int `json:"consecutive_errors,omitempty" yaml:"consecutive_errors,omitempty"`
	IntervalSeconds    int `json:"interval_seconds,omitempty" yaml:"interval_seconds,omitempty"`
	BaseEjectionTimeSeconds int `json:"base_ejection_time_seconds,omitempty" yaml:"base_ejection_time_seconds,omitempty"`
}

// LinkerdInfo contains Linkerd service mesh information
type LinkerdInfo struct {
	// Control plane
	ControlPlanePods  int `json:"control_plane_pods" yaml:"control_plane_pods"`
	ControlPlaneReady int `json:"control_plane_ready" yaml:"control_plane_ready"`

	// Proxy status
	ProxyStatus []LinkerdProxyStatus `json:"proxy_status,omitempty" yaml:"proxy_status,omitempty"`

	// Pods without proxy in mesh-enabled namespaces
	PodsWithoutProxy []PodInfo `json:"pods_without_proxy,omitempty" yaml:"pods_without_proxy,omitempty"`

	// Proxy versions found
	ProxyVersions []string `json:"proxy_versions,omitempty" yaml:"proxy_versions,omitempty"`

	// Service latencies (P99 in ms)
	ServiceLatencies map[string]float64 `json:"service_latencies,omitempty" yaml:"service_latencies,omitempty"`

	// Service error rates
	ServiceErrorRates map[string]float64 `json:"service_error_rates,omitempty" yaml:"service_error_rates,omitempty"`
}

// LinkerdProxyStatus represents the status of a Linkerd proxy
type LinkerdProxyStatus struct {
	PodName   string `json:"pod_name" yaml:"pod_name"`
	Namespace string `json:"namespace" yaml:"namespace"`
	Status    string `json:"status" yaml:"status"` // healthy, unhealthy
	Version   string `json:"version,omitempty" yaml:"version,omitempty"`
}
