// Pod 日志与日志分析

import { http, HttpResponse } from 'msw'
import { getLogsForPod, analyzeLogs, mockRbacAnalysis } from '../data/index'
import { store } from '../store'

export const analysisHandlers = [
  http.get('/api/v1/clusters/:cluster/namespaces/:namespace/pods/:name/logs', ({ params }) => {
    // 优先 store.pods，否则按名字匹配
    const pod = store.pods.find((p) => p.metadata.name === params.name && p.metadata.namespace === params.namespace)
    const podName = pod ? pod.metadata.name : (params.name as string)
    return HttpResponse.json({ logs: getLogsForPod(podName) })
  }),
  http.get('/api/v1/clusters/:cluster/namespaces/:namespace/pods/:name/logs/analysis', ({ params }) => {
    const pod = store.pods.find((p) => p.metadata.name === params.name && p.metadata.namespace === params.namespace)
    const podName = pod ? pod.metadata.name : (params.name as string)
    const logs = getLogsForPod(podName)
    return HttpResponse.json(analyzeLogs(logs))
  }),
  http.post('/api/v1/analysis/logs', async ({ request }) => {
    const body = await request.json() as { logs: string }
    return HttpResponse.json(analyzeLogs(body.logs || ''))
  }),
  http.get('/api/v1/clusters/:cluster/rbac/analysis', ({ params }) => {
    const k = params.cluster === 'production' ? 'production' : 'kindtest'
    return HttpResponse.json(mockRbacAnalysis[k])
  }),
]