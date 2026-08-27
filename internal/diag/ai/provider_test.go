package ai

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Set test environment variables
	os.Setenv("KUDIG_AI_PROVIDER", "test-provider")
	os.Setenv("KUDIG_AI_MODEL", "gpt-4-test")
	os.Setenv("KUDIG_AI_LANGUAGE", "en")
	defer func() {
		os.Unsetenv("KUDIG_AI_PROVIDER")
		os.Unsetenv("KUDIG_AI_MODEL")
		os.Unsetenv("KUDIG_AI_LANGUAGE")
	}()

	config := LoadConfig()

	if config.Provider != "test-provider" {
		t.Errorf("expected provider to be test-provider, got %s", config.Provider)
	}

	if config.Model != "gpt-4-test" {
		t.Errorf("expected model to be gpt-4-test, got %s", config.Model)
	}

	if config.Language != "en" {
		t.Errorf("expected language to be en, got %s", config.Language)
	}
}

func TestLoadConfig_Defaults(t *testing.T) {
	// Clear environment
	os.Unsetenv("KUDIG_AI_PROVIDER")
	os.Unsetenv("KUDIG_AI_MODEL")
	os.Unsetenv("KUDIG_AI_LANGUAGE")

	config := LoadConfig()

	if config.Provider != "openai" {
		t.Errorf("expected default provider to be openai, got %s", config.Provider)
	}

	if config.Model != "gpt-4" {
		t.Errorf("expected default model to be gpt-4, got %s", config.Model)
	}

	if config.Language != "zh" {
		t.Errorf("expected default language to be zh, got %s", config.Language)
	}
}

func TestNewOpenAIProvider_NoAPIKey(t *testing.T) {
	config := &Config{
		APIKey:   "",
		Provider: "openai",
	}

	_, err := NewOpenAIProvider(config)
	if err == nil {
		t.Error("expected error when API key is empty")
	}
}

func TestNewFactory(t *testing.T) {
	config := &Config{
		Provider: "openai",
		APIKey:   "test-key",
	}

	factory := NewFactory(config)
	if factory == nil {
		t.Fatal("expected factory to not be nil")
	}

	_, err := factory.CreateProvider()
	// Will fail because test-key is invalid, but should create provider
	// We just verify it doesn't panic
	_ = err
}

func TestFactory_NoAPIKey(t *testing.T) {
	config := &Config{
		Provider: "openai",
		APIKey:   "",
	}

	factory := NewFactory(config)
	_, err := factory.CreateProvider()
	if err == nil {
		t.Error("expected error when API key is not configured")
	}
}

func TestOpenAIProvider_Name(t *testing.T) {
	config := &Config{
		APIKey: "test-key",
		Model:  "gpt-4",
	}

	provider, err := NewOpenAIProvider(config)
	if err != nil {
		t.Skipf("Skipping test: %v", err)
	}

	if provider.Name() != "openai" {
		t.Errorf("expected name to be openai, got %s", provider.Name())
	}
}

func TestGetLocalizedMessage(t *testing.T) {
	config := &Config{Language: "zh"}
	factory := NewFactory(config)
	provider, _ := NewOpenAIProvider(&Config{APIKey: "test", Language: "zh"})
	_ = factory

	// We can't easily test this without a real provider, but we can verify it doesn't panic
	msg := provider.getLocalizedMessage("no_issues_found")
	if msg == "" {
		t.Error("expected non-empty message")
	}
}

func TestFactoryMimoDefaults(t *testing.T) {
	config := &Config{Provider: "mimo", APIKey: "sk-test", Model: "gpt-4"}

	factory := NewFactory(config)
	provider, err := factory.CreateProvider()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	openaiProvider, ok := provider.(*OpenAIProvider)
	if !ok {
		t.Fatalf("expected *OpenAIProvider, got %T", provider)
	}
	if openaiProvider.config.BaseURL != "https://api.xiaomimimo.com/v1" {
		t.Errorf("expected mimo default BaseURL for sk- key, got %s", openaiProvider.config.BaseURL)
	}
	if openaiProvider.config.Model != "mimo-v2.5" {
		t.Errorf("expected mimo default model mimo-v2.5, got %s", openaiProvider.config.Model)
	}
}

func TestFactoryMimoTokenPlanEndpoint(t *testing.T) {
	// tp- 前缀为 Token Plan 套餐，需自动路由到专属端点
	config := &Config{Provider: "mimo", APIKey: "tp-test", Model: "gpt-4"}

	provider, err := NewFactory(config).CreateProvider()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	p := provider.(*OpenAIProvider)
	if p.config.BaseURL != "https://token-plan-cn.xiaomimimo.com/v1" {
		t.Errorf("expected token-plan endpoint for tp- key, got %s", p.config.BaseURL)
	}
}

func TestFactoryMimoExplicitConfigWins(t *testing.T) {
	config := &Config{
		Provider: "MIMO", // 大小写不敏感
		APIKey:   "tp-test",
		BaseURL:  "https://proxy.example.com/v1",
		Model:    "mimo-v2.5-pro",
	}

	provider, err := NewFactory(config).CreateProvider()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	p := provider.(*OpenAIProvider)
	if p.config.BaseURL != "https://proxy.example.com/v1" || p.config.Model != "mimo-v2.5-pro" {
		t.Errorf("explicit BaseURL/Model should win, got %s / %s", p.config.BaseURL, p.config.Model)
	}
}

func TestFactoryMimoNoAPIKey(t *testing.T) {
	config := &Config{Provider: "mimo", APIKey: ""}
	_, err := NewFactory(config).CreateProvider()
	if err == nil {
		t.Error("expected error when mimo API key is empty")
	}
}

func TestFactoryOllamaNoAPIKeyAllowed(t *testing.T) {
	config := &Config{Provider: "ollama", APIKey: "", Model: "gpt-4"}

	provider, err := NewFactory(config).CreateProvider()
	if err != nil {
		t.Fatalf("ollama should work without API key: %v", err)
	}
	p := provider.(*OpenAIProvider)
	if p.config.BaseURL != "http://localhost:11434/v1" {
		t.Errorf("expected ollama default BaseURL, got %s", p.config.BaseURL)
	}
	if p.config.Model != "llama3" {
		t.Errorf("expected ollama default model llama3, got %s", p.config.Model)
	}
}

func TestParseAnalysisResponse_FencedChineseJSON(t *testing.T) {
	p := &OpenAIProvider{config: &Config{Language: "zh"}}
	// MiMo 实测会返回 Markdown 围栏 + 中文键名的 JSON
	content := "```json\n{\n  \"摘要\": \"kubelet 10250 未监听\",\n  \"根因分析\": \"kubelet 服务异常\",\n  \"修复建议\": [\n    {\"标题\": \"重启 kubelet\", \"命令\": \"systemctl restart kubelet\", \"风险等级\": \"低\"}\n  ],\n  \"置信度\": 0.95\n}\n```"

	res, err := p.parseAnalysisResponse(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Summary != "kubelet 10250 未监听" {
		t.Errorf("summary not mapped, got %q", res.Summary)
	}
	if res.RootCause != "kubelet 服务异常" {
		t.Errorf("root cause not mapped, got %q", res.RootCause)
	}
	if len(res.Suggestions) != 1 || res.Suggestions[0].Command != "systemctl restart kubelet" {
		t.Errorf("suggestions not mapped, got %+v", res.Suggestions)
	}
	if res.Confidence < 0.94 || res.Confidence > 0.96 {
		t.Errorf("confidence not mapped, got %v", res.Confidence)
	}
}

func TestParseAnalysisResponse_PlainTextFallback(t *testing.T) {
	p := &OpenAIProvider{config: &Config{Language: "en"}}
	res, err := p.parseAnalysisResponse("just plain text analysis")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Summary != "just plain text analysis" || res.Confidence != 0.5 {
		t.Errorf("fallback broken, got %+v", res)
	}
}

func TestGetEnvIntInvalidFallsBackToDefault(t *testing.T) {
	os.Setenv("KUDIG_AI_TIMEOUT", "30s") // 非法整数
	defer os.Unsetenv("KUDIG_AI_TIMEOUT")

	if got := getEnvInt("KUDIG_AI_TIMEOUT", 30); got != 30 {
		t.Errorf("invalid int should fall back to default 30, got %d", got)
	}

	os.Setenv("KUDIG_AI_TEMPERATURE", "abc")
	defer os.Unsetenv("KUDIG_AI_TEMPERATURE")
	if got := getEnvFloat("KUDIG_AI_TEMPERATURE", 0.3); got != 0.3 {
		t.Errorf("invalid float should fall back to default 0.3, got %v", got)
	}
}
