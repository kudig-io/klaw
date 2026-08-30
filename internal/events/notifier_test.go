package events

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/kudig-io/klaw/internal/messaging"
)

// ---- 测试替身 ----

type sentMessage struct {
	channelID string
	content   string
}

// fakeCommunicator 记录所有发送，用于断言投递结果
type fakeCommunicator struct {
	mu   sync.Mutex
	name string
	sent []sentMessage
}

func (f *fakeCommunicator) Name() string  { return f.name }
func (f *fakeCommunicator) Start() error  { return nil }
func (f *fakeCommunicator) Stop() error   { return nil }
func (f *fakeCommunicator) RegisterHandler(_ messaging.MessageHandler) {}
func (f *fakeCommunicator) IsHealthy() bool { return true }
func (f *fakeCommunicator) SendMessage(channelID string, resp *messaging.Response) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.sent = append(f.sent, sentMessage{channelID: channelID, content: resp.Content})
	return nil
}
func (f *fakeCommunicator) count() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.sent)
}
func (f *fakeCommunicator) last() sentMessage {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.sent[len(f.sent)-1]
}

// fakeSource 满足 Source 接口的最小实现
type fakeSource struct {
	*BaseSource
}

func (f *fakeSource) Start(_ context.Context) error { return nil }
func (f *fakeSource) Stop() error                   { return nil }

// newTestNotifier 构造带两个 fake 平台的 Notifier
func newTestNotifier(t *testing.T, opts NotifierOptions) (*Notifier, *fakeCommunicator, *fakeCommunicator) {
	t.Helper()
	dingtalk := &fakeCommunicator{name: "dingtalk"}
	feishu := &fakeCommunicator{name: "feishu"}
	mgr := messaging.NewManager(nil)
	mgr.RegisterCommunicator("dingtalk", dingtalk)
	mgr.RegisterCommunicator("feishu", feishu)

	em := NewManager()
	src := &fakeSource{BaseSource: NewBaseSource("kubernetes-prod")}
	em.Register(src)

	return NewNotifier(mgr, em, opts), dingtalk, feishu
}

func testEvent(reason string) *Event {
	return &Event{
		ID:           fmt.Sprintf("evt-%d", time.Now().UnixNano()),
		Type:         EventTypeWarning,
		ResourceType: ResourcePod,
		ResourceName: "nginx-1",
		Namespace:    "default",
		Cluster:      "prod",
		Reason:       reason,
		Message:      "test message",
		Timestamp:    time.Now(),
	}
}

// ---- E-1 回归：channels key 规范化 + 定向投递 + 不再硬编码 dingtalk ----

func TestNotifierRoutesToSubscribedChannels(t *testing.T) {
	n, dingtalk, feishu := newTestNotifier(t, NotifierOptions{})
	if err := n.SubscribeSource("kubernetes-prod", "ops-alert"); err != nil {
		t.Fatalf("subscribe: %v", err)
	}

	// 事件携带的 Cluster 是集群名（不带 "kubernetes-" 前缀），必须仍能命中订阅
	n.handleEvent(testEvent("BackOff"))

	if dingtalk.count() != 1 {
		t.Fatalf("dingtalk should receive exactly 1 message, got %d", dingtalk.count())
	}
	if feishu.count() != 1 {
		t.Fatalf("feishu should receive exactly 1 message, got %d", feishu.count())
	}
	if got := dingtalk.last().channelID; got != "ops-alert" {
		t.Errorf("expected channel ops-alert, got %q", got)
	}
}

func TestNotifierBroadcastWhenNoChannels(t *testing.T) {
	n, dingtalk, feishu := newTestNotifier(t, NotifierOptions{})
	if err := n.SubscribeSource("kubernetes-prod"); err != nil {
		t.Fatalf("subscribe: %v", err)
	}

	n.handleEvent(testEvent("BackOff"))

	if dingtalk.count() != 1 || feishu.count() != 1 {
		t.Fatalf("broadcast should reach all platforms, got dingtalk=%d feishu=%d", dingtalk.count(), feishu.count())
	}
	if !strings.Contains(dingtalk.last().content, "BackOff") {
		t.Errorf("message content should be markdown of the event, got %q", dingtalk.last().content)
	}
}

// ---- E-2 回归：去重 / 聚合 / 限流接线 ----

func TestNotifierDedupSuppressesDuplicates(t *testing.T) {
	n, dingtalk, _ := newTestNotifier(t, NotifierOptions{DedupWindow: time.Second})
	if err := n.SubscribeSource("kubernetes-prod"); err != nil {
		t.Fatalf("subscribe: %v", err)
	}

	n.handleEvent(testEvent("BackOff"))
	n.handleEvent(testEvent("BackOff")) // 同 key，窗口内重复

	if dingtalk.count() != 1 {
		t.Fatalf("duplicate event must be suppressed, got %d deliveries", dingtalk.count())
	}
}

func TestNotifierAggregation(t *testing.T) {
	n, dingtalk, _ := newTestNotifier(t, NotifierOptions{AggregateWindow: time.Second, AggregateThreshold: 3})
	if err := n.SubscribeSource("kubernetes-prod"); err != nil {
		t.Fatalf("subscribe: %v", err)
	}

	n.handleEvent(testEvent("BackOff"))    // 首条透传
	n.handleEvent(testEvent("BackOff"))    // 攒批
	n.handleEvent(testEvent("BackOff"))    // 达到阈值 → 聚合事件

	if dingtalk.count() != 2 {
		t.Fatalf("expected 2 deliveries (first + aggregated), got %d", dingtalk.count())
	}
	agg := dingtalk.last()
	if !strings.Contains(agg.content, "similar events") {
		t.Errorf("second delivery should be aggregated event, got %q", agg.content)
	}
}

func TestNotifierRateLimit(t *testing.T) {
	n, dingtalk, _ := newTestNotifier(t, NotifierOptions{RateLimit: 1, Burst: 1})
	if err := n.SubscribeSource("kubernetes-prod"); err != nil {
		t.Fatalf("subscribe: %v", err)
	}

	n.handleEvent(testEvent("BackOff"))
	n.handleEvent(testEvent("BackOff")) // 同 key 第二条：burst 已耗尽且时间未推进

	if dingtalk.count() != 1 {
		t.Fatalf("second same-key event should be rate limited, got %d deliveries", dingtalk.count())
	}

	// 不同 key 不受影响（各自独立桶）
	other := testEvent("OOMKilled")
	n.handleEvent(other)
	if dingtalk.count() != 2 {
		t.Fatalf("different key should not be affected, got %d deliveries", dingtalk.count())
	}
}

// ---- E-3 回归：handler panic 不拖垮进程、不影响其他 handler ----

func TestEmitRecoversHandlerPanic(t *testing.T) {
	src := &fakeSource{BaseSource: NewBaseSource("panic-src")}

	received := make(chan string, 1)
	src.Subscribe(func(e *Event) { panic("boom") })
	src.Subscribe(func(e *Event) { received <- e.Reason })

	done := make(chan struct{})
	go func() {
		defer close(done)
		src.emit(&Event{Reason: "BackOff"}) // 若 panic 未恢复，此处进程崩溃
	}()

	select {
	case reason := <-received:
		if reason != "BackOff" {
			t.Errorf("unexpected reason %q", reason)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("second handler did not receive event")
	}
	<-done
}

// ---- UnsubscribeSource 真正退订 ----

func TestUnsubscribeSourceStopsDelivery(t *testing.T) {
	em := NewManager()
	src := &fakeSource{BaseSource: NewBaseSource("kubernetes-prod")}
	em.Register(src)

	dingtalk := &fakeCommunicator{name: "dingtalk"}
	mgr := messaging.NewManager(nil)
	mgr.RegisterCommunicator("dingtalk", dingtalk)
	n := NewNotifier(mgr, em, NotifierOptions{})

	if err := n.SubscribeSource("kubernetes-prod"); err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	if err := n.UnsubscribeSource("kubernetes-prod"); err != nil {
		t.Fatalf("unsubscribe: %v", err)
	}

	// 直接经源 emit（异步 goroutine），等待后断言无投递
	src.emit(testEvent("BackOff"))
	deadline := time.After(500 * time.Millisecond)
	for dingtalk.count() == 0 {
		select {
		case <-deadline:
			return // 未收到即通过
		case <-time.After(10 * time.Millisecond):
		}
	}
	t.Fatalf("event delivered after unsubscribe: %+v", dingtalk.last())
}
