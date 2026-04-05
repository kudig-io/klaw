// MSW (Mock Service Worker) handlers - API mock

import { http, HttpResponse } from 'msw'
import {
  mockClusters,
  mockClusterStatus,
  mockNamespaces,
  mockPods,
  mockNodes,
  mockDeployments,
  mockEvents,
  mockNodeMetrics,
  mockMetricsHistory,
  mockAlerts,
} from './data'

export const handlers = [
  // Cluster APIs
  http.get('/api/clusters', () => {
    return HttpResponse.json(mockClusters)
  }),

  http.get('/api/clusters/:name', ({ params }) => {
    const cluster = mockClusters.find(c => c.name === params.name)
    if (!cluster) {
      return new HttpResponse(null, { status: 404 })
    }
    return HttpResponse.json(cluster)
  }),

  http.get('/api/clusters/:name/status', ({ params }) => {
    return HttpResponse.json({
      ...mockClusterStatus,
      cluster: params.name,
    })
  }),

  http.get('/api/clusters/:name/metrics', ({ params }) => {
    return HttpResponse.json({
      clusterName: params.name,
      timestamp: new Date().toISOString(),
      nodes: { total: 3, ready: 3, notReady: 0 },
      pods: { total: 12, running: 10, pending: 2, failed: 0 },
      resources: {
        totalCPU: '12',
        totalMemory: '24Gi',
        usedCPU: '4',
        usedMemory: '8Gi',
      },
    })
  }),

  http.get('/api/clusters/:name/namespaces', () => {
    return HttpResponse.json(mockNamespaces)
  }),

  // Pod APIs
  http.get('/api/clusters/:cluster/pods', ({ request }) => {
    const url = new URL(request.url)
    const namespace = url.searchParams.get('namespace')
    if (namespace) {
      return HttpResponse.json(mockPods.filter(p => p.metadata.namespace === namespace))
    }
    return HttpResponse.json(mockPods)
  }),

  http.get('/api/clusters/:cluster/namespaces/:namespace/pods', ({ params }) => {
    const pods = mockPods.filter(p => p.metadata.namespace === params.namespace)
    return HttpResponse.json(pods)
  }),

  http.get('/api/clusters/:cluster/namespaces/:namespace/pods/:name', ({ params }) => {
    const pod = mockPods.find(p => 
      p.metadata.name === params.name && 
      p.metadata.namespace === params.namespace
    )
    if (!pod) {
      return new HttpResponse(null, { status: 404 })
    }
    return HttpResponse.json(pod)
  }),

  http.get('/api/clusters/:cluster/namespaces/:namespace/pods/:name/logs', () => {
    return HttpResponse.json({
      logs: '[mock] 127.0.0.1 - - [01/Apr/2026:12:00:00 +0000] "GET / HTTP/1.1" 200 612\n[mock] 127.0.0.1 - - [01/Apr/2026:12:00:01 +0000] "GET /api HTTP/1.1" 200 1024',
    })
  }),

  http.delete('/api/clusters/:cluster/namespaces/:namespace/pods/:name', ({ params }) => {
    return HttpResponse.json({
      message: `Pod ${params.name} deleted successfully`,
    })
  }),

  // Node APIs
  http.get('/api/clusters/:cluster/nodes', () => {
    return HttpResponse.json(mockNodes)
  }),

  http.get('/api/clusters/:cluster/nodes/:name', ({ params }) => {
    const node = mockNodes.find(n => n.metadata.name === params.name)
    if (!node) {
      return new HttpResponse(null, { status: 404 })
    }
    return HttpResponse.json(node)
  }),

  http.get('/api/clusters/:cluster/nodes/metrics', () => {
    return HttpResponse.json(mockNodeMetrics)
  }),

  // Deployment APIs
  http.get('/api/clusters/:cluster/namespaces/:namespace/deployments', ({ params }) => {
    const deployments = mockDeployments.filter(d => d.metadata.namespace === params.namespace)
    return HttpResponse.json(deployments)
  }),

  http.get('/api/clusters/:cluster/namespaces/:namespace/deployments/:name', ({ params }) => {
    const deployment = mockDeployments.find(d => 
      d.metadata.name === params.name && 
      d.metadata.namespace === params.namespace
    )
    if (!deployment) {
      return new HttpResponse(null, { status: 404 })
    }
    return HttpResponse.json(deployment)
  }),

  http.get('/api/clusters/:cluster/namespaces/:namespace/deployments/:name/status', ({ params }) => {
    const deployment = mockDeployments.find(d => 
      d.metadata.name === params.name && 
      d.metadata.namespace === params.namespace
    )
    if (!deployment) {
      return new HttpResponse(null, { status: 404 })
    }
    return HttpResponse.json({
      name: deployment.metadata.name,
      namespace: deployment.metadata.namespace,
      replicas: deployment.status.replicas,
      availableReplicas: deployment.status.availableReplicas,
      readyReplicas: deployment.status.readyReplicas,
      updatedReplicas: deployment.status.updatedReplicas,
      conditions: deployment.status.conditions,
    })
  }),

  http.get('/api/clusters/:cluster/namespaces/:namespace/deployments/:name/pods', ({ params }) => {
    const deployment = mockDeployments.find(d => 
      d.metadata.name === params.name && 
      d.metadata.namespace === params.namespace
    )
    if (!deployment) {
      return new HttpResponse(null, { status: 404 })
    }
    // 返回与该 deployment 相关的 pods（简化逻辑）
    const appLabel = deployment.metadata.labels?.app
    const pods = mockPods.filter(p => 
      p.metadata.namespace === params.namespace &&
      p.metadata.name.includes(appLabel || '')
    )
    return HttpResponse.json(pods)
  }),

  http.post('/api/clusters/:cluster/namespaces/:namespace/deployments/:name/scale', async ({ params, request }) => {
    const body = await request.json() as { replicas: number }
    return HttpResponse.json({
      message: 'Deployment scaled successfully',
      replicas: body.replicas,
    })
  }),

  http.post('/api/clusters/:cluster/namespaces/:namespace/deployments/:name/restart', ({ params }) => {
    return HttpResponse.json({
      message: `Deployment ${params.name} restarted successfully`,
    })
  }),

  // Event APIs
  http.get('/api/clusters/:cluster/events', ({ request }) => {
    const url = new URL(request.url)
    const namespace = url.searchParams.get('namespace')
    if (namespace) {
      return HttpResponse.json(mockEvents.filter(e => e.metadata.namespace === namespace))
    }
    return HttpResponse.json(mockEvents)
  }),

  http.get('/api/clusters/:cluster/namespaces/:namespace/events', ({ params }) => {
    const events = mockEvents.filter(e => e.metadata.namespace === params.namespace)
    return HttpResponse.json(events)
  }),

  // Monitoring APIs
  http.get('/api/monitoring/:cluster/status', ({ params }) => {
    return HttpResponse.json({
      cluster: params.cluster,
      active: true,
      dataPoints: 100,
    })
  }),

  http.get('/api/monitoring/:cluster/alerts', ({ params }) => {
    const alerts = mockAlerts.filter(a => a.cluster === params.cluster)
    return HttpResponse.json(alerts)
  }),

  http.get('/api/monitoring/:cluster/history', ({ params }) => {
    return HttpResponse.json(mockMetricsHistory.map(h => ({
      ...h,
      clusterName: params.cluster,
    })))
  }),
]

// 模拟延迟的 handlers（用于测试加载状态）
export const slowHandlers = [
  http.get('/api/clusters', async () => {
    await new Promise(resolve => setTimeout(resolve, 1000))
    return HttpResponse.json(mockClusters)
  }),
]

// 模拟错误的 handlers（用于测试错误处理）
export const errorHandlers = [
  http.get('/api/clusters', () => {
    return new HttpResponse(
      JSON.stringify({ error: 'Internal Server Error' }),
      { status: 500 }
    )
  }),

  http.get('/api/clusters/:name/status', () => {
    return new HttpResponse(
      JSON.stringify({ error: 'Cluster not reachable' }),
      { status: 503 }
    )
  }),
]
