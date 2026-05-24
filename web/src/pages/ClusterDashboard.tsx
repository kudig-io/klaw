import React, { useState, useEffect } from 'react'
import { analysisApi, clusterApi, type RBACAnalysis } from '../lib/api'
import { cn, formatDate } from '../lib/utils'
import { RefreshCw, Loader2 } from 'lucide-react'

const ClusterDashboard: React.FC = () => {
  const [clusters, setClusters] = useState<any[]>([])
  const [statuses, setStatuses] = useState<Record<string, any>>({})
  const [rbacSummaries, setRbacSummaries] = useState<Record<string, RBACAnalysis>>({})
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

      const [statusResults, rbacResults] = await Promise.all([
        Promise.all(statusPromises),
        Promise.all(rbacPromises),
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
    } catch (err: any) {
      const errorMsg = err.response?.data?.error || err.message || String(err)
      setError(`Failed to fetch cluster data: ${errorMsg}`)
      console.error('Error fetching cluster data:', err)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchData()
  }, [])

  const getNodeStatusIcon = (ready: number, total: number) => {
    if (ready === total) return '✅'
    if (ready > 0) return '⚠️'
    return '❌'
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
          Try Again
        </button>
      </div>
    )
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold">Cluster Overview</h1>
        <button
          onClick={fetchData}
          className="btn btn-secondary flex items-center space-x-2"
        >
          <RefreshCw className="h-4 w-4" />
          <span>Refresh</span>
        </button>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {clusters.map((cluster) => {
          const status = statuses[cluster.name]
          return (
            <div key={cluster.name} className="card p-6">
              <div className="flex items-center justify-between mb-4">
                <h2 className="text-xl font-semibold">{cluster.name}</h2>
                <span className="text-sm text-gray-500 dark:text-gray-400">
                  {status ? formatDate(status.timestamp) : 'Loading...'}
                </span>
              </div>

              <div className="space-y-4">
                {status ? (
                  <>
                    <div className="bg-gray-50 dark:bg-gray-800 rounded-lg p-4">
                      <h3 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Nodes</h3>
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
                                ? 'bg-success-500 w-full' 
                                : status.nodes.ready > 0 
                                ? 'bg-warning-500' 
                                : 'bg-danger-500',
                              `w-[${(status.nodes.ready / status.nodes.total) * 100}%]`
                            )}
                          />
                        </div>
                      </div>
                    </div>

                    <div className="bg-gray-50 dark:bg-gray-800 rounded-lg p-4">
                      <h3 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Pods</h3>
                      <div className="grid grid-cols-3 gap-2 mb-2">
                        <div className="text-center">
                          <div className="text-sm text-gray-500 dark:text-gray-400">Running</div>
                          <div className="font-semibold text-success-600 dark:text-success-400">
                            {status.pods.running}
                          </div>
                        </div>
                        <div className="text-center">
                          <div className="text-sm text-gray-500 dark:text-gray-400">Pending</div>
                          <div className="font-semibold text-warning-600 dark:text-warning-400">
                            {status.pods.pending}
                          </div>
                        </div>
                        <div className="text-center">
                          <div className="text-sm text-gray-500 dark:text-gray-400">Failed</div>
                          <div className="font-semibold text-danger-600 dark:text-danger-400">
                            {status.pods.failed}
                          </div>
                        </div>
                      </div>
                      <div className="text-center text-sm text-gray-500 dark:text-gray-400">
                        Total: {status.pods.total}
                      </div>
                    </div>

                    <div className="bg-gray-50 dark:bg-gray-800 rounded-lg p-4">
                      <h3 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">RBAC Summary</h3>
                      {rbacSummaries[cluster.name] ? (
                        <div className="grid grid-cols-2 gap-2 text-sm">
                          <div>
                            <div className="text-gray-500 dark:text-gray-400">Roles</div>
                            <div className="font-semibold">{rbacSummaries[cluster.name].totalRoles}</div>
                          </div>
                          <div>
                            <div className="text-gray-500 dark:text-gray-400">ClusterRoles</div>
                            <div className="font-semibold">{rbacSummaries[cluster.name].totalClusterRoles}</div>
                          </div>
                          <div>
                            <div className="text-gray-500 dark:text-gray-400">Bindings</div>
                            <div className="font-semibold">{rbacSummaries[cluster.name].totalBindings}</div>
                          </div>
                          <div>
                            <div className="text-gray-500 dark:text-gray-400">ClusterBindings</div>
                            <div className="font-semibold">{rbacSummaries[cluster.name].totalClusterBindings}</div>
                          </div>
                        </div>
                      ) : (
                        <div className="text-sm text-gray-500 dark:text-gray-400">RBAC analysis unavailable</div>
                      )}
                    </div>

                    <div className="flex space-x-2">
                      <button className="btn btn-primary flex-1 text-sm">
                        View Details
                      </button>
                      <button className="btn btn-secondary text-sm">
                        View Metrics
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
