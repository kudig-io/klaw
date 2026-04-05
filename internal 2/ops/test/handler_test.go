package ops_test

import (
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
	help := handler.HandleCommand("help")

	if help == "" {
		t.Error("ShowHelp() returned empty help message")
	}

	expectedHelp := "Available commands:"
	if help[:len(expectedHelp)] != expectedHelp {
		t.Errorf("ShowHelp() = %v, want %v", help, expectedHelp)
	}
}
