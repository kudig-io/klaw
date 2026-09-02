// 内存可变 store：handler 中的写操作（删 pod、扩缩容、ack/resolve 告警等）直接改这里。
// 模块加载时从基础数据深拷贝一份。test 用例之间的 mutation 不会跨文件级 module reset
// （不同测试文件 vitest 各自 module 隔离；同一文件内测试顺序也可能影响——这是 mock 而非真实存储，
// 用户刷新浏览器可重置。）

import { mockPods, mockAlertRecords, deploymentsStore, mockAuditLogs, mockTenants, mockTenantUsers, backupsStore } from './data/index'

export interface AlertRecordMutable {
  id: string
  cluster: string
  ruleId: string
  ruleName: string
  ruleType: string
  resourceKind: string
  resourceName: string
  namespace?: string
  severity: 'info' | 'warning' | 'error' | 'critical'
  value: string | number | boolean
  threshold: string | number | boolean
  operator: string
  message: string
  acknowledged: boolean
  resolved: boolean
  createdAt: string
  acknowledgedAt?: string
  resolvedAt?: string
}

export const store = {
  pods: structuredClone(mockPods),
  deployments: structuredClone(deploymentsStore),
  alertRecords: structuredClone(mockAlertRecords) as AlertRecordMutable[],
  auditLogs: structuredClone(mockAuditLogs),
  tenants: structuredClone(mockTenants),
  users: structuredClone(mockTenantUsers),
  backups: structuredClone(backupsStore),
}

let alertSeq = store.alertRecords.length
let backupSeq = store.backups.length
let tenantSeq = store.tenants.length
let userSeq = store.users.length
let auditSeq = store.auditLogs.length

const now = () => new Date().toISOString()

export const id = (prefix: string) => `${prefix}-${Date.now().toString(36)}-${(Math.random() * 1e6 | 0).toString(36)}`

export function appendAudit(entry: { eventType: string; category: string; severity: string; user: string; action: string; resource: Record<string, string>; result: string; cluster?: string; namespace?: string }) {
  auditSeq += 1
  store.auditLogs.unshift({
    id: `audit-${Date.now().toString(36)}-${auditSeq}`,
    timestamp: now(),
    eventType: entry.eventType,
    category: entry.category,
    severity: entry.severity,
    source: 'klaw-web',
    user: entry.user,
    action: entry.action,
    resource: { ...entry.resource, ...(entry.cluster ? { cluster: entry.cluster } : {}), ...(entry.namespace ? { namespace: entry.namespace } : {}) },
    result: entry.result,
  } as never)
  if (store.auditLogs.length > 200) store.auditLogs.length = 200
}

export function nextAlertId() {
  alertSeq += 1
  return `alert-new-${Date.now().toString(36)}-${alertSeq}`
}
export function nextBackupName() {
  const d = new Date()
  const stamp = `${d.getFullYear()}${String(d.getMonth() + 1).padStart(2, '0')}${String(d.getDate()).padStart(2, '0')}-${String(d.getHours()).padStart(2, '0')}${String(d.getMinutes()).padStart(2, '0')}${String(d.getSeconds()).padStart(2, '0')}`
  return `klaw-manual-${stamp}`
}
export function nextTenantId() {
  tenantSeq += 1
  return `tenant-new-${Date.now().toString(36)}-${tenantSeq}`
}
export function nextUserId() {
  userSeq += 1
  return `tu-new-${Date.now().toString(36)}-${userSeq}`
}
export { now }