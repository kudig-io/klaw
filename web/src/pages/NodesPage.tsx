import React, { useState, useEffect } from 'react'
import { clusterApi, nodeApi } from '../lib/api'
import { formatDate, getStatusColor } from '../lib/utils'
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
      setError('Failed to fetch clusters')
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
      setError('Failed to fetch nodes')
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
        <h1 className="text-2xl font-bold">Nodes Management</h1>
        <div className="flex items-center space-x-4">
          <select
            value={selectedCluster}
            onChange={(e) => setSelectedCluster(e.target.value)}
            className="input"
          >
            <option value="">Select Cluster</option>
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
            className="btn btn-secondary flex items-center space-x-2"
          >
            <RefreshCw className="h-4 w-4" />
            <span>Refresh</span>
          </button>
        </div>
      </div>

      {error && (
        <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded mb-4">
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
            return (
              <div key={node.metadata.name} className="card p-6">
                <div className="flex items-start justify-between mb-4">
                  <div className="flex items-center space-x-3">
                    <Server className="h-6 w-6 text-primary-600 dark:text-primary-400" />
                    <div>
                      <h2 className="text-lg font-semibold">{node.metadata.name}</h2>
                      <div className="flex items-center space-x-2 mt-1">
                        <span className={`inline-block h-2 w-2 rounded-full ${getStatusColor(status)}`} />
                        <span className="text-sm text-gray-600 dark:text-gray-400">
                          {status === 'True' ? 'Ready' : 'Not Ready'}
                        </span>
                      </div>
                    </div>
                  </div>
                  <div className="text-sm text-gray-500 dark:text-gray-400">
                    {formatDate(node.metadata.creationTimestamp)}
                  </div>
                </div>

                <div className="grid grid-cols-2 gap-4 mb-4">
                  <div className="bg-gray-50 dark:bg-gray-800 rounded-lg p-4">
                    <div className="flex items-center space-x-2 mb-2">
                      <Cpu className="h-4 w-4 text-gray-500 dark:text-gray-400" />
                      <h3 className="text-sm font-medium">CPU</h3>
                    </div>
                    <div className="text-lg font-semibold">
                      {nodeMetric ? nodeMetric.cpu : node.status.capacity.cpu}
                    </div>
                  </div>
                  <div className="bg-gray-50 dark:bg-gray-800 rounded-lg p-4">
                    <div className="flex items-center space-x-2 mb-2">
                      <HardDrive className="h-4 w-4 text-gray-500 dark:text-gray-400" />
                      <h3 className="text-sm font-medium">Memory</h3>
                    </div>
                    <div className="text-lg font-semibold">
                      {nodeMetric ? nodeMetric.memory : node.status.capacity.memory}
                    </div>
                  </div>
                </div>

                <div className="bg-gray-50 dark:bg-gray-800 rounded-lg p-4">
                  <h3 className="text-sm font-medium mb-3">Conditions</h3>
                  <div className="space-y-2">
                    {node.status.conditions.map((condition: any) => (
                      <div key={condition.type} className="flex items-center justify-between">
                        <span className="text-sm">{condition.type}</span>
                        <span className={`text-sm font-medium ${getStatusColor(condition.status === 'True' ? 'Ready' : 'NotReady')}`}>
                          {condition.status === 'True' ? 'Ready' : 'Not Ready'}
                        </span>
                      </div>
                    ))}
                  </div>
                </div>
              </div>
            )
          })}
          {nodes.length === 0 && (
            <div className="text-center py-12 text-gray-500 dark:text-gray-400 col-span-full">
              No nodes found
            </div>
          )}
        </div>
      )}
    </div>
  )
}

export default NodesPage
