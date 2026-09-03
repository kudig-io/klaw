import { useState, useEffect } from 'react'
import { useSearchParams } from 'react-router-dom'
import { clusterApi, serviceApi, type Service } from '../lib/api'
import { ClusterSelector } from '../components/ClusterSelector'
import { NamespaceSelector } from '../components/NamespaceSelector'
import { RefreshButton } from '../components/RefreshButton'
import { ServiceDetailDrawer } from '../components/ServiceDetailDrawer'
import { Trash2, Globe, Network, Info } from 'lucide-react'
import { useToast } from '../contexts/ToastContext'

const ALL_NAMESPACES = '_all' // Special value for all namespaces

export function ServicesPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const { showToast } = useToast()

  const [clusters, setClusters] = useState<Array<{ name: string }>>([])
  const [selectedCluster, setSelectedCluster] = useState(searchParams.get('cluster') || '')
  const [selectedNamespace, setSelectedNamespace] = useState(searchParams.get('namespace') || '')
  
  const [services, setServices] = useState<Service[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const [selectedService, setSelectedService] = useState<Service | null>(null)
  const [isDrawerOpen, setIsDrawerOpen] = useState(false)

  // Load clusters on mount
  useEffect(() => {
    loadClusters()
  }, [])

  // Update URL when selection changes
  useEffect(() => {
    const params: Record<string, string> = {}
    if (selectedCluster) params.cluster = selectedCluster
    if (selectedNamespace && selectedNamespace !== ALL_NAMESPACES) params.namespace = selectedNamespace
    setSearchParams(params)
    
    if (selectedCluster) {
      loadServices()
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

  async function loadServices() {
    if (!selectedCluster) return
    
    setIsLoading(true)
    try {
      const ns = selectedNamespace === ALL_NAMESPACES ? '' : selectedNamespace
      const response = await serviceApi.listServices(selectedCluster, ns)
      setServices(response.data || [])
    } catch (error) {
      console.error('Failed to load services:', error)
      showToast('加载服务列表失败', 'error')
      setServices([])
    } finally {
      setIsLoading(false)
    }
  }

  async function handleDeleteService(service: Service) {
    if (!selectedCluster) return
    
    if (!confirm(`确定要删除服务"${service.metadata.name}"吗？`)) {
      return
    }

    try {
      await serviceApi.deleteService(
        selectedCluster,
        service.metadata.namespace,
        service.metadata.name
      )
      showToast(`服务"${service.metadata.name}"已删除`, 'success')
      loadServices()
    } catch (error) {
      console.error('Failed to delete service:', error)
      showToast('删除服务失败', 'error')
    }
  }

  function handleViewService(service: Service) {
    setSelectedService(service)
    setIsDrawerOpen(true)
  }

  function formatAge(timestamp: string): string {
    const date = new Date(timestamp)
    const now = new Date()
    const diff = now.getTime() - date.getTime()
    
    const seconds = Math.floor(diff / 1000)
    const minutes = Math.floor(seconds / 60)
    const hours = Math.floor(minutes / 60)
    const days = Math.floor(hours / 24)
    
    if (days > 0) return `${days}d`
    if (hours > 0) return `${hours}h`
    if (minutes > 0) return `${minutes}m`
    return `${seconds}s`
  }

  function getServiceTypeColor(type: string): string {
    switch (type) {
      case 'LoadBalancer':
        return 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-300'
      case 'NodePort':
        return 'bg-purple-100 text-purple-800 dark:bg-purple-900/30 dark:text-purple-300'
      case 'ClusterIP':
        return 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-300'
      case 'ExternalName':
        return 'bg-orange-100 text-orange-800 dark:bg-orange-900/30 dark:text-orange-300'
      default:
        return 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300'
    }
  }

  return (
    <div className="p-6">
      <div className="mb-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold text-gray-900 dark:text-white">服务管理</h1>
            <p className="text-gray-600 dark:text-gray-400 mt-1">
              管理 Kubernetes 服务（Service）及其端点
            </p>
          </div>
          <RefreshButton onClick={loadServices} isLoading={isLoading} />
        </div>
      </div>

      {/* Filters */}
      <div className="card p-4 mb-6">
        <div className="flex flex-wrap items-center gap-4">
          <ClusterSelector
            clusters={clusters.map(c => c.name)}
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

      {/* Services Table */}
      <div className="card overflow-hidden">
        {isLoading ? (
          <div className="p-12 text-center">
            <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-primary-500"></div>
            <p className="mt-4 text-gray-600 dark:text-gray-400">正在加载服务…</p>
          </div>
        ) : services.length === 0 ? (
          <div className="p-12 text-center">
            <Network className="w-12 h-12 text-gray-300 dark:text-gray-600 mx-auto mb-4" />
            <h3 className="text-lg font-medium text-gray-900 dark:text-white mb-2">未找到服务</h3>
            <p className="text-gray-600 dark:text-gray-400">
              {selectedNamespace && selectedNamespace !== ALL_NAMESPACES
                ? `命名空间"${selectedNamespace}"下暂无服务`
                : '当前集群下暂无服务'}
            </p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm text-left">
              <thead className="bg-gray-50 dark:bg-gray-700/50 text-gray-900 dark:text-white">
                <tr>
                  <th className="px-4 py-3 font-semibold">名称</th>
                  <th className="px-4 py-3 font-semibold">命名空间</th>
                  <th className="px-4 py-3 font-semibold">类型</th>
                  <th className="px-4 py-3 font-semibold">集群 IP</th>
                  <th className="px-4 py-3 font-semibold">外部 IP</th>
                  <th className="px-4 py-3 font-semibold">端口</th>
                  <th className="px-4 py-3 font-semibold">选择器</th>
                  <th className="px-4 py-3 font-semibold">存续时间</th>
                  <th className="px-4 py-3 font-semibold text-right">操作</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200 dark:divide-gray-700">
                {services.map((service) => (
                  <tr 
                    key={`${service.metadata.namespace}-${service.metadata.name}`}
                    className="hover:bg-gray-50 dark:hover:bg-gray-700/30 transition-colors"
                  >
                    <td className="px-4 py-3 font-medium text-gray-900 dark:text-white">
                      {service.metadata.name}
                    </td>
                    <td className="px-4 py-3">
                      <span className="inline-flex items-center px-2 py-1 rounded text-xs font-medium bg-gray-100 text-gray-700 dark:bg-gray-700 dark:text-gray-300">
                        {service.metadata.namespace}
                      </span>
                    </td>
                    <td className="px-4 py-3">
                      <span className={`inline-flex items-center px-2 py-1 rounded text-xs font-medium ${getServiceTypeColor(service.spec.type)}`}>
                        {service.spec.type}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-gray-600 dark:text-gray-400">
                      {service.spec.clusterIP || '-'}
                    </td>
                    <td className="px-4 py-3 text-gray-600 dark:text-gray-400 font-mono text-xs">
                      {service.status.loadBalancer?.ingress?.[0]?.ip || service.spec.externalIPs?.join(', ') || '-'}
                    </td>
                    <td className="px-4 py-3">
                      <div className="flex flex-wrap gap-1">
                        {service.spec.ports?.slice(0, 2).map((port, idx) => (
                          <span 
                            key={idx}
                            className="inline-flex items-center px-2 py-0.5 rounded text-xs bg-primary-50 text-primary-700 dark:bg-primary-900/30 dark:text-primary-300"
                          >
                            {port.port}:{port.targetPort}/{port.protocol}
                          </span>
                        ))}
                        {(service.spec.ports?.length || 0) > 2 && (
                          <span className="text-xs text-gray-500 dark:text-gray-400">
                            +{service.spec.ports!.length - 2}
                          </span>
                        )}
                      </div>
                    </td>
                    <td className="px-4 py-3">
                      {service.spec.selector ? (
                        <div className="flex flex-wrap gap-1">
                          {Object.entries(service.spec.selector).slice(0, 1).map(([key, value]) => (
                            <span 
                              key={key}
                              className="inline-flex items-center px-2 py-0.5 rounded text-xs bg-indigo-50 text-indigo-700 dark:bg-indigo-900/30 dark:text-indigo-300"
                            >
                              {key}={value}
                            </span>
                          ))}
                          {Object.keys(service.spec.selector).length > 1 && (
                            <span className="text-xs text-gray-500 dark:text-gray-400">
                              +{Object.keys(service.spec.selector).length - 1}
                            </span>
                          )}
                        </div>
                      ) : (
                        <span className="text-gray-400 dark:text-gray-600">-</span>
                      )}
                    </td>
                    <td className="px-4 py-3 text-gray-600 dark:text-gray-400">
                      {formatAge(service.metadata.creationTimestamp)}
                    </td>
                    <td className="px-4 py-3">
                      <div className="flex items-center justify-end gap-2">
                        <button
                          onClick={() => handleViewService(service)}
                          className="p-1.5 text-gray-600 hover:text-primary-600 hover:bg-primary-50 rounded-lg transition-colors"
                          title="查看详情"
                        >
                          <Info className="w-4 h-4" />
                        </button>
                        <button
                          onClick={() => handleDeleteService(service)}
                          className="p-1.5 text-gray-600 hover:text-red-600 hover:bg-red-50 rounded-lg transition-colors"
                          title="删除服务"
                        >
                          <Trash2 className="w-4 h-4" />
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Service Detail Drawer */}
      {selectedService && (
        <ServiceDetailDrawer
          isOpen={isDrawerOpen}
          onClose={() => setIsDrawerOpen(false)}
          service={selectedService}
          cluster={selectedCluster}
        />
      )}
    </div>
  )
}
