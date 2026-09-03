// Ingress + NetworkPolicy + 网络分析 handlers

import { http, HttpResponse } from 'msw'
import { mockIngresses, mockNetworkPolicies, derive } from '../data/index'

export const networkHandlers = [
  http.get('/api/v1/clusters/:cluster/ingresses', () => HttpResponse.json(mockIngresses)),
  http.get('/api/v1/clusters/:cluster/namespaces/:namespace/ingresses', ({ params }) =>
    HttpResponse.json(mockIngresses.filter((i) => i.metadata.namespace === params.namespace))),
  http.get('/api/v1/clusters/:cluster/networkpolicies', () => HttpResponse.json(mockNetworkPolicies)),
  http.get('/api/v1/clusters/:cluster/namespaces/:namespace/networkpolicies', ({ params }) =>
    HttpResponse.json(mockNetworkPolicies.filter((p) => p.metadata.namespace === params.namespace))),
  http.get('/api/v1/analysis/network', () => HttpResponse.json(derive.networkAnalysis())),
]
