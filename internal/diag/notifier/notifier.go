// Package notifier provides webhook notification capabilities for diagnostic alerts
package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/kudig-io/klaw/internal/diag/types"
)

// Notifier defines the interface for sending notifications
type Notifier interface {
	// Send sends a notification with the given title and message
	Send(title, message string, issues []types.Issue) error
	// Name returns the notifier name
	Name() string
}

// Config holds configuration for notifiers
type Config struct {
	// Slack webhook URL
	SlackWebhookURL string
	// DingTalk webhook URL
	DingTalkWebhookURL string
	// WeChat Work webhook URL
	WeChatWebhookURL string
	// Minimum severity to trigger notification (critical, warning, info)
	MinSeverity types.Severity
}

// NewConfigFromEnv creates a new Config from environment variables
func NewConfigFromEnv() *Config {
	minSev := types.SeverityCritical // Default
	if sev := os.Getenv("KUDIG_NOTIFY_MIN_SEVERITY"); sev != "" {
		switch sev {
		case "critical":
			minSev = types.SeverityCritical
		case "warning":
			minSev = types.SeverityWarning
		case "info":
			minSev = types.SeverityInfo
		}
	}

	return &Config{
		SlackWebhookURL:    os.Getenv("KUDIG_SLACK_WEBHOOK_URL"),
		DingTalkWebhookURL: os.Getenv("KUDIG_DINGTALK_WEBHOOK_URL"),
		WeChatWebhookURL:   os.Getenv("KUDIG_WECHAT_WEBHOOK_URL"),
		MinSeverity:        minSev,
	}
}

// IsEnabled returns true if any notifier is configured
func (c *Config) IsEnabled() bool {
	return c.SlackWebhookURL != "" || c.DingTalkWebhookURL != "" || c.WeChatWebhookURL != ""
}

// ShouldNotify returns true if the given issues should trigger a notification
func (c *Config) ShouldNotify(issues []types.Issue) bool {
	if !c.IsEnabled() {
		return false
	}

	for _, issue := range issues {
		if issue.Severity <= c.MinSeverity {
			return true
		}
	}
	return false
}

// MultiNotifier sends notifications to multiple channels
type MultiNotifier struct {
	Notifiers []Notifier
}

// NewMultiNotifier creates a new MultiNotifier from config
func NewMultiNotifier(config *Config) *MultiNotifier {
	mn := &MultiNotifier{}

	if config.SlackWebhookURL != "" {
		mn.Notifiers = append(mn.Notifiers, NewSlackNotifier(config.SlackWebhookURL))
	}

	if config.DingTalkWebhookURL != "" {
		mn.Notifiers = append(mn.Notifiers, NewDingTalkNotifier(config.DingTalkWebhookURL))
	}

	if config.WeChatWebhookURL != "" {
		mn.Notifiers = append(mn.Notifiers, NewWeChatNotifier(config.WeChatWebhookURL))
	}

	return mn
}

// Send sends notifications to all configured channels
func (m *MultiNotifier) Send(title, message string, issues []types.Issue) []error {
	var errors []error
	for _, n := range m.Notifiers {
		if err := n.Send(title, message, issues); err != nil {
			errors = append(errors, fmt.Errorf("%s: %w", n.Name(), err))
		}
	}
	return errors
}

// SlackNotifier sends notifications to Slack
type SlackNotifier struct {
	webhookURL string
	client     *http.Client
}

// NewSlackNotifier creates a new Slack notifier
func NewSlackNotifier(webhookURL string) *SlackNotifier {
	return &SlackNotifier{
		webhookURL: webhookURL,
		client:     &http.Client{Timeout: 10 * time.Second},
	}
}

// Name returns the notifier name
func (s *SlackNotifier) Name() string {
	return "Slack"
}

// Send sends a notification to Slack
func (s *SlackNotifier) Send(title, message string, issues []types.Issue) error {
	// Format issues for Slack
	var fields []map[string]interface{}
	for _, issue := range issues {
		if issue.Severity > types.SeverityInfo {
			continue // Skip info level for Slack notifications
		}
		color := "#36a64f" // green
		if issue.Severity == types.SeverityWarning {
			color = "#ff9900" // orange
		} else if issue.Severity == types.SeverityCritical {
			color = "#ff0000" // red
		}

		fields = append(fields, map[string]interface{}{
			"title": issue.CNName,
			"value": issue.Details,
			"short": false,
			"color": color,
		})
	}

	payload := map[string]interface{}{
		"text": title,
		"attachments": []map[string]interface{}{
			{
				"color":  "#ff0000",
				"title":  message,
				"fields": fields,
				"footer": "Kudig Kubernetes Diagnostic",
				"ts":     time.Now().Unix(),
			},
		},
	}

	return s.sendPayload(payload)
}

func (s *SlackNotifier) sendPayload(payload map[string]interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	resp, err := s.client.Post(s.webhookURL, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// DingTalkNotifier sends notifications to DingTalk
type DingTalkNotifier struct {
	webhookURL string
	client     *http.Client
}

// NewDingTalkNotifier creates a new DingTalk notifier
func NewDingTalkNotifier(webhookURL string) *DingTalkNotifier {
	return &DingTalkNotifier{
		webhookURL: webhookURL,
		client:     &http.Client{Timeout: 10 * time.Second},
	}
}

// Name returns the notifier name
func (d *DingTalkNotifier) Name() string {
	return "DingTalk"
}

// Send sends a notification to DingTalk
func (d *DingTalkNotifier) Send(title, message string, issues []types.Issue) error {
	// Format issues for DingTalk markdown
	var markdownContent strings.Builder
	markdownContent.WriteString(fmt.Sprintf("## %s\n\n", title))
	markdownContent.WriteString(fmt.Sprintf("**%s**\n\n", message))

	if len(issues) > 0 {
		markdownContent.WriteString("### 发现问题\n\n")
		for _, issue := range issues {
			if issue.Severity > types.SeverityInfo {
				continue
			}
			severityEmoji := "🔴"
			if issue.Severity == types.SeverityWarning {
				severityEmoji = "🟡"
			}
			markdownContent.WriteString(fmt.Sprintf("%s **%s**: %s\n\n", severityEmoji, issue.CNName, issue.Details))
		}
	}

	payload := map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]string{
			"title": title,
			"text":  markdownContent.String(),
		},
	}

	return d.sendPayload(payload)
}

func (d *DingTalkNotifier) sendPayload(payload map[string]interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	resp, err := d.client.Post(d.webhookURL, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// WeChatNotifier sends notifications to WeChat Work
type WeChatNotifier struct {
	webhookURL string
	client     *http.Client
}

// NewWeChatNotifier creates a new WeChat Work notifier
func NewWeChatNotifier(webhookURL string) *WeChatNotifier {
	return &WeChatNotifier{
		webhookURL: webhookURL,
		client:     &http.Client{Timeout: 10 * time.Second},
	}
}

// Name returns the notifier name
func (w *WeChatNotifier) Name() string {
	return "WeChat Work"
}

// Send sends a notification to WeChat Work
func (w *WeChatNotifier) Send(title, message string, issues []types.Issue) error {
	// Format issues for WeChat markdown
	var content strings.Builder
	content.WriteString(fmt.Sprintf("%s\n\n", title))
	content.WriteString(fmt.Sprintf("%s\n\n", message))

	if len(issues) > 0 {
		content.WriteString("发现问题:\n")
		for _, issue := range issues {
			if issue.Severity > types.SeverityInfo {
				continue
			}
			severityText := "[严重]"
			if issue.Severity == types.SeverityWarning {
				severityText = "[警告]"
			}
			content.WriteString(fmt.Sprintf("%s %s: %s\n", severityText, issue.CNName, issue.Details))
		}
	}

	payload := map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]string{
			"content": content.String(),
		},
	}

	return w.sendPayload(payload)
}

func (w *WeChatNotifier) sendPayload(payload map[string]interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	resp, err := w.client.Post(w.webhookURL, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
