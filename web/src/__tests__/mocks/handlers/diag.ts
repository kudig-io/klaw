// 集群诊断与 SOS handlers

import { http, HttpResponse } from 'msw'
import { buildDiagResponse, mockSosStatus, mockSosFallback, mockNodes } from '../data/index'
import { store } from '../store'

export const diagHandlers = [
  http.get('/api/v1/diag/run', ({ request }) => {
    const url = new URL(request.url)
    const node = url.searchParams.get('node') || undefined
    // 先以 base mock issue 集为底；再基于 store 当前状态追加动态 issue
    const base = buildDiagResponse(node)
    const dynamic: any[] = []
    // 如果 worker2 MemoryPressure 真实在 nodes 里反映，加入对应 issue
    const worker2 = mockNodes.find((n) => n.metadata.name === 'kind-test-worker2')
    if (worker2?.status.conditions.find((c: any) => c.type === 'MemoryPressure' && c.status === 'True')) {
      // 已包含
    }
    // 如果 mall-frontend 有 CrashLoop pod，加入 issue
    const crashLoop = store.pods.find((p) => p.status.phase === 'CrashLoopBackOff' && p.metadata.namespace === 'mall-prod')
    if (crashLoop) {
      dynamic.push({
        severity: 'warning', cn_name: '容器 CrashLoopBackOff', en_name: 'Container CrashLoop',
        analyzer_name: 'pod-restart', location: `${crashLoop.metadata.namespace}/${crashLoop.metadata.name}`,
        details: `Pod ${crashLoop.metadata.name} 当前处于 CrashLoopBackOff 状态。`,
        remediation: { suggestion: 'kubectl logs 查看上一次容器退出原因。' },
      })
    }
    return HttpResponse.json({
      ...base,
      issues: [...base.issues, ...dynamic],
    })
  }),
]

export const sosHandlers = [
  http.get('/api/v1/sos/status', () => HttpResponse.json(mockSosStatus)),
  http.get('/api/v1/sos/status-fallback', () => HttpResponse.json(mockSosFallback)),
]