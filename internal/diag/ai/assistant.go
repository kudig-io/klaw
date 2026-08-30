package ai

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/kudig-io/klaw/internal/diag/types"
)

// Assistant provides AI-assisted diagnostic capabilities
type Assistant struct {
	provider Provider
	config   *Config
}

// NewAssistant creates a new AI assistant
func NewAssistant(config *Config) (*Assistant, error) {
	if config == nil {
		config = LoadConfig()
	}

	factory := NewFactory(config)
	provider, err := factory.CreateProvider()
	if err != nil {
		return nil, err
	}

	return &Assistant{
		provider: provider,
		config:   config,
	}, nil
}

// AnalyzeWithAI performs AI analysis on diagnostic results
func (a *Assistant) AnalyzeWithAI(ctx context.Context, issues []types.Issue, hostname string) (*AnalysisResult, error) {
	if len(issues) == 0 {
		return &AnalysisResult{
			Summary:     a.getLocalizedMessage("no_issues"),
			RootCause:   "",
			Suggestions: []FixSuggestion{},
			Severity:    types.SeverityInfo,
			Confidence:  1.0,
			Language:    a.config.Language,
		}, nil
	}

	ctx, cancel := context.WithTimeout(ctx, time.Duration(a.config.Timeout)*time.Second)
	defer cancel()

	return a.provider.Analyze(ctx, issues, hostname)
}

// ExplainIssue provides detailed explanation for a single issue
func (a *Assistant) ExplainIssue(ctx context.Context, issue types.Issue) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(a.config.Timeout)*time.Second)
	defer cancel()

	// Create a single issue analysis
	issues := []types.Issue{issue}
	result, err := a.provider.Analyze(ctx, issues, "")
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "**%s**", result.Summary)
	if result.RootCause != "" {
		fmt.Fprintf(&sb, "\n\n**根因分析:**\n%s", result.RootCause)
	}
	// 之前此处只输出"修复建议:"标题后即截断，现在把建议真正拼接出来
	if len(result.Suggestions) > 0 {
		sb.WriteString("\n\n**修复建议:**\n")
		for i, s := range result.Suggestions {
			fmt.Fprintf(&sb, "%d. %s", i+1, s.Title)
			if s.Description != "" {
				fmt.Fprintf(&sb, " — %s", s.Description)
			}
			if s.Command != "" {
				fmt.Fprintf(&sb, "\n   命令: `%s`", s.Command)
			}
			sb.WriteString("\n")
		}
	}

	return sb.String(), nil
}

// GetFixCommands gets fix commands for an issue
func (a *Assistant) GetFixCommands(ctx context.Context, issue types.Issue) ([]FixSuggestion, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(a.config.Timeout)*time.Second)
	defer cancel()

	return a.provider.SuggestFixes(ctx, issue)
}

// IsAvailable checks if AI assistant is available
func (a *Assistant) IsAvailable() bool {
	return a.provider != nil && a.config.APIKey != ""
}

func (a *Assistant) getLocalizedMessage(key string) string {
	messages := map[string]map[string]string{
		"no_issues": {
			"zh": "🎉 恭喜！系统未检测到明显问题。",
			"en": "🎉 Congratulations! No significant issues detected.",
		},
	}

	if msg, ok := messages[key][a.config.Language]; ok {
		return msg
	}
	return messages[key]["en"]
}
