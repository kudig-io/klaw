// 集群 / 节点 / Pod / Deployment /Event handlers

import { http, HttpResponse } from 'msw'
import {
  mockClusters, mockNamespaces, mockEvents, mockNodes, mockNodeMetrics, mockDeployments, derive, deploymentsStore,
} from '../data/index'
import { store } from '../store'

export const clusterHandlers = [
  http.get('/api/v1/clusters', () => HttpResponse.json(mockClusters)),
  http.get('/api/v1/clusters/:name', ({ params }) => {
    const c = mockClusters.find((c) => c.name === params.name)
    return c ? HttpResponse.json(c) : new HttpResponse(null, { status: 404 })
  }),
  http.get('/api/v1/clusters/:name/status', ({ params }) =>
    HttpResponse.json(derive.clusterStatus(params.name as string))),
  http.get('/api/v1/clusters/:name/metrics', ({ params }) =>
    HttpResponse.json(derive.clusterMetrics(params.name as string))),
  http.get('/api/v1/clusters/:name/namespaces', () => HttpResponse.json(mockNamespaces)),
  http.get('/api/v1/clusters/:cluster/deployments', () => HttpResponse.json(mockDeployments)),
]

export const nodeHandlers = [
  http.get('/api/v1/clusters/:cluster/nodes', () => HttpResponse.json(mockNodes)),
  // /nodes/metrics 必须早于 /nodes/:name（避免被当成节点名）
  http.get('/api/v1/clusters/:cluster/nodes/metrics', () => HttpResponse.json(mockNodeMetrics)),
  http.get('/api/v1/clusters/:cluster/nodes/:name', ({ params }) => {
    const node = mockNodes.find((n) => n.metadata.name === params.name)
    return node ? HttpResponse.json(node) : new HttpResponse(null, { status: 404 })
  }),
]

export const podHandlers = [
  http.get('/api/v1/clusters/:cluster/pods', ({ request }) => {
    const url = new URL(request.url)
    const namespace = url.searchParams.get('namespace')
    const pods = namespace ? store.pods.filter((p) => p.metadata.namespace === namespace) : store.pods
    return HttpResponse.json(pods)
  }),
  http.get('/api/v1/clusters/:cluster/namespaces/:namespace/pods', ({ params }) => {
    return HttpResponse.json(store.pods.filter((p) => p.metadata.namespace === params.namespace))
  }),
  http.get('/api/v1/clusters/:cluster/namespaces/:namespace/pods/:name', ({ params }) => {
    const pod = store.pods.find((p) => p.metadata.name === params.name && p.metadata.namespace === params.namespace)
    return pod ? HttpResponse.json(pod) : new HttpResponse(null, { status: 404 })
  }),
  http.delete('/api/v1/clusters/:cluster/namespaces/:namespace/pods/:name', ({ params }) => {
    const idx = store.pods.findIndex((p) => p.metadata.name === params.name && p.metadata.namespace === params.namespace)
    if (idx >= 0) store.pods.splice(idx, 1)
    return HttpResponse.json({ message: `Pod ${params.name} deleted successfully` })
  }),
]

export const deploymentHandlers = [
  http.get('/api/v1/clusters/:cluster/namespaces/:namespace/deployments', ({ params }) => {
    return HttpResponse.json(deploymentsStore.filter((d) => d.metadata.namespace === params.namespace))
  }),
  http.get('/api/v1/clusters/:cluster/namespaces/:namespace/deployments/:name', ({ params }) => {
    const d = deploymentsStore.find((d) => d.metadata.name === params.name && d.metadata.namespace === params.namespace)
    return d ? HttpResponse.json(d) : new HttpResponse(null, { status: 404 })
  }),
  http.get('/api/v1/clusters/:cluster/namespaces/:namespace/deployments/:name/status', ({ params }) => {
    const d = deploymentsStore.find((d) => d.metadata.name === params.name && d.metadata.namespace === params.namespace)
    if (!d) return new HttpResponse(null, { status: 404 })
    return HttpResponse.json({
      name: d.metadata.name, namespace: d.metadata.namespace,
      replicas: d.status.replicas, availableReplicas: d.status.availableReplicas,
      readyReplicas: d.status.readyReplicas, updatedReplicas: d.status.updatedReplicas,
      conditions: d.status.conditions,
    })
  }),
  http.get('/api/v1/clusters/:cluster/namespaces/:namespace/deployments/:name/pods', ({ params }) => {
    const d = deploymentsStore.find((d) => d.metadata.name === params.name && d.metadata.namespace === params.namespace)
    if (!d) return HttpResponse.json([])
    const appLabel = d.metadata.labels?.app
    return HttpResponse.json(
      store.pods.filter((p) => p.metadata.namespace === params.namespace && appLabel && p.metadata.labels?.app === appLabel)
    )
  }),
  http.post('/api/v1/clusters/:cluster/namespaces/:namespace/deployments/:name/scale', async ({ params, request }) => {
    const body = await request.json() as { replicas: number }
    const d = deploymentsStore.find((d) => d.metadata.name === params.name && d.metadata.namespace === params.namespace)
    if (d) {
      d.spec.replicas = body.replicas
      d.status.replicas = body.replicas
      d.status.updatedReplicas = body.replicas
      // pending replicas (ready/available unchanged until new pods come up)
    }
    return HttpResponse.json({ message: 'Deployment scaled successfully', replicas: body.replicas })
  }),
  http.post('/api/v1/clusters/:cluster/namespaces/:namespace/deployments/:name/restart', ({ params }) => {
    return HttpResponse.json({ message: `Deployment ${params.name} restarted successfully` })
  }),
]

export const eventHandlers = [
  http.get('/api/v1/clusters/:cluster/events', ({ request }) => {
    const url = new URL(request.url)
    const namespace = url.searchParams.get('namespace')
    return HttpResponse.json(namespace ? mockEvents.filter((e: any) => e.metadata.namespace === namespace) : mockEvents)
  }),
  http.get('/api/v1/clusters/:cluster/namespaces/:namespace/events', ({ params }) => {
    return HttpResponse.json(mockEvents.filter((e: any) => e.metadata.namespace === params.namespace))
  }),
]