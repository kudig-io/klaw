// 派生层：所有计数 / 统计 / endpoints 等都从基础数据实时派生
// 保证 mock 数据集的内部一致性（改一处全联动）

import type { Pod, Node } from '../../../lib/api'
import { mockPods } from './workloads'
import { mockNodes, mockNodeMetrics } from './nodes'
import { mockServices } from './services'
import { mockAlertRules, mockAlertRecords, mockMetricsHistory } from './observability'
import { mockTenants, mockTenantUsers, mockAuditLogs } from './governance'
import { mockBackups } from './backups'

const MS = 1000

// ── ClusterStatus ─────────────────────────────────────────
export function deriveClusterStatus(clusterName: string) {
  const clusterNodes: Node[] = mockNodes
  const clusterPods: Pod[] = mockPods.filter((p) => p.metadata.namespace !== 'production-mock') // 全部 kind-test
  if (clusterName === 'kind-test') {
    const ready = clusterNodes.filter((n) => {
      const cond = n.status.conditions.find((c) => c.type === 'Ready')
      return cond?.status === 'True'
    }).length
    const running = clusterPods.filter((p) => p.status.phase === 'Running').length
    const pending = clusterPods.filter((p) => p.status.phase === 'Pending').length
    const failed = clusterPods.filter((p) => ['Failed', 'CrashLoopBackOff'].includes(p.status.phase)).length
    const total = clusterPods.length
    return {
      cluster: clusterName,
      nodes: { total: clusterNodes.length, ready, notReady: clusterNodes.length - ready },
      pods: { total, running, pending, failed },
      timestamp: new Date().toISOString(),
    }
  }
  // production：静态概览（演示）
  return {
    cluster: clusterName,
    nodes: { total: 3, ready: 3, notReady: 0 },
    pods: { total: 38, running: 36, pending: 1, failed: 1 },
    timestamp: new Date().toISOString(),
  }
}

// ── ClusterMetrics ────────────────────────────────────────
export function deriveClusterMetrics(clusterName: string) {
  if (clusterName === 'kind-test') {
    const totalCPU = mockNodes.reduce((acc, n) => acc + parseInt(n.status.capacity.cpu, 10), 0)
    const totalMemory = mockNodes.reduce((acc, n) => {
      const m = n.status.capacity.memory
      return acc + (m.endsWith('Gi') ? parseFloat(m) : 0)
    }, 0)
    return {
      clusterName,
      timestamp: new Date().toISOString(),
      nodes: { total: mockNodes.length, ready: 3, notReady: 0 },
      pods: deriveClusterStatus(clusterName).pods,
      resources: { totalCPU: `${totalCPU}`, totalMemory: `${totalMemory}Gi`, usedCPU: '4', usedMemory: '7.4Gi' },
    }
  }
  return {
    clusterName,
    timestamp: new Date().toISOString(),
    nodes: { total: 3, ready: 3, notReady: 0 },
    pods: { total: 38, running: 36, pending: 1, failed: 1 },
    resources: { totalCPU: '32', totalMemory: '128Gi', usedCPU: '12', usedMemory: '48Gi' },
  }
}

// ── ServiceEndpoints（selector × pods 派生）───────────────
export function deriveServiceEndpoints(clusterName: string, namespace: string, serviceName: string) {
  const svc = mockServices.find((s) => s.metadata.namespace === namespace && s.metadata.name === serviceName)
  if (!svc) return { serviceName, namespace, endpoints: [] }
  const selector = svc.spec.selector || {}
  const nsPods = mockPods.filter((p) => p.metadata.namespace === namespace)
  const matches = (labels: Record<string, string> = {}) =>
    Object.entries(selector).every(([k, v]) => labels[k] === v)
  const ready = nsPods.filter((p) => p.status.phase === 'Running' && !!p.status.podIP && matches(p.metadata.labels))
  const notReady = nsPods.filter((p) => p.status.phase !== 'Running' && matches(p.metadata.labels))
  const port = svc.spec.ports?.[0]
  return {
    serviceName,
    namespace,
    endpoints: [{
      addresses: ready.map((p) => ({
        ip: p.status.podIP,
        nodeName: p.spec.nodeName,
        targetRef: { kind: 'Pod', name: p.metadata.name, namespace: p.metadata.namespace },
      })),
      notReadyAddresses: notReady.map((p) => ({
        ip: p.status.podIP || '',
        nodeName: p.spec.nodeName || undefined,
        targetRef: { kind: 'Pod', name: p.metadata.name, namespace: p.metadata.namespace },
      })),
      ports: port ? [{ port: port.targetPort, protocol: port.protocol, name: port.name }] : [],
    }],
  }
}

// ── AlertStats / AlertHistory ─────────────────────────────
export function deriveAlertStats(cluster: string) {
  const records = mockAlertRecords.filter((r) => r.cluster === cluster)
  const active = records.filter((r) => !r.resolved).length
  const recent24h = records.filter((r) => Date.now() - new Date(r.createdAt).getTime() <= 24 * 3600 * MS).length
  const bySeverity: Record<string, number> = {}
  records.filter((r) => !r.resolved).forEach((r) => { bySeverity[r.severity] = (bySeverity[r.severity] || 0) + 1 })
  const byStatus: Record<string, number> = {
    active: records.filter((r) => !r.resolved && !r.acknowledged).length,
    acknowledged: records.filter((r) => !r.resolved && r.acknowledged).length,
    resolved: records.filter((r) => r.resolved).length,
  }
  return { total: records.length, active, bySeverity, byStatus, recent24h }
}

export function deriveAlertHistory(cluster: string, limit = 50) {
  return mockAlertRecords
    .filter((r) => r.cluster === cluster)
    .sort((a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime())
    .slice(0, limit)
}

// ── TenantStats ───────────────────────────────────────────
export function deriveTenantStats() {
  const usersByRole: Record<string, number> = {}
  mockTenantUsers.forEach((u) => { usersByRole[u.role] = (usersByRole[u.role] || 0) + 1 })
  const totalNamespaces = new Set(mockTenants.flatMap((t) => t.namespaces)).size
  return {
    totalTenants: mockTenants.length,
    totalUsers: mockTenantUsers.length,
    totalNamespaces,
    usersByRole,
  }
}

// ── AuditStats ────────────────────────────────────────────
export function deriveAuditStats() {
  const byEventType: Record<string, number> = {}
  const bySeverity: Record<string, number> = {}
  const byCategory: Record<string, number> = {}
  const byUser: Record<string, number> = {}
  mockAuditLogs.forEach((l) => {
    byEventType[l.eventType] = (byEventType[l.eventType] || 0) + 1
    bySeverity[l.severity] = (bySeverity[l.severity] || 0) + 1
    byCategory[l.category] = (byCategory[l.category] || 0) + 1
    byUser[l.user] = (byUser[l.user] || 0) + 1
  })
  return {
    totalLogs: mockAuditLogs.length,
    byEventType, bySeverity, byCategory, byUser,
    recent24h: mockAuditLogs.filter((l) => Date.now() - new Date(l.timestamp).getTime() <= 24 * 3600 * MS).length,
  }
}

// ── BackupSummary ─────────────────────────────────────────
export function deriveBackupSummary(cluster: string) {
  const list = mockBackups.filter((b) => b.cluster === cluster)
  const byPhase: Record<string, number> = {}
  const byMode: Record<string, number> = {}
  list.forEach((b) => {
    byPhase[b.phase] = (byPhase[b.phase] || 0) + 1
    byMode[b.spec.backupMode] = (byMode[b.spec.backupMode] || 0) + 1
  })
  return {
    total: list.length,
    byPhase,
    byMode,
    recent24h: list.filter((b) => Date.now() - new Date(b.createdAt).getTime() <= 24 * 3600 * MS).length,
  }
}

// ── MonitorHistory（kind-test）── 派生避免常量
export const mockMetricsHistoryDerived = mockMetricsHistory

// 导出派生项汇总
export const derived = {
  clusterStatus: deriveClusterStatus,
  clusterMetrics: deriveClusterMetrics,
  serviceEndpoints: deriveServiceEndpoints,
  alertStats: deriveAlertStats,
  alertHistory: deriveAlertHistory,
  tenantStats: deriveTenantStats,
  auditStats: deriveAuditStats,
  backupSummary: deriveBackupSummary,
}

// 重新导出最常用字段，便于 handlers 直接访问
export { mockAlertRules, mockAlertRecords, mockMetricsHistory as mockMetricsHistoryList, mockTenants, mockTenantUsers, mockAuditLogs, mockBackups }
export { mockPods, mockNodes, mockNodeMetrics, mockServices }