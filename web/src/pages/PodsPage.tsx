import React, { useState, useEffect } from 'react'
import { clusterApi, podApi } from '../lib/api'
import { getStatusColor, formatDate } from '../lib/utils'
import { Search, RefreshCw, Loader2, ChevronDown, ChevronUp, Trash2 } from 'lucide-react'

const PodsPage: React.FC = () => {
  const [clusters, setClusters] = useState<any[]>([])
  const [selectedCluster, setSelectedCluster] = useState<string>('')
  const [namespaces, setNamespaces] = useState<any[]>([])
  const [selectedNamespace, setSelectedNamespace] = useState<string>('')
  const [pods, setPods] = useState<any[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [expandedPod, setExpandedPod] = useState<string | null>(null)
  const [podLogs, setPodLogs] = useState<Record<string, string>>({})
  const [logsLoading, setLogsLoading] = useState<Record<string, boolean>>({})
  const [searchTerm, setSearchTerm] = useState('')

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
      fetchNamespaces()
    }
  }, [selectedCluster])

  const fetchNamespaces = async () => {
    try {
      const response = await clusterApi.getNamespaces(selectedCluster)
      setNamespaces(response.data)
      setSelectedNamespace('default')
    } catch (err) {
      setError('Failed to fetch namespaces')
      console.error('Error fetching namespaces:', err)
    }
  }

  useEffect(() => {
    if (selectedCluster && selectedNamespace) {
      fetchPods()
    }
  }, [selectedCluster, selectedNamespace])

  const fetchPods = async () => {
    try {
      setLoading(true)
      setError(null)
      const response = await podApi.listPods(selectedCluster, selectedNamespace)
      setPods(response.data)
    } catch (err) {
      setError('Failed to fetch pods')
      console.error('Error fetching pods:', err)
    } finally {
      setLoading(false)
    }
  }

  const fetchPodLogs = async (podName: string) => {
    try {
      setLogsLoading({ ...logsLoading, [podName]: true })
      const response = await podApi.getPodLogs(selectedCluster, selectedNamespace, podName, 100)
      setPodLogs({ ...podLogs, [podName]: response.data.logs })
    } catch (err) {
      console.error('Error fetching pod logs:', err)
    } finally {
      setLogsLoading({ ...logsLoading, [podName]: false })
    }
  }

  const deletePod = async (podName: string) => {
    if (!confirm(`Are you sure you want to delete pod ${podName}?`)) {
      return
    }

    try {
      await podApi.deletePod(selectedCluster, selectedNamespace, podName)
      fetchPods()
    } catch (err) {
      setError('Failed to delete pod')
      console.error('Error deleting pod:', err)
    }
  }

  const togglePodDetails = (podName: string) => {
    if (expandedPod === podName) {
      setExpandedPod(null)
    } else {
      setExpandedPod(podName)
      if (!podLogs[podName]) {
        fetchPodLogs(podName)
      }
    }
  }

  const filteredPods = pods.filter(pod => 
    pod.metadata.name.toLowerCase().includes(searchTerm.toLowerCase())
  )

  return (
    <div>
      <div className="flex flex-col md:flex-row md:items-center justify-between mb-6 gap-4">
        <h1 className="text-2xl font-bold">Pods Management</h1>
        <div className="flex flex-col md:flex-row gap-4">
          <div className="flex items-center space-x-2">
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
          </div>
          <div className="flex items-center space-x-2">
            <select
              value={selectedNamespace}
              onChange={(e) => setSelectedNamespace(e.target.value)}
              className="input"
              disabled={!selectedCluster}
            >
              <option value="">Select Namespace</option>
              {namespaces.map((ns) => (
                <option key={ns.metadata.name} value={ns.metadata.name}>
                  {ns.metadata.name}
                </option>
              ))}
            </select>
          </div>
          <button
            onClick={fetchPods}
            className="btn btn-secondary flex items-center space-x-2 whitespace-nowrap"
          >
            <RefreshCw className="h-4 w-4" />
            <span>Refresh</span>
          </button>
        </div>
      </div>

      <div className="flex items-center mb-4">
        <div className="relative flex-1 max-w-md">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
          <input
            type="text"
            placeholder="Search pods..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="input pl-10"
          />
        </div>
        <div className="ml-4 text-sm text-gray-500 dark:text-gray-400">
          {filteredPods.length} pods
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
        <div className="overflow-x-auto">
          <table className="w-full border-collapse">
            <thead>
              <tr className="bg-gray-100 dark:bg-gray-800">
                <th className="px-6 py-3 text-left text-sm font-semibold text-gray-700 dark:text-gray-300">
                  Pod Name
                </th>
                <th className="px-6 py-3 text-left text-sm font-semibold text-gray-700 dark:text-gray-300">
                  Status
                </th>
                <th className="px-6 py-3 text-left text-sm font-semibold text-gray-700 dark:text-gray-300">
                  Node
                </th>
                <th className="px-6 py-3 text-left text-sm font-semibold text-gray-700 dark:text-gray-300">
                  IP
                </th>
                <th className="px-6 py-3 text-left text-sm font-semibold text-gray-700 dark:text-gray-300">
                  Created
                </th>
                <th className="px-6 py-3 text-right text-sm font-semibold text-gray-700 dark:text-gray-300">
                  Actions
                </th>
              </tr>
            </thead>
            <tbody>
              {filteredPods.map((pod) => (
                <React.Fragment key={pod.metadata.name}>
                  <tr className="border-b border-gray-200 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-800/50">
                    <td className="px-6 py-4 text-sm font-medium">
                      {pod.metadata.name}
                    </td>
                    <td className="px-6 py-4">
                      <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium">
                        <span className={`inline-block h-2 w-2 rounded-full mr-1 ${getStatusColor(pod.status.phase)}`} />
                        {pod.status.phase}
                      </span>
                    </td>
                    <td className="px-6 py-4 text-sm">
                      {pod.spec.nodeName || '-'}
                    </td>
                    <td className="px-6 py-4 text-sm">
                      {pod.status.podIP || '-'}
                    </td>
                    <td className="px-6 py-4 text-sm text-gray-500 dark:text-gray-400">
                      {formatDate(pod.metadata.creationTimestamp)}
                    </td>
                    <td className="px-6 py-4 text-right">
                      <div className="flex items-center justify-end space-x-2">
                        <button
                          onClick={() => togglePodDetails(pod.metadata.name)}
                          className="text-primary-600 hover:text-primary-800 dark:text-primary-400 dark:hover:text-primary-300"
                        >
                          {expandedPod === pod.metadata.name ? (
                            <ChevronUp className="h-5 w-5" />
                          ) : (
                            <ChevronDown className="h-5 w-5" />
                          )}
                        </button>
                        <button
                          onClick={() => deletePod(pod.metadata.name)}
                          className="text-danger-600 hover:text-danger-800 dark:text-danger-400 dark:hover:text-danger-300"
                          title="Delete Pod"
                        >
                          <Trash2 className="h-5 w-5" />
                        </button>
                      </div>
                    </td>
                  </tr>
                  {expandedPod === pod.metadata.name && (
                    <tr className="bg-gray-50 dark:bg-gray-800/30 border-b border-gray-200 dark:border-gray-700">
                      <td colSpan={6} className="px-6 py-4">
                        <div className="bg-gray-100 dark:bg-gray-900 rounded-lg p-4">
                          <h3 className="text-sm font-semibold mb-2">Logs for {pod.metadata.name}</h3>
                          {logsLoading[pod.metadata.name] ? (
                            <div className="flex items-center justify-center py-8">
                              <Loader2 className="h-5 w-5 animate-spin text-primary-600" />
                            </div>
                          ) : (
                            <pre className="text-xs overflow-auto max-h-60 whitespace-pre-wrap text-gray-700 dark:text-gray-300">
                              {podLogs[pod.metadata.name] || 'Loading logs...'}
                            </pre>
                          )}
                        </div>
                      </td>
                    </tr>
                  )}
                </React.Fragment>
              ))}
            </tbody>
          </table>
          {filteredPods.length === 0 && (
            <div className="text-center py-12 text-gray-500 dark:text-gray-400">
              No pods found
            </div>
          )}
        </div>
      )}
    </div>
  )
}

export default PodsPage
