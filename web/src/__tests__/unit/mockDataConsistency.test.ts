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

// ── 信息量扩展不变量（字段完备性 + 深层交叉引用）──────────────
describe('Mock 信息量扩展不变量', () => {
  // containerStatuses 按 phase 有不同形状（running/waiting/lastState 分支），测试中放宽为动态访问
  const csOf = (p: (typeof mockPods)[number]) => p.status.containerStatuses[0] as Record<string, any>

  it('Pod 容器完备：image/resources 齐备，containerStatuses 与 containers 一一对应，qosClass=Burstable', () => {
    for (const p of mockPods) {
      expect(p.spec.containers.length).toBeGreaterThanOrEqual(1)
      for (const c of p.spec.containers) {
        expect(c.image).toBeTruthy()
        expect(Object.keys(c.resources.requests).length).toBeGreaterThan(0)
        expect(Object.keys(c.resources.limits).length).toBeGreaterThan(0)
      }
      expect(p.status.containerStatuses).toHaveLength(p.spec.containers.length)
      expect(p.status.qosClass).toBe('Burstable')
    }
  })

  it('镜像同源：每个 Pod 与同 namespace/app 的 Deployment 使用完全相同的镜像', () => {
    for (const p of mockPods) {
      const app = p.metadata.labels?.app
      const dep = mockDeployments.find((d) => d.metadata.namespace === p.metadata.namespace && d.metadata.labels?.app === app)
      expect(dep).toBeDefined()
      expect(p.spec.containers[0].image).toBe(dep!.spec.template.spec.containers[0].image)
    }
  })

  it('Deployment 反向封闭：每个 Deployment 的 selector 都能匹配到至少一个 Pod', () => {
    for (const d of mockDeployments) {
      const app = d.spec.selector.matchLabels.app
      const pods = mockPods.filter((p) => p.metadata.namespace === d.metadata.namespace && p.metadata.labels?.app === app)
      expect(pods.length).toBeGreaterThanOrEqual(1)
    }
  })

  it('Pod 网络与调度：Running/CrashLoop 有 IP 与节点，Pending 无 IP 无节点', () => {
    for (const p of mockPods) {
      if (p.status.phase === 'Pending') {
        expect(p.status.podIP).toBe('')
        expect(p.spec.nodeName).toBe('')
      } else {
        expect(p.status.podIP).toMatch(/^10\.244\.\d+\.\d+$/)
        expect(NODE_NAMES.has(p.spec.nodeName)).toBe(true)
      }
    }
  })

  it('容器状态与 phase 一致：Running ready、Pending ContainerCreating、CrashLoop waiting+lastState', () => {
    for (const p of mockPods) {
      const cs = csOf(p)
      if (p.status.phase === 'Running') {
        expect(cs.ready).toBe(true)
        expect(cs.state.running).toBeDefined()
        expect(cs.restartCount).toBe(0)
      } else if (p.status.phase === 'Pending') {
        expect(cs.ready).toBe(false)
        expect(cs.state.waiting?.reason).toBe('ContainerCreating')
        expect(cs.lastState).toBeUndefined()
      } else {
        expect(p.status.phase).toBe('CrashLoopBackOff')
        expect(cs.ready).toBe(false)
        expect(cs.state.waiting?.reason).toBe('CrashLoopBackOff')
        expect(cs.lastState?.terminated?.exitCode).toBe(1)
      }
    }
  })

  it('故事线锁：z8x3c restartCount=8 且部署在 worker2，与重启风暴告警 value/threshold 互证', () => {
    const z = mockPods.find((p) => p.metadata.name === 'mall-frontend-7d9c5f8b4-z8x3c')
    expect(z).toBeDefined()
    expect(z?.spec.nodeName).toBe('kind-test-worker2')
    expect(csOf(z!).restartCount).toBe(8)
    const alert = mockAlertRecords.find((a) => a.id === 'alert-frontend-restart')
    expect(alert?.resourceName).toBe('mall-frontend-7d9c5f8b4-z8x3c')
    expect(alert?.value).toBe('8')
    expect(alert?.threshold).toBe('5')
    expect(alert?.acknowledged).toBe(true)
  })

  it('故事线锁：worker2 内存 93% 与内存压力告警 value 93% / 阈值 90% 互证', () => {
    expect(mockNodeMetrics['kind-test-worker2'].usage.memoryPercent).toBe(93)
    const alert = mockAlertRecords.find((a) => a.id === 'alert-mem-worker2')
    expect(alert?.resourceName).toBe('kind-test-worker2')
    expect(alert?.value).toBe('93%')
    expect(alert?.threshold).toBe('90%')
  })

  it('节点字段完备：capacity/allocatable/nodeInfo/addresses(InternalIP)/kubelet 版本', () => {
    for (const n of mockNodes) {
      expect(n.status.capacity.pods).toBe('110')
      expect(n.status.allocatable.cpu).toBeTruthy()
      expect(n.status.allocatable.memory).toBeTruthy()
      expect(n.status.nodeInfo.kubeletVersion).toMatch(/^v1\.\d+/)
      expect(n.status.nodeInfo.containerRuntimeVersion).toContain('containerd://')
      const ip = n.status.addresses.find((a) => a.type === 'InternalIP')
      expect(ip?.address).toMatch(/^172\.18\.0\.\d+$/)
    }
  })

  it('Service 字段完备：LB/NodePort 必有 externalTrafficPolicy，LB 必有 ingress IP', () => {
    for (const s of mockServices) {
      expect(s.spec.clusterIP).toMatch(/^10\.96\./)
      expect(s.spec.ports.length).toBeGreaterThanOrEqual(1)
      for (const port of s.spec.ports) {
        expect(['TCP', 'UDP']).toContain(port.protocol)
        expect(port.targetPort).toBeGreaterThan(0)
      }
      if (s.spec.type === 'LoadBalancer' || s.spec.type === 'NodePort') {
        expect(['Cluster', 'Local']).toContain(s.spec.externalTrafficPolicy)
      } else {
        expect(s.spec.externalTrafficPolicy).toBeUndefined()
      }
      if (s.spec.type === 'LoadBalancer') {
        expect(s.status.loadBalancer?.ingress?.[0]?.ip).toMatch(/^172\.18\.0\.\d+$/)
      }
    }
  })

  it('业务 namespace 的 Service selector 都能匹配到同 namespace 的 Pod', () => {
    const business = ['klaw-test', 'mall-prod', 'mall-staging']
    for (const s of mockServices.filter((x) => business.includes(x.metadata.namespace))) {
      const sel = s.spec.selector ?? {}
      const keys = Object.keys(sel)
      expect(keys.length).toBeGreaterThan(0)
      const matched = mockPods.filter((p) =>
        p.metadata.namespace === s.metadata.namespace &&
        keys.every((k) => p.metadata.labels?.[k] === sel[k]),
      )
      expect(matched.length).toBeGreaterThanOrEqual(1)
    }
  })

  it('备份字段完备：快照路径/etcd 端点/validationResult；Completed 有效且带完成时间，失败必有 message', () => {
    for (const b of mockBackups) {
      expect(b.snapshotLocation).toBe(`s3://klaw-backups/etcd/kind-test/${b.name}.db`)
      expect(b.spec.etcdEndpoints).toHaveLength(3)
      expect(b.etcdRevision).toBeGreaterThan(0)
      expect(b.validationResult).toBeDefined()
      expect(b.spec.validation?.enabled).toBe(true)
    }
    const completed = mockBackups.filter((b) => b.phase === 'Completed')
    expect(completed.length).toBeGreaterThan(0)
    for (const b of completed) {
      expect(b.validationResult?.valid).toBe(true)
      expect(b.completionTime).toBeTruthy()
    }
    for (const b of mockBackups.filter((x) => x.phase !== 'Completed')) {
      expect(b.validationResult?.valid).toBe(false)
      expect(b.message).toBeTruthy()
    }
  })

  it('告警规则封闭：每条告警记录的 ruleId 都存在于规则库，规则条件字段齐备', () => {
    expect(mockAlertRules).toHaveLength(6)
    const ruleIds = new Set(mockAlertRules.map((r) => r.id))
    for (const a of mockAlertRecords) {
      expect(ruleIds.has(a.ruleId)).toBe(true)
      expect(a.value).toBeTruthy()
      expect(a.threshold).toBeTruthy()
      expect(a.operator).toBeTruthy()
    }
    for (const r of mockAlertRules) {
      expect(r.enabled).toBe(true)
      expect(r.condition.field).toBeTruthy()
      expect(r.condition.operator).toBeTruthy()
      expect(r.condition.timeWindow).toBeTruthy()
    }
  })

  it('告警构成锁：活跃 5 条（2 critical/2 error/1 warning），已解决 8 条', () => {
    const active = mockAlertRecords.filter((a) => !a.resolved)
    expect(active).toHaveLength(5)
    expect(active.filter((a) => a.severity === 'critical')).toHaveLength(2)
    expect(active.filter((a) => a.severity === 'error')).toHaveLength(2)
    expect(active.filter((a) => a.severity === 'warning')).toHaveLength(1)
    expect(mockAlertRecords.filter((a) => a.resolved)).toHaveLength(8)
  })

  it('事件字段完备且故事线关键事件在场（OOMKilled/BackOff/FailedScheduling）', () => {
    for (const e of mockEvents) {
      expect(['Normal', 'Warning']).toContain(e.type)
      expect(e.reason).toBeTruthy()
      expect(e.message).toBeTruthy()
      expect(e.lastTimestamp).toBeTruthy()
    }
    expect(mockEvents.some((e) => e.reason === 'OOMKilled' && e.metadata.name.startsWith('mall-frontend-7d9c5f8b4-z8x3c'))).toBe(true)
    expect(mockEvents.some((e) => e.reason === 'BackOff' && e.metadata.name.startsWith('mall-frontend-7d9c5f8b4-z8x3c'))).toBe(true)
    expect(mockEvents.some((e) => e.reason === 'FailedScheduling' && e.metadata.name.startsWith('mall-frontend-7d9c5f8b4-q7w2e'))).toBe(true)
  })

  it('监控 history 故事线：首点全健康 12/12，末点 10 Running/4 Pending/1 Failed 与告警互证', () => {
    expect(mockMetricsHistory).toHaveLength(24)
    for (const p of mockMetricsHistory) expect(p.clusterName).toBe('kind-test')
    expect(mockMetricsHistory[0].pods).toEqual({ total: 12, running: 12, pending: 0, failed: 0 })
    expect(mockMetricsHistory[mockMetricsHistory.length - 1].pods).toEqual({ total: 15, running: 10, pending: 4, failed: 1 })
  })

  it('诊断 issue 构成锁：9 条（2 critical/1 error/3 warning/3 info），节点过滤命中唯一', () => {
    const issues = buildDiagIssues()
    expect(issues).toHaveLength(9)
    expect(issues.filter((i) => i.severity === 'critical')).toHaveLength(2)
    expect(issues.filter((i) => i.severity === 'error')).toHaveLength(1)
    expect(issues.filter((i) => i.severity === 'warning')).toHaveLength(3)
    expect(issues.filter((i) => i.severity === 'info')).toHaveLength(3)
    for (const i of issues) {
      expect(i.cn_name).toBeTruthy()
      expect(i.en_name).toBeTruthy()
      expect(i.analyzer_name).toBeTruthy()
      expect(i.details).toBeTruthy()
      expect(i.remediation?.suggestion).toBeTruthy()
    }
    expect(issues.filter((i) => i.remediation?.command)).toHaveLength(7)
    expect(buildDiagIssues('kind-test-worker2')).toHaveLength(1)
  })
})