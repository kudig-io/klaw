package feishu

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/kudig-io/klaw/internal/config"
	"github.com/kudig-io/klaw/internal/kubernetes"
)

// Client 飞书客户端
type Client struct {
	config      config.FeishuConfig
	k8sManager  *kubernetes.Manager
	appID       string
	appSecret   string
	accessToken string
	tokenExpiry time.Time
}

// NewClient 创建飞书客户端
func NewClient(cfg config.FeishuConfig) (*Client, error) {
	return &Client{
		config:    cfg,
		appID:     cfg.AppID,
		appSecret: cfg.AppSecret,
	}, nil
}

// SetK8sManager 设置Kubernetes管理器
func (c *Client) SetK8sManager(manager *kubernetes.Manager) {
	c.k8sManager = manager
}

// Start 启动飞书客户端
func (c *Client) Start() {
	// 这里可以实现webhook服务器，接收飞书消息
	// 为了简单起见，我们先实现发送消息的功能
	fmt.Println("Feishu client started")

	// 初始化获取access token
	if err := c.refreshAccessToken(); err != nil {
		fmt.Printf("Failed to refresh access token: %v\n", err)
	}
}

// refreshAccessToken 刷新access token
func (c *Client) refreshAccessToken() error {
	// 构建请求URL
	url := "https://open.feishu.cn/open-apis/auth/v3/app_access_token/internal"

	// 构建请求体
	requestBody := map[string]string{
		"app_id":     c.appID,
		"app_secret": c.appSecret,
	}

	// 编码请求体
	data, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %v", err)
	}

	// 发送请求
	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		return fmt.Errorf("failed to refresh access token: %v", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to refresh access token, status code: %d", resp.StatusCode)
	}

	// 解析响应
	var response struct {
		Code int `json:"code"`
		Data struct {
			AppAccessToken string `json:"app_access_token"`
			Expire         int    `json:"expire"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	if response.Code != 0 {
		return fmt.Errorf("failed to refresh access token, code: %d", response.Code)
	}

	// 更新access token和过期时间
	c.accessToken = response.Data.AppAccessToken
	c.tokenExpiry = time.Now().Add(time.Duration(response.Data.Expire-300) * time.Second)

	return nil
}

// getAccessToken 获取access token
func (c *Client) getAccessToken() (string, error) {
	// 检查access token是否过期
	if time.Now().After(c.tokenExpiry) {
		if err := c.refreshAccessToken(); err != nil {
			return "", err
		}
	}

	return c.accessToken, nil
}

// SendMessage 发送消息到飞书
func (c *Client) SendMessage(message string) error {
	// 获取access token
	token, err := c.getAccessToken()
	if err != nil {
		return err
	}

	// 构建请求URL
	url := "https://open.feishu.cn/open-apis/im/v1/messages"

	// 构建请求头
	headers := map[string]string{
		"Authorization": "Bearer " + token,
		"Content-Type":  "application/json",
	}

	// 构建请求体
	requestBody := map[string]interface{}{
		"receive_id_type": "chat_id",
		"receive_id":      "oc_abcdefg", // 这里需要替换为实际的聊天ID
		"content":         fmt.Sprintf(`{\"text\":\"%s\"}`, message),
		"msg_type":        "text",
	}

	// 编码请求体
	data, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %v", err)
	}

	// 创建请求
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	// 设置请求头
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
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

// HandleMessage 处理接收到的消息
func (c *Client) HandleMessage(message string) (string, error) {
	// 这里可以实现消息处理逻辑
	// 解析命令，执行操作，返回结果
	return "收到消息: " + message, nil
}

// SendImage 发送图片到飞书
func (c *Client) SendImage(imageData []byte, message string) error {
	// 获取access token
	token, err := c.getAccessToken()
	if err != nil {
		return err
	}

	// 构建请求URL
	url := "https://open.feishu.cn/open-apis/im/v1/messages"

	// 构建请求头
	headers := map[string]string{
		"Authorization": "Bearer " + token,
		"Content-Type":  "application/json",
	}

	// 构建请求体
	requestBody := map[string]interface{}{
		"receive_id_type": "chat_id",
		"receive_id":      "oc_abcdefg",
		"msg_type":       "post",
		"content": map[string]interface{}{
			"post": map[string]interface{}{
				"zh_cn": map[string]interface{}{
					"title":   "Kubernetes监控图表",
					"content": [][]map[string]interface{}{
						{
							{
								"tag":  "text",
								"text": message,
							},
						},
						{
							{
								"tag":      "img",
								"img_key":  base64.StdEncoding.EncodeToString(imageData),
								"alt": map[string]interface{}{
									"tag":     "plain_text",
									"content": "监控图表",
								},
							},
						},
					},
				},
			},
		},
	}

	// 编码请求体
	data, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %v", err)
	}

	// 创建请求
	req, err := http.NewRequest("POST", url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	// 设置请求头
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
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

// SendChart 发送图表到飞书
func (c *Client) SendChart(chartData []byte, title string) error {
	return c.SendImage(chartData, title)
}
