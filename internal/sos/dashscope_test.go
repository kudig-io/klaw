package sos

import (
	"encoding/base64"
	"testing"

	"github.com/kudig-io/klaw/internal/config"
)

func TestBuildRealtimeURL(t *testing.T) {
	got := BuildRealtimeURL(config.SOSDashscopeConfig{
		WorkspaceID: "ws123", Region: "cn-beijing", Model: "m1",
	})
	want := "wss://ws123.cn-beijing.maas.aliyuncs.com/api-ws/v1/realtime?model=m1"
	if got != want {
		t.Fatalf("got %s want %s", got, want)
	}
}

func TestBuildSessionUpdate(t *testing.T) {
	ev := BuildSessionUpdate("Ethan", "instr", []ToolDefinition{{Type: "function", Name: "t1"}})
	if ev["type"] != "session.update" {
		t.Fatalf("type = %v", ev["type"])
	}
	sess := ev["session"].(map[string]any)
	if sess["voice"] != "Ethan" || sess["instructions"] != "instr" {
		t.Fatalf("session = %v", sess)
	}
	if sess["turn_detection"].(map[string]any)["type"] != "semantic_vad" {
		t.Fatal("expected semantic_vad")
	}
	audio := sess["audio"].(map[string]any)
	in := audio["input"].(map[string]any)
	out := audio["output"].(map[string]any)
	if in["format"] != "pcm" || in["sample_rate"] != 16000 || out["sample_rate"] != 24000 {
		t.Fatalf("audio formats wrong: %v", audio)
	}
}

func TestEncodeAudioAppend(t *testing.T) {
	pcm := []byte{1, 2, 3}
	ev := EncodeAudioAppend(pcm)
	if ev["type"] != "input_audio_buffer.append" {
		t.Fatal("wrong event type")
	}
	if ev["audio"].(string) != base64.StdEncoding.EncodeToString(pcm) {
		t.Fatal("wrong base64 payload")
	}
}

func TestTranslateUpstream(t *testing.T) {
	// session.created -> session 帧
	evs, _, _ := TranslateUpstream(map[string]any{
		"type": "session.created", "session": map[string]any{"model": "m", "voice": "v"},
	})
	if len(evs) != 1 || evs[0].Type != "session" || evs[0].Model != "m" {
		t.Fatalf("session.created -> %+v", evs)
	}
	// speech_started -> 打断帧
	evs, _, _ = TranslateUpstream(map[string]any{"type": "input_audio_buffer.speech_started"})
	if len(evs) != 1 || evs[0].Type != "speech_started" {
		t.Fatalf("speech_started -> %+v", evs)
	}
	// 音频 delta -> 二进制
	b64 := base64.StdEncoding.EncodeToString([]byte{9, 9})
	evs, audio, _ := TranslateUpstream(map[string]any{"type": "response.audio.delta", "delta": b64})
	if len(evs) != 0 || len(audio) != 2 || audio[0] != 9 {
		t.Fatalf("audio.delta -> evs=%+v audio=%v", evs, audio)
	}
	// 助理字幕 delta
	evs, _, _ = TranslateUpstream(map[string]any{"type": "response.audio_transcript.delta", "delta": "你好"})
	if evs[0].Type != "assistant.transcript.delta" || evs[0].Delta != "你好" {
		t.Fatalf("transcript.delta -> %+v", evs)
	}
	// 用户字幕 delta / done
	evs, _, _ = TranslateUpstream(map[string]any{"type": "conversation.item.input_audio_transcription.delta", "delta": "问"})
	if evs[0].Type != "user.transcript.delta" {
		t.Fatalf("user delta -> %+v", evs)
	}
	// 工具调用
	_, _, calls := TranslateUpstream(map[string]any{
		"type":    "response.function_call_arguments.done",
		"call_id": "c1", "name": "list_pods", "arguments": `{"namespace":"default"}`,
	})
	if len(calls) != 1 || calls[0].Name != "list_pods" || calls[0].CallID != "c1" {
		t.Fatalf("function call -> %+v", calls)
	}
	// 错误事件
	evs, _, _ = TranslateUpstream(map[string]any{"type": "error", "error": map[string]any{"message": "bad"}})
	if evs[0].Type != "error" || evs[0].Message != "bad" {
		t.Fatalf("error -> %+v", evs)
	}
	// response.done -> 回合结束
	evs, _, _ = TranslateUpstream(map[string]any{"type": "response.done"})
	if evs[0].Type != "response.done" {
		t.Fatalf("response.done -> %+v", evs)
	}
	// 未知事件静默忽略
	evs, audio, calls = TranslateUpstream(map[string]any{"type": "rate_limits.updated"})
	if len(evs) != 0 || audio != nil || calls != nil {
		t.Fatal("unknown events should be dropped")
	}
}

func TestBuildFunctionCallOutput(t *testing.T) {
	ev := BuildFunctionCallOutput("c1", `{"ok":true}`)
	if ev["type"] != "conversation.item.create" {
		t.Fatal("wrong type")
	}
	item := ev["item"].(map[string]any)
	if item["call_id"] != "c1" || item["output"] != `{"ok":true}` {
		t.Fatalf("item = %v", item)
	}
}
