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
      setError('Failed to load tenant data')
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
        setError('Failed to load clusters')
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
      setError('Failed to create tenant')
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
      setError('Failed to create tenant user')
    }
  }

  const deleteTenant = async (tenant: Tenant) => {
    if (!confirm(`Delete tenant ${tenant.name}?`)) return
    await tenancyApi.deleteTenant(tenant.id)
    await loadData()
  }

  const deleteUser = async (user: TenantUser) => {
    if (!confirm(`Delete user ${user.username}?`)) return
    await tenancyApi.deleteUser(user.id)
    await loadData()
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Tenants</h1>
          <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
            Multi-tenant management and audit trail integrated from the guardian platform plan.
          </p>
        </div>
        <select className="input max-w-xs" value={selectedCluster} onChange={(e) => setSelectedCluster(e.target.value)}>
          <option value="">All Clusters</option>
          {clusters.map((cluster) => (
            <option key={cluster.name} value={cluster.name}>{cluster.name}</option>
          ))}
        </select>
      </div>

      {error && <div className="rounded-lg border border-red-200 bg-red-50 text-red-700 px-4 py-3">{error}</div>}

      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <div className="card p-5">
          <div className="text-sm text-gray-500">Tenants</div>
          <div className="text-2xl font-semibold mt-2">{stats?.totalTenants ?? 0}</div>
        </div>
        <div className="card p-5">
          <div className="text-sm text-gray-500">Users</div>
          <div className="text-2xl font-semibold mt-2">{stats?.totalUsers ?? 0}</div>
        </div>
        <div className="card p-5">
          <div className="text-sm text-gray-500">Namespaces</div>
          <div className="text-2xl font-semibold mt-2">{stats?.totalNamespaces ?? 0}</div>
        </div>
        <div className="card p-5">
          <div className="text-sm text-gray-500">Audit Logs 24h</div>
          <div className="text-2xl font-semibold mt-2">{auditStats?.recent24h ?? 0}</div>
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
                <span>Tenant Inventory</span>
              </h2>
              <div className="space-y-3">
                {tenants.map((tenant) => (
                  <div key={tenant.id} className="rounded-lg border border-gray-200 dark:border-gray-800 p-4">
                    <div className="flex items-start justify-between gap-4">
                      <div>
                        <div className="font-medium">{tenant.name}</div>
                        <div className="text-sm text-gray-500 mt-1">{tenant.description || 'No description'}</div>
                      </div>
                      {tenant.id !== 'default' && (
                        <button onClick={() => deleteTenant(tenant)} className="text-danger-600 hover:text-danger-700">
                          <Trash2 className="h-4 w-4" />
                        </button>
                      )}
                    </div>
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-3 mt-4 text-sm">
                      <div>
                        <div className="text-gray-500">Cluster</div>
                        <div>{tenant.cluster || 'Unassigned'}</div>
                      </div>
                      <div>
                        <div className="text-gray-500">Namespaces</div>
                        <div>{tenant.namespaces.join(', ')}</div>
                      </div>
                      <div>
                        <div className="text-gray-500">Default Role</div>
                        <div>{tenant.rbac.defaultRole}</div>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </div>

            <div className="card p-6">
              <h2 className="text-lg font-semibold flex items-center gap-2 mb-4">
                <Users className="h-5 w-5 text-primary-600" />
                <span>Tenant Users</span>
              </h2>
              <div className="space-y-3">
                {users.map((user) => (
                  <div key={user.id} className="rounded-lg border border-gray-200 dark:border-gray-800 p-4 flex items-center justify-between">
                    <div>
                      <div className="font-medium">{user.username}</div>
                      <div className="text-sm text-gray-500 mt-1">
                        {(user.subjectKind || 'User')} · {user.subjectName || user.username} · {user.role}
                      </div>
                      <div className="text-sm text-gray-500 mt-1">
                        {user.email || 'No email'} · NS: {user.namespaces?.join(', ') || 'tenant default'}
                        {user.subjectKind === 'ServiceAccount' && ` · SA NS: ${user.subjectNamespace || 'default'}`}
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
              <h2 className="text-lg font-semibold mb-4">Audit Trail</h2>
              <div className="space-y-3">
                {auditLogs.map((log) => (
                  <div key={log.id} className="rounded-lg border border-gray-200 dark:border-gray-800 p-4">
                    <div className="flex items-start justify-between gap-4">
                      <div>
                        <div className="font-medium">{log.action}</div>
                        <div className="text-sm text-gray-500 mt-1">{log.user || 'system'} · {log.result}</div>
                      </div>
                      <div className="text-sm text-gray-500 whitespace-nowrap">{formatDate(log.timestamp)}</div>
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
                <span>Create Tenant</span>
              </h2>
              <form onSubmit={createTenant} className="space-y-4">
                <select className="input" value={tenantForm.cluster || selectedCluster} onChange={(e) => setTenantForm((prev) => ({ ...prev, cluster: e.target.value }))} required>
                  <option value="">Select Cluster</option>
                  {clusters.map((cluster) => (
                    <option key={cluster.name} value={cluster.name}>{cluster.name}</option>
                  ))}
                </select>
                <input className="input" value={tenantForm.name} onChange={(e) => setTenantForm((prev) => ({ ...prev, name: e.target.value }))} placeholder="Tenant name" required />
                <input className="input" value={tenantForm.description} onChange={(e) => setTenantForm((prev) => ({ ...prev, description: e.target.value }))} placeholder="Description" />
                <textarea className="input min-h-[84px]" value={tenantForm.namespaces} onChange={(e) => setTenantForm((prev) => ({ ...prev, namespaces: e.target.value }))} placeholder="default, team-a" />
                <div className="grid grid-cols-2 gap-3">
                  <input className="input" value={tenantForm.cpu} onChange={(e) => setTenantForm((prev) => ({ ...prev, cpu: e.target.value }))} placeholder="CPU" />
                  <input className="input" value={tenantForm.memory} onChange={(e) => setTenantForm((prev) => ({ ...prev, memory: e.target.value }))} placeholder="Memory" />
                  <input className="input" value={tenantForm.pods} onChange={(e) => setTenantForm((prev) => ({ ...prev, pods: e.target.value }))} placeholder="Pods" />
                  <input className="input" value={tenantForm.services} onChange={(e) => setTenantForm((prev) => ({ ...prev, services: e.target.value }))} placeholder="Services" />
                </div>
                <input className="input" value={tenantForm.persistentVolumeClaims} onChange={(e) => setTenantForm((prev) => ({ ...prev, persistentVolumeClaims: e.target.value }))} placeholder="PVCs" />
                <select className="input" value={tenantForm.defaultRole} onChange={(e) => setTenantForm((prev) => ({ ...prev, defaultRole: e.target.value }))}>
                  <option value="view">view</option>
                  <option value="edit">edit</option>
                  <option value="admin">admin</option>
                </select>
                <button type="submit" className="btn btn-primary w-full">Create Tenant</button>
              </form>
            </div>

            <div className="card p-6">
              <h2 className="text-lg font-semibold flex items-center gap-2 mb-4">
                <Plus className="h-5 w-5 text-primary-600" />
                <span>Add Tenant User</span>
              </h2>
              <form onSubmit={createUser} className="space-y-4">
                <select className="input" value={userForm.tenantId} onChange={(e) => setUserForm((prev) => ({ ...prev, tenantId: e.target.value }))} required>
                  <option value="">Select Tenant</option>
                  {tenants.map((tenant) => (
                    <option key={tenant.id} value={tenant.id}>{tenant.name}</option>
                  ))}
                </select>
                <input className="input" value={userForm.username} onChange={(e) => setUserForm((prev) => ({ ...prev, username: e.target.value }))} placeholder="Username" required />
                <input className="input" value={userForm.email} onChange={(e) => setUserForm((prev) => ({ ...prev, email: e.target.value }))} placeholder="Email" />
                <select className="input" value={userForm.subjectKind} onChange={(e) => setUserForm((prev) => ({ ...prev, subjectKind: e.target.value as TenantSubjectKind, subjectNamespace: e.target.value === 'ServiceAccount' ? prev.subjectNamespace : '' }))}>
                  <option value="User">User</option>
                  <option value="Group">Group</option>
                  <option value="ServiceAccount">ServiceAccount</option>
                </select>
                <input
                  className="input"
                  value={userForm.subjectName}
                  onChange={(e) => setUserForm((prev) => ({ ...prev, subjectName: e.target.value }))}
                  placeholder={userForm.subjectKind === 'Group' ? 'Group name' : userForm.subjectKind === 'ServiceAccount' ? 'ServiceAccount name' : 'Subject name (optional)'}
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
                    placeholder="ServiceAccount namespace"
                  />
                )}
                <select className="input" value={userForm.role} onChange={(e) => setUserForm((prev) => ({ ...prev, role: e.target.value }))}>
                  <option value="viewer">viewer</option>
                  <option value="editor">editor</option>
                  <option value="admin">admin</option>
                </select>
                <button type="submit" className="btn btn-primary w-full">Add User</button>
              </form>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

export default TenantsPage
