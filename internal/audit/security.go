package audit

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type SecurityEvent struct {
	ID                string              `json:"id"`
	Timestamp         time.Time           `json:"timestamp"`
	Type              string              `json:"type"`
	Severity          string              `json:"severity"`
	Title             string              `json:"title"`
	Description       string              `json:"description"`
	Source            string              `json:"source"`
	AffectedResources []map[string]string `json:"affectedResources,omitempty"`
	Remediation       string              `json:"remediation,omitempty"`
	Status            string              `json:"status"`
	Acknowledged      bool                `json:"acknowledged"`
	AcknowledgedBy    string              `json:"acknowledgedBy,omitempty"`
	AcknowledgedAt    *time.Time          `json:"acknowledgedAt,omitempty"`
	Resolved          bool                `json:"resolved"`
	ResolvedBy        string              `json:"resolvedBy,omitempty"`
	ResolvedAt        *time.Time          `json:"resolvedAt,omitempty"`
}

type SecurityEventFilter struct {
	Type       string `json:"type,omitempty"`
	Severity   string `json:"severity,omitempty"`
	Status     string `json:"status,omitempty"`
	Unresolved bool   `json:"unresolved,omitempty"`
	Limit      int    `json:"limit,omitempty"`
}

type ComplianceReport struct {
	ID         string        `json:"id"`
	Timestamp  time.Time     `json:"timestamp"`
	ReportType string        `json:"reportType"`
	Status     string        `json:"status"`
	Summary    ReportSummary `json:"summary"`
	Findings   []Finding     `json:"findings"`
	Score      int           `json:"score"`
}

type ReportSummary struct {
	TotalFindings    int `json:"totalFindings"`
	CriticalFindings int `json:"criticalFindings"`
	HighFindings     int `json:"highFindings"`
	MediumFindings   int `json:"mediumFindings"`
	LowFindings      int `json:"lowFindings"`
	InfoFindings     int `json:"infoFindings"`
}

type Finding struct {
	Category       string            `json:"category"`
	Severity       string            `json:"severity"`
	Title          string            `json:"title"`
	Description    string            `json:"description"`
	Resource       map[string]string `json:"resource,omitempty"`
	Recommendation string            `json:"recommendation,omitempty"`
}

func (l *Logger) CreateSecurityEvent(event SecurityEvent) SecurityEvent {
	l.mu.Lock()
	defer l.mu.Unlock()

	if event.ID == "" {
		event.ID = uuid.New().String()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}
	if event.Status == "" {
		event.Status = "open"
	}
	l.securityEvents = append(l.securityEvents, event)
	_ = l.saveSecurityLocked()
	return event
}

func (l *Logger) GetSecurityEvents(filter SecurityEventFilter) []SecurityEvent {
	l.mu.RLock()
	defer l.mu.RUnlock()

	var result []SecurityEvent
	for _, e := range l.securityEvents {
		if filter.Type != "" && e.Type != filter.Type {
			continue
		}
		if filter.Severity != "" && e.Severity != filter.Severity {
			continue
		}
		if filter.Status != "" && e.Status != filter.Status {
			continue
		}
		if filter.Unresolved && e.Resolved {
			continue
		}
		result = append(result, e)
		if filter.Limit > 0 && len(result) >= filter.Limit {
			break
		}
	}
	return result
}

func (l *Logger) GetSecurityEvent(id string) (SecurityEvent, bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	for _, e := range l.securityEvents {
		if e.ID == id {
			return e, true
		}
	}
	return SecurityEvent{}, false
}

func (l *Logger) AcknowledgeSecurityEvent(id, by string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	for i := range l.securityEvents {
		if l.securityEvents[i].ID == id {
			now := time.Now().UTC()
			l.securityEvents[i].Acknowledged = true
			l.securityEvents[i].AcknowledgedBy = by
			l.securityEvents[i].AcknowledgedAt = &now
			return l.saveSecurityLocked()
		}
	}
	return fmt.Errorf("security event %s not found", id)
}

func (l *Logger) ResolveSecurityEvent(id, by string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	for i := range l.securityEvents {
		if l.securityEvents[i].ID == id {
			now := time.Now().UTC()
			l.securityEvents[i].Resolved = true
			l.securityEvents[i].ResolvedBy = by
			l.securityEvents[i].ResolvedAt = &now
			l.securityEvents[i].Status = "resolved"
			return l.saveSecurityLocked()
		}
	}
	return fmt.Errorf("security event %s not found", id)
}

func (l *Logger) ListComplianceReports() []ComplianceReport {
	l.mu.RLock()
	defer l.mu.RUnlock()

	result := make([]ComplianceReport, len(l.complianceReports))
	for i, r := range l.complianceReports {
		result[i] = r
	}
	return result
}

func (l *Logger) GenerateComplianceReport(ctx context.Context, clientset kubernetes.Interface, reportType string) (*ComplianceReport, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	var findings []Finding

	for _, check := range []func(context.Context, kubernetes.Interface) ([]Finding, error){
		l.checkPodSecurity,
		l.checkNodeSecurity,
		l.checkNamespaceSecurity,
		l.checkSecretSecurity,
		l.checkRBACSecurity,
		l.checkNetworkSecurity,
	} {
		f, err := check(ctx, clientset)
		if err != nil {
			return nil, err
		}
		findings = append(findings, f...)
	}

	report := ComplianceReport{
		ID:         uuid.New().String(),
		Timestamp:  time.Now().UTC(),
		ReportType: reportType,
		Status:     "completed",
		Summary:    buildSummary(findings),
		Findings:   findings,
		Score:      calculateComplianceScore(findings),
	}
	l.complianceReports = append(l.complianceReports, report)
	_ = l.saveComplianceLocked()
	return &report, nil
}

func (l *Logger) checkPodSecurity(ctx context.Context, clientset kubernetes.Interface) ([]Finding, error) {
	pods, err := clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	var findings []Finding
	for _, pod := range pods.Items {
		res := map[string]string{"kind": "pod", "namespace": pod.Namespace, "name": pod.Name}
		for _, c := range pod.Spec.Containers {
			if c.SecurityContext != nil && c.SecurityContext.Privileged != nil && *c.SecurityContext.Privileged {
				findings = append(findings, Finding{
					Category:       "pod-security",
					Severity:       "critical",
					Title:          "Privileged container detected",
					Description:    fmt.Sprintf("Container %s in pod %s/%s runs in privileged mode", c.Name, pod.Namespace, pod.Name),
					Resource:       res,
					Recommendation: "Remove privileged flag and use specific capabilities instead",
				})
			}
			if c.SecurityContext == nil {
				findings = append(findings, Finding{
					Category:       "pod-security",
					Severity:       "medium",
					Title:          "No securityContext defined",
					Description:    fmt.Sprintf("Container %s in pod %s/%s has no securityContext", c.Name, pod.Namespace, pod.Name),
					Resource:       res,
					Recommendation: "Define a securityContext with runAsNonRoot, readOnlyRootFilesystem, and drop all capabilities",
				})
			}
			if len(c.Resources.Limits) == 0 && len(c.Resources.Requests) == 0 {
				findings = append(findings, Finding{
					Category:       "pod-security",
					Severity:       "medium",
					Title:          "No resource limits defined",
					Description:    fmt.Sprintf("Container %s in pod %s/%s has no resource limits or requests", c.Name, pod.Namespace, pod.Name),
					Resource:       res,
					Recommendation: "Set CPU and memory requests and limits to prevent resource exhaustion",
				})
			}
		}
	}
	return findings, nil
}

func (l *Logger) checkNodeSecurity(ctx context.Context, clientset kubernetes.Interface) ([]Finding, error) {
	nodes, err := clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	var findings []Finding
	for _, node := range nodes.Items {
		res := map[string]string{"kind": "node", "name": node.Name}
		if len(node.Spec.Taints) == 0 {
			findings = append(findings, Finding{
				Category:       "node-security",
				Severity:       "low",
				Title:          "Node has no taints",
				Description:    fmt.Sprintf("Node %s has no taints, allowing any pod to be scheduled", node.Name),
				Resource:       res,
				Recommendation: "Apply appropriate taints to restrict workload placement",
			})
		}
		hasRole := false
		for label := range node.Labels {
			if strings.HasPrefix(label, "node-role.kubernetes.io/") {
				hasRole = true
				break
			}
		}
		if !hasRole {
			findings = append(findings, Finding{
				Category:       "node-security",
				Severity:       "info",
				Title:          "No node role labels",
				Description:    fmt.Sprintf("Node %s has no role labels", node.Name),
				Resource:       res,
				Recommendation: "Assign node role labels for better scheduling control",
			})
		}
	}
	return findings, nil
}

func (l *Logger) checkNamespaceSecurity(ctx context.Context, clientset kubernetes.Interface) ([]Finding, error) {
	namespaces, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	var findings []Finding
	for _, ns := range namespaces.Items {
		res := map[string]string{"kind": "namespace", "name": ns.Name}
		if ns.Name == "default" {
			findings = append(findings, Finding{
				Category:       "namespace-security",
				Severity:       "medium",
				Title:          "Workloads in default namespace",
				Description:    "The default namespace should not be used for production workloads",
				Resource:       res,
				Recommendation: "Move workloads to dedicated namespaces",
			})
		}
		if len(ns.Labels) <= 1 {
			findings = append(findings, Finding{
				Category:       "namespace-security",
				Severity:       "low",
				Title:          "Namespace has minimal labels",
				Description:    fmt.Sprintf("Namespace %s has no organizational labels", ns.Name),
				Resource:       res,
				Recommendation: "Add labels for environment, team, and compliance tracking",
			})
		}
	}
	return findings, nil
}

func (l *Logger) checkSecretSecurity(ctx context.Context, clientset kubernetes.Interface) ([]Finding, error) {
	secrets, err := clientset.CoreV1().Secrets("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	var findings []Finding
	for _, secret := range secrets.Items {
		if secret.Type != corev1.SecretTypeOpaque {
			continue
		}
		res := map[string]string{"kind": "secret", "namespace": secret.Namespace, "name": secret.Name}
		if len(secret.Annotations) == 0 {
			findings = append(findings, Finding{
				Category:       "secret-security",
				Severity:       "low",
				Title:          "Secret without annotations",
				Description:    fmt.Sprintf("Secret %s/%s has no annotations for tracking or rotation", secret.Namespace, secret.Name),
				Resource:       res,
				Recommendation: "Add annotations for creation date, owner, and rotation schedule",
			})
		}
	}
	return findings, nil
}

func (l *Logger) checkRBACSecurity(ctx context.Context, clientset kubernetes.Interface) ([]Finding, error) {
	crbs, err := clientset.RbacV1().ClusterRoleBindings().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	var findings []Finding
	for _, crb := range crbs.Items {
		if crb.RoleRef.Name == "cluster-admin" {
			for _, subject := range crb.Subjects {
				if subject.Kind == "User" || subject.Kind == "ServiceAccount" {
					findings = append(findings, Finding{
						Category:       "rbac-security",
						Severity:       "high",
						Title:          "Cluster-admin binding detected",
						Description:    fmt.Sprintf("Subject %s/%s bound to cluster-admin via %s", subject.Kind, subject.Name, crb.Name),
						Resource:       map[string]string{"kind": "clusterrolebinding", "name": crb.Name},
						Recommendation: "Use more restrictive roles with specific permissions",
					})
				}
			}
		}
	}
	return findings, nil
}

func (l *Logger) checkNetworkSecurity(ctx context.Context, clientset kubernetes.Interface) ([]Finding, error) {
	namespaces, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	var findings []Finding
	for _, ns := range namespaces.Items {
		if ns.Name == "kube-system" || ns.Name == "kube-public" || ns.Name == "kube-node-lease" {
			continue
		}
		policies, err := clientset.NetworkingV1().NetworkPolicies(ns.Name).List(ctx, metav1.ListOptions{})
		if err != nil {
			continue
		}
		hasDefaultDeny := false
		for _, np := range policies.Items {
			for _, pt := range np.Spec.PolicyTypes {
				if pt == networkingv1.PolicyTypeIngress && len(np.Spec.Ingress) == 0 &&
					len(np.Spec.PodSelector.MatchLabels) == 0 && np.Spec.PodSelector.MatchExpressions == nil {
					hasDefaultDeny = true
				}
			}
		}
		if !hasDefaultDeny && len(policies.Items) == 0 {
			findings = append(findings, Finding{
				Category:       "network-security",
				Severity:       "medium",
				Title:          "No network policies",
				Description:    fmt.Sprintf("Namespace %s has no network policies defined", ns.Name),
				Resource:       map[string]string{"kind": "namespace", "name": ns.Name},
				Recommendation: "Implement default-deny network policies and whitelist required traffic",
			})
		}
	}
	return findings, nil
}

func buildSummary(findings []Finding) ReportSummary {
	var summary ReportSummary
	summary.TotalFindings = len(findings)
	for _, f := range findings {
		switch strings.ToLower(f.Severity) {
		case "critical":
			summary.CriticalFindings++
		case "high":
			summary.HighFindings++
		case "medium":
			summary.MediumFindings++
		case "low":
			summary.LowFindings++
		case "info":
			summary.InfoFindings++
		}
	}
	return summary
}

func calculateComplianceScore(findings []Finding) int {
	score := 100
	for _, f := range findings {
		switch strings.ToLower(f.Severity) {
		case "critical":
			score -= 10
		case "high":
			score -= 5
		case "medium":
			score -= 3
		case "low":
			score -= 1
		}
	}
	if score < 0 {
		score = 0
	}
	return score
}
