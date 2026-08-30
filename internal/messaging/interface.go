package messaging

import (
	"context"
	"sync"
)

// Message 表示收到的消息
type Message struct {
	ID        string
	Content   string
	SenderID  string
	SenderName string
	ChannelID string
	ChannelName string
	Timestamp int64
	IsGroup   bool
	Mentioned bool // 是否被@提及
}

// Response 表示发送的响应
type Response struct {
	Content string
	Format  FormatType
	// 交互式组件（按钮、菜单等）
	Interactive *InteractiveComponent
}

// FormatType 消息格式类型
type FormatType string

const (
	FormatPlain     FormatType = "plain"     // 纯文本
	FormatMarkdown  FormatType = "markdown"  // Markdown
	FormatJSON      FormatType = "json"      // JSON 代码块
	FormatTable     FormatType = "table"     // 表格
	FormatImage     FormatType = "image"     // 图片
)

// InteractiveComponent 交互式组件
type InteractiveComponent struct {
	Type       ComponentType
	Buttons    []Button
	Menus      []Menu
}

type ComponentType string

const (
	ComponentButtons ComponentType = "buttons"
	ComponentMenus   ComponentType = "menus"
)

type Button struct {
	ID    string
	Label string
	Style string // primary, danger, default
	Value string
}

type Menu struct {
	ID      string
	Label   string
	Options []MenuOption
}

type MenuOption struct {
	Label string
	Value string
}

// Communicator 通信平台接口
type Communicator interface {
	// Name 返回通信平台名称
	Name() string
	
	// Start 启动通信客户端
	Start() error
	
	// Stop 停止通信客户端
	Stop() error
	
	// SendMessage 发送消息到指定频道/用户
	SendMessage(channelID string, response *Response) error
	
	// RegisterHandler 注册消息处理器
	RegisterHandler(handler MessageHandler)
	
	// IsHealthy 检查连接健康状态
	IsHealthy() bool
}

// MessageHandler 消息处理函数类型
type MessageHandler func(ctx context.Context, msg *Message) (*Response, error)

// CommunicatorFactory 通信平台工厂
type CommunicatorFactory interface {
	// Create 根据配置创建通信平台实例
	Create(config interface{}) (Communicator, error)
	// SupportedTypes 返回支持的通信平台类型
	SupportedTypes() []string
}

// CommunicatorRegistry 通信平台注册中心
type CommunicatorRegistry struct {
	factories map[string]CommunicatorFactory
}

// NewCommunicatorRegistry 创建注册中心
func NewCommunicatorRegistry() *CommunicatorRegistry {
	return &CommunicatorRegistry{
		factories: make(map[string]CommunicatorFactory),
	}
}

// Register 注册工厂
func (r *CommunicatorRegistry) Register(platformType string, factory CommunicatorFactory) {
	r.factories[platformType] = factory
}

// Get 获取工厂
func (r *CommunicatorRegistry) Get(platformType string) (CommunicatorFactory, bool) {
	factory, ok := r.factories[platformType]
	return factory, ok
}

// Manager 通信平台管理器（并发安全：注册与发送可在不同 goroutine 进行）
type Manager struct {
	mu            sync.RWMutex
	registry      *CommunicatorRegistry
	communicators map[string]Communicator
	handlers      []MessageHandler
}

// NewManager 创建管理器
func NewManager(registry *CommunicatorRegistry) *Manager {
	return &Manager{
		registry:      registry,
		communicators: make(map[string]Communicator),
		handlers:      make([]MessageHandler, 0),
	}
}

// RegisterCommunicator 注册通信平台
func (m *Manager) RegisterCommunicator(name string, comm Communicator) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.communicators[name] = comm
}

// GetCommunicator 获取通信平台
func (m *Manager) GetCommunicator(name string) (Communicator, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	comm, ok := m.communicators[name]
	return comm, ok
}

// RegisterGlobalHandler 注册全局消息处理器
func (m *Manager) RegisterGlobalHandler(handler MessageHandler) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.handlers = append(m.handlers, handler)
}

// snapshot 复制当前平台与 handler 列表，避免持锁调用外部代码
func (m *Manager) snapshot() (map[string]Communicator, []MessageHandler) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	comms := make(map[string]Communicator, len(m.communicators))
	for name, comm := range m.communicators {
		comms[name] = comm
	}
	handlers := make([]MessageHandler, len(m.handlers))
	copy(handlers, m.handlers)
	return comms, handlers
}

// StartAll 启动所有通信平台
func (m *Manager) StartAll() error {
	comms, handlers := m.snapshot()
	for name, comm := range comms {
		for _, handler := range handlers {
			comm.RegisterHandler(handler)
		}

		if err := comm.Start(); err != nil {
			return &StartError{Platform: name, Err: err}
		}
	}
	return nil
}

// StopAll 停止所有通信平台
func (m *Manager) StopAll() error {
	comms, _ := m.snapshot()
	var errs []error
	for name, comm := range comms {
		if err := comm.Stop(); err != nil {
			errs = append(errs, &StopError{Platform: name, Err: err})
		}
	}
	if len(errs) > 0 {
		return &MultiError{Errors: errs}
	}
	return nil
}

// SendToAll 发送消息到所有平台
func (m *Manager) SendToAll(channelID string, response *Response) error {
	comms, _ := m.snapshot()
	var errs []error
	for name, comm := range comms {
		if err := comm.SendMessage(channelID, response); err != nil {
			errs = append(errs, &SendError{Platform: name, Err: err})
		}
	}
	if len(errs) > 0 {
		return &MultiError{Errors: errs}
	}
	return nil
}

// 错误类型定义

type StartError struct {
	Platform string
	Err      error
}

func (e *StartError) Error() string {
	return "failed to start " + e.Platform + ": " + e.Err.Error()
}

type StopError struct {
	Platform string
	Err      error
}

func (e *StopError) Error() string {
	return "failed to stop " + e.Platform + ": " + e.Err.Error()
}

type SendError struct {
	Platform string
	Err      error
}

func (e *SendError) Error() string {
	return "failed to send via " + e.Platform + ": " + e.Err.Error()
}

type MultiError struct {
	Errors []error
}

func (e *MultiError) Error() string {
	var msgs []string
	for _, err := range e.Errors {
		msgs = append(msgs, err.Error())
	}
	return "multiple errors: " + joinStrings(msgs, "; ")
}

func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}
