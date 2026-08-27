package sos

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	corev1 "k8s.io/api/core/v1"

	"github.com/kudig-io/klaw/internal/config"
)

type nopReader struct{}

func (nopReader) ListPods(_, _ string) ([]corev1.Pod, error)         { return nil, nil }
func (nopReader) ListNodes(string) ([]corev1.Node, error)            { return nil, nil }
func (nopReader) ListEvents(_, _ string) ([]corev1.Event, error)     { return nil, nil }
func (nopReader) GetPodLogs(_, _, _ string, _ int64) (string, error) { return "", nil }

// fakeUpstream 模拟 DashScope：回放预置事件，收到上行音频后退出；
// toolCall 模式下改为下发一次工具调用并持续记录上行消息
type fakeUpstream struct {
	t         *testing.T
	toolCall  bool   // true：下发 response.function_call_arguments.done 并持续服务
	toolName  string // 工具名，缺省 get_cluster_status
	toolArgs  string // 工具参数 JSON，缺省 {}
	waitAudio bool   // true：等首条 input_audio_buffer.append 后再下发工具调用
	mu        sync.Mutex
	received  []map[string]any
	conns     int // 累计连接数
}

func (f *fakeUpstream) appendReceived(m map[string]any) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.received = append(f.received, m)
}

func (f *fakeUpstream) hasReceived(typ string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, m := range f.received {
		if m["type"] == typ {
			return true
		}
	}
	return false
}

// connCount 累计连接数（验证重连已发生）
func (f *fakeUpstream) connCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.conns
}

// lastOf 返回最后一条指定类型的消息
func (f *fakeUpstream) lastOf(typ string) map[string]any {
	f.mu.Lock()
	defer f.mu.Unlock()
	for i := len(f.received) - 1; i >= 0; i-- {
		if f.received[i]["type"] == typ {
			return f.received[i]
		}
	}
	return nil
}

func (f *fakeUpstream) serve(w http.ResponseWriter, r *http.Request) {
	c, err := (&websocket.Upgrader{}).Upgrade(w, r, nil)
	if err != nil {
		f.t.Logf("upgrade: %v", err)
		return
	}
	defer c.Close()
	f.mu.Lock()
	f.conns++
	f.mu.Unlock()
	if f.toolCall {
		name := f.toolName
		if name == "" {
			name = "get_cluster_status"
		}
		args := f.toolArgs
		if args == "" {
			args = "{}"
		}
		emit := func() {
			_ = c.WriteJSON(map[string]any{
				"type":      "response.function_call_arguments.done",
				"call_id":   "c1",
				"name":      name,
				"arguments": args,
			})
		}
		if !f.waitAudio {
			emit()
		}
	} else {
		// 第一条消息应为 session.update；这里直接回放事件
		_ = c.WriteJSON(map[string]any{"type": "session.created", "session": map[string]any{"model": "m", "voice": "v"}})
		_ = c.WriteJSON(map[string]any{"type": "response.audio.delta", "delta": "AQI="}) // base64 {1,2}
		_ = c.WriteJSON(map[string]any{"type": "response.audio_transcript.delta", "delta": "你好"})
		_ = c.WriteJSON(map[string]any{"type": "input_audio_buffer.speech_started"})
	}
	pending := f.toolCall && f.waitAudio // 等首条 append 再下发工具调用
	for {
		mt, data, err := c.ReadMessage()
		if err != nil {
			return
		}
		if mt == websocket.TextMessage {
			var m map[string]any
			_ = json.Unmarshal(data, &m)
			f.appendReceived(m)
			if pending && m["type"] == "input_audio_buffer.append" {
				name := f.toolName
				if name == "" {
					name = "get_cluster_status"
				}
				args := f.toolArgs
				if args == "" {
					args = "{}"
				}
				_ = c.WriteJSON(map[string]any{
					"type":      "response.function_call_arguments.done",
					"call_id":   "c1",
					"name":      name,
					"arguments": args,
				})
				pending = false
			}
			if !f.toolCall && m["type"] == "input_audio_buffer.append" {
				return // 收到浏览器音频即结束
			}
		}
	}
}

func testSOSConfig() config.SOSConfig {
	return config.SOSConfig{
		Enabled: true,
		Dashscope: config.SOSDashscopeConfig{
			APIKey: "k", WorkspaceID: "ws", Region: "cn-beijing",
			Model: "m", Voice: "Ethan",
		},
	}
}

func TestManagerStatus(t *testing.T) {
	m, err := NewManager(testSOSConfig(), nopReader{}, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	st := m.Status()
	if !st.Enabled || !st.Ready || st.FAQCount < 1 || st.Model != "m" || st.Voice != "Ethan" {
		t.Fatalf("status = %+v", st)
	}
	cfg := testSOSConfig()
	cfg.Dashscope.APIKey = ""
	m2, err := NewManager(cfg, nopReader{}, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	if m2.Status().Ready {
		t.Fatal("expected not ready without api key")
	}
}

func TestSessionBridge(t *testing.T) {
	up := &fakeUpstream{t: t}
	upSrv := httptest.NewServer(http.HandlerFunc(up.serve))
	defer upSrv.Close()
	u := strings.Replace(upSrv.URL, "http://", "ws://", 1)

	orig := Dial
	origRetry := retryDialDelay
	defer func() { Dial = orig; retryDialDelay = origRetry }()
	// 缩短重连延迟：上游收到音频后关闭触发的重连在本测试内完成，避免遗留 goroutine
	retryDialDelay = 20 * time.Millisecond
	Dial = func(ctx context.Context, c config.SOSConfig) (*websocket.Conn, error) {
		conn, _, err := websocket.DefaultDialer.DialContext(ctx, u, nil)
		return conn, err
	}

	m, err := NewManager(testSOSConfig(), nopReader{}, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(http.HandlerFunc(m.HandleSessionWS))
	defer srv.Close()

	wsURL := strings.Replace(srv.URL, "http://", "ws://", 1)
	browser, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer browser.Close()
	_ = browser.SetReadDeadline(time.Now().Add(5 * time.Second))

	// 1) 收到 session 帧
	var ev ClientEvent
	if err := browser.ReadJSON(&ev); err != nil || ev.Type != "session" {
		t.Fatalf("first event = %+v err=%v", ev, err)
	}
	// 2) 收到音频二进制帧 {1,2}
	mt, data, err := browser.ReadMessage()
	if err != nil || mt != websocket.BinaryMessage || len(data) != 2 {
		t.Fatalf("audio frame mt=%d data=%v err=%v", mt, data, err)
	}
	// 3) 助理字幕
	if err := browser.ReadJSON(&ev); err != nil || ev.Type != "assistant.transcript.delta" || ev.Delta != "你好" {
		t.Fatalf("subtitle = %+v err=%v", ev, err)
	}
	// 4) 打断事件
	if err := browser.ReadJSON(&ev); err != nil || ev.Type != "speech_started" {
		t.Fatalf("speech_started = %+v err=%v", ev, err)
	}
	// 5) 上行音频 → 上游收到 append
	_ = browser.WriteMessage(websocket.BinaryMessage, []byte{1, 0, 2, 0})
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if up.hasReceived("input_audio_buffer.append") {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if !up.hasReceived("input_audio_buffer.append") {
		t.Fatal("upstream never received input_audio_buffer.append")
	}
	// 上游收到音频后关闭会触发一次重连；等重连完成再结束，避免遗留 goroutine
	deadline = time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if up.connCount() >= 2 {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func TestHandleSessionNotConfigured(t *testing.T) {
	cfg := config.SOSConfig{Enabled: false}
	m, err := NewManager(cfg, nopReader{}, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/sos/session", nil)
	rec := httptest.NewRecorder()
	m.HandleSessionWS(rec, req)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}
}

// TestSessionToolCallFlow 工具调用闭环：上游下发 function_call_arguments.done 后，
// 后端回传 conversation.item.create(function_call_output)，并追加 response.create 触发模型继续作答
func TestSessionToolCallFlow(t *testing.T) {
	up := &fakeUpstream{t: t, toolCall: true}
	upSrv := httptest.NewServer(http.HandlerFunc(up.serve))
	defer upSrv.Close()
	u := strings.Replace(upSrv.URL, "http://", "ws://", 1)

	orig := Dial
	defer func() { Dial = orig }()
	Dial = func(ctx context.Context, c config.SOSConfig) (*websocket.Conn, error) {
		conn, _, err := websocket.DefaultDialer.DialContext(ctx, u, nil)
		return conn, err
	}

	m, err := NewManager(testSOSConfig(), nopReader{}, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(http.HandlerFunc(m.HandleSessionWS))
	defer srv.Close()

	browser, _, err := websocket.DefaultDialer.Dial(strings.Replace(srv.URL, "http://", "ws://", 1), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer browser.Close()
	_ = browser.SetReadDeadline(time.Now().Add(5 * time.Second))

	// 浏览器收到 tool_call 通知
	var ev ClientEvent
	if err := browser.ReadJSON(&ev); err != nil || ev.Type != "tool_call" || ev.Name != "get_cluster_status" {
		t.Fatalf("tool_call event = %+v err=%v", ev, err)
	}

	// 上游收到 function_call_output 与 response.create 两条消息
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if up.hasReceived("conversation.item.create") && up.hasReceived("response.create") {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if !up.hasReceived("conversation.item.create") {
		t.Fatal("upstream never received conversation.item.create(function_call_output)")
	}
	if !up.hasReceived("response.create") {
		t.Fatal("upstream never received response.create after tool output")
	}
}

// TestSessionUpstreamReconnect 上游第一次连接正常关闭后自动重连一次，浏览器会话续存
func TestSessionUpstreamReconnect(t *testing.T) {
	var mu sync.Mutex
	connCount := 0
	upSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := (&websocket.Upgrader{}).Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer c.Close()
		mu.Lock()
		connCount++
		n := connCount
		mu.Unlock()
		if n == 1 {
			// 读取 session.update 后立即正常关闭，模拟上游断线
			_, _, _ = c.ReadMessage()
			_ = c.WriteControl(websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""), time.Now().Add(time.Second))
			time.Sleep(100 * time.Millisecond)
			return
		}
		// 第二次连接：正常服务，回放事件
		_ = c.WriteJSON(map[string]any{"type": "session.created", "session": map[string]any{"model": "m2", "voice": "v2"}})
		_ = c.WriteJSON(map[string]any{"type": "response.audio_transcript.delta", "delta": "重连成功"})
		for {
			if _, _, err := c.ReadMessage(); err != nil {
				return
			}
		}
	}))
	defer upSrv.Close()
	u := strings.Replace(upSrv.URL, "http://", "ws://", 1)

	origDial, origRetry := Dial, retryDialDelay
	defer func() { Dial = origDial; retryDialDelay = origRetry }()
	retryDialDelay = 10 * time.Millisecond
	Dial = func(ctx context.Context, c config.SOSConfig) (*websocket.Conn, error) {
		conn, _, err := websocket.DefaultDialer.DialContext(ctx, u, nil)
		return conn, err
	}

	m, err := NewManager(testSOSConfig(), nopReader{}, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(http.HandlerFunc(m.HandleSessionWS))
	defer srv.Close()

	browser, _, err := websocket.DefaultDialer.Dial(strings.Replace(srv.URL, "http://", "ws://", 1), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer browser.Close()
	_ = browser.SetReadDeadline(time.Now().Add(10 * time.Second))

	// 第一次连接的 session 帧
	var ev ClientEvent
	if err := browser.ReadJSON(&ev); err != nil || ev.Type != "session" {
		t.Fatalf("first session = %+v err=%v", ev, err)
	}
	// 重连后第二次连接的回放事件：会话续存
	for {
		if err := browser.ReadJSON(&ev); err != nil {
			t.Fatalf("session lost after upstream reconnect: %v", err)
		}
		if ev.Type == "assistant.transcript.delta" && ev.Delta == "重连成功" {
			break
		}
	}
	// 主动结束会话，避免读协程泄漏
	_ = browser.WriteJSON(map[string]any{"type": "end"})
}

// TestSessionIdleTimeout 空闲超时：无音频活动超过阈值时下发 error 帧并关闭会话
func TestSessionIdleTimeout(t *testing.T) {
	upSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := (&websocket.Upgrader{}).Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer c.Close()
		// 保持连接但不发任何数据
		for {
			if _, _, err := c.ReadMessage(); err != nil {
				return
			}
		}
	}))
	defer upSrv.Close()
	u := strings.Replace(upSrv.URL, "http://", "ws://", 1)

	origDial := Dial
	origIdle, origCheck := idleTimeout, idleCheckInterval
	defer func() {
		Dial = origDial
		idleTimeout, idleCheckInterval = origIdle, origCheck
	}()
	idleTimeout, idleCheckInterval = 30*time.Millisecond, 30*time.Millisecond
	Dial = func(ctx context.Context, c config.SOSConfig) (*websocket.Conn, error) {
		conn, _, err := websocket.DefaultDialer.DialContext(ctx, u, nil)
		return conn, err
	}

	m, err := NewManager(testSOSConfig(), nopReader{}, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(http.HandlerFunc(m.HandleSessionWS))
	defer srv.Close()

	browser, _, err := websocket.DefaultDialer.Dial(strings.Replace(srv.URL, "http://", "ws://", 1), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer browser.Close()
	_ = browser.SetReadDeadline(time.Now().Add(5 * time.Second))

	var ev ClientEvent
	if err := browser.ReadJSON(&ev); err != nil || ev.Type != "session.idle_timeout" {
		t.Fatalf("expected session.idle_timeout frame, got %+v err=%v", ev, err)
	}
	// 会话应已关闭：后续读取收到关闭/错误
	_ = browser.SetReadDeadline(time.Now().Add(3 * time.Second))
	if _, _, err := browser.ReadMessage(); err == nil {
		t.Fatal("expected connection closed after idle timeout")
	}
}

// wsPair 建立一对本地 WebSocket 连接（浏览器端 / 上游端）
func wsPair(t *testing.T) (browser, upstream *websocket.Conn) {
	t.Helper()
	upCh := make(chan *websocket.Conn, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := (&websocket.Upgrader{}).Upgrade(w, r, nil)
		if err != nil {
			return
		}
		upCh <- c
	}))
	t.Cleanup(srv.Close)
	conn, _, err := websocket.DefaultDialer.Dial(strings.Replace(srv.URL, "http://", "ws://", 1), nil)
	if err != nil {
		t.Fatal(err)
	}
	browser = conn
	select {
	case upstream = <-upCh:
	case <-time.After(2 * time.Second):
		t.Fatal("upstream side never upgraded")
	}
	return browser, upstream
}

// TestDispatchToolOutputAndJSON 工具成功/失败路径：回传 function_call_output + response.create；
// 错误信息经 json.Marshal 生成合法 JSON
func TestDispatchToolOutputAndJSON(t *testing.T) {
	m, err := NewManager(testSOSConfig(), nopReader{}, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	browser, upstream := wsPair(t)
	defer browser.Close()
	defer upstream.Close()
	s := &session{m: m, browser: browser, upstream: upstream, closed: make(chan struct{})}

	// 会话发往上游的消息到达 browser 端（wsPair 中 upstream 的对端），在此读取
	msgs := make(chan map[string]any, 16)
	go func() {
		for {
			_, data, err := browser.ReadMessage()
			if err != nil {
				return
			}
			var mm map[string]any
			_ = json.Unmarshal(data, &mm)
			msgs <- mm
		}
	}()

	// 成功路径：get_cluster_status
	s.dispatchTool(FunctionCall{CallID: "c1", Name: "get_cluster_status", Arguments: "{}"})
	var out1, cc1 map[string]any
	select {
	case out1 = <-msgs:
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting function_call_output")
	}
	select {
	case cc1 = <-msgs:
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting response.create")
	}
	if out1["type"] != "conversation.item.create" || cc1["type"] != "response.create" {
		t.Fatalf("unexpected messages: %v / %v", out1, cc1)
	}

	// 失败路径：未知工具，错误信息必须为合法 JSON
	s.dispatchTool(FunctionCall{CallID: "c2", Name: "no_such_tool", Arguments: "{}"})
	var out2, cc2 map[string]any
	select {
	case out2 = <-msgs:
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting error function_call_output")
	}
	select {
	case cc2 = <-msgs:
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting response.create after error output")
	}
	item, _ := out2["item"].(map[string]any)
	if item == nil {
		t.Fatalf("missing item: %v", out2)
	}
	outStr, _ := item["output"].(string)
	var parsed map[string]string
	if err := json.Unmarshal([]byte(outStr), &parsed); err != nil {
		t.Fatalf("tool error output is not valid JSON: %q err=%v", outStr, err)
	}
	if parsed["error"] == "" {
		t.Fatalf("expected error field in output, got %v", parsed)
	}
	if cc2["type"] != "response.create" {
		t.Fatalf("expected response.create, got %v", cc2)
	}
}

// TestManagerAuditCallback 审计回调：建连成功/会话结束/工具调用三处记录元数据；
// Dial 失败路径不 panic；未注入回调时各路径安全无操作
func TestManagerAuditCallback(t *testing.T) {
	var mu sync.Mutex
	var actions []string
	record := func(action, detail string) {
		mu.Lock()
		defer mu.Unlock()
		actions = append(actions, action+"|"+detail)
	}
	hasAction := func(prefix string) bool {
		mu.Lock()
		defer mu.Unlock()
		for _, a := range actions {
			if strings.HasPrefix(a, prefix) {
				return true
			}
		}
		return false
	}

	// Dial 失败路径：不 panic，且不记录会话开始
	m, err := NewManager(testSOSConfig(), nopReader{}, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	m.SetAuditLog(record)
	origDial := Dial
	defer func() { Dial = origDial }()
	Dial = func(ctx context.Context, c config.SOSConfig) (*websocket.Conn, error) {
		return nil, errors.New("dial refused")
	}
	srv := httptest.NewServer(http.HandlerFunc(m.HandleSessionWS))
	defer srv.Close()
	browser, _, err := websocket.DefaultDialer.Dial(strings.Replace(srv.URL, "http://", "ws://", 1), nil)
	if err != nil {
		t.Fatal(err)
	}
	_ = browser.SetReadDeadline(time.Now().Add(3 * time.Second))
	var ev ClientEvent
	_ = browser.ReadJSON(&ev) // error 帧
	browser.Close()
	if hasAction("sos.session_start") {
		t.Fatal("session_start should not be recorded when upstream dial fails")
	}

	// 完整会话路径：建连成功 + 工具调用 + 会话结束均有审计元数据
	up := &fakeUpstream{t: t, toolCall: true}
	upSrv := httptest.NewServer(http.HandlerFunc(up.serve))
	defer upSrv.Close()
	upURL := strings.Replace(upSrv.URL, "http://", "ws://", 1)
	Dial = func(ctx context.Context, c config.SOSConfig) (*websocket.Conn, error) {
		conn, _, err := websocket.DefaultDialer.DialContext(ctx, upURL, nil)
		return conn, err
	}
	browser2, _, err := websocket.DefaultDialer.Dial(strings.Replace(srv.URL, "http://", "ws://", 1), nil)
	if err != nil {
		t.Fatal(err)
	}
	_ = browser2.SetReadDeadline(time.Now().Add(5 * time.Second))
	if err := browser2.ReadJSON(&ev); err != nil || ev.Type != "tool_call" {
		t.Fatalf("expected tool_call, got %+v err=%v", ev, err)
	}
	browser2.WriteJSON(map[string]any{"type": "end"})
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if hasAction("sos.session_start") && hasAction("sos.tool_call") && hasAction("sos.session_end") {
			return
		}
		time.Sleep(30 * time.Millisecond)
	}
	t.Fatalf("audit actions incomplete: %v", actions)
}

// recordingReader 记录工具实际收到的集群名
type recordingReader struct {
	nopReader
	mu      sync.Mutex
	cluster string
}

func (r *recordingReader) ListNodes(cluster string) ([]corev1.Node, error) {
	r.mu.Lock()
	r.cluster = cluster
	r.mu.Unlock()
	return nil, nil
}

func (r *recordingReader) lastCluster() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.cluster
}

// TestSessionStartClusterOverride 浏览器 start 指令选定集群后，工具调用作用于该集群
func TestSessionStartClusterOverride(t *testing.T) {
	up := &fakeUpstream{t: t, toolCall: true, waitAudio: true}
	upSrv := httptest.NewServer(http.HandlerFunc(up.serve))
	defer upSrv.Close()
	u := strings.Replace(upSrv.URL, "http://", "ws://", 1)

	orig := Dial
	defer func() { Dial = orig }()
	Dial = func(ctx context.Context, c config.SOSConfig) (*websocket.Conn, error) {
		conn, _, err := websocket.DefaultDialer.DialContext(ctx, u, nil)
		return conn, err
	}

	rec := &recordingReader{}
	m, err := NewManager(testSOSConfig(), rec, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(http.HandlerFunc(m.HandleSessionWS))
	defer srv.Close()

	browser, _, err := websocket.DefaultDialer.Dial(strings.Replace(srv.URL, "http://", "ws://", 1), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer browser.Close()
	_ = browser.SetReadDeadline(time.Now().Add(5 * time.Second))

	// 先选集群再发音频（音频触发上游下发工具调用）
	if err := browser.WriteJSON(map[string]any{"type": "start", "cluster": "c1"}); err != nil {
		t.Fatal(err)
	}
	if err := browser.WriteMessage(websocket.BinaryMessage, []byte{1, 0}); err != nil {
		t.Fatal(err)
	}

	var ev ClientEvent
	if err := browser.ReadJSON(&ev); err != nil || ev.Type != "tool_call" {
		t.Fatalf("expected tool_call, got %+v err=%v", ev, err)
	}
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if rec.lastCluster() == "c1" {
			return
		}
		time.Sleep(30 * time.Millisecond)
	}
	t.Fatalf("tool never executed on cluster c1, got %q", rec.lastCluster())
}

// TestDispatchToolExecutionTimeout 工具执行受 toolExecTimeout 约束，超时返回错误 JSON
func TestDispatchToolExecutionTimeout(t *testing.T) {
	up := &fakeUpstream{t: t, toolCall: true, toolName: "get_pod_logs",
		toolArgs: `{"namespace":"default","pod":"p"}`}
	upSrv := httptest.NewServer(http.HandlerFunc(up.serve))
	defer upSrv.Close()
	u := strings.Replace(upSrv.URL, "http://", "ws://", 1)

	origDial := Dial
	origTimeout := toolExecTimeout
	defer func() { Dial = origDial; toolExecTimeout = origTimeout }()
	toolExecTimeout = 80 * time.Millisecond
	Dial = func(ctx context.Context, c config.SOSConfig) (*websocket.Conn, error) {
		conn, _, err := websocket.DefaultDialer.DialContext(ctx, u, nil)
		return conn, err
	}

	// 注入慢工具：GetPodLogs 阻塞 5s，远超 toolExecTimeout，验证抢占式超时
	slow := slowLogsReader{}
	m, err := NewManager(testSOSConfig(), slow, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(http.HandlerFunc(m.HandleSessionWS))
	defer srv.Close()

	browser, _, err := websocket.DefaultDialer.Dial(strings.Replace(srv.URL, "http://", "ws://", 1), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer browser.Close()
	_ = browser.SetReadDeadline(time.Now().Add(5 * time.Second))

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if msg := up.lastOf("conversation.item.create"); msg != nil {
			item, _ := msg["item"].(map[string]any)
			out, _ := item["output"].(string)
			if !strings.Contains(out, "deadline") {
				t.Fatalf("expected timeout error in output, got %s", out)
			}
			return
		}
		time.Sleep(30 * time.Millisecond)
	}
	t.Fatal("upstream never received function_call_output")
}

// slowLogsReader GetPodLogs 长时间阻塞（验证抢占式超时后 goroutine 经缓冲 channel 安全退出）
type slowLogsReader struct {
	nopReader
}

func (slowLogsReader) GetPodLogs(_, _, _ string, _ int64) (string, error) {
	time.Sleep(5 * time.Second)
	return "", nil
}

// TestAuditCallbackPanicSafe 审计回调 panic 不外泄
func TestAuditCallbackPanicSafe(t *testing.T) {
	m, err := NewManager(testSOSConfig(), nopReader{}, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	m.SetAuditLog(func(_, _ string) { panic("boom") })
	m.audit("sos.session_start", "") // 不应 panic
}

// TestSessionBridgeGLM provider=glm 时桥接全链路（事件回放与 dashscope 路径一致）
func TestSessionBridgeGLM(t *testing.T) {
	up := &fakeUpstream{t: t}
	upSrv := httptest.NewServer(http.HandlerFunc(up.serve))
	defer upSrv.Close()
	u := strings.Replace(upSrv.URL, "http://", "ws://", 1)

	orig := Dial
	defer func() { Dial = orig }()
	Dial = func(ctx context.Context, c config.SOSConfig) (*websocket.Conn, error) {
		if c.Provider != "glm" {
			t.Errorf("expected provider glm, got %q", c.Provider)
		}
		conn, _, err := websocket.DefaultDialer.DialContext(ctx, u, nil)
		return conn, err
	}

	m, err := NewManager(glmBridgeConfig(), nopReader{}, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	if !m.Status().Ready {
		t.Fatal("expected glm manager ready")
	}
	srv := httptest.NewServer(http.HandlerFunc(m.HandleSessionWS))
	defer srv.Close()

	browser, _, err := websocket.DefaultDialer.Dial(strings.Replace(srv.URL, "http://", "ws://", 1), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer browser.Close()
	_ = browser.SetReadDeadline(time.Now().Add(5 * time.Second))

	var ev ClientEvent
	if err := browser.ReadJSON(&ev); err != nil || ev.Type != "session" {
		t.Fatalf("first event = %+v err=%v", ev, err)
	}
	mt, data, err := browser.ReadMessage()
	if err != nil || mt != websocket.BinaryMessage || len(data) != 2 {
		t.Fatalf("audio frame mt=%d data=%v err=%v", mt, data, err)
	}
}

func glmBridgeConfig() config.SOSConfig {
	return config.SOSConfig{
		Enabled:  true,
		Provider: "glm",
		GLM:      config.SOSGlmConfig{APIKey: "id.secret", Model: "glm-realtime"},
	}
}
