// Mock 数据聚合导出
//
// 1) 保留旧的导出名（mockClusters / mockClusterStatus / mockNamespaces / mockPods / mockNodes /
//    mockDeployments / mockEvents / mockNodeMetrics / mockMetricsHistory / mockAlerts）以保证向后兼容；
// 2) 暴露新数据：mockServices / mockAlertRules / mockAlertRecords / mockTenants / mockTenantUsers /
//    mockAuditLogs / mockBackups / mockSosStatus / mockRbacAnalysis。

export { mockClusters, mockNamespaces } from './clusters'
export { mockNodes, mockNodeMetrics } from './nodes'
export { mockPods, mockDeployments, deploymentsStore } from './workloads'
export { mockServices } from './services'
export { mockEvents, mockAlertRules, mockAlertRecords, mockMetricsHistory, mockAlerts } from './observability'
export { mockTenants, mockTenantUsers, mockAuditLogs, mockRbacAnalysis } from './governance'
export { mockBackups, backupsStore } from './backups'
export { mockSosStatus, mockSosFallback } from './sos'
export { getLogsForPod, analyzeLogs } from './logs'
export { buildDiagResponse, buildDiagIssues } from './diag'

// 兼容旧导出：mockClusterStatus 现已用派生函数 deriveClusterStatus('kind-test')
// 但单测里通过 API 响应拿到 status，不需要直接引用 mockClusterStatus。
// 仍然导出一个静态版本以兼容潜在旧 import。
import { derived } from './derive'
import type { ClusterStatus } from '../../../lib/api'
export const mockClusterStatus: ClusterStatus = derived.clusterStatus('kind-test')
export const mockMetricsHistoryDerived = derived.clusterMetrics

// 派生工具（handlers 使用）
export const derive = derived