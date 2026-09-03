import React, { useState, useEffect } from 'react'
import { clusterApi, podApi, type LogAnalysis } from '../lib/api'
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
  const [podAnalysis, setPodAnalysis] = useState<Record<string, LogAnalysis>>({})
  const [logsLoading, setLogsLoading] = useState<Record<string, boolean>>({})
  const [analysisLoading, setAnalysisLoading] = useState<Record<string, boolean>>({})
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
      setError('获取容器组列表失败')
      console.error('Error fetching pods:', err)
    } finally {
      setLoading(false)
    }
  }

  const getPodNamespace = (pod: any) => selectedNamespace || pod.metadata.namespace

  const getPodRestarts = (pod: any) =>
    (pod.status.containerStatuses || []).reduce((sum: number, cs: any) => sum + (cs.restartCount || 0), 0)

  const fetchPodLogs = async (pod: any) => {
    const podName = pod.metadata.name
    const namespace = getPodNamespace(pod)
    try {
      setLogsLoading((prev) => ({ ...prev, [podName]: true }))
      const response = await podApi.getPodLogs(selectedCluster, namespace, podName, 100)
      setPodLogs((prev) => ({ ...prev, [podName]: response.data.logs }))
    } catch (err) {
      console.error('Error fetching pod logs:', err)
    } finally {
      setLogsLoading((prev) => ({ ...prev, [podName]: false }))
    }
  }

  const fetchPodAnalysis = async (pod: any) => {
    const podName = pod.metadata.name
    const namespace = getPodNamespace(pod)
    try {
      setAnalysisLoading((prev) => ({ ...prev, [podName]: true }))
      const response = await podApi.analyzePodLogs(selectedCluster, namespace, podName, 200)
      setPodAnalysis((prev) => ({ ...prev, [podName]: response.data }))
    } catch (err) {
      console.error('Error analyzing pod logs:', err)
    } finally {
      setAnalysisLoading((prev) => ({ ...prev, [podName]: false }))
    }
  }

  const deletePod = async (pod: any) => {
    const podName = pod.metadata.name
    const namespace = getPodNamespace(pod)
    if (!confirm(`确定要删除容器组 ${podName} 吗？`)) {
      return
    }

    try {
      await podApi.deletePod(selectedCluster, namespace, podName)
      fetchPods()
    } catch (err) {
      setError('删除容器组失败')
      console.error('Error deleting pod:', err)
    }
  }

  const togglePodDetails = (pod: any) => {
    const podName = pod.metadata.name
    if (expandedPod === podName) {
      setExpandedPod(null)
    } else {
      setExpandedPod(podName)
      if (!podLogs[podName]) {
        fetchPodLogs(pod)
      }
      if (!podAnalysis[podName]) {
        fetchPodAnalysis(pod)
      }
    }
  }

  const filteredPods = pods.filter(pod => 
    pod.metadata.name.toLowerCase().includes(searchTerm.toLowerCase())
  )

  return (
    <div>
      <div className="flex flex-col md:flex-row md:items-center justify-between mb-6 gap-4">
        <h1 className="text-2xl font-bold">容器组（Pod）管理</h1>
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
            onClick={fetchPods}
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
            placeholder="搜索容器组…"
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="input pl-10"
          />
        </div>
        <div className="ml-4 text-sm text-gray-500 dark:text-gray-400">
          共 {filteredPods.length} 个容器组
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
          <table className="w-full min-w-[880px] border-collapse text-sm">
            <thead>
              <tr className="bg-gray-100 dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700">
                {['容器组名称', '命名空间', '状态', '节点', 'IP 地址', '重启', '容器', '创建时间'].map((label) => (
                  <th
                    key={label}
                    className="px-3 py-2 text-left text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-gray-400 whitespace-nowrap"
                  >
                    {label}
                  </th>
                ))}
                <th className="sticky right-0 bg-gray-100 dark:bg-gray-800 px-3 py-2 text-right text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-gray-400 shadow-[-6px_0_8px_-6px_rgba(0,0,0,0.15)]">
                  操作
                </th>
              </tr>
            </thead>
            <tbody>
              {filteredPods.map((pod) => (
                <React.Fragment key={pod.metadata.name}>
                  <tr className="group border-b border-gray-200 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-800/50">
                    <td className="px-3 py-2 font-mono font-medium max-w-[240px] truncate" title={pod.metadata.name}>
                      {pod.metadata.name}
                    </td>
                    <td className="px-3 py-2 text-gray-500 dark:text-gray-400 whitespace-nowrap">
                      {getPodNamespace(pod)}
                    </td>
                    <td className="px-3 py-2">
                      <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium whitespace-nowrap">
                        <span className={`inline-block h-2 w-2 rounded-full mr-1 ${getStatusColor(pod.status.phase)}`} />
                        {pod.status.phase}
                      </span>
                    </td>
                    <td className="px-3 py-2 font-mono max-w-[160px] truncate" title={pod.spec.nodeName || '-'}>
                      {pod.spec.nodeName || '-'}
                    </td>
                    <td className="px-3 py-2 font-mono whitespace-nowrap">
                      {pod.status.podIP || '-'}
                    </td>
                    <td className={`px-3 py-2 font-mono ${getPodRestarts(pod) > 0 ? 'font-semibold text-danger-600 dark:text-danger-400' : ''}`}>
                      {getPodRestarts(pod)}
                    </td>
                    <td className="px-3 py-2 font-mono text-xs text-gray-500 dark:text-gray-400 max-w-[220px] truncate" title={pod.spec.containers?.[0]?.image || '-'}>
                      {pod.spec.containers?.[0]?.image || '-'}
                    </td>
                    <td className="px-3 py-2 text-xs text-gray-500 dark:text-gray-400 whitespace-nowrap">
                      {formatDate(pod.metadata.creationTimestamp)}
                    </td>
                    <td className="sticky right-0 bg-white dark:bg-gray-900 group-hover:bg-gray-50 dark:group-hover:bg-gray-800 px-3 py-2 text-right shadow-[-6px_0_8px_-6px_rgba(0,0,0,0.15)]">
                      <div className="flex items-center justify-end space-x-1">
                        <button
                          onClick={() => togglePodDetails(pod)}
                          aria-label={expandedPod === pod.metadata.name ? '收起详情' : '展开详情'}
                          className="p-1 rounded text-primary-600 hover:text-primary-800 dark:text-primary-400 dark:hover:text-primary-300"
                        >
                          {expandedPod === pod.metadata.name ? (
                            <ChevronUp className="h-5 w-5" />
                          ) : (
                            <ChevronDown className="h-5 w-5" />
                          )}
                        </button>
                        <button
                          onClick={() => deletePod(pod)}
                          aria-label={`删除容器组 ${pod.metadata.name}`}
                          title="删除容器组"
                          className="p-1 rounded text-danger-600 hover:text-danger-800 dark:text-danger-400 dark:hover:text-danger-300"
                        >
                          <Trash2 className="h-5 w-5" />
                        </button>
                      </div>
                    </td>
                  </tr>
                  {expandedPod === pod.metadata.name && (
                    <tr className="bg-gray-50 dark:bg-gray-800/30 border-b border-gray-200 dark:border-gray-700">
                      <td colSpan={9} className="px-6 py-4">
                        <div className="bg-gray-100 dark:bg-gray-900 rounded-lg p-4">
                          <div className="grid grid-cols-2 md:grid-cols-4 gap-3 mb-4">
                            <div className="bg-white dark:bg-gray-800 rounded p-3">
                              <div className="text-xs text-gray-500 dark:text-gray-400">Pod IP</div>
                              <div className="text-sm font-semibold">{pod.status.podIP || '-'}</div>
                            </div>
                            <div className="bg-white dark:bg-gray-800 rounded p-3">
                              <div className="text-xs text-gray-500 dark:text-gray-400">QoS 等级</div>
                              <div className="text-sm font-semibold">{pod.status.qosClass || '-'}</div>
                            </div>
                            <div className="bg-white dark:bg-gray-800 rounded p-3">
                              <div className="text-xs text-gray-500 dark:text-gray-400">启动时间</div>
                              <div className="text-sm font-semibold">{pod.status.startTime ? formatDate(pod.status.startTime) : '-'}</div>
                            </div>
                            <div className="bg-white dark:bg-gray-800 rounded p-3">
                              <div className="text-xs text-gray-500 dark:text-gray-400">累计重启</div>
                              <div className={`text-sm font-semibold ${getPodRestarts(pod) > 0 ? 'text-red-600' : ''}`}>{getPodRestarts(pod)}</div>
                            </div>
                          </div>

                          {pod.spec.containers && pod.spec.containers.length > 0 && (
                            <div className="mb-4">
                              <h4 className="text-sm font-semibold mb-2">容器明细</h4>
                              <div className="space-y-2">
                                {pod.spec.containers.map((c: any) => {
                                  const cs = (pod.status.containerStatuses || []).find((s: any) => s.name === c.name)
                                  const state = cs?.state || {}
                                  const stateText = state.running
                                    ? 'Running'
                                    : state.waiting
                                      ? `Waiting（${state.waiting.reason}）`
                                      : state.terminated
                                        ? `Terminated（${state.terminated.reason}）`
                                        : '-'
                                  return (
                                    <div key={c.name} className="bg-white dark:bg-gray-800 rounded p-3 text-sm">
                                      <div className="flex items-center justify-between mb-1">
                                        <span className="font-medium">{c.name}</span>
                                        <span className="text-xs text-gray-500 dark:text-gray-400">
                                          {stateText} · 重启 {cs?.restartCount || 0} 次
                                        </span>
                                      </div>
                                      <div className="text-xs text-gray-500 dark:text-gray-400 mb-1">镜像：{c.image}</div>
                                      {(c.resources?.requests || c.resources?.limits) && (
                                        <div className="text-xs text-gray-500 dark:text-gray-400">
                                          {c.resources?.requests && (
                                            <span className="mr-3">
                                              Requests：CPU {c.resources.requests.cpu || '-'} / 内存 {c.resources.requests.memory || '-'}
                                            </span>
                                          )}
                                          {c.resources?.limits && (
                                            <span>Limits：CPU {c.resources.limits.cpu || '-'} / 内存 {c.resources.limits.memory || '-'}</span>
                                          )}
                                        </div>
                                      )}
                                      {state.waiting?.message && (
                                        <div className="text-xs text-red-600 mt-1">{state.waiting.message}</div>
                                      )}
                                    </div>
                                  )
                                })}
                              </div>
                            </div>
                          )}

                          {pod.status.conditions && pod.status.conditions.length > 0 && (
                            <div className="mb-4">
                              <h4 className="text-sm font-semibold mb-2">状态条件（Conditions）</h4>
                              <div className="grid grid-cols-2 md:grid-cols-4 gap-2">
                                {pod.status.conditions.map((cond: any) => (
                                  <div key={cond.type} className="bg-white dark:bg-gray-800 rounded p-2 text-xs">
                                    <div className="font-medium">{cond.type}</div>
                                    <div className={cond.status === 'True' ? 'text-green-600' : 'text-red-600'}>
                                      {cond.status}
                                      {cond.reason ? ` · ${cond.reason}` : ''}
                                    </div>
                                    {cond.message && (
                                      <div className="text-gray-500 dark:text-gray-400 mt-0.5">{cond.message}</div>
                                    )}
                                  </div>
                                ))}
                              </div>
                            </div>
                          )}

                          <h3 className="text-sm font-semibold mb-2">{pod.metadata.name} 的日志</h3>
                          <div className="text-xs text-gray-500 dark:text-gray-400 mb-3">
                            命名空间：{getPodNamespace(pod)}
                          </div>
                          {analysisLoading[pod.metadata.name] ? (
                            <div className="flex items-center justify-center py-3">
                              <Loader2 className="h-4 w-4 animate-spin text-primary-600" />
                            </div>
                          ) : podAnalysis[pod.metadata.name] && (
                            <div className="grid grid-cols-2 md:grid-cols-4 gap-3 mb-4">
                              <div className="bg-white dark:bg-gray-800 rounded p-3">
                                <div className="text-xs text-gray-500 dark:text-gray-400">错误</div>
                                <div className="text-lg font-semibold text-red-600">{podAnalysis[pod.metadata.name].errorCount}</div>
                              </div>
                              <div className="bg-white dark:bg-gray-800 rounded p-3">
                                <div className="text-xs text-gray-500 dark:text-gray-400">警告</div>
                                <div className="text-lg font-semibold text-yellow-600">{podAnalysis[pod.metadata.name].warningCount}</div>
                              </div>
                              <div className="bg-white dark:bg-gray-800 rounded p-3">
                                <div className="text-xs text-gray-500 dark:text-gray-400">安全事件</div>
                                <div className="text-lg font-semibold text-orange-600">{podAnalysis[pod.metadata.name].securityEvents?.length || 0}</div>
                              </div>
                              <div className="bg-white dark:bg-gray-800 rounded p-3">
                                <div className="text-xs text-gray-500 dark:text-gray-400">慢请求</div>
                                <div className="text-lg font-semibold text-blue-600">{podAnalysis[pod.metadata.name].performanceMetrics.slowRequests?.length || 0}</div>
                              </div>
                            </div>
                          )}
                          {logsLoading[pod.metadata.name] ? (
                            <div className="flex items-center justify-center py-8">
                              <Loader2 className="h-5 w-5 animate-spin text-primary-600" />
                            </div>
                          ) : (
                            <pre className="text-xs overflow-auto max-h-60 whitespace-pre-wrap text-gray-700 dark:text-gray-300">
                              {podLogs[pod.metadata.name] || '日志加载中…'}
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
              未找到容器组
            </div>
          )}
        </div>
      )}
    </div>
  )
}

export default PodsPage
