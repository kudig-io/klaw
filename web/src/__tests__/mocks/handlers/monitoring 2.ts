// 监控 + 告警（rules / evaluate / history / stats / acknowledge / resolve）

import { http, HttpResponse } from 'msw'
import { mockMetricsHistory, mockAlertRules, mockAlertRecords, mockAlerts, derive } from '../data/index'
import { store, appendAudit, nextAlertId, now } from '../store'

export const monitorHandlers = [
  http.get('/api/v1/clusters/:cluster/monitor/status', ({ params }) => {
    const history = mockMetricsHistory.filter((h) => h.clusterName === params.cluster)
    const rules = mockAlertRules.filter((r) => r.cluster === 'kind-test' || r.cluster === undefined)
    return HttpResponse.json({
      cluster: params.cluster,
      active: true,
      dataPoints: history.length,
      interval: '5m',
      evalInterval: '1m',
      rulesEnabled: rules.filter((r) => r.enabled).length,
      rulesTotal: rules.length,
      lastEvaluation: now(),
      timestamp: now(),
    })
  }),
  http.get('/api/v1/clusters/:cluster/monitor/alerts', ({ params }) => {
    return HttpResponse.json(mockAlerts.filter((a: any) => a.cluster === params.cluster))
  }),
  http.get('/api/v1/clusters/:cluster/monitor/history', ({ params }) => {
    return HttpResponse.json(
      mockMetricsHistory.map((h) => ({ ...h, clusterName: params.cluster }))
    )
  }),
]

export const alertHandlers = [
  http.get('/api/v1/clusters/:cluster/alerts/rules', () => HttpResponse.json(mockAlertRules.filter((r) => r.cluster === 'kind-test' || r.cluster === undefined))),
  http.post('/api/v1/clusters/:cluster/alerts/rules', () => HttpResponse.json({ message: 'rule created' }, { status: 201 })),
  http.put('/api/v1/clusters/:cluster/alerts/rules/:id', () => HttpResponse.json({ message: 'rule updated' })),
  http.delete('/api/v1/clusters/:cluster/alerts/rules/:id', () => new HttpResponse(null, { status: 204 })),

  http.get('/api/v1/clusters/:cluster/alerts/history', ({ params, request }) => {
    const url = new URL(request.url)
    const limit = parseInt(url.searchParams.get('limit') || '50', 10)
    return HttpResponse.json(derive.alertHistory(params.cluster as string, limit))
  }),
  http.get('/api/v1/clusters/:cluster/alerts/stats', ({ params }) =>
    HttpResponse.json(derive.alertStats(params.cluster as string))),

  // 求值：根据当前 store 状态评估规则，生成新的 active alert（去重）
  http.post('/api/v1/clusters/:cluster/alerts/evaluate', ({ params }) => {
    const triggered: any[] = []
    const ts = now()
    for (const rule of mockAlertRules.filter((r) => r.cluster === 'kind-test' || !r.cluster)) {
      if (!rule.enabled) continue
      if (rule.id === 'rule-memory-pressure') {
        // worker2 MemoryPressure → critical
        const exists = store.alertRecords.find((r) => !r.resolved && r.ruleId === rule.id && r.resourceName === 'kind-test-worker2')
        if (!exists) {
          const rec = {
            id: nextAlertId(), cluster: params.cluster as string, ruleId: rule.id, ruleName: rule.name, ruleType: 'node',
            resourceKind: 'Node', resourceName: 'kind-test-worker2', severity: 'critical' as const,
            value: '93%', threshold: '90%', operator: '>', message: 'kind-test-worker2 内存使用 93%（手动求值触发）',
            acknowledged: false, resolved: false, createdAt: ts,
          }
          store.alertRecords.unshift(rec)
          triggered.push(rec)
        }
      } else if (rule.id === 'rule-pod-restart-storm') {
        const exists = store.alertRecords.find((r) => !r.resolved && r.ruleId === rule.id && r.resourceName === 'mall-frontend-7d9c5f8b4-z8x3c')
        if (!exists) {
          const rec = {
            id: nextAlertId(), cluster: params.cluster as string, ruleId: rule.id, ruleName: rule.name, ruleType: 'pod',
            resourceKind: 'Pod', resourceName: 'mall-frontend-7d9c5f8b4-z8x3c', namespace: 'mall-prod',
            severity: 'critical' as const, value: '8', threshold: '5', operator: '>',
            message: 'mall-frontend-7d9c5f8b4-z8x3c 30 分钟内重启 8 次',
            acknowledged: false, resolved: false, createdAt: ts,
          }
          store.alertRecords.unshift(rec)
          triggered.push(rec)
        }
      }
    }
    return HttpResponse.json(triggered)
  }),

  http.post('/api/v1/clusters/:cluster/alerts/:id/acknowledge', ({ params }) => {
    const rec = store.alertRecords.find((r) => r.id === params.id)
    if (!rec) return new HttpResponse(null, { status: 404 })
    rec.acknowledged = true
    rec.acknowledgedAt = now()
    appendAudit({
      eventType: 'alert.acknowledge', category: 'tenancy', severity: 'info', user: 'oncall',
      action: `ack alert ${rec.id}`, resource: { kind: 'Alert', name: rec.resourceName },
      cluster: params.cluster as string, namespace: rec.namespace,
      result: 'success',
    })
    return HttpResponse.json(rec)
  }),
  http.post('/api/v1/clusters/:cluster/alerts/:id/resolve', ({ params }) => {
    const rec = store.alertRecords.find((r) => r.id === params.id)
    if (!rec) return new HttpResponse(null, { status: 404 })
    rec.resolved = true
    rec.resolvedAt = now()
    appendAudit({
      eventType: 'alert.resolve', category: 'tenancy', severity: 'info', user: 'oncall',
      action: `resolve alert ${rec.id}`, resource: { kind: 'Alert', name: rec.resourceName },
      cluster: params.cluster as string, namespace: rec.namespace,
      result: 'success',
    })
    return HttpResponse.json(rec)
  }),
]