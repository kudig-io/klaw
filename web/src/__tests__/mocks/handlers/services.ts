// Service + Endpoints handlers

import { http, HttpResponse } from 'msw'
import { mockServices, derive } from '../data/index'
import { store, appendAudit } from '../store'

export const serviceHandlers = [
  http.get('/api/v1/clusters/:cluster/services', () => HttpResponse.json(mockServices)),
  http.get('/api/v1/clusters/:cluster/namespaces/:namespace/services', ({ params }) =>
    HttpResponse.json(mockServices.filter((s) => s.metadata.namespace === params.namespace))),
  http.get('/api/v1/clusters/:cluster/namespaces/:namespace/services/:name', ({ params }) => {
    const s = mockServices.find((s) => s.metadata.name === params.name && s.metadata.namespace === params.namespace)
    return s ? HttpResponse.json(s) : new HttpResponse(null, { status: 404 })
  }),
  http.get('/api/v1/clusters/:cluster/namespaces/:namespace/services/:name/endpoints', ({ params }) =>
    HttpResponse.json(derive.serviceEndpoints(params.cluster as string, params.namespace as string, params.name as string))),
  http.delete('/api/v1/clusters/:cluster/namespaces/:namespace/services/:name', ({ params }) => {
    const idx = mockServices.findIndex((s) => s.metadata.name === params.name && s.metadata.namespace === params.namespace)
    if (idx >= 0) mockServices.splice(idx, 1)
    appendAudit({
      eventType: 'service.delete', category: 'resource', severity: 'info', user: 'oncall',
      action: `delete service ${params.name}`, resource: { kind: 'Service', name: params.name as string },
      cluster: params.cluster as string, namespace: params.namespace as string,
      result: 'success',
    })
    return HttpResponse.json({ message: `Service ${params.name} deleted successfully` })
  }),
]