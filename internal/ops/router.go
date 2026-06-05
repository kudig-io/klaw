package ops

import (
	"context"
	"fmt"
	"strings"

	"github.com/kudig-io/klaw/internal/messaging"
)

// CommandRouter 命令路由器
type CommandRouter struct {
	handler     *Handler
	prefix      string // 命令前缀，如 "klaw"
	mentionName string // 被@时的名称
}

// NewCommandRouter 创建命令路由器
func NewCommandRouter(handler *Handler) *CommandRouter {
	return &CommandRouter{
		handler:     handler,
		prefix:      "klaw",
		mentionName: "Klaw",
	}
}

// SetPrefix 设置命令前缀
func (r *CommandRouter) SetPrefix(prefix string) {
	r.prefix = prefix
}

// SetMentionName 设置@名称
func (r *CommandRouter) SetMentionName(name string) {
	r.mentionName = name
}

// HandleMessage 实现 messaging.MessageHandler 接口
func (r *CommandRouter) HandleMessage(ctx context.Context, msg *messaging.Message) (*messaging.Response, error) {
	// 检查是否是命令
	command, ok := r.extractCommand(msg.Content, msg.Mentioned)
	if !ok {
		// 不是命令，忽略
		return nil, nil
	}

	command = r.ExpandCommand(command)

	// 执行命令
	result, err := r.handler.HandleCommand(command)
	if err != nil {
		return &messaging.Response{
			Content: fmt.Sprintf("**Error:** %s", err.Error()),
			Format:  messaging.FormatMarkdown,
		}, nil
	}

	// 返回结果
	return &messaging.Response{
		Content: result,
		Format:  r.detectFormat(result),
	}, nil
}

// extractCommand 从消息内容中提取命令
func (r *CommandRouter) extractCommand(content string, mentioned bool) (string, bool) {
	content = strings.TrimSpace(content)

	// 如果被@提及，移除@部分
	if mentioned {
		// 移除 "@Klaw " 或 "@Klaw" 前缀
		mentionPattern := "@" + r.mentionName
		if idx := strings.Index(content, mentionPattern); idx != -1 {
			content = content[:idx] + content[idx+len(mentionPattern):]
			content = strings.TrimSpace(content)
		}
	}

	// 检查前缀
	prefixPattern := r.prefix + " "
	if !strings.HasPrefix(content, prefixPattern) && content != r.prefix {
		// 不匹配前缀
		return "", false
	}

	// 提取命令部分
	command := strings.TrimPrefix(content, prefixPattern)
	command = strings.TrimSpace(command)

	if command == "" {
		// 只有前缀，返回帮助
		return "help", true
	}

	return command, true
}

// detectFormat 检测响应格式
func (r *CommandRouter) detectFormat(content string) messaging.FormatType {
	// 如果包含 Markdown 语法，使用 Markdown
	if strings.Contains(content, "**") ||
		strings.Contains(content, "```") ||
		strings.Contains(content, "|") ||
		strings.Contains(content, "#") {
		return messaging.FormatMarkdown
	}

	// 如果包含 JSON，使用 JSON 格式
	if strings.HasPrefix(strings.TrimSpace(content), "{") ||
		strings.HasPrefix(strings.TrimSpace(content), "[") {
		return messaging.FormatJSON
	}

	return messaging.FormatPlain
}

// HandleFunc 返回可直接使用的 HandleFunc
func (r *CommandRouter) HandleFunc() messaging.MessageHandler {
	return r.HandleMessage
}

// ShowHelp 显示帮助信息（统一 Markdown 表格格式，与 kudig 对齐）
func (r *CommandRouter) ShowHelp() string {
	return renderHelpMarkdown()
}

// RegisterShortcuts 注册命令缩写
func (r *CommandRouter) RegisterShortcuts() map[string]string {
	return map[string]string{
		"c":    "cluster",
		"p":    "pod",
		"n":    "node",
		"d":    "deployment",
		"s":    "service",
		"svc":  "service",
		"r":    "rbac",
		"m":    "monitor",
		"h":    "help",
		"ls":   "list",
		"desc": "describe",
		"log":  "logs",
		"del":  "delete",
		"rm":   "delete",
	}
}

// ExpandCommand 展开缩写命令
func (r *CommandRouter) ExpandCommand(command string) string {
	shortcuts := r.RegisterShortcuts()
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return command
	}

	// 展开第一部分
	if expanded, ok := shortcuts[parts[0]]; ok {
		parts[0] = expanded
	}

	// 展开第二部分（如果存在）
	if len(parts) > 1 {
		if expanded, ok := shortcuts[parts[1]]; ok {
			parts[1] = expanded
		}
	}

	return strings.Join(parts, " ")
}
