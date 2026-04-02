package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config 应用配置结构
type Config struct {
	Kubernetes KubernetesConfig `yaml:"kubernetes"`
	Messaging  MessagingConfig  `yaml:"messaging"`
	OpenClaw   OpenClawConfig   `yaml:"openclaw"`
	Server     ServerConfig     `yaml:"server"`
	Events     EventConfig      `yaml:"events"`
}

// KubernetesConfig Kubernetes配置
type KubernetesConfig struct {
	Clusters []ClusterConfig `yaml:"clusters"`
}

// ClusterConfig 集群配置
type ClusterConfig struct {
	Name       string `yaml:"name" json:"name"`
	Kubeconfig string `yaml:"kubeconfig" json:"kubeconfig"`
	Context    string `yaml:"context" json:"context"`
}

// MessagingConfig 消息平台配置
type MessagingConfig struct {
	DingTalk DingTalkConfig `yaml:"dingtalk"`
	Feishu   FeishuConfig   `yaml:"feishu"`
}

// DingTalkConfig 钉钉配置
type DingTalkConfig struct {
	Enabled      bool   `yaml:"enabled"`
	AppKey       string `yaml:"app_key"`
	AppSecret    string `yaml:"app_secret"`
	Webhook      string `yaml:"webhook"`
	Secret       string `yaml:"secret"`
	WebhookPort  int    `yaml:"webhook_port"`
}

// FeishuConfig 飞书配置
type FeishuConfig struct {
	Enabled   bool   `yaml:"enabled"`
	AppID     string `yaml:"app_id"`
	AppSecret string `yaml:"app_secret"`
}

// OpenClawConfig OpenClaw配置
type OpenClawConfig struct {
	Enabled bool   `yaml:"enabled"`
	Skills  string `yaml:"skills"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port int `yaml:"port"`
}

// EventConfig 事件监听配置
type EventConfig struct {
	Enabled        bool          `yaml:"enabled"`
	WatchTypes     []string      `yaml:"watch_types"`     // 监听的资源类型: pod, deployment, service
	Namespaces     []string      `yaml:"namespaces"`      // 监听的命名空间，为空表示所有
	EventTypes     []string      `yaml:"event_types"`     // 监听的事件类型: Normal, Warning, Error
	Reasons        []string      `yaml:"reasons"`         // 关注的原因
	ExcludeReasons []string      `yaml:"exclude_reasons"` // 排除的原因
	MinSeverity    string        `yaml:"min_severity"`    // 最小严重级别: info, warning, critical
	RateLimit      int           `yaml:"rate_limit"`      // 每秒最大事件数
	DedupWindow    int           `yaml:"dedup_window"`    // 去重窗口（秒）
	MuteDuration   int           `yaml:"mute_duration"`   // 相同事件静音时长（分钟）
	Channels       []string      `yaml:"channels"`        // 推送的频道列表
}

// Load 加载配置文件
func Load(path string) (*Config, error) {
	// 检查配置文件是否存在
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found: %s", path)
	}

	// 读取配置文件
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	// 解析配置文件
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config file: %v", err)
	}

	// 设置默认值
	if config.Server.Port == 0 {
		config.Server.Port = 8080
	}

	return &config, nil
}
