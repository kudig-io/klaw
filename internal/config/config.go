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
}

// KubernetesConfig Kubernetes配置
type KubernetesConfig struct {
	Clusters []ClusterConfig `yaml:"clusters"`
}

// ClusterConfig 集群配置
type ClusterConfig struct {
	Name      string `yaml:"name"`
	Kubeconfig string `yaml:"kubeconfig"`
	Context    string `yaml:"context"`
}

// MessagingConfig 消息平台配置
type MessagingConfig struct {
	DingTalk DingTalkConfig `yaml:"dingtalk"`
	Feishu   FeishuConfig   `yaml:"feishu"`
}

// DingTalkConfig 钉钉配置
type DingTalkConfig struct {
	Enabled bool   `yaml:"enabled"`
	AppKey  string `yaml:"app_key"`
	AppSecret string `yaml:"app_secret"`
	Webhook   string `yaml:"webhook"`
	Secret    string `yaml:"secret"`
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
