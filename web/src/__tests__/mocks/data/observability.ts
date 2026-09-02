// 可观测数据：Events、AlertRules、AlertRecords、MonitorHistory
//
// 故事线时间轴：recent 24h 内 worker2 内存压力触发 critical 告警，frontend
// Pending 触发 error 告警，frontend-...-z8x3c CrashLoopBackOff 触发 critical 告警；
// 历史告警记录事件缓解轨迹。

import { minutesAgo, hoursAgo, daysAgo } from '../time'

// ── Events ─────────────────────────────────────────────────
export const mockEvents = [
  // 默认 namespace（节点事件）
  { metadata: { name: 'kind-test-worker2.18a2c', namespace: 'default' }, type: 'Normal', reason: 'NodeHasInsufficientMemory', message: 'Node kind-test-worker2 status is now: MemoryPressure=true', lastTimestamp: minutesAgo(40) },
  { metadata: { name: 'kind-test-worker2.18a2d', namespace: 'default' }, type: 'Warning', reason: 'NodeNotReady', message: 'Node kind-test-worker2 not ready for 12s (kubelet down)', lastTimestamp: daysAgo(3) },

  // klaw-test 常规事件
  { metadata: { name: 'nginx-6b66fbbd46-abc12.1', namespace: 'klaw-test' }, type: 'Normal', reason: 'Scheduled', message: 'Successfully assigned nginx-6b66fbbd46-abc12 to kind-test-worker', lastTimestamp: daysAgo(2) },
  { metadata: { name: 'nginx-6b66fbbd46-def34.1', namespace: 'klaw-test' }, type: 'Normal', reason: 'Pulled', message: 'Successfully pulled image "nginx:alpine"', lastTimestamp: daysAgo(2) },
  { metadata: { name: 'frontend-58cb7f74c8-xyz78.1', namespace: 'klaw-test' }, type: 'Normal', reason: 'Scheduled', message: 'Successfully assigned frontend-58cb7f74c8-xyz78 to kind-test-worker', lastTimestamp: daysAgo(1) },
  { metadata: { name: 'frontend-58cb7f74c8-pqr90.1', namespace: 'klaw-test' }, type: 'Warning', reason: 'FailedScheduling', message: '0/3 nodes are available: 2 node(s) had untolerated taint, 1 node(s) insufficient memory.', lastTimestamp: hoursAgo(3) },

  // mall-prod 故事线事件
  { metadata: { name: 'mall-frontend-7d9c5f8b4-q7w2e.1', namespace: 'mall-prod' }, type: 'Warning', reason: 'FailedScheduling', message: '0/3 nodes are available: 2 Insufficient memory, 1 untolerated taint node.kubernetes.io/memory-pressure.', lastTimestamp: hoursAgo(2) },
  { metadata: { name: 'mall-frontend-7d9c5f8b4-q7w2e.2', namespace: 'mall-prod' }, type: 'Warning', reason: 'FailedScheduling', message: 'pod has unbound immediate PersistentVolumeClaims (repeated 4 times)', lastTimestamp: hoursAgo(2) },
  { metadata: { name: 'mall-frontend-7d9c5f8b4-z8x3c.1', namespace: 'mall-prod' }, type: 'Warning', reason: 'BackOff', message: 'Back-off restarting failed container mall-frontend', lastTimestamp: minutesAgo(45) },
  { metadata: { name: 'mall-frontend-7d9c5f8b4-z8x3c.2', namespace: 'mall-prod' }, type: 'Warning', reason: 'OOMKilled', message: 'Container mall-frontend exceeded memory limits (1024Mi), killed', lastTimestamp: minutesAgo(48) },
  { metadata: { name: 'mall-frontend-7d9c5f8b4-z8x3c.3', namespace: 'mall-prod' }, type: 'Normal', reason: 'Pulled', message: 'Successfully pulled image "registry.local/mall/frontend:v2.4.1"', lastTimestamp: hoursAgo(2) },
  { metadata: { name: 'mall-frontend-7d9c5f8b4-z8x3c.4', namespace: 'mall-prod' }, type: 'Warning', reason: 'Failed', message: 'Error: ImagePullBackOff (previous attempt: exit code 134)', lastTimestamp: hoursAgo(2) },
  { metadata: { name: 'mall-frontend.1', namespace: 'mall-prod' }, type: 'Normal', reason: 'ScalingReplicaSet', message: 'Scaled up replica set mall-frontend-7d9c5f8b4 from 1 to 3', lastTimestamp: hoursAgo(2) },
  { metadata: { name: 'order-service-6c9d8e7f5-d3e4f.1', namespace: 'mall-prod' }, type: 'Normal', reason: 'Scheduled', message: 'Successfully assigned order-service-6c9d8e7f5-d3e4f to kind-test-worker2', lastTimestamp: daysAgo(4) },

  // mall-staging 常规
  { metadata: { name: 'cart-service-stg.1', namespace: 'mall-staging' }, type: 'Warning', reason: 'Unhealthy', message: 'Liveness probe failed: HTTP 500', lastTimestamp: daysAgo(1) },

  // data-platform
  { metadata: { name: 'spark-driver.1', namespace: 'data-platform' }, type: 'Normal', reason: 'Created', message: 'Created container spark-driver', lastTimestamp: daysAgo(1) },
  { metadata: { name: 'kafka-broker.1', namespace: 'data-platform' }, type: 'Warning', reason: 'NetworkNotReady', message: 'network is not ready: container runtime network not ready', lastTimestamp: daysAgo(8) },
]

// ── AlertRules ────────────────────────────────────────────
export const mockAlertRules = [
  { id: 'rule-memory-pressure', cluster: 'kind-test', name: 'NodeMemoryPressure', description: '节点内存压力超过 90%', enabled: true, severity: 'critical', condition: { type: 'node', field: 'memoryPressure', operator: '==', threshold: true, timeWindow: '5m' }, actions: ['notify:oncall', 'webhook:incident'], createdAt: daysAgo(60), updatedAt: daysAgo(7) },
  { id: 'rule-pod-restart-storm', cluster: 'kind-test', name: 'PodRestartStorm', description: 'Pod 5 分钟内重启超过 5 次', enabled: true, severity: 'critical', condition: { type: 'pod', field: 'restartCount', operator: '>', threshold: 5, timeWindow: '5m' }, actions: ['notify:oncall'], createdAt: daysAgo(60), updatedAt: daysAgo(14) },
  { id: 'rule-pod-pending', cluster: 'kind-test', name: 'PodPendingTooLong', description: 'Pod 调度 Pending 超过 10 分钟', enabled: true, severity: 'warning', condition: { type: 'pod', field: 'phase', operator: '==', threshold: 'Pending', timeWindow: '10m' }, actions: ['notify:oncall'], createdAt: daysAgo(60), updatedAt: daysAgo(3) },
  { id: 'rule-deployment-unavailable', cluster: 'kind-test', name: 'DeploymentUnavailable', description: 'Deployment 可用副本数低于预期', enabled: true, severity: 'error', condition: { type: 'deployment', field: 'availableReplicas', operator: '<', threshold: 1, timeWindow: '2m' }, actions: ['notify:oncall'], createdAt: daysAgo(60), updatedAt: daysAgo(2) },
  { id: 'rule-backup-stale', cluster: 'kind-test', name: 'BackupStale', description: 'etcd 备份超过 26 小时未更新', enabled: true, severity: 'warning', condition: { type: 'backup', field: 'ageHours', operator: '>', threshold: 26, timeWindow: '1h' }, actions: ['notify:oncall'], createdAt: daysAgo(60), updatedAt: daysAgo(7) },
  { id: 'rule-pvc-high', cluster: 'kind-test', name: 'PVCUsageHigh', description: 'PVC 使用率超过 85%', enabled: true, severity: 'warning', condition: { type: 'pvc', field: 'usagePercent', operator: '>', threshold: 85, timeWindow: '5m' }, actions: [], createdAt: daysAgo(45), updatedAt: daysAgo(10) },
]

// ── AlertRecords ──────────────────────────────────────────
export const mockAlertRecords = [
  // ACTIVE — 当前事故
  { id: 'alert-mem-worker2', cluster: 'kind-test', ruleId: 'rule-memory-pressure', ruleName: 'NodeMemoryPressure', ruleType: 'node', resourceKind: 'Node', resourceName: 'kind-test-worker2', severity: 'critical', value: '93%', threshold: '90%', operator: '>', message: 'kind-test-worker2 内存使用 93%（7.4Gi/8Gi），触发 MemoryPressure', acknowledged: false, resolved: false, createdAt: minutesAgo(38) },
  { id: 'alert-frontend-pending', cluster: 'kind-test', ruleId: 'rule-pod-pending', ruleName: 'PodPendingTooLong', ruleType: 'pod', resourceKind: 'Pod', resourceName: 'mall-frontend-7d9c5f8b4-q7w2e', namespace: 'mall-prod', severity: 'error', value: '25m', threshold: '10m', operator: '>', message: 'Pod mall-frontend-7d9c5f8b4-q7w2e 调度 Pending 25 分钟：2 节点内存不足', acknowledged: false, resolved: false, createdAt: minutesAgo(25) },
  { id: 'alert-frontend-restart', cluster: 'kind-test', ruleId: 'rule-pod-restart-storm', ruleName: 'PodRestartStorm', ruleType: 'pod', resourceKind: 'Pod', resourceName: 'mall-frontend-7d9c5f8b4-z8x3c', namespace: 'mall-prod', severity: 'critical', value: '8', threshold: '5', operator: '>', message: 'mall-frontend-7d9c5f8b4-z8x3c 30 分钟内重启 8 次（OOMKilled → CrashLoopBackOff）', acknowledged: true, acknowledgedAt: minutesAgo(15), resolved: false, createdAt: minutesAgo(18) },
  { id: 'alert-frontend-unavail', cluster: 'kind-test', ruleId: 'rule-deployment-unavailable', ruleName: 'DeploymentUnavailable', ruleType: 'deployment', resourceKind: 'Deployment', resourceName: 'mall-frontend', namespace: 'mall-prod', severity: 'error', value: '1/3', threshold: 'availableReplicas >= replicas', operator: '<', message: 'Deployment mall-frontend 仅 1/3 副本可用（ProgressProgress延期）', acknowledged: false, resolved: false, createdAt: minutesAgo(12) },
  { id: 'alert-spark-pending', cluster: 'kind-test', ruleId: 'rule-deployment-unavailable', ruleName: 'DeploymentUnavailable', ruleType: 'deployment', resourceKind: 'Deployment', resourceName: 'spark-driver', namespace: 'data-platform', severity: 'warning', value: '0/2', threshold: 'availableReplicas >= 1', operator: '<', message: 'Deployment spark-driver 0/2 副本可用（节点调度资源不足）', acknowledged: true, acknowledgedAt: hoursAgo(2), resolved: false, createdAt: hoursAgo(3) },

  // 历史 — 已 ACK 或已 RESOLVED
  { id: 'alert-mem-worker2-hist', cluster: 'kind-test', ruleId: 'rule-memory-pressure', ruleName: 'NodeMemoryPressure', ruleType: 'node', resourceKind: 'Node', resourceName: 'kind-test-worker2', severity: 'critical', value: '91%', threshold: '90%', operator: '>', message: 'kind-test-worker2 内存使用 91%', acknowledged: true, acknowledgedAt: daysAgo(2), resolved: true, resolvedAt: daysAgo(1), createdAt: daysAgo(2) },
  { id: 'alert-pvc-redis-hist', cluster: 'kind-test', ruleId: 'rule-pvc-high', ruleName: 'PVCUsageHigh', ruleType: 'pvc', resourceKind: 'PVC', resourceName: 'mall-redis-data', namespace: 'mall-prod', severity: 'warning', value: '88%', threshold: '85%', operator: '>', message: 'PVC mall-redis-data 使用率 88%', acknowledged: true, resolved: true, resolvedAt: hoursAgo(20), createdAt: hoursAgo(28) },
  { id: 'alert-pending-stg', cluster: 'kind-test', ruleId: 'rule-pod-pending', ruleName: 'PodPendingTooLong', ruleType: 'pod', resourceKind: 'Pod', resourceName: 'cart-service-stg-29', namespace: 'mall-staging', severity: 'warning', value: '14m', threshold: '10m', operator: '>', message: 'Pod cart-service-stg-29 调度 Pending（ImagePullBackOff）', acknowledged: true, resolved: true, resolvedAt: daysAgo(1), createdAt: daysAgo(1) },
  { id: 'alert-restart-prod', cluster: 'production', ruleId: 'rule-pod-restart-storm', ruleName: 'PodRestartStorm', ruleType: 'pod', resourceKind: 'Pod', resourceName: 'mall-api-7c8d9e6f-x9y2z', namespace: 'prod-mall', severity: 'critical', value: '6', threshold: '5', operator: '>', message: 'mall-api-7c8d9e6f-x9y2z 30 分钟内重启 6 次', acknowledged: true, resolved: true, resolvedAt: daysAgo(2), createdAt: daysAgo(2) },
  { id: 'alert-notready-worker', cluster: 'kind-test', ruleId: 'rule-memory-pressure', ruleName: 'NodeNotReady', ruleType: 'node', resourceKind: 'Node', resourceName: 'kind-test-worker', severity: 'critical', value: 'True', threshold: 'Ready=False', operator: '==', message: 'Node kind-test-worker NotReady (kubelet 短暂掉线)', acknowledged: true, resolved: true, resolvedAt: daysAgo(3), createdAt: daysAgo(3) },
  { id: 'alert-deploy-stg', cluster: 'kind-test', ruleId: 'rule-deployment-unavailable', ruleName: 'DeploymentUnavailable', ruleType: 'deployment', resourceKind: 'Deployment', resourceName: 'cart-service-stg', namespace: 'mall-staging', severity: 'warning', value: '0/1', threshold: '1', operator: '<', message: 'Deployment cart-service-stg 0/1 副本可用', acknowledged: true, resolved: true, resolvedAt: daysAgo(1), createdAt: daysAgo(1) },
  { id: 'alert-backup-stale', cluster: 'kind-test', ruleId: 'rule-backup-stale', ruleName: 'BackupStale', ruleType: 'backup', resourceKind: 'Backup', resourceName: 'klaw-daily-20260829', severity: 'warning', value: '28h', threshold: '26h', operator: '>', message: '最近一次成功备份已超过 28 小时', acknowledged: true, resolved: true, resolvedAt: daysAgo(2), createdAt: daysAgo(2) },
  { id: 'alert-mem-prev', cluster: 'kind-test', ruleId: 'rule-memory-pressure', ruleName: 'NodeMemoryPressure', ruleType: 'node', resourceKind: 'Node', resourceName: 'kind-test-worker2', severity: 'warning', value: '88%', threshold: '85%', operator: '>', message: 'kind-test-worker2 内存使用 88%（首次触发）', acknowledged: true, resolved: true, resolvedAt: daysAgo(2), createdAt: daysAgo(2) },
]

// ── MonitorHistory ────────────────────────────────────────
// 24 个数据点（每 5 分钟一次），以此刻倒推；故事时间线：1h 前事故发生，pending 0→4、failed 0→1。
// 用 hoursAgo + offset 派生；为避免循环，data 文件生成时即时写入时间戳。
function ts(agoMs: number) { return new Date(Date.now() - agoMs).toISOString() }

export const mockMetricsHistory = (() => {
  const points: Array<{ clusterName: string; timestamp: string; nodes: { total: number; ready: number; notReady: number }; pods: { total: number; running: number; pending: number; failed: number } }> = []
  const STEP = 5 * 60 * 1000 // 5 min
  const STEPS = 24 // 2h
  for (let i = 0; i < STEPS; i++) {
    const ago = (STEPS - i) * STEP
    const isAfterIncident = ago <= 60 * 60 * 1000 // 最近 1 小时
    // 事故前 14/13/0/0（含 cart-stg；事故后 Running 10、Pend 4、Fail 1）
    points.push({
      clusterName: 'kind-test',
      timestamp: ts(ago),
      nodes: { total: 3, ready: 3, notReady: 0 },
      pods: isAfterIncident
        ? { total: 15, running: 10, pending: 4, failed: 1 }
        : { total: 12, running: 12, pending: 0, failed: 0 },
    })
  }
  return points
})()

// 旧字段，保留向后兼容旧 handler（monitoringApi.getAlerts 仍用旧形状）
export const mockAlerts = mockAlertRecords
  .filter((r) => !r.resolved)
  .map((r) => ({
    id: r.id,
    cluster: r.cluster,
    type: r.ruleType,
    level: r.severity,
    message: r.message,
    createdAt: r.createdAt,
    resolved: r.resolved,
  }))