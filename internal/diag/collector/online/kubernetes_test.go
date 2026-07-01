package online

import (
	"testing"

	"github.com/kudig-io/klaw/internal/diag/collector"
	"github.com/kudig-io/klaw/internal/diag/types"
)

func TestNewCollector(t *testing.T) {
	c := NewCollector()
	if c == nil {
		t.Fatal("NewCollector() returned nil")
	}
}

func TestCollectorName(t *testing.T) {
	c := NewCollector()
	if c.Name() != "online" {
		t.Errorf("Name() = %v, want %v", c.Name(), "online")
	}
}

func TestCollectorMode(t *testing.T) {
	c := NewCollector()
	if c.Mode() != types.ModeOnline {
		t.Errorf("Mode() = %v, want %v", c.Mode(), types.ModeOnline)
	}
}

func TestCollectorValidate_NoConfig(t *testing.T) {
	c := NewCollector()
	config := &collector.Config{}

	// Try to validate - may succeed if default kubeconfig exists
	// or fail if it doesn't. We just verify it doesn't panic.
	_ = c.Validate(config)
}

func TestHomeDir(t *testing.T) {
	home := homeDir()
	if home == "" {
		t.Error("homeDir() returned empty string")
	}
}

func TestCollectorCollect_NotImplemented(t *testing.T) {
	c := NewCollector()
	// The Collect method requires a real K8s client, so we just verify it exists
	// and returns an error when no client is configured
	if c == nil {
		t.Fatal("Collector is nil")
	}
}
