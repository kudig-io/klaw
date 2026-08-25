package sos

import (
	"context"
	"encoding/json"
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

// fakeUpstream 模拟 DashScope：回放预置事件，收到上行音频后退出
type fakeUpstream struct {
	t        *testing.T
	mu       sync.Mutex
	received []map[string]any
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

func (f *fakeUpstream) serve(w http.ResponseWriter, r *http.Request) {
	c, err := (&websocket.Upgrader{}).Upgrade(w, r, nil)
	if err != nil {
		f.t.Logf("upgrade: %v", err)
		return
	}
	defer c.Close()
	// 第一条消息应为 session.update；这里直接回放事件
	_ = c.WriteJSON(map[string]any{"type": "session.created", "session": map[string]any{"model": "m", "voice": "v"}})
	_ = c.WriteJSON(map[string]any{"type": "response.audio.delta", "delta": "AQI="}) // base64 {1,2}
	_ = c.WriteJSON(map[string]any{"type": "response.audio_transcript.delta", "delta": "你好"})
	_ = c.WriteJSON(map[string]any{"type": "input_audio_buffer.speech_started"})
	for {
		mt, data, err := c.ReadMessage()
		if err != nil {
			return
		}
		if mt == websocket.TextMessage {
			var m map[string]any
			_ = json.Unmarshal(data, &m)
			f.appendReceived(m)
			if m["type"] == "input_audio_buffer.append" {
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
	defer func() { Dial = orig }()
	Dial = func(ctx context.Context, c config.SOSDashscopeConfig) (*websocket.Conn, error) {
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
			return // 成功
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatal("upstream never received input_audio_buffer.append")
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
