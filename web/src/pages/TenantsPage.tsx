import { useEffect, useState } from 'react'
import { auditApi, clusterApi, tenancyApi, type AuditLog, type AuditStatistics, type Cluster, type Tenant, type TenantStatistics, type TenantUser } from '../lib/api'
import { formatDate } from '../lib/utils'
import { Loader2, Plus, Shield, Trash2, Users } from 'lucide-react'

const defaultTenant = {
  cluster: '',
  name: '',
  description: '',
  namespaces: 'default',
  cpu: '10',
  memory: '20Gi',
  pods: '100',
  services: '50',
  persistentVolumeClaims: '20',
  defaultRole: 'view',
}

type TenantSubjectKind = NonNullable<TenantUser['subjectKind']>

type TenantUserForm = {
  tenantId: string
  username: string
  email: string
  role: string
  namespaces: string
  subjectKind: TenantSubjectKind
  subjectName: string
  subjectNamespace: string
}

const defaultUser: TenantUserForm = {
  tenantId: '',
  username: '',
  email: '',
  role: 'viewer',
  namespaces: '',
  subjectKind: 'User',
  subjectName: '',
  subjectNamespace: '',
}

const TenantsPage = () => {
  const [tenants, setTenants] = useState<Tenant[]>([])
  const [users, setUsers] = useState<TenantUser[]>([])
  const [stats, setStats] = useState<TenantStatistics | null>(null)
  const [auditLogs, setAuditLogs] = useState<AuditLog[]>([])
  const [auditStats, setAuditStats] = useState<AuditStatistics | null>(null)
  const [tenantForm, setTenantForm] = useState(defaultTenant)
  const [userForm, setUserForm] = useState(defaultUser)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [clusters, setClusters] = useState<Cluster[]>([])
  const [selectedCluster, setSelectedCluster] = useState('')

  const selectedTenant = tenants.find((tenant) => tenant.id === userForm.tenantId)

  const loadData = async (cluster = selectedCluster) => {
    setLoading(true)
    setError(null)
    try {
      const [tenantsRes, usersRes, statsRes, logsRes, auditStatsRes] = await Promise.all([
        tenancyApi.listTenants(cluster ? { cluster } : undefined),
        tenancyApi.listUsers(),
        tenancyApi.stats(),
        auditApi.listLogs({ category: 'tenancy', limit: 20 }),
        auditApi.stats(),
      ])
      setTenants(tenantsRes.data)
      setUsers(usersRes.data)
      setStats(statsRes.data)
      setAuditLogs(logsRes.data)
      setAuditStats(auditStatsRes.data)
    } catch (err) {
      console.error(err)
      setError('加载多租户数据失败')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    const loadInitial = async () => {
      try {
        const response = await clusterApi.getClusters()
        setClusters(response.data)
        const nextCluster = response.data[0]?.name || ''
        setSelectedCluster(nextCluster)
        await loadData(nextCluster)
      } catch (err) {
        console.error(err)
        setError('加载集群列表失败')
      }
    }
    void loadInitial()
  }, [])

  useEffect(() => {
    if (selectedCluster) {
      void loadData(selectedCluster)
    }
  }, [selectedCluster])

  const createTenant = async (e: React.FormEvent) => {
    e.preventDefault()
    try {
      await tenancyApi.createTenant({
        cluster: tenantForm.cluster || selectedCluster,
        name: tenantForm.name,
        description: tenantForm.description,
        namespaces: tenantForm.namespaces.split(',').map((item) => item.trim()).filter(Boolean),
        resourceQuotas: {
          cpu: tenantForm.cpu,
          memory: tenantForm.memory,
          pods: tenantForm.pods,
          services: tenantForm.services,
          persistentVolumeClaims: tenantForm.persistentVolumeClaims,
        },
        networkPolicies: {
          enabled: true,
          defaultDeny: true,
        },
        rbac: {
          enabled: true,
          defaultRole: tenantForm.defaultRole,
        },
      })
      setTenantForm(defaultTenant)
      await loadData()
    } catch (err) {
      console.error(err)
      setError('创建租户失败')
    }
  }

  const createUser = async (e: React.FormEvent) => {
    e.preventDefault()
    try {
      await tenancyApi.createUser({
        tenantId: userForm.tenantId,
        username: userForm.username,
        email: userForm.email,
        role: userForm.role,
        namespaces: userForm.namespaces.split(',').map((item) => item.trim()).filter(Boolean),
        subjectKind: userForm.subjectKind,
        subjectName: userForm.subjectName.trim() || undefined,
        subjectNamespace: userForm.subjectKind === 'ServiceAccount' ? userForm.subjectNamespace.trim() || undefined : undefined,
      })
      setUserForm(defaultUser)
      await loadData()
    } catch (err) {
      console.error(err)
      setError('创建租户用户失败')
    }
  }

  const deleteTenant = async (tenant: Tenant) => {
    if (!confirm(`确定要删除租户 ${tenant.name} 吗？`)) return
    await tenancyApi.deleteTenant(tenant.id)
    await loadData()
  }

  const deleteUser = async (user: TenantUser) => {
    if (!confirm(`确定要删除用户 ${user.username} 吗？`)) return
    await tenancyApi.deleteUser(user.id)
    await loadData()
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">多租户</h1>
          <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
            提供租户隔离、配额与审计日志的统一管理。
          </p>
        </div>
        <select className="input max-w-xs" value={selectedCluster} onChange={(e) => setSelectedCluster(e.target.value)}>
          <option value="">全部集群</option>
          {clusters.map((cluster) => (
            <option key={cluster.name} value={cluster.name}>{cluster.name}</option>
          ))}
        </select>
      </div>

      {error && <div className="rounded-lg border border-danger-200 dark:border-danger-800 bg-danger-50 dark:bg-danger-900/20 text-danger-700 dark:text-danger-300 px-4 py-3">{error}</div>}

      <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-5 gap-4">
        <div className="card p-5">
          <div className="text-sm text-gray-500 dark:text-gray-400">租户数</div>
          <div className="text-2xl font-semibold mt-2">{stats?.totalTenants ?? 0}</div>
        </div>
        <div className="card p-5">
          <div className="text-sm text-gray-500 dark:text-gray-400">用户数</div>
          <div className="text-2xl font-semibold mt-2">{stats?.totalUsers ?? 0}</div>
        </div>
        <div className="card p-5">
          <div className="text-sm text-gray-500 dark:text-gray-400">命名空间数</div>
          <div className="text-2xl font-semibold mt-2">{stats?.totalNamespaces ?? 0}</div>
        </div>
        <div className="card p-5">
          <div className="text-sm text-gray-500 dark:text-gray-400">近 24 小时审计</div>
          <div className="text-2xl font-semibold mt-2">{auditStats?.recent24h ?? 0}</div>
        </div>
        <div className="card p-5">
          <div className="text-sm text-gray-500 dark:text-gray-400">审计日志总数</div>
          <div className="text-2xl font-semibold mt-2">{auditStats?.totalLogs ?? 0}</div>
        </div>
      </div>

      {loading ? (
        <div className="flex items-center justify-center py-12">
          <Loader2 className="h-6 w-6 animate-spin text-primary-600" />
        </div>
      ) : (
        <div className="grid grid-cols-1 xl:grid-cols-3 gap-6">
          <div className="xl:col-span-2 space-y-6">
            <div className="card p-6">
              <h2 className="text-lg font-semibold flex items-center gap-2 mb-4">
                <Shield className="h-5 w-5 text-primary-600" />
                <span>租户列表</span>
              </h2>
              <div className="space-y-3">
                {tenants.map((tenant) => (
                  <div key={tenant.id} className="rounded-lg border border-gray-200 dark:border-gray-800 p-4">
                    <div className="flex items-start justify-between gap-4">
                      <div>
                        <div className="flex items-center gap-2 flex-wrap">
                          <span className="font-medium">{tenant.name}</span>
                          <span className={`px-1.5 py-0.5 rounded text-xs font-medium ${tenant.networkPolicies.enabled ? 'bg-success-100 text-success-700 dark:bg-success-900/30 dark:text-success-400' : 'bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-400'}`}>
                            网络策略 {tenant.networkPolicies.enabled ? '启用' : '停用'}
                          </span>
                          {tenant.networkPolicies.defaultDeny && (
                            <span className="px-1.5 py-0.5 rounded text-xs font-medium bg-primary-100 text-primary-700 dark:bg-primary-900/30 dark:text-primary-400">默认拒绝</span>
                          )}
                          <span className="px-1.5 py-0.5 rounded text-xs font-medium bg-info-100 text-info-700 dark:bg-info-900/30 dark:text-info-400">RBAC {tenant.rbac.defaultRole}</span>
                        </div>
                        <div className="text-sm text-gray-500 dark:text-gray-400 mt-1">{tenant.description || '暂无描述'}</div>
                      </div>
                      {tenant.id !== 'default' && (
                        <button onClick={() => deleteTenant(tenant)} className="text-danger-600 hover:text-danger-700">
                          <Trash2 className="h-4 w-4" />
                        </button>
                      )}
                    </div>
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-3 mt-4 text-sm">
                      <div>
                        <div className="text-gray-500 dark:text-gray-400">所属集群</div>
                        <div>{tenant.cluster || '未分配'}</div>
                      </div>
                      <div>
                        <div className="text-gray-500 dark:text-gray-400">命名空间</div>
                        <div>{tenant.namespaces.join(', ')}</div>
                      </div>
                    </div>
                    <div className="grid grid-cols-2 md:grid-cols-5 gap-3 mt-3 text-sm">
                      <div>
                        <div className="text-gray-500 dark:text-gray-400">CPU 配额</div>
                        <div className="font-medium">{tenant.resourceQuotas.cpu} 核</div>
                      </div>
                      <div>
                        <div className="text-gray-500 dark:text-gray-400">内存配额</div>
                        <div className="font-medium">{tenant.resourceQuotas.memory}</div>
                      </div>
                      <div>
                        <div className="text-gray-500 dark:text-gray-400">容器组上限</div>
                        <div className="font-medium">{tenant.resourceQuotas.pods}</div>
                      </div>
                      <div>
                        <div className="text-gray-500 dark:text-gray-400">服务上限</div>
                        <div className="font-medium">{tenant.resourceQuotas.services}</div>
                      </div>
                      <div>
                        <div className="text-gray-500 dark:text-gray-400">PVC 上限</div>
                        <div className="font-medium">{tenant.resourceQuotas.persistentVolumeClaims}</div>
                      </div>
                    </div>
                    <div className="text-xs text-gray-500 dark:text-gray-400 mt-3">
                      创建 {formatDate(tenant.createdAt)} · 更新 {formatDate(tenant.updatedAt)}
                    </div>
                  </div>
                ))}
              </div>
            </div>

            <div className="card p-6">
              <h2 className="text-lg font-semibold flex items-center gap-2 mb-4">
                <Users className="h-5 w-5 text-primary-600" />
                <span>租户用户</span>
              </h2>
              {stats?.usersByRole && Object.keys(stats.usersByRole).length > 0 && (
                <div className="flex items-center gap-2 flex-wrap mb-4">
                  {Object.entries(stats.usersByRole).map(([role, count]) => (
                    <span key={role} className="px-1.5 py-0.5 rounded bg-gray-100 dark:bg-gray-800 text-xs">
                      {role} × {count}
                    </span>
                  ))}
                </div>
              )}
              <div className="space-y-3">
                {users.map((user) => (
                  <div key={user.id} className="rounded-lg border border-gray-200 dark:border-gray-800 p-4 flex items-center justify-between">
                    <div>
                      <div className="font-medium">{user.username}</div>
                      <div className="text-sm text-gray-500 dark:text-gray-400 mt-1">
                        {(user.subjectKind || 'User')} · {user.subjectName || user.username} · {user.role}
                      </div>
                      <div className="text-sm text-gray-500 dark:text-gray-400 mt-1">
                        {user.email || '未填写邮箱'} · 命名空间：{user.namespaces?.join(', ') || '租户默认'}
                        {user.subjectKind === 'ServiceAccount' && ` · SA 命名空间：${user.subjectNamespace || 'default'}`}
                      </div>
                    </div>
                    <button onClick={() => deleteUser(user)} className="text-danger-600 hover:text-danger-700">
                      <Trash2 className="h-4 w-4" />
                    </button>
                  </div>
                ))}
              </div>
            </div>

            <div className="card p-6">
              <h2 className="text-lg font-semibold mb-4">审计日志</h2>
              <div className="space-y-3">
                {auditLogs.map((log) => (
                  <div key={log.id} className="rounded-lg border border-gray-200 dark:border-gray-800 p-4">
                    <div className="flex items-start justify-between gap-4">
                      <div>
                        <div className="flex items-center gap-2 flex-wrap">
                          <span className="font-mono text-xs text-gray-500 dark:text-gray-400">{log.eventType}</span>
                          <span className={`px-1.5 py-0.5 rounded text-xs font-medium ${log.severity === 'warning' ? 'bg-warning-100 text-warning-700 dark:bg-warning-900/30 dark:text-warning-400' : log.severity === 'error' ? 'bg-danger-100 text-danger-700 dark:bg-danger-900/30 dark:text-danger-400' : 'bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-400'}`}>{log.severity}</span>
                          <span className={`px-1.5 py-0.5 rounded text-xs font-medium ${log.result === 'success' ? 'bg-success-100 text-success-700 dark:bg-success-900/30 dark:text-success-400' : log.result === 'failure' ? 'bg-danger-100 text-danger-700 dark:bg-danger-900/30 dark:text-danger-400' : 'bg-warning-100 text-warning-700 dark:bg-warning-900/30 dark:text-warning-400'}`}>{log.result}</span>
                        </div>
                        <div className="font-medium mt-1">{log.action}</div>
                        <div className="text-sm text-gray-500 dark:text-gray-400 mt-1">
                          {log.user || '系统'} · 来源 {log.source}
                          {log.resource?.kind && ` · ${log.resource.kind} ${log.resource.namespace ? `${log.resource.namespace}/` : ''}${log.resource.name}`}
                          {log.ipAddress && ` · IP ${log.ipAddress}`}
                        </div>
                        {log.details && Object.keys(log.details).length > 0 && (
                          <div className="text-xs text-gray-500 dark:text-gray-400 mt-1 font-mono break-all">
                            {Object.entries(log.details).map(([k, v]) => `${k}: ${String(v)}`).join(' · ')}
                          </div>
                        )}
                      </div>
                      <div className="text-sm text-gray-500 dark:text-gray-400 whitespace-nowrap">{formatDate(log.timestamp)}</div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>

          <div className="space-y-6">
            <div className="card p-6">
              <h2 className="text-lg font-semibold flex items-center gap-2 mb-4">
                <Plus className="h-5 w-5 text-primary-600" />
                <span>创建租户</span>
              </h2>
              <form onSubmit={createTenant} className="space-y-4">
                <select className="input" value={tenantForm.cluster || selectedCluster} onChange={(e) => setTenantForm((prev) => ({ ...prev, cluster: e.target.value }))} required>
                  <option value="">选择集群</option>
                  {clusters.map((cluster) => (
                    <option key={cluster.name} value={cluster.name}>{cluster.name}</option>
                  ))}
                </select>
                <input className="input" value={tenantForm.name} onChange={(e) => setTenantForm((prev) => ({ ...prev, name: e.target.value }))} placeholder="租户名称" required />
                <input className="input" value={tenantForm.description} onChange={(e) => setTenantForm((prev) => ({ ...prev, description: e.target.value }))} placeholder="描述" />
                <textarea className="input min-h-[84px]" value={tenantForm.namespaces} onChange={(e) => setTenantForm((prev) => ({ ...prev, namespaces: e.target.value }))} placeholder="default, team-a" />
                <div className="grid grid-cols-2 gap-3">
                  <input className="input" value={tenantForm.cpu} onChange={(e) => setTenantForm((prev) => ({ ...prev, cpu: e.target.value }))} placeholder="CPU" />
                  <input className="input" value={tenantForm.memory} onChange={(e) => setTenantForm((prev) => ({ ...prev, memory: e.target.value }))} placeholder="内存" />
                  <input className="input" value={tenantForm.pods} onChange={(e) => setTenantForm((prev) => ({ ...prev, pods: e.target.value }))} placeholder="容器组" />
                  <input className="input" value={tenantForm.services} onChange={(e) => setTenantForm((prev) => ({ ...prev, services: e.target.value }))} placeholder="服务" />
                </div>
                <input className="input" value={tenantForm.persistentVolumeClaims} onChange={(e) => setTenantForm((prev) => ({ ...prev, persistentVolumeClaims: e.target.value }))} placeholder="存储卷声明（PVC）" />
                <select className="input" value={tenantForm.defaultRole} onChange={(e) => setTenantForm((prev) => ({ ...prev, defaultRole: e.target.value }))}>
                  <option value="view">view</option>
                  <option value="edit">edit</option>
                  <option value="admin">admin</option>
                </select>
                <button type="submit" className="btn btn-primary w-full">创建租户</button>
              </form>
            </div>

            <div className="card p-6">
              <h2 className="text-lg font-semibold flex items-center gap-2 mb-4">
                <Plus className="h-5 w-5 text-primary-600" />
                <span>添加租户用户</span>
              </h2>
              <form onSubmit={createUser} className="space-y-4">
                <select className="input" value={userForm.tenantId} onChange={(e) => setUserForm((prev) => ({ ...prev, tenantId: e.target.value }))} required>
                  <option value="">选择租户</option>
                  {tenants.map((tenant) => (
                    <option key={tenant.id} value={tenant.id}>{tenant.name}</option>
                  ))}
                </select>
                <input className="input" value={userForm.username} onChange={(e) => setUserForm((prev) => ({ ...prev, username: e.target.value }))} placeholder="用户名" required />
                <input className="input" value={userForm.email} onChange={(e) => setUserForm((prev) => ({ ...prev, email: e.target.value }))} placeholder="邮箱" />
                <select className="input" value={userForm.subjectKind} onChange={(e) => setUserForm((prev) => ({ ...prev, subjectKind: e.target.value as TenantSubjectKind, subjectNamespace: e.target.value === 'ServiceAccount' ? prev.subjectNamespace : '' }))}>
                  <option value="User">User</option>
                  <option value="Group">Group</option>
                  <option value="ServiceAccount">ServiceAccount</option>
                </select>
                <input
                  className="input"
                  value={userForm.subjectName}
                  onChange={(e) => setUserForm((prev) => ({ ...prev, subjectName: e.target.value }))}
                  placeholder={userForm.subjectKind === 'Group' ? '用户组名称' : userForm.subjectKind === 'ServiceAccount' ? 'ServiceAccount 名称' : '主体名称（可选）'}
                />
                <textarea
                  className="input min-h-[72px]"
                  value={userForm.namespaces}
                  onChange={(e) => setUserForm((prev) => ({ ...prev, namespaces: e.target.value }))}
                  placeholder={selectedTenant ? selectedTenant.namespaces.join(', ') : 'default, team-a'}
                />
                {userForm.subjectKind === 'ServiceAccount' && (
                  <input
                    className="input"
                    value={userForm.subjectNamespace}
                    onChange={(e) => setUserForm((prev) => ({ ...prev, subjectNamespace: e.target.value }))}
                    placeholder="ServiceAccount 所在命名空间"
                  />
                )}
                <select className="input" value={userForm.role} onChange={(e) => setUserForm((prev) => ({ ...prev, role: e.target.value }))}>
                  <option value="viewer">viewer</option>
                  <option value="editor">editor</option>
                  <option value="admin">admin</option>
                </select>
                <button type="submit" className="btn btn-primary w-full">添加用户</button>
              </form>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

export default TenantsPage
