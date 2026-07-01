//go:build linux && cgo

package probe

import (
	"context"
	"fmt"
	"time"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/perf"
	"github.com/cilium/ebpf/rlimit"
	"k8s.io/klog/v2"
)

type ProbeManager struct {
	collection *ebpf.Collection
	links      []link.Link
	events     chan Event
	stopCh     chan struct{}
	readers    []*perf.Reader
}

func NewProbeManager(config *Config) (*ProbeManager, error) {
	if config == nil {
		config = DefaultConfig()
	}
	if err := rlimit.RemoveMemlock(); err != nil {
		return nil, fmt.Errorf("failed to remove memlock limit: %w", err)
	}
	pm := &ProbeManager{
		events: make(chan Event, 10000),
		stopCh: make(chan struct{}),
	}
	return pm, nil
}

func (pm *ProbeManager) Start(ctx context.Context) error {
	klog.InfoS("eBPF diagnostics is not yet implemented; probes will not collect data")
	go pm.run(ctx)
	return nil
}

func (pm *ProbeManager) Stop() error {
	klog.InfoS("Stopping eBPF probes")
	close(pm.stopCh)
	for _, reader := range pm.readers {
		reader.Close()
	}
	for _, l := range pm.links {
		l.Close()
	}
	if pm.collection != nil {
		pm.collection.Close()
	}
	close(pm.events)
	return nil
}

func (pm *ProbeManager) run(ctx context.Context) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-pm.stopCh:
			return
		case <-ticker.C:
		}
	}
}

func (pm *ProbeManager) Events() <-chan Event {
	return pm.events
}

func (pm *ProbeManager) GetTCPStats() (*TCPStats, error) {
	return &TCPStats{}, nil
}

func (pm *ProbeManager) GetDNSStats() (*DNSStats, error) {
	return &DNSStats{}, nil
}

func (pm *ProbeManager) GetFileIOStats() (*FileIOStats, error) {
	return &FileIOStats{}, nil
}
