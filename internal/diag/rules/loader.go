package rules

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Loader loads rules from YAML files
type Loader struct {
	ruleSets []*RuleSet
}

// NewLoader creates a new rule loader
func NewLoader() *Loader {
	return &Loader{
		ruleSets: make([]*RuleSet, 0),
	}
}

// LoadFile loads rules from a single YAML file
func (l *Loader) LoadFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read rule file %s: %w", path, err)
	}

	var ruleSet RuleSet
	if err := yaml.Unmarshal(data, &ruleSet); err != nil {
		return fmt.Errorf("failed to parse rule file %s: %w", path, err)
	}

	// Validate rules
	for i, rule := range ruleSet.Rules {
		if rule.ID == "" {
			return fmt.Errorf("rule %d in %s: missing id", i, path)
		}
		if rule.Name == "" {
			ruleSet.Rules[i].Name = rule.ID
		}
		if rule.Severity == "" {
			ruleSet.Rules[i].Severity = "info"
		}
		// Enable by default if not specified
		if !rule.Enabled {
			ruleSet.Rules[i].Enabled = true
		}
	}

	l.ruleSets = append(l.ruleSets, &ruleSet)
	return nil
}

// LoadDir loads all YAML files from a directory
func (l *Loader) LoadDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read rules directory %s: %w", dir, err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		ext := filepath.Ext(entry.Name())
		if ext != ".yaml" && ext != ".yml" {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		if err := l.LoadFile(path); err != nil {
			return err
		}
	}

	return nil
}

// LoadBuiltin loads built-in default rules
func (l *Loader) LoadBuiltin() error {
	builtinRules := &RuleSet{
		Version:     "1.0",
		Name:        "builtin",
		Description: "Built-in diagnostic rules",
		Rules: []Rule{
			{
				ID:          "HIGH_CPU_LOAD",
				Name:        "CPU负载过高",
				Description: "检查CPU负载是否超过核心数的80%",
				Category:    "system",
				Severity:    "warning",
				Enabled:     true,
				Condition: Condition{
					Type:      "metric_threshold",
					Metric:    "load_avg_1min",
					Operator:  "gt",
					Threshold: 0.8, // Will be multiplied by CPU cores
				},
				Remediation: "检查高CPU使用进程: top -c; 分析是否有异常进程",
			},
			{
				ID:          "HIGH_MEMORY_USAGE",
				Name:        "内存使用率过高",
				Description: "检查内存使用率是否超过90%",
				Category:    "system",
				Severity:    "warning",
				Enabled:     true,
				Condition: Condition{
					Type:      "metric_threshold",
					Metric:    "mem_used_percent",
					Operator:  "gt",
					Threshold: 90,
				},
				Remediation: "检查内存使用: free -h; 识别内存占用高的进程",
			},
			{
				ID:          "DISK_SPACE_LOW",
				Name:        "磁盘空间不足",
				Description: "检查磁盘使用率是否超过85%",
				Category:    "system",
				Severity:    "warning",
				Enabled:     true,
				Condition: Condition{
					Type:      "metric_threshold",
					Metric:    "disk_used_percent",
					Operator:  "gt",
					Threshold: 85,
				},
				Remediation: "清理磁盘空间: df -h; 查找大文件: du -sh /*",
			},
			{
				ID:          "OOM_KILLER",
				Name:        "OOM Killer事件",
				Description: "检查是否有OOM Killer杀死进程",
				Category:    "kernel",
				Severity:    "critical",
				Enabled:     true,
				Condition: Condition{
					Type:    "file_contains",
					File:    "logs/dmesg.log",
					Pattern: "Out of memory|oom-kill|invoked oom-killer",
				},
				Remediation: "增加内存或优化Pod资源限制; 检查: dmesg | grep -i oom",
			},
			{
				ID:          "KERNEL_PANIC",
				Name:        "内核Panic",
				Description: "检查是否有内核Panic事件",
				Category:    "kernel",
				Severity:    "critical",
				Enabled:     true,
				Condition: Condition{
					Type:    "file_contains",
					File:    "logs/dmesg.log",
					Pattern: "Kernel panic|kernel BUG|BUG:",
				},
				Remediation: "检查硬件和驱动; 更新内核版本",
			},
			{
				ID:          "KUBELET_NOT_RUNNING",
				Name:        "Kubelet未运行",
				Description: "检查Kubelet服务是否运行",
				Category:    "kubernetes",
				Severity:    "critical",
				Enabled:     true,
				Condition: Condition{
					Type:    "file_contains",
					File:    "service_status",
					Pattern: "kubelet.*inactive|kubelet.*dead",
				},
				Remediation: "启动Kubelet: systemctl start kubelet",
			},
			{
				ID:          "CONTAINER_RUNTIME_DOWN",
				Name:        "容器运行时未运行",
				Description: "检查容器运行时服务是否运行",
				Category:    "kubernetes",
				Severity:    "critical",
				Enabled:     true,
				Condition: Condition{
					Type: "or",
					Or: []Condition{
						{
							Type:    "file_contains",
							File:    "service_status",
							Pattern: "containerd.*inactive|containerd.*dead",
						},
						{
							Type:    "file_contains",
							File:    "service_status",
							Pattern: "docker.*inactive|docker.*dead",
						},
					},
				},
				Remediation: "启动容器运行时: systemctl start containerd",
			},
			{
				ID:          "PLEG_UNHEALTHY",
				Name:        "PLEG不健康",
				Description: "检查Kubelet PLEG状态",
				Category:    "kubernetes",
				Severity:    "critical",
				Enabled:     true,
				Condition: Condition{
					Type:    "file_contains",
					File:    "logs/kubelet.log",
					Pattern: "PLEG is not healthy",
				},
				Remediation: "重启容器运行时: systemctl restart containerd",
			},
			{
				ID:          "CNI_ERROR",
				Name:        "CNI网络错误",
				Description: "检查CNI网络插件错误",
				Category:    "network",
				Severity:    "critical",
				Enabled:     true,
				Condition: Condition{
					Type:    "file_contains",
					File:    "logs/kubelet.log",
					Pattern: "CNI.*failed|cni.*error|NetworkPlugin.*error",
				},
				Remediation: "检查CNI配置: ls /etc/cni/net.d/; 重启网络插件",
			},
			{
				ID:          "CERTIFICATE_ERROR",
				Name:        "证书错误",
				Description: "检查证书相关错误",
				Category:    "kubernetes",
				Severity:    "critical",
				Enabled:     true,
				Condition: Condition{
					Type:    "file_contains",
					File:    "logs/kubelet.log",
					Pattern: "certificate.*expired|x509.*certificate",
				},
				Remediation: "更新证书: kubeadm certs renew all",
			},
		},
	}

	l.ruleSets = append(l.ruleSets, builtinRules)
	return nil
}

// GetAllRules returns all loaded rules
func (l *Loader) GetAllRules() []Rule {
	var rules []Rule
	for _, rs := range l.ruleSets {
		for _, r := range rs.Rules {
			if r.Enabled {
				rules = append(rules, r)
			}
		}
	}
	return rules
}

// GetRulesByCategory returns rules filtered by category
func (l *Loader) GetRulesByCategory(category string) []Rule {
	var rules []Rule
	for _, rs := range l.ruleSets {
		for _, r := range rs.Rules {
			if r.Enabled && r.Category == category {
				rules = append(rules, r)
			}
		}
	}
	return rules
}

// GetRuleByID returns a rule by its ID
func (l *Loader) GetRuleByID(id string) *Rule {
	for _, rs := range l.ruleSets {
		for _, r := range rs.Rules {
			if r.ID == id {
				return &r
			}
		}
	}
	return nil
}
