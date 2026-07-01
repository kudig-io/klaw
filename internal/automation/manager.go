package automation

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/kudig-io/klaw/internal/storage"
	"k8s.io/client-go/kubernetes"
)

type Manager struct {
	store     *storage.Store
	scripts   []Script
	history   []ScriptExecution
	mu        sync.RWMutex
	clientset kubernetes.Interface
}

func NewManager(store *storage.Store) *Manager {
	m := &Manager{store: store}
	m.load()
	if len(m.scripts) == 0 {
		m.scripts = defaultScripts()
		_ = m.saveLocked()
	}
	return m
}

func (m *Manager) WithClientset(cs kubernetes.Interface) *Manager {
	m.clientset = cs
	return m
}

func (m *Manager) List(filter ScriptFilter) []Script {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []Script
	for _, s := range m.scripts {
		if filter.Type != "" && s.Type != filter.Type {
			continue
		}
		if filter.Enabled != nil && s.Enabled != *filter.Enabled {
			continue
		}
		result = append(result, s)
		if filter.Limit > 0 && len(result) >= filter.Limit {
			break
		}
	}
	return result
}

func (m *Manager) Get(id string) (*Script, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, s := range m.scripts {
		if s.ID == id {
			copy := s
			return &copy, nil
		}
	}
	return nil, fmt.Errorf("script not found: %s", id)
}

func (m *Manager) Add(script Script) (*Script, error) {
	if script.Name == "" {
		return nil, fmt.Errorf("script name is required")
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	if script.ID == "" {
		script.ID = fmt.Sprintf("script-%s", uuid.New().String()[:8])
	}
	if script.Type == "" {
		script.Type = ScriptTypeCustom
	}
	if script.Timeout == 0 {
		script.Timeout = 300
	}
	now := time.Now()
	script.CreatedAt = now
	script.UpdatedAt = now

	m.scripts = append([]Script{script}, m.scripts...)
	_ = m.saveLocked()
	copy := script
	return &copy, nil
}

func (m *Manager) Update(id string, updates Script) (*Script, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i := range m.scripts {
		if m.scripts[i].ID != id {
			continue
		}
		s := &m.scripts[i]
		if updates.Name != "" {
			s.Name = updates.Name
		}
		if updates.Description != "" {
			s.Description = updates.Description
		}
		s.Enabled = updates.Enabled
		if updates.Schedule != "" {
			s.Schedule = updates.Schedule
		}
		if updates.Timeout > 0 {
			s.Timeout = updates.Timeout
		}
		if updates.Parameters != nil {
			s.Parameters = updates.Parameters
		}
		s.UpdatedAt = time.Now()
		_ = m.saveLocked()
		copy := *s
		return &copy, nil
	}
	return nil, fmt.Errorf("script not found: %s", id)
}

func (m *Manager) Delete(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, s := range m.scripts {
		if s.ID == id {
			m.scripts = append(m.scripts[:i], m.scripts[i+1:]...)
			return m.saveLocked()
		}
	}
	return fmt.Errorf("script not found: %s", id)
}

func (m *Manager) Execute(ctx context.Context, scriptID, trigger string, params map[string]interface{}) (*ScriptExecution, error) {
	script, err := m.Get(scriptID)
	if err != nil {
		return nil, err
	}

	merged := mergeParams(script.Parameters, params)
	now := time.Now()
	exec := &ScriptExecution{
		ID:         fmt.Sprintf("exec-%s", uuid.New().String()[:8]),
		ScriptID:   scriptID,
		ScriptName: script.Name,
		Trigger:    trigger,
		Parameters: merged,
		Status:     StatusRunning,
		StartTime:  now,
	}

	output, err := m.dispatch(ctx, script, merged)
	endTime := time.Now()
	exec.EndTime = &endTime
	exec.Duration = int(endTime.Sub(now).Seconds())

	if err != nil {
		exec.Status = StatusFailed
		exec.Error = err.Error()
	} else {
		exec.Status = StatusSuccess
		exec.Output = output
	}

	m.mu.Lock()
	m.history = append([]ScriptExecution{*exec}, m.history...)
	if len(m.history) > 1000 {
		m.history = m.history[:1000]
	}
	_ = m.saveHistoryLocked()
	m.mu.Unlock()

	return exec, nil
}

func (m *Manager) History(limit int) []ScriptExecution {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if limit > 0 && limit < len(m.history) {
		return append([]ScriptExecution(nil), m.history[:limit]...)
	}
	return append([]ScriptExecution(nil), m.history...)
}

func (m *Manager) Statistics() Statistics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := Statistics{
		ByScript:  map[string]int{},
		ByTrigger: map[string]int{},
	}
	var totalDur int
	for _, e := range m.history {
		stats.Total++
		if e.Status == StatusSuccess {
			stats.Successful++
		} else if e.Status == StatusFailed {
			stats.Failed++
		}
		stats.ByScript[e.ScriptName]++
		stats.ByTrigger[e.Trigger]++
		totalDur += e.Duration
	}
	if stats.Total > 0 {
		stats.AvgDuration = float64(totalDur) / float64(stats.Total)
	}
	return stats
}

func (m *Manager) dispatch(ctx context.Context, script *Script, params map[string]interface{}) (string, error) {
	switch script.Type {
	case ScriptTypeBuiltin:
		return m.executeBuiltin(ctx, script.Script, params)
	case ScriptTypeCustom:
		return m.executeCustom(ctx, script.Script, script.Timeout, params)
	default:
		return "", fmt.Errorf("unknown script type: %s", script.Type)
	}
}

func (m *Manager) executeCustom(ctx context.Context, command string, timeoutSec int, params map[string]interface{}) (string, error) {
	if command == "" {
		return "", fmt.Errorf("custom script is empty")
	}
	if timeoutSec <= 0 {
		timeoutSec = 300
	}
	runCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSec)*time.Second)
	defer cancel()

	envs := os.Environ()
	for k, v := range params {
		envs = append(envs, fmt.Sprintf("%s=%v", strings.ToUpper(k), v))
	}

	cmd := exec.CommandContext(runCtx, "bash", "-c", command)
	cmd.Env = envs
	out, err := cmd.CombinedOutput()
	if err != nil {
		if runCtx.Err() == context.DeadlineExceeded {
			return string(out), fmt.Errorf("script timeout after %ds", timeoutSec)
		}
		return string(out), fmt.Errorf("script failed: %w", err)
	}
	return string(out), nil
}

func mergeParams(base, override map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range base {
		result[k] = v
	}
	for k, v := range override {
		result[k] = v
	}
	return result
}

func (m *Manager) load() {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, _ = m.store.GetJSON("automation", "scripts", &m.scripts)
	_, _ = m.store.GetJSON("automation", "history", &m.history)
}

func (m *Manager) saveLocked() error {
	return m.store.PutJSON("automation", "scripts", m.scripts)
}

func (m *Manager) saveHistoryLocked() error {
	return m.store.PutJSON("automation", "history", m.history)
}

func defaultScripts() []Script {
	now := time.Now()
	defaults := []Script{
		{ID: "cleanup-evicted-pods", Name: "清理已驱逐的 Pod", Description: "自动清理所有处于 Evicted 状态的 Pod", Enabled: true, Type: ScriptTypeBuiltin, Script: "cleanup-evicted-pods", Schedule: "0 */6 * * *", Timeout: 300, CreatedAt: now, UpdatedAt: now},
		{ID: "restart-crashing-pods", Name: "重启崩溃的 Pod", Description: "自动重启崩溃循环的 Pod", Enabled: false, Type: ScriptTypeBuiltin, Script: "restart-crashing-pods", Schedule: "*/5 * * * *", Timeout: 300, CreatedAt: now, UpdatedAt: now},
		{ID: "scale-deployment", Name: "扩展 Deployment", Description: "根据资源使用率自动扩展 Deployment", Enabled: false, Type: ScriptTypeBuiltin, Script: "scale-deployment", Schedule: "*/2 * * * *", Timeout: 300, CreatedAt: now, UpdatedAt: now},
		{ID: "check-node-health", Name: "检查节点健康", Description: "定期检查节点健康状态并报告", Enabled: true, Type: ScriptTypeBuiltin, Script: "check-node-health", Schedule: "*/10 * * * *", Timeout: 300, CreatedAt: now, UpdatedAt: now},
		{ID: "backup-configmaps", Name: "备份 ConfigMaps", Description: "定期备份所有 ConfigMaps", Enabled: false, Type: ScriptTypeBuiltin, Script: "backup-configmaps", Schedule: "0 3 * * *", Timeout: 300, CreatedAt: now, UpdatedAt: now},
		{ID: "update-image-tags", Name: "更新镜像标签", Description: "批量更新 Deployment 中的镜像标签", Enabled: false, Type: ScriptTypeBuiltin, Script: "update-image-tags", Schedule: "", Timeout: 300, CreatedAt: now, UpdatedAt: now},
	}
	return defaults
}
