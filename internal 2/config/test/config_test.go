package config_test

import (
	"testing"

	"github.com/kudig-io/klaw/internal/config"
)

func TestLoadConfig(t *testing.T) {
	// 测试加载配置文件
	cfg, err := config.Load("../../../configs/config.yaml.example")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// 验证配置是否正确加载
	if cfg.Server.Port != 8080 {
		t.Errorf("Expected server port 8080, got %d", cfg.Server.Port)
	}

	if len(cfg.Kubernetes.Clusters) == 0 {
		t.Error("Expected at least one cluster in config")
	}
}
