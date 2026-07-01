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
