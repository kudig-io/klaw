import { describe, expect, it } from 'vitest'
import {
  float32ToPcm16,
  initialSosState,
  pcm16ToFloat32,
  reduceSosSession,
} from '../../lib/sosProtocol'

describe('sosProtocol', () => {
  it('pcm 转换往返', () => {
    const input = new Float32Array([0, 0.5, -0.5, 1, -1])
    const pcm = float32ToPcm16(input)
    const back = pcm16ToFloat32(pcm)
    expect(back[1]).toBeCloseTo(0.5, 2)
    expect(back[2]).toBeCloseTo(-0.5, 2)
  })

  it('session 事件记录 model/voice', () => {
    const s = reduceSosSession(initialSosState, {
      kind: 'event',
      event: { type: 'session', model: 'm1', voice: 'Ethan' },
    })
    expect(s.model).toBe('m1')
    expect(s.voice).toBe('Ethan')
    expect(s.status).toBe('connected')
  })

  it('字幕 delta 累积，response.done 归档', () => {
    let s = initialSosState
    s = reduceSosSession(s, { kind: 'event', event: { type: 'user.transcript.done', delta: '集群健康吗' } })
    s = reduceSosSession(s, { kind: 'event', event: { type: 'assistant.transcript.delta', delta: '让我' } })
    s = reduceSosSession(s, { kind: 'event', event: { type: 'assistant.transcript.delta', delta: '查一下' } })
    expect(s.assistantText).toBe('让我查一下')
    expect(s.messages).toHaveLength(0)
    s = reduceSosSession(s, { kind: 'event', event: { type: 'response.done' } })
    expect(s.messages).toEqual([
      { role: 'user', text: '集群健康吗' },
      { role: 'assistant', text: '让我查一下' },
    ])
    expect(s.assistantText).toBe('')
    expect(s.userText).toBe('')
  })

  it('speech_started 停止播报标记', () => {
    const speaking = reduceSosSession(initialSosState, {
      kind: 'event',
      event: { type: 'assistant.transcript.delta', delta: 'x' },
    })
    expect(speaking.speaking).toBe(true)
    const s = reduceSosSession(speaking, {
      kind: 'event',
      event: { type: 'speech_started' },
    })
    expect(s.speaking).toBe(false)
  })

  it('error 事件记录错误信息', () => {
    const s = reduceSosSession(initialSosState, {
      kind: 'event',
      event: { type: 'error', message: '连接失败' },
    })
    expect(s.status).toBe('error')
    expect(s.error).toBe('连接失败')
  })
})
