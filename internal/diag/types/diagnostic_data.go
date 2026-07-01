// Package types defines core data structures for klaw
package types

import (
	"io"
	"time"

	"k8s.io/client-go/kubernetes"
)

// DataMode represents the data collection mode
type DataMode int

const (
	// ModeOffline indicates offline analysis of collected diagnostic data
	ModeOffline DataMode = iota + 1
	// ModeOnline indicates real-time diagnosis via K8s API
	ModeOnline
	// ModeHybrid indicates combination of offline and online modes
	ModeHybrid
)

// String returns the string representation of DataMode
func (m DataMode) String() string {
	switch m {
	case ModeOffline:
		return "offline"
	case ModeOnline:
		return "online"
	case ModeHybrid:
		return "hybrid"
	default:
		return "unknown"
	}
}

// DiagnosticData is the container for all diagnostic information
type DiagnosticData struct {
	// Mode indicates how the data was collected
	Mode DataMode `json:"mode" yaml:"mode"`

	// Timestamp when the data was collected
	Timestamp time.Time `json:"timestamp" yaml:"timestamp"`

	// DiagnosePath is the path to the diagnostic directory (offline mode)
	DiagnosePath string `json:"diagnose_path,omitempty" yaml:"diagnose_path,omitempty"`

	// NodeInfo contains node identification information
	NodeInfo NodeInfo `json:"node_info" yaml:"node_info"`

	// SystemMetrics contains system resource metrics
	SystemMetrics *SystemMetrics `json:"system_metrics,omitempty" yaml:"system_metrics,omitempty"`

	// RawFiles contains raw file contents from offline mode
	// key: relative file path, value: file content
	RawFiles map[string][]byte `json:"-" yaml:"-"`

	// LogStreams provides streaming access to log files
	// key: log file identifier, value: reader
	LogStreams map[string]io.Reader `json:"-" yaml:"-"`

	// K8sClient is the Kubernetes client (online mode only)
	K8sClient kubernetes.Interface `json:"-" yaml:"-"`

	// Namespace to focus on (online mode)
	Namespace string `json:"namespace,omitempty" yaml:"namespace,omitempty"`

	// NodeName to focus on (online mode)
	NodeName string `json:"node_name,omitempty" yaml:"node_name,omitempty"`

	// IstioInfo contains Istio service mesh information (online mode)
	IstioInfo *IstioInfo `json:"istio_info,omitempty" yaml:"istio_info,omitempty"`

	// LinkerdInfo contains Linkerd service mesh information (online mode)
	LinkerdInfo *LinkerdInfo `json:"linkerd_info,omitempty" yaml:"linkerd_info,omitempty"`
}

// NodeInfo contains basic node identification
type NodeInfo struct {
	// Hostname of the node
	Hostname string `json:"hostname" yaml:"hostname"`

	// KernelVersion of the node
	KernelVersion string `json:"kernel_version,omitempty" yaml:"kernel_version,omitempty"`

	// OSImage (e.g., "CentOS Linux 7")
	OSImage string `json:"os_image,omitempty" yaml:"os_image,omitempty"`

	// ContainerRuntime (e.g., "containerd://1.6.0")
	ContainerRuntime string `json:"container_runtime,omitempty" yaml:"container_runtime,omitempty"`

	// KubeletVersion
	KubeletVersion string `json:"kubelet_version,omitempty" yaml:"kubelet_version,omitempty"`
}

// SystemMetrics contains system resource metrics
type SystemMetrics struct {
	// CPU metrics
	CPUCores     int     `json:"cpu_cores" yaml:"cpu_cores"`
	LoadAvg1Min  float64 `json:"load_avg_1min" yaml:"load_avg_1min"`
	LoadAvg5Min  float64 `json:"load_avg_5min" yaml:"load_avg_5min"`
	LoadAvg15Min float64 `json:"load_avg_15min" yaml:"load_avg_15min"`

	// Memory metrics (in KB)
	MemTotal     int64 `json:"mem_total" yaml:"mem_total"`
	MemAvailable int64 `json:"mem_available" yaml:"mem_available"`
	MemFree      int64 `json:"mem_free" yaml:"mem_free"`
	SwapTotal    int64 `json:"swap_total" yaml:"swap_total"`
	SwapFree     int64 `json:"swap_free" yaml:"swap_free"`

	// Disk metrics
	DiskUsage []DiskUsage `json:"disk_usage,omitempty" yaml:"disk_usage,omitempty"`

	// Network metrics
	ConntrackCurrent int64 `json:"conntrack_current" yaml:"conntrack_current"`
	ConntrackMax     int64 `json:"conntrack_max" yaml:"conntrack_max"`
}

// DiskUsage contains disk usage information for a mount point
type DiskUsage struct {
	MountPoint   string  `json:"mount_point" yaml:"mount_point"`
	Device       string  `json:"device" yaml:"device"`
	TotalBytes   int64   `json:"total_bytes" yaml:"total_bytes"`
	UsedBytes    int64   `json:"used_bytes" yaml:"used_bytes"`
	UsedPercent  float64 `json:"used_percent" yaml:"used_percent"`
	InodeUsed    int64   `json:"inode_used" yaml:"inode_used"`
	InodeTotal   int64   `json:"inode_total" yaml:"inode_total"`
	InodePercent float64 `json:"inode_percent" yaml:"inode_percent"`
}

// ServiceStatus represents the status of a system service
type ServiceStatus struct {
	Name    string `json:"name" yaml:"name"`
	Status  string `json:"status" yaml:"status"` // running, stopped, failed, unknown
	Enabled bool   `json:"enabled" yaml:"enabled"`
}

// NewDiagnosticData creates a new DiagnosticData instance
func NewDiagnosticData(mode DataMode) *DiagnosticData {
	return &DiagnosticData{
		Mode:       mode,
		Timestamp:  time.Now(),
		RawFiles:   make(map[string][]byte),
		LogStreams: make(map[string]io.Reader),
	}
}

// GetRawFile returns the content of a raw file
func (d *DiagnosticData) GetRawFile(path string) ([]byte, bool) {
	if d.RawFiles == nil {
		return nil, false
	}
	content, ok := d.RawFiles[path]
	return content, ok
}

// SetRawFile sets the content of a raw file
func (d *DiagnosticData) SetRawFile(path string, content []byte) {
	if d.RawFiles == nil {
		d.RawFiles = make(map[string][]byte)
	}
	d.RawFiles[path] = content
}

// HasK8sClient returns true if K8s client is available
func (d *DiagnosticData) HasK8sClient() bool {
	return d.K8sClient != nil
}
