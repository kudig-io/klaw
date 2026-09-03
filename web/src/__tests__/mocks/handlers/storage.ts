// PVC + PV + StorageClass + 存储分析 handlers

import { http, HttpResponse } from 'msw'
import { mockPVCs, mockPVs, mockStorageClasses, derive } from '../data/index'

export const storageHandlers = [
  http.get('/api/v1/clusters/:cluster/persistentvolumeclaims', () => HttpResponse.json(mockPVCs)),
  http.get('/api/v1/clusters/:cluster/namespaces/:namespace/persistentvolumeclaims', ({ params }) =>
    HttpResponse.json(mockPVCs.filter((c) => c.metadata.namespace === params.namespace))),
  http.get('/api/v1/clusters/:cluster/persistentvolumes', () => HttpResponse.json(mockPVs)),
  http.get('/api/v1/clusters/:cluster/storageclasses', () => HttpResponse.json(mockStorageClasses)),
  http.get('/api/v1/analysis/storage', () => HttpResponse.json(derive.storageAnalysis())),
]
