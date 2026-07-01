// Package ai provides AI/LLM integration for diagnostic assistance
package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/kudig-io/klaw/internal/diag/types"
	"github.com/sashabaranov/go-openai"
)

// Provider defines the interface for AI providers
type Provider interface {
	// Analyze analyzes issues and returns insights
	Analyze(ctx context.Context, issues []types.Issue, hostname string) (*AnalysisResult, error)
	
	// GenerateSummary generates a summary of diagnostic results
	GenerateSummary(ctx context.Context, issues []types.Issue) (string, error)
	
	// SuggestFixes suggests fixes for issues
	SuggestFixes(ctx context.Context, issue types.Issue) ([]FixSuggestion, error)
	
	// Name returns the provider name
	Name() string
}

// AnalysisResult contains the AI analysis result
type AnalysisResult struct {
	Summary      string           `json:"summary"`
	RootCause    string           `json:"root_cause"`
	Suggestions  []FixSuggestion  `json:"suggestions"`
	Severity     types.Severity   `json:"severity"`
	Confidence   float64          `json:"confidence"`
	Language     string           `json:"language"`
}

// FixSuggestion contains a suggested fix
type FixSuggestion struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Command     string `json:"command,omitempty"`
	Risk        string `json:"risk"` // low, medium, high
	AutoFix     bool   `json:"auto_fix"`
}

// Config holds AI provider configuration
type Config struct {
	Provider    string `env:"KUDIG_AI_PROVIDER" default:"openai"` // openai, qwen, ollama
	APIKey      string `env:"KUDIG_AI_API_KEY"`
	BaseURL     string `env:"KUDIG_AI_BASE_URL"` // for custom endpoints like Ollama
	Model       string `env:"KUDIG_AI_MODEL" default:"gpt-4"`
	Timeout     int    `env:"KUDIG_AI_TIMEOUT" default:"30"`
	Language    string `env:"KUDIG_AI_LANGUAGE" default:"zh"` // zh or en
	MaxTokens   int    `env:"KUDIG_AI_MAX_TOKENS" default:"2000"`
	Temperature float64 `env:"KUDIG_AI_TEMPERATURE" default:"0.3"`
}

// LoadConfig loads AI configuration from environment
func LoadConfig() *Config {
	return &Config{
		Provider:    getEnv("KUDIG_AI_PROVIDER", "openai"),
		APIKey:      getEnv("KUDIG_AI_API_KEY", ""),
		BaseURL:     getEnv("KUDIG_AI_BASE_URL", ""),
		Model:       getEnv("KUDIG_AI_MODEL", "gpt-4"),
		Timeout:     getEnvInt("KUDIG_AI_TIMEOUT", 30),
		Language:    getEnv("KUDIG_AI_LANGUAGE", "zh"),
		MaxTokens:   getEnvInt("KUDIG_AI_MAX_TOKENS", 2000),
		Temperature: getEnvFloat("KUDIG_AI_TEMPERATURE", 0.3),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var result int
		fmt.Sscanf(value, "%d", &result)
		return result
	}
	return defaultValue
}

func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		var result float64
		fmt.Sscanf(value, "%f", &result)
		return result
	}
	return defaultValue
}

// OpenAIProvider implements Provider using OpenAI API
type OpenAIProvider struct {
	client *openai.Client
	config *Config
}

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider(config *Config) (*OpenAIProvider, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("AI API key not configured")
	}

	clientConfig := openai.DefaultConfig(config.APIKey)
	if config.BaseURL != "" {
		clientConfig.BaseURL = config.BaseURL
	}

	return &OpenAIProvider{
		client: openai.NewClientWithConfig(clientConfig),
		config: config,
	}, nil
}

// Name returns the provider name
func (p *OpenAIProvider) Name() string {
	return "openai"
}

// Analyze analyzes issues using OpenAI
func (p *OpenAIProvider) Analyze(ctx context.Context, issues []types.Issue, hostname string) (*AnalysisResult, error) {
	if len(issues) == 0 {
		return &AnalysisResult{
			Summary:     p.getLocalizedMessage("no_issues_found"),
			RootCause:   "",
			Suggestions: []FixSuggestion{},
			Severity:    types.SeverityInfo,
			Confidence:  1.0,
			Language:    p.config.Language,
		}, nil
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(p.config.Timeout)*time.Second)
	defer cancel()

	// Prepare prompt
	prompt := p.buildAnalysisPrompt(issues, hostname)

	// Create chat completion
	resp, err := p.client.CreateChatCompletion(timeoutCtx, openai.ChatCompletionRequest{
		Model:       p.config.Model,
		MaxTokens:   p.config.MaxTokens,
		Temperature: float32(p.config.Temperature),
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: p.getSystemPrompt(),
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get AI analysis: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from AI")
	}

	// Parse response
	return p.parseAnalysisResponse(resp.Choices[0].Message.Content)
}

// GenerateSummary generates a summary
func (p *OpenAIProvider) GenerateSummary(ctx context.Context, issues []types.Issue) (string, error) {
	if len(issues) == 0 {
		return p.getLocalizedMessage("system_healthy"), nil
	}

	prompt := p.buildSummaryPrompt(issues)

	resp, err := p.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:       p.config.Model,
		MaxTokens:   500,
		Temperature: 0.3,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: p.getLocalizedMessage("summary_system_prompt"),
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
	})
	if err != nil {
		return "", err
	}

	if len(resp.Choices) > 0 {
		return resp.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("no response")
}

// SuggestFixes suggests fixes for an issue
func (p *OpenAIProvider) SuggestFixes(ctx context.Context, issue types.Issue) ([]FixSuggestion, error) {
	prompt := p.buildFixPrompt(issue)

	resp, err := p.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:       p.config.Model,
		MaxTokens:   1000,
		Temperature: 0.2,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: p.getLocalizedMessage("fix_system_prompt"),
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
	})
	if err != nil {
		return nil, err
	}

	if len(resp.Choices) > 0 {
		return p.parseFixResponse(resp.Choices[0].Message.Content)
	}

	return nil, fmt.Errorf("no response")
}

func (p *OpenAIProvider) getSystemPrompt() string {
	if p.config.Language == "zh" {
		return `你是一位 Kubernetes 诊断专家。分析诊断结果并提供：
1. 问题摘要（中文）
2. 根因分析
3. 修复建议（包含风险等级）
4. 置信度（0-1）

请以 JSON 格式返回结果。`
	}
	return `You are a Kubernetes diagnostic expert. Analyze the diagnostic results and provide:
1. Issue summary (English)
2. Root cause analysis
3. Fix suggestions with risk levels
4. Confidence score (0-1)

Please return results in JSON format.`
}

func (p *OpenAIProvider) buildAnalysisPrompt(issues []types.Issue, hostname string) string {
	var sb strings.Builder
	
	if p.config.Language == "zh" {
		sb.WriteString(fmt.Sprintf("主机: %s\n", hostname))
		sb.WriteString(fmt.Sprintf("发现 %d 个问题:\n\n", len(issues)))
	} else {
		sb.WriteString(fmt.Sprintf("Host: %s\n", hostname))
		sb.WriteString(fmt.Sprintf("Found %d issues:\n\n", len(issues)))
	}

	for i, issue := range issues {
		sb.WriteString(fmt.Sprintf("%d. [%s] %s\n", i+1, issue.Severity, issue.CNName))
		sb.WriteString(fmt.Sprintf("   %s\n", issue.Details))
		if issue.Location != "" {
			sb.WriteString(fmt.Sprintf("   Location: %s\n", issue.Location))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func (p *OpenAIProvider) buildSummaryPrompt(issues []types.Issue) string {
	severityCount := make(map[types.Severity]int)
	for _, issue := range issues {
		severityCount[issue.Severity]++
	}

	var sb strings.Builder
	if p.config.Language == "zh" {
		sb.WriteString(fmt.Sprintf("严重: %d, 警告: %d, 信息: %d\n\n",
			severityCount[types.SeverityCritical],
			severityCount[types.SeverityWarning],
			severityCount[types.SeverityInfo]))
		sb.WriteString("请生成一句简洁的中文总结。")
	} else {
		sb.WriteString(fmt.Sprintf("Critical: %d, Warning: %d, Info: %d\n\n",
			severityCount[types.SeverityCritical],
			severityCount[types.SeverityWarning],
			severityCount[types.SeverityInfo]))
		sb.WriteString("Please generate a concise English summary.")
	}

	return sb.String()
}

func (p *OpenAIProvider) buildFixPrompt(issue types.Issue) string {
	var sb strings.Builder
	if p.config.Language == "zh" {
		sb.WriteString(fmt.Sprintf("问题: %s\n", issue.CNName))
		sb.WriteString(fmt.Sprintf("详情: %s\n", issue.Details))
		sb.WriteString(fmt.Sprintf("位置: %s\n", issue.Location))
		sb.WriteString("请提供具体的修复命令和建议。")
	} else {
		sb.WriteString(fmt.Sprintf("Issue: %s\n", issue.CNName))
		sb.WriteString(fmt.Sprintf("Details: %s\n", issue.Details))
		sb.WriteString(fmt.Sprintf("Location: %s\n", issue.Location))
		sb.WriteString("Please provide specific fix commands and suggestions.")
	}
	return sb.String()
}

func (p *OpenAIProvider) parseAnalysisResponse(content string) (*AnalysisResult, error) {
	// Try to parse as JSON
	var result AnalysisResult
	if err := json.Unmarshal([]byte(content), &result); err == nil {
		return &result, nil
	}

	// If not valid JSON, create a simple result
	return &AnalysisResult{
		Summary:     content,
		RootCause:   "",
		Suggestions: []FixSuggestion{},
		Severity:    types.SeverityWarning,
		Confidence:  0.5,
		Language:    p.config.Language,
	}, nil
}

func (p *OpenAIProvider) parseFixResponse(content string) ([]FixSuggestion, error) {
	// Try to parse as JSON array
	var suggestions []FixSuggestion
	if err := json.Unmarshal([]byte(content), &suggestions); err == nil {
		return suggestions, nil
	}

	// If not valid JSON, return single suggestion
	return []FixSuggestion{{
		Title:       "AI Suggestion",
		Description: content,
		Risk:        "medium",
		AutoFix:     false,
	}}, nil
}

func (p *OpenAIProvider) getLocalizedMessage(key string) string {
	messages := map[string]map[string]string{
		"no_issues_found": {
			"zh": "未发现明显问题",
			"en": "No significant issues found",
		},
		"system_healthy": {
			"zh": "系统健康状况良好",
			"en": "System is healthy",
		},
		"summary_system_prompt": {
			"zh": "生成一句简洁的诊断总结。",
			"en": "Generate a concise diagnostic summary.",
		},
		"fix_system_prompt": {
			"zh": "提供具体的 Kubernetes 问题修复建议。",
			"en": "Provide specific Kubernetes issue fix suggestions.",
		},
	}

	if msg, ok := messages[key][p.config.Language]; ok {
		return msg
	}
	return messages[key]["en"]
}

// Factory creates AI providers
type Factory struct {
	config *Config
}

// NewFactory creates a new AI provider factory
func NewFactory(config *Config) *Factory {
	return &Factory{config: config}
}

// CreateProvider creates an AI provider based on configuration
func (f *Factory) CreateProvider() (Provider, error) {
	if f.config.APIKey == "" {
		return nil, fmt.Errorf("AI API key not configured, set KUDIG_AI_API_KEY")
	}

	switch f.config.Provider {
	case "openai", "qwen", "ollama":
		return NewOpenAIProvider(f.config)
	default:
		return NewOpenAIProvider(f.config)
	}
}
