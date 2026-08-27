package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config 应用配置结构
type Config struct {
	Kubernetes KubernetesConfig `yaml:"kubernetes"`
	Messaging  MessagingConfig  `yaml:"messaging"`
	OpenClaw   OpenClawConfig   `yaml:"openclaw"`
	Server     ServerConfig     `yaml:"server"`
	Events     EventConfig      `yaml:"events"`
	SOS        SOSConfig        `yaml:"sos"`
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
	Enabled     bool   `yaml:"enabled"`
	AppKey      string `yaml:"app_key"`
	AppSecret   string `yaml:"app_secret"`
	Webhook     string `yaml:"webhook"`
	Secret      string `yaml:"secret"`
	WebhookPort int    `yaml:"webhook_port"`
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
	Port int        `yaml:"port"`
	Auth AuthConfig `yaml:"auth"`
	CORS CORSConfig `yaml:"cors"`
}

// AuthConfig API 认证配置
type AuthConfig struct {
	Enabled bool   `yaml:"enabled"`
	Token   string `yaml:"token"` // 建议通过环境变量 KLAW_API_TOKEN 注入，避免落盘
}

// CORSConfig 跨域配置
type CORSConfig struct {
	AllowedOrigins []string `yaml:"allowed_origins"`
}

// EventConfig 事件监听配置
type EventConfig struct {
	Enabled        bool     `yaml:"enabled"`
	WatchTypes     []string `yaml:"watch_types"`     // 监听的资源类型: pod, deployment, service
	Namespaces     []string `yaml:"namespaces"`      // 监听的命名空间，为空表示所有
	EventTypes     []string `yaml:"event_types"`     // 监听的事件类型: Normal, Warning, Error
	Reasons        []string `yaml:"reasons"`         // 关注的原因
	ExcludeReasons []string `yaml:"exclude_reasons"` // 排除的原因
	MinSeverity    string   `yaml:"min_severity"`    // 最小严重级别: info, warning, critical
	RateLimit      int      `yaml:"rate_limit"`      // 每秒最大事件数
	DedupWindow    int      `yaml:"dedup_window"`    // 去重窗口（秒）
	MuteDuration   int      `yaml:"mute_duration"`   // 相同事件静音时长（分钟）
	Channels       []string `yaml:"channels"`        // 推送的频道列表
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
	if config.SOS.Dashscope.Region == "" {
		config.SOS.Dashscope.Region = "cn-beijing"
	}
	if config.SOS.Dashscope.Model == "" {
		config.SOS.Dashscope.Model = "qwen3.5-omni-plus-realtime"
	}
	if config.SOS.Dashscope.Voice == "" {
		config.SOS.Dashscope.Voice = "Ethan"
	}

	// 环境变量覆盖（敏感信息不落盘，由 Secret 注入）
	applyEnvOverrides(&config)

	return &config, nil
}

// applyEnvOverrides 从环境变量覆盖敏感配置（优先级高于配置文件）
func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("KLAW_API_TOKEN"); v != "" {
		cfg.Server.Auth.Token = v
	}
	if v := os.Getenv("KLAW_DINGTALK_APP_KEY"); v != "" {
		cfg.Messaging.DingTalk.AppKey = v
	}
	if v := os.Getenv("KLAW_DINGTALK_APP_SECRET"); v != "" {
		cfg.Messaging.DingTalk.AppSecret = v
	}
	if v := os.Getenv("KLAW_DINGTALK_WEBHOOK"); v != "" {
		cfg.Messaging.DingTalk.Webhook = v
	}
	if v := os.Getenv("KLAW_DINGTALK_SECRET"); v != "" {
		cfg.Messaging.DingTalk.Secret = v
	}
	if v := os.Getenv("KLAW_FEISHU_APP_ID"); v != "" {
		cfg.Messaging.Feishu.AppID = v
	}
	if v := os.Getenv("KLAW_FEISHU_APP_SECRET"); v != "" {
		cfg.Messaging.Feishu.AppSecret = v
	}
	if v := os.Getenv("KLAW_SOS_DASHSCOPE_API_KEY"); v != "" {
		cfg.SOS.Dashscope.APIKey = v
	}
	if v := os.Getenv("KLAW_SOS_GLM_API_KEY"); v != "" {
		cfg.SOS.GLM.APIKey = v
	}
	// provider 归一化：大小写不敏感，缺省为 dashscope
	switch strings.ToLower(cfg.SOS.Provider) {
	case "", "dashscope":
		cfg.SOS.Provider = "dashscope"
	case "glm":
		cfg.SOS.Provider = "glm"
	}
	if cfg.SOS.Provider == "glm" && cfg.SOS.GLM.Model == "" {
		cfg.SOS.GLM.Model = "glm-realtime"
	}
}

// SOSConfig SOS 语音应急快速对话配置
type SOSConfig struct {
	Enabled            bool               `yaml:"enabled"`
	Provider           string             `yaml:"provider"` // dashscope（默认）| glm
	Dashscope          SOSDashscopeConfig `yaml:"dashscope"`
	GLM                SOSGlmConfig       `yaml:"glm"`
	FAQFile            string             `yaml:"faq_file"`            // 外部语料路径；为空时使用内嵌默认语料
	InstructionsPrefix string             `yaml:"instructions_prefix"` // 追加在默认系统提示之前
}

// SOSDashscopeConfig 阿里云百炼 DashScope Realtime 配置
type SOSDashscopeConfig struct {
	APIKey      string `yaml:"api_key"`      // 建议通过环境变量 KLAW_SOS_DASHSCOPE_API_KEY 注入，避免落盘
	WorkspaceID string `yaml:"workspace_id"` // 百炼 Workspace ID（端点子域名）
	Region      string `yaml:"region"`       // cn-beijing / ap-southeast-1
	Model       string `yaml:"model"`
	Voice       string `yaml:"voice"`
}

// SOSGlmConfig 智谱 GLM-Realtime 配置
type SOSGlmConfig struct {
	APIKey string `yaml:"api_key"` // 形如 {id}.{secret}；建议通过环境变量 KLAW_SOS_GLM_API_KEY 注入
	Model  string `yaml:"model"`
	Voice  string `yaml:"voice"` // 可选；GLM-Realtime 无标准音色字段，非空时随会话下发
}
