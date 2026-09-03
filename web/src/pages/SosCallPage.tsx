import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { ArrowLeft, Loader2, Mic, MicOff, PhoneOff, Siren } from 'lucide-react'
import { fetchSosStatus, type SosStatus } from '../lib/sosApi'
import { useSosSession } from '../hooks/useSosSession'
import { cn } from '../lib/utils'

const statusText: Record<string, string> = {
  idle: '准备就绪',
  connecting: '连接中…',
  connected: '通话中',
  error: '出错了',
  ended: '已结束',
}

// SosCallPage SOS 全屏语音通话页（类电话聊天）：头像动画 + 双向字幕 + 静音/挂断控制
export default function SosCallPage() {
  const [statusInfo, setStatusInfo] = useState<SosStatus | null>(null)
  const { state, start, hangup, toggleMute } = useSosSession()

  useEffect(() => {
    fetchSosStatus().then(setStatusInfo).catch(() => setStatusInfo(null))
  }, [])

  useEffect(() => {
    if (statusInfo?.ready && state.status === 'idle') {
      void start()
    }
  }, [statusInfo, state.status, start])

  return (
    <div className="flex min-h-screen flex-col bg-gray-950 text-white">
      {/* 顶栏 */}
      <div className="flex items-center justify-between p-4">
        <Link to="/" className="rounded-md p-2 hover:bg-white/10" aria-label="返回">
          <ArrowLeft className="h-5 w-5" />
        </Link>
        <div className="flex items-center gap-2 text-sm text-gray-300">
          <Siren className="h-4 w-4 text-red-500" />
          SOS 应急语音助手
        </div>
        <div className="w-9" />
      </div>

      {/* 主体 */}
      <div className="flex flex-1 flex-col items-center justify-center gap-8 px-6">
        {statusInfo && !statusInfo.ready ? (
          <div className="max-w-md space-y-3 rounded-lg bg-white/5 p-6 text-center text-sm leading-6 text-gray-300">
            <p className="text-base font-medium text-white">SOS 未就绪，请检查配置：</p>
            <p>
              在 <code className="rounded bg-black/40 px-1">config.yaml</code> 开启
              <code className="mx-1 rounded bg-black/40 px-1">sos.enabled</code> 并配置
              <code className="mx-1 rounded bg-black/40 px-1">sos.dashscope.workspace_id</code>；
              API Key 建议用环境变量
              <code className="mx-1 rounded bg-black/40 px-1">KLAW_SOS_DASHSCOPE_API_KEY</code>
              注入后重启服务。
            </p>
          </div>
        ) : (
          <>
            {/* 头像 + 呼吸动画 */}
            <div
              className={cn(
                'flex h-32 w-32 items-center justify-center rounded-full ring-1 ring-white/10',
                'bg-red-500',
                state.status === 'connected' && 'animate-pulse',
              )}
            >
              {state.status === 'connecting' ? (
                <Loader2 className="h-12 w-12 animate-spin" />
              ) : (
                <Siren className="h-12 w-12" />
              )}
            </div>
            <div className="text-sm text-gray-400">
              {statusText[state.status] ?? state.status}
              {state.error ? ` · ${state.error}` : ''}
            </div>

            {/* 工具调用状态：应急场景下给用户“正在查集群”的实时反馈 */}
            {state.toolCall && (
              <div className="flex items-center gap-2 rounded-full bg-white/5 px-4 py-1.5 text-xs text-gray-300">
                <Loader2 className="h-3 w-3 animate-spin text-red-400" />
                正在查询集群：{state.toolCall}
              </div>
            )}

            {/* 字幕区 */}
            <div className="w-full max-w-xl space-y-2 overflow-y-auto">
              {(state.messages ?? []).map((m, i) => (
                <p
                  key={i}
                  className={cn(
                    'text-sm',
                    m.role === 'user' ? 'text-right text-red-300' : 'text-left text-gray-200',
                  )}
                >
                  {m.text}
                </p>
              ))}
              {state.userText && <p className="text-right text-sm text-red-300">{state.userText}</p>}
              {state.assistantText && (
                <p className="text-left text-sm text-gray-200">{state.assistantText}</p>
              )}
            </div>
          </>
        )}
      </div>

      {/* 控制条 */}
      {statusInfo?.ready && (
        <div className="flex items-center justify-center gap-8 pb-10">
          <button
            onClick={toggleMute}
            aria-label={state.muted ? '取消静音' : '静音'}
            className={cn(
              'flex h-14 w-14 items-center justify-center rounded-full transition-colors',
              state.muted ? 'bg-gray-600 text-white' : 'bg-white/10 text-white hover:bg-white/20',
            )}
          >
            {state.muted ? <MicOff className="h-6 w-6" /> : <Mic className="h-6 w-6" />}
          </button>
          <button
            onClick={hangup}
            aria-label="挂断"
            className="flex h-16 w-16 items-center justify-center rounded-full bg-red-600 transition-colors hover:bg-red-700 active:bg-red-800"
          >
            <PhoneOff className="h-7 w-7" />
          </button>
        </div>
      )}
    </div>
  )
}
