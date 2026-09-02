// 租户 + 用户 + 审计 handlers

import { http, HttpResponse } from 'msw'
import { derive } from '../data/index'
import { store, appendAudit, nextTenantId, nextUserId, now } from '../store'

export const tenancyHandlers = [
  // Tenants
  http.get('/api/v1/tenants', ({ request }) => {
    const url = new URL(request.url)
    const cluster = url.searchParams.get('cluster')
    const name = url.searchParams.get('name')
    const ns = url.searchParams.get('namespace')
    let list = store.tenants
    if (cluster) list = list.filter((t) => t.cluster === cluster)
    if (name) list = list.filter((t) => t.name.includes(name))
    if (ns) list = list.filter((t) => t.namespaces.includes(ns))
    return HttpResponse.json(list)
  }),
  http.get('/api/v1/tenants/:id', ({ params }) => {
    const t = store.tenants.find((t) => t.id === params.id)
    return t ? HttpResponse.json(t) : new HttpResponse(null, { status: 404 })
  }),
  http.post('/api/v1/tenants', async ({ request }) => {
    const body = await request.json() as any
    const tenant = {
      id: nextTenantId(),
      cluster: body.cluster,
      name: body.name,
      description: body.description,
      namespaces: body.namespaces || [],
      resourceQuotas: body.resourceQuotas || { cpu: '8', memory: '16Gi', pods: '100', services: '40', persistentVolumeClaims: '20' },
      networkPolicies: body.networkPolicies || { enabled: true, defaultDeny: false },
      rbac: body.rbac || { enabled: true, defaultRole: 'view' },
      createdAt: now(),
      updatedAt: now(),
    }
    store.tenants.unshift(tenant)
    appendAudit({
      eventType: 'tenant.create', category: 'tenancy', severity: 'info', user: 'oncall',
      action: `create tenant ${tenant.name}`, resource: { kind: 'Tenant', name: tenant.name },
      cluster: tenant.cluster, result: 'success',
    })
    return HttpResponse.json(tenant, { status: 201 })
  }),
  http.put('/api/v1/tenants/:id', ({ params }) => {
    const t = store.tenants.find((t) => t.id === params.id)
    if (!t) return new HttpResponse(null, { status: 404 })
    t.updatedAt = now()
    appendAudit({
      eventType: 'tenant.update', category: 'tenancy', severity: 'info', user: 'oncall',
      action: `update tenant ${t.name}`, resource: { kind: 'Tenant', name: t.name },
      cluster: t.cluster, result: 'success',
    })
    return HttpResponse.json(t)
  }),
  http.delete('/api/v1/tenants/:id', ({ params }) => {
    const idx = store.tenants.findIndex((t) => t.id === params.id)
    if (idx < 0) return new HttpResponse(null, { status: 404 })
    const [removed] = store.tenants.splice(idx, 1)
    appendAudit({
      eventType: 'tenant.delete', category: 'tenancy', severity: 'warning', user: 'oncall',
      action: `delete tenant ${removed.name}`, resource: { kind: 'Tenant', name: removed.name },
      cluster: removed.cluster, result: 'success',
    })
    return HttpResponse.json({ message: `Tenant ${removed.name} deleted` })
  }),
  http.get('/api/v1/tenants/stats', () => HttpResponse.json(derive.tenantStats())),

  // Users
  http.get('/api/v1/tenant-users', ({ request }) => {
    const url = new URL(request.url)
    const tenantId = url.searchParams.get('tenantId')
    const role = url.searchParams.get('role')
    let list = store.users
    if (tenantId) list = list.filter((u) => u.tenantId === tenantId)
    if (role) list = list.filter((u) => u.role === role)
    return HttpResponse.json(list)
  }),
  http.post('/api/v1/tenant-users', async ({ request }) => {
    const body = await request.json() as any
    const user = {
      id: nextUserId(),
      tenantId: body.tenantId,
      username: body.username,
      email: body.email,
      role: body.role,
      namespaces: body.namespaces || [],
      subjectKind: body.subjectKind || 'User',
      subjectName: body.subjectName || body.username,
      subjectNamespace: body.subjectKind === 'ServiceAccount' ? body.subjectNamespace : undefined,
      createdAt: now(),
    }
    store.users.unshift(user)
    appendAudit({
      eventType: 'tenant.user.create', category: 'tenancy', severity: 'info', user: 'oncall',
      action: `add user ${user.username}`, resource: { kind: 'TenantUser', name: user.username },
      result: 'success',
    })
    return HttpResponse.json(user, { status: 201 })
  }),
  http.delete('/api/v1/tenant-users/:id', ({ params }) => {
    const idx = store.users.findIndex((u) => u.id === params.id)
    if (idx < 0) return new HttpResponse(null, { status: 404 })
    const [removed] = store.users.splice(idx, 1)
    appendAudit({
      eventType: 'tenant.user.delete', category: 'tenancy', severity: 'warning', user: 'oncall',
      action: `remove user ${removed.username}`, resource: { kind: 'TenantUser', name: removed.username },
      result: 'success',
    })
    return HttpResponse.json({ message: `User ${removed.username} removed` })
  }),
]

export const auditHandlers = [
  http.get('/api/v1/audit/logs', ({ request }) => {
    const url = new URL(request.url)
    const category = url.searchParams.get('category')
    const eventType = url.searchParams.get('eventType')
    const severity = url.searchParams.get('severity')
    const user = url.searchParams.get('user')
    const limit = parseInt(url.searchParams.get('limit') || '50', 10)
    let list = [...store.auditLogs]
    if (category) list = list.filter((l) => l.category === category)
    if (eventType) list = list.filter((l) => l.eventType === eventType)
    if (severity) list = list.filter((l) => l.severity === severity)
    if (user) list = list.filter((l) => l.user === user)
    return HttpResponse.json(list.slice(0, limit))
  }),
  http.get('/api/v1/audit/stats', () => HttpResponse.json(derive.auditStats())),
]