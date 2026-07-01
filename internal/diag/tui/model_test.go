package tui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/kudig-io/klaw/internal/diag/types"
)

func containsEmoji(title string, severity types.Severity) bool {
	switch severity {
	case types.SeverityCritical:
		return strings.Contains(title, "🔴")
	case types.SeverityWarning:
		return strings.Contains(title, "🟡")
	case types.SeverityInfo:
		return strings.Contains(title, "🟢")
	}
	return false
}

func TestNewModel(t *testing.T) {
	m := NewModel()
	if m.state != StateMenu {
		t.Errorf("Expected initial state StateMenu, got %d", m.state)
	}
	if m.onlineMode != true {
		t.Error("Expected onlineMode to be true by default")
	}
}

func TestModel_Init(t *testing.T) {
	m := NewModel()
	cmd := m.Init()
	if cmd != nil {
		t.Error("Expected nil Init command")
	}
}

func TestMenuItem_FilterValue(t *testing.T) {
	item := MenuItem{Title: "Test Item"}
	if item.FilterValue() != "Test Item" {
		t.Errorf("Expected FilterValue 'Test Item', got '%s'", item.FilterValue())
	}
}

func TestIssueItem_FilterValue(t *testing.T) {
	item := IssueItem{Issue: types.Issue{CNName: "CPU过高", ENName: "HIGH_CPU"}}
	expected := "CPU过高 HIGH_CPU"
	if item.FilterValue() != expected {
		t.Errorf("Expected FilterValue '%s', got '%s'", expected, item.FilterValue())
	}
}

func TestIssueItem_Title(t *testing.T) {
	tests := []struct {
		severity types.Severity
		expected string
	}{
		{types.SeverityCritical, "严重问题"},
		{types.SeverityWarning, "警告问题"},
		{types.SeverityInfo, "信息问题"},
	}
	for _, tc := range tests {
		item := IssueItem{Issue: types.Issue{Severity: tc.severity, CNName: tc.expected}}
		title := item.Title()
		if !containsEmoji(title, tc.severity) || !strings.Contains(title, tc.expected) {
			t.Errorf("Title should contain severity icon and '%s', got '%s'", tc.expected, title)
		}
	}
}

func TestIssueItem_Description(t *testing.T) {
	item := IssueItem{Issue: types.Issue{
		ENName:  "HIGH_CPU",
		Details: "CPU usage is 95%",
	}}
	desc := item.Description()
	if desc != "[HIGH_CPU] CPU usage is 95%" {
		t.Errorf("Unexpected description: %s", desc)
	}
}

func TestIssueItem_Description_Truncate(t *testing.T) {
	longDetails := ""
	for i := 0; i < 100; i++ {
		longDetails += "x"
	}
	item := IssueItem{Issue: types.Issue{
		ENName:  "TEST",
		Details: longDetails,
	}}
	desc := item.Description()
	if len(desc) > 60 {
		t.Errorf("Description should be truncated, got length %d: %s", len(desc), desc)
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"short", 10, "short"},
		{"a very long string that exceeds max", 10, "a very ..."},
		{"exact", 5, "exact"},
	}
	for _, tc := range tests {
		result := truncate(tc.input, tc.maxLen)
		if result != tc.expected {
			t.Errorf("truncate(%q, %d) = %q, expected %q", tc.input, tc.maxLen, result, tc.expected)
		}
	}
}

func TestCountIssuesBySeverity(t *testing.T) {
	issues := []types.Issue{
		{Severity: types.SeverityCritical},
		{Severity: types.SeverityCritical},
		{Severity: types.SeverityWarning},
		{Severity: types.SeverityInfo},
		{Severity: types.SeverityInfo},
		{Severity: types.SeverityInfo},
	}
	critical, warning, info := countIssuesBySeverity(issues)
	if critical != 2 {
		t.Errorf("Expected 2 critical, got %d", critical)
	}
	if warning != 1 {
		t.Errorf("Expected 1 warning, got %d", warning)
	}
	if info != 3 {
		t.Errorf("Expected 3 info, got %d", info)
	}
}

func TestCountIssuesBySeverity_Empty(t *testing.T) {
	critical, warning, info := countIssuesBySeverity(nil)
	if critical != 0 || warning != 0 || info != 0 {
		t.Errorf("Expected all zeros for nil issues, got %d/%d/%d", critical, warning, info)
	}
}

// Update and View tests

func TestUpdate_WindowSizeMsg(t *testing.T) {
	m := NewModel()
	msg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updated, _ := m.Update(msg)
	um := updated.(Model)
	if um.width != 120 {
		t.Errorf("Expected width 120, got %d", um.width)
	}
	if um.height != 40 {
		t.Errorf("Expected height 40, got %d", um.height)
	}
}

func TestUpdate_SpinnerTickMsg(t *testing.T) {
	m := NewModel()
	m.state = StateDiagnosing
	m.diagnosing = true
	msg := m.spinner.Tick()
	updated, _ := m.Update(msg)
	um := updated.(Model)
	_ = um
}

func TestUpdate_DiagnosisCompleteMsg_Success(t *testing.T) {
	m := NewModel()
	m.state = StateDiagnosing
	m.diagnosing = true
	issues := []types.Issue{
		{Severity: types.SeverityCritical, CNName: "严重问题", ENName: "CRIT1"},
		{Severity: types.SeverityWarning, CNName: "警告问题", ENName: "WARN1"},
	}
	data := types.NewDiagnosticData(types.ModeOffline)
	msg := DiagnosisCompleteMsg{Issues: issues, Data: data, Error: nil}
	updated, _ := m.Update(msg)
	um := updated.(Model)
	if um.diagnosing {
		t.Error("diagnosing should be false after completion")
	}
	if um.state != StateResults {
		t.Errorf("Expected StateResults, got %d", um.state)
	}
	if len(um.issues) != 2 {
		t.Errorf("Expected 2 issues, got %d", len(um.issues))
	}
	if um.diagError != nil {
		t.Errorf("Expected nil error, got %v", um.diagError)
	}
}

func TestUpdate_DiagnosisCompleteMsg_Error(t *testing.T) {
	m := NewModel()
	m.state = StateDiagnosing
	m.diagnosing = true
	msg := DiagnosisCompleteMsg{Error: fmt.Errorf("collection failed")}
	updated, _ := m.Update(msg)
	um := updated.(Model)
	if um.diagError == nil {
		t.Error("Expected error to be set")
	}
	if !strings.Contains(um.message, "collection failed") {
		t.Errorf("Message should contain error, got: %s", um.message)
	}
}

func TestUpdate_DiagnosisCompleteMsg_EmptyIssues(t *testing.T) {
	m := NewModel()
	m.state = StateDiagnosing
	msg := DiagnosisCompleteMsg{Issues: []types.Issue{}}
	updated, _ := m.Update(msg)
	um := updated.(Model)
	if um.state != StateResults {
		t.Errorf("Expected StateResults, got %d", um.state)
	}
	if len(um.issues) != 0 {
		t.Errorf("Expected 0 issues, got %d", len(um.issues))
	}
}

// View tests

func TestView_Menu(t *testing.T) {
	m := NewModel()
	view := m.View()
	if !strings.Contains(view, "klaw") {
		t.Error("Menu view should contain 'klaw'")
	}
}

func TestView_Diagnosing(t *testing.T) {
	m := NewModel()
	m.state = StateDiagnosing
	m.diagnosing = true
	m.onlineMode = true
	m.nodeName = "node-1"
	view := m.View()
	if !strings.Contains(view, "诊断中") {
		t.Error("Diagnosing view should contain '诊断中'")
	}
	if !strings.Contains(view, "node-1") {
		t.Error("Diagnosing view should show node name")
	}
}

func TestView_Diagnosing_Offline(t *testing.T) {
	m := NewModel()
	m.state = StateDiagnosing
	m.diagnosing = true
	m.onlineMode = false
	view := m.View()
	if !strings.Contains(view, "离线分析") {
		t.Error("Offline diagnosing view should contain '离线分析'")
	}
}

func TestView_Results(t *testing.T) {
	m := NewModel()
	m.state = StateResults
	m.issues = []types.Issue{
		{Severity: types.SeverityCritical, CNName: "严重", ENName: "CRIT"},
		{Severity: types.SeverityWarning, CNName: "警告", ENName: "WARN"},
		{Severity: types.SeverityInfo, CNName: "信息", ENName: "INFO"},
	}
	view := m.View()
	if !strings.Contains(view, "3") {
		t.Error("Results view should show issue count")
	}
	if !strings.Contains(view, "致命") {
		t.Error("Results should show critical count")
	}
}

func TestView_IssueDetail(t *testing.T) {
	m := NewModel()
	m.state = StateIssueDetail
	m.selectedIssue = &types.Issue{
		Severity: types.SeverityCritical,
		CNName:   "测试问题",
		ENName:   "TEST_ISSUE",
		Details:  "这是一个测试问题的详情",
		Location: "/etc/kubernetes/config",
		Remediation: &types.Remediation{
			Suggestion: "重启服务",
			Command:    "systemctl restart kubelet",
		},
	}
	view := m.View()
	if !strings.Contains(view, "测试问题") {
		t.Error("Detail view should contain issue CNName")
	}
	if !strings.Contains(view, "TEST_ISSUE") {
		t.Error("Detail view should contain issue ENName")
	}
	if !strings.Contains(view, "重启服务") {
		t.Error("Detail view should contain remediation suggestion")
	}
	if !strings.Contains(view, "systemctl restart kubelet") {
		t.Error("Detail view should contain remediation command")
	}
}

func TestView_IssueDetail_WithMetadata(t *testing.T) {
	m := NewModel()
	m.state = StateIssueDetail
	m.selectedIssue = &types.Issue{
		CNName:   "元数据问题",
		ENName:   "META_ISSUE",
		Severity: types.SeverityWarning,
		Metadata: map[string]string{
			"pod":       "my-pod",
			"namespace": "default",
		},
	}
	view := m.View()
	if !strings.Contains(view, "my-pod") {
		t.Error("Detail view should show metadata pod")
	}
	if !strings.Contains(view, "default") {
		t.Error("Detail view should show metadata namespace")
	}
}

func TestView_IssueDetail_NilIssue(t *testing.T) {
	m := NewModel()
	m.state = StateIssueDetail
	m.selectedIssue = nil
	view := m.View()
	if !strings.Contains(view, "未选择问题") {
		t.Error("Detail view with nil issue should show '未选择问题'")
	}
}

func TestView_DefaultState(t *testing.T) {
	m := NewModel()
	m.state = AppState(99)
	view := m.View()
	if !strings.Contains(view, "klaw") {
		t.Error("Default view should fall back to menu")
	}
}

// Key handling tests

func TestUpdate_KeyMsg_Quit(t *testing.T) {
	m := NewModel()
	m.state = StateMenu
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Error("Expected quit command for ctrl+c")
	}
	_ = updated
}

func TestUpdate_KeyMsg_ResultsBack(t *testing.T) {
	m := NewModel()
	m.state = StateResults
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	um := updated.(Model)
	if um.state != StateMenu {
		t.Errorf("Expected StateMenu after 'b', got %d", um.state)
	}
}

func TestUpdate_KeyMsg_IssueDetailBack(t *testing.T) {
	m := NewModel()
	m.state = StateIssueDetail
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	um := updated.(Model)
	if um.state != StateResults {
		t.Errorf("Expected StateResults after 'b', got %d", um.state)
	}
}

func TestUpdate_KeyMsg_ResultsQuit(t *testing.T) {
	m := NewModel()
	m.state = StateResults
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Error("Expected quit command for 'q' in results")
	}
}

func TestUpdate_KeyMsg_IssueDetailQuit(t *testing.T) {
	m := NewModel()
	m.state = StateIssueDetail
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Error("Expected quit command for 'q' in issue detail")
	}
}

func TestUpdate_KeyMsg_ResultsSelectIssue(t *testing.T) {
	m := NewModel()
	m.state = StateResults
	m.issues = []types.Issue{
		{Severity: types.SeverityCritical, CNName: "严重", ENName: "CRIT"},
	}
	items := make([]list.Item, 1)
	items[0] = IssueItem{Issue: m.issues[0]}
	m.issuesList.SetItems(items)
	m.issuesList.Select(0)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	um := updated.(Model)
	if um.state != StateIssueDetail {
		t.Errorf("Expected StateIssueDetail after Enter, got %d", um.state)
	}
	if um.selectedIssue == nil {
		t.Error("Expected selectedIssue to be set")
	}
}

// Styles

func TestStyles_NotNil(t *testing.T) {
	styles := []lipgloss.Style{
		titleStyle, itemStyle, selectedItemStyle, infoStyle,
		errorStyle, warningStyle, successStyle, criticalStyle, descriptionStyle,
	}
	for i, s := range styles {
		if s.GetWidth() == 0 && s.GetHeight() == 0 && s.GetForeground() == (lipgloss.Color("")) && s.GetBackground() == (lipgloss.Color("")) {
			t.Logf("Style %d may be unstyled (this is informational)", i)
		}
	}
}
