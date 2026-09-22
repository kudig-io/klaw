package ops

import (
	"fmt"
	"strings"
)

// handler_service.go Service 命令实现：列表、详情与 Endpoints。

// listServices 列出 Service
func (h *Handler) listServices(clusterName, namespace string) (string, error) {
	if err := h.requireKubernetes(); err != nil {
		return "", err
	}

	services, err := h.resources.ListServices(clusterName, namespace)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## Services in `%s`\n\n", namespace))
	sb.WriteString("| Name | Type | Cluster-IP | Ports |\n")
	sb.WriteString("|------|------|------------|-------|\n")
	for _, svc := range services {
		ports := []string{}
		for _, p := range svc.Spec.Ports {
			ports = append(ports, fmt.Sprintf("%d/%s", p.Port, p.Protocol))
		}
		if len(ports) == 0 {
			ports = append(ports, "<none>")
		}
		sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n",
			svc.Name, svc.Spec.Type, svc.Spec.ClusterIP, strings.Join(ports, ",")))
	}
	if len(services) == 0 {
		sb.WriteString("| *(none)* | | | |\n")
	}

	return sb.String(), nil
}

// describeService 描述 Service
func (h *Handler) describeService(clusterName, namespace, serviceName string) (string, error) {
	if err := h.requireKubernetes(); err != nil {
		return "", err
	}

	service, err := h.resources.GetService(clusterName, namespace, serviceName)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## Service: %s\n\n", service.Name))
	sb.WriteString("| Property | Value |\n")
	sb.WriteString("|----------|-------|\n")
	sb.WriteString(fmt.Sprintf("| Namespace | %s |\n", service.Namespace))
	sb.WriteString(fmt.Sprintf("| Type | %s |\n", service.Spec.Type))
	sb.WriteString(fmt.Sprintf("| ClusterIP | %s |\n", service.Spec.ClusterIP))
	sb.WriteString(fmt.Sprintf("| Ports | %d |\n", len(service.Spec.Ports)))

	return sb.String(), nil
}

// getServiceEndpoints 获取 Service Endpoints
func (h *Handler) getServiceEndpoints(clusterName, namespace, serviceName string) (string, error) {
	if err := h.requireKubernetes(); err != nil {
		return "", err
	}

	endpoints, err := h.resources.GetServiceEndpoints(clusterName, namespace, serviceName)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## Endpoints for `%s`\n\n", serviceName))
	sb.WriteString("| IP | Port |\n")
	sb.WriteString("|----|------|\n")
	hasEndpoints := false
	for _, subset := range endpoints.Subsets {
		for _, address := range subset.Addresses {
			for _, port := range subset.Ports {
				sb.WriteString(fmt.Sprintf("| %s | %d/%s |\n", address.IP, port.Port, port.Protocol))
				hasEndpoints = true
			}
		}
	}
	if !hasEndpoints {
		sb.WriteString("| *(none)* | |\n")
	}

	return sb.String(), nil
}
