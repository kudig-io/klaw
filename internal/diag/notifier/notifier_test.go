package notifier

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kudig-io/klaw/internal/diag/types"
)

func TestNewConfigFromEnv(t *testing.T) {
	// Save and restore environment variables
	oldSlackURL := ""
	oldMinSev := ""
	
	// Set test environment variables
	t.Setenv("KUDIG_SLACK_WEBHOOK_URL", "https://hooks.slack.com/test")
	t.Setenv("KUDIG_NOTIFY_MIN_SEVERITY", "warning")

	config := NewConfigFromEnv()

	if config.SlackWebhookURL != "https://hooks.slack.com/test" {
		t.Errorf("Expected Slack URL to be set, got '%s'", config.SlackWebhookURL)
	}

	if config.MinSeverity != types.SeverityWarning {
		t.Errorf("Expected min severity warning, got %v", config.MinSeverity)
	}

	// Reset
	if oldSlackURL != "" {
		t.Setenv("KUDIG_SLACK_WEBHOOK_URL", oldSlackURL)
	} else {
		t.Setenv("KUDIG_SLACK_WEBHOOK_URL", "")
	}
	if oldMinSev != "" {
		t.Setenv("KUDIG_NOTIFY_MIN_SEVERITY", oldMinSev)
	} else {
		t.Setenv("KUDIG_NOTIFY_MIN_SEVERITY", "")
	}
}

func TestConfig_IsEnabled(t *testing.T) {
	tests := []struct {
		name     string
		config   Config
		expected bool
	}{
		{
			name:     "Empty config",
			config:   Config{},
			expected: false,
		},
		{
			name: "Slack enabled",
			config: Config{
				SlackWebhookURL: "https://hooks.slack.com/test",
			},
			expected: true,
		},
		{
			name: "DingTalk enabled",
			config: Config{
				DingTalkWebhookURL: "https://oapi.dingtalk.com/test",
			},
			expected: true,
		},
		{
			name: "WeChat enabled",
			config: Config{
				WeChatWebhookURL: "https://qyapi.weixin.qq.com/test",
			},
			expected: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.config.IsEnabled() != tc.expected {
				t.Errorf("Expected IsEnabled() = %v", tc.expected)
			}
		})
	}
}

func TestConfig_ShouldNotify(t *testing.T) {
	config := &Config{
		SlackWebhookURL: "https://hooks.slack.com/test",
		MinSeverity:     types.SeverityWarning,
	}

	tests := []struct {
		name     string
		issues   []types.Issue
		expected bool
	}{
		{
			name:     "No issues",
			issues:   []types.Issue{},
			expected: false,
		},
		{
			name: "Only info issues",
			issues: []types.Issue{
				{Severity: types.SeverityInfo},
			},
			expected: false,
		},
		{
			name: "Has warning issue",
			issues: []types.Issue{
				{Severity: types.SeverityWarning},
			},
			expected: true,
		},
		{
			name: "Has critical issue",
			issues: []types.Issue{
				{Severity: types.SeverityCritical},
			},
			expected: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if config.ShouldNotify(tc.issues) != tc.expected {
				t.Errorf("Expected ShouldNotify() = %v for %s", tc.expected, tc.name)
			}
		})
	}
}

func TestNewMultiNotifier(t *testing.T) {
	config := &Config{
		SlackWebhookURL:    "https://hooks.slack.com/test",
		DingTalkWebhookURL: "https://oapi.dingtalk.com/test",
		WeChatWebhookURL:   "https://qyapi.weixin.qq.com/test",
	}

	mn := NewMultiNotifier(config)

	if len(mn.Notifiers) != 3 {
		t.Errorf("Expected 3 notifiers, got %d", len(mn.Notifiers))
	}
}

func TestSlackNotifier_Name(t *testing.T) {
	n := NewSlackNotifier("https://hooks.slack.com/test")
	if n.Name() != "Slack" {
		t.Errorf("Expected name 'Slack', got '%s'", n.Name())
	}
}

func TestDingTalkNotifier_Name(t *testing.T) {
	n := NewDingTalkNotifier("https://oapi.dingtalk.com/test")
	if n.Name() != "DingTalk" {
		t.Errorf("Expected name 'DingTalk', got '%s'", n.Name())
	}
}

func TestWeChatNotifier_Name(t *testing.T) {
	n := NewWeChatNotifier("https://qyapi.weixin.qq.com/test")
	if n.Name() != "WeChat Work" {
		t.Errorf("Expected name 'WeChat Work', got '%s'", n.Name())
	}
}

func TestMultiNotifier_Send_NoNotifiers(t *testing.T) {
	mn := &MultiNotifier{}
	errors := mn.Send("Test", "Message", nil)

	if len(errors) != 0 {
		t.Errorf("Expected 0 errors with no notifiers, got %d", len(errors))
	}
}

func TestNewConfigFromEnv_DefaultSeverity(t *testing.T) {
	t.Setenv("KUDIG_NOTIFY_MIN_SEVERITY", "")
	t.Setenv("KUDIG_SLACK_WEBHOOK_URL", "")
	t.Setenv("KUDIG_DINGTALK_WEBHOOK_URL", "")
	t.Setenv("KUDIG_WECHAT_WEBHOOK_URL", "")

	config := NewConfigFromEnv()
	if config.MinSeverity != types.SeverityCritical {
		t.Errorf("Expected default severity Critical, got %v", config.MinSeverity)
	}
	if config.IsEnabled() {
		t.Error("Expected IsEnabled() = false with no URLs")
	}
}

func TestNewConfigFromEnv_InfoSeverity(t *testing.T) {
	t.Setenv("KUDIG_NOTIFY_MIN_SEVERITY", "info")
	config := NewConfigFromEnv()
	if config.MinSeverity != types.SeverityInfo {
		t.Errorf("Expected severity Info, got %v", config.MinSeverity)
	}
}

func TestNewConfigFromEnv_CriticalSeverity(t *testing.T) {
	t.Setenv("KUDIG_NOTIFY_MIN_SEVERITY", "critical")
	config := NewConfigFromEnv()
	if config.MinSeverity != types.SeverityCritical {
		t.Errorf("Expected severity Critical, got %v", config.MinSeverity)
	}
}

func TestNewConfigFromEnv_DingTalkAndWeChat(t *testing.T) {
	t.Setenv("KUDIG_DINGTALK_WEBHOOK_URL", "https://oapi.dingtalk.com/test")
	t.Setenv("KUDIG_WECHAT_WEBHOOK_URL", "https://qyapi.weixin.qq.com/test")
	t.Setenv("KUDIG_SLACK_WEBHOOK_URL", "")
	t.Setenv("KUDIG_NOTIFY_MIN_SEVERITY", "")

	config := NewConfigFromEnv()
	if !config.IsEnabled() {
		t.Error("Expected IsEnabled() = true")
	}
	if config.DingTalkWebhookURL != "https://oapi.dingtalk.com/test" {
		t.Errorf("DingTalk URL not set correctly")
	}
	if config.WeChatWebhookURL != "https://qyapi.weixin.qq.com/test" {
		t.Errorf("WeChat URL not set correctly")
	}
}

func TestShouldNotify_Disabled(t *testing.T) {
	config := &Config{}
	if config.ShouldNotify([]types.Issue{{Severity: types.SeverityCritical}}) {
		t.Error("ShouldNotify should return false when no notifiers configured")
	}
}

func TestMultiNotifier_EmptyConfig(t *testing.T) {
	mn := NewMultiNotifier(&Config{})
	if len(mn.Notifiers) != 0 {
		t.Errorf("Expected 0 notifiers for empty config, got %d", len(mn.Notifiers))
	}
}

func TestSlackNotifier_Send_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	n := NewSlackNotifier(server.URL)
	issues := []types.Issue{
		{Severity: types.SeverityCritical, CNName: "Test Issue", Details: "test details"},
		{Severity: types.SeverityWarning, CNName: "Warn Issue", Details: "warn details"},
	}
	if err := n.Send("Title", "Message", issues); err != nil {
		t.Fatalf("Send() error = %v", err)
	}
}

func TestSlackNotifier_Send_EmptyIssues(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	n := NewSlackNotifier(server.URL)
	if err := n.Send("Title", "Message", nil); err != nil {
		t.Fatalf("Send() error = %v", err)
	}
}

func TestSlackNotifier_Send_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	n := NewSlackNotifier(server.URL)
	err := n.Send("Title", "Message", nil)
	if err == nil {
		t.Fatal("Expected error for 500 status")
	}
}

func TestSlackNotifier_Send_InvalidURL(t *testing.T) {
	n := NewSlackNotifier("http://127.0.0.1:0/invalid")
	err := n.Send("Title", "Message", nil)
	if err == nil {
		t.Fatal("Expected error for unreachable URL")
	}
}

func TestDingTalkNotifier_Send_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	n := NewDingTalkNotifier(server.URL)
	issues := []types.Issue{
		{Severity: types.SeverityCritical, CNName: "严重问题", Details: "test"},
		{Severity: types.SeverityWarning, CNName: "警告问题", Details: "warn"},
	}
	if err := n.Send("诊断标题", "诊断消息", issues); err != nil {
		t.Fatalf("Send() error = %v", err)
	}
}

func TestDingTalkNotifier_Send_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer server.Close()

	n := NewDingTalkNotifier(server.URL)
	err := n.Send("Title", "Msg", nil)
	if err == nil {
		t.Fatal("Expected error for 502 status")
	}
}

func TestWeChatNotifier_Send_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	n := NewWeChatNotifier(server.URL)
	issues := []types.Issue{
		{Severity: types.SeverityCritical, CNName: "严重", Details: "details"},
	}
	if err := n.Send("标题", "消息", issues); err != nil {
		t.Fatalf("Send() error = %v", err)
	}
}

func TestWeChatNotifier_Send_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	n := NewWeChatNotifier(server.URL)
	err := n.Send("Title", "Msg", nil)
	if err == nil {
		t.Fatal("Expected error for 500 status")
	}
}

func TestMultiNotifier_Send_AllSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	mn := &MultiNotifier{
		Notifiers: []Notifier{
			NewSlackNotifier(server.URL),
			NewDingTalkNotifier(server.URL),
		},
	}
	errs := mn.Send("Title", "Message", nil)
	if len(errs) != 0 {
		t.Errorf("Expected 0 errors, got %d: %v", len(errs), errs)
	}
}

func TestMultiNotifier_Send_PartialFailure(t *testing.T) {
	goodServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer goodServer.Close()

	badServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer badServer.Close()

	mn := &MultiNotifier{
		Notifiers: []Notifier{
			NewSlackNotifier(goodServer.URL),
			NewDingTalkNotifier(badServer.URL),
		},
	}
	errs := mn.Send("Title", "Message", nil)
	if len(errs) != 1 {
		t.Errorf("Expected 1 error, got %d", len(errs))
	}
}
