import { useCallback, useEffect, useReducer, useRef } from 'react'
import { buildSosSessionUrl } from '../lib/sosApi'
import {
  initialSosState,
  pcm16ToFloat32,
  reduceSosSession,
  type SosServerEvent,
} from '../lib/sosProtocol'
import pcmProcessorUrl from '../worklets/pcm-processor.ts?url'

// useSosSession SOS 语音会话：WS 连接、AudioWorklet 采集上行、24k 播放队列下行、语义打断停播
export function useSosSession() {
  const [state, dispatch] = useReducer(reduceSosSession, initialSosState)

  const wsRef = useRef<WebSocket | null>(null)
  const ctxRef = useRef<AudioContext | null>(null)
  const streamRef = useRef<MediaStream | null>(null)
  const nodeRef = useRef<AudioWorkletNode | null>(null)
  const sourcesRef = useRef<Set<AudioBufferSourceNode>>(new Set())
  const erroredRef = useRef(false) // 已发生错误时，onclose 不降级为 ended（保留错误状态与文案）

  const stopPlayback = useCallback(() => {
    sourcesRef.current.forEach((src) => {
      try {
        src.stop()
      } catch {
        // 已结束的 source 忽略
      }
    })
    sourcesRef.current.clear()
  }, [])

  const stopPlaybackRef = useRef(stopPlayback)
  stopPlaybackRef.current = stopPlayback

  const cleanup = useCallback(() => {
    if (nodeRef.current) {
      nodeRef.current.port.onmessage = null
      nodeRef.current.disconnect()
      nodeRef.current = null
    }
    streamRef.current?.getTracks().forEach((t) => t.stop())
    streamRef.current = null
    void ctxRef.current?.close()
    ctxRef.current = null
    wsRef.current?.close()
    wsRef.current = null
    stopPlayback()
  }, [stopPlayback])

  const start = useCallback(async () => {
    dispatch({ kind: 'reset' })
    erroredRef.current = false
    dispatch({ kind: 'status', status: 'connecting' })
    try {
      const stream = await navigator.mediaDevices.getUserMedia({
        audio: { echoCancellation: true, noiseSuppression: true, autoGainControl: true },
      })
      streamRef.current = stream
      const ctx = new AudioContext()
      ctxRef.current = ctx
      await ctx.audioWorklet.addModule(pcmProcessorUrl)
      const node = new AudioWorkletNode(ctx, 'pcm-processor')
      nodeRef.current = node
      node.port.onmessage = (e: MessageEvent<ArrayBuffer>) => {
        if (wsRef.current?.readyState === WebSocket.OPEN) {
          wsRef.current.send(e.data)
        }
      }
      ctx.createMediaStreamSource(stream).connect(node)

      const ws = new WebSocket(buildSosSessionUrl())
      wsRef.current = ws
      ws.binaryType = 'arraybuffer'
      ws.onopen = () => {
        // 会话开始帧（可选携带 cluster 选定目标集群，缺省为默认集群）
        ws.send(JSON.stringify({ type: 'start' }))
      }
      ws.onmessage = (ev) => {
        if (typeof ev.data === 'string') {
          const event = JSON.parse(ev.data) as SosServerEvent
          // 智能打断：收到用户开口信号立即停播本地音频
          if (event.type === 'speech_started') {
            stopPlaybackRef.current()
          }
          if (event.type === 'error') {
            erroredRef.current = true
          }
          dispatch({ kind: 'event', event })
          return
        }
        // 下行 24k PCM：调度播放
        const audioCtx = ctxRef.current
        if (!audioCtx) return
        const floats = pcm16ToFloat32(ev.data as ArrayBuffer)
        const buf = audioCtx.createBuffer(1, floats.length, 24000)
        buf.copyToChannel(floats as Float32Array<ArrayBuffer>, 0)
        const src = audioCtx.createBufferSource()
        src.buffer = buf
        src.connect(audioCtx.destination)
        src.onended = () => sourcesRef.current.delete(src)
        sourcesRef.current.add(src)
        src.start()
      }
      ws.onclose = () => {
        if (!erroredRef.current) {
          dispatch({ kind: 'status', status: 'ended' })
        }
      }
      ws.onerror = () => {
        erroredRef.current = true
        dispatch({ kind: 'error', error: 'WebSocket 连接失败' })
      }
    } catch (err) {
      dispatch({ kind: 'error', error: err instanceof Error ? err.message : '无法启动语音会话' })
      cleanup()
    }
  }, [cleanup])

  const hangup = useCallback(() => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify({ type: 'end' }))
    }
    cleanup()
    dispatch({ kind: 'status', status: 'ended' })
  }, [cleanup])

  const toggleMute = useCallback(() => {
    const muted = !state.muted
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify({ type: muted ? 'mute' : 'unmute' }))
    }
    dispatch({ kind: 'mute', muted })
  }, [state.muted])

  useEffect(() => cleanup, [cleanup])

  return { state, start, hangup, toggleMute }
}
