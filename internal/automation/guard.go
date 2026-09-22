package automation

import (
	"fmt"
	"regexp"
	"sync"
)

// Guard 自定义脚本的危险命令防护：
// 在 bash -c 执行前对命令文本做模式匹配，命中即拦截并记审计日志。
// 默认开启；可通过 automation.guard.enabled: false 关闭（不推荐）。
type Guard struct {
	mu       sync.RWMutex
	enabled  bool
	patterns []dangerousPattern
}

type dangerousPattern struct {
	re   *regexp.Regexp
	desc string
}

// defaultDangerousPatterns 覆盖不可逆的破坏性操作（删根、格盘、写设备、
// 远程脚本直接落地执行、关机、抹除历史/定时任务等），避免误伤常规运维命令。
func defaultDangerousPatterns() []dangerousPattern {
	rules := []struct {
		pattern string
		desc    string
	}{
		{`:\(\)\s*\{.*\|.*&.*\}\s*;\s*:`, "fork 炸弹"},
		// rm -rf 作用于根/家目录本身（/、/*、~、~/、~/* 等，首参数或任意位置）
		{`rm\s+(-[a-zA-Z]+\s+)+(/|~/|~)\*?(\s.*|$)`, "递归删除根目录"},
		{`rm\s+(-[a-zA-Z]+\s+)+(.*\s)?(/|~/|~)\*?\s*$`, "递归删除根目录"},
		// rm -rf 作用于系统一级目录（/usr、/etc、/var 及其子路径等，任意参数位置）
		{`rm\s+(-[a-zA-Z]+\s+)+(.*\s)?/(usr|etc|var|home|boot|bin|sbin|lib|lib64|opt|root|dev|proc|sys|run|srv|mnt|media)(/.*)?\s*$`, "递归删除系统目录"},
		{`\bmkfs(\.\w+)?\b`, "格式化文件系统"},
		{`\b(wipefs|fdisk|sfdisk|cfdisk|parted)\b`, "磁盘分区/擦除"},
		{`\bdd\b[^|]*\bof=/dev/`, "dd 覆写块设备"},
		{`>\s*/dev/(sd|nvme|hd|vd)`, "重定向覆写块设备"},
		{`\b(curl|wget)\b[^|]*\|\s*(ba|z|da|fi)?sh\b`, "远程脚本直接执行"},
		{`\b(shutdown|reboot|halt|poweroff)\b`, "关机/重启"},
		{`\binit\s+[06]\b`, "切换运行级（关机/重启）"},
		{`chmod\s+-R\s+777\s+/(?:\s|$)`, "根目录递归开放写权限"},
		{`\bcrontab\s+-r\b`, "抹除定时任务"},
		{`\bhistory\s+-c\b`, "抹除命令历史"},
	}
	patterns := make([]dangerousPattern, 0, len(rules))
	for _, r := range rules {
		re, err := regexp.Compile(r.pattern)
		if err != nil {
			// 内置规则必须可编译；失败直接跳过而不是 panic
			continue
		}
		patterns = append(patterns, dangerousPattern{re: re, desc: r.desc})
	}
	return patterns
}

func NewGuard(enabled bool) *Guard {
	return &Guard{enabled: enabled, patterns: defaultDangerousPatterns()}
}

func (g *Guard) Enabled() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.enabled
}

func (g *Guard) SetEnabled(enabled bool) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.enabled = enabled
}

// Check 返回 nil 表示放行；命中时返回说明具体规则的错误
func (g *Guard) Check(command string) error {
	if g == nil || !g.Enabled() {
		return nil
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	for _, p := range g.patterns {
		if p.re.MatchString(command) {
			return fmt.Errorf("命令被安全防护拦截：命中危险模式「%s」，如确认需要执行请在配置中关闭 automation.guard.enabled", p.desc)
		}
	}
	return nil
}
