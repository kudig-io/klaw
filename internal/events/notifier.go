package events

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/kudig-io/klaw/internal/messaging"
)

// NotifierOptions 通知器可调参数，零值字段使用内置默认。
type NotifierOptions struct {
	// RateLimit 每秒最大事件数（默认 10）；Burst 为令牌桶容量（默认 2×RateLimit）
	RateLimit int
	Burst     int
	// DedupWindow 相同事件（同集群/命名空间/资源/原因）的去重窗口，0 表示不去重。
	// 窗口自事件首次出现起算，窗口内的重复事件直接丢弃。
	DedupWindow time.Duration
	// AggregateWindow / AggregateThreshold 事件聚合：窗口内同类事件达到阈值后
	// 合并为一条聚合事件推送，0 阈值表示不聚合。
	AggregateWindow    time.Duration
	AggregateThreshold int
}

// Notifier 事件通知器：去重 → 限流 → 聚合 → 路由到通信平台
type Notifier struct {
	commManager  *messaging.Manager
	eventManager *Manager
	rateLimiter  *RateLimiter
	dedup        *EventDedup
	aggregator   *EventAggregator
	mu           sync.RWMutex
	channels     map[string][]string   // 事件源（及其集群名）-> 通信频道列表
	subscribed   map[string]EventHandler // 事件源 -> 已注册的 handler（用于可靠退订）
}

// RateLimiter 速率限制器（令牌桶，按 key 维度）
type RateLimiter struct {
	mu       sync.Mutex
	buckets  map[string]*tokenBucket
	rate     int // 每秒允许的请求数
	capacity int // 桶容量
}

type tokenBucket struct {
	tokens     int
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

// cleanup 惰性清理长期不活跃的桶，防止 map 随事件源维度无限增长
func (rl *RateLimiter) cleanup(idle time.Duration) {
	deadline := time.Now().Add(-idle)
	for k, b := range rl.buckets {
		if b.lastUpdate.Before(deadline) {
			delete(rl.buckets, k)
		}
	}
}

// NewNotifier 创建事件通知器
func NewNotifier(commManager *messaging.Manager, eventManager *Manager, opts NotifierOptions) *Notifier {
	rate := opts.RateLimit
	if rate <= 0 {
		rate = 10
	}
	burst := opts.Burst
	if burst <= 0 {
		burst = rate * 2
	}
	n := &Notifier{
		commManager:  commManager,
		eventManager: eventManager,
		rateLimiter:  NewRateLimiter(rate, burst),
		channels:     make(map[string][]string),
		subscribed:   make(map[string]EventHandler),
	}
	if opts.DedupWindow > 0 {
		n.dedup = NewEventDedup(opts.DedupWindow)
	}
	if opts.AggregateThreshold > 0 {
		window := opts.AggregateWindow
		if window <= 0 {
			window = time.Minute
		}
		n.aggregator = NewEventAggregator(window, opts.AggregateThreshold)
	}
	return n
}

// SubscribeSource 订阅事件源并绑定通信频道；不指定频道时广播到所有平台
func (n *Notifier) SubscribeSource(sourceName string, channelIDs ...string) error {
	source, ok := n.eventManager.Get(sourceName)
	if !ok {
		return fmt.Errorf("event source %s not found", sourceName)
	}

	n.mu.Lock()
	// 同时按源名与集群名（剥去 "kubernetes-" 前缀）建索引：
	// 事件携带的 Cluster 字段是集群名，而 source.Name() 带 "kubernetes-" 前缀
	for _, key := range notifierChannelKeys(sourceName) {
		n.channels[key] = channelIDs
	}
	// 保存 handler 引用，退订时传入同一 func 值才能命中
	handler := EventHandler(n.handleEvent)
	n.subscribed[sourceName] = handler
	n.mu.Unlock()

	source.Subscribe(handler)
	return nil
}

// UnsubscribeSource 取消订阅事件源
func (n *Notifier) UnsubscribeSource(sourceName string) error {
	source, ok := n.eventManager.Get(sourceName)
	if !ok {
		return fmt.Errorf("event source %s not found", sourceName)
	}

	n.mu.Lock()
	handler, ok := n.subscribed[sourceName]
	if ok {
		delete(n.subscribed, sourceName)
	}
	for _, key := range notifierChannelKeys(sourceName) {
		delete(n.channels, key)
	}
	n.mu.Unlock()

	if ok {
		source.Unsubscribe(handler)
	}
	return nil
}

// notifierChannelKeys 生成事件的频道路由键：源名本身 + 剥去 "kubernetes-" 前缀的集群名
func notifierChannelKeys(sourceName string) []string {
	keys := []string{sourceName}
	if cluster := strings.TrimPrefix(sourceName, "kubernetes-"); cluster != sourceName {
		keys = append(keys, cluster)
	}
	return keys
}

// handleEvent 处理事件：去重 → 限流 → 聚合 → 路由
func (n *Notifier) handleEvent(event *Event) {
	// 1. 去重：窗口内的相同事件直接丢弃（不消耗限流配额）
	if n.dedup != nil && n.dedup.IsDuplicate(event) {
		return
	}

	// 2. 限流：按 集群/资源类型/原因 维度限速
	key := fmt.Sprintf("%s/%s/%s", event.Cluster, event.ResourceType, event.Reason)
	if !n.rateLimiter.Allow(key) {
		fmt.Printf("Event rate limited: %s\n", key)
		return
	}

	// 3. 聚合：窗口内同类事件攒批，达到阈值后合并推送
	if n.aggregator != nil {
		aggregated := n.aggregator.Add(event)
		if aggregated == nil {
			return // 未达阈值，继续攒批
		}
		event = aggregated
	}

	// 4. 路由：有指定频道则定向发送，否则广播到所有平台
	n.mu.RLock()
	channels := n.channels[event.Cluster]
	n.mu.RUnlock()

	if len(channels) == 0 {
		n.broadcast(event)
		return
	}
	for _, channelID := range channels {
		n.sendToChannel(channelID, event)
	}
}

// broadcast 广播事件到所有通信平台
func (n *Notifier) broadcast(event *Event) {
	if n.commManager == nil {
		fmt.Printf("[Event] %s\n", event.ToSummary())
		return
	}
	response := &messaging.Response{
		Content: event.ToMarkdown(),
		Format:  messaging.FormatMarkdown,
	}
	if err := n.commManager.SendToAll("", response); err != nil {
		fmt.Printf("Failed to broadcast event: %v\n", err)
	}
}

// sendToChannel 发送事件到指定频道（所有已注册平台）
func (n *Notifier) sendToChannel(channelID string, event *Event) {
	if n.commManager == nil {
		fmt.Printf("[Event] %s\n", event.ToSummary())
		return
	}
	response := &messaging.Response{
		Content: event.ToMarkdown(),
		Format:  messaging.FormatMarkdown,
	}
	if err := n.commManager.SendToAll(channelID, response); err != nil {
		fmt.Printf("Failed to send event to channel %s: %v\n", channelID, err)
	}
}

// EventAggregator 事件聚合器：窗口内同类事件达到阈值后合并为一条聚合事件
type EventAggregator struct {
	mu        sync.Mutex
	events    map[string]*aggregatedEvent
	window    time.Duration
	threshold int
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

// Add 添加事件；返回值：nil 表示已攒批（未达阈值），非 nil 表示应推送的事件
// （首条事件原样透传以便及时告警，达到阈值时返回聚合事件）
func (ea *EventAggregator) Add(event *Event) *Event {
	ea.mu.Lock()
	defer ea.mu.Unlock()

	key := fmt.Sprintf("%s/%s/%s/%s", event.Cluster, event.Namespace, event.ResourceType, event.Reason)

	agg, exists := ea.events[key]
	if !exists {
		// 首条事件立即透传，保证严重问题不被延迟
		ea.events[key] = &aggregatedEvent{
			event:     event,
			count:     1,
			firstTime: time.Now(),
			lastTime:  time.Now(),
		}
		return event
	}
	if time.Since(agg.firstTime) > ea.window {
		// 窗口过期，重新起算
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

	// 达到阈值，返回聚合事件并重置计数
	if agg.count >= ea.threshold {
		count := agg.count
		agg.count = 0
		agg.firstTime = time.Now()

		return &Event{
			ID:           event.ID,
			Type:         event.Type,
			ResourceType: event.ResourceType,
			ResourceName: fmt.Sprintf("%s (and %d similar events)", event.ResourceName, count-1),
			Namespace:    event.Namespace,
			Cluster:      event.Cluster,
			Reason:       event.Reason,
			Message:      fmt.Sprintf("[%d similar events in %v] %s", count, ea.window, event.Message),
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

// EventDedup 事件去重器：窗口自事件首次出现起算，窗口内重复事件直接丢弃
type EventDedup struct {
	mu     sync.Mutex
	seen   map[string]time.Time
	window time.Duration
}

// NewEventDedup 创建事件去重器
func NewEventDedup(window time.Duration) *EventDedup {
	return &EventDedup{
		seen:   make(map[string]time.Time),
		window: window,
	}
}

// IsDuplicate 检查是否是重复事件（非重复时记录本次时间）
func (ed *EventDedup) IsDuplicate(event *Event) bool {
	ed.mu.Lock()
	defer ed.mu.Unlock()

	key := dedupKey(event)
	now := time.Now()

	// 惰性清理过期记录
	for k, t := range ed.seen {
		if now.Sub(t) > ed.window {
			delete(ed.seen, k)
		}
	}

	if lastSeen, ok := ed.seen[key]; ok {
		if now.Sub(lastSeen) < ed.window {
			return true
		}
	}

	ed.seen[key] = now
	return false
}
