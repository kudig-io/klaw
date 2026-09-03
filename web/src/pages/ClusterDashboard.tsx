import React, { useState, useEffect } from 'react'
import { analysisApi, clusterApi, type RBACAnalysis } from '../lib/api'
import { cn, formatDate } from '../lib/utils'
import { RefreshCw, Loader2, CheckCircle2, AlertTriangle, XCircle } from 'lucide-react'

const ClusterDashboard: React.FC = () => {
  const [clusters, setClusters] = useState<any[]>([])
  const [statuses, setStatuses] = useState<Record<string, any>>({})
  const [rbacSummaries, setRbacSummaries] = useState<Record<string, RBACAnalysis>>({})
  const [metricsSummaries, setMetricsSummaries] = useState<Record<string, any>>({})
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const fetchData = async () => {
    try {
      setLoading(true)
      setError(null)
      
      const clustersResponse = await clusterApi.getClusters()
      setClusters(clustersResponse.data)

      const statusPromises = clustersResponse.data.map(async (cluster: any) => {
        const statusResponse = await clusterApi.getClusterStatus(cluster.name)
        return { [cluster.name]: statusResponse.data }
      })
      const rbacPromises = clustersResponse.data.map(async (cluster: any) => {
        try {
          const response = await analysisApi.analyzeRBAC(cluster.name)
          return { [cluster.name]: response.data }
        } catch {
          return { [cluster.name]: null }
        }
      })
      const metricsPromises = clustersResponse.data.map(async (cluster: any) => {
        try {
          const response = await clusterApi.getClusterMetrics(cluster.name)
          return { [cluster.name]: response.data }
        } catch {
          return { [cluster.name]: null }
        }
      })

      const [statusResults, rbacResults, metricsResults] = await Promise.all([
        Promise.all(statusPromises),
        Promise.all(rbacPromises),
        Promise.all(metricsPromises),
      ])
      const statusMap: Record<string, any> = {}
      statusResults.forEach((result: Record<string, any>) => {
        Object.assign(statusMap, result)
      })
      setStatuses(statusMap)

      const rbacMap: Record<string, RBACAnalysis> = {}
      rbacResults.forEach((result: Record<string, any>) => {
        Object.assign(rbacMap, result)
      })
      setRbacSummaries(rbacMap)

      const metricsMap: Record<string, any> = {}
      metricsResults.forEach((result: Record<string, any>) => {
        Object.assign(metricsMap, result)
      })
      setMetricsSummaries(metricsMap)
    } catch (err: any) {
      const errorMsg = err.response?.data?.error || err.message || String(err)
      setError(`集群数据获取失败：${errorMsg}`)
      console.error('Error fetching cluster data:', err)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchData()
  }, [])

  const getNodeStatusIcon = (ready: number, total: number) => {
    if (ready === total) return <CheckCircle2 className="h-4 w-4 text-success-600 dark:text-success-400" aria-hidden="true" />
    if (ready > 0) return <AlertTriangle className="h-4 w-4 text-warning-600 dark:text-warning-400" aria-hidden="true" />
    return <XCircle className="h-4 w-4 text-danger-600 dark:text-danger-400" aria-hidden="true" />
  }

  const getResourcePercent = (used: string, total: string) => {
    const u = parseFloat(used)
    const t = parseFloat(total)
    if (Number.isNaN(u) || Number.isNaN(t) || t === 0) return null
    return Math.min(Math.round((u / t) * 100), 100)
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-[60vh]">
        <Loader2 className="h-8 w-8 animate-spin text-primary-600" />
      </div>
    )
  }

  if (error) {
    return (
      <div className="text-center py-12">
        <div className="text-danger-500 mb-4">{error}</div>
        <button 
          onClick={fetchData}
          className="btn btn-primary"
        >
          重试
        </button>
      </div>
    )
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold">集群概览</h1>
        <button
          onClick={fetchData}
          className="btn btn-secondary flex items-center space-x-2"
        >
          <RefreshCw className="h-4 w-4" />
          <span>刷新</span>
        </button>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {clusters.map((cluster) => {
          const status = statuses[cluster.name]
          const metrics = metricsSummaries[cluster.name]
          const res = metrics?.resources
          return (
            <div key={cluster.name} className="card p-6">
              <div className="flex items-center justify-between mb-4">
                <h2 className="text-xl font-semibold">{cluster.name}</h2>
                <span className="text-sm text-gray-500 dark:text-gray-400">
                  {status ? formatDate(status.timestamp) : '加载中…'}
                </span>
              </div>

              <div className="space-y-4">
                {status ? (
                  <>
                    <div className="bg-gray-50 dark:bg-gray-800 rounded-lg p-4">
                      <h3 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">节点（Node）</h3>
                      <div className="flex items-center justify-between">
                        <div className="flex items-center space-x-2">
                          <span className="text-lg font-semibold">
                            {status.nodes.ready}/{status.nodes.total}
                          </span>
                          <span className="text-sm text-gray-500 dark:text-gray-400">
                            {getNodeStatusIcon(status.nodes.ready, status.nodes.total)}
                          </span>
                        </div>
                        <div className="w-24 bg-gray-200 dark:bg-gray-700 rounded-full h-2">
                          <div
                            className={cn(
                              'h-2 rounded-full transition-all duration-300',
                              status.nodes.ready === status.nodes.total
                                ? 'bg-success-500'
                                : status.nodes.ready > 0
                                ? 'bg-warning-500'
                                : 'bg-danger-500'
                            )}
                            style={{ width: `${(status.nodes.ready / status.nodes.total) * 100}%` }}
                          />
                        </div>
                      </div>
                    </div>

                    <div className="bg-gray-50 dark:bg-gray-800 rounded-lg p-4">
                      <h3 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">容器组（Pod）</h3>
                      <div className="grid grid-cols-3 gap-2 mb-2">
                        <div className="text-center">
                          <div className="text-sm text-gray-500 dark:text-gray-400">运行中</div>
                          <div className="font-semibold text-success-600 dark:text-success-400">
                            {status.pods.running}
                          </div>
                        </div>
                        <div className="text-center">
                          <div className="text-sm text-gray-500 dark:text-gray-400">等待中</div>
                          <div className="font-semibold text-warning-600 dark:text-warning-400">
                            {status.pods.pending}
                          </div>
                        </div>
                        <div className="text-center">
                          <div className="text-sm text-gray-500 dark:text-gray-400">失败</div>
                          <div className="font-semibold text-danger-600 dark:text-danger-400">
                            {status.pods.failed}
                          </div>
                        </div>
                      </div>
                      <div className="text-center text-sm text-gray-500 dark:text-gray-400">
                        共 {status.pods.total} 个
                      </div>
                    </div>

                    <div className="bg-gray-50 dark:bg-gray-800 rounded-lg p-4">
                      <h3 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">资源用量（Metrics）</h3>
                      {res ? (
                        <div className="space-y-3">
                          <div>
                            <div className="flex items-center justify-between text-sm mb-1">
                              <span className="text-gray-500 dark:text-gray-400">CPU（核）</span>
                              <span className="font-medium">
                                {res.usedCPU} / {res.totalCPU}
                              </span>
                            </div>
                            <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-1.5">
                              <div
                                className={cn(
                                  'h-1.5 rounded-full',
                                  (getResourcePercent(res.usedCPU, res.totalCPU) ?? 0) >= 80 ? 'bg-danger-500' : 'bg-primary-500'
                                )}
                                style={{ width: `${getResourcePercent(res.usedCPU, res.totalCPU) ?? 0}%` }}
                              />
                            </div>
                          </div>
                          <div>
                            <div className="flex items-center justify-between text-sm mb-1">
                              <span className="text-gray-500 dark:text-gray-400">内存（Gi）</span>
                              <span className="font-medium">
                                {res.usedMemory} / {res.totalMemory}
                              </span>
                            </div>
                            <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-1.5">
                              <div
                                className={cn(
                                  'h-1.5 rounded-full',
                                  (getResourcePercent(res.usedMemory, res.totalMemory) ?? 0) >= 80 ? 'bg-danger-500' : 'bg-primary-500'
                                )}
                                style={{ width: `${getResourcePercent(res.usedMemory, res.totalMemory) ?? 0}%` }}
                              />
                            </div>
                          </div>
                          <div className="text-xs text-gray-500 dark:text-gray-400 text-right">
                            采样时间：{formatDate(metrics.timestamp)}
                          </div>
                        </div>
                      ) : (
                        <div className="text-sm text-gray-500 dark:text-gray-400">资源指标暂不可用</div>
                      )}
                    </div>

                    <div className="bg-gray-50 dark:bg-gray-800 rounded-lg p-4">
                      <h3 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">RBAC（角色权限）概览</h3>
                      {rbacSummaries[cluster.name] ? (
                        <div className="grid grid-cols-2 gap-2 text-sm">
                          <div>
                            <div className="text-gray-500 dark:text-gray-400">角色（Role）</div>
                            <div className="font-semibold">{rbacSummaries[cluster.name].totalRoles}</div>
                          </div>
                          <div>
                            <div className="text-gray-500 dark:text-gray-400">集群角色（ClusterRole）</div>
                            <div className="font-semibold">{rbacSummaries[cluster.name].totalClusterRoles}</div>
                          </div>
                          <div>
                            <div className="text-gray-500 dark:text-gray-400">角色绑定（Binding）</div>
                            <div className="font-semibold">{rbacSummaries[cluster.name].totalBindings}</div>
                          </div>
                          <div>
                            <div className="text-gray-500 dark:text-gray-400">集群角色绑定（ClusterBinding）</div>
                            <div className="font-semibold">{rbacSummaries[cluster.name].totalClusterBindings}</div>
                          </div>
                        </div>
                      ) : (
                        <div className="text-sm text-gray-500 dark:text-gray-400">RBAC（角色权限）分析暂不可用</div>
                      )}
                    </div>

                    <div className="flex space-x-2">
                      <button className="btn btn-primary flex-1 text-sm">
                        查看详情
                      </button>
                      <button className="btn btn-secondary text-sm">
                        查看指标
                      </button>
                    </div>
                  </>
                ) : (
                  <div className="flex items-center justify-center h-32">
                    <Loader2 className="h-6 w-6 animate-spin text-gray-400" />
                  </div>
                )}
              </div>
            </div>
          )
        })}
      </div>
    </div>
  )
}

export default ClusterDashboard
