package probe

import (
	"encoding/binary"
	"fmt"
	"net"
	"time"
)

type EventType uint32

const (
	EventTypeTCPConnect EventType = iota
	EventTypeTCPDisconnect
	EventTypeTCPRetransmit
	EventTypeDNSEntry
	EventTypeDNSExit
	EventTypeSyscallEntry
	EventTypeSyscallExit
	EventTypeFileOpen
	EventTypeFileClose
	EventTypePacketDrop
)

type Event struct {
	Type      EventType
	Timestamp uint64
	PID       uint32
	Comm      [16]byte
	CPU       uint32
	Data      []byte
}

type TCPConnectEvent struct {
	Saddr   uint32
	Daddr   uint32
	Sport   uint16
	Dport   uint16
	Latency uint64
}

type DNSQueryEvent struct {
	Query      [256]byte
	QueryLen   uint32
	DNSLatency uint64
}

type SyscallEvent struct {
	SyscallID uint32
	Latency   uint64
}

type FileIOEvent struct {
	Filename [256]byte
	Op       uint32
	Size     uint64
	Latency  uint64
}

type PacketDropEvent struct {
	Reason uint32
	Saddr  uint32
	Daddr  uint32
	Sport  uint16
	Dport  uint16
	Proto  uint8
}

type Config struct {
	EnableTCP        bool
	EnableDNS        bool
	EnableSyscall    bool
	EnableFileIO     bool
	EnablePacketDrop bool
	Duration         time.Duration
}

type TCPStats struct {
	TotalConnections   uint64
	ActiveConnections  uint64
	Retransmits        uint64
	AvgLatencyMicrosec uint64
	MaxLatencyMicrosec uint64
}

type DNSStats struct {
	TotalQueries       uint64
	FailedQueries      uint64
	AvgLatencyMicrosec uint64
	MaxLatencyMicrosec uint64
}

type FileIOStats struct {
	TotalReads        uint64
	TotalWrites       uint64
	TotalReadBytes    uint64
	TotalWriteBytes   uint64
	AvgLatencyMicrosec uint64
}

func DefaultConfig() *Config {
	return &Config{
		EnableTCP:        true,
		EnableDNS:        true,
		EnableSyscall:    false,
		EnableFileIO:     true,
		EnablePacketDrop: true,
		Duration:         30 * time.Second,
	}
}

func ParseTCPConnectEvent(data []byte) (*TCPConnectEvent, error) {
	if len(data) < 16 {
		return nil, fmt.Errorf("insufficient data for TCP connect event")
	}
	return &TCPConnectEvent{
		Saddr:   binary.LittleEndian.Uint32(data[0:4]),
		Daddr:   binary.LittleEndian.Uint32(data[4:8]),
		Sport:   binary.LittleEndian.Uint16(data[8:10]),
		Dport:   binary.LittleEndian.Uint16(data[10:12]),
		Latency: binary.LittleEndian.Uint64(data[12:20]),
	}, nil
}

func ParseDNSQueryEvent(data []byte) (*DNSQueryEvent, error) {
	if len(data) < 264 {
		return nil, fmt.Errorf("insufficient data for DNS query event")
	}
	event := &DNSQueryEvent{
		QueryLen:   binary.LittleEndian.Uint32(data[256:260]),
		DNSLatency: binary.LittleEndian.Uint64(data[260:268]),
	}
	copy(event.Query[:], data[0:256])
	return event, nil
}

func FormatIP(ip uint32) string {
	return net.IP([]byte{
		byte(ip),
		byte(ip >> 8),
		byte(ip >> 16),
		byte(ip >> 24),
	}).String()
}
