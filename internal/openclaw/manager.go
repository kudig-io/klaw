package openclaw

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kudig-io/klaw/internal/config"
	"github.com/kudig-io/klaw/internal/kubernetes"
)

// Manager OpenClaw管理器
type Manager struct {
	config     config.OpenClawConfig
	k8sManager *kubernetes.Manager
	skillsPath string
}

// NewManager 创建OpenClaw管理器
func NewManager(cfg config.OpenClawConfig, k8sManager *kubernetes.Manager) (*Manager, error) {
	m := &Manager{
		config:     cfg,
		k8sManager: k8sManager,
		skillsPath: cfg.Skills,
	}

	// 检查技能路径是否存在
	if _, err := os.Stat(m.skillsPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("skills path not found: %s", m.skillsPath)
	}

	return m, nil
}

// Start 启动OpenClaw管理器
func (m *Manager) Start() {
	fmt.Println("OpenClaw manager started")
	// 这里可以实现技能加载和管理逻辑
}

// GetSkills 获取所有技能
func (m *Manager) GetSkills() ([]string, error) {
	var skills []string

	// 遍历技能目录
	err := filepath.Walk(m.skillsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 检查是否是目录
		if info.IsDir() {
			// 检查是否包含SKILL.md文件
			skillFile := filepath.Join(path, "SKILL.md")
			if _, err := os.Stat(skillFile); err == nil {
				// 获取技能名称（目录名）
				skillName := filepath.Base(path)
				skills = append(skills, skillName)
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk skills path: %v", err)
	}

	return skills, nil
}

// ExecuteSkill 执行技能
func (m *Manager) ExecuteSkill(skillName, command string) (string, error) {
	// 这里可以实现技能执行逻辑
	return fmt.Sprintf("Executed skill %s with command: %s", skillName, command), nil
}
