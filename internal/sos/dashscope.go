package sos

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"

	"github.com/kudig-io/klaw/internal/config"
)

// BuildRealtimeURL 组装 DashScope Realtime WebSocket 地址（百炼专属端点）
func BuildRealtimeURL(c config.SOSDashscopeConfig) string {
	return fmt.Sprintf("wss://%s.%s.maas.aliyuncs.com/api-ws/v1/realtime?model=%s",
		c.WorkspaceID, c.Region, c.Model)
}

// DialRealtime 建立与 DashScope 的 WebSocket 连接（Bearer 鉴权）
func DialRealtime(ctx context.Context, c config.SOSDashscopeConfig) (*websocket.Conn, error) {
	if c.WorkspaceID == "" || c.APIKey == "" {
		return nil, fmt.Errorf("sos.dashscope workspace_id/api_key not configured")
	}
	header := http.Header{}
	header.Set("Authorization", "Bearer "+c.APIKey)
	d := &websocket.Dialer{HandshakeTimeout: 10 * time.Second}
	conn, _, err := d.DialContext(ctx, BuildRealtimeURL(c), header)
	if err != nil {
		return nil, fmt.Errorf("dial dashscope: %w", err)
	}
	return conn, nil
}

// BuildSessionUpdate 构造 session.update：音色/语义打断/音频格式/instructions/tools
// 注：input_audio_transcription 用于开启用户语音转写（双向字幕）；若上游报错可移除该字段，
// 用户字幕会降级，其余功能不受影响。
func BuildSessionUpdate(voice, instructions string, tools []ToolDefinition) map[string]any {
	return map[string]any{
		"type": "session.update",
		"session": map[string]any{
			"voice":        voice,
			"instructions": instructions,
			"tools":        tools,
			"turn_detection": map[string]any{
				"type": "semantic_vad",
			},
			"audio": map[string]any{
				"input":  map[string]any{"format": "pcm", "sample_rate": 16000},
				"output": map[string]any{"format": "pcm", "sample_rate": 24000},
			},
			"input_audio_transcription": map[string]any{"model": "qwen3-asr-flash"},
		},
	}
}

// EncodeAudioAppend 将一段裸 PCM（16k 小端 Int16）编码为 input_audio_buffer.append 事件
func EncodeAudioAppend(pcm []byte) map[string]any {
	return map[string]any{
		"type":  "input_audio_buffer.append",
		"audio": base64.StdEncoding.EncodeToString(pcm),
	}
}

// BuildFunctionCallOutput 构造 conversation.item.create(function_call_output) 回传事件
func BuildFunctionCallOutput(callID, output string) map[string]any {
	return map[string]any{
		"type": "conversation.item.create",
		"item": map[string]any{
			"type":    "function_call_output",
			"call_id": callID,
			"output":  output,
		},
	}
}

// ClientEvent 发给浏览器的文本帧（小写驼峰）
type ClientEvent struct {
	Type    string `json:"type"`
	Delta   string `json:"delta,omitempty"`
	Name    string `json:"name,omitempty"`
	Message string `json:"message,omitempty"`
	Model   string `json:"model,omitempty"`
	Voice   string `json:"voice,omitempty"`
}

// FunctionCall 上游要求执行的工具调用
type FunctionCall struct {
	CallID    string
	Name      string
	Arguments string
}

func str(v any) string {
	s, _ := v.(string)
	return s
}

// TranslateUpstream 将上游事件翻译为：浏览器文本帧 / 待下发音频（裸 PCM）/ 待执行工具调用
func TranslateUpstream(raw map[string]any) (events []ClientEvent, audio []byte, calls []FunctionCall) {
	switch raw["type"] {
	case "session.created", "session.updated":
		ev := ClientEvent{Type: "session"}
		if sess, ok := raw["session"].(map[string]any); ok {
			ev.Model = str(sess["model"])
			ev.Voice = str(sess["voice"])
		}
		events = append(events, ev)
	case "input_audio_buffer.speech_started":
		// 语义打断：通知浏览器立即停播本地音频
		events = append(events, ClientEvent{Type: "speech_started"})
	case "conversation.item.input_audio_transcription.delta":
		events = append(events, ClientEvent{Type: "user.transcript.delta", Delta: str(raw["delta"])})
	case "conversation.item.input_audio_transcription.completed":
		events = append(events, ClientEvent{Type: "user.transcript.done", Delta: str(raw["transcript"])})
	case "response.audio.delta":
		if b, err := base64.StdEncoding.DecodeString(str(raw["delta"])); err == nil {
			audio = b
		}
	case "response.audio_transcript.delta":
		events = append(events, ClientEvent{Type: "assistant.transcript.delta", Delta: str(raw["delta"])})
	case "response.function_call_arguments.done":
		calls = append(calls, FunctionCall{
			CallID:    str(raw["call_id"]),
			Name:      str(raw["name"]),
			Arguments: str(raw["arguments"]),
		})
	case "response.done":
		events = append(events, ClientEvent{Type: "response.done"})
	case "error":
		msg := str(raw["message"])
		if e, ok := raw["error"].(map[string]any); ok && msg == "" {
			msg = str(e["message"])
		}
		events = append(events, ClientEvent{Type: "error", Message: msg})
	}
	return
}
