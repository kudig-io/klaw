package offline

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/kudig-io/klaw/internal/diag/collector"
)

func TestNewCollector(t *testing.T) {
	c := NewCollector()
	if c == nil {
		t.Fatal("NewCollector() returned nil")
	}
	if c.Name() != "offline" {
		t.Errorf("Name() = %v, want %v", c.Name(), "offline")
	}
}

func TestCollectorMode(t *testing.T) {
	c := NewCollector()
	if c.Mode().String() != "offline" {
		t.Errorf("Mode() = %v, want offline", c.Mode())
	}
}

func TestCollectorValidate_EmptyPath(t *testing.T) {
	c := NewCollector()
	config := &collector.Config{
		DiagnosePath: "",
	}

	err := c.Validate(config)
	if err == nil {
		t.Error("Expected error for empty path")
	}
}

func TestCollectorValidate_NonExistent(t *testing.T) {
	c := NewCollector()
	config := &collector.Config{
		DiagnosePath: "/nonexistent/path/12345",
	}

	err := c.Validate(config)
	if err == nil {
		t.Error("Expected error for non-existent path")
	}
}

func TestCollectorValidate_NotDirectory(t *testing.T) {
	c := NewCollector()

	// Create a temp file
	tmpfile, err := os.CreateTemp("", "test")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	config := &collector.Config{
		DiagnosePath: tmpfile.Name(),
	}

	err = c.Validate(config)
	if err == nil {
		t.Error("Expected error for file (not directory)")
	}
}

func TestCollectorValidate_Valid(t *testing.T) {
	c := NewCollector()

	// Create a temp directory
	tmpdir, err := os.MkdirTemp("", "kudig-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpdir)

	config := &collector.Config{
		DiagnosePath: tmpdir,
	}

	err = c.Validate(config)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestCollectorCollect_InvalidPath(t *testing.T) {
	c := NewCollector()
	ctx := context.Background()
	config := &collector.Config{
		DiagnosePath: "/nonexistent",
	}

	_, err := c.Collect(ctx, config)
	if err == nil {
		t.Error("Expected error for invalid path")
	}
}

func TestCollectorCollect_Valid(t *testing.T) {
	c := NewCollector()
	ctx := context.Background()

	// Create a temp directory with some test files
	tmpdir, err := os.MkdirTemp("", "kudig-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpdir)

	// Create test files
	testFiles := []string{
		"system_info",
		"system_status",
		"service_status",
		"memory_info",
	}
	for _, file := range testFiles {
		path := filepath.Join(tmpdir, file)
		if err := os.WriteFile(path, []byte("test content"), 0600); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	config := &collector.Config{
		DiagnosePath: tmpdir,
	}

	data, err := c.Collect(ctx, config)
	if err != nil {
		t.Fatalf("Collect() error = %v", err)
	}
	if data == nil {
		t.Fatal("Collect() returned nil data")
	}
	if data.DiagnosePath != tmpdir {
		t.Errorf("DiagnosePath = %v, want %v", data.DiagnosePath, tmpdir)
	}
}

func TestCollectorCollect_ContextCancel(t *testing.T) {
	c := NewCollector()

	// Create a temp directory
	tmpdir, err := os.MkdirTemp("", "kudig-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpdir)

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	config := &collector.Config{
		DiagnosePath: tmpdir,
	}

	_, err = c.Collect(ctx, config)
	if err == nil {
		t.Error("Expected error for cancelled context")
	}
}
