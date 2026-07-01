//go:build linux && cgo

package probe

import (
	"context"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config == nil {
		t.Fatal("expected config to not be nil")
	}

	if !config.EnableTCP {
		t.Error("expected EnableTCP to be true")
	}

	if !config.EnableDNS {
		t.Error("expected EnableDNS to be true")
	}

	if config.EnableSyscall {
		t.Error("expected EnableSyscall to be false")
	}

	if config.Duration != 30*time.Second {
		t.Errorf("expected Duration to be 30s, got %v", config.Duration)
	}
}

func TestNewProbeManager(t *testing.T) {
	config := DefaultConfig()
	pm, err := NewProbeManager(config)

	if err != nil {
		// eBPF may not be available in test environment
		t.Skipf("eBPF not available: %v", err)
	}

	if pm == nil {
		t.Fatal("expected probe manager to not be nil")
	}
}

func TestNewProbeManager_NilConfig(t *testing.T) {
	pm, err := NewProbeManager(nil)

	if err != nil {
		t.Skipf("eBPF not available: %v", err)
	}

	if pm == nil {
		t.Fatal("expected probe manager to not be nil")
	}
}

func TestProbeManager_StartStop(t *testing.T) {
	config := &Config{
		EnableTCP:        false,
		EnableDNS:        false,
		EnableSyscall:    false,
		EnableFileIO:     false,
		EnablePacketDrop: false,
		Duration:         1 * time.Second,
	}

	pm, err := NewProbeManager(config)
	if err != nil {
		t.Skipf("eBPF not available: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = pm.Start(ctx)
	if err != nil {
		t.Errorf("expected no error starting probe manager, got %v", err)
	}

	// Let it run briefly
	time.Sleep(100 * time.Millisecond)

	err = pm.Stop()
	if err != nil {
		t.Errorf("expected no error stopping probe manager, got %v", err)
	}
}

func TestParseTCPConnectEvent(t *testing.T) {
	// Valid data
	data := make([]byte, 20)
	data[0] = 192
	data[1] = 168
	data[2] = 1
	data[3] = 1 // 192.168.1.1

	event, err := ParseTCPConnectEvent(data)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if event == nil {
		t.Fatal("expected event to not be nil")
	}

	// Invalid data
	_, err = ParseTCPConnectEvent([]byte{1, 2, 3})
	if err == nil {
		t.Error("expected error for insufficient data")
	}
}

func TestParseDNSQueryEvent(t *testing.T) {
	// Valid data
	data := make([]byte, 268)
	copy(data[0:256], []byte("example.com"))

	event, err := ParseDNSQueryEvent(data)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if event == nil {
		t.Fatal("expected event to not be nil")
	}

	// Invalid data
	_, err = ParseDNSQueryEvent([]byte{1, 2, 3})
	if err == nil {
		t.Error("expected error for insufficient data")
	}
}

func TestFormatIP(t *testing.T) {
	// Test IP 192.168.1.1 (little endian: 0x0101A8C0)
	ip := uint32(0x0101A8C0)
	formatted := FormatIP(ip)

	expected := "192.168.1.1"
	if formatted != expected {
		t.Errorf("expected %s, got %s", expected, formatted)
	}
}

func TestTCPStats(t *testing.T) {
	stats := &TCPStats{
		TotalConnections:   100,
		ActiveConnections:  10,
		Retransmits:        5,
		AvgLatencyMicrosec: 1000,
		MaxLatencyMicrosec: 5000,
	}

	if stats.TotalConnections != 100 {
		t.Errorf("expected 100 connections, got %d", stats.TotalConnections)
	}
}

func TestDNSStats(t *testing.T) {
	stats := &DNSStats{
		TotalQueries:       50,
		FailedQueries:      2,
		AvgLatencyMicrosec: 500,
		MaxLatencyMicrosec: 2000,
	}

	if stats.TotalQueries != 50 {
		t.Errorf("expected 50 queries, got %d", stats.TotalQueries)
	}
}

func TestFileIOStats(t *testing.T) {
	stats := &FileIOStats{
		TotalReads:      1000,
		TotalWrites:     500,
		TotalReadBytes:  1024000,
		TotalWriteBytes: 512000,
		AvgLatencyMicrosec: 100,
	}

	if stats.TotalReads != 1000 {
		t.Errorf("expected 1000 reads, got %d", stats.TotalReads)
	}
}
