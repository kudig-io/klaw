package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/kudig-io/klaw/internal/diag/analyzer"
	"github.com/kudig-io/klaw/internal/diag/collector"
	"github.com/kudig-io/klaw/internal/diag/types"
)

// DiagnosisCompleteMsg is sent when diagnosis is complete
type DiagnosisCompleteMsg struct {
	Issues []types.Issue
	Data   *types.DiagnosticData
	Error  error
}

// startDiagnosis starts the diagnosis process
func (m Model) startDiagnosis() tea.Cmd {
	return func() tea.Msg {
		mode := types.ModeOnline
		var cfg *collector.Config

		if m.onlineMode {
			mode = types.ModeOnline
			cfg = collector.NewOnlineConfig(m.kubeconfig, m.nodeName)
			cfg.Namespace = m.namespace
		} else {
			mode = types.ModeOffline
			cfg = collector.NewOfflineConfig(m.diagnosePath)
		}

		coll, ok := collector.GetCollector(mode)
		if !ok {
			return DiagnosisCompleteMsg{
				Error: fmt.Errorf("%s mode collector not available", mode),
			}
		}

		if err := coll.Validate(cfg); err != nil {
			return DiagnosisCompleteMsg{
				Error: fmt.Errorf("collector validation failed: %w", err),
			}
		}

		data, err := coll.Collect(m.context, cfg)
		if err != nil {
			return DiagnosisCompleteMsg{
				Error: fmt.Errorf("data collection failed: %w", err),
			}
		}

		results, err := analyzer.DefaultRegistry.ExecuteAll(m.context, data)
		if err != nil {
			return DiagnosisCompleteMsg{
				Error: fmt.Errorf("analysis failed: %w", err),
			}
		}

		issues := analyzer.CollectIssues(results)
		if issues == nil {
			issues = []types.Issue{}
		}

		return DiagnosisCompleteMsg{
			Issues: issues,
			Data:   data,
			Error:  nil,
		}
	}
}

// RunTUI starts the terminal user interface
func RunTUI() error {
	model := NewModel()
	p := tea.NewProgram(model, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
