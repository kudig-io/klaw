import { useState, useEffect } from 'react'
import { useSearchParams } from 'react-router-dom'
import {
  clusterApi,
  storageApi,
  type PersistentVolumeClaim,
  type PersistentVolume,
  type StorageClass,
  type StorageAnalysis,
} from '../lib/api'
import { ClusterSelector } from '../components/ClusterSelector'
import { NamespaceSelector } from '../components/NamespaceSelector'
import { RefreshButton } from '../components/RefreshButton'
import { Database, HardDrive, Boxes, PieChart } from 'lucide-react'
import { useToast } from '../contexts/ToastContext'

const ALL_NAMESPACES = '_all' // Special value for all namespaces

const ACCESS_MODE_SHORT: Record<string, string> = {
  ReadWriteOnce: 'RWO',
  ReadOnlyMany: 'ROX',
  ReadWriteMany: 'RWX',
  ReadWriteOncePod: 'RWOP',
}

function formatBytes(bytes: number): string {
  if (!Number.isFinite(bytes) || bytes <= 0) return '-'
  const units = ['B', 'Ki', 'Mi', 'Gi', 'Ti', 'Pi']
  let value = bytes
  let i = 0
  while (value >= 1024 && i < units.length - 1) {
    value /= 1024
    i++
  }
  return `${value >= 100 ? value.toFixed(0) : value.toFixed(1)} ${units[i]}`
}

export function StoragePage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const { showToast } = useToast()

  const [clusters, setClusters] = useState<Array<{ name: string }>>([])
  const [selectedCluster, setSelectedCluster] = useState(searchParams.get('cluster') || '')
  const [selectedNamespace, setSelectedNamespace] = useState(searchParams.get('namespace') || '')

  const [pvcs, setPvcs] = useState<PersistentVolumeClaim[]>([])
  const [pvs, setPvs] = useState<PersistentVolume[]>([])
  const [storageClasses, setStorageClasses] = useState<StorageClass[]>([])
  const [analysis, setAnalysis] = useState<StorageAnalysis | null>(null)
  const [isLoading, setIsLoading] = useState(false)

  useEffect(() => {
    loadClusters()
  }, [])

  useEffect(() => {
    const params: Record<string, string> = {}
    if (selectedCluster) params.cluster = selectedCluster
    if (selectedNamespace && selectedNamespace !== ALL_NAMESPACES) params.namespace = selectedNamespace
    setSearchParams(params)

    if (selectedCluster) {
      loadStorage()
    }
  }, [selectedCluster, selectedNamespace])

  async function loadClusters() {
    try {
      const response = await clusterApi.getClusters()
      setClusters(response.data)
      if (response.data.length > 0 && !selectedCluster) {
        setSelectedCluster(response.data[0].name)
      }
    } catch (error) {
      console.error('Failed to load clusters:', error)
      showToast('加载集群列表失败', 'error')
    }
  }

  async function loadStorage() {
    if (!selectedCluster) return

    setIsLoading(true)
    try {
      const ns = selectedNamespace === ALL_NAMESPACES ? '' : selectedNamespace
      const [pvcRes, pvRes, scRes, analysisRes] = await Promise.all([
        storageApi.listPVCs(selectedCluster, ns),
        storageApi.listPVs(selectedCluster),
        storageApi.listStorageClasses(selectedCluster),
        storageApi.getStorageAnalysis(),
      ])
      setPvcs(pvcRes.data || [])
      setPvs(pvRes.data || [])
      setStorageClasses(scRes.data || [])
      setAnalysis(analysisRes.data)
    } catch (error) {
      console.error('Failed to load storage resources:', error)
      showToast('加载存储资源失败', 'error')
      setPvcs([])
      setPvs([])
      setStorageClasses([])
      setAnalysis(null)
    } finally {
      setIsLoading(false)
    }
  }

  function formatAge(timestamp: string): string {
    const date = new Date(timestamp)
    const diff = Date.now() - date.getTime()
    const seconds = Math.floor(diff / 1000)
    const minutes = Math.floor(seconds / 60)
    const hours = Math.floor(minutes / 60)
    const days = Math.floor(hours / 24)
    if (days > 0) return `${days}d`
    if (hours > 0) return `${hours}h`
    if (minutes > 0) return `${minutes}m`
    return `${seconds}s`
  }

  function formatAccessModes(modes: string[]): string {
    return modes.map((m) => ACCESS_MODE_SHORT[m] || m).join(', ')
  }

  function getPvcPhaseColor(phase: string): string {
    switch (phase) {
      case 'Bound':
        return 'bg-success-100 text-success-800 dark:bg-success-900/30 dark:text-success-300'
      case 'Pending':
        return 'bg-warning-100 text-warning-800 dark:bg-warning-900/30 dark:text-warning-300'
      case 'Lost':
        return 'bg-danger-100 text-danger-800 dark:bg-danger-900/30 dark:text-danger-300'
      case 'Terminating':
        return 'bg-warning-100 text-warning-800 dark:bg-warning-900/30 dark:text-warning-300'
      default:
        return 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300'
    }
  }

  function getPvPhaseColor(phase: string): string {
    switch (phase) {
      case 'Available':
        return 'bg-info-100 text-info-800 dark:bg-info-900/30 dark:text-info-300'
      case 'Bound':
        return 'bg-success-100 text-success-800 dark:bg-success-900/30 dark:text-success-300'
      case 'Released':
        return 'bg-warning-100 text-warning-800 dark:bg-warning-900/30 dark:text-warning-300'
      case 'Failed':
        return 'bg-danger-100 text-danger-800 dark:bg-danger-900/30 dark:text-danger-300'
      default:
        return 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300'
    }
  }

  function isDefaultStorageClass(sc: StorageClass): boolean {
    return sc.metadata.annotations?.['storageclass.kubernetes.io/is-default-class'] === 'true'
  }

  function getPvSource(pv: PersistentVolume): string {
    if (pv.spec.csi) return pv.spec.csi.driver
    if (pv.spec.nfs) return `NFS · ${pv.spec.nfs.server}`
    if (pv.spec.hostPath) return `hostPath · ${pv.spec.hostPath.path}`
    if (pv.spec.local) return `local · ${pv.spec.local.path}`
    return '-'
  }

  function getClaimRef(pv: PersistentVolume): string {
    if (!pv.spec.claimRef) return '-'
    return `${pv.spec.claimRef.namespace}/${pv.spec.claimRef.name}`
  }

  const capacity = analysis?.storageCapacity
  const capacityPercent =
    capacity && capacity.totalBytes > 0 ? Math.min(100, Math.round((capacity.usedBytes / capacity.totalBytes) * 100)) : 0

  return (
    <div className="p-6">
      <div className="mb-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold text-gray-900 dark:text-white">存储管理</h1>
            <p className="text-gray-600 dark:text-gray-400 mt-1">
              管理持久卷声明（PVC）、持久卷（PV）与存储类（StorageClass），分析集群存储容量
            </p>
          </div>
          <RefreshButton onClick={loadStorage} isLoading={isLoading} />
        </div>
      </div>

      {/* Filters */}
      <div className="card p-4 mb-6">
        <div className="flex flex-wrap items-center gap-4">
          <ClusterSelector
            clusters={clusters.map((c) => c.name)}
            selected={selectedCluster}
            onSelect={setSelectedCluster}
          />

          <NamespaceSelector
            cluster={selectedCluster}
            selected={selectedNamespace || ALL_NAMESPACES}
            onSelect={setSelectedNamespace}
            showAllNamespaces={true}
          />
        </div>
        <p className="mt-2 text-xs text-gray-500 dark:text-gray-400">
          命名空间筛选仅作用于 PVC 列表；PV 与 StorageClass 为集群级资源
        </p>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
        <div className="card p-5">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-600 dark:text-gray-400">PVC 声明</p>
              <p className="text-2xl font-bold text-gray-900 dark:text-white mt-1">{pvcs.length}</p>
            </div>
            <div className="p-2.5 rounded-lg bg-info-100 dark:bg-info-900/30">
              <Database className="h-5 w-5 text-info-600 dark:text-info-300" />
            </div>
          </div>
        </div>
        <div className="card p-5">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-600 dark:text-gray-400">持久卷 PV</p>
              <p className="text-2xl font-bold text-gray-900 dark:text-white mt-1">{pvs.length}</p>
            </div>
            <div className="p-2.5 rounded-lg bg-success-100 dark:bg-success-900/30">
              <HardDrive className="h-5 w-5 text-success-600 dark:text-success-300" />
            </div>
          </div>
        </div>
        <div className="card p-5">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-600 dark:text-gray-400">存储类</p>
              <p className="text-2xl font-bold text-gray-900 dark:text-white mt-1">{storageClasses.length}</p>
            </div>
            <div className="p-2.5 rounded-lg bg-primary-100 dark:bg-primary-900/30">
              <Boxes className="h-5 w-5 text-primary-600 dark:text-primary-300" />
            </div>
          </div>
        </div>
        <div className="card p-5">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-600 dark:text-gray-400">容量使用</p>
              <p className="text-lg font-bold text-gray-900 dark:text-white mt-1">
                {capacity && capacity.totalBytes > 0
                  ? `${formatBytes(capacity.usedBytes)} / ${formatBytes(capacity.totalBytes)}`
                  : '—'}
              </p>
              {capacity && capacity.totalBytes > 0 && (
                <div className="mt-2 w-40 h-1.5 rounded-full bg-gray-200 dark:bg-gray-700 overflow-hidden">
                  <div className="h-full rounded-full bg-primary-500" style={{ width: `${capacityPercent}%` }} />
                </div>
              )}
            </div>
            <div className="p-2.5 rounded-lg bg-warning-100 dark:bg-warning-900/30">
              <PieChart className="h-5 w-5 text-warning-600 dark:text-warning-300" />
            </div>
          </div>
        </div>
      </div>

      {/* PVC Table */}
      <div className="mb-4 flex items-center gap-2">
        <Database className="h-4 w-4 text-gray-500" />
        <h2 className="text-base font-semibold text-gray-900 dark:text-white">持久卷声明（PVC）</h2>
      </div>
      <div className="card overflow-hidden mb-8">
        {isLoading ? (
          <div className="p-12 text-center">
            <div className="inline-block animate-spin rounded-full h-8 w-8 border-2 border-primary-500 border-t-transparent"></div>
            <p className="mt-4 text-gray-600 dark:text-gray-400">正在加载存储资源…</p>
          </div>
        ) : pvcs.length === 0 ? (
          <div className="p-12 text-center">
            <Database className="w-12 h-12 text-gray-300 dark:text-gray-600 mx-auto mb-4" />
            <h3 className="text-lg font-medium text-gray-900 dark:text-white mb-2">未找到 PVC</h3>
            <p className="text-gray-600 dark:text-gray-400">当前范围下暂无持久卷声明</p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm text-left">
              <thead className="bg-gray-50 dark:bg-gray-700/50 text-gray-900 dark:text-white">
                <tr>
                  <th className="px-4 py-3 font-semibold">名称</th>
                  <th className="px-4 py-3 font-semibold">命名空间</th>
                  <th className="px-4 py-3 font-semibold">状态</th>
                  <th className="px-4 py-3 font-semibold">容量</th>
                  <th className="px-4 py-3 font-semibold">访问模式</th>
                  <th className="px-4 py-3 font-semibold">存储类</th>
                  <th className="px-4 py-3 font-semibold">绑定卷</th>
                  <th className="px-4 py-3 font-semibold">存续时间</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200 dark:divide-gray-700">
                {pvcs.map((pvc) => (
                  <tr
                    key={`${pvc.metadata.namespace}-${pvc.metadata.name}`}
                    className="hover:bg-gray-50 dark:hover:bg-gray-700/30 transition-colors"
                  >
                    <td className="px-4 py-3 font-medium text-gray-900 dark:text-white">
                      {pvc.metadata.name}
                      {pvc.metadata.deletionTimestamp && (
                        <span className="ml-2 text-xs text-warning-600 dark:text-warning-400">删除中</span>
                      )}
                    </td>
                    <td className="px-4 py-3">
                      <span className="inline-flex items-center px-2 py-1 rounded text-xs font-medium bg-gray-100 text-gray-700 dark:bg-gray-700 dark:text-gray-300">
                        {pvc.metadata.namespace}
                      </span>
                    </td>
                    <td className="px-4 py-3">
                      <span className={`inline-flex items-center px-2 py-1 rounded text-xs font-medium ${getPvcPhaseColor(pvc.status.phase)}`}>
                        {pvc.status.phase}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-gray-600 dark:text-gray-400">
                      {pvc.status.capacity?.storage || pvc.spec.resources.requests.storage}
                    </td>
                    <td className="px-4 py-3 text-gray-600 dark:text-gray-400 font-mono text-xs">
                      {formatAccessModes(pvc.spec.accessModes)}
                    </td>
                    <td className="px-4 py-3 text-gray-600 dark:text-gray-400 font-mono text-xs">
                      {pvc.spec.storageClassName || '（默认）'}
                    </td>
                    <td className="px-4 py-3 text-gray-600 dark:text-gray-400 font-mono text-xs">
                      {pvc.spec.volumeName || '-'}
                    </td>
                    <td className="px-4 py-3 text-gray-600 dark:text-gray-400">
                      {formatAge(pvc.metadata.creationTimestamp)}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* PV Table */}
      <div className="mb-4 flex items-center gap-2">
        <HardDrive className="h-4 w-4 text-gray-500" />
        <h2 className="text-base font-semibold text-gray-900 dark:text-white">持久卷（PV）</h2>
      </div>
      <div className="card overflow-hidden mb-8">
        {!isLoading && pvs.length === 0 ? (
          <div className="p-12 text-center">
            <HardDrive className="w-12 h-12 text-gray-300 dark:text-gray-600 mx-auto mb-4" />
            <h3 className="text-lg font-medium text-gray-900 dark:text-white mb-2">未找到持久卷</h3>
            <p className="text-gray-600 dark:text-gray-400">当前集群下暂无持久卷</p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm text-left">
              <thead className="bg-gray-50 dark:bg-gray-700/50 text-gray-900 dark:text-white">
                <tr>
                  <th className="px-4 py-3 font-semibold">名称</th>
                  <th className="px-4 py-3 font-semibold">状态</th>
                  <th className="px-4 py-3 font-semibold">容量</th>
                  <th className="px-4 py-3 font-semibold">访问模式</th>
                  <th className="px-4 py-3 font-semibold">回收策略</th>
                  <th className="px-4 py-3 font-semibold">存储类</th>
                  <th className="px-4 py-3 font-semibold">绑定声明</th>
                  <th className="px-4 py-3 font-semibold">来源</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200 dark:divide-gray-700">
                {pvs.map((pv) => (
                  <tr
                    key={pv.metadata.name}
                    className="hover:bg-gray-50 dark:hover:bg-gray-700/30 transition-colors"
                  >
                    <td className="px-4 py-3 font-medium text-gray-900 dark:text-white">{pv.metadata.name}</td>
                    <td className="px-4 py-3">
                      <span className={`inline-flex items-center px-2 py-1 rounded text-xs font-medium ${getPvPhaseColor(pv.status.phase)}`}>
                        {pv.status.phase}
                      </span>
                      {pv.status.reason && (
                        <div className="text-xs text-gray-500 dark:text-gray-400 mt-0.5">{pv.status.reason}</div>
                      )}
                    </td>
                    <td className="px-4 py-3 text-gray-600 dark:text-gray-400">{pv.spec.capacity.storage}</td>
                    <td className="px-4 py-3 text-gray-600 dark:text-gray-400 font-mono text-xs">
                      {formatAccessModes(pv.spec.accessModes)}
                    </td>
                    <td className="px-4 py-3 text-gray-600 dark:text-gray-400">
                      {pv.spec.persistentVolumeReclaimPolicy || '-'}
                    </td>
                    <td className="px-4 py-3 text-gray-600 dark:text-gray-400 font-mono text-xs">
                      {pv.spec.storageClassName || '-'}
                    </td>
                    <td className="px-4 py-3 text-gray-600 dark:text-gray-400 font-mono text-xs">
                      {getClaimRef(pv)}
                    </td>
                    <td className="px-4 py-3 text-gray-600 dark:text-gray-400 font-mono text-xs">
                      {getPvSource(pv)}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* StorageClass Table */}
      <div className="mb-4 flex items-center gap-2">
        <Boxes className="h-4 w-4 text-gray-500" />
        <h2 className="text-base font-semibold text-gray-900 dark:text-white">存储类（StorageClass）</h2>
      </div>
      <div className="card overflow-hidden mb-8">
        {!isLoading && storageClasses.length === 0 ? (
          <div className="p-12 text-center">
            <Boxes className="w-12 h-12 text-gray-300 dark:text-gray-600 mx-auto mb-4" />
            <h3 className="text-lg font-medium text-gray-900 dark:text-white mb-2">未找到存储类</h3>
            <p className="text-gray-600 dark:text-gray-400">当前集群下暂无 StorageClass</p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm text-left">
              <thead className="bg-gray-50 dark:bg-gray-700/50 text-gray-900 dark:text-white">
                <tr>
                  <th className="px-4 py-3 font-semibold">名称</th>
                  <th className="px-4 py-3 font-semibold">Provisioner</th>
                  <th className="px-4 py-3 font-semibold">回收策略</th>
                  <th className="px-4 py-3 font-semibold">绑定模式</th>
                  <th className="px-4 py-3 font-semibold">可扩容</th>
                  <th className="px-4 py-3 font-semibold">挂载选项</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200 dark:divide-gray-700">
                {storageClasses.map((sc) => (
                  <tr
                    key={sc.metadata.name}
                    className="hover:bg-gray-50 dark:hover:bg-gray-700/30 transition-colors"
                  >
                    <td className="px-4 py-3 font-medium text-gray-900 dark:text-white">
                      {sc.metadata.name}
                      {isDefaultStorageClass(sc) && (
                        <span className="ml-2 inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-primary-100 text-primary-800 dark:bg-primary-900/30 dark:text-primary-300">
                          默认
                        </span>
                      )}
                    </td>
                    <td className="px-4 py-3 text-gray-600 dark:text-gray-400 font-mono text-xs">
                      {sc.provisioner}
                    </td>
                    <td className="px-4 py-3 text-gray-600 dark:text-gray-400">
                      {sc.reclaimPolicy || 'Delete'}
                    </td>
                    <td className="px-4 py-3 text-gray-600 dark:text-gray-400">
                      {sc.volumeBindingMode || 'Immediate'}
                    </td>
                    <td className="px-4 py-3">
                      {sc.allowVolumeExpansion ? (
                        <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-success-100 text-success-800 dark:bg-success-900/30 dark:text-success-300">
                          支持
                        </span>
                      ) : (
                        <span className="text-gray-400 dark:text-gray-600">-</span>
                      )}
                    </td>
                    <td className="px-4 py-3 text-gray-600 dark:text-gray-400 font-mono text-xs">
                      {sc.mountOptions && sc.mountOptions.length > 0 ? sc.mountOptions.join(', ') : '-'}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Storage Analysis */}
      {analysis && (
        <>
          <div className="mb-4 flex items-center gap-2">
            <PieChart className="h-4 w-4 text-gray-500" />
            <h2 className="text-base font-semibold text-gray-900 dark:text-white">存储分析</h2>
          </div>
          <div className="grid grid-cols-1 lg:grid-cols-4 gap-6">
            <div className="card p-5">
              <h3 className="text-sm font-semibold text-gray-900 dark:text-white mb-3">PV 状态分布</h3>
              <div className="space-y-2">
                {Object.entries(analysis.pvByStatus).map(([status, count]) => (
                  <div key={status} className="flex items-center justify-between text-sm">
                    <span className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ${getPvPhaseColor(status)}`}>
                      {status}
                    </span>
                    <span className="font-semibold text-gray-900 dark:text-white">{count}</span>
                  </div>
                ))}
                {Object.keys(analysis.pvByStatus).length === 0 && (
                  <p className="text-sm text-gray-500 dark:text-gray-400">暂无数据</p>
                )}
              </div>
            </div>
            <div className="card p-5">
              <h3 className="text-sm font-semibold text-gray-900 dark:text-white mb-3">PVC 状态分布</h3>
              <div className="space-y-2">
                {Object.entries(analysis.pvcByStatus).map(([status, count]) => (
                  <div key={status} className="flex items-center justify-between text-sm">
                    <span className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ${getPvcPhaseColor(status)}`}>
                      {status}
                    </span>
                    <span className="font-semibold text-gray-900 dark:text-white">{count}</span>
                  </div>
                ))}
                {Object.keys(analysis.pvcByStatus).length === 0 && (
                  <p className="text-sm text-gray-500 dark:text-gray-400">暂无数据</p>
                )}
              </div>
            </div>
            <div className="card p-5">
              <h3 className="text-sm font-semibold text-gray-900 dark:text-white mb-3">PV 按存储类</h3>
              <div className="space-y-2">
                {Object.entries(analysis.pvByStorageClass).map(([sc, count]) => (
                  <div key={sc} className="flex items-center justify-between text-sm">
                    <span className="font-mono text-xs text-gray-700 dark:text-gray-300 truncate">{sc}</span>
                    <span className="font-semibold text-gray-900 dark:text-white shrink-0 ml-2">{count}</span>
                  </div>
                ))}
                {Object.keys(analysis.pvByStorageClass).length === 0 && (
                  <p className="text-sm text-gray-500 dark:text-gray-400">暂无数据</p>
                )}
              </div>
            </div>
            <div className="card p-5">
              <h3 className="text-sm font-semibold text-gray-900 dark:text-white mb-3">存储类按 Provisioner</h3>
              <div className="space-y-2">
                {Object.entries(analysis.scByProvisioner).map(([provisioner, count]) => (
                  <div key={provisioner} className="flex items-center justify-between text-sm">
                    <span className="font-mono text-xs text-gray-700 dark:text-gray-300 truncate">{provisioner}</span>
                    <span className="font-semibold text-gray-900 dark:text-white shrink-0 ml-2">{count}</span>
                  </div>
                ))}
                {Object.keys(analysis.scByProvisioner).length === 0 && (
                  <p className="text-sm text-gray-500 dark:text-gray-400">暂无数据</p>
                )}
              </div>
            </div>
          </div>

          {capacity && capacity.totalBytes > 0 && (
            <div className="card p-5 mt-6">
              <h3 className="text-sm font-semibold text-gray-900 dark:text-white mb-3">存储容量</h3>
              <div className="h-2.5 rounded-full bg-gray-200 dark:bg-gray-700 overflow-hidden">
                <div className="h-full rounded-full bg-primary-500" style={{ width: `${capacityPercent}%` }} />
              </div>
              <div className="mt-3 grid grid-cols-3 gap-4 text-sm">
                <div>
                  <p className="text-gray-500 dark:text-gray-400">总容量</p>
                  <p className="font-semibold text-gray-900 dark:text-white">{formatBytes(capacity.totalBytes)}</p>
                </div>
                <div>
                  <p className="text-gray-500 dark:text-gray-400">已申请</p>
                  <p className="font-semibold text-gray-900 dark:text-white">{formatBytes(capacity.usedBytes)}</p>
                </div>
                <div>
                  <p className="text-gray-500 dark:text-gray-400">可用</p>
                  <p className="font-semibold text-gray-900 dark:text-white">{formatBytes(capacity.availableBytes)}</p>
                </div>
              </div>
            </div>
          )}
        </>
      )}
    </div>
  )
}
