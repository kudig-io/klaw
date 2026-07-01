//go:build !linux || !cgo

package probe

import (
	"context"
	"errors"
)

type ProbeManager struct{}

func NewProbeManager(config *Config) (*ProbeManager, error) {
	return nil, errors.New("eBPF probes require Linux with cgo support")
}

func (pm *ProbeManager) Start(ctx context.Context) error {
	return errors.New("eBPF not available on this platform")
}

func (pm *ProbeManager) Stop() error {
	return nil
}

func (pm *ProbeManager) Events() <-chan Event {
	return nil
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
