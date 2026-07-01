// Package tui provides a Terminal User Interface for klaw using bubbletea
package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/kudig-io/klaw/internal/diag/types"
)

// Styles for TUI
var (
	titleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		PaddingLeft(2).
		PaddingRight(2).
		MarginBottom(1)

	itemStyle = lipgloss.NewStyle().
		PaddingLeft(2)

	selectedItemStyle = lipgloss.NewStyle().
		PaddingLeft(2).
		Foreground(lipgloss.Color("#7D56F4")).
		Bold(true)

	infoStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666666"))

	errorStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF5555"))

	warningStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFAA00"))

	successStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#55AA55"))

	criticalStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF0000")).
		Bold(true)

	descriptionStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#AAAAAA"))
)

// MenuItem represents a menu item
type MenuItem struct {
	Title       string
	Description string
	Action      func() tea.Cmd
}

// FilterValue implements list.Item interface
func (i MenuItem) FilterValue() string { return i.Title }

// IssueItem wraps types.Issue for list display
type IssueItem struct {
	Issue types.Issue
}

// FilterValue implements list.Item interface
func (i IssueItem) FilterValue() string {
	return i.Issue.CNName + " " + i.Issue.ENName
}

// Title returns the issue title
func (i IssueItem) Title() string {
	severityIcon := "•"
	if i.Issue.Severity == types.SeverityCritical {
		severityIcon = "🔴"
	} else if i.Issue.Severity == types.SeverityWarning {
		severityIcon = "🟡"
	} else if i.Issue.Severity == types.SeverityInfo {
		severityIcon = "🟢"
	}
	return fmt.Sprintf("%s %s", severityIcon, i.Issue.CNName)
}

// Description returns the issue description
func (i IssueItem) Description() string {
	return fmt.Sprintf("[%s] %s", i.Issue.ENName, truncate(i.Issue.Details, 50))
}

// Model is the TUI application model
type Model struct {
	// State
	state       AppState
	width       int
	height      int
	context     context.Context
	cancelFunc  context.CancelFunc

	// Menu
	menu        list.Model
	menuItems   []MenuItem
	selectedIdx int

	// Diagnosis
	issues      []types.Issue
	issuesList  list.Model
	selectedIssue *types.Issue
	spinner     spinner.Model
	diagnosing  bool
	diagError   error
	diagResult  *types.DiagnosticData

	// Inputs
	kubeconfig   string
	nodeName     string
	namespace    string
	diagnosePath string
	onlineMode   bool

	// Messages
	message     string
}

// AppState represents the current state of the TUI
type AppState int

const (
	StateMenu AppState = iota
	StateDiagnosing
	StateResults
	StateIssueDetail
	StateHistory
	StateSettings
)

// NewModel creates a new TUI model
func NewModel() Model {
	ctx, cancel := context.WithCancel(context.Background())

	// Create spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4"))

	// Create menu items
	menuItems := []list.Item{
		MenuItem{
			Title:       "🚀 在线诊断",
			Description: "连接 Kubernetes 集群进行实时诊断",
		},
		MenuItem{
			Title:       "📁 离线分析",
			Description: "分析已收集的诊断数据",
		},
		MenuItem{
			Title:       "📜 历史记录",
			Description: "查看和对比历史诊断结果",
		},
		MenuItem{
			Title:       "⚙️  设置",
			Description: "配置诊断参数和选项",
		},
		MenuItem{
			Title:       "❌ 退出",
			Description: "退出 klaw",
		},
	}

	// Create menu list
	menuDelegate := list.NewDefaultDelegate()
	menuDelegate.Styles.SelectedTitle = selectedItemStyle
	menuDelegate.Styles.SelectedDesc = descriptionStyle

	menu := list.New(menuItems, menuDelegate, 50, 15)
	menu.Title = "主菜单"
	menu.SetShowStatusBar(false)
	menu.SetFilteringEnabled(false)
	menu.Styles.Title = titleStyle

	// Create issues list (empty initially)
	issueDelegate := list.NewDefaultDelegate()
	issueDelegate.Styles.SelectedTitle = selectedItemStyle
	issueDelegate.Styles.SelectedDesc = descriptionStyle

	issuesList := list.New([]list.Item{}, issueDelegate, 80, 20)
	issuesList.Title = "诊断结果"
	issuesList.SetShowStatusBar(true)
	issuesList.Styles.Title = titleStyle

	return Model{
		state:      StateMenu,
		context:    ctx,
		cancelFunc: cancel,
		spinner:    s,
		menu:       menu,
		issuesList: issuesList,
		onlineMode: true,
	}
}

// Init initializes the TUI
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyMsg(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.menu.SetWidth(msg.Width)
		m.issuesList.SetWidth(msg.Width)
		return m, nil

	case DiagnosisCompleteMsg:
		m.diagnosing = false
		m.issues = msg.Issues
		m.diagResult = msg.Data
		if msg.Error != nil {
			m.diagError = msg.Error
			m.message = errorStyle.Render(fmt.Sprintf("诊断失败: %v", msg.Error))
		} else {
			m.message = successStyle.Render(fmt.Sprintf("诊断完成！发现 %d 个问题", len(msg.Issues)))
			// Update issues list
			items := make([]list.Item, len(msg.Issues))
			for i, issue := range msg.Issues {
				items[i] = IssueItem{Issue: issue}
			}
			m.issuesList.SetItems(items)
			m.state = StateResults
		}
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	// Update sub-models based on state
	switch m.state {
	case StateMenu:
		var cmd tea.Cmd
		m.menu, cmd = m.menu.Update(msg)
		return m, cmd
	case StateResults:
		var cmd tea.Cmd
		m.issuesList, cmd = m.issuesList.Update(msg)
		return m, cmd
	}

	return m, nil
}

// View renders the TUI
func (m Model) View() string {
	switch m.state {
	case StateMenu:
		return m.viewMenu()
	case StateDiagnosing:
		return m.viewDiagnosing()
	case StateResults:
		return m.viewResults()
	case StateIssueDetail:
		return m.viewIssueDetail()
	default:
		return m.viewMenu()
	}
}

func (m Model) viewMenu() string {
	var b strings.Builder

	// Header
	b.WriteString(titleStyle.Render("  🔧 klaw - Kubernetes 诊断工具  "))
	b.WriteString("\n\n")

	// Menu
	b.WriteString(m.menu.View())
	b.WriteString("\n\n")

	// Info
	b.WriteString(infoStyle.Render("↑/↓: 选择 | Enter: 确认 | q: 退出"))

	if m.message != "" {
		b.WriteString("\n\n")
		b.WriteString(m.message)
	}

	return b.String()
}

func (m Model) viewDiagnosing() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("  🔧 klaw - 诊断中...  "))
	b.WriteString("\n\n")

	b.WriteString(m.spinner.View())
	b.WriteString(" 正在执行诊断...")
	b.WriteString("\n\n")

	if m.onlineMode {
		b.WriteString(infoStyle.Render(fmt.Sprintf("模式: 在线诊断 | 节点: %s", m.nodeName)))
	} else {
		b.WriteString(infoStyle.Render("模式: 离线分析"))
	}

	b.WriteString("\n\n")
	b.WriteString(infoStyle.Render("按 Ctrl+C 取消"))

	return b.String()
}

func (m Model) viewResults() string {
	var b strings.Builder

	// Header with summary
	summary := fmt.Sprintf("  📊 诊断结果 - %d 个问题  ", len(m.issues))
	b.WriteString(titleStyle.Render(summary))
	b.WriteString("\n\n")

	// Summary by severity
	critical, warning, info := countIssuesBySeverity(m.issues)
	b.WriteString(fmt.Sprintf("%s %s %s\n\n",
		criticalStyle.Render(fmt.Sprintf("🔴 致命: %d", critical)),
		warningStyle.Render(fmt.Sprintf("🟡 警告: %d", warning)),
		successStyle.Render(fmt.Sprintf("🟢 信息: %d", info)),
	))

	// Issues list
	b.WriteString(m.issuesList.View())
	b.WriteString("\n")

	// Help
	b.WriteString(infoStyle.Render("↑/↓: 选择 | Enter: 详情 | b: 返回 | q: 退出"))

	return b.String()
}

func (m Model) viewIssueDetail() string {
	if m.selectedIssue == nil {
		return "未选择问题"
	}

	var b strings.Builder

	issue := m.selectedIssue

	// Header
	title := fmt.Sprintf("  %s  ", issue.CNName)
	b.WriteString(titleStyle.Render(title))
	b.WriteString("\n\n")

	// Severity
	sevStyle := infoStyle
	if issue.Severity == types.SeverityCritical {
		sevStyle = criticalStyle
	} else if issue.Severity == types.SeverityWarning {
		sevStyle = warningStyle
	}
	b.WriteString(fmt.Sprintf("严重级别: %s\n", sevStyle.Render(issue.Severity.String())))

	// Code (ENName)
	b.WriteString(fmt.Sprintf("问题代码: %s\n", infoStyle.Render(issue.ENName)))
	b.WriteString("\n")

	// Details
	b.WriteString(descriptionStyle.Render("详细信息:"))
	b.WriteString("\n")
	b.WriteString(itemStyle.Render(issue.Details))
	b.WriteString("\n\n")

	// Location
	if issue.Location != "" {
		b.WriteString(descriptionStyle.Render("位置:"))
		b.WriteString("\n")
		b.WriteString(itemStyle.Render(issue.Location))
		b.WriteString("\n\n")
	}

	// Remediation
	if issue.Remediation != nil && issue.Remediation.Suggestion != "" {
		b.WriteString(descriptionStyle.Render("修复建议:"))
		b.WriteString("\n")
		b.WriteString(itemStyle.Render(issue.Remediation.Suggestion))
		if issue.Remediation.Command != "" {
			b.WriteString("\n")
			b.WriteString(itemStyle.Render(fmt.Sprintf("命令: %s", issue.Remediation.Command)))
		}
		b.WriteString("\n\n")
	}

	// Metadata
	if len(issue.Metadata) > 0 {
		b.WriteString(descriptionStyle.Render("附加信息:"))
		b.WriteString("\n")
		for k, v := range issue.Metadata {
			b.WriteString(itemStyle.Render(fmt.Sprintf("%s: %s", k, v)))
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	// Back hint
	b.WriteString(infoStyle.Render("按 b 返回"))

	return b.String()
}

func (m Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.state {
	case StateMenu:
		switch msg.String() {
		case "q", "ctrl+c":
			m.cancelFunc()
			return m, tea.Quit
		case "enter":
			if i, ok := m.menu.SelectedItem().(MenuItem); ok {
				switch i.Title {
				case "🚀 在线诊断":
					m.onlineMode = true
					m.state = StateDiagnosing
					m.diagnosing = true
					return m, tea.Batch(
						m.spinner.Tick,
						m.startDiagnosis(),
					)
				case "📁 离线分析":
					m.onlineMode = false
					m.state = StateDiagnosing
					m.diagnosing = true
					return m, tea.Batch(
						m.spinner.Tick,
						m.startDiagnosis(),
					)
				case "❌ 退出":
					return m, tea.Quit
				}
			}
		}

	case StateResults:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "b":
			m.state = StateMenu
			return m, nil
		case "enter":
			if i, ok := m.issuesList.SelectedItem().(IssueItem); ok {
				m.selectedIssue = &i.Issue
				m.state = StateIssueDetail
			}
			return m, nil
		}

	case StateIssueDetail:
		switch msg.String() {
		case "b":
			m.state = StateResults
			return m, nil
		case "q":
			return m, tea.Quit
		}
	}

	return m, nil
}

// Helper functions

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func countIssuesBySeverity(issues []types.Issue) (critical, warning, info int) {
	for _, issue := range issues {
		switch issue.Severity {
		case types.SeverityCritical:
			critical++
		case types.SeverityWarning:
			warning++
		case types.SeverityInfo:
			info++
		}
	}
	return
}
