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

const formatDuration = (start?: string, end?: string) => {
  if (!start || !end) return '-'
  const seconds = Math.max(0, Math.round((new Date(end).getTime() - new Date(start).getTime()) / 1000))
  if (seconds < 60) return `${seconds}s`
  const minutes = Math.floor(seconds / 60)
  return `${minutes}m ${seconds % 60}s`
}

const getPhaseBadgeClass = (phase: string) => {
  if (phase === 'Completed') return 'bg-success-100 text-success-700 dark:bg-success-900/30 dark:text-success-400'
  if (phase === 'Failed') return 'bg-danger-100 text-danger-700 dark:bg-danger-900/30 dark:text-danger-400'
  if (phase === 'PartiallyFailed') return 'bg-warning-100 text-warning-700 dark:bg-warning-900/30 dark:text-warning-400'
  return 'bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-400'
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
        setError('加载集群列表失败')
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
      setError('加载备份列表失败')
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
      setError(err?.response?.data?.error || '创建备份失败')
    } finally {
      setSubmitting(false)
    }
  }

  const handleDelete = async (backup: BackupItem) => {
    if (!confirm(`确定要删除备份 ${backup.name} 吗？`)) return
    try {
      await backupApi.delete(selectedCluster, backup.name)
      await loadBackups(selectedCluster)
    } catch (err) {
      console.error(err)
      setError('删除备份失败')
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold">备份恢复</h1>
          <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
            集成 etcd 备份能力，为集群提供统一的管理入口。
          </p>
        </div>
        <div className="flex items-center gap-3">
          <select value={selectedCluster} onChange={(e) => setSelectedCluster(e.target.value)} className="input w-44 shrink-0">
            <option value="">选择集群</option>
            {clusters.map((cluster: any) => (
              <option key={cluster.name} value={cluster.name}>{cluster.name}</option>
            ))}
          </select>
          <button onClick={() => selectedCluster && loadBackups(selectedCluster)} className="btn btn-secondary flex items-center gap-2 whitespace-nowrap shrink-0">
            <RefreshCw className="h-4 w-4" />
            <span>刷新</span>
          </button>
        </div>
      </div>

      {error && (
        <div className="rounded-lg border border-danger-200 dark:border-danger-800 bg-danger-50 dark:bg-danger-900/20 text-danger-700 dark:text-danger-300 px-4 py-3">
          {error}
        </div>
      )}

      <div className="grid grid-cols-2 lg:grid-cols-6 gap-4">
        <div className="card p-5">
          <div className="text-sm text-gray-500 dark:text-gray-400">备份总数</div>
          <div className="text-2xl font-semibold mt-2">{summary?.total ?? 0}</div>
        </div>
        <div className="card p-5">
          <div className="text-sm text-gray-500 dark:text-gray-400">已完成</div>
          <div className="text-2xl font-semibold mt-2">{summary?.byPhase?.Completed ?? 0}</div>
        </div>
        <div className="card p-5">
          <div className="text-sm text-gray-500 dark:text-gray-400">部分失败</div>
          <div className="text-2xl font-semibold mt-2 text-warning-600 dark:text-warning-400">{summary?.byPhase?.PartiallyFailed ?? 0}</div>
        </div>
        <div className="card p-5">
          <div className="text-sm text-gray-500 dark:text-gray-400">失败</div>
          <div className="text-2xl font-semibold mt-2 text-danger-600 dark:text-danger-400">{summary?.byPhase?.Failed ?? 0}</div>
        </div>
        <div className="card p-5">
          <div className="text-sm text-gray-500 dark:text-gray-400">增量备份</div>
          <div className="text-2xl font-semibold mt-2">{summary?.byMode?.Incremental ?? 0}</div>
        </div>
        <div className="card p-5">
          <div className="text-sm text-gray-500 dark:text-gray-400">近 24 小时</div>
          <div className="text-2xl font-semibold mt-2">{summary?.recent24h ?? 0}</div>
        </div>
      </div>

      <div className="grid grid-cols-1 xl:grid-cols-3 gap-6">
        <div className="xl:col-span-2 card p-6">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-lg font-semibold flex items-center gap-2">
              <DatabaseBackup className="h-5 w-5 text-primary-600" />
              <span>备份列表</span>
            </h2>
          </div>

          {loading ? (
            <div className="flex items-center justify-center py-12">
              <Loader2 className="h-6 w-6 animate-spin text-primary-600" />
            </div>
          ) : backups.length === 0 ? (
            <div className="text-sm text-gray-500 dark:text-gray-400 py-8 text-center">
              当前集群还没有备份记录。
            </div>
          ) : (
            <div className="space-y-3">
              {backups.map((backup) => (
                <div key={backup.name} className="rounded-lg border border-gray-200 dark:border-gray-800 p-4">
                  <div className="flex items-start justify-between gap-4">
                    <div>
                      <div className="flex items-center gap-2 flex-wrap">
                        <span className="font-medium">{backup.name}</span>
                        <span className={`px-1.5 py-0.5 rounded text-xs font-medium ${getPhaseBadgeClass(backup.phase)}`}>
                          {backup.phase}
                        </span>
                      </div>
                      <div className="text-sm text-gray-500 dark:text-gray-400 mt-1">
                        {backup.spec.backupMode} · {formatBytes(backup.snapshotSize)} · 耗时 {formatDuration(backup.startTime, backup.completionTime)}
                      </div>
                    </div>
                    <button
                      onClick={() => handleDelete(backup)}
                      className="text-danger-600 hover:text-danger-700"
                      aria-label={`删除备份 ${backup.name}`}
                      title="删除备份"
                    >
                      <Trash2 className="h-4 w-4" />
                    </button>
                  </div>

                  <div className="grid grid-cols-1 md:grid-cols-2 gap-3 mt-4 text-sm">
                    <div>
                      <div className="text-gray-500 dark:text-gray-400">创建时间</div>
                      <div>{formatDate(backup.createdAt)}</div>
                    </div>
                    <div>
                      <div className="text-gray-500 dark:text-gray-400">etcd 版本</div>
                      <div>{backup.etcdRevision}</div>
                    </div>
                    <div className="md:col-span-2">
                      <div className="text-gray-500 dark:text-gray-400">快照位置</div>
                      <div className="break-all">{backup.snapshotLocation}</div>
                    </div>
                    <div>
                      <div className="text-gray-500 dark:text-gray-400">存储位置</div>
                      <div>
                        {backup.spec.storageLocation.provider} / {backup.spec.storageLocation.bucket} / {backup.spec.storageLocation.region}
                      </div>
                    </div>
                    <div>
                      <div className="text-gray-500 dark:text-gray-400">凭证引用（Secret）</div>
                      <div className="font-mono">{backup.spec.storageLocation.credentialsSecret}</div>
                    </div>
                    <div>
                      <div className="text-gray-500 dark:text-gray-400">数据校验</div>
                      <div>
                        {backup.validationResult?.message || '未启用'}
                        {backup.validationResult?.hash && (
                          <span className="ml-2 font-mono text-xs text-gray-500 dark:text-gray-400">hash: {backup.validationResult.hash}</span>
                        )}
                      </div>
                    </div>
                    <div>
                      <div className="text-gray-500 dark:text-gray-400">校验配置</div>
                      <div>
                        {backup.spec.validation?.enabled ? '启用' : '停用'}
                        {backup.spec.validation?.consistencyCheck ? ' · 一致性检查 启用' : ''}
                      </div>
                    </div>
                    <div>
                      <div className="text-gray-500 dark:text-gray-400">etcd 端点</div>
                      <div>{backup.spec.etcdEndpoints?.length ?? 0} 个</div>
                    </div>
                    {backup.message && (
                      <div className="md:col-span-2 text-danger-600 dark:text-danger-400">{backup.message}</div>
                    )}
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>

        <div className="card p-6">
          <h2 className="text-lg font-semibold mb-4 flex items-center gap-2">
            <Plus className="h-5 w-5 text-primary-600" />
            <span>创建备份</span>
          </h2>

          <form onSubmit={handleCreate} className="space-y-4">
            <div>
              <label className="block text-sm font-medium mb-1">名称</label>
              <input
                className="input"
                value={request.name}
                onChange={(e) => updateRequest({ name: e.target.value })}
                placeholder="daily-backup-20260523"
                required
              />
            </div>

            <div>
              <label className="block text-sm font-medium mb-1">备份模式</label>
              <select
                className="input"
                value={request.backupMode}
                onChange={(e) => updateRequest({ backupMode: e.target.value as CreateBackupRequest['backupMode'] })}
              >
                <option value="Full">全量（Full）</option>
                <option value="Incremental">增量（Incremental）</option>
              </select>
            </div>

            <div>
              <label className="block text-sm font-medium mb-1">存储提供商</label>
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
              <label className="block text-sm font-medium mb-1">存储桶</label>
              <input
                className="input"
                value={request.storageLocation.bucket}
                onChange={(e) => updateRequest({ storageLocation: { ...request.storageLocation, bucket: e.target.value } })}
                required
              />
            </div>

            <div>
              <label className="block text-sm font-medium mb-1">地域</label>
              <input
                className="input"
                value={request.storageLocation.region}
                onChange={(e) => updateRequest({ storageLocation: { ...request.storageLocation, region: e.target.value } })}
                required
              />
            </div>

            <div>
              <label className="block text-sm font-medium mb-1">路径前缀</label>
              <input
                className="input"
                value={request.storageLocation.prefix}
                onChange={(e) => updateRequest({ storageLocation: { ...request.storageLocation, prefix: e.target.value } })}
                placeholder="etcd-backups"
              />
            </div>

            <div>
              <label className="block text-sm font-medium mb-1">凭证引用（Secret）</label>
              <input
                className="input"
                value={request.storageLocation.credentialsSecret}
                onChange={(e) => updateRequest({ storageLocation: { ...request.storageLocation, credentialsSecret: e.target.value } })}
                required
              />
            </div>

            <div>
              <label className="block text-sm font-medium mb-1">etcd 地址</label>
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
              <span>启用数据校验</span>
            </label>

            <label className="flex items-center gap-2 text-sm">
              <input
                type="checkbox"
                checked={request.validation.consistencyCheck}
                onChange={(e) => updateRequest({
                  validation: { ...request.validation, consistencyCheck: e.target.checked },
                })}
              />
              <span>启用一致性检查</span>
            </label>

            <button type="submit" className="btn btn-primary w-full flex items-center justify-center gap-2" disabled={submitting || !selectedCluster}>
              {submitting ? <Loader2 className="h-4 w-4 animate-spin" /> : <Plus className="h-4 w-4" />}
              <span>{submitting ? '创建中…' : '创建备份'}</span>
            </button>
          </form>
        </div>
      </div>
    </div>
  )
}

export default BackupsPage
