import React, { useEffect, useMemo, useState } from 'react'
import { backupApi, clusterApi, type BackupItem, type BackupSummary, type CreateBackupRequest } from '../lib/api'
import { formatDate } from '../lib/utils'
import { DatabaseBackup, Loader2, Plus, RefreshCw, Trash2 } from 'lucide-react'

const defaultRequest: CreateBackupRequest = {
  name: '',
  backupMode: 'Full',
  etcdEndpoints: [],
  storageLocation: {
    provider: 'S3',
    bucket: '',
    region: '',
    prefix: '',
    endpoint: '',
    credentialsSecret: '',
  },
  validation: {
    enabled: true,
    consistencyCheck: true,
  },
}

const formatBytes = (value: number) => {
  if (!value) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let size = value
  let unitIndex = 0
  while (size >= 1024 && unitIndex < units.length - 1) {
    size /= 1024
    unitIndex++
  }
  return `${size.toFixed(size >= 10 ? 0 : 1)} ${units[unitIndex]}`
}

const BackupsPage: React.FC = () => {
  const [clusters, setClusters] = useState<any[]>([])
  const [selectedCluster, setSelectedCluster] = useState('')
  const [backups, setBackups] = useState<BackupItem[]>([])
  const [summary, setSummary] = useState<BackupSummary | null>(null)
  const [request, setRequest] = useState<CreateBackupRequest>(defaultRequest)
  const [loading, setLoading] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const loadClusters = async () => {
      try {
        const response = await clusterApi.getClusters()
        setClusters(response.data)
        if (response.data.length > 0) {
          setSelectedCluster(response.data[0].name)
        }
      } catch (err) {
        console.error(err)
        setError('Failed to load clusters')
      }
    }
    loadClusters()
  }, [])

  useEffect(() => {
    if (selectedCluster) {
      void loadBackups(selectedCluster)
    }
  }, [selectedCluster])

  const loadBackups = async (cluster: string) => {
    setLoading(true)
    setError(null)
    try {
      const [listResponse, summaryResponse] = await Promise.all([
        backupApi.list(cluster),
        backupApi.summary(cluster),
      ])
      setBackups(listResponse.data)
      setSummary(summaryResponse.data)
    } catch (err) {
      console.error(err)
      setError('Failed to load backups')
    } finally {
      setLoading(false)
    }
  }

  const filteredEndpoints = useMemo(
    () => request.etcdEndpoints?.join(', ') ?? '',
    [request.etcdEndpoints]
  )

  const updateRequest = (patch: Partial<CreateBackupRequest>) => {
    setRequest((prev) => ({ ...prev, ...patch }))
  }

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!selectedCluster) return

    setSubmitting(true)
    setError(null)
    try {
      await backupApi.create(selectedCluster, request)
      setRequest(defaultRequest)
      await loadBackups(selectedCluster)
    } catch (err: any) {
      console.error(err)
      setError(err?.response?.data?.error || 'Failed to create backup')
    } finally {
      setSubmitting(false)
    }
  }

  const handleDelete = async (backup: BackupItem) => {
    if (!confirm(`Delete backup ${backup.name}?`)) return
    try {
      await backupApi.delete(selectedCluster, backup.name)
      await loadBackups(selectedCluster)
    } catch (err) {
      console.error(err)
      setError('Failed to delete backup')
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold">Backups</h1>
          <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
            Integrates the etcd backup domain into Klaw with a unified management surface.
          </p>
        </div>
        <div className="flex items-center gap-3">
          <select value={selectedCluster} onChange={(e) => setSelectedCluster(e.target.value)} className="input">
            <option value="">Select Cluster</option>
            {clusters.map((cluster: any) => (
              <option key={cluster.name} value={cluster.name}>{cluster.name}</option>
            ))}
          </select>
          <button onClick={() => selectedCluster && loadBackups(selectedCluster)} className="btn btn-secondary flex items-center gap-2">
            <RefreshCw className="h-4 w-4" />
            <span>Refresh</span>
          </button>
        </div>
      </div>

      {error && (
        <div className="rounded-lg border border-red-200 bg-red-50 text-red-700 px-4 py-3">
          {error}
        </div>
      )}

      <div className="grid grid-cols-1 lg:grid-cols-4 gap-4">
        <div className="card p-5">
          <div className="text-sm text-gray-500">Total Backups</div>
          <div className="text-2xl font-semibold mt-2">{summary?.total ?? 0}</div>
        </div>
        <div className="card p-5">
          <div className="text-sm text-gray-500">Completed</div>
          <div className="text-2xl font-semibold mt-2">{summary?.byPhase?.Completed ?? 0}</div>
        </div>
        <div className="card p-5">
          <div className="text-sm text-gray-500">Incremental</div>
          <div className="text-2xl font-semibold mt-2">{summary?.byMode?.Incremental ?? 0}</div>
        </div>
        <div className="card p-5">
          <div className="text-sm text-gray-500">Recent 24h</div>
          <div className="text-2xl font-semibold mt-2">{summary?.recent24h ?? 0}</div>
        </div>
      </div>

      <div className="grid grid-cols-1 xl:grid-cols-3 gap-6">
        <div className="xl:col-span-2 card p-6">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-lg font-semibold flex items-center gap-2">
              <DatabaseBackup className="h-5 w-5 text-primary-600" />
              <span>Backup Inventory</span>
            </h2>
          </div>

          {loading ? (
            <div className="flex items-center justify-center py-12">
              <Loader2 className="h-6 w-6 animate-spin text-primary-600" />
            </div>
          ) : backups.length === 0 ? (
            <div className="text-sm text-gray-500 dark:text-gray-400 py-8 text-center">
              No backups created for this cluster yet.
            </div>
          ) : (
            <div className="space-y-3">
              {backups.map((backup) => (
                <div key={backup.name} className="rounded-lg border border-gray-200 dark:border-gray-800 p-4">
                  <div className="flex items-start justify-between gap-4">
                    <div>
                      <div className="font-medium">{backup.name}</div>
                      <div className="text-sm text-gray-500 mt-1">
                        {backup.spec.backupMode} · {backup.phase} · {formatBytes(backup.snapshotSize)}
                      </div>
                    </div>
                    <button
                      onClick={() => handleDelete(backup)}
                      className="text-danger-600 hover:text-danger-700"
                      title="Delete backup"
                    >
                      <Trash2 className="h-4 w-4" />
                    </button>
                  </div>

                  <div className="grid grid-cols-1 md:grid-cols-2 gap-3 mt-4 text-sm">
                    <div>
                      <div className="text-gray-500">Created</div>
                      <div>{formatDate(backup.createdAt)}</div>
                    </div>
                    <div>
                      <div className="text-gray-500">Revision</div>
                      <div>{backup.etcdRevision}</div>
                    </div>
                    <div className="md:col-span-2">
                      <div className="text-gray-500">Snapshot Location</div>
                      <div className="break-all">{backup.snapshotLocation}</div>
                    </div>
                    <div>
                      <div className="text-gray-500">Storage</div>
                      <div>{backup.spec.storageLocation.provider} / {backup.spec.storageLocation.bucket}</div>
                    </div>
                    <div>
                      <div className="text-gray-500">Validation</div>
                      <div>{backup.validationResult?.message || 'Disabled'}</div>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>

        <div className="card p-6">
          <h2 className="text-lg font-semibold mb-4 flex items-center gap-2">
            <Plus className="h-5 w-5 text-primary-600" />
            <span>Create Backup</span>
          </h2>

          <form onSubmit={handleCreate} className="space-y-4">
            <div>
              <label className="block text-sm font-medium mb-1">Name</label>
              <input
                className="input"
                value={request.name}
                onChange={(e) => updateRequest({ name: e.target.value })}
                placeholder="daily-backup-20260523"
                required
              />
            </div>

            <div>
              <label className="block text-sm font-medium mb-1">Mode</label>
              <select
                className="input"
                value={request.backupMode}
                onChange={(e) => updateRequest({ backupMode: e.target.value as CreateBackupRequest['backupMode'] })}
              >
                <option value="Full">Full</option>
                <option value="Incremental">Incremental</option>
              </select>
            </div>

            <div>
              <label className="block text-sm font-medium mb-1">Storage Provider</label>
              <select
                className="input"
                value={request.storageLocation.provider}
                onChange={(e) => updateRequest({
                  storageLocation: { ...request.storageLocation, provider: e.target.value as CreateBackupRequest['storageLocation']['provider'] },
                })}
              >
                <option value="S3">S3</option>
                <option value="OSS">OSS</option>
                <option value="GCS">GCS</option>
                <option value="Azure">Azure</option>
              </select>
            </div>

            <div>
              <label className="block text-sm font-medium mb-1">Bucket</label>
              <input
                className="input"
                value={request.storageLocation.bucket}
                onChange={(e) => updateRequest({ storageLocation: { ...request.storageLocation, bucket: e.target.value } })}
                required
              />
            </div>

            <div>
              <label className="block text-sm font-medium mb-1">Region</label>
              <input
                className="input"
                value={request.storageLocation.region}
                onChange={(e) => updateRequest({ storageLocation: { ...request.storageLocation, region: e.target.value } })}
                required
              />
            </div>

            <div>
              <label className="block text-sm font-medium mb-1">Prefix</label>
              <input
                className="input"
                value={request.storageLocation.prefix}
                onChange={(e) => updateRequest({ storageLocation: { ...request.storageLocation, prefix: e.target.value } })}
                placeholder="etcd-backups"
              />
            </div>

            <div>
              <label className="block text-sm font-medium mb-1">Credentials Secret</label>
              <input
                className="input"
                value={request.storageLocation.credentialsSecret}
                onChange={(e) => updateRequest({ storageLocation: { ...request.storageLocation, credentialsSecret: e.target.value } })}
                required
              />
            </div>

            <div>
              <label className="block text-sm font-medium mb-1">Etcd Endpoints</label>
              <textarea
                className="input min-h-[92px]"
                value={filteredEndpoints}
                onChange={(e) => updateRequest({
                  etcdEndpoints: e.target.value
                    .split(',')
                    .map((item) => item.trim())
                    .filter(Boolean),
                })}
                placeholder="https://etcd-0:2379, https://etcd-1:2379"
              />
            </div>

            <label className="flex items-center gap-2 text-sm">
              <input
                type="checkbox"
                checked={request.validation.enabled}
                onChange={(e) => updateRequest({
                  validation: { ...request.validation, enabled: e.target.checked },
                })}
              />
              <span>Enable Validation</span>
            </label>

            <label className="flex items-center gap-2 text-sm">
              <input
                type="checkbox"
                checked={request.validation.consistencyCheck}
                onChange={(e) => updateRequest({
                  validation: { ...request.validation, consistencyCheck: e.target.checked },
                })}
              />
              <span>Enable Consistency Check</span>
            </label>

            <button type="submit" className="btn btn-primary w-full flex items-center justify-center gap-2" disabled={submitting || !selectedCluster}>
              {submitting ? <Loader2 className="h-4 w-4 animate-spin" /> : <Plus className="h-4 w-4" />}
              <span>{submitting ? 'Creating...' : 'Create Backup'}</span>
            </button>
          </form>
        </div>
      </div>
    </div>
  )
}

export default BackupsPage
