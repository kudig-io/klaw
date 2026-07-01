// Package collector defines the data collection layer interfaces
package collector

import (
	"context"

	"github.com/kudig-io/klaw/internal/diag/types"
)

// Collector is the interface for data collectors
type Collector interface {
	// Collect gathers diagnostic data from the source
	Collect(ctx context.Context, config *Config) (*types.DiagnosticData, error)

	// Mode returns the data collection mode
	Mode() types.DataMode

	// Name returns the collector name
	Name() string

	// Validate checks if the collector can operate with given config
	Validate(config *Config) error
}

// Config holds configuration for data collection
type Config struct {
	// DiagnosePath is the path to diagnostic directory (offline mode)
	DiagnosePath string `json:"diagnose_path,omitempty" yaml:"diagnose_path,omitempty"`

	// Kubeconfig is the path to kubeconfig file (online mode)
	Kubeconfig string `json:"kubeconfig,omitempty" yaml:"kubeconfig,omitempty"`

	// Context is the kubeconfig context to use
	Context string `json:"context,omitempty" yaml:"context,omitempty"`

	// Namespace to focus on
	Namespace string `json:"namespace,omitempty" yaml:"namespace,omitempty"`

	// NodeName to focus on
	NodeName string `json:"node_name,omitempty" yaml:"node_name,omitempty"`

	// AllNodes indicates whether to collect from all nodes
	AllNodes bool `json:"all_nodes,omitempty" yaml:"all_nodes,omitempty"`

	// MetricsURL is the Prometheus metrics endpoint (optional)
	MetricsURL string `json:"metrics_url,omitempty" yaml:"metrics_url,omitempty"`

	// Files lists specific files to read (offline mode)
	Files []string `json:"files,omitempty" yaml:"files,omitempty"`

	// Timeout for collection operations
	TimeoutSeconds int `json:"timeout_seconds,omitempty" yaml:"timeout_seconds,omitempty"`
}

// NewConfig creates a new Config with defaults
func NewConfig() *Config {
	return &Config{
		TimeoutSeconds: 60,
	}
}

// NewOfflineConfig creates a config for offline mode
func NewOfflineConfig(diagnosePath string) *Config {
	return &Config{
		DiagnosePath:   diagnosePath,
		TimeoutSeconds: 60,
	}
}

// NewOnlineConfig creates a config for online mode
func NewOnlineConfig(kubeconfig, nodeName string) *Config {
	return &Config{
		Kubeconfig:     kubeconfig,
		NodeName:       nodeName,
		TimeoutSeconds: 60,
	}
}

// Factory creates collectors based on mode.
type Factory struct {
	collectors map[types.DataMode]Collector
}

// NewCollectorFactory creates a new factory.
func NewCollectorFactory() *Factory {
	return &Factory{
		collectors: make(map[types.DataMode]Collector),
	}
}

// Register adds a collector for a specific mode
func (f *Factory) Register(collector Collector) {
	f.collectors[collector.Mode()] = collector
}

// Get returns a collector for the specified mode
func (f *Factory) Get(mode types.DataMode) (Collector, bool) {
	c, ok := f.collectors[mode]
	return c, ok
}

// DefaultFactory is the global collector factory
var DefaultFactory = NewCollectorFactory()

// RegisterCollector registers a collector with the default factory
func RegisterCollector(collector Collector) {
	DefaultFactory.Register(collector)
}

// GetCollector gets a collector from the default factory
func GetCollector(mode types.DataMode) (Collector, bool) {
	return DefaultFactory.Get(mode)
}
