package scanner

import (
	"testing"
)

func TestNewImageScanner(t *testing.T) {
	s := NewImageScanner()
	if s == nil {
		t.Fatal("Expected non-nil scanner")
	}
	if s.ScannerType != "trivy" {
		t.Errorf("Expected scanner type 'trivy', got '%s'", s.ScannerType)
	}
	if len(s.SeverityLevels) != 3 {
		t.Errorf("Expected 3 severity levels, got %d", len(s.SeverityLevels))
	}
}

func TestImageScanner_isScannerAvailable(t *testing.T) {
	s := NewImageScanner()
	// trivy is likely not available in test environment
	available := s.IsAvailable()
	// Just verify it doesn't panic - result depends on environment
	_ = available
}

func TestImageScanner_ParseTrivyOutput_Empty(t *testing.T) {
	s := NewImageScanner()
	result, err := s.parseTrivyOutput("test:latest", []byte(`{}`))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result.Image != "test:latest" {
		t.Errorf("Expected image 'test:latest', got '%s'", result.Image)
	}
	if result.Summary.Total != 0 {
		t.Errorf("Expected 0 total vulnerabilities, got %d", result.Summary.Total)
	}
}

func TestImageScanner_ParseTrivyOutput_WithVulns(t *testing.T) {
	s := NewImageScanner()
	input := `{
		"Results": [
			{
				"Vulnerabilities": [
					{
						"VulnerabilityID": "CVE-2023-1234",
						"Severity": "HIGH",
						"Title": "Test vulnerability",
						"Description": "A test vuln",
						"PkgName": "test-pkg",
						"InstalledVersion": "1.0.0",
						"FixedVersion": "1.0.1",
						"References": ["https://example.com"]
					},
					{
						"VulnerabilityID": "CVE-2023-5678",
						"Severity": "CRITICAL",
						"Title": "Critical test vuln",
						"Description": "Critical",
						"PkgName": "critical-pkg",
						"InstalledVersion": "2.0.0"
					}
				]
			}
		]
	}`
	result, err := s.parseTrivyOutput("nginx:latest", []byte(input))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(result.Vulnerabilities) != 2 {
		t.Fatalf("Expected 2 vulnerabilities, got %d", len(result.Vulnerabilities))
	}
	if result.Summary.Critical != 1 {
		t.Errorf("Expected 1 critical, got %d", result.Summary.Critical)
	}
	if result.Summary.High != 1 {
		t.Errorf("Expected 1 high, got %d", result.Summary.High)
	}
	if result.Summary.Total != 2 {
		t.Errorf("Expected total 2, got %d", result.Summary.Total)
	}
	vuln := result.Vulnerabilities[0]
	if vuln.ID != "CVE-2023-1234" {
		t.Errorf("Expected ID 'CVE-2023-1234', got '%s'", vuln.ID)
	}
	if vuln.Package != "test-pkg" {
		t.Errorf("Expected package 'test-pkg', got '%s'", vuln.Package)
	}
}

func TestImageScanner_ParseTrivyOutput_InvalidJSON(t *testing.T) {
	s := NewImageScanner()
	result, err := s.parseTrivyOutput("test:latest", []byte(`invalid json`))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result.Summary.Total != 0 {
		t.Errorf("Expected 0 vulnerabilities for invalid JSON, got %d", result.Summary.Total)
	}
}

func TestMockScanResult(t *testing.T) {
	result := MockScanResult("test:latest")
	if result.Image != "test:latest" {
		t.Errorf("Expected image 'test:latest', got '%s'", result.Image)
	}
	if len(result.Vulnerabilities) != 1 {
		t.Errorf("Expected 1 vulnerability in mock, got %d", len(result.Vulnerabilities))
	}
	if result.Summary.High != 1 {
		t.Errorf("Expected 1 high vulnerability, got %d", result.Summary.High)
	}
}

func TestFormatResult(t *testing.T) {
	result := &ScanResult{
		Image:   "nginx:latest",
		Scanner: "trivy",
		Vulnerabilities: []Vulnerability{
			{ID: "CVE-2023-0001", Severity: "CRITICAL", Title: "Test", Package: "pkg1", Version: "1.0"},
		},
		Summary: ScanSummary{
			Critical: 1,
			Total:    1,
		},
	}
	formatted := FormatResult(result)
	if formatted == "" {
		t.Error("Expected non-empty formatted result")
	}
}
