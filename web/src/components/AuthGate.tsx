import { useState, type FormEvent, type ReactNode } from 'react'
import { Loader2, LockKeyhole, ShieldCheck } from 'lucide-react'
import { useAuth } from '../contexts/AuthContext'
import { getAuthToken } from '../lib/auth'
import { cn } from '../lib/utils'

// 认证门：未认证（或用户主动管理凭证）时遮罩应用；checking 阶段显示探测中
export default function AuthGate({ children }: { children: ReactNode }) {
  const { status, hasToken, signIn, signOut, gateOpen, closeGate } = useAuth()
  const [tokenInput, setTokenInput] = useState('')
  const [error, setError] = useState('')
  const [verifying, setVerifying] = useState(false)

  if (status === 'checking') {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50 dark:bg-gray-950">
        <div className="flex items-center gap-2 text-sm text-gray-500 dark:text-gray-400">
          <Loader2 className="h-4 w-4 animate-spin" aria-hidden="true" />
          正在验证访问凭证…
        </div>
      </div>
    )
  }

  const isManage = status === 'authenticated' && gateOpen
  const locked = status === 'required'

  if (!locked && !isManage) {
    return <>{children}</>
  }

  const handleSubmit = async (event: FormEvent) => {
    event.preventDefault()
    const token = tokenInput.trim()
    if (!token || verifying) return
    setVerifying(true)
    setError('')
    const ok = await signIn(token)
    setVerifying(false)
    if (!ok) {
      setError('Token 无效或已过期，请重试')
      setTokenInput('')
    }
  }

  const handleClear = async () => {
    setTokenInput('')
    setError('')
    closeGate()
    await signOut()
  }

  return (
    <div
      className={cn(
        'fixed inset-0 z-50 flex items-center justify-center p-4',
        isManage ? 'bg-gray-900/50 backdrop-blur-sm' : 'bg-gray-50 dark:bg-gray-950',
      )}
      role="dialog"
      aria-modal="true"
      aria-label={isManage ? '访问凭证管理' : '访问认证'}
    >
      <div className="w-full max-w-sm rounded-xl border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900 p-8 shadow-xl">
        <div className="w-10 h-10 rounded-lg bg-primary-600 flex items-center justify-center mb-5">
          {isManage ? (
            <ShieldCheck className="h-5 w-5 text-white" aria-hidden="true" />
          ) : (
            <LockKeyhole className="h-5 w-5 text-white" aria-hidden="true" />
          )}
        </div>
        <h2 className="text-lg font-semibold tracking-tight text-gray-900 dark:text-white">
          {isManage ? '访问凭证' : 'Klaw 访问控制'}
        </h2>
        <p className="mt-1.5 text-sm text-gray-500 dark:text-gray-400">
          {isManage
            ? '当前已通过 API 认证，可更换或清除本地保存的 Token。'
            : '此控制台已开启 API 认证，请输入 Access Token 以继续。'}
        </p>

        {isManage && hasToken && (
          <p className="mt-3 text-xs font-mono text-gray-400 dark:text-gray-500">
            已保存 Token：…{getAuthToken().slice(-4)}
          </p>
        )}

        <form onSubmit={handleSubmit} className="mt-5 space-y-3">
          <label htmlFor="klaw-auth-token" className="sr-only">
            Access Token
          </label>
          <input
            id="klaw-auth-token"
            type="password"
            value={tokenInput}
            onChange={(event) => setTokenInput(event.target.value)}
            placeholder="粘贴 Access Token"
            autoComplete="off"
            autoFocus
            className="w-full px-3 py-2 rounded-md border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-800 text-sm text-gray-900 dark:text-gray-100 placeholder:text-gray-400 dark:placeholder:text-gray-500 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
          />
          {error && (
            <p className="text-sm text-danger-600 dark:text-danger-500" role="alert">
              {error}
            </p>
          )}
          <button
            type="submit"
            disabled={verifying || !tokenInput.trim()}
            className="w-full flex items-center justify-center gap-2 px-4 py-2 rounded-md bg-primary-600 text-white text-sm font-medium hover:bg-primary-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors duration-150"
          >
            {verifying && <Loader2 className="h-4 w-4 animate-spin" aria-hidden="true" />}
            {isManage ? '更换 Token' : '验证并进入'}
          </button>
        </form>

        {isManage && (
          <div className="mt-3 flex items-center justify-between">
            <button
              onClick={closeGate}
              className="text-sm text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200 transition-colors duration-150"
            >
              关闭
            </button>
            {hasToken && (
              <button
                onClick={handleClear}
                className="text-sm text-danger-600 hover:text-danger-700 dark:text-danger-500 dark:hover:text-danger-400 transition-colors duration-150"
              >
                清除本地 Token
              </button>
            )}
          </div>
        )}
      </div>
    </div>
  )
}
