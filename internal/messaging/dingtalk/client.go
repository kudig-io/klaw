package dingtalk

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/kudig-io/klaw/internal/config"
	"github.com/kudig-io/klaw/internal/kubernetes"
)

// Client 钉钉客户端
type Client struct {
	config      config.DingTalkConfig
	k8sManager  *kubernetes.Manager
	webhook     string
	secret      string
	accessToken string
}

// NewClient 创建钉钉客户端
func NewClient(cfg config.DingTalkConfig) (*Client, error) {
	return &Client{
		config:  cfg,
		webhook: cfg.Webhook,
		secret:  cfg.Secret,
	}, nil
}

// SetK8sManager 设置Kubernetes管理器
func (c *Client) SetK8sManager(manager *kubernetes.Manager) {
	c.k8sManager = manager
}

// Start 启动钉钉客户端
func (c *Client) Start() {
	// 这里可以实现webhook服务器，接收钉钉消息
	// 为了简单起见，我们先实现发送消息的功能
	fmt.Println("DingTalk client started")
}

// SendMessage 发送消息到钉钉
func (c *Client) SendMessage(message string) error {
	// 生成签名
	timestamp := time.Now().UnixMilli()
	signature := c.generateSignature(timestamp)

	// 构建请求URL
	requestURL := fmt.Sprintf("%s&timestamp=%d&sign=%s", c.webhook, timestamp, signature)

	// 构建请求体
	requestBody := map[string]interface{}{
		"msgtype": "text",
		"text": map[string]string{
			"content": message,
		},
	}

	// 编码请求体
	data, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %v", err)
	}

	// 发送请求
	resp, err := http.Post(requestURL, "application/json", nil)
	if err != nil {
		return fmt.Errorf("failed to send message: %v", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send message, status code: %d", resp.StatusCode)
	}

	return nil
}

// generateSignature 生成签名
func (c *Client) generateSignature(timestamp int64) string {
	// 构建签名字符串
	stringToSign := fmt.Sprintf("%d\n%s", timestamp, c.secret)

	// 计算HMAC-SHA256
	h := hmac.New(sha256.New, []byte(c.secret))
	h.Write([]byte(stringToSign))

	// 进行base64编码
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))

	// URL编码
	signature = url.QueryEscape(signature)

	return signature
}

// HandleMessage 处理接收到的消息
func (c *Client) HandleMessage(message string) (string, error) {
	// 这里可以实现消息处理逻辑
	// 解析命令，执行操作，返回结果
	return "收到消息: " + message, nil
}

// SendImage 发送图片到钉钉
func (c *Client) SendImage(imageData []byte, message string) error {
	// 生成签名
	timestamp := time.Now().UnixMilli()
	signature := c.generateSignature(timestamp)

	// 构建请求URL
	requestURL := fmt.Sprintf("%s&timestamp=%d&sign=%s", c.webhook, timestamp, signature)

	// 构建请求体
	requestBody := map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]string{
			"title": "Kubernetes监控图表",
			"text":  fmt.Sprintf("## %s\n\n![监控图表](data:image/png;base64,%s)", message, base64.StdEncoding.EncodeToString(imageData)),
		},
	}

	// 编码请求体
	data, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %v", err)
	}

	// 发送请求
	resp, err := http.Post(requestURL, "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to send image: %v", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send image, status code: %d", resp.StatusCode)
	}

	return nil
}

// SendChart 发送图表到钉钉
func (c *Client) SendChart(chartData []byte, title string) error {
	return c.SendImage(chartData, title)
}
