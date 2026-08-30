package events

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// EventType 事件类型
type EventType string

const (
	EventTypeNormal   EventType = "Normal"
	EventTypeWarning  EventType = "Warning"
	EventTypeError    EventType = "Error"
	EventTypeCreate   EventType = "Created"
	EventTypeUpdate   EventType = "Updated"
	EventTypeDelete   EventType = "Deleted"
)

// ResourceType 资源类型
type ResourceType string

const (
	ResourcePod         ResourceType = "Pod"
	ResourceDeployment  ResourceType = "Deployment"
	ResourceService     ResourceType = "Service"
	ResourceNode        ResourceType = "Node"
	ResourceConfigMap   ResourceType = "ConfigMap"
	ResourceSecret      ResourceType = "Secret"
	ResourceIngress     ResourceType = "Ingress"
	ResourcePersistentVolume ResourceType = "PersistentVolume"
	ResourcePersistentVolumeClaim ResourceType = "PersistentVolumeClaim"
)

// Event 统一事件结构
type Event struct {
	ID            string
	Type          EventType
	ResourceType  ResourceType
	ResourceName  string
	Namespace     string
	Cluster       string
	Reason        string
	Message       string
	Timestamp     time.Time
	Count         int32
	InvolvedObject InvolvedObject
	Labels        map[string]string
	Annotations   map[string]string
}

// InvolvedObject 涉及的对象
type InvolvedObject struct {
	Kind      string
	Name      string
	Namespace string
	UID       string
	APIVersion string
}

// Severity 事件严重级别
type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityCritical Severity = "critical"
)

// GetSeverity 根据事件类型获取严重级别
func (e *Event) GetSeverity() Severity {
	switch e.Type {
	case EventTypeError:
		return SeverityCritical
	case EventTypeWarning:
		return SeverityWarning
	default:
		return SeverityInfo
	}
}

// ToMarkdown 转换为 Markdown 格式
func (e *Event) ToMarkdown() string {
	icon := "ℹ️"
	
	switch e.GetSeverity() {
	case SeverityCritical:
		icon = "🔴"
	case SeverityWarning:
		icon = "🟡"
	default:
		icon = "🟢"
	}
	
	return fmt.Sprintf(`%s **%s** - %s

**资源：** %s/%s
**命名空间：** %s
**集群：** %s
**原因：** %s
**时间：** %s
**消息：**
> %s`,
		icon,
		e.Type,
		e.ResourceType,
		e.InvolvedObject.Kind,
		e.ResourceName,
		e.Namespace,
		e.Cluster,
		e.Reason,
		e.Timestamp.Format("2006-01-02 15:04:05"),
		e.Message,
	)
}

// ToSummary 返回事件摘要
func (e *Event) ToSummary() string {
	icon := "ℹ️"
	switch e.GetSeverity() {
	case SeverityCritical:
		icon = "🔴"
	case SeverityWarning:
		icon = "🟡"
	}
	
	return fmt.Sprintf("%s [%s] %s/%s: %s", 
		icon, 
		e.Type, 
		e.InvolvedObject.Kind, 
		e.ResourceName, 
		e.Reason,
	)
}

// FilterConfig 事件过滤配置
type FilterConfig struct {
	Namespaces     []string       // 关注的命名空间，为空表示所有
	ResourceTypes  []ResourceType // 关注的资源类型
	EventTypes     []EventType    // 关注的事件类型
	Reasons        []string       // 关注的原因（如 BackOff、Unhealthy）
	ExcludeReasons []string       // 排除的原因
	MinSeverity    Severity       // 最小严重级别
}

// ShouldFilter 检查事件是否应该被过滤
func (f *FilterConfig) ShouldFilter(event *Event) bool {
	// 检查命名空间
	if len(f.Namespaces) > 0 {
		found := false
		for _, ns := range f.Namespaces {
			if ns == event.Namespace || ns == "*" {
				found = true
				break
			}
		}
		if !found {
			return true
		}
	}
	
	// 检查资源类型
	if len(f.ResourceTypes) > 0 {
		found := false
		for _, rt := range f.ResourceTypes {
			if rt == event.ResourceType {
				found = true
				break
			}
		}
		if !found {
			return true
		}
	}
	
	// 检查事件类型
	// 兼容历史配置：K8s Event 只有 Normal/Warning 两类，
	// 老版本文档中的 "Error" 类型实际不会出现，归一化映射到 Warning
	if len(f.EventTypes) > 0 {
		found := false
		for _, et := range f.EventTypes {
			if et == event.Type || (et == EventTypeError && event.Type == EventTypeWarning) {
				found = true
				break
			}
		}
		if !found {
			return true
		}
	}
	
	// 检查原因
	if len(f.Reasons) > 0 {
		found := false
		for _, reason := range f.Reasons {
			if reason == event.Reason {
				found = true
				break
			}
		}
		if !found {
			return true
		}
	}
	
	// 检查排除的原因
	for _, reason := range f.ExcludeReasons {
		if reason == event.Reason {
			return true
		}
	}
	
	// 检查严重级别
	severity := event.GetSeverity()
	if f.MinSeverity != "" {
		severityOrder := map[Severity]int{
			SeverityInfo:     0,
			SeverityWarning:  1,
			SeverityCritical: 2,
		}
		if severityOrder[severity] < severityOrder[f.MinSeverity] {
			return true
		}
	}
	
	return false
}

// Source 事件源接口
type Source interface {
	// Name 返回事件源名称
	Name() string
	
	// Start 启动事件监听
	Start(ctx context.Context) error
	
	// Stop 停止事件监听
	Stop() error
	
	// Subscribe 订阅事件
	Subscribe(handler EventHandler)
	
	// Unsubscribe 取消订阅
	Unsubscribe(handler EventHandler)
	
	// SetFilter 设置事件过滤器
	SetFilter(filter *FilterConfig)
	
	// IsHealthy 检查健康状态
	IsHealthy() bool
}

// EventHandler 事件处理函数类型
type EventHandler func(event *Event)

// BaseSource 基础事件源实现
type BaseSource struct {
	name        string
	handlers    []EventHandler
	filter      *FilterConfig
	mu          sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
	running     bool
}

// NewBaseSource 创建基础事件源
func NewBaseSource(name string) *BaseSource {
	return &BaseSource{
		name:     name,
		handlers: make([]EventHandler, 0),
		filter:   &FilterConfig{},
	}
}

// Name 返回名称
func (s *BaseSource) Name() string {
	return s.name
}

// Subscribe 订阅事件
func (s *BaseSource) Subscribe(handler EventHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers = append(s.handlers, handler)
}

// Unsubscribe 取消订阅
func (s *BaseSource) Unsubscribe(handler EventHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// 简单的线性查找移除
	for i, h := range s.handlers {
		// 比较函数指针（简化处理）
		if fmt.Sprintf("%p", h) == fmt.Sprintf("%p", handler) {
			s.handlers = append(s.handlers[:i], s.handlers[i+1:]...)
			break
		}
	}
}

// SetFilter 设置过滤器
func (s *BaseSource) SetFilter(filter *FilterConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.filter = filter
}

// IsHealthy 检查健康状态
func (s *BaseSource) IsHealthy() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// emit 发送事件到所有订阅者
func (s *BaseSource) emit(event *Event) {
	s.mu.RLock()
	filter := s.filter
	handlers := make([]EventHandler, len(s.handlers))
	copy(handlers, s.handlers)
	s.mu.RUnlock()
	
	// 应用过滤器
	if filter != nil && filter.ShouldFilter(event) {
		return
	}
	
	// 发送到所有订阅者（异步；单个 handler 的 panic 不能拖垮进程）
	for _, handler := range handlers {
		go func(h EventHandler) {
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("Event handler panic (recovered): %v\n", r)
				}
			}()
			h(event)
		}(handler)
	}
}

// Manager 事件源管理器
type Manager struct {
	sources map[string]Source
	mu      sync.RWMutex
}

// NewManager 创建事件源管理器
func NewManager() *Manager {
	return &Manager{
		sources: make(map[string]Source),
	}
}

// Register 注册事件源
func (m *Manager) Register(source Source) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sources[source.Name()] = source
}

// Unregister 注销事件源
func (m *Manager) Unregister(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if source, ok := m.sources[name]; ok {
		source.Stop()
		delete(m.sources, name)
	}
}

// Get 获取事件源
func (m *Manager) Get(name string) (Source, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	source, ok := m.sources[name]
	return source, ok
}

// StartAll 启动所有事件源
func (m *Manager) StartAll(ctx context.Context) error {
	m.mu.RLock()
	sources := make([]Source, 0, len(m.sources))
	for _, s := range m.sources {
		sources = append(sources, s)
	}
	m.mu.RUnlock()
	
	for _, source := range sources {
		if err := source.Start(ctx); err != nil {
			return fmt.Errorf("failed to start source %s: %v", source.Name(), err)
		}
	}
	
	return nil
}

// StopAll 停止所有事件源
func (m *Manager) StopAll() {
	m.mu.RLock()
	sources := make([]Source, 0, len(m.sources))
	for _, s := range m.sources {
		sources = append(sources, s)
	}
	m.mu.RUnlock()
	
	for _, source := range sources {
		source.Stop()
	}
}
