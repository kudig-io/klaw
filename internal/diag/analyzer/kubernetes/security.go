// Package kubernetes provides Kubernetes component analyzers
package kubernetes

import (
	"context"
	"strings"

	"github.com/kudig-io/klaw/internal/diag/analyzer"
	"github.com/kudig-io/klaw/internal/diag/types"
)

// CertificateAnalyzer checks for certificate expiration
type CertificateAnalyzer struct {
	*analyzer.BaseAnalyzer
}

// NewCertificateAnalyzer creates a new certificate analyzer
func NewCertificateAnalyzer() *CertificateAnalyzer {
	return &CertificateAnalyzer{
		BaseAnalyzer: analyzer.NewBaseAnalyzer(
			"kubernetes.certificate",
			"检查证书状态",
			"kubernetes",
			[]types.DataMode{types.ModeOffline, types.ModeOnline},
		),
	}
}

// Analyze performs certificate analysis
func (a *CertificateAnalyzer) Analyze(ctx context.Context, data *types.DiagnosticData) ([]types.Issue, error) {
	var issues []types.Issue

	kubeletLog, ok := data.GetRawFile("logs/kubelet.log")
	if !ok {
		kubeletLog, ok = data.GetRawFile("daemon_status/kubelet_status")
		if !ok {
			return issues, nil
		}
	}

	content := string(kubeletLog)

	if strings.Contains(content, "certificate has expired") {
		issue := types.NewIssue(
			types.SeverityCritical,
			"证书已过期",
			"CERTIFICATE_EXPIRED",
			"Kubelet证书已过期",
			"logs/kubelet.log",
		).WithRemediation("更新证书: kubeadm certs renew all; systemctl restart kubelet")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	} else if strings.Contains(content, "certificate will expire") {
		issue := types.NewIssue(
			types.SeverityWarning,
			"证书即将过期",
			"CERTIFICATE_EXPIRING",
			"Kubelet证书即将过期",
			"logs/kubelet.log",
		).WithRemediation("尽快更新证书: kubeadm certs renew all")
		issue.AnalyzerName = a.Name()
		issues = append(issues, *issue)
	}

	return issues, nil
}
