package sos

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"github.com/kudig-io/klaw/internal/config"
)

const writeWait = 10 * time.Second

var (
	idleTimeout       = 5 * time.Minute  // 会话空闲超时（包级变量，便于测试注入）
	idleCheckInterval = 30 * time.Second // 空闲检查轮询间隔（与 idleTimeout 同步注入）
	retryDialDelay    = 2 * time.Second  // 上游重连拨号延迟（便于测试注入）
	toolExecTimeout   = 20 * time.Second // 单个工具执行超时（避免 k8s API 挂起导致 goroutine 无界存活）
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
	auditLog        func(action, detail string) // 可选审计回调（仅元数据，不含音频/转写）
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

// SetAuditLog 注入可选审计回调（不改变构造签名）；仅记录操作元数据，禁止记录音频/转写内容
func (m *Manager) SetAuditLog(fn func(action, detail string)) { m.auditLog = fn }

// audit 安全触发审计回调（未注入时空操作；回调 panic 不击穿会话生命周期）
func (m *Manager) audit(action, detail string) {
	if m.auditLog == nil {
		return
	}
	defer func() {
		if r := recover(); r != nil {
			log.Printf("sos: audit callback panic ignored: %v", r)
		}
	}()
	m.auditLog(action, detail)
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
		log.Printf("sos: connect dashscope failed: %v", err)
		_ = browser.WriteJSON(ClientEvent{Type: "error", Message: "语音服务连接失败，请检查 SOS 配置与网络"})
		_ = browser.Close()
		return
	}
	m.audit("sos.session_start", fmt.Sprintf("model=%s", m.cfg.Dashscope.Model))
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
	cluster    string // 浏览器 start 指令选定的集群（空为默认集群）
	reconnDone bool   // 上游断线只自动重连一次
	closed     chan struct{}
	closeOnce  sync.Once
}

func (s *session) run(ctx context.Context) {
	defer s.closeAll()
	// 下发会话配置（三层兜底：instructions 含语料 + tools）
	_ = s.upstream.WriteJSON(BuildSessionUpdate(s.m.cfg.Dashscope.Voice, s.m.instr, s.m.tools.Definitions()))

	errCh := make(chan error, 2)
	// 上游正常关闭也视为链路中断（spec 要求自动重连一次），用哨兵错误与浏览器侧的 nil 区分
	watchUpstream := func() { errCh <- s.watchUpstreamRead() }
	go func() { errCh <- s.readBrowser() }()
	go watchUpstream()

	timer := time.NewTicker(idleCheckInterval)
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
				_ = s.writeBrowser(ClientEvent{Type: "session.idle_timeout", Message: "会话长时间无操作，已自动结束"})
				return
			}
		case err := <-errCh:
			if err == nil {
				return // 浏览器侧正常关闭（end 指令）
			}
			// 上游读错误或正常关闭且尚未重连过：自动重连一次
			if (err == errUpstreamRead || err == errUpstreamClosed) && !s.reconnDone {
				s.reconnDone = true
				if err == errUpstreamRead {
					log.Printf("sos: upstream read error, reconnecting once")
				} else {
					log.Printf("sos: upstream closed, reconnecting once")
				}
				if s.reconnectUpstream(ctx) {
					go watchUpstream()
					continue
				}
			}
			_ = s.writeBrowser(ClientEvent{Type: "error", Message: "连接已断开，请刷新页面重试"})
			return
		}
	}
}

// watchUpstreamRead 包装 readUpstream：正常关闭（nil）翻译为哨兵错误，交由主循环触发重连
func (s *session) watchUpstreamRead() error {
	if err := s.readUpstream(); err == nil {
		return errUpstreamClosed
	}
	return errUpstreamRead
}

var (
	errUpstreamRead   = websocketError("upstream read error")
	errUpstreamClosed = websocketError("upstream closed normally")
)

type websocketError string

func (e websocketError) Error() string { return string(e) }

func (s *session) reconnectUpstream(ctx context.Context) bool {
	s.wmuU.Lock()
	_ = s.upstream.Close()
	s.wmuU.Unlock()
	time.Sleep(retryDialDelay)
	conn, err := Dial(ctx, s.m.cfg.Dashscope)
	if err != nil {
		return false
	}
	// 替换连接与下发 session.update 必须持写锁：readBrowser 并发读写该字段
	s.wmuU.Lock()
	s.upstream = conn
	_ = conn.WriteJSON(BuildSessionUpdate(s.m.cfg.Dashscope.Voice, s.m.instr, s.m.tools.Definitions()))
	s.wmuU.Unlock()
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
				Type    string `json:"type"`
				Cluster string `json:"cluster"`
			}
			if err := json.Unmarshal(data, &msg); err != nil {
				continue
			}
			switch msg.Type {
			case "end":
				return nil
			case "start":
				// 会话开始时选定目标集群（空为默认集群），后续工具调用生效
				s.muAudio.Lock()
				s.cluster = msg.Cluster
				s.muAudio.Unlock()
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

// readUpstream 读上游事件并翻译下发；工具调用进程内异步执行
func (s *session) readUpstream() error {
	for {
		_, data, err := s.upstream.ReadMessage()
		if err != nil {
			// 上游正常关闭：优雅结束（返回 nil），仅异常读错误才上报 errUpstreamRead
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) ||
				errors.Is(err, websocket.ErrCloseSent) {
				return nil
			}
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
			// 异步执行：避免工具耗时阻塞上行读循环（写锁已保证并发安全）
			go s.dispatchTool(call)
		}
	}
}

// dispatchTool 进程内执行集群工具并回传结果
func (s *session) dispatchTool(call FunctionCall) {
	s.m.audit("sos.tool_call", "tool="+call.Name)
	s.muAudio.Lock()
	cluster := s.cluster
	s.muAudio.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), toolExecTimeout)
	defer cancel()
	out, err := s.m.tools.ExecuteForCluster(ctx, cluster, call.Name, json.RawMessage(call.Arguments))
	if err != nil {
		// 用 json.Marshal 构造，避免手工拼接产生非法 JSON
		b, _ := json.Marshal(map[string]string{"error": err.Error()})
		out = string(b)
	}
	// 通知浏览器当前工具调用（字幕区展示）
	_ = s.writeBrowser(ClientEvent{Type: "tool_call", Name: call.Name})
	s.wmuU.Lock()
	_ = s.upstream.WriteJSON(BuildFunctionCallOutput(call.CallID, out))
	// OpenAI Realtime 兼容协议要求：工具输出回传后须发送 response.create 触发模型继续作答
	_ = s.upstream.WriteJSON(map[string]any{"type": "response.create"})
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
		s.m.audit("sos.session_end", "")
		_ = s.browser.Close()
		_ = s.upstream.Close()
	})
}
