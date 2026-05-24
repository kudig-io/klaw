package api

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/gorilla/mux"
)

// UnifiedResourceInfo mirrors the generic resource shape used by the fusion plan.
type UnifiedResourceInfo struct {
	Name              string            `json:"name"`
	Namespace         string            `json:"namespace,omitempty"`
	Kind              string            `json:"kind"`
	CreationTimestamp time.Time         `json:"creationTimestamp"`
	Labels            map[string]string `json:"labels,omitempty"`
	Annotations       map[string]string `json:"annotations,omitempty"`
	Status            string            `json:"status,omitempty"`
	Raw               interface{}       `json:"raw,omitempty"`
}

// UnifiedResourceList provides a consistent response envelope for generic resource APIs.
type UnifiedResourceList struct {
	Items        []UnifiedResourceInfo `json:"items"`
	Total        int                   `json:"total"`
	ResourceKind string                `json:"resourceKind"`
}

func (s *Server) setupUnifiedV1Routes() {
	s.router.HandleFunc("/api/v1/clusters", s.handleGetClusters).Methods("GET")
	s.router.HandleFunc("/api/v1/clusters/{name}", s.handleGetCluster).Methods("GET")
	s.router.HandleFunc("/api/v1/clusters/{name}/status", s.handleGetClusterStatus).Methods("GET")
	s.router.HandleFunc("/api/v1/clusters/{name}/metrics", s.handleGetClusterMetrics).Methods("GET")
	s.router.HandleFunc("/api/v1/clusters/{name}/namespaces", s.handleGetNamespaces).Methods("GET")

	s.router.HandleFunc("/api/v1/clusters/{cluster}/pods", s.handleListAllPods).Methods("GET")
	s.router.HandleFunc("/api/v1/clusters/{cluster}/namespaces/{namespace}/pods", s.handleListPods).Methods("GET")
	s.router.HandleFunc("/api/v1/clusters/{cluster}/namespaces/{namespace}/pods/{name}", s.handleGetPod).Methods("GET")
	s.router.HandleFunc("/api/v1/clusters/{cluster}/namespaces/{namespace}/pods/{name}/logs", s.handleGetPodLogs).Methods("GET")
	s.router.HandleFunc("/api/v1/clusters/{cluster}/namespaces/{namespace}/pods/{name}/logs/analysis", s.handleAnalyzePodLogs).Methods("GET")
	s.router.HandleFunc("/api/v1/clusters/{cluster}/namespaces/{namespace}/pods/{name}", s.handleDeletePod).Methods("DELETE")

	s.router.HandleFunc("/api/v1/clusters/{cluster}/nodes", s.handleListNodes).Methods("GET")
	s.router.HandleFunc("/api/v1/clusters/{cluster}/nodes/{name}", s.handleGetNode).Methods("GET")
	s.router.HandleFunc("/api/v1/clusters/{cluster}/nodes/metrics", s.handleGetNodeMetrics).Methods("GET")

	s.router.HandleFunc("/api/v1/clusters/{cluster}/deployments", s.handleListAllDeployments).Methods("GET")
	s.router.HandleFunc("/api/v1/clusters/{cluster}/namespaces/{namespace}/deployments", s.handleListDeployments).Methods("GET")
	s.router.HandleFunc("/api/v1/clusters/{cluster}/namespaces/{namespace}/deployments/{name}", s.handleGetDeployment).Methods("GET")
	s.router.HandleFunc("/api/v1/clusters/{cluster}/namespaces/{namespace}/deployments/{name}/scale", s.handleScaleDeployment).Methods("POST")
	s.router.HandleFunc("/api/v1/clusters/{cluster}/namespaces/{namespace}/deployments/{name}/restart", s.handleRestartDeployment).Methods("POST")
	s.router.HandleFunc("/api/v1/clusters/{cluster}/namespaces/{namespace}/deployments/{name}/pods", s.handleGetDeploymentPods).Methods("GET")
	s.router.HandleFunc("/api/v1/clusters/{cluster}/namespaces/{namespace}/deployments/{name}/status", s.handleGetDeploymentStatus).Methods("GET")

	s.router.HandleFunc("/api/v1/clusters/{cluster}/services", s.handleListAllServices).Methods("GET")
	s.router.HandleFunc("/api/v1/clusters/{cluster}/namespaces/{namespace}/services", s.handleListServices).Methods("GET")
	s.router.HandleFunc("/api/v1/clusters/{cluster}/namespaces/{namespace}/services/{name}", s.handleGetService).Methods("GET")
	s.router.HandleFunc("/api/v1/clusters/{cluster}/namespaces/{namespace}/services/{name}/endpoints", s.handleGetServiceEndpoints).Methods("GET")
	s.router.HandleFunc("/api/v1/clusters/{cluster}/namespaces/{namespace}/services/{name}", s.handleDeleteService).Methods("DELETE")

	s.router.HandleFunc("/api/v1/clusters/{cluster}/events", s.handleGetEvents).Methods("GET")
	s.router.HandleFunc("/api/v1/clusters/{cluster}/namespaces/{namespace}/events", s.handleGetEvents).Methods("GET")

	s.router.HandleFunc("/api/v1/clusters/{cluster}/monitor/status", s.handleGetMonitorStatus).Methods("GET")
	s.router.HandleFunc("/api/v1/clusters/{cluster}/monitor/alerts", s.handleGetMonitorAlerts).Methods("GET")
	s.router.HandleFunc("/api/v1/clusters/{cluster}/monitor/history", s.handleGetMetricsHistory).Methods("GET")
	s.router.HandleFunc("/api/v1/clusters/{cluster}/alerts/rules", s.handleGetAlertRules).Methods("GET")
	s.router.HandleFunc("/api/v1/clusters/{cluster}/alerts/rules", s.handleCreateAlertRule).Methods("POST")
	s.router.HandleFunc("/api/v1/clusters/{cluster}/alerts/rules/{id}", s.handleUpdateAlertRule).Methods("PUT")
	s.router.HandleFunc("/api/v1/clusters/{cluster}/alerts/rules/{id}", s.handleDeleteAlertRule).Methods("DELETE")
	s.router.HandleFunc("/api/v1/clusters/{cluster}/alerts/evaluate", s.handleEvaluateAlerts).Methods("POST")
	s.router.HandleFunc("/api/v1/clusters/{cluster}/alerts/history", s.handleGetAlertHistory).Methods("GET")
	s.router.HandleFunc("/api/v1/clusters/{cluster}/alerts/stats", s.handleGetAlertStats).Methods("GET")
	s.router.HandleFunc("/api/v1/clusters/{cluster}/alerts/{id}/acknowledge", s.handleAcknowledgeAlert).Methods("POST")
	s.router.HandleFunc("/api/v1/clusters/{cluster}/alerts/{id}/resolve", s.handleResolveAlertRecord).Methods("POST")
	s.router.HandleFunc("/api/v1/clusters/{cluster}/backups", s.handleListBackups).Methods("GET")
	s.router.HandleFunc("/api/v1/clusters/{cluster}/backups", s.handleCreateBackup).Methods("POST")
	s.router.HandleFunc("/api/v1/clusters/{cluster}/backups/summary", s.handleBackupSummary).Methods("GET")
	s.router.HandleFunc("/api/v1/clusters/{cluster}/backups/{name}", s.handleGetBackup).Methods("GET")
	s.router.HandleFunc("/api/v1/clusters/{cluster}/backups/{name}", s.handleDeleteBackup).Methods("DELETE")
	s.router.HandleFunc("/api/v1/tenants", s.handleListTenants).Methods("GET")
	s.router.HandleFunc("/api/v1/tenants", s.handleCreateTenant).Methods("POST")
	s.router.HandleFunc("/api/v1/tenants/stats", s.handleTenantStatistics).Methods("GET")
	s.router.HandleFunc("/api/v1/tenants/{id}", s.handleGetTenant).Methods("GET")
	s.router.HandleFunc("/api/v1/tenants/{id}", s.handleUpdateTenant).Methods("PUT")
	s.router.HandleFunc("/api/v1/tenants/{id}", s.handleDeleteTenant).Methods("DELETE")
	s.router.HandleFunc("/api/v1/tenant-users", s.handleListTenantUsers).Methods("GET")
	s.router.HandleFunc("/api/v1/tenant-users", s.handleCreateTenantUser).Methods("POST")
	s.router.HandleFunc("/api/v1/tenant-users/{id}", s.handleDeleteTenantUser).Methods("DELETE")
	s.router.HandleFunc("/api/v1/audit/logs", s.handleAuditLogs).Methods("GET")
	s.router.HandleFunc("/api/v1/audit/stats", s.handleAuditStats).Methods("GET")
	s.router.HandleFunc("/api/v1/clusters/{cluster}/rbac/analysis", s.handleAnalyzeRBAC).Methods("GET")
	s.router.HandleFunc("/api/v1/analysis/logs", s.handleAnalyzeRawLogs).Methods("POST")

	s.router.HandleFunc("/api/v1/clusters/{cluster}/resources/{kind}", s.handleListUnifiedResources).Methods("GET")
	s.router.HandleFunc("/api/v1/clusters/{cluster}/namespaces/{namespace}/resources/{kind}", s.handleListUnifiedResources).Methods("GET")
	s.router.HandleFunc("/api/v1/clusters/{cluster}/resources/{kind}/{name}", s.handleGetUnifiedResource).Methods("GET")
	s.router.HandleFunc("/api/v1/clusters/{cluster}/namespaces/{namespace}/resources/{kind}/{name}", s.handleGetUnifiedResource).Methods("GET")
}

func (s *Server) handleListUnifiedResources(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clusterName := vars["cluster"]
	kind := normalizeResourceKind(vars["kind"])
	namespace := vars["namespace"]

	items, err := s.listUnifiedResources(clusterName, namespace, kind)
	if err != nil {
		s.respondError(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.respondJSON(w, UnifiedResourceList{
		Items:        items,
		Total:        len(items),
		ResourceKind: kind,
	}, http.StatusOK)
}

func (s *Server) handleGetUnifiedResource(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clusterName := vars["cluster"]
	kind := normalizeResourceKind(vars["kind"])
	name := vars["name"]
	namespace := vars["namespace"]

	item, err := s.getUnifiedResource(clusterName, namespace, kind, name)
	if err != nil {
		s.respondError(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.respondJSON(w, item, http.StatusOK)
}

func (s *Server) listUnifiedResources(clusterName, namespace, kind string) ([]UnifiedResourceInfo, error) {
	switch kind {
	case "namespaces":
		if namespace != "" {
			return nil, errClusterScopedResource(kind)
		}
		namespaces, err := s.resources.ListNamespaces(clusterName)
		if err != nil {
			return nil, err
		}
		items := make([]UnifiedResourceInfo, 0, len(namespaces))
		for _, ns := range namespaces {
			items = append(items, toUnifiedNamespace(ns))
		}
		return items, nil
	case "nodes":
		if namespace != "" {
			return nil, errClusterScopedResource(kind)
		}
		nodes, err := s.resources.ListNodes(clusterName)
		if err != nil {
			return nil, err
		}
		items := make([]UnifiedResourceInfo, 0, len(nodes))
		for _, node := range nodes {
			items = append(items, toUnifiedNode(node))
		}
		return items, nil
	case "pods":
		pods, err := s.resources.ListPods(clusterName, namespace)
		if err != nil {
			return nil, err
		}
		items := make([]UnifiedResourceInfo, 0, len(pods))
		for _, pod := range pods {
			items = append(items, toUnifiedPod(pod))
		}
		return items, nil
	case "deployments":
		deployments, err := s.resources.ListDeployments(clusterName, namespace)
		if err != nil {
			return nil, err
		}
		items := make([]UnifiedResourceInfo, 0, len(deployments))
		for _, deployment := range deployments {
			items = append(items, toUnifiedDeployment(deployment))
		}
		return items, nil
	case "services":
		services, err := s.resources.ListServices(clusterName, namespace)
		if err != nil {
			return nil, err
		}
		items := make([]UnifiedResourceInfo, 0, len(services))
		for _, service := range services {
			items = append(items, toUnifiedService(service))
		}
		return items, nil
	case "events":
		events, err := s.resources.ListEvents(clusterName, namespace)
		if err != nil {
			return nil, err
		}
		items := make([]UnifiedResourceInfo, 0, len(events))
		for _, event := range events {
			items = append(items, toUnifiedEvent(event))
		}
		return items, nil
	case "configmaps":
		configMaps, err := s.resources.ListConfigMaps(clusterName, namespace)
		if err != nil {
			return nil, err
		}
		items := make([]UnifiedResourceInfo, 0, len(configMaps))
		for _, configMap := range configMaps {
			items = append(items, toUnifiedConfigMap(configMap))
		}
		return items, nil
	case "statefulsets":
		statefulSets, err := s.resources.ListStatefulSets(clusterName, namespace)
		if err != nil {
			return nil, err
		}
		items := make([]UnifiedResourceInfo, 0, len(statefulSets))
		for _, statefulSet := range statefulSets {
			items = append(items, toUnifiedStatefulSet(statefulSet))
		}
		return items, nil
	case "ingresses":
		ingresses, err := s.resources.ListIngresses(clusterName, namespace)
		if err != nil {
			return nil, err
		}
		items := make([]UnifiedResourceInfo, 0, len(ingresses))
		for _, ingress := range ingresses {
			items = append(items, toUnifiedIngress(ingress))
		}
		return items, nil
	default:
		return nil, errUnsupportedResource(kind)
	}
}

func (s *Server) getUnifiedResource(clusterName, namespace, kind, name string) (*UnifiedResourceInfo, error) {
	switch kind {
	case "namespaces":
		if namespace != "" {
			return nil, errClusterScopedResource(kind)
		}
		ns, err := s.resources.GetNamespace(clusterName, name)
		if err != nil {
			return nil, err
		}
		item := toUnifiedNamespace(*ns)
		return &item, nil
	case "nodes":
		if namespace != "" {
			return nil, errClusterScopedResource(kind)
		}
		node, err := s.resources.GetNode(clusterName, name)
		if err != nil {
			return nil, err
		}
		item := toUnifiedNode(*node)
		return &item, nil
	case "pods":
		if namespace == "" {
			return nil, errNamespaceRequired(kind)
		}
		pod, err := s.resources.GetPod(clusterName, namespace, name)
		if err != nil {
			return nil, err
		}
		item := toUnifiedPod(*pod)
		return &item, nil
	case "deployments":
		if namespace == "" {
			return nil, errNamespaceRequired(kind)
		}
		deployment, err := s.resources.GetDeployment(clusterName, namespace, name)
		if err != nil {
			return nil, err
		}
		item := toUnifiedDeployment(*deployment)
		return &item, nil
	case "services":
		if namespace == "" {
			return nil, errNamespaceRequired(kind)
		}
		service, err := s.resources.GetService(clusterName, namespace, name)
		if err != nil {
			return nil, err
		}
		item := toUnifiedService(*service)
		return &item, nil
	case "configmaps":
		if namespace == "" {
			return nil, errNamespaceRequired(kind)
		}
		configMap, err := s.resources.GetConfigMap(clusterName, namespace, name)
		if err != nil {
			return nil, err
		}
		item := toUnifiedConfigMap(*configMap)
		return &item, nil
	case "statefulsets":
		if namespace == "" {
			return nil, errNamespaceRequired(kind)
		}
		statefulSet, err := s.resources.GetStatefulSet(clusterName, namespace, name)
		if err != nil {
			return nil, err
		}
		item := toUnifiedStatefulSet(*statefulSet)
		return &item, nil
	case "ingresses":
		if namespace == "" {
			return nil, errNamespaceRequired(kind)
		}
		ingress, err := s.resources.GetIngress(clusterName, namespace, name)
		if err != nil {
			return nil, err
		}
		item := toUnifiedIngress(*ingress)
		return &item, nil
	default:
		return nil, errUnsupportedResource(kind)
	}
}

func normalizeResourceKind(kind string) string {
	kind = strings.ToLower(strings.TrimSpace(kind))
	switch kind {
	case "namespace":
		return "namespaces"
	case "node":
		return "nodes"
	case "pod":
		return "pods"
	case "deployment":
		return "deployments"
	case "service":
		return "services"
	case "event":
		return "events"
	case "configmap":
		return "configmaps"
	case "statefulset":
		return "statefulsets"
	case "ingress":
		return "ingresses"
	default:
		return kind
	}
}

func errUnsupportedResource(kind string) error {
	return &routeError{message: "unsupported resource kind: " + kind}
}

func errNamespaceRequired(kind string) error {
	return &routeError{message: "namespace is required for resource kind: " + kind}
}

func errClusterScopedResource(kind string) error {
	return &routeError{message: "resource kind is cluster-scoped: " + kind}
}

type routeError struct {
	message string
}

func (e *routeError) Error() string {
	return e.message
}

func toUnifiedNamespace(ns corev1.Namespace) UnifiedResourceInfo {
	return UnifiedResourceInfo{
		Name:              ns.Name,
		Kind:              "Namespace",
		CreationTimestamp: ns.CreationTimestamp.Time,
		Labels:            ns.Labels,
		Annotations:       ns.Annotations,
		Status:            string(ns.Status.Phase),
		Raw:               ns,
	}
}

func toUnifiedNode(node corev1.Node) UnifiedResourceInfo {
	return UnifiedResourceInfo{
		Name:              node.Name,
		Kind:              "Node",
		CreationTimestamp: node.CreationTimestamp.Time,
		Labels:            node.Labels,
		Annotations:       node.Annotations,
		Status:            summarizeNodeStatus(node.Status.Conditions),
		Raw:               node,
	}
}

func toUnifiedPod(pod corev1.Pod) UnifiedResourceInfo {
	return UnifiedResourceInfo{
		Name:              pod.Name,
		Namespace:         pod.Namespace,
		Kind:              "Pod",
		CreationTimestamp: pod.CreationTimestamp.Time,
		Labels:            pod.Labels,
		Annotations:       pod.Annotations,
		Status:            string(pod.Status.Phase),
		Raw:               pod,
	}
}

func toUnifiedDeployment(deployment appsv1.Deployment) UnifiedResourceInfo {
	desired := int32(0)
	if deployment.Spec.Replicas != nil {
		desired = *deployment.Spec.Replicas
	}

	status := strings.TrimSpace(
		strings.Join([]string{
			"ready=" + int32ToString(deployment.Status.ReadyReplicas),
			"available=" + int32ToString(deployment.Status.AvailableReplicas),
			"desired=" + int32ToString(desired),
		}, ", "),
	)

	return UnifiedResourceInfo{
		Name:              deployment.Name,
		Namespace:         deployment.Namespace,
		Kind:              "Deployment",
		CreationTimestamp: deployment.CreationTimestamp.Time,
		Labels:            deployment.Labels,
		Annotations:       deployment.Annotations,
		Status:            status,
		Raw:               deployment,
	}
}

func toUnifiedService(service corev1.Service) UnifiedResourceInfo {
	status := string(service.Spec.Type)
	if len(service.Status.LoadBalancer.Ingress) > 0 {
		status += " (loadBalancerReady)"
	}

	return UnifiedResourceInfo{
		Name:              service.Name,
		Namespace:         service.Namespace,
		Kind:              "Service",
		CreationTimestamp: service.CreationTimestamp.Time,
		Labels:            service.Labels,
		Annotations:       service.Annotations,
		Status:            status,
		Raw:               service,
	}
}

func toUnifiedEvent(event corev1.Event) UnifiedResourceInfo {
	return UnifiedResourceInfo{
		Name:              event.Name,
		Namespace:         event.Namespace,
		Kind:              "Event",
		CreationTimestamp: event.CreationTimestamp.Time,
		Labels:            event.Labels,
		Annotations:       event.Annotations,
		Status:            strings.TrimSpace(event.Type + " " + event.Reason),
		Raw:               event,
	}
}

func toUnifiedConfigMap(configMap corev1.ConfigMap) UnifiedResourceInfo {
	return UnifiedResourceInfo{
		Name:              configMap.Name,
		Namespace:         configMap.Namespace,
		Kind:              "ConfigMap",
		CreationTimestamp: configMap.CreationTimestamp.Time,
		Labels:            configMap.Labels,
		Annotations:       configMap.Annotations,
		Status:            "keys=" + strconv.Itoa(len(configMap.Data)+len(configMap.BinaryData)),
		Raw:               configMap,
	}
}

func toUnifiedStatefulSet(statefulSet appsv1.StatefulSet) UnifiedResourceInfo {
	desired := int32(0)
	if statefulSet.Spec.Replicas != nil {
		desired = *statefulSet.Spec.Replicas
	}

	return UnifiedResourceInfo{
		Name:              statefulSet.Name,
		Namespace:         statefulSet.Namespace,
		Kind:              "StatefulSet",
		CreationTimestamp: statefulSet.CreationTimestamp.Time,
		Labels:            statefulSet.Labels,
		Annotations:       statefulSet.Annotations,
		Status:            "ready=" + int32ToString(statefulSet.Status.ReadyReplicas) + ", desired=" + int32ToString(desired),
		Raw:               statefulSet,
	}
}

func toUnifiedIngress(ingress networkingv1.Ingress) UnifiedResourceInfo {
	hostCount := 0
	for _, rule := range ingress.Spec.Rules {
		if rule.Host != "" {
			hostCount++
		}
	}

	return UnifiedResourceInfo{
		Name:              ingress.Name,
		Namespace:         ingress.Namespace,
		Kind:              "Ingress",
		CreationTimestamp: ingress.CreationTimestamp.Time,
		Labels:            ingress.Labels,
		Annotations:       ingress.Annotations,
		Status:            "hosts=" + strconv.Itoa(hostCount),
		Raw:               ingress,
	}
}

func summarizeNodeStatus(conditions []corev1.NodeCondition) string {
	for _, condition := range conditions {
		if condition.Type == corev1.NodeReady {
			return string(condition.Status)
		}
	}

	return string(metav1.ConditionUnknown)
}

func int32ToString(v int32) string {
	return strconv.FormatInt(int64(v), 10)
}
