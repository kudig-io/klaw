package events

import (
	"fmt"
	"sync"
	"time"

	"github.com/kudig-io/klaw/internal/messaging"
)

// Notifier 事件通知器
type Notifier struct {
	commManager    *messaging.Manager
	eventManager   *Manager
	rateLimiter    *RateLimiter
	mu             sync.RWMutex
	channels       map[string][]string // 事件源 -> 通信频道列表
	muteDuration   time.Duration
}

// RateLimiter 速率限制器
type RateLimiter struct {
	mu       sync.Mutex
	buckets  map[string]*tokenBucket
	rate     int           // 每秒允许的请求数
	capacity int           // 桶容量
}

type tokenBucket struct {
	tokens    int
	lastUpdate time.Time
}

// NewRateLimiter 创建速率限制器
func NewRateLimiter(rate, capacity int) *RateLimiter {
	return &RateLimiter{
		buckets:  make(map[string]*tokenBucket),
		rate:     rate,
		capacity: capacity,
	}
}

// Allow 检查是否允许请求
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	bucket, exists := rl.buckets[key]
	if !exists {
		bucket = &tokenBucket{
			tokens:     rl.capacity - 1,
			lastUpdate: time.Now(),
		}
		rl.buckets[key] = bucket
		return true
	}
	
	// 计算新增的 token
	now := time.Now()
	elapsed := now.Sub(bucket.lastUpdate).Seconds()
	bucket.tokens += int(elapsed * float64(rl.rate))
	if bucket.tokens > rl.capacity {
		bucket.tokens = rl.capacity
	}
	bucket.lastUpdate = now
	
	if bucket.tokens > 0 {
		bucket.tokens--
		return true
	}
	
	return false
}

// NewNotifier 创建事件通知器
func NewNotifier(commManager *messaging.Manager, eventManager *Manager) *Notifier {
	return &Notifier{
		commManager:  commManager,
		eventManager: eventManager,
		rateLimiter:  NewRateLimiter(10, 20), // 每秒10个，最多累积20个
		channels:     make(map[string][]string),
		muteDuration: 5 * time.Minute, // 相同事件静音5分钟
	}
}

// SetMuteDuration 设置事件静音时长
func (n *Notifier) SetMuteDuration(duration time.Duration) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.muteDuration = duration
}

// SubscribeSource 订阅事件源
func (n *Notifier) SubscribeSource(sourceName string, channelIDs ...string) error {
	source, ok := n.eventManager.Get(sourceName)
	if !ok {
		return fmt.Errorf("event source %s not found", sourceName)
	}
	
	n.mu.Lock()
	n.channels[sourceName] = channelIDs
	n.mu.Unlock()
	
	// 订阅事件
	source.Subscribe(n.handleEvent)
	
	return nil
}

// UnsubscribeSource 取消订阅事件源
func (n *Notifier) UnsubscribeSource(sourceName string) error {
	_, ok := n.eventManager.Get(sourceName)
	if !ok {
		return fmt.Errorf("event source %s not found", sourceName)
	}
	
	n.mu.Lock()
	delete(n.channels, sourceName)
	n.mu.Unlock()
	
	// 取消订阅（简化处理，实际应该保存 handler 引用）
	// source.Unsubscribe(n.handleEvent)
	
	return nil
}

// handleEvent 处理事件
func (n *Notifier) handleEvent(event *Event) {
	// 速率限制检查
	key := fmt.Sprintf("%s/%s/%s", event.Cluster, event.ResourceType, event.Reason)
	if !n.rateLimiter.Allow(key) {
		// 被限流，记录日志但不发送
		fmt.Printf("Event rate limited: %s\n", key)
		return
	}
	
	n.mu.RLock()
	channels := n.channels[event.Cluster]
	n.mu.RUnlock()
	
	if len(channels) == 0 {
		// 如果没有配置特定频道，发送到所有平台
		n.broadcast(event)
	} else {
		// 发送到指定频道
		for _, channelID := range channels {
			n.sendToChannel(channelID, event)
		}
	}
}

// broadcast 广播事件到所有通信平台
func (n *Notifier) broadcast(event *Event) {
	// 发送到所有通信平台
	// 这里简化处理，实际应该遍历所有已注册的 communicator
	// n.commManager.SendToAll("", event.ToMarkdown())
	
	fmt.Printf("[Event] %s\n", event.ToSummary())
}

// sendToChannel 发送事件到指定频道
func (n *Notifier) sendToChannel(channelID string, event *Event) {
	response := &messaging.Response{
		Content: event.ToMarkdown(),
		Format:  messaging.FormatMarkdown,
	}
	
	// 获取通信平台
	comm, ok := n.commManager.GetCommunicator("dingtalk")
	if ok {
		if err := comm.SendMessage(channelID, response); err != nil {
			fmt.Printf("Failed to send event to channel %s: %v\n", channelID, err)
		}
	}
}

// EventAggregator 事件聚合器
type EventAggregator struct {
	mu         sync.Mutex
	events     map[string]*aggregatedEvent
	window     time.Duration
	threshold  int
}

type aggregatedEvent struct {
	event     *Event
	count     int
	firstTime time.Time
	lastTime  time.Time
}

// NewEventAggregator 创建事件聚合器
func NewEventAggregator(window time.Duration, threshold int) *EventAggregator {
	return &EventAggregator{
		events:    make(map[string]*aggregatedEvent),
		window:    window,
		threshold: threshold,
	}
}

// Add 添加事件
func (ea *EventAggregator) Add(event *Event) *Event {
	ea.mu.Lock()
	defer ea.mu.Unlock()
	
	key := fmt.Sprintf("%s/%s/%s/%s", event.Cluster, event.Namespace, event.ResourceType, event.Reason)
	
	agg, exists := ea.events[key]
	if !exists || time.Since(agg.firstTime) > ea.window {
		ea.events[key] = &aggregatedEvent{
			event:     event,
			count:     1,
			firstTime: time.Now(),
			lastTime:  time.Now(),
		}
		return event
	}
	
	agg.count++
	agg.lastTime = time.Now()
	
	// 如果超过阈值，返回聚合事件
	if agg.count >= ea.threshold {
		agg.count = 0
		agg.firstTime = time.Now()
		
		// 创建聚合事件
		return &Event{
			ID:           event.ID,
			Type:         event.Type,
			ResourceType: event.ResourceType,
			ResourceName: fmt.Sprintf("%s (and %d similar events)", event.ResourceName, ea.threshold-1),
			Namespace:    event.Namespace,
			Cluster:      event.Cluster,
			Reason:       event.Reason,
			Message:      fmt.Sprintf("[%d similar events in %v] %s", ea.threshold, ea.window, event.Message),
			Timestamp:    time.Now(),
			InvolvedObject: event.InvolvedObject,
		}
	}
	
	return nil // 不发送，等待聚合
}

// dedupKey 生成去重 key
func dedupKey(event *Event) string {
	return fmt.Sprintf("%s:%s:%s:%s:%s", 
		event.Cluster, 
		event.Namespace, 
		event.ResourceType, 
		event.ResourceName, 
		event.Reason,
	)
}

// EventDedup 事件去重器
type EventDedup struct {
	mu      sync.Mutex
	seen    map[string]time.Time
	window  time.Duration
}

// NewEventDedup 创建事件去重器
func NewEventDedup(window time.Duration) *EventDedup {
	return &EventDedup{
		seen:   make(map[string]time.Time),
		window: window,
	}
}

// IsDuplicate 检查是否是重复事件
func (ed *EventDedup) IsDuplicate(event *Event) bool {
	ed.mu.Lock()
	defer ed.mu.Unlock()
	
	key := dedupKey(event)
	
	// 清理过期记录
	now := time.Now()
	for k, t := range ed.seen {
		if now.Sub(t) > ed.window {
			delete(ed.seen, k)
		}
	}
	
	// 检查是否已存在
	if lastSeen, ok := ed.seen[key]; ok {
		if now.Sub(lastSeen) < ed.window {
			return true
		}
	}
	
	// 记录本次事件
	ed.seen[key] = now
	return false
}
