package rules

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewLoader(t *testing.T) {
	l := NewLoader()
	if l == nil {
		t.Fatal("NewLoader() returned nil")
	}
}

func TestLoaderLoadFile(t *testing.T) {
	l := NewLoader()

	// Create temp directory and file
	tmpDir, err := os.MkdirTemp("", "rules-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test rule file
	ruleContent := `
version: "1.0"
name: "test-rules"
rules:
  - id: TEST_RULE
    name: "Test Rule"
    category: "test"
    severity: "warning"
    enabled: true
    condition:
      type: "file_contains"
      file: "test.log"
      pattern: "error"
`
	ruleFile := filepath.Join(tmpDir, "test.yaml")
	if err := os.WriteFile(ruleFile, []byte(ruleContent), 0600); err != nil {
		t.Fatalf("Failed to write rule file: %v", err)
	}

	// Load the file
	if err := l.LoadFile(ruleFile); err != nil {
		t.Fatalf("LoadFile() error = %v", err)
	}

	// Check rules were loaded
	rules := l.GetAllRules()
	if len(rules) != 1 {
		t.Errorf("Expected 1 rule, got %d", len(rules))
	}
	if rules[0].ID != "TEST_RULE" {
		t.Errorf("Rule ID = %v, want %v", rules[0].ID, "TEST_RULE")
	}
}

func TestLoaderLoadFileNotFound(t *testing.T) {
	l := NewLoader()
	err := l.LoadFile("/nonexistent/file.yaml")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestLoaderLoadFileInvalidYAML(t *testing.T) {
	l := NewLoader()

	// Create temp file with invalid YAML
	tmpFile, err := os.CreateTemp("", "invalid-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString("invalid: yaml: content: ["); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	err = l.LoadFile(tmpFile.Name())
	if err == nil {
		t.Error("Expected error for invalid YAML")
	}
}

func TestLoaderLoadFileMissingID(t *testing.T) {
	l := NewLoader()

	// Create temp file with rule missing ID
	tmpFile, err := os.CreateTemp("", "invalid-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	ruleContent := `
version: "1.0"
name: "test-rules"
rules:
  - name: "Test Rule"
    category: "test"
    severity: "warning"
`
	if err := os.WriteFile(tmpFile.Name(), []byte(ruleContent), 0600); err != nil {
		t.Fatalf("Failed to write rule file: %v", err)
	}

	err = l.LoadFile(tmpFile.Name())
	if err == nil {
		t.Error("Expected error for rule missing ID")
	}
}

func TestLoaderLoadDir(t *testing.T) {
	l := NewLoader()

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "rules-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create multiple rule files
	for i, name := range []string{"rules1.yaml", "rules2.yml"} {
		ruleContent := `
version: "1.0"
name: "test-rules"
rules:
  - id: TEST_RULE_` + string(rune('0'+i)) + `
    name: "Test Rule"
    category: "test"
    severity: "warning"
    enabled: true
    condition:
      type: "file_contains"
      file: "test.log"
      pattern: "error"
`
		ruleFile := filepath.Join(tmpDir, name)
		if err := os.WriteFile(ruleFile, []byte(ruleContent), 0600); err != nil {
			t.Fatalf("Failed to write rule file: %v", err)
		}
	}

	// Create a non-yaml file (should be ignored)
	otherFile := filepath.Join(tmpDir, "readme.txt")
	if err := os.WriteFile(otherFile, []byte("readme"), 0600); err != nil {
		t.Fatalf("Failed to write other file: %v", err)
	}

	// Load directory
	if err := l.LoadDir(tmpDir); err != nil {
		t.Fatalf("LoadDir() error = %v", err)
	}

	// Check rules were loaded
	rules := l.GetAllRules()
	if len(rules) != 2 {
		t.Errorf("Expected 2 rules, got %d", len(rules))
	}
}

func TestLoaderLoadBuiltin(t *testing.T) {
	l := NewLoader()
	if err := l.LoadBuiltin(); err != nil {
		t.Fatalf("LoadBuiltin() error = %v", err)
	}

	rules := l.GetAllRules()
	if len(rules) == 0 {
		t.Error("Expected built-in rules to be loaded")
	}
}

func TestLoaderGetRule(t *testing.T) {
	l := NewLoader()

	// Create and load a rule
	tmpFile, err := os.CreateTemp("", "rules-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	ruleContent := `
version: "1.0"
name: "test-rules"
rules:
  - id: MY_RULE
    name: "My Rule"
    category: "test"
    severity: "warning"
    enabled: true
    condition:
      type: "file_contains"
      file: "test.log"
      pattern: "error"
`
	if err := os.WriteFile(tmpFile.Name(), []byte(ruleContent), 0600); err != nil {
		t.Fatalf("Failed to write rule file: %v", err)
	}

	if err := l.LoadFile(tmpFile.Name()); err != nil {
		t.Fatalf("LoadFile() error = %v", err)
	}

	// Get existing rule
	rule := l.GetRuleByID("MY_RULE")
	if rule == nil {
		t.Fatal("Expected to find MY_RULE")
	}
	if rule.ID != "MY_RULE" {
		t.Errorf("Rule ID = %v, want %v", rule.ID, "MY_RULE")
	}

	// Get nonexistent rule
	nonexistent := l.GetRuleByID("NONEXISTENT")
	if nonexistent != nil {
		t.Error("Should not find nonexistent rule")
	}
}

func TestLoaderDefaultFields(t *testing.T) {
	l := NewLoader()

	// Create temp file with minimal rule
	tmpFile, err := os.CreateTemp("", "rules-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	ruleContent := `
version: "1.0"
name: "test-rules"
rules:
  - id: MINIMAL_RULE
    condition:
      type: "file_contains"
      file: "test.log"
      pattern: "error"
`
	if err := os.WriteFile(tmpFile.Name(), []byte(ruleContent), 0600); err != nil {
		t.Fatalf("Failed to write rule file: %v", err)
	}

	if err := l.LoadFile(tmpFile.Name()); err != nil {
		t.Fatalf("LoadFile() error = %v", err)
	}

	rule := l.GetRuleByID("MINIMAL_RULE")
	if rule == nil {
		t.Fatal("Expected to find MINIMAL_RULE")
	}

	// Check defaults
	if rule.Name != "MINIMAL_RULE" {
		t.Errorf("Name = %v, want %v", rule.Name, "MINIMAL_RULE")
	}
	if rule.Severity != "info" {
		t.Errorf("Severity = %v, want %v", rule.Severity, "info")
	}
	if !rule.Enabled {
		t.Error("Enabled should be true by default")
	}
}
