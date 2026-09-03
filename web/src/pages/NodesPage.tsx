import React, { useState, useEffect } from 'react'
import { clusterApi, nodeApi } from '../lib/api'
import { cn, formatDate, getStatusColor } from '../lib/utils'
import { RefreshCw, Loader2, Server, Cpu, HardDrive } from 'lucide-react'

const NodesPage: React.FC = () => {
  const [clusters, setClusters] = useState<any[]>([])
  const [selectedCluster, setSelectedCluster] = useState<string>('')
  const [nodes, setNodes] = useState<any[]>([])
  const [metrics, setMetrics] = useState<Record<string, any>>({})
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    fetchClusters()
  }, [])

  const fetchClusters = async () => {
    try {
      const response = await clusterApi.getClusters()
      setClusters(response.data)
      if (response.data.length > 0) {
        setSelectedCluster(response.data[0].name)
      }
    } catch (err) {
      setError('获取集群列表失败')
      console.error('Error fetching clusters:', err)
    }
  }

  useEffect(() => {
    if (selectedCluster) {
      fetchNodes()
      fetchNodeMetrics()
    }
  }, [selectedCluster])

  const fetchNodes = async () => {
    try {
      setLoading(true)
      setError(null)
      const response = await nodeApi.listNodes(selectedCluster)
      setNodes(response.data)
    } catch (err) {
      setError('获取节点列表失败')
      console.error('Error fetching nodes:', err)
    } finally {
      setLoading(false)
    }
  }

  const fetchNodeMetrics = async () => {
    try {
      const response = await nodeApi.getNodeMetrics(selectedCluster)
      setMetrics(response.data)
    } catch (err) {
      console.error('Error fetching node metrics:', err)
    }
  }

  const getNodeStatus = (node: any) => {
    const readyCondition = node.status.conditions.find(
      (cond: any) => cond.type === 'Ready'
    )
    return readyCondition ? readyCondition.status : 'Unknown'
  }

  return (
    <div>
      <div className="flex flex-col md:flex-row md:items-center justify-between mb-6 gap-4">
        <h1 className="text-2xl font-bold">节点（Node）管理</h1>
        <div className="flex items-center space-x-4">
          <select
            value={selectedCluster}
            onChange={(e) => setSelectedCluster(e.target.value)}
            className="input w-44 shrink-0"
          >
            <option value="">选择集群</option>
            {clusters.map((cluster) => (
              <option key={cluster.name} value={cluster.name}>
                {cluster.name}
              </option>
            ))}
          </select>
          <button
            onClick={() => {
              fetchNodes()
              fetchNodeMetrics()
            }}
            className="btn btn-secondary flex items-center space-x-2 whitespace-nowrap"
          >
            <RefreshCw className="h-4 w-4" />
            <span>刷新</span>
          </button>
        </div>
      </div>

      {error && (
        <div className="bg-red-50 dark:bg-red-950/40 border border-red-200 dark:border-red-800 text-red-700 dark:text-red-400 px-4 py-3 rounded mb-4">
          {error}
        </div>
      )}

      {loading ? (
        <div className="flex items-center justify-center min-h-[40vh]">
          <Loader2 className="h-8 w-8 animate-spin text-primary-600" />
        </div>
      ) : (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {nodes.map((node) => {
            const nodeMetric = metrics[node.metadata.name]
            const status = getNodeStatus(node)
            const internalIP = node.status.addresses?.find((a: any) => a.type === 'InternalIP')?.address
            const isControlPlane = Object.keys(node.metadata.labels || {}).some((k) => k === 'node-role.kubernetes.io/control-plane')
            return (
              <div key={node.metadata.name} className="card p-6">
                <div className="flex items-start justify-between gap-3 mb-4">
                  <div className="flex items-center space-x-3 min-w-0">
                    <Server className="h-6 w-6 text-primary-600 dark:text-primary-400 shrink-0" />
                    <div className="min-w-0">
                      <h2 className="text-lg font-semibold truncate" title={node.metadata.name}>{node.metadata.name}</h2>
                      <div className="flex flex-wrap items-center gap-x-2 gap-y-1 mt-1">
                        <span className={`inline-block h-2 w-2 rounded-full shrink-0 ${getStatusColor(status)}`} />
                        <span className="text-sm text-gray-600 dark:text-gray-400 whitespace-nowrap">
                          {status === 'True' ? '就绪' : '未就绪'}
                        </span>
                        {isControlPlane && (
                          <span className="px-1.5 py-0.5 rounded bg-primary-100 text-primary-700 dark:bg-primary-900/30 dark:text-primary-300 text-xs whitespace-nowrap">
                            control-plane
                          </span>
                        )}
                        {internalIP && <span className="text-sm font-mono text-gray-500 dark:text-gray-400">{internalIP}</span>}
                      </div>
                    </div>
                  </div>
                  <div className="text-sm text-gray-500 dark:text-gray-400 whitespace-nowrap shrink-0">
                    {formatDate(node.metadata.creationTimestamp)}
                  </div>
                </div>

                <div className="grid grid-cols-2 md:grid-cols-3 gap-4 mb-4">
                  <div className="bg-gray-50 dark:bg-gray-800 rounded-lg p-4">
                    <div className="flex items-center space-x-2 mb-2">
                      <Cpu className="h-4 w-4 text-gray-500 dark:text-gray-400" />
                      <h3 className="text-sm font-medium">CPU</h3>
                    </div>
                    <div className="text-lg font-semibold">
                      {nodeMetric ? nodeMetric.cpu : node.status.capacity.cpu}
                    </div>
                    {nodeMetric?.usage && (
                      <div className="mt-2">
                        <div className="flex items-center justify-between text-xs text-gray-500 dark:text-gray-400">
                          <span>使用率</span>
                          <span className={nodeMetric.usage.cpuPercent >= 80 ? 'text-red-600 font-medium' : ''}>{nodeMetric.usage.cpuPercent}%</span>
                        </div>
                        <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-1.5 mt-1">
                          <div
                            className={`h-1.5 rounded-full ${nodeMetric.usage.cpuPercent >= 80 ? 'bg-red-500' : 'bg-primary-500'}`}
                            style={{ width: `${Math.min(nodeMetric.usage.cpuPercent, 100)}%` }}
                          />
                        </div>
                      </div>
                    )}
                  </div>
                  <div className="bg-gray-50 dark:bg-gray-800 rounded-lg p-4">
                    <div className="flex items-center space-x-2 mb-2">
                      <HardDrive className="h-4 w-4 text-gray-500 dark:text-gray-400" />
                      <h3 className="text-sm font-medium">内存</h3>
                    </div>
                    <div className="text-lg font-semibold">
                      {nodeMetric ? nodeMetric.memory : node.status.capacity.memory}
                    </div>
                    {nodeMetric?.usage && (
                      <div className="mt-2">
                        <div className="flex items-center justify-between text-xs text-gray-500 dark:text-gray-400">
                          <span>使用率</span>
                          <span className={nodeMetric.usage.memoryPercent >= 80 ? 'text-red-600 font-medium' : ''}>{nodeMetric.usage.memoryPercent}%</span>
                        </div>
                        <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-1.5 mt-1">
                          <div
                            className={`h-1.5 rounded-full ${nodeMetric.usage.memoryPercent >= 80 ? 'bg-red-500' : 'bg-primary-500'}`}
                            style={{ width: `${Math.min(nodeMetric.usage.memoryPercent, 100)}%` }}
                          />
                        </div>
                      </div>
                    )}
                  </div>
                  <div className="bg-gray-50 dark:bg-gray-800 rounded-lg p-4">
                    <h3 className="text-sm font-medium mb-2">Pods</h3>
                    <div className="text-lg font-semibold">{nodeMetric?.usage?.pods ?? '-'}</div>
                    <div className="text-xs text-gray-500 dark:text-gray-400 mt-1">
                      可分配 {node.status.allocatable?.pods ?? '-'}
                    </div>
                  </div>
                </div>

                <div className="bg-gray-50 dark:bg-gray-800 rounded-lg p-4 mb-4">
                  <h3 className="text-sm font-medium mb-3">节点信息</h3>
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-2 text-sm">
                    <div className="min-w-0"><span className="text-gray-500 dark:text-gray-400">操作系统：</span><span className="font-mono break-words">{node.status.nodeInfo?.osImage || '-'}</span></div>
                    <div className="min-w-0"><span className="text-gray-500 dark:text-gray-400">架构：</span><span className="font-mono break-words">{node.status.nodeInfo?.architecture || '-'}</span></div>
                    <div className="min-w-0"><span className="text-gray-500 dark:text-gray-400">容器运行时：</span><span className="font-mono break-words">{node.status.nodeInfo?.containerRuntimeVersion || '-'}</span></div>
                    <div className="min-w-0"><span className="text-gray-500 dark:text-gray-400">kubelet：</span><span className="font-mono break-words">{node.status.nodeInfo?.kubeletVersion || '-'}</span></div>
                    <div className="min-w-0"><span className="text-gray-500 dark:text-gray-400">可分配 CPU：</span><span className="font-mono break-words">{node.status.allocatable?.cpu || '-'}</span></div>
                    <div className="min-w-0"><span className="text-gray-500 dark:text-gray-400">可分配内存：</span><span className="font-mono break-words">{node.status.allocatable?.memory || '-'}</span></div>
                  </div>
                </div>

                <div className="bg-gray-50 dark:bg-gray-800 rounded-lg p-4">
                  <h3 className="text-sm font-medium mb-3">状态条件（Conditions）</h3>
                  <div className="space-y-2">
                    {node.status.conditions.map((condition: any) => {
                      const isReady = condition.type === 'Ready'
                      const pressure = !isReady && condition.status === 'True'
                      const label = pressure ? '压力告警' : isReady ? (condition.status === 'True' ? '就绪' : '未就绪') : '正常'
                      const colorClass = pressure || (isReady && condition.status !== 'True')
                        ? 'text-danger-600 dark:text-danger-400'
                        : isReady
                          ? 'text-success-600 dark:text-success-400'
                          : 'text-gray-500 dark:text-gray-400'
                      return (
                        <div key={condition.type} className="flex items-center justify-between">
                          <span className="text-sm">{condition.type}</span>
                          <span className={`text-sm font-medium ${colorClass}`}>{label}</span>
                        </div>
                      )
                    })}
                  </div>
                </div>
              </div>
            )
          })}
          {nodes.length === 0 && (
            <div className="text-center py-12 text-gray-500 dark:text-gray-400 col-span-full">
              暂无节点
            </div>
          )}
        </div>
      )}
    </div>
  )
}

export default NodesPage
