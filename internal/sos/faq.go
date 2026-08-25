// Package sos 提供 SOS 语音应急快速对话：语料注入、集群工具与 DashScope Realtime 会话桥接。
package sos

import (
	_ "embed"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:embed faq/seed.yaml
var seedFAQ []byte

// FAQEntry 预置问答语料条目
type FAQEntry struct {
	ID       string   `yaml:"id"`
	Question string   `yaml:"question"`
	Answer   string   `yaml:"answer"`
	Tags     []string `yaml:"tags"`
}

type faqFile struct {
	FAQs []FAQEntry `yaml:"faqs"`
}

// LoadFAQs 加载语料；path 为空时使用内嵌默认语料
func LoadFAQs(path string) ([]FAQEntry, error) {
	data := seedFAQ
	if path != "" {
		b, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read faq file: %w", err)
		}
		data = b
	}
	var f faqFile
	if err := yaml.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("parse faq file: %w", err)
	}
	if len(f.FAQs) == 0 {
		return nil, fmt.Errorf("no faq entries loaded")
	}
	return f.FAQs, nil
}

// defaultInstructionsPrefix 三层兜底策略的系统提示
const defaultInstructionsPrefix = `你是 Klaw——Kubernetes 智能运维平台的 SOS 应急语音助手。回答遵守三层优先级：
1) 命中下方标准问答语料中的主题时，严格按语料口径回答；
2) 涉及集群实时状态、异常排查、应急诊断时，必须先调用工具查询真实数据再回答；
3) 其他问题用通用知识回答，并明确说明"这是通用建议，非当前集群实测数据"。
回答用简体中文，口语化、简洁，适合语音播报，一次不超过 5 句。`

// BuildInstructions 组装注入 Realtime 会话的 system instructions（三层兜底第 1 层）
func BuildInstructions(prefix string, faqs []FAQEntry) string {
	var b strings.Builder
	b.WriteString(defaultInstructionsPrefix)
	if prefix != "" {
		b.WriteString("\n")
		b.WriteString(prefix)
	}
	b.WriteString("\n\n## 标准问答语料（命中主题时严格按语料口径回答）")
	for i, f := range faqs {
		fmt.Fprintf(&b, "\n\n### %d. Q: %s\nA: %s", i+1, f.Question, strings.TrimSpace(f.Answer))
	}
	return b.String()
}
