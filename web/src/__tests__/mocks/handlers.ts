// MSW handlers 入口（兼容旧 import { handlers } from '../mocks/handlers'）
// 慢速/错误 handlers 仍按旧格式内联，保持向后兼容。

import { http, HttpResponse } from 'msw'
import { mockClusters } from './data'
import { handlers as aggregatedHandlers, analysisHandlers } from './handlers/index'

export const handlers = aggregatedHandlers

// 慢速 handlers（加载状态测试用）
export const slowHandlers = [
  ...analysisHandlers,
  http.get('/api/v1/clusters', async () => {
    await new Promise((resolve) => setTimeout(resolve, 1000))
    return HttpResponse.json(mockClusters)
  }),
]

// 错误 handlers（错误处理测试用）
export const errorHandlers = [
  http.get('/api/v1/clusters', () =>
    new HttpResponse(JSON.stringify({ error: 'Internal Server Error' }), { status: 500 })),
  http.get('/api/v1/clusters/:name/status', () =>
    new HttpResponse(JSON.stringify({ error: 'Cluster not reachable' }), { status: 503 })),
]