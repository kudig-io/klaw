import { v1Api } from './api'

export interface SosStatus {
  enabled: boolean
  ready: boolean
  model: string
  voice: string
  faq_count: number
}

export async function fetchSosStatus(): Promise<SosStatus> {
  const { data } = await v1Api.get<SosStatus>('/sos/status')
  return data
}

// 浏览器 WebSocket 无法携带 Authorization 头，token 以查询参数传递
export function buildSosSessionUrl(): string {
  const proto = window.location.protocol === 'https:' ? 'wss' : 'ws'
  const token = localStorage.getItem('klaw_token') ?? ''
  const q = token ? `?token=${encodeURIComponent(token)}` : ''
  return `${proto}://${window.location.host}/api/v1/sos/session${q}`
}
