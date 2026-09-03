import React, { useState, useEffect } from 'react'
import { clusterApi, deploymentApi, Deployment, DeploymentStatus } from '../lib/api'
import { cn, formatDate } from '../lib/utils'
import { Search, RefreshCw, Loader2, ChevronDown, ChevronUp, RotateCcw, Plus, Minus, Server, Box } from 'lucide-react'

const DeploymentsPage: React.FC = () => {
  const [clusters, setClusters] = useState<any[]>([])
  const [selectedCluster, setSelectedCluster] = useState<string>('')
  const [namespaces, setNamespaces] = useState<any[]>([])
  const [selectedNamespace, setSelectedNamespace] = useState<string>('')
  const [deployments, setDeployments] = useState<Deployment[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [expandedDeployment, setExpandedDeployment] = useState<string | null>(null)
  const [deploymentStatus, setDeploymentStatus] = useState<Record<string, DeploymentStatus>>({})
  const [statusLoading, setStatusLoading] = useState<Record<string, boolean>>({})
  const [searchTerm, setSearchTerm] = useState('')
  const [scalingDeployment, setScalingDeployment] = useState<string | null>(null)
  const [restartingDeployment, setRestartingDeployment] = useState<string | null>(null)

  const getDeploymentNamespace = (deployment: Deployment) => selectedNamespace || deployment.metadata.namespace

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
      fetchNamespaces()
    }
  }, [selectedCluster])

  const fetchNamespaces = async () => {
    try {
      const response = await clusterApi.getNamespaces(selectedCluster)
      setNamespaces(response.data)
      // 默认选择 "All Namespaces" (空字符串)
      setSelectedNamespace('')
    } catch (err) {
      setError('获取命名空间列表失败')
      console.error('Error fetching namespaces:', err)
    }
  }

  useEffect(() => {
    if (selectedCluster) {
      fetchDeployments()
    }
  }, [selectedCluster, selectedNamespace])

  const fetchDeployments = async () => {
    try {
      setLoading(true)
      setError(null)
      const response = await deploymentApi.listDeployments(selectedCluster, selectedNamespace)
      setDeployments(response.data)
    } catch (err) {
      setError('获取部署列表失败')
      console.error('Error fetching deployments:', err)
    } finally {
      setLoading(false)
    }
  }

  const fetchDeploymentStatus = async (deployment: Deployment) => {
    const deploymentName = deployment.metadata.name
    const namespace = getDeploymentNamespace(deployment)
    try {
      setStatusLoading((prev) => ({ ...prev, [deploymentName]: true }))
      const response = await deploymentApi.getDeploymentStatus(selectedCluster, namespace, deploymentName)
      setDeploymentStatus((prev) => ({ ...prev, [deploymentName]: response.data }))
    } catch (err) {
      console.error('Error fetching deployment status:', err)
    } finally {
      setStatusLoading((prev) => ({ ...prev, [deploymentName]: false }))
    }
  }

  const toggleDeploymentDetails = (deployment: Deployment) => {
    const deploymentName = deployment.metadata.name
    if (expandedDeployment === deploymentName) {
      setExpandedDeployment(null)
    } else {
      setExpandedDeployment(deploymentName)
      if (!deploymentStatus[deploymentName]) {
        fetchDeploymentStatus(deployment)
      }
    }
  }

  const scaleDeployment = async (deployment: Deployment, currentReplicas: number, delta: number) => {
    const deploymentName = deployment.metadata.name
    const namespace = getDeploymentNamespace(deployment)
    const newReplicas = Math.max(0, currentReplicas + delta)
    if (newReplicas === currentReplicas) return

    try {
      setScalingDeployment(deploymentName)
      await deploymentApi.scaleDeployment(selectedCluster, namespace, deploymentName, newReplicas)
      // 刷新 Deployment 列表
      await fetchDeployments()
      // 刷新状态
      await fetchDeploymentStatus(deployment)
    } catch (err) {
      setError('调整副本数失败')
      console.error('Error scaling deployment:', err)
    } finally {
      setScalingDeployment(null)
    }
  }

  const restartDeployment = async (deployment: Deployment) => {
    const deploymentName = deployment.metadata.name
    const namespace = getDeploymentNamespace(deployment)
    if (!confirm(`确定要重启部署 ${deploymentName} 吗？`)) {
      return
    }

    try {
      setRestartingDeployment(deploymentName)
      await deploymentApi.restartDeployment(selectedCluster, namespace, deploymentName)
      alert(`部署 ${deploymentName} 已重启`)
    } catch (err) {
      setError('重启部署失败')
      console.error('Error restarting deployment:', err)
    } finally {
      setRestartingDeployment(null)
    }
  }

  const getStatusColor = (deployment: Deployment) => {
    const available = deployment.status.availableReplicas || 0
    const desired = deployment.spec.replicas || 0
    
    if (available === desired && desired > 0) {
      return 'bg-green-500'
    } else if (available > 0) {
      return 'bg-yellow-500'
    } else if (desired === 0) {
      return 'bg-gray-400'
    } else {
      return 'bg-red-500'
    }
  }

  const getStatusText = (deployment: Deployment) => {
    const available = deployment.status.availableReplicas || 0
    const desired = deployment.spec.replicas || 0
    
    if (available === desired && desired > 0) {
      return '可用'
    } else if (available > 0) {
      return '变更中'
    } else if (desired === 0) {
      return '已缩容至 0'
    } else {
      return '不可用'
    }
  }

  const filteredDeployments = deployments.filter(deployment => 
    deployment.metadata.name.toLowerCase().includes(searchTerm.toLowerCase())
  )

  return (
    <div>
      <div className="flex flex-col md:flex-row md:items-center justify-between mb-6 gap-4">
        <h1 className="text-2xl font-bold">部署管理</h1>
        <div className="flex flex-col md:flex-row gap-4">
          <div className="flex items-center space-x-2">
            <select
              value={selectedCluster}
              onChange={(e) => setSelectedCluster(e.target.value)}
              className="input"
            >
              <option value="">选择集群</option>
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
              <option value="">全部命名空间</option>
              {namespaces.map((ns) => (
                <option key={ns.metadata.name} value={ns.metadata.name}>
                  {ns.metadata.name}
                </option>
              ))}
            </select>
          </div>
          <button
            onClick={fetchDeployments}
            className="btn btn-secondary flex items-center space-x-2 whitespace-nowrap"
          >
            <RefreshCw className="h-4 w-4" />
            <span>刷新</span>
          </button>
        </div>
      </div>

      <div className="flex items-center mb-4">
        <div className="relative flex-1 max-w-md">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
          <input
            type="text"
            placeholder="搜索部署…"
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="input pl-10"
          />
        </div>
        <div className="ml-4 text-sm text-gray-500 dark:text-gray-400">
          共 {filteredDeployments.length} 个部署
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
                  部署名称
                </th>
                <th className="px-6 py-3 text-left text-sm font-semibold text-gray-700 dark:text-gray-300">
                  命名空间
                </th>
                <th className="px-6 py-3 text-left text-sm font-semibold text-gray-700 dark:text-gray-300">
                  状态
                </th>
                <th className="px-6 py-3 text-left text-sm font-semibold text-gray-700 dark:text-gray-300">
                  副本
                </th>
                <th className="px-6 py-3 text-left text-sm font-semibold text-gray-700 dark:text-gray-300">
                  镜像
                </th>
                <th className="px-6 py-3 text-left text-sm font-semibold text-gray-700 dark:text-gray-300">
                  创建时间
                </th>
                <th className="px-6 py-3 text-right text-sm font-semibold text-gray-700 dark:text-gray-300">
                  操作
                </th>
              </tr>
            </thead>
            <tbody>
              {filteredDeployments.map((deployment) => (
                <React.Fragment key={deployment.metadata.name}>
                  <tr className="border-b border-gray-200 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-800/50">
                    <td className="px-6 py-4 text-sm font-medium">
                      <div className="flex items-center space-x-2">
                        <Box className="h-4 w-4 text-primary-600" />
                        <span>{deployment.metadata.name}</span>
                      </div>
                    </td>
                    <td className="px-6 py-4 text-sm text-gray-500 dark:text-gray-400">
                      {getDeploymentNamespace(deployment)}
                    </td>
                    <td className="px-6 py-4">
                      <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium">
                        <span className={`inline-block h-2 w-2 rounded-full mr-1 ${getStatusColor(deployment)}`} />
                        {getStatusText(deployment)}
                      </span>
                    </td>
                    <td className="px-6 py-4 text-sm">
                      <div className="flex items-center space-x-2">
                        <button
                          onClick={() => scaleDeployment(
                            deployment,
                            deployment.spec.replicas || 0,
                            -1
                          )}
                          disabled={scalingDeployment === deployment.metadata.name || (deployment.spec.replicas || 0) <= 0}
                          className="p-1 rounded hover:bg-gray-200 dark:hover:bg-gray-700 disabled:opacity-50"
                        >
                          <Minus className="h-4 w-4" />
                        </button>
                        <span className="font-medium min-w-[3rem] text-center">
                          {deployment.status.availableReplicas || 0}/{deployment.spec.replicas || 0}
                        </span>
                        <button
                          onClick={() => scaleDeployment(
                            deployment,
                            deployment.spec.replicas || 0,
                            1
                          )}
                          disabled={scalingDeployment === deployment.metadata.name}
                          className="p-1 rounded hover:bg-gray-200 dark:hover:bg-gray-700 disabled:opacity-50"
                        >
                          <Plus className="h-4 w-4" />
                        </button>
                      </div>
                    </td>
                    <td className="px-6 py-4 text-sm">
                      <span className="text-gray-600 dark:text-gray-400 truncate max-w-[200px] inline-block">
                        {deployment.spec.template.spec.containers[0]?.image || '-'}
                      </span>
                    </td>
                    <td className="px-6 py-4 text-sm text-gray-500 dark:text-gray-400">
                      {formatDate(deployment.metadata.creationTimestamp)}
                    </td>
                    <td className="px-6 py-4 text-right">
                      <div className="flex items-center justify-end space-x-2">
                        <button
                          onClick={() => restartDeployment(deployment)}
                          disabled={restartingDeployment === deployment.metadata.name}
                          className="p-2 rounded hover:bg-gray-200 dark:hover:bg-gray-700 disabled:opacity-50"
                          title="重启部署"
                        >
                          <RotateCcw className={cn(
                            "h-4 w-4",
                            restartingDeployment === deployment.metadata.name && "animate-spin"
                          )} />
                        </button>
                        <button
                          onClick={() => toggleDeploymentDetails(deployment)}
                          className="text-primary-600 hover:text-primary-800 dark:text-primary-400 dark:hover:text-primary-300"
                        >
                          {expandedDeployment === deployment.metadata.name ? (
                            <ChevronUp className="h-5 w-5" />
                          ) : (
                            <ChevronDown className="h-5 w-5" />
                          )}
                        </button>
                      </div>
                    </td>
                  </tr>
                  {expandedDeployment === deployment.metadata.name && (
                    <tr className="bg-gray-50 dark:bg-gray-800/30 border-b border-gray-200 dark:border-gray-700">
                      <td colSpan={7} className="px-6 py-4">
                        <div className="bg-gray-100 dark:bg-gray-900 rounded-lg p-4">
                          <h3 className="text-sm font-semibold mb-3 flex items-center">
                            <Server className="h-4 w-4 mr-2" />
                            部署详情：{deployment.metadata.name}
                          </h3>

                          <div className="grid grid-cols-2 md:grid-cols-3 gap-4 mb-4">
                            <div className="bg-white dark:bg-gray-800 rounded p-3">
                              <div className="text-xs text-gray-500 dark:text-gray-400">命名空间</div>
                              <div className="text-sm font-semibold">{getDeploymentNamespace(deployment)}</div>
                            </div>
                            <div className="bg-white dark:bg-gray-800 rounded p-3">
                              <div className="text-xs text-gray-500 dark:text-gray-400">标签选择器</div>
                              <div className="text-sm font-semibold font-mono break-all">
                                {Object.entries(deployment.spec.selector.matchLabels).map(([k, v]) => `${k}=${v}`).join(', ')}
                              </div>
                            </div>
                            <div className="bg-white dark:bg-gray-800 rounded p-3">
                              <div className="text-xs text-gray-500 dark:text-gray-400">标签</div>
                              <div className="text-sm font-semibold font-mono break-all">
                                {Object.entries(deployment.metadata.labels || {}).map(([k, v]) => `${k}=${v}`).join(', ') || '-'}
                              </div>
                            </div>
                          </div>
                          
                          {statusLoading[deployment.metadata.name] ? (
                            <div className="flex items-center justify-center py-8">
                              <Loader2 className="h-5 w-5 animate-spin text-primary-600" />
                            </div>
                          ) : deploymentStatus[deployment.metadata.name] ? (
                            <div className="space-y-4">
                              <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                                <div className="bg-white dark:bg-gray-800 rounded p-3">
                                  <div className="text-xs text-gray-500 dark:text-gray-400">期望副本</div>
                                  <div className="text-lg font-semibold">{deployment.spec.replicas || 0}</div>
                                </div>
                                <div className="bg-white dark:bg-gray-800 rounded p-3">
                                  <div className="text-xs text-gray-500 dark:text-gray-400">可用副本</div>
                                  <div className="text-lg font-semibold text-green-600">
                                    {deploymentStatus[deployment.metadata.name].availableReplicas}
                                  </div>
                                </div>
                                <div className="bg-white dark:bg-gray-800 rounded p-3">
                                  <div className="text-xs text-gray-500 dark:text-gray-400">就绪副本</div>
                                  <div className="text-lg font-semibold text-blue-600">
                                    {deploymentStatus[deployment.metadata.name].readyReplicas}
                                  </div>
                                </div>
                                <div className="bg-white dark:bg-gray-800 rounded p-3">
                                  <div className="text-xs text-gray-500 dark:text-gray-400">已更新副本</div>
                                  <div className="text-lg font-semibold">
                                    {deploymentStatus[deployment.metadata.name].updatedReplicas}
                                  </div>
                                </div>
                              </div>
                              
                              {deploymentStatus[deployment.metadata.name].conditions.length > 0 && (
                                <div>
                                  <h4 className="text-xs font-semibold text-gray-500 dark:text-gray-400 mb-2">状态条件（Conditions）</h4>
                                  <div className="space-y-2">
                                    {deploymentStatus[deployment.metadata.name].conditions.map((condition, idx) => (
                                      <div key={idx} className="bg-white dark:bg-gray-800 rounded p-2 text-sm">
                                        <div className="flex items-center justify-between">
                                          <span className="font-medium">{condition.type}</span>
                                          <span className={cn(
                                            "px-2 py-0.5 rounded text-xs",
                                            condition.status === 'True' 
                                              ? "bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400"
                                              : "bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-400"
                                          )}>
                                            {condition.status}
                                          </span>
                                        </div>
                                        {condition.reason && (
                                          <div className="text-xs text-gray-500 mt-1">{condition.reason}</div>
                                        )}
                                        {condition.message && (
                                          <div className="text-xs text-gray-400 mt-0.5">{condition.message}</div>
                                        )}
                                      </div>
                                    ))}
                                  </div>
                                </div>
                              )}

                              <div>
                                <h4 className="text-xs font-semibold text-gray-500 dark:text-gray-400 mb-2">容器（Containers）</h4>
                                <div className="space-y-2">
                                  {deployment.spec.template.spec.containers.map((container, idx) => (
                                    <div key={idx} className="bg-white dark:bg-gray-800 rounded p-2 text-sm">
                                      <div className="flex items-center justify-between">
                                        <span className="font-medium">{container.name}</span>
                                      </div>
                                      <div className="text-xs text-gray-500 mt-1 font-mono truncate">{container.image}</div>
                                      {(container.resources?.requests || container.resources?.limits) && (
                                        <div className="text-xs text-gray-500 dark:text-gray-400 mt-1">
                                          {container.resources?.requests && (
                                            <span className="mr-3">
                                              Requests：CPU {container.resources.requests.cpu || '-'} / 内存 {container.resources.requests.memory || '-'}
                                            </span>
                                          )}
                                          {container.resources?.limits && (
                                            <span>Limits：CPU {container.resources.limits.cpu || '-'} / 内存 {container.resources.limits.memory || '-'}</span>
                                          )}
                                        </div>
                                      )}
                                    </div>
                                  ))}
                                </div>
                              </div>
                            </div>
                          ) : (
                            <div className="text-center py-8 text-gray-500">暂无状态信息</div>
                          )}
                        </div>
                      </td>
                    </tr>
                  )}
                </React.Fragment>
              ))}
            </tbody>
          </table>
          {filteredDeployments.length === 0 && (
            <div className="text-center py-12 text-gray-500 dark:text-gray-400">
              未找到部署
            </div>
          )}
        </div>
      )}
    </div>
  )
}

export default DeploymentsPage
