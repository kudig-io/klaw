// Package reporter provides SARIF output format for security scanning integration
package reporter

import (
	"encoding/json"
	"fmt"

	"github.com/kudig-io/klaw/internal/diag/types"
)

// SarifReporter generates reports in SARIF-like format for security scanning integration
type SarifReporter struct{}

// NewSarifReporter creates a new SARIF reporter
func NewSarifReporter() *SarifReporter {
	return &SarifReporter{}
}

// Format returns the reporter format
func (r *SarifReporter) Format() string {
	return "sarif"
}

// ContentType returns the MIME type
func (r *SarifReporter) ContentType() string {
	return "application/sarif+json"
}

// FileExtension returns the file extension
func (r *SarifReporter) FileExtension() string {
	return ".sarif"
}

// Generate creates a SARIF-like report
func (r *SarifReporter) Generate(issues []types.Issue, metadata *ReportMetadata) ([]byte, error) {
	// Filter security issues
	securityIssues := FilterSecurityIssues(issues)

	// Build SARIF-like structure
	report := map[string]interface{}{
		"$schema": "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
		"version": "2.1.0",
		"runs": []map[string]interface{}{
			{
				"tool": map[string]interface{}{
					"driver": map[string]interface{}{
						"name":            "klaw",
						"informationUri":  "https://github.com/kudig-io/klaw",
						"fullName":        "klaw - Kubernetes All-in-One Operations Platform",
						"version":         "2.0.0",
						"organization":    "kudig-io",
						"rules":           buildSarifRules(securityIssues),
					},
				},
				"results": buildSarifResults(securityIssues),
				"invocations": []map[string]interface{}{
					{
						"executionSuccessful": true,
						"machine":             metadata.Hostname,
					},
				},
			},
		},
	}

	// Marshal to JSON
	output, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal SARIF report: %w", err)
	}

	return output, nil
}

// buildSarifRules builds SARIF rules from issues
func buildSarifRules(issues []types.Issue) []map[string]interface{} {
	rulesMap := make(map[string]map[string]interface{})

	for _, issue := range issues {
		if _, exists := rulesMap[issue.ENName]; exists {
			continue
		}

		rule := map[string]interface{}{
			"id":               issue.ENName,
			"name":             issue.CNName,
			"shortDescription": map[string]string{"text": issue.CNName},
			"fullDescription":  map[string]string{"text": issue.Details},
			"defaultConfiguration": map[string]interface{}{
				"level": severityToSarifLevel(issue.Severity),
			},
		}

		if issue.Remediation != nil {
			rule["help"] = map[string]interface{}{
				"text":     issue.Remediation.Suggestion,
				"markdown": fmt.Sprintf("**修复建议**\n\n%s", issue.Remediation.Suggestion),
			}
		}

		rulesMap[issue.ENName] = rule
	}

	rules := make([]map[string]interface{}, 0, len(rulesMap))
	for _, rule := range rulesMap {
		rules = append(rules, rule)
	}
	return rules
}

// buildSarifResults builds SARIF results from issues
func buildSarifResults(issues []types.Issue) []map[string]interface{} {
	results := make([]map[string]interface{}, 0, len(issues))

	for _, issue := range issues {
		result := map[string]interface{}{
			"ruleId":  issue.ENName,
			"level":   severityToSarifLevel(issue.Severity),
			"message": map[string]string{"text": issue.Details},
			"locations": []map[string]interface{}{
				{
					"physicalLocation": map[string]interface{}{
						"artifactLocation": map[string]string{
							"uri": issue.Location,
						},
					},
					"logicalLocations": []map[string]string{
						{
							"fullyQualifiedName": issue.Location,
							"name":               issue.CNName,
						},
					},
				},
			},
		}

		// Add fix if remediation has command
		if issue.Remediation != nil && issue.Remediation.Command != "" {
			result["fixes"] = []map[string]interface{}{
				{
					"description": map[string]string{"text": issue.Remediation.Suggestion},
					"artifactChanges": []map[string]interface{}{
						{
							"artifactLocation": map[string]string{"uri": issue.Location},
							"replacements": []map[string]interface{}{
								{
									"deletedRegion": map[string]interface{}{
										"snippet": map[string]string{
											"text": fmt.Sprintf("执行命令: %s", issue.Remediation.Command),
										},
									},
								},
							},
						},
					},
				},
			}
		}

		results = append(results, result)
	}

	return results
}

// isSecurityIssue checks if an issue is security-related
func isSecurityIssue(issue types.Issue) bool {
	securityAnalyzers := map[string]bool{
		"security.cis":        true,
		"security.rbac":       true,
		"kubernetes.cis":      true,
		"kubernetes.rbac":     true,
		"servicemesh.istio":   true,
		"servicemesh.linkerd": true,
	}

	if securityAnalyzers[issue.AnalyzerName] {
		return true
	}

	// Check for security-related keywords in code
	securityKeywords := []string{
		"CIS", "RBAC", "SECURITY", "PRIVILEGE", "ROOT",
		"PASSWORD", "TOKEN", "CERTIFICATE", "TLS", "MTLS",
		"AUTHORIZATION", "AUTHENTICATION", "AUDIT",
	}

	for _, keyword := range securityKeywords {
		if containsString(issue.ENName, keyword) {
			return true
		}
	}

	return false
}

// severityToSarifLevel converts klaw severity to SARIF level
func severityToSarifLevel(severity types.Severity) string {
	switch severity {
	case types.SeverityCritical:
		return "error"
	case types.SeverityWarning:
		return "warning"
	case types.SeverityInfo:
		return "note"
	default:
		return "none"
	}
}

// containsString checks if string contains substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// SarifReportGenerator provides high-level SARIF generation
type SarifReportGenerator struct {
	Reporter *SarifReporter
}

// NewSarifReportGenerator creates a new SARIF report generator
func NewSarifReportGenerator() *SarifReportGenerator {
	return &SarifReportGenerator{
		Reporter: NewSarifReporter(),
	}
}

// GenerateFromIssues generates a SARIF report from issues
func (g *SarifReportGenerator) GenerateFromIssues(issues []types.Issue, target string) ([]byte, error) {
	metadata := &ReportMetadata{
		Hostname: target,
		Mode:     "sarif",
		Engine:   "klaw-security",
	}
	return g.Reporter.Generate(issues, metadata)
}

// FilterSecurityIssues filters only security-related issues
func FilterSecurityIssues(issues []types.Issue) []types.Issue {
	var securityIssues []types.Issue
	for _, issue := range issues {
		if isSecurityIssue(issue) {
			securityIssues = append(securityIssues, issue)
		}
	}
	return securityIssues
}

func init() {
	RegisterReporter(NewSarifReporter())
}
