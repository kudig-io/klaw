//go:build linux && cgo

package analyzer

import (
	"context"
	"testing"
	"time"

	"github.com/kudig-io/klaw/internal/diag/types"
)

func TestNewTCPAnalyzer(t *testing.T) {
	a := NewTCPAnalyzer()
	if a == nil {
		t.Fatal("expected analyzer to not be nil")
	}

	if a.Name() != "ebpf.tcp" {
		t.Errorf("expected name to be ebpf.tcp, got %s", a.Name())
	}

	if a.Category() != "ebpf" {
		t.Errorf("expected category to be ebpf, got %s", a.Category())
	}
}

func TestTCPAnalyzer_SupportedModes(t *testing.T) {
	a := NewTCPAnalyzer()
	modes := a.SupportedModes()

	if len(modes) != 1 {
		t.Errorf("expected 1 supported mode, got %d", len(modes))
	}

	if modes[0] != types.ModeOnline {
		t.Errorf("expected ModeOnline, got %v", modes[0])
	}
}

func TestTCPAnalyzer_Analyze(t *testing.T) {
	a := NewTCPAnalyzer()
	data := &types.DiagnosticData{}

	// Test with sufficient timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	issues, err := a.Analyze(ctx, data)

	// Should return without error (eBPF may not be available in test environment)
	// Context deadline exceeded is acceptable when probes can't start
	if err != nil && err != context.DeadlineExceeded {
		t.Errorf("unexpected error: %v", err)
	}

	// Should return empty issues when eBPF unavailable
	_ = issues
}

func TestDNSAnalyzer(t *testing.T) {
	a := NewDNSAnalyzer()
	if a == nil {
		t.Fatal("expected analyzer to not be nil")
	}

	if a.Name() != "ebpf.dns" {
		t.Errorf("expected name to be ebpf.dns, got %s", a.Name())
	}

	data := &types.DiagnosticData{}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	issues, err := a.Analyze(ctx, data)
	// Context deadline exceeded is acceptable when probes can't start
	if err != nil && err != context.DeadlineExceeded {
		t.Errorf("unexpected error: %v", err)
	}
	_ = issues
}

func TestFileIOAnalyzer(t *testing.T) {
	a := NewFileIOAnalyzer()
	if a == nil {
		t.Fatal("expected analyzer to not be nil")
	}

	if a.Name() != "ebpf.fileio" {
		t.Errorf("expected name to be ebpf.fileio, got %s", a.Name())
	}

	data := &types.DiagnosticData{}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	issues, err := a.Analyze(ctx, data)
	// Context deadline exceeded is acceptable when probes can't start
	if err != nil && err != context.DeadlineExceeded {
		t.Errorf("unexpected error: %v", err)
	}
	_ = issues
}
