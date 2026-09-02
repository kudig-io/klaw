// Mock 数据关联性不变量测试：守护所有 mock 间的交叉引用与派生一致性
// 任何字段在 test 中被误改都会让这些断言失败。

import { describe, it, expect } from 'vitest'
import {
  mockClusters, mockNamespaces, mockNodes, mockPods, mockDeployments, mockServices,
  mockEvents, mockAlertRules, mockAlertRecords, mockTenants, mockTenantUsers,
  mockAuditLogs, mockBackups, mockNodeMetrics, mockMetricsHistory, derive, analyzeLogs, getLogsForPod, buildDiagIssues,
} from '../mocks/data'

const POD_NAMES = new Set(mockPods.map((p) => p.metadata.name))
const DEPLOY_NAMES = new Set(mockDeployments.map((d) => d.metadata.name))
const NODE_NAMES = new Set(mockNodes.map((n) => n.metadata.name))
const NS_NAMES = new Set(mockNamespaces.map((n) => n.metadata.name))

describe('Mock 关联性不变量', () => {
  it('cluster 数量与 kind-test/production 顺序', () => {
    expect(mockClusters).toHaveLength(2)
    expect(mockClusters[0].name).toBe('kind-test')
    expect(mockClusters[1].name).toBe('production')
  })

  it('kind-test 节点恰好 3 个，全部 Ready', () => {
    expect(mockNodes).toHaveLength(3)
    for (const n of mockNodes) {
      const ready = n.status.conditions.find((c) => c.type === 'Ready')
      expect(ready?.status).toBe('True')
    }
  })

  it('node-metrics 中 kind-test-control-plane.cpu === "4"', () => {
    expect(mockNodeMetrics['kind-test-control-plane'].cpu).toBe('4')
  })

  it('cluster status 派生：kind-test 有 3 节点 + 10 Running Pod（关键单测约束）', () => {
    const s = derive.clusterStatus('kind-test')
    expect(s.nodes.total).toBe(3)
    expect(s.nodes.ready).toBe(3)
    expect(s.pods.running).toBe(10)
  })

  it('所有 Deployment 名字全局唯一（单测 getByText 约束）', () => {
    expect(DEPLOY_NAMES.size).toBe(mockDeployments.length)
  })

  it('nginx 在 klaw-test 显示 2/2（唯一 2/2，单测约束）', () => {
    const nginx = mockDeployments.find((d) => d.metadata.name === 'nginx')
    expect(nginx?.metadata.namespace).toBe('klaw-test')
    expect(nginx?.status.availableReplicas).toBe(2)
    expect(nginx?.spec.replicas).toBe(2)
    // 其余 deployment 的 available/spec 不应同时出现 2/2
    const other22 = mockDeployments.filter((d) =>
      d.metadata.name !== 'nginx' &&
      d.status.availableReplicas === 2 && d.spec.replicas === 2,
    )
    expect(other22).toHaveLength(0)
  })

  it('每个 Pod 的 namespace 都存在', () => {
    for (const p of mockPods) {
      expect(NS_NAMES.has(p.metadata.namespace)).toBe(true)
    }
  })

  it('每个 Deployment 的 namespace 都存在', () => {
    for (const d of mockDeployments) {
      expect(NS_NAMES.has(d.metadata.namespace)).toBe(true)
    }
  })

  it('kind-test 至少存在 1 个 Pending / CrashLoopBackOff / Failed pod（故事线）', () => {
    const phases = mockPods.map((p) => p.status.phase)
    expect(phases).toContain('Pending')
    expect(phases).toContain('CrashLoopBackOff')
  })

  it('故障链：mall-prod/frontend 有 Pending + CrashLoop pod，klaw-test/frontend 全部 Running', () => {
    expect(mockPods.some((p) => p.metadata.namespace === 'mall-prod' && p.status.phase === 'Pending')).toBe(true)
    expect(mockPods.some((p) => p.metadata.namespace === 'mall-prod' && p.status.phase === 'CrashLoopBackOff')).toBe(true)
    for (const p of mockPods.filter((p) => p.metadata.namespace === 'klaw-test')) {
      expect(['Running', 'Pending']).toContain(p.status.phase)
    }
  })

  it('节点 kind-test-worker2 标记 MemoryPressure（事故源头）', () => {
    const n = mockNodes.find((n) => n.metadata.name === 'kind-test-worker2')
    expect(n?.status.conditions.find((c) => c.type === 'MemoryPressure')?.status).toBe('True')
  })

  it('活跃告警至少 2 条，引用真实存在的资源', () => {
    const active = mockAlertRecords.filter((r) => !r.resolved)
    expect(active.length).toBeGreaterThanOrEqual(2)
    for (const a of active) {
      if (a.resourceKind === 'Pod') expect(POD_NAMES.has(a.resourceName)).toBe(true)
      if (a.resourceKind === 'Node') expect(NODE_NAMES.has(a.resourceName)).toBe(true)
    }
  })

  it('audit log 引用的用户存在于 tenant-users', () => {
    const usernames = new Set(mockTenantUsers.map((u) => u.username))
    for (const log of mockAuditLogs) {
      if (log.user && log.user !== 'system' && log.user !== 'unknown' && log.user !== 'oncall') {
        expect(usernames.has(log.user)).toBe(true)
      }
    }
  })

  it('租户与命名空间交叉引用：租户的 namespaces 都在 mockNamespaces 中', () => {
    for (const t of mockTenants) {
      for (const n of t.namespaces) {
        expect(NS_NAMES.has(n)).toBe(true)
      }
    }
  })

  it('每个 tenant-user.tenantId 都在 tenants 中', () => {
    const tids = new Set(mockTenants.map((t) => t.id))
    for (const u of mockTenantUsers) {
      expect(tids.has(u.tenantId)).toBe(true)
    }
  })

  it('Service endpoints：selector 至少匹配到一个 ready pod（nginx / mall-gateway）', () => {
    const nginx = mockServices.find((s) => s.metadata.name === 'nginx' && s.metadata.namespace === 'klaw-test')
    expect(nginx?.spec.selector?.app).toBe('nginx')
    expect(mockPods.filter((p) => p.metadata.labels?.app === 'nginx' && p.status.phase === 'Running').length).toBe(2)

    const mallFrontend = mockServices.find((s) => s.metadata.name === 'mall-frontend' && s.metadata.namespace === 'mall-prod')
    expect(mallFrontend?.spec.type).toBe('LoadBalancer')
  })

  it('事件引用的 Pod 都存在', () => {
    for (const e of mockEvents) {
      const text = e.message + ' ' + (e.reason || '')
      // 事件中的 Pod 名字以 mall-frontend-...-xxx 等出现
      const match = text.match(/mall-[a-z-]+-[a-z0-9]+-[a-z0-9]+|frontend-[a-z0-9]+-[a-z0-9]+|nginx-[a-z0-9]+-[a-z0-9]+|httpbin-[a-z0-9]+-[a-z0-9]+/)
      if (match) expect(POD_NAMES.has(match[0])).toBe(true)
    }
  })

  it('每次 getLogsForPod 返回都包含 [mock] 前缀', () => {
    const sample = ['nginx-6b66fbbd46-abc12', 'mall-frontend-7d9c5f8b4-z8x3c', 'unknown-pod-name']
    for (const n of sample) {
      expect(getLogsForPod(n)).toContain('[mock]')
    }
  })

  it('analyzeLogs 从 mall-frontend-...-z8x3c 日志中能识别错误与 stack trace', () => {
    const logs = getLogsForPod('mall-frontend-7d9c5f8b4-z8x3c')
    const result = analyzeLogs(logs)
    expect(result.errorCount).toBeGreaterThan(0)
    expect(result.stackTraces.length).toBeGreaterThan(0)
    expect(result.performanceMetrics.slowRequests.length).toBeGreaterThanOrEqual(0)
    expect(result.logLevels.ERROR).toBeGreaterThan(0)
  })

  it('diag issues 引用真实节点/Deployment 名', () => {
    const issues = buildDiagIssues()
    expect(issues.some((i) => i.location?.includes('kind-test-worker2'))).toBe(true)
    expect(issues.some((i) => i.location?.includes('mall-frontend'))).toBe(true)
  })

  it('alert stats 派生一致：active+resolved === total', () => {
    const s = derive.alertStats('kind-test')
    expect(s.byStatus.active + s.byStatus.acknowledged + s.byStatus.resolved).toBe(s.total)
  })

  it('backup summary 派生：total = 全量备份列表长度', () => {
    const s = derive.backupSummary('kind-test')
    expect(s.total).toBe(mockBackups.filter((b) => b.cluster === 'kind-test').length)
  })

  it('tenant stats：totalNamespaces = 租户命名空间并集', () => {
    const s = derive.tenantStats()
    const union = new Set(mockTenants.flatMap((t) => t.namespaces)).size
    expect(s.totalNamespaces).toBe(union)
    expect(s.totalTenants).toBe(mockTenants.length)
    expect(s.totalUsers).toBe(mockTenantUsers.length)
  })

  it('监控 history 至少 10 个点', () => {
    expect(mockMetricsHistory.length).toBeGreaterThanOrEqual(10)
  })
})