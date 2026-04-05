package dingtalk

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/kudig-io/klaw/internal/messaging"
)

const (
	PlatformName = "dingtalk"
)

// Config 钉钉配置
type Config struct {
	Enabled      bool   `yaml:"enabled" json:"enabled"`
	AppKey       string `yaml:"app_key" json:"app_key"`
	AppSecret    string `yaml:"app_secret" json:"app_secret"`
	Webhook      string `yaml:"webhook" json:"webhook"`
	Secret       string `yaml:"secret" json:"secret"`
	WebhookPort  int    `yaml:"webhook_port" json:"webhook_port"` // 接收消息的端口
}

// Plugin 钉钉通信插件
type Plugin struct {
	config      Config
	handlers    []messaging.MessageHandler
	server      *http.Server
	webhookPath string
	mu          sync.RWMutex
	running     bool
}

// NewPlugin 创建钉钉插件
func NewPlugin(config Config) *Plugin {
	return &Plugin{
		config:      config,
		handlers:    make([]messaging.MessageHandler, 0),
		webhookPath: "/webhook/dingtalk",
	}
}

// Name 返回平台名称
func (p *Plugin) Name() string {
	return PlatformName
}

// Start 启动钉钉插件
func (p *Plugin) Start() error {
	if !p.config.Enabled {
		return nil
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if p.running {
		return nil
	}

	// 创建 HTTP 服务器接收钉钉消息
	mux := http.NewServeMux()
	mux.HandleFunc(p.webhookPath, p.handleWebhook)

	port := p.config.WebhookPort
	if port == 0 {
		port = 8081 // 默认端口
	}

	p.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	p.running = true

	// 在后台启动服务器
	go func() {
		if err := p.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("DingTalk webhook server error: %v\n", err)
		}
	}()

	fmt.Printf("DingTalk plugin started, webhook: http://localhost:%d%s\n", port, p.webhookPath)
	return nil
}

// Stop 停止钉钉插件
func (p *Plugin) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.running {
		return nil
	}

	p.running = false

	if p.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return p.server.Shutdown(ctx)
	}

	return nil
}

// SendMessage 发送消息到钉钉
func (p *Plugin) SendMessage(channelID string, response *messaging.Response) error {
	if !p.config.Enabled {
		return fmt.Errorf("dingtalk plugin is not enabled")
	}

	switch response.Format {
	case messaging.FormatMarkdown:
		return p.sendMarkdownMessage(response.Content)
	case messaging.FormatImage:
		return p.sendImageMessage(response.Content)
	default:
		return p.sendTextMessage(response.Content)
	}
}

// RegisterHandler 注册消息处理器
func (p *Plugin) RegisterHandler(handler messaging.MessageHandler) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.handlers = append(p.handlers, handler)
}

// IsHealthy 检查健康状态
func (p *Plugin) IsHealthy() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.running
}

// DingTalkWebhookMessage 钉钉 Webhook 消息结构
type DingTalkWebhookMessage struct {
	MsgType string `json:"msgtype"`
	Text    struct {
		Content string `json:"content"`
	} `json:"text"`
	SenderStaffID   string `json:"senderStaffId"`
	SenderNick      string `json:"senderNick"`
	ConversationID  string `json:"conversationId"`
	ConversationTitle string `json:"conversationTitle"`
	CreateAt        int64  `json:"createAt"`
	ChatbotUserID   string `json:"chatbotUserId"`
	AtUsers         []struct {
		StaffID string `json:"staffId"`
	} `json:"atUsers"`
	IsAdmin    bool `json:"isAdmin"`
	IsInAtList bool `json:"isInAtList"`
}

// handleWebhook 处理钉钉 Webhook 回调
func (p *Plugin) handleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 验证签名（如果配置了 Secret）
	if p.config.Secret != "" {
		timestamp := r.Header.Get("timestamp")
		sign := r.Header.Get("sign")
		
		if !p.verifySignature(timestamp, sign) {
			http.Error(w, "Invalid signature", http.StatusUnauthorized)
			return
		}
	}

	// 解析消息
	var dingMsg DingTalkWebhookMessage
	if err := json.NewDecoder(r.Body).Decode(&dingMsg); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// 转换为内部消息格式
	msg := &messaging.Message{
		ID:          fmt.Sprintf("%d", dingMsg.CreateAt),
		Content:     strings.TrimSpace(dingMsg.Text.Content),
		SenderID:    dingMsg.SenderStaffID,
		SenderName:  dingMsg.SenderNick,
		ChannelID:   dingMsg.ConversationID,
		ChannelName: dingMsg.ConversationTitle,
		Timestamp:   dingMsg.CreateAt,
		IsGroup:     dingMsg.ConversationID != "",
		Mentioned:   dingMsg.IsInAtList,
	}

	// 异步处理消息
	go p.processMessage(msg)

	// 立即返回成功响应（钉钉要求 200ms 内响应）
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// processMessage 处理消息
func (p *Plugin) processMessage(msg *messaging.Message) {
	ctx := context.Background()
	
	p.mu.RLock()
	handlers := make([]messaging.MessageHandler, len(p.handlers))
	copy(handlers, p.handlers)
	p.mu.RUnlock()

	for _, handler := range handlers {
		response, err := handler(ctx, msg)
		if err != nil {
			// 发送错误响应
			p.sendTextMessage(fmt.Sprintf("❌ 执行出错: %v", err))
			continue
		}
		
		if response != nil {
			// 发送响应
			if err := p.SendMessage(msg.ChannelID, response); err != nil {
				fmt.Printf("Failed to send response: %v\n", err)
			}
		}
	}
}

// verifySignature 验证签名
func (p *Plugin) verifySignature(timestamp, sign string) bool {
	expectedSign := p.generateSignature(timestamp)
	return sign == expectedSign
}

// generateSignature 生成签名
func (p *Plugin) generateSignature(timestamp string) string {
	stringToSign := timestamp + "\n" + p.config.Secret
	h := hmac.New(sha256.New, []byte(p.config.Secret))
	h.Write([]byte(stringToSign))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// sendTextMessage 发送文本消息
func (p *Plugin) sendTextMessage(content string) error {
	return p.sendToWebhook(map[string]interface{}{
		"msgtype": "text",
		"text": map[string]string{
			"content": content,
		},
	})
}

// sendMarkdownMessage 发送 Markdown 消息
func (p *Plugin) sendMarkdownMessage(content string) error {
	return p.sendToWebhook(map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]string{
			"title": "Klaw 消息",
			"text":  content,
		},
	})
}

// sendImageMessage 发送图片消息
func (p *Plugin) sendImageMessage(base64Image string) error {
	return p.sendToWebhook(map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]string{
			"title": "Klaw 图表",
			"text":  fmt.Sprintf("![图表](%s)", base64Image),
		},
	})
}

// sendToWebhook 发送到钉钉 Webhook
func (p *Plugin) sendToWebhook(payload map[string]interface{}) error {
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	signature := p.generateSignature(timestamp)

	// 构建请求 URL
	webhookURL := p.config.Webhook
	if webhookURL == "" {
		return fmt.Errorf("webhook URL is not configured")
	}

	requestURL := fmt.Sprintf("%s&timestamp=%s&sign=%s", webhookURL, timestamp, url.QueryEscape(signature))

	// 编码请求体
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %v", err)
	}

	// 发送请求
	resp, err := http.Post(requestURL, "application/json", strings.NewReader(string(data)))
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// Factory 钉钉插件工厂
type Factory struct{}

// Create 创建钉钉插件
func (f *Factory) Create(config interface{}) (messaging.Communicator, error) {
	cfg, ok := config.(Config)
	if !ok {
		return nil, fmt.Errorf("invalid config type for dingtalk")
	}
	return NewPlugin(cfg), nil
}

// SupportedTypes 返回支持的类型
func (f *Factory) SupportedTypes() []string {
	return []string{PlatformName}
}
