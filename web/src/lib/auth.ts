const TOKEN_KEY = 'klaw_token'

export const UNAUTHORIZED_EVENT = 'klaw:unauthorized'

export function getAuthToken(): string {
  return localStorage.getItem(TOKEN_KEY) ?? ''
}

export function setAuthToken(token: string): void {
  localStorage.setItem(TOKEN_KEY, token)
}

export function clearAuthToken(): void {
  localStorage.removeItem(TOKEN_KEY)
}

// 401 时由 axios 拦截器广播，AuthContext 监听后弹出认证门
export function notifyUnauthorized(): void {
  window.dispatchEvent(new CustomEvent(UNAUTHORIZED_EVENT))
}

export function onUnauthorized(callback: () => void): () => void {
  const handler = () => callback()
  window.addEventListener(UNAUTHORIZED_EVENT, handler)
  return () => window.removeEventListener(UNAUTHORIZED_EVENT, handler)
}

// 与 main.tsx / App.tsx 的 Mock 判定保持一致：环境变量或 localStorage 开关
export function isMockMode(): boolean {
  return import.meta.env.VITE_USE_MOCK === 'true' || localStorage.getItem('USE_MOCK') === 'true'
}
