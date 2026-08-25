// SOS 会话纯逻辑：状态机 reducer 与 PCM 转换（无副作用，便于单测）

export type SosConnStatus = 'idle' | 'connecting' | 'connected' | 'error' | 'ended'

export interface TranscriptEntry {
  role: 'user' | 'assistant'
  text: string
}

export interface SosSessionState {
  status: SosConnStatus
  error: string
  model: string
  voice: string
  userText: string
  assistantText: string
  muted: boolean
  speaking: boolean
  messages: TranscriptEntry[]
}

export const initialSosState: SosSessionState = {
  status: 'idle',
  error: '',
  model: '',
  voice: '',
  userText: '',
  assistantText: '',
  muted: false,
  speaking: false,
  messages: [],
}

export type SosServerEvent =
  | { type: 'session'; model?: string; voice?: string }
  | { type: 'error'; message?: string }
  | { type: 'speech_started' }
  | { type: 'response.done' }
  | { type: 'tool_call'; name?: string }
  | { type: 'user.transcript.delta'; delta?: string }
  | { type: 'user.transcript.done'; delta?: string }
  | { type: 'assistant.transcript.delta'; delta?: string }

export type SosAction =
  | { kind: 'status'; status: SosConnStatus }
  | { kind: 'event'; event: SosServerEvent }
  | { kind: 'mute'; muted: boolean }
  | { kind: 'error'; error: string }
  | { kind: 'reset' }

export function reduceSosSession(state: SosSessionState, action: SosAction): SosSessionState {
  switch (action.kind) {
    case 'status':
      return { ...state, status: action.status }
    case 'mute':
      return { ...state, muted: action.muted }
    case 'error':
      return { ...state, status: 'error', error: action.error }
    case 'reset':
      return { ...initialSosState }
    case 'event':
      return reduceEvent(state, action.event)
  }
}

function reduceEvent(state: SosSessionState, ev: SosServerEvent): SosSessionState {
  switch (ev.type) {
    case 'session':
      return {
        ...state,
        status: 'connected',
        model: ev.model ?? state.model,
        voice: ev.voice ?? state.voice,
      }
    case 'error':
      return { ...state, status: 'error', error: ev.message ?? '未知错误' }
    case 'speech_started':
      // 智能打断：用户开口，停止本地播报
      return { ...state, speaking: false }
    case 'user.transcript.delta':
      return { ...state, userText: state.userText + (ev.delta ?? '') }
    case 'user.transcript.done':
      return { ...state, userText: ev.delta ?? state.userText }
    case 'assistant.transcript.delta':
      return { ...state, assistantText: state.assistantText + (ev.delta ?? ''), speaking: true }
    case 'response.done': {
      const messages = [...state.messages]
      if (state.userText) messages.push({ role: 'user', text: state.userText })
      if (state.assistantText) messages.push({ role: 'assistant', text: state.assistantText })
      return { ...state, messages, userText: '', assistantText: '', speaking: false }
    }
    default:
      return state
  }
}

// Float32 [-1,1] -> 16k 小端 Int16 PCM
export function float32ToPcm16(input: Float32Array): ArrayBuffer {
  const buf = new ArrayBuffer(input.length * 2)
  const view = new DataView(buf)
  for (let i = 0; i < input.length; i++) {
    const s = Math.max(-1, Math.min(1, input[i]))
    view.setInt16(i * 2, s < 0 ? s * 0x8000 : s * 0x7fff, true)
  }
  return buf
}

// 24k 小端 Int16 PCM -> Float32 [-1,1]
export function pcm16ToFloat32(buffer: ArrayBuffer): Float32Array {
  const view = new DataView(buffer)
  const out = new Float32Array(view.byteLength / 2)
  for (let i = 0; i < out.length; i++) {
    out[i] = view.getInt16(i * 2, true) / 0x8000
  }
  return out
}
