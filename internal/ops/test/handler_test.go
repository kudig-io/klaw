package ops_test

import (
	"strings"
	"testing"

	"github.com/kudig-io/klaw/internal/ops"
)

func TestHandler_HandleCommand(t *testing.T) {
	handler := ops.NewHandler(nil, nil)

	tests := []struct {
		name    string
		command string
		wantErr bool
	}{
		{
			name:    "help command",
			command: "help",
			wantErr: false,
		},
		{
			name:    "empty command",
			command: "",
			wantErr: true,
		},
		{
			name:    "unknown command",
			command: "unknown",
			wantErr: true,
		},
		{
			name:    "cluster status command",
			command: "cluster status test",
			wantErr: true,
		},
		{
			name:    "pod list command",
			command: "pod list test default",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := handler.HandleCommand(tt.command)
			if (err != nil) != tt.wantErr {
				t.Errorf("HandleCommand() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && result == "" {
				t.Errorf("HandleCommand() returned empty result")
			}
		})
	}
}

func TestHandler_ShowHelp(t *testing.T) {
	handler := ops.NewHandler(nil, nil)
	help, err := handler.HandleCommand("help")
	if err != nil {
		t.Fatalf("HandleCommand(help) error = %v", err)
	}

	if help == "" {
		t.Error("ShowHelp() returned empty help message")
	}

	expectedHelp := "## Klaw ChatOps Commands"
	if len(help) < len(expectedHelp) || help[:len(expectedHelp)] != expectedHelp {
		t.Errorf("ShowHelp() = %v, want prefix %v", help, expectedHelp)
	}
}

func TestHandler_DeletePodDestructiveGuard(t *testing.T) {
	handler := ops.NewHandler(nil, nil)

	// 默认关闭：delete 命令直接拒绝（即使 k8s 未初始化也不应触发删除路径）
	_, err := handler.HandleCommand("pod delete c default p")
	if err == nil {
		t.Fatal("destructive command must be rejected by default")
	}
	if !strings.Contains(err.Error(), "allow_destructive") {
		t.Errorf("rejection should mention allow_destructive, got: %v", err)
	}

	// 显式开启后：走到 k8s 初始化检查（nil manager 报 kubernetes manager not initialized）
	handler.SetAllowDestructive(true)
	_, err = handler.HandleCommand("pod delete c default p")
	if err == nil {
		t.Fatal("expected kubernetes manager error when enabled but manager is nil")
	}
	if !strings.Contains(err.Error(), "kubernetes manager not initialized") {
		t.Errorf("expected kubernetes manager error, got: %v", err)
	}
}
