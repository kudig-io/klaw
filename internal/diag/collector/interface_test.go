package collector

import (
	"context"
	"testing"

	"github.com/kudig-io/klaw/internal/diag/types"
)

func TestNewConfig(t *testing.T) {
	config := NewConfig()

	if config.TimeoutSeconds != 60 {
		t.Errorf("TimeoutSeconds = %v, want %v", config.TimeoutSeconds, 60)
	}
}

func TestNewOfflineConfig(t *testing.T) {
	config := NewOfflineConfig("/tmp/test")

	if config.DiagnosePath != "/tmp/test" {
		t.Errorf("DiagnosePath = %v, want %v", config.DiagnosePath, "/tmp/test")
	}
	if config.TimeoutSeconds != 60 {
		t.Errorf("TimeoutSeconds = %v, want %v", config.TimeoutSeconds, 60)
	}
}

func TestNewOnlineConfig(t *testing.T) {
	config := NewOnlineConfig("/tmp/kubeconfig", "test-node")

	if config.Kubeconfig != "/tmp/kubeconfig" {
		t.Errorf("Kubeconfig = %v, want %v", config.Kubeconfig, "/tmp/kubeconfig")
	}
	if config.NodeName != "test-node" {
		t.Errorf("NodeName = %v, want %v", config.NodeName, "test-node")
	}
	if config.TimeoutSeconds != 60 {
		t.Errorf("TimeoutSeconds = %v, want %v", config.TimeoutSeconds, 60)
	}
}

func TestCollectorFactory(t *testing.T) {
	factory := NewCollectorFactory()

	// Create mock collector
	mock := &mockCollector{mode: types.ModeOffline}

	// Register
	factory.Register(mock)

	// Get
	got, ok := factory.Get(types.ModeOffline)
	if !ok {
		t.Error("Expected to find registered collector")
	}
	if got.Name() != "mock" {
		t.Errorf("Name = %v, want %v", got.Name(), "mock")
	}

	// Get nonexistent
	_, ok = factory.Get(types.ModeOnline)
	if ok {
		t.Error("Should not find nonexistent collector")
	}
}

func TestRegisterCollector(t *testing.T) {
	// Create new factory for test
	DefaultFactory = NewCollectorFactory()

	mock := &mockCollector{mode: types.ModeOffline}
	RegisterCollector(mock)

	col, ok := GetCollector(types.ModeOffline)
	if !ok {
		t.Error("Expected to find registered collector")
	}
	if col.Name() != "mock" {
		t.Errorf("Name = %v, want %v", col.Name(), "mock")
	}
}

// mockCollector is a test double
type mockCollector struct {
	mode types.DataMode
}

func (m *mockCollector) Name() string         { return "mock" }
func (m *mockCollector) Mode() types.DataMode { return m.mode }
func (m *mockCollector) Validate(_ *Config) error {
	return nil
}
func (m *mockCollector) Collect(_ context.Context, _ *Config) (*types.DiagnosticData, error) {
	return types.NewDiagnosticData(types.ModeOffline), nil
}
