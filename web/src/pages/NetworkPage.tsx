import { useState, useEffect } from 'react'
import { useSearchParams } from 'react-router-dom'
import { clusterApi, networkApi, type Ingress, type NetworkPolicy, type NetworkAnalysis } from '../lib/api'
import { ClusterSelector } from '../components/ClusterSelector'
import { NamespaceSelector } from '../components/NamespaceSelector'
import { RefreshButton } from '../components/RefreshButton'
import { Globe, Network, Shield, Layers, ArrowLeftRight, Lock } from 'lucide-react'
import { useToast } from '../contexts/ToastContext'

const ALL_NAMESPACES = '_all' // Special value for all namespaces

export function NetworkPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const { showToast } = useToast()

  const [clusters, setClusters] = useState<Array<{ name: string }>>([])
  const [selectedCluster, setSelectedCluster] = useState(searchParams.get('cluster') || '')
  const [selectedNamespace, setSelectedNamespace] = useState(searchParams.get('namespace') || '')

  const [ingresses, setIngresses] = useState<Ingress[]>([])
  const [policies, setPolicies] = useState<NetworkPolicy[]>([])
  const [analysis, setAnalysis] = useState<NetworkAnalysis | null>(null)
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
      loadNetwork()
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

  async function loadNetwork() {
    if (!selectedCluster) return

    setIsLoading(true)
    try {
      const ns = selectedNamespace === ALL_NAMESPACES ? '' : selectedNamespace
      const [ingressRes, policyRes, analysisRes] = await Promise.all([
        networkApi.listIngresses(selectedCluster, ns),
        networkApi.listNetworkPolicies(selectedCluster, ns),
        networkApi.getNetworkAnalysis(),
      ])
      setIngresses(ingressRes.data || [])
      setPolicies(policyRes.data || [])
      setAnalysis(analysisRes.data)
    } catch (error) {
      console.error('Failed to load network resources:', error)
      showToast('加载网络资源失败', 'error')
      setIngresses([])
      setPolicies([])
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

  function getIngressHost(ingress: Ingress): string {
    const hosts = ingress.spec.rules.filter((r) => r.host).map((r) => r.host as string)
    if (hosts.length === 0) {
      return ingress.spec.defaultBackend ? '（默认后端）' : '-'
    }
    return hosts.length > 1 ? `${hosts[0]} +${hosts.length - 1}` : hosts[0]
  }

  function getIngressPathCount(ingress: Ingress): number {
    return ingress.spec.rules.reduce((acc, r) => acc + (r.http?.paths.length || 0), 0)
  }

  function getIngressTLS(ingress: Ingress): string | null {
    const tls = ingress.spec.tls?.[0]
    if (!tls) return null
    return tls.secretName || '已启用'
  }

  function getIngressAddress(ingress: Ingress): string {
    const addr = ingress.status.loadBalancer?.ingress?.[0]
    return addr?.ip || addr?.hostname || '等待分配'
  }

  function getPodSelectorSummary(np: NetworkPolicy): string {
    const labels = Object.entries(np.spec.podSelector.matchLabels || {})
    const hasExpr = (np.spec.podSelector.matchExpressions?.length || 0) > 0
    if (labels.length === 0 && !hasExpr) return '所有 Pod'
    const parts = labels.map(([k, v]) => `${k}=${v}`)
    if (hasExpr) parts.push('表达式选择')
    return parts.join(', ')
  }

  function getPolicyRuleSummary(np: NetworkPolicy): string {
    const parts: string[] = []
    if (np.spec.policyTypes.includes('Ingress')) {
      const rules = np.spec.ingress || []
      parts.push(rules.length === 0 ? '入站全部拒绝' : `入站 ${rules.length} 条`)
    }
    if (np.spec.policyTypes.includes('Egress')) {
      const rules = np.spec.egress || []
      parts.push(rules.length === 0 ? '出站全部拒绝' : `出站 ${rules.length} 条`)
    }
    return parts.join(' · ') || '-'
  }

  function getPolicyTypeColor(type: string): string {
    switch (type) {
      case 'Ingress':
        return 'bg-info-100 text-info-800 dark:bg-info-900/30 dark:text-info-300'
      case 'Egress':
        return 'bg-primary-100 text-primary-800 dark:bg-primary-900/30 dark:text-primary-300'
      default:
        return 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300'
    }
  }

  function getServiceTypeColor(type: string): string {
    switch (type) {
      case 'LoadBalancer':
        return 'bg-info-100 text-info-800 dark:bg-info-900/30 dark:text-info-300'
      case 'NodePort':
        return 'bg-primary-100 text-primary-800 dark:bg-primary-900/30 dark:text-primary-300'
      case 'ClusterIP':
        return 'bg-success-100 text-success-800 dark:bg-success-900/30 dark:text-success-300'
      default:
        return 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300'
    }
  }

  const policyNamespaceCount = new Set(policies.map((p) => p.metadata.namespace)).size

  return (
    <div className="p-6">
      <div className="mb-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold text-gray-900 dark:text-white">网络管理</h1>
            <p className="text-gray-600 dark:text-gray-400 mt-1">
              管理 Ingress 入口规则与 NetworkPolicy 网络策略，分析集群网络暴露面
            </p>
          </div>
          <RefreshButton onClick={loadNetwork} isLoading={isLoading} />
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
      </div>

      {/* Stats */}
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
        <div className="card p-5">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-600 dark:text-gray-400">Ingress 规则</p>
              <p className="text-2xl font-bold text-gray-900 dark:text-white mt-1">{ingresses.length}</p>
            </div>
            <div className="p-2.5 rounded-lg bg-info-100 dark:bg-info-900/30">
              <Globe className="h-5 w-5 text-info-600 dark:text-info-300" />
            </div>
          </div>
        </div>
        <div className="card p-5">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-600 dark:text-gray-400">网络策略</p>
              <p className="text-2xl font-bold text-gray-900 dark:text-white mt-1">{policies.length}</p>
            </div>
            <div className="p-2.5 rounded-lg bg-success-100 dark:bg-success-900/30">
              <Shield className="h-5 w-5 text-success-600 dark:text-success-300" />
            </div>
          </div>
        </div>
        <div className="card p-5">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-600 dark:text-gray-400">策略覆盖命名空间</p>
              <p className="text-2xl font-bold text-gray-900 dark:text-white mt-1">{policyNamespaceCount}</p>
            </div>
            <div className="p-2.5 rounded-lg bg-primary-100 dark:bg-primary-900/30">
              <Layers className="h-5 w-5 text-primary-600 dark:text-primary-300" />
            </div>
          </div>
        </div>
        <div className="card p-5">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-600 dark:text-gray-400">对外暴露服务</p>
              <p className="text-2xl font-bold text-gray-900 dark:text-white mt-1">
                {analysis ? analysis.exposedServices.length : '-'}
              </p>
            </div>
            <div className="p-2.5 rounded-lg bg-warning-100 dark:bg-warning-900/30">
              <ArrowLeftRight className="h-5 w-5 text-warning-600 dark:text-warning-300" />
            </div>
          </div>
        </div>
      </div>

      {/* Ingress Table */}
      <div className="mb-4 flex items-center gap-2">
        <Globe className="h-4 w-4 text-gray-500" />
        <h2 className="text-base font-semibold text-gray-900 dark:text-white">Ingress 入口规则</h2>
      </div>
      <div className="card overflow-hidden mb-8">
        {isLoading ? (
          <div className="p-12 text-center">
            <div className="inline-block animate-spin rounded-full h-8 w-8 border-2 border-primary-500 border-t-transparent"></div>
            <p className="mt-4 text-gray-600 dark:text-gray-400">正在加载网络资源…</p>
          </div>
        ) : ingresses.length === 0 ? (
          <div className="p-12 text-center">
            <Globe className="w-12 h-12 text-gray-300 dark:text-gray-600 mx-auto mb-4" />
            <h3 className="text-lg font-medium text-gray-900 dark:text-white mb-2">未找到 Ingress</h3>
            <p className="text-gray-600 dark:text-gray-400">当前范围下暂无 Ingress 规则</p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm text-left">
              <thead className="bg-gray-50 dark:bg-gray-700/50 text-gray-900 dark:text-white">
                <tr>
                  <th className="px-4 py-3 font-semibold">名称</th>
                  <th className="px-4 py-3 font-semibold">命名空间</th>
                  <th className="px-4 py-3 font-semibold">Ingress 类</th>
                  <th className="px-4 py-3 font-semibold">Host</th>
                  <th className="px-4 py-3 font-semibold">路径</th>
                  <th className="px-4 py-3 font-semibold">TLS</th>
                  <th className="px-4 py-3 font-semibold">地址</th>
                  <th className="px-4 py-3 font-semibold">存续时间</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200 dark:divide-gray-700">
                {ingresses.map((ingress) => (
                  <tr
                    key={`${ingress.metadata.namespace}-${ingress.metadata.name}`}
                    className="hover:bg-gray-50 dark:hover:bg-gray-700/30 transition-colors"
                  >
                    <td className="px-4 py-3 font-medium text-gray-900 dark:text-white">
                      {ingress.metadata.name}
                    </td>
                    <td className="px-4 py-3">
                      <span className="inline-flex items-center px-2 py-1 rounded text-xs font-medium bg-gray-100 text-gray-700 dark:bg-gray-700 dark:text-gray-300">
                        {ingress.metadata.namespace}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-gray-600 dark:text-gray-400 font-mono text-xs">
                      {ingress.spec.ingressClassName || '-'}
                    </td>
                    <td className="px-4 py-3 text-gray-600 dark:text-gray-400 font-mono text-xs">
                      {getIngressHost(ingress)}
                    </td>
                    <td className="px-4 py-3 text-gray-600 dark:text-gray-400">
                      {getIngressPathCount(ingress)}
                    </td>
                    <td className="px-4 py-3">
                      {getIngressTLS(ingress) ? (
                        <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded text-xs font-medium bg-success-100 text-success-800 dark:bg-success-900/30 dark:text-success-300">
                          <Lock className="h-3 w-3" />
                          {getIngressTLS(ingress)}
                        </span>
                      ) : (
                        <span className="text-gray-400 dark:text-gray-600">-</span>
                      )}
                    </td>
                    <td className="px-4 py-3 text-gray-600 dark:text-gray-400 font-mono text-xs">
                      {getIngressAddress(ingress)}
                    </td>
                    <td className="px-4 py-3 text-gray-600 dark:text-gray-400">
                      {formatAge(ingress.metadata.creationTimestamp)}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* NetworkPolicy Table */}
      <div className="mb-4 flex items-center gap-2">
        <Shield className="h-4 w-4 text-gray-500" />
        <h2 className="text-base font-semibold text-gray-900 dark:text-white">NetworkPolicy 网络策略</h2>
      </div>
      <div className="card overflow-hidden mb-8">
        {!isLoading && policies.length === 0 ? (
          <div className="p-12 text-center">
            <Shield className="w-12 h-12 text-gray-300 dark:text-gray-600 mx-auto mb-4" />
            <h3 className="text-lg font-medium text-gray-900 dark:text-white mb-2">未找到网络策略</h3>
            <p className="text-gray-600 dark:text-gray-400">当前范围下暂无 NetworkPolicy</p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm text-left">
              <thead className="bg-gray-50 dark:bg-gray-700/50 text-gray-900 dark:text-white">
                <tr>
                  <th className="px-4 py-3 font-semibold">名称</th>
                  <th className="px-4 py-3 font-semibold">命名空间</th>
                  <th className="px-4 py-3 font-semibold">类型</th>
                  <th className="px-4 py-3 font-semibold">目标 Pod 选择器</th>
                  <th className="px-4 py-3 font-semibold">规则摘要</th>
                  <th className="px-4 py-3 font-semibold">存续时间</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200 dark:divide-gray-700">
                {policies.map((np) => (
                  <tr
                    key={`${np.metadata.namespace}-${np.metadata.name}`}
                    className="hover:bg-gray-50 dark:hover:bg-gray-700/30 transition-colors"
                  >
                    <td className="px-4 py-3 font-medium text-gray-900 dark:text-white">
                      {np.metadata.name}
                    </td>
                    <td className="px-4 py-3">
                      <span className="inline-flex items-center px-2 py-1 rounded text-xs font-medium bg-gray-100 text-gray-700 dark:bg-gray-700 dark:text-gray-300">
                        {np.metadata.namespace}
                      </span>
                    </td>
                    <td className="px-4 py-3">
                      <div className="flex flex-wrap gap-1">
                        {np.spec.policyTypes.map((t) => (
                          <span
                            key={t}
                            className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ${getPolicyTypeColor(t)}`}
                          >
                            {t}
                          </span>
                        ))}
                      </div>
                    </td>
                    <td className="px-4 py-3 text-gray-600 dark:text-gray-400 font-mono text-xs">
                      {getPodSelectorSummary(np)}
                    </td>
                    <td className="px-4 py-3 text-gray-600 dark:text-gray-400">
                      {getPolicyRuleSummary(np)}
                    </td>
                    <td className="px-4 py-3 text-gray-600 dark:text-gray-400">
                      {formatAge(np.metadata.creationTimestamp)}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Network Analysis */}
      {analysis && (
        <>
          <div className="mb-4 flex items-center gap-2">
            <Network className="h-4 w-4 text-gray-500" />
            <h2 className="text-base font-semibold text-gray-900 dark:text-white">网络分析</h2>
          </div>
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 mb-6">
            <div className="card p-5">
              <h3 className="text-sm font-semibold text-gray-900 dark:text-white mb-3">服务类型分布</h3>
              <div className="space-y-2">
                {Object.entries(analysis.servicesByType).map(([type, count]) => (
                  <div key={type} className="flex items-center justify-between text-sm">
                    <span className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ${getServiceTypeColor(type)}`}>
                      {type}
                    </span>
                    <span className="font-semibold text-gray-900 dark:text-white">{count}</span>
                  </div>
                ))}
                {Object.keys(analysis.servicesByType).length === 0 && (
                  <p className="text-sm text-gray-500 dark:text-gray-400">暂无数据</p>
                )}
              </div>
            </div>
            <div className="card p-5">
              <h3 className="text-sm font-semibold text-gray-900 dark:text-white mb-3">策略分布（按命名空间）</h3>
              <div className="space-y-2">
                {Object.entries(analysis.policiesByNamespace).map(([ns, names]) => (
                  <div key={ns} className="flex items-center justify-between text-sm gap-2">
                    <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-gray-100 text-gray-700 dark:bg-gray-700 dark:text-gray-300 shrink-0">
                      {ns}
                    </span>
                    <span className="text-gray-500 dark:text-gray-400 truncate" title={names.join(', ')}>
                      {names.join(', ')}
                    </span>
                    <span className="font-semibold text-gray-900 dark:text-white shrink-0">{names.length}</span>
                  </div>
                ))}
                {Object.keys(analysis.policiesByNamespace).length === 0 && (
                  <p className="text-sm text-gray-500 dark:text-gray-400">暂无数据</p>
                )}
              </div>
            </div>
            <div className="card p-5">
              <h3 className="text-sm font-semibold text-gray-900 dark:text-white mb-3">Ingress Host 分布</h3>
              <div className="space-y-2">
                {Object.entries(analysis.ingressesByHost).map(([host, names]) => (
                  <div key={host} className="text-sm">
                    <div className="font-mono text-xs text-gray-700 dark:text-gray-300 truncate" title={host}>
                      {host}
                    </div>
                    <div className="text-xs text-gray-500 dark:text-gray-400">{names.join(', ')}</div>
                  </div>
                ))}
                {Object.keys(analysis.ingressesByHost).length === 0 && (
                  <p className="text-sm text-gray-500 dark:text-gray-400">暂无数据</p>
                )}
              </div>
            </div>
          </div>

          <div className="card overflow-hidden">
            <div className="px-4 py-3 border-b border-gray-200 dark:border-gray-700">
              <h3 className="text-sm font-semibold text-gray-900 dark:text-white">
                对外暴露服务（LoadBalancer / NodePort）
              </h3>
            </div>
            {analysis.exposedServices.length === 0 ? (
              <div className="p-8 text-center text-sm text-gray-500 dark:text-gray-400">
                当前集群没有对外暴露的服务
              </div>
            ) : (
              <div className="overflow-x-auto">
                <table className="w-full text-sm text-left">
                  <thead className="bg-gray-50 dark:bg-gray-700/50 text-gray-900 dark:text-white">
                    <tr>
                      <th className="px-4 py-3 font-semibold">名称</th>
                      <th className="px-4 py-3 font-semibold">命名空间</th>
                      <th className="px-4 py-3 font-semibold">类型</th>
                      <th className="px-4 py-3 font-semibold">端口</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-gray-200 dark:divide-gray-700">
                    {analysis.exposedServices.map((svc) => (
                      <tr
                        key={`${svc.namespace}-${svc.name}`}
                        className="hover:bg-gray-50 dark:hover:bg-gray-700/30 transition-colors"
                      >
                        <td className="px-4 py-3 font-medium text-gray-900 dark:text-white">{svc.name}</td>
                        <td className="px-4 py-3">
                          <span className="inline-flex items-center px-2 py-1 rounded text-xs font-medium bg-gray-100 text-gray-700 dark:bg-gray-700 dark:text-gray-300">
                            {svc.namespace}
                          </span>
                        </td>
                        <td className="px-4 py-3">
                          <span className={`inline-flex items-center px-2 py-1 rounded text-xs font-medium ${getServiceTypeColor(svc.type)}`}>
                            {svc.type}
                          </span>
                        </td>
                        <td className="px-4 py-3">
                          <div className="flex flex-wrap gap-1">
                            {svc.ports.map((port, idx) => (
                              <span
                                key={idx}
                                className="inline-flex items-center px-2 py-0.5 rounded text-xs bg-primary-50 text-primary-700 dark:bg-primary-900/30 dark:text-primary-300 font-mono"
                              >
                                {port.port}
                                {port.nodePort ? `:${port.nodePort}` : ''}/{port.protocol}
                              </span>
                            ))}
                          </div>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </div>
        </>
      )}
    </div>
  )
}
