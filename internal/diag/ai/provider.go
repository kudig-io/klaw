// Package ai provides AI/LLM integration for diagnostic assistance
package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
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
	// Provider: openai, qwen, ollama, mimo（均为 OpenAI 兼容协议；mimo 为小米 MiMo 开放平台）
	// mimo 需区分计费套餐：sk- 按量 key 用 api.xiaomimimo.com；tp- Token Plan key 用 token-plan-cn.xiaomimimo.com
	Provider    string `env:"KUDIG_AI_PROVIDER" default:"openai"`
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
		if result, err := strconv.Atoi(value); err == nil {
			return result
		}
		// 非法值不能静默返回 0（会导致 Timeout=0 立即超时），回退默认值
		fmt.Printf("⚠ Invalid %s=%q, using default %d\n", key, value, defaultValue)
	}
	return defaultValue
}

func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if result, err := strconv.ParseFloat(value, 64); err == nil {
			return result
		}
		fmt.Printf("⚠ Invalid %s=%q, using default %v\n", key, value, defaultValue)
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
	// ollama 为本地服务，通常无鉴权，允许空 APIKey
	if config.APIKey == "" && config.Provider != "ollama" {
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

	// MiMo 默认开启深度思考（reasoning_content 会挤占生成预算），
	// go-openai 无法注入 thinking 字段，故 mimo 走专用 HTTP 路径显式关闭
	if p.config.Provider == "mimo" {
		return p.mimoAnalyze(timeoutCtx, issues, hostname)
	}

	// Prepare prompt
	prompt := p.buildAnalysisPrompt(issues, hostname)

	// Create chat completion（注意：JSON 输出请勿使用 Markdown 围栏）
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

// mimoAnalyze MiMo 专用调用路径：请求体额外注入 thinking.disabled（关闭深度思考，
// 避免 reasoning_content 挤占生成预算导致正文为空），响应仍为 OpenAI 兼容格式。
func (p *OpenAIProvider) mimoAnalyze(ctx context.Context, issues []types.Issue, hostname string) (*AnalysisResult, error) {
	body := map[string]any{
		"model": p.config.Model,
		"messages": []map[string]string{
			{"role": "system", "content": p.getSystemPrompt()},
			{"role": "user", "content": p.buildAnalysisPrompt(issues, hostname)},
		},
		"max_tokens":  p.config.MaxTokens,
		"temperature": p.config.Temperature,
		"stream":      false,
		"thinking":    map[string]string{"type": "disabled"},
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to build mimo request: %w", err)
	}

	endpoint := strings.TrimRight(p.config.BaseURL, "/") + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(string(payload)))
	if err != nil {
		return nil, fmt.Errorf("failed to create mimo request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.config.APIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get AI analysis: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20))
	if err != nil {
		return nil, fmt.Errorf("failed to read mimo response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get AI analysis: status %d, body: %.200s", resp.StatusCode, string(data))
	}

	var chatResp openai.ChatCompletionResponse
	if err := json.Unmarshal(data, &chatResp); err != nil {
		return nil, fmt.Errorf("failed to parse mimo response: %w", err)
	}
	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("no response from AI")
	}

	return p.parseAnalysisResponse(chatResp.Choices[0].Message.Content)
}

func (p *OpenAIProvider) getSystemPrompt() string {
	if p.config.Language == "zh" {
		return `你是一位 Kubernetes 诊断专家。分析诊断结果并提供：
1. 问题摘要（中文）
2. 根因分析
3. 修复建议（包含风险等级）
4. 置信度（0-1）

请直接输出 JSON（不要包裹在 Markdown 代码块中），字段：摘要、根因分析、修复建议（数组，元素含标题/描述/命令/风险等级）、置信度。`
	}
	return `You are a Kubernetes diagnostic expert. Analyze the diagnostic results and provide:
1. Issue summary (English)
2. Root cause analysis
3. Fix suggestions with risk levels
4. Confidence score (0-1)

Return the result directly as JSON (no Markdown code fences), with fields: summary, root_cause, suggestions (array of {title, description, command, risk}), confidence.`
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

// parseAnalysisResponse 解析 LLM 返回的分析结果。
// 兼容两种常见输出：标准 JSON（英文字段名）与包裹在 Markdown 围栏内、
// 使用中文键名的 JSON；两者都失败时降级为纯文本摘要。
func (p *OpenAIProvider) parseAnalysisResponse(content string) (*AnalysisResult, error) {
	body := stripCodeFence(content)

	var result AnalysisResult
	// 注意：中文键名对英文 tag 全是未知字段，Unmarshal 会"成功"但得到零值，
	// 因此必须同时校验 Summary 非空，否则会短路掉下方的中文键映射分支
	if err := json.Unmarshal([]byte(body), &result); err == nil && result.Summary != "" {
		return &result, nil
	}

	// LLM 常用中文键名（如“摘要”“根因分析”）返回，做一次映射兼容
	var raw map[string]any
	if err := json.Unmarshal([]byte(body), &raw); err == nil {
		if mapped := mapChineseKeys(raw); mapped != nil {
			return mapped, nil
		}
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

// stripCodeFence 去除包裹 JSON 的 Markdown 代码围栏
func stripCodeFence(content string) string {
	s := strings.TrimSpace(content)
	if strings.HasPrefix(s, "```") {
		if idx := strings.Index(s, "\n"); idx > 0 {
			s = s[idx+1:]
		}
		s = strings.TrimSuffix(s, "```")
	}
	return strings.TrimSpace(s)
}

// mapChineseKeys 将中文键名的分析结果映射为 AnalysisResult；无法识别时返回 nil
func mapChineseKeys(raw map[string]any) *AnalysisResult {
	pick := func(keys ...string) string {
		for _, k := range keys {
			if v, ok := raw[k].(string); ok && v != "" {
				return v
			}
		}
		return ""
	}
	summary := pick("摘要", "问题摘要", "summary")
	if summary == "" {
		return nil
	}
	res := &AnalysisResult{
		Summary:     summary,
		RootCause:   pick("根因分析", "根因", "root_cause"),
		Suggestions: []FixSuggestion{},
		Severity:    types.SeverityWarning,
		Confidence:  0.7,
	}
	if conf, ok := raw["置信度"].(float64); ok {
		if conf <= 1 {
			res.Confidence = conf
		} else { // 百分制
			res.Confidence = conf / 100
		}
	}
	if arr, ok := raw["修复建议"].([]any); ok {
		for _, item := range arr {
			m, ok := item.(map[string]any)
			if !ok {
				continue
			}
			sug := FixSuggestion{Title: strVal(m, "标题", "建议")}
			sug.Description = strVal(m, "描述", "说明")
			sug.Command = strVal(m, "命令")
			sug.Risk = strVal(m, "风险等级", "风险")
			res.Suggestions = append(res.Suggestions, sug)
		}
	}
	return res
}

func strVal(m map[string]any, keys ...string) string {
	for _, k := range keys {
		if v, ok := m[k].(string); ok {
			return v
		}
	}
	return ""
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

// CreateProvider creates an AI provider based on configuration.
// 所有 provider 均走 OpenAI 兼容协议，工厂仅负责按 provider 补齐默认 BaseURL/Model 与鉴权要求。
func (f *Factory) CreateProvider() (Provider, error) {
	switch strings.ToLower(f.config.Provider) {
	case "mimo":
		if f.config.APIKey == "" {
			return nil, fmt.Errorf("AI API key not configured, set KUDIG_AI_API_KEY")
		}
		// 小米 MiMo 开放平台（https://mimo.mi.com），Bearer 鉴权；未显式指定 BaseURL 时按 key 前缀推断：
		// tp- 为 Token Plan 订阅套餐（专属端点），其余为按量付费 sk- key
		if f.config.BaseURL == "" {
			if strings.HasPrefix(f.config.APIKey, "tp-") {
				f.config.BaseURL = "https://token-plan-cn.xiaomimimo.com/v1"
			} else {
				f.config.BaseURL = "https://api.xiaomimimo.com/v1"
			}
		}
		if f.config.Model == "" || f.config.Model == "gpt-4" {
			f.config.Model = "mimo-v2.5"
		}
		return NewOpenAIProvider(f.config)
	case "ollama":
		// 本地 Ollama 无需鉴权；显式设置的 BaseURL/Model 优先
		if f.config.BaseURL == "" {
			f.config.BaseURL = "http://localhost:11434/v1"
		}
		if f.config.Model == "" || f.config.Model == "gpt-4" {
			f.config.Model = "llama3"
		}
		return NewOpenAIProvider(f.config)
	default: // openai、qwen 及其他 OpenAI 兼容服务
		if f.config.APIKey == "" {
			return nil, fmt.Errorf("AI API key not configured, set KUDIG_AI_API_KEY")
		}
		return NewOpenAIProvider(f.config)
	}
}
