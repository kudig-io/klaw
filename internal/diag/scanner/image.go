// Package scanner provides image security scanning capabilities
package scanner

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// ImageScanner provides image vulnerability scanning
type ImageScanner struct {
	// Scanner type: trivy, snyk, etc.
	ScannerType string

	// Severity levels to report
	SeverityLevels []string
}

// NewImageScanner creates a new image scanner
func NewImageScanner() *ImageScanner {
	return &ImageScanner{
		ScannerType:    "trivy",
		SeverityLevels: []string{"CRITICAL", "HIGH", "MEDIUM"},
	}
}

// Vulnerability represents a found vulnerability
type Vulnerability struct {
	ID          string   `json:"id"`
	Severity    string   `json:"severity"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Package     string   `json:"package"`
	Version     string   `json:"version"`
	FixedIn     string   `json:"fixed_in,omitempty"`
	References  []string `json:"references,omitempty"`
}

// ScanResult contains scan results for an image
type ScanResult struct {
	Image           string          `json:"image"`
	Scanner         string          `json:"scanner"`
	ScanTime        string          `json:"scan_time"`
	Vulnerabilities []Vulnerability `json:"vulnerabilities"`
	Summary         ScanSummary     `json:"summary"`
}

// ScanSummary provides vulnerability counts by severity
type ScanSummary struct {
	Critical int `json:"critical"`
	High     int `json:"high"`
	Medium   int `json:"medium"`
	Low      int `json:"low"`
	Total    int `json:"total"`
}

// ScanImage scans a container image for vulnerabilities
func (s *ImageScanner) ScanImage(ctx context.Context, image string) (*ScanResult, error) {
	// Check if trivy is available
	if !s.IsAvailable() {
		return nil, fmt.Errorf("scanner not available: %s", s.ScannerType)
	}

	// Run trivy scan
	cmd := exec.CommandContext(ctx, "trivy", "image",
		"--format", "json",
		"--severity", strings.Join(s.SeverityLevels, ","),
		image)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("scan failed: %w", err)
	}

	// Parse results
	return s.parseTrivyOutput(image, output)
}

// IsAvailable checks if the scanner binary is available
func (s *ImageScanner) IsAvailable() bool {
	_, err := exec.LookPath(s.ScannerType)
	return err == nil
}

// parseTrivyOutput parses trivy JSON output
func (s *ImageScanner) parseTrivyOutput(image string, output []byte) (*ScanResult, error) {
	result := &ScanResult{
		Image:           image,
		Scanner:         s.ScannerType,
		ScanTime:        "now",
		Vulnerabilities: make([]Vulnerability, 0),
		Summary:         ScanSummary{},
	}

	// Parse trivy output
	var trivyResult struct {
		Results []struct {
			Vulnerabilities []struct {
				VulnerabilityID string `json:"VulnerabilityID"`
				Severity        string `json:"Severity"`
				Title           string `json:"Title"`
				Description     string `json:"Description"`
				PkgName         string `json:"PkgName"`
				InstalledVersion string `json:"InstalledVersion"`
				FixedVersion    string `json:"FixedVersion"`
				References      []string `json:"References"`
			} `json:"Vulnerabilities"`
		} `json:"Results"`
	}

	if err := json.Unmarshal(output, &trivyResult); err != nil {
		// If parsing fails, return empty result
		return result, nil
	}

	// Convert to our format
	for _, r := range trivyResult.Results {
		for _, v := range r.Vulnerabilities {
			vuln := Vulnerability{
				ID:          v.VulnerabilityID,
				Severity:    v.Severity,
				Title:       v.Title,
				Description: v.Description,
				Package:     v.PkgName,
				Version:     v.InstalledVersion,
				FixedIn:     v.FixedVersion,
				References:  v.References,
			}
			result.Vulnerabilities = append(result.Vulnerabilities, vuln)

			// Update summary
			switch v.Severity {
			case "CRITICAL":
				result.Summary.Critical++
			case "HIGH":
				result.Summary.High++
			case "MEDIUM":
				result.Summary.Medium++
			case "LOW":
				result.Summary.Low++
			}
			result.Summary.Total++
		}
	}

	return result, nil
}

// FormatResult formats scan result for display
func FormatResult(result *ScanResult) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("\n🔒 镜像安全扫描: %s\n", result.Image))
	sb.WriteString("=" + strings.Repeat("=", 50) + "\n")

	// Summary
	sb.WriteString(fmt.Sprintf("\n漏洞统计:\n"))
	sb.WriteString(fmt.Sprintf("  🔴 致命: %d\n", result.Summary.Critical))
	sb.WriteString(fmt.Sprintf("  🟠 高危: %d\n", result.Summary.High))
	sb.WriteString(fmt.Sprintf("  🟡 中危: %d\n", result.Summary.Medium))
	sb.WriteString(fmt.Sprintf("  🟢 低危: %d\n", result.Summary.Low))
	sb.WriteString(fmt.Sprintf("  总计:   %d\n", result.Summary.Total))

	// Top vulnerabilities
	if len(result.Vulnerabilities) > 0 {
		sb.WriteString("\n关键漏洞:\n")
		for i, v := range result.Vulnerabilities {
			if i >= 10 {
				sb.WriteString(fmt.Sprintf("\n... 还有 %d 个漏洞\n", len(result.Vulnerabilities)-10))
				break
			}
			sb.WriteString(fmt.Sprintf("\n  [%s] %s\n", v.Severity, v.ID))
			sb.WriteString(fmt.Sprintf("    %s\n", v.Title))
			sb.WriteString(fmt.Sprintf("    包: %s %s", v.Package, v.Version))
			if v.FixedIn != "" {
				sb.WriteString(fmt.Sprintf(" (修复版本: %s)", v.FixedIn))
			}
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// MockScanResult returns a mock scan result for testing
func MockScanResult(image string) *ScanResult {
	return &ScanResult{
		Image:    image,
		Scanner:  "trivy",
		ScanTime: "now",
		Vulnerabilities: []Vulnerability{
			{
				ID:       "CVE-2023-1234",
				Severity: "HIGH",
				Title:    "Example vulnerability",
				Package:  "example-package",
				Version:  "1.0.0",
				FixedIn:  "1.0.1",
			},
		},
		Summary: ScanSummary{
			High:  1,
			Total: 1,
		},
	}
}
