// 治理数据：租户、用户、审计日志、RBAC

import { daysAgo, hoursAgo, minutesAgo } from '../time'

// ── Tenants ───────────────────────────────────────────────
export const mockTenants = [
  {
    id: 'tenant-platform',
    cluster: 'kind-test',
    name: 'platform',
    description: '平台组：基础设施与 CI/CD 工作负载',
    namespaces: ['klaw-test', 'default', 'kube-system', 'ingress-nginx'],
    resourceQuotas: { cpu: '32', memory: '64Gi', pods: '500', services: '100', persistentVolumeClaims: '50' },
    networkPolicies: { enabled: true, defaultDeny: false },
    rbac: { enabled: true, defaultRole: 'admin' },
    createdAt: daysAgo(180),
    updatedAt: daysAgo(7),
  },
  {
    id: 'tenant-mall',
    cluster: 'kind-test',
    name: 'mall',
    description: '商城业务组：电商核心链路',
    namespaces: ['mall-prod', 'mall-staging'],
    resourceQuotas: { cpu: '64', memory: '128Gi', pods: '300', services: '80', persistentVolumeClaims: '40' },
    networkPolicies: { enabled: true, defaultDeny: true },
    rbac: { enabled: true, defaultRole: 'editor' },
    createdAt: daysAgo(45),
    updatedAt: hoursAgo(2),
  },
  {
    id: 'tenant-data',
    cluster: 'kind-test',
    name: 'data',
    description: '数据组：Spark / Kafka / Flink 工作负载',
    namespaces: ['data-platform'],
    resourceQuotas: { cpu: '128', memory: '256Gi', pods: '200', services: '40', persistentVolumeClaims: '60' },
    networkPolicies: { enabled: true, defaultDeny: true },
    rbac: { enabled: true, defaultRole: 'editor' },
    createdAt: daysAgo(30),
    updatedAt: daysAgo(1),
  },
]

// ── TenantUsers ───────────────────────────────────────────
export const mockTenantUsers = [
  { id: 'tu-001', tenantId: 'tenant-platform', username: 'admin', email: 'admin@kudig.io', role: 'admin', namespaces: ['*'], subjectKind: 'User', subjectName: 'admin', createdAt: daysAgo(180) },
  { id: 'tu-002', tenantId: 'tenant-platform', username: 'ci-bot', email: 'ci-bot@kudig.io', role: 'editor', namespaces: ['klaw-test'], subjectKind: 'ServiceAccount', subjectName: 'ci-bot', subjectNamespace: 'klaw-test', createdAt: daysAgo(120) },
  { id: 'tu-003', tenantId: 'tenant-mall', username: 'zhangwei', email: 'zhangwei@kudig.io', role: 'admin', namespaces: ['mall-prod', 'mall-staging'], subjectKind: 'User', subjectName: 'zhangwei', createdAt: daysAgo(45) },
  { id: 'tu-004', tenantId: 'tenant-mall', username: 'lina', email: 'lina@kudig.io', role: 'developer', namespaces: ['mall-prod'], subjectKind: 'User', subjectName: 'lina', createdAt: daysAgo(30) },
  { id: 'tu-005', tenantId: 'tenant-mall', username: 'chenming', email: 'chenming@kudig.io', role: 'viewer', namespaces: ['mall-staging'], subjectKind: 'User', subjectName: 'chenming', createdAt: daysAgo(20) },
  { id: 'tu-006', tenantId: 'tenant-data', username: 'wangqiang', email: 'wangqiang@kudig.io', role: 'admin', namespaces: ['data-platform'], subjectKind: 'User', subjectName: 'wangqiang', createdAt: daysAgo(30) },
  { id: 'tu-007', tenantId: 'tenant-data', username: 'liunan', email: 'liunan@kudig.io', role: 'developer', namespaces: ['data-platform'], subjectKind: 'User', subjectName: 'liunan', createdAt: daysAgo(25) },
  { id: 'tu-008', tenantId: 'tenant-mall', username: 'frontend-team', email: 'frontend@kudig.io', role: 'developer', namespaces: ['mall-prod'], subjectKind: 'Group', subjectName: 'frontend-team', createdAt: daysAgo(20) },
  ]

// ── AuditLogs ─────────────────────────────────────────────
export const mockAuditLogs = [
  // tenancy
  { id: 'audit-001', timestamp: minutesAgo(15), eventType: 'alert.acknowledge', category: 'tenancy', severity: 'info', source: 'klaw-web', user: 'wangqiang', action: 'ack alert alert-frontend-restart', resource: { kind: 'Alert', name: 'alert-frontend-restart', namespace: 'mall-prod', cluster: 'kind-test' }, result: 'success', details: { note: 'oncall 处置中' }, ipAddress: '10.0.0.21', userAgent: 'Mozilla/5.0' },
  { id: 'audit-002', timestamp: minutesAgo(38), eventType: 'alert.create', category: 'tenancy', severity: 'warning', source: 'klaw-evaluator', user: 'system', action: 'rule-memory-pressure triggered alert-mem-worker2', resource: { kind: 'Node', name: 'kind-test-worker2', cluster: 'kind-test' }, result: 'success' },
  { id: 'audit-003', timestamp: hoursAgo(2), eventType: 'deployment.scale', category: 'resource', severity: 'warning', source: 'klaw-web', user: 'zhangwei', action: 'scale Deployment mall-frontend to 3', resource: { kind: 'Deployment', name: 'mall-frontend', namespace: 'mall-prod', cluster: 'kind-test' }, result: 'success' },
  { id: 'audit-004', timestamp: hoursAgo(2), eventType: 'pod.delete', category: 'resource', severity: 'warning', source: 'kubectl', user: 'wangqiang', action: 'delete Pod mall-frontend-7d9c5f8b4-k4m9p', resource: { kind: 'Pod', name: 'mall-frontend-7d9c5f8b4-k4m9p', namespace: 'mall-prod', cluster: 'kind-test' }, result: 'success', ipAddress: '10.0.0.21' },
  { id: 'audit-005', timestamp: hoursAgo(3), eventType: 'tenant.user.create', category: 'tenancy', severity: 'info', source: 'klaw-web', user: 'zhangwei', action: 'add user liunan to tenant-data', resource: { kind: 'Tenant', name: 'tenant-data', cluster: 'kind-test' }, result: 'success' },
  { id: 'audit-006', timestamp: hoursAgo(5), eventType: 'tenant.update', category: 'tenancy', severity: 'info', source: 'klaw-web', user: 'zhangwei', action: 'update tenant-mall quotas cpu 32->64', resource: { kind: 'Tenant', name: 'tenant-mall', cluster: 'kind-test' }, result: 'success' },
  { id: 'audit-007', timestamp: hoursAgo(8), eventType: 'backup.create', category: 'backup', severity: 'info', source: 'klaw-scheduler', user: 'system', action: 'create backup klaw-daily-20260830-150000', resource: { kind: 'Backup', name: 'klaw-daily-20260830-150000', cluster: 'kind-test' }, result: 'success' },
  { id: 'audit-008', timestamp: hoursAgo(11), eventType: 'backup.create', category: 'backup', severity: 'warning', source: 'klaw-scheduler', user: 'system', action: 'create backup klaw-daily-20260830-030000 (partial failure: etcd snapshot timeout)', resource: { kind: 'Backup', name: 'klaw-daily-20260830-030000', cluster: 'kind-test' }, result: 'partial', details: { reason: 'etcd snapshot timeout after 30s' } },
  { id: 'audit-009', timestamp: daysAgo(1), eventType: 'tenant.create', category: 'tenancy', severity: 'info', source: 'klaw-web', user: 'admin', action: 'create tenant tenant-data', resource: { kind: 'Tenant', name: 'tenant-data', cluster: 'kind-test' }, result: 'success' },
  { id: 'audit-010', timestamp: daysAgo(1), eventType: 'rbac.role.grant', category: 'security', severity: 'info', source: 'klaw-web', user: 'admin', action: 'grant RoleBinding frontend-team -> mall-prod-admin', resource: { kind: 'RoleBinding', name: 'frontend-team-mall-prod-admin', namespace: 'mall-prod', cluster: 'kind-test' }, result: 'success' },
  { id: 'audit-011', timestamp: daysAgo(1), eventType: 'tenant.user.delete', category: 'tenancy', severity: 'warning', source: 'klaw-web', user: 'admin', action: 'remove user test-user', resource: { kind: 'TenantUser', name: 'test-user', cluster: 'kind-test' }, result: 'success' },
  { id: 'audit-012', timestamp: daysAgo(2), eventType: 'backup.delete', category: 'backup', severity: 'info', source: 'klaw-web', user: 'wangqiang', action: 'delete backup klaw-daily-20260825', resource: { kind: 'Backup', name: 'klaw-daily-20260825', cluster: 'kind-test' }, result: 'success' },
  { id: 'audit-013', timestamp: daysAgo(2), eventType: 'login.failed', category: 'security', severity: 'warning', source: 'klaw-web', user: 'unknown', action: 'failed login attempt username=ci-bot from 198.51.100.66', resource: { kind: 'Auth', name: 'ci-bot', cluster: 'kind-test' }, result: 'failure', ipAddress: '198.51.100.66', details: { reason: 'invalid password' } },
  { id: 'audit-014', timestamp: daysAgo(3), eventType: 'tenant.create', category: 'tenancy', severity: 'info', source: 'klaw-web', user: 'admin', action: 'create tenant tenant-mall', resource: { kind: 'Tenant', name: 'tenant-mall', cluster: 'kind-test' }, result: 'success' },
]

// ── RBAC Analysis（按租户/用户派生 + 系统数）───────────────
export const mockRbacAnalysis = {
  kindtest: {
    totalRoles: 24,
    totalClusterRoles: 73,
    totalBindings: 28,
    totalClusterBindings: 59,
    rolesByNamespace: {
      'klaw-test': ['klaw-test-admin', 'klaw-test-editor', 'klaw-test-viewer'],
      'mall-prod': ['mall-admin', 'mall-developer', 'mall-viewer', 'mall-secrets-reader'],
      'mall-staging': ['mall-admin-staging', 'mall-viewer-staging'],
      'data-platform': ['data-admin', 'data-developer'],
    },
    bindingsBySubject: {
      admin: [{ type: 'User', name: 'admin', namespace: '*', role: 'admin', roleKind: 'ClusterRole' }],
      zhangwei: [{ type: 'User', name: 'zhangwei', namespace: 'mall-prod', role: 'mall-admin', roleKind: 'Role' }],
      wangqiang: [{ type: 'User', name: 'wangqiang', namespace: 'data-platform', role: 'data-admin', roleKind: 'Role' }],
      'ci-bot': [{ type: 'ServiceAccount', name: 'ci-bot', namespace: 'klaw-test', role: 'klaw-test-editor', roleKind: 'Role' }],
    },
    bindingsByRole: {
      'mall-admin': [{ type: 'User', name: 'zhangwei', namespace: 'mall-prod', subjects: ['zhangwei'] }],
      'mall-developer': [{ type: 'User', name: 'lina', namespace: 'mall-prod', subjects: ['lina', 'frontend-team'] }],
    },
    timestamp: new Date(Date.now() - 60_000).toISOString(),
  },
  production: {
    totalRoles: 18,
    totalClusterRoles: 73,
    totalBindings: 16,
    totalClusterBindings: 59,
    rolesByNamespace: {
      'prod-mall': ['prod-admin', 'prod-developer'],
      default: ['admin'],
    },
    bindingsBySubject: {
      'prod-admin': [{ type: 'User', name: 'prod-admin', namespace: 'prod-mall', role: 'prod-admin', roleKind: 'Role' }],
    },
    bindingsByRole: {
      'prod-admin': [{ type: 'User', name: 'prod-admin', namespace: 'prod-mall', subjects: ['prod-admin'] }],
    },
    timestamp: new Date(Date.now() - 60_000).toISOString(),
  },
}