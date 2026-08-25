package sos

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"github.com/kudig-io/klaw/internal/config"
)

const (
	idleTimeout    = 5 * time.Minute // 会话空闲超时
	writeWait      = 10 * time.Second
	retryDialDelay = 2 * time.Second
)

// Dial 可替换的上游建连函数（测试注入假 DashScope）
var Dial = func(ctx context.Context, c config.SOSDashscopeConfig) (*websocket.Conn, error) {
	return DialRealtime(ctx, c)
}

// StatusResponse GET /api/v1/sos/status 响应
type StatusResponse struct {
	Enabled  bool   `json:"enabled"`
	Ready    bool   `json:"ready"`
	Model    string `json:"model"`
	Voice    string `json:"voice"`
	FAQCount int    `json:"faq_count"`
}

// Manager SOS 会话管理器：持有配置、语料 instructions 与工具执行器
type Manager struct {
	cfg             config.SOSConfig
	instr           string
	faqs            []FAQEntry
	tools           *ToolExecutor
	allowedOrigins  []string
	browserUpgrader websocket.Upgrader
}

func NewManager(cfg config.SOSConfig, reader ClusterReader, clusterName string, allowedOrigins []string) (*Manager, error) {
	faqs, err := LoadFAQs(cfg.FAQFile)
	if err != nil {
		return nil, err
	}
	m := &Manager{
		cfg:            cfg,
		instr:          BuildInstructions(cfg.InstructionsPrefix, faqs),
		faqs:           faqs,
		tools:          NewToolExecutor(reader, clusterName),
		allowedOrigins: allowedOrigins,
	}
	m.browserUpgrader = websocket.Upgrader{CheckOrigin: m.checkWSOrigin}
	return m, nil
}

// checkWSOrigin 仅放行同源与 CORS 白名单内的 Origin（浏览器 WS 无法携带自定义鉴权头，
// Origin 校验是跨站防护的第一道关口）
func (m *Manager) checkWSOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true // 非浏览器客户端
	}
	if u, err := url.Parse(origin); err == nil && r.Host != "" && u.Host == r.Host {
		return true // 同源
	}
	for _, o := range m.allowedOrigins {
		if o == "*" || o == origin {
			return true
		}
	}
	return false
}

// Status 返回 SOS 可用性：enabled 且 DashScope 配置完整才 ready
func (m *Manager) Status() StatusResponse {
	ready := m.cfg.Enabled && m.cfg.Dashscope.APIKey != "" && m.cfg.Dashscope.WorkspaceID != ""
	return StatusResponse{
		Enabled: m.cfg.Enabled, Ready: ready,
		Model: m.cfg.Dashscope.Model, Voice: m.cfg.Dashscope.Voice,
		FAQCount: len(m.faqs),
	}
}

// HandleSessionWS 升级浏览器连接并启动桥接会话
func (m *Manager) HandleSessionWS(w http.ResponseWriter, r *http.Request) {
	if !m.cfg.Enabled || m.cfg.Dashscope.APIKey == "" || m.cfg.Dashscope.WorkspaceID == "" {
		http.Error(w, `{"error":"sos not configured"}`, http.StatusServiceUnavailable)
		return
	}
	browser, err := m.browserUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	upstream, err := Dial(r.Context(), m.cfg.Dashscope)
	if err != nil {
		_ = browser.WriteJSON(ClientEvent{Type: "error", Message: "connect dashscope failed: " + err.Error()})
		_ = browser.Close()
		return
	}
	s := &session{m: m, browser: browser, upstream: upstream,
		lastAudio: time.Now(), closed: make(chan struct{})}
	s.run(r.Context())
}

// session 单次语音会话：浏览器 WS <-> DashScope WS 双向桥接
type session struct {
	m        *Manager
	browser  *websocket.Conn
	upstream *websocket.Conn

	wmuB, wmuU sync.Mutex // 两把连接各自的写锁
	muAudio    sync.Mutex
	lastAudio  time.Time
	muted      bool
	reconnDone bool // 上游断线只自动重连一次
	closed     chan struct{}
	closeOnce  sync.Once
}

func (s *session) run(ctx context.Context) {
	defer s.closeAll()
	// 下发会话配置（三层兜底：instructions 含语料 + tools）
	_ = s.upstream.WriteJSON(BuildSessionUpdate(s.m.cfg.Dashscope.Voice, s.m.instr, s.m.tools.Definitions()))

	errCh := make(chan error, 2)
	go func() { errCh <- s.readBrowser() }()
	go func() { errCh <- s.readUpstream() }()

	timer := time.NewTicker(30 * time.Second)
	defer timer.Stop()
	for {
		select {
		case <-s.closed:
			return
		case <-ctx.Done():
			return
		case <-timer.C:
			s.muAudio.Lock()
			idle := time.Since(s.lastAudio) > idleTimeout
			s.muAudio.Unlock()
			if idle {
				_ = s.writeBrowser(ClientEvent{Type: "error", Message: "session idle timeout"})
				return
			}
		case err := <-errCh:
			if err == nil {
				return // 对端正常关闭
			}
			// 上游读错误且尚未重连过：自动重连一次
			if err == errUpstreamRead && !s.reconnDone {
				s.reconnDone = true
				log.Printf("sos: upstream read error: %v, reconnecting once", err)
				if s.reconnectUpstream(ctx) {
					go func() { errCh <- s.readUpstream() }()
					continue
				}
			}
			_ = s.writeBrowser(ClientEvent{Type: "error", Message: "connection lost"})
			return
		}
	}
}

var errUpstreamRead = websocketError("upstream read error")

type websocketError string

func (e websocketError) Error() string { return string(e) }

func (s *session) reconnectUpstream(ctx context.Context) bool {
	_ = s.upstream.Close()
	time.Sleep(retryDialDelay)
	conn, err := Dial(ctx, s.m.cfg.Dashscope)
	if err != nil {
		return false
	}
	s.upstream = conn
	_ = conn.WriteJSON(BuildSessionUpdate(s.m.cfg.Dashscope.Voice, s.m.instr, s.m.tools.Definitions()))
	return true
}

func (s *session) touch() {
	s.muAudio.Lock()
	s.lastAudio = time.Now()
	s.muAudio.Unlock()
}

// readBrowser 读浏览器帧：文本=控制指令，二进制=上行 PCM
func (s *session) readBrowser() error {
	for {
		mt, data, err := s.browser.ReadMessage()
		if err != nil {
			return err
		}
		switch mt {
		case websocket.TextMessage:
			var msg struct {
				Type string `json:"type"`
			}
			if err := json.Unmarshal(data, &msg); err != nil {
				continue
			}
			switch msg.Type {
			case "end":
				return nil
			case "mute":
				s.muAudio.Lock()
				s.muted = true
				s.muAudio.Unlock()
			case "unmute":
				s.muAudio.Lock()
				s.muted = false
				s.muAudio.Unlock()
			}
		case websocket.BinaryMessage:
			s.touch()
			s.muAudio.Lock()
			muted := s.muted
			s.muAudio.Unlock()
			if muted {
				continue
			}
			s.wmuU.Lock()
			_ = s.upstream.WriteJSON(EncodeAudioAppend(data))
			s.wmuU.Unlock()
		}
	}
}

// readUpstream 读上游事件并翻译下发；工具调用进程内执行
func (s *session) readUpstream() error {
	for {
		_, data, err := s.upstream.ReadMessage()
		if err != nil {
			return errUpstreamRead
		}
		var raw map[string]any
		if err := json.Unmarshal(data, &raw); err != nil {
			continue
		}
		events, audio, calls := TranslateUpstream(raw)
		for _, ev := range events {
			_ = s.writeBrowser(ev)
		}
		if len(audio) > 0 {
			s.wmuB.Lock()
			_ = s.browser.SetWriteDeadline(time.Now().Add(writeWait))
			_ = s.browser.WriteMessage(websocket.BinaryMessage, audio)
			s.wmuB.Unlock()
		}
		for _, call := range calls {
			s.dispatchTool(call)
		}
	}
}

// dispatchTool 进程内执行集群工具并回传结果
func (s *session) dispatchTool(call FunctionCall) {
	out, err := s.m.tools.Execute(context.Background(), call.Name, json.RawMessage(call.Arguments))
	if err != nil {
		out = `{"error":"` + err.Error() + `"}`
	}
	// 通知浏览器当前工具调用（字幕区展示）
	_ = s.writeBrowser(ClientEvent{Type: "tool_call", Name: call.Name})
	s.wmuU.Lock()
	_ = s.upstream.WriteJSON(BuildFunctionCallOutput(call.CallID, out))
	s.wmuU.Unlock()
}

func (s *session) writeBrowser(ev ClientEvent) error {
	s.wmuB.Lock()
	defer s.wmuB.Unlock()
	_ = s.browser.SetWriteDeadline(time.Now().Add(writeWait))
	return s.browser.WriteJSON(ev)
}

func (s *session) closeAll() {
	s.closeOnce.Do(func() {
		close(s.closed)
		_ = s.browser.Close()
		_ = s.upstream.Close()
	})
}
