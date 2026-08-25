package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestSOSDefaults(t *testing.T) {
	p := writeTemp(t, "sos:\n  enabled: true\n")
	cfg, err := Load(p)
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.SOS.Enabled {
		t.Fatal("expected sos.enabled=true")
	}
	if cfg.SOS.Dashscope.Region != "cn-beijing" {
		t.Fatalf("expected default region cn-beijing, got %s", cfg.SOS.Dashscope.Region)
	}
	if cfg.SOS.Dashscope.Model != "qwen3.5-omni-plus-realtime" {
		t.Fatalf("expected default model, got %s", cfg.SOS.Dashscope.Model)
	}
	if cfg.SOS.Dashscope.Voice != "Ethan" {
		t.Fatalf("expected default voice Ethan, got %s", cfg.SOS.Dashscope.Voice)
	}
}

func TestSOSEnvOverride(t *testing.T) {
	p := writeTemp(t, "sos:\n  enabled: true\n  dashscope:\n    api_key: from-file\n")
	t.Setenv("KLAW_SOS_DASHSCOPE_API_KEY", "from-env")
	cfg, err := Load(p)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.SOS.Dashscope.APIKey != "from-env" {
		t.Fatalf("expected env override, got %s", cfg.SOS.Dashscope.APIKey)
	}
}

func TestSOSDisabledByDefault(t *testing.T) {
	cfg, err := Load(writeTemp(t, "{}\n"))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.SOS.Enabled {
		t.Fatal("expected sos disabled by default")
	}
}
