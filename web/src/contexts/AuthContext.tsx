import { createContext, useCallback, useContext, useEffect, useState } from 'react'
import type { ReactNode } from 'react'
import { v1Api } from '../lib/api'
import { clearAuthToken, getAuthToken, isMockMode, onUnauthorized, setAuthToken } from '../lib/auth'

export type AuthStatus = 'checking' | 'authenticated' | 'required'

interface AuthContextType {
  status: AuthStatus
  /** 后端是否开启了 API 认证（verify 响应回报；Mock 模式下恒为 false） */
  authEnabled: boolean
  /** 本地是否存有 Token */
  hasToken: boolean
  /** 保存 Token 并验证；验证失败会清除 Token 并返回 false */
  signIn: (token: string) => Promise<boolean>
  /** 清除本地 Token 并重新探测认证状态 */
  signOut: () => Promise<void>
  /** 打开凭证管理弹窗（已认证状态下） */
  openGate: () => void
  closeGate: () => void
  gateOpen: boolean
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

interface VerifyResponse {
  authenticated: boolean
  authEnabled: boolean
}

async function requestVerify(): Promise<{ ok: boolean; authEnabled: boolean }> {
  try {
    const { data } = await v1Api.get<VerifyResponse>('/auth/verify')
    return { ok: true, authEnabled: Boolean(data.authEnabled) }
  } catch {
    return { ok: false, authEnabled: true }
  }
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [status, setStatus] = useState<AuthStatus>('checking')
  const [authEnabled, setAuthEnabled] = useState(true)
  const [hasToken, setHasToken] = useState(() => getAuthToken() !== '')
  const [gateOpen, setGateOpen] = useState(false)

  const verify = useCallback(async () => {
    const { ok, authEnabled: enabled } = await requestVerify()
    setAuthEnabled(enabled)
    setStatus(ok ? 'authenticated' : 'required')
  }, [])

  useEffect(() => {
    // Mock 模式走 MSW，不校验真实后端凭证
    if (isMockMode()) {
      setAuthEnabled(false)
      setStatus('authenticated')
      return
    }
    void verify()
    // 会话中途 Token 失效（后端轮换/超时）时重新弹出认证门
    return onUnauthorized(() => {
      setStatus('required')
      setGateOpen(false)
    })
  }, [verify])

  const signIn = useCallback(async (token: string) => {
    setAuthToken(token)
    setHasToken(true)
    const { ok, authEnabled: enabled } = await requestVerify()
    if (ok) {
      setAuthEnabled(enabled)
      setStatus('authenticated')
      setGateOpen(false)
      return true
    }
    clearAuthToken()
    setHasToken(false)
    return false
  }, [])

  const signOut = useCallback(async () => {
    clearAuthToken()
    setHasToken(false)
    await verify()
  }, [verify])

  const openGate = useCallback(() => setGateOpen(true), [])
  const closeGate = useCallback(() => setGateOpen(false), [])

  return (
    <AuthContext.Provider
      value={{ status, authEnabled, hasToken, signIn, signOut, openGate, closeGate, gateOpen }}
    >
      {children}
    </AuthContext.Provider>
  )
}

// Context + useXxx 同文件导出是本仓库约定（与 ToastContext 一致），HMR 边界代价可接受
// eslint-disable-next-line react-refresh/only-export-components
export function useAuth(): AuthContextType {
  const context = useContext(AuthContext)
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider')
  }
  return context
}
