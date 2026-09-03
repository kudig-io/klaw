import { useState, useEffect } from 'react'
import { X, Globe, Network, Link2, Server, MapPin, Copy } from 'lucide-react'
import { type Service, serviceApi, type ServiceEndpoints } from '../lib/api'
import { useToast } from '../contexts/ToastContext'

interface ServiceDetailDrawerProps {
  isOpen: boolean
  onClose: () => void
  service: Service
  cluster: string
}

export function ServiceDetailDrawer({ isOpen, onClose, service, cluster }: ServiceDetailDrawerProps) {
  const { showToast } = useToast()
  const [endpoints, setEndpoints] = useState<ServiceEndpoints | null>(null)
  const [isLoadingEndpoints, setIsLoadingEndpoints] = useState(false)
  const [activeTab, setActiveTab] = useState<'overview' | 'ports' | 'endpoints'>('overview')
  const tabLabels = { overview: '概览', ports: '端口', endpoints: '端点' } as const

  useEffect(() => {
    if (isOpen && cluster) {
      loadEndpoints()
    }
  }, [isOpen, cluster, service])

  async function loadEndpoints() {
    setIsLoadingEndpoints(true)
    try {
      const response = await serviceApi.getServiceEndpoints(
        cluster,
        service.metadata.namespace,
        service.metadata.name
      )
      setEndpoints(response.data)
    } catch (error) {
      console.error('Failed to load endpoints:', error)
    } finally {
      setIsLoadingEndpoints(false)
    }
  }

  function copyToClipboard(text: string) {
    navigator.clipboard.writeText(text)
    showToast('已复制到剪贴板', 'success')
  }

  function formatAge(timestamp: string): string {
    const date = new Date(timestamp)
    const now = new Date()
    const diff = now.getTime() - date.getTime()
    
    const seconds = Math.floor(diff / 1000)
    const minutes = Math.floor(seconds / 60)
    const hours = Math.floor(minutes / 60)
    const days = Math.floor(hours / 24)
    
    if (days > 0) return `${days} 天`
    if (hours > 0) return `${hours} 小时`
    if (minutes > 0) return `${minutes} 分钟`
    return `${seconds} 秒`
  }

  if (!isOpen) return null

  return (
    <div className="fixed inset-0 z-50 flex justify-end">
      {/* Backdrop */}
      <div 
        className="absolute inset-0 bg-black/50 backdrop-blur-sm"
        onClick={onClose}
      />
      
      {/* Drawer */}
      <div className="relative w-full max-w-2xl bg-white dark:bg-gray-800 border-l border-gray-200 dark:border-gray-700 flex flex-col h-full">
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-gray-200 dark:border-gray-700">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-primary-100 dark:bg-primary-900/30 rounded-lg">
              <Globe className="w-5 h-5 text-primary-600 dark:text-primary-400" />
            </div>
            <div>
              <h2 className="text-lg font-semibold text-gray-900 dark:text-white">
                {service.metadata.name}
              </h2>
              <p className="text-sm text-gray-500 dark:text-gray-400">
                命名空间 {service.metadata.namespace}
              </p>
            </div>
          </div>
          <button
            onClick={onClose}
            aria-label="关闭详情"
            className="p-2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Tabs */}
        <div className="flex border-b border-gray-200 dark:border-gray-700 px-6">
          {(['overview', 'ports', 'endpoints'] as const).map((tab) => (
            <button
              key={tab}
              onClick={() => setActiveTab(tab)}
              className={`px-4 py-3 text-sm font-medium border-b-2 transition-colors ${
                activeTab === tab
                  ? 'border-primary-500 text-primary-600 dark:text-primary-400'
                  : 'border-transparent text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white'
              }`}
            >
              {tabLabels[tab]}
            </button>
          ))}
        </div>

        {/* Content */}
        <div className="flex-1 overflow-y-auto p-6">
          {activeTab === 'overview' && (
            <div className="space-y-6">
              {/* Basic Info */}
              <section>
                <h3 className="text-sm font-medium text-gray-900 dark:text-white mb-3 flex items-center gap-2">
                  <Network className="w-4 h-4" />
                  基本信息
                </h3>
                <div className="bg-gray-50 dark:bg-gray-700/50 rounded-lg p-4 space-y-3">
                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <label className="text-xs text-gray-500 dark:text-gray-400">类型</label>
                      <p className="text-sm font-medium text-gray-900 dark:text-white">
                        {service.spec.type}
                      </p>
                    </div>
                    <div>
                      <label className="text-xs text-gray-500 dark:text-gray-400">集群 IP</label>
                      <div className="flex items-center gap-2">
                        <p className="text-sm font-medium text-gray-900 dark:text-white font-mono">
                          {service.spec.clusterIP || '-'}
                        </p>
                        {service.spec.clusterIP && (
                          <button
                            onClick={() => copyToClipboard(service.spec.clusterIP)}
                            aria-label={`复制 ClusterIP ${service.spec.clusterIP}`}
                            className="text-gray-400 hover:text-primary-600"
                          >
                            <Copy className="w-3 h-3" />
                          </button>
                        )}
                      </div>
                    </div>
                    <div>
                      <label className="text-xs text-gray-500 dark:text-gray-400">存续时间</label>
                      <p className="text-sm font-medium text-gray-900 dark:text-white">
                        {formatAge(service.metadata.creationTimestamp)}
                      </p>
                    </div>
                    <div>
                      <label className="text-xs text-gray-500 dark:text-gray-400">命名空间</label>
                      <p className="text-sm font-medium text-gray-900 dark:text-white">
                        {service.metadata.namespace}
                      </p>
                    </div>
                    <div>
                      <label className="text-xs text-gray-500 dark:text-gray-400">会话亲和</label>
                      <p className="text-sm font-medium text-gray-900 dark:text-white">
                        {service.spec.sessionAffinity || 'None'}
                      </p>
                    </div>
                    <div>
                      <label className="text-xs text-gray-500 dark:text-gray-400">外部流量策略</label>
                      <p className="text-sm font-medium text-gray-900 dark:text-white">
                        {service.spec.externalTrafficPolicy || '-'}
                      </p>
                    </div>
                  </div>
                </div>
              </section>

              {/* External IPs */}
              {(service.spec.externalIPs?.length || 0) > 0 && (
                <section>
                  <h3 className="text-sm font-medium text-gray-900 dark:text-white mb-3 flex items-center gap-2">
                    <MapPin className="w-4 h-4" />
                    外部 IP
                  </h3>
                  <div className="bg-gray-50 dark:bg-gray-700/50 rounded-lg p-4">
                    <div className="flex flex-wrap gap-2">
                      {service.spec.externalIPs!.map((ip, idx) => (
                        <span 
                          key={idx}
                          className="inline-flex items-center gap-1 px-2 py-1 rounded text-xs font-medium bg-warning-100 text-warning-800 dark:bg-warning-900/30 dark:text-warning-300"
                        >
                          {ip}
                          <button
                            onClick={() => copyToClipboard(ip)}
                            aria-label={`复制外部 IP ${ip}`}
                            className="hover:text-warning-600"
                          >
                            <Copy className="w-3 h-3" />
                          </button>
                        </span>
                      ))}
                    </div>
                  </div>
                </section>
              )}

              {/* Load Balancer Ingress */}
              {(service.status.loadBalancer?.ingress?.length || 0) > 0 && (
                <section>
                  <h3 className="text-sm font-medium text-gray-900 dark:text-white mb-3 flex items-center gap-2">
                    <Globe className="w-4 h-4" />
                    负载均衡（Load Balancer）
                  </h3>
                  <div className="bg-gray-50 dark:bg-gray-700/50 rounded-lg p-4">
                    <div className="space-y-2">
                      {service.status.loadBalancer!.ingress!.map((ingress, idx) => (
                        <div key={idx} className="flex items-center gap-2">
                          {ingress.ip && (
                            <span className="inline-flex items-center gap-1 px-2 py-1 rounded text-xs font-medium bg-info-100 text-info-800 dark:bg-info-900/30 dark:text-info-300">
                              IP：{ingress.ip}
                              <button
                                onClick={() => copyToClipboard(ingress.ip!)}
                                aria-label={`复制 Ingress IP ${ingress.ip}`}
                                className="hover:text-info-600"
                              >
                                <Copy className="w-3 h-3" />
                              </button>
                            </span>
                          )}
                          {ingress.hostname && (
                            <span className="inline-flex items-center gap-1 px-2 py-1 rounded text-xs font-medium bg-primary-100 text-primary-800 dark:bg-primary-900/30 dark:text-primary-300">
                              主机名：{ingress.hostname}
                              <button
                                onClick={() => copyToClipboard(ingress.hostname!)}
                                aria-label={`复制主机名 ${ingress.hostname}`}
                                className="hover:text-primary-600"
                              >
                                <Copy className="w-3 h-3" />
                              </button>
                            </span>
                          )}
                        </div>
                      ))}
                    </div>
                  </div>
                </section>
              )}

              {/* Selector */}
              {service.spec.selector && Object.keys(service.spec.selector).length > 0 && (
                <section>
                  <h3 className="text-sm font-medium text-gray-900 dark:text-white mb-3 flex items-center gap-2">
                    <Link2 className="w-4 h-4" />
                    选择器
                  </h3>
                  <div className="bg-gray-50 dark:bg-gray-700/50 rounded-lg p-4">
                    <div className="flex flex-wrap gap-2">
                      {Object.entries(service.spec.selector).map(([key, value]) => (
                        <span 
                          key={key}
                          className="inline-flex items-center px-2 py-1 rounded text-xs font-medium bg-primary-100 text-primary-800 dark:bg-primary-900/30 dark:text-primary-300"
                        >
                          {key}: {value}
                        </span>
                      ))}
                    </div>
                  </div>
                </section>
              )}

              {/* Labels */}
              {service.metadata.labels && Object.keys(service.metadata.labels).length > 0 && (
                <section>
                  <h3 className="text-sm font-medium text-gray-900 dark:text-white mb-3">
                    标签（Labels）
                  </h3>
                  <div className="bg-gray-50 dark:bg-gray-700/50 rounded-lg p-4">
                    <div className="flex flex-wrap gap-2">
                      {Object.entries(service.metadata.labels).map(([key, value]) => (
                        <span 
                          key={key}
                          className="inline-flex items-center px-2 py-1 rounded text-xs font-medium bg-gray-200 text-gray-700 dark:bg-gray-600 dark:text-gray-300"
                        >
                          {key}={value}
                        </span>
                      ))}
                    </div>
                  </div>
                </section>
              )}

              {/* Annotations */}
              {service.metadata.annotations && Object.keys(service.metadata.annotations).length > 0 && (
                <section>
                  <h3 className="text-sm font-medium text-gray-900 dark:text-white mb-3">
                    注解（Annotations）
                  </h3>
                  <div className="bg-gray-50 dark:bg-gray-700/50 rounded-lg p-4 space-y-2">
                    {Object.entries(service.metadata.annotations).map(([key, value]) => (
                      <div key={key} className="text-xs">
                        <span className="font-medium text-gray-700 dark:text-gray-300">{key}:</span>
                        <span className="text-gray-600 dark:text-gray-400 ml-2">{value}</span>
                      </div>
                    ))}
                  </div>
                </section>
              )}
            </div>
          )}

          {activeTab === 'ports' && (
            <div className="space-y-4">
              <h3 className="text-sm font-medium text-gray-900 dark:text-white flex items-center gap-2">
                <Server className="w-4 h-4" />
                服务端口
              </h3>
              
              {service.spec.ports && service.spec.ports.length > 0 ? (
                <div className="space-y-3">
                  {service.spec.ports.map((port, idx) => (
                    <div 
                      key={idx}
                      className="bg-gray-50 dark:bg-gray-700/50 rounded-lg p-4"
                    >
                      <div className="flex items-center justify-between mb-3">
                        <span className="text-sm font-medium text-gray-900 dark:text-white">
                          {port.name || `端口 ${idx + 1}`}
                        </span>
                        <span className="text-xs px-2 py-1 rounded bg-primary-100 text-primary-800 dark:bg-primary-900/30 dark:text-primary-300">
                          {port.protocol}
                        </span>
                      </div>
                      <div className="grid grid-cols-3 gap-4 text-sm">
                        <div>
                          <label className="text-xs text-gray-500 dark:text-gray-400">端口</label>
                          <p className="font-medium text-gray-900 dark:text-white font-mono">
                            {port.port}
                          </p>
                        </div>
                        <div>
                          <label className="text-xs text-gray-500 dark:text-gray-400">目标端口</label>
                          <p className="font-medium text-gray-900 dark:text-white font-mono">
                            {port.targetPort}
                          </p>
                        </div>
                        {port.nodePort && (
                          <div>
                            <label className="text-xs text-gray-500 dark:text-gray-400">节点端口</label>
                            <p className="font-medium text-gray-900 dark:text-white font-mono">
                              {port.nodePort}
                            </p>
                          </div>
                        )}
                      </div>
                    </div>
                  ))}
                </div>
              ) : (
                <p className="text-gray-500 dark:text-gray-400 text-sm">未定义端口</p>
              )}
            </div>
          )}

          {activeTab === 'endpoints' && (
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <h3 className="text-sm font-medium text-gray-900 dark:text-white flex items-center gap-2">
                  <Network className="w-4 h-4" />
                  端点
                </h3>
                <button
                  onClick={loadEndpoints}
                  disabled={isLoadingEndpoints}
                  className="text-xs text-primary-600 hover:text-primary-700 dark:text-primary-400 dark:hover:text-primary-300"
                >
                  {isLoadingEndpoints ? '加载中…' : '刷新'}
                </button>
              </div>

              {isLoadingEndpoints ? (
                <div className="text-center py-8">
                  <div className="inline-block animate-spin rounded-full h-6 w-6 border-2 border-primary-500 border-t-transparent"></div>
                </div>
              ) : endpoints?.endpoints && endpoints.endpoints.length > 0 ? (
                <div className="space-y-4">
                  {endpoints.endpoints.map((subset, idx) => (
                    <div key={idx} className="bg-gray-50 dark:bg-gray-700/50 rounded-lg p-4">
                      {/* Ready Addresses */}
                      {subset.addresses && subset.addresses.length > 0 && (
                        <div className="mb-4">
                          <h4 className="text-xs font-medium text-success-600 dark:text-success-400 mb-2">
                            就绪（{subset.addresses.length}）
                          </h4>
                          <div className="space-y-2">
                            {subset.addresses.map((addr, addrIdx) => (
                              <div 
                                key={addrIdx}
                                className="flex items-center justify-between text-sm bg-white dark:bg-gray-800 p-2 rounded"
                              >
                                <div className="flex items-center gap-2">
                                  <span className="w-2 h-2 rounded-full bg-success-500"></span>
                                  <span className="font-mono text-gray-900 dark:text-white">{addr.ip}</span>
                                  {addr.targetRef && (
                                    <span className="text-xs text-gray-500 dark:text-gray-400">
                                      ({addr.targetRef.kind}: {addr.targetRef.name})
                                    </span>
                                  )}
                                </div>
                                <button
                                  onClick={() => copyToClipboard(addr.ip)}
                                  aria-label={`复制端点 IP ${addr.ip}`}
                                  className="text-gray-400 hover:text-primary-600"
                                >
                                  <Copy className="w-3 h-3" />
                                </button>
                              </div>
                            ))}
                          </div>
                        </div>
                      )}

                      {/* Not Ready Addresses */}
                      {subset.notReadyAddresses && subset.notReadyAddresses.length > 0 && (
                        <div>
                          <h4 className="text-xs font-medium text-danger-600 dark:text-danger-400 mb-2">
                            未就绪（{subset.notReadyAddresses.length}）
                          </h4>
                          <div className="space-y-2">
                            {subset.notReadyAddresses.map((addr, addrIdx) => (
                              <div 
                                key={addrIdx}
                                className="flex items-center justify-between text-sm bg-white dark:bg-gray-800 p-2 rounded"
                              >
                                <div className="flex items-center gap-2">
                                  <span className="w-2 h-2 rounded-full bg-danger-500"></span>
                                  <span className="font-mono text-gray-900 dark:text-white">{addr.ip}</span>
                                  {addr.targetRef && (
                                    <span className="text-xs text-gray-500 dark:text-gray-400">
                                      ({addr.targetRef.kind}: {addr.targetRef.name})
                                    </span>
                                  )}
                                </div>
                              </div>
                            ))}
                          </div>
                        </div>
                      )}

                      {/* Ports */}
                      {subset.ports && subset.ports.length > 0 && (
                        <div className="mt-4 pt-4 border-t border-gray-200 dark:border-gray-600">
                          <h4 className="text-xs font-medium text-gray-500 dark:text-gray-400 mb-2">
                            端点端口
                          </h4>
                          <div className="flex flex-wrap gap-2">
                            {subset.ports.map((port, portIdx) => (
                              <span 
                                key={portIdx}
                                className="inline-flex items-center px-2 py-1 rounded text-xs font-medium bg-gray-200 text-gray-700 dark:bg-gray-600 dark:text-gray-300"
                              >
                                {port.name || '未命名'}: {port.port}/{port.protocol}
                              </span>
                            ))}
                          </div>
                        </div>
                      )}
                    </div>
                  ))}
                </div>
              ) : (
                <div className="text-center py-8 text-gray-500 dark:text-gray-400">
                  <Network className="w-12 h-12 mx-auto mb-3 opacity-50" />
                  <p>未找到端点</p>
                  <p className="text-xs mt-1">服务选择器可能未匹配到任何容器组（Pod）</p>
                </div>
              )}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
