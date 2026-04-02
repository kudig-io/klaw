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
    showToast('Copied to clipboard', 'success')
  }

  function formatAge(timestamp: string): string {
    const date = new Date(timestamp)
    const now = new Date()
    const diff = now.getTime() - date.getTime()
    
    const seconds = Math.floor(diff / 1000)
    const minutes = Math.floor(seconds / 60)
    const hours = Math.floor(minutes / 60)
    const days = Math.floor(hours / 24)
    
    if (days > 0) return `${days} day${days > 1 ? 's' : ''}`
    if (hours > 0) return `${hours} hour${hours > 1 ? 's' : ''}`
    if (minutes > 0) return `${minutes} minute${minutes > 1 ? 's' : ''}`
    return `${seconds} second${seconds > 1 ? 's' : ''}`
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
      <div className="relative w-full max-w-2xl bg-white dark:bg-slate-800 shadow-2xl animate-in slide-in-from-right duration-300 flex flex-col h-full">
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-slate-200 dark:border-slate-700">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-primary-100 dark:bg-primary-900/30 rounded-lg">
              <Globe className="w-5 h-5 text-primary-600 dark:text-primary-400" />
            </div>
            <div>
              <h2 className="text-lg font-semibold text-slate-900 dark:text-white">
                {service.metadata.name}
              </h2>
              <p className="text-sm text-slate-500 dark:text-slate-400">
                {service.metadata.namespace} namespace
              </p>
            </div>
          </div>
          <button
            onClick={onClose}
            className="p-2 text-slate-400 hover:text-slate-600 dark:hover:text-slate-300 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-700 transition-colors"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Tabs */}
        <div className="flex border-b border-slate-200 dark:border-slate-700 px-6">
          {(['overview', 'ports', 'endpoints'] as const).map((tab) => (
            <button
              key={tab}
              onClick={() => setActiveTab(tab)}
              className={`px-4 py-3 text-sm font-medium border-b-2 transition-colors ${
                activeTab === tab
                  ? 'border-primary-500 text-primary-600 dark:text-primary-400'
                  : 'border-transparent text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white'
              }`}
            >
              {tab.charAt(0).toUpperCase() + tab.slice(1)}
            </button>
          ))}
        </div>

        {/* Content */}
        <div className="flex-1 overflow-y-auto p-6">
          {activeTab === 'overview' && (
            <div className="space-y-6">
              {/* Basic Info */}
              <section>
                <h3 className="text-sm font-medium text-slate-900 dark:text-white mb-3 flex items-center gap-2">
                  <Network className="w-4 h-4" />
                  Basic Information
                </h3>
                <div className="bg-slate-50 dark:bg-slate-700/50 rounded-lg p-4 space-y-3">
                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <label className="text-xs text-slate-500 dark:text-slate-400">Type</label>
                      <p className="text-sm font-medium text-slate-900 dark:text-white">
                        {service.spec.type}
                      </p>
                    </div>
                    <div>
                      <label className="text-xs text-slate-500 dark:text-slate-400">Cluster IP</label>
                      <div className="flex items-center gap-2">
                        <p className="text-sm font-medium text-slate-900 dark:text-white font-mono">
                          {service.spec.clusterIP || '-'}
                        </p>
                        {service.spec.clusterIP && (
                          <button
                            onClick={() => copyToClipboard(service.spec.clusterIP)}
                            className="text-slate-400 hover:text-primary-600"
                          >
                            <Copy className="w-3 h-3" />
                          </button>
                        )}
                      </div>
                    </div>
                    <div>
                      <label className="text-xs text-slate-500 dark:text-slate-400">Age</label>
                      <p className="text-sm font-medium text-slate-900 dark:text-white">
                        {formatAge(service.metadata.creationTimestamp)}
                      </p>
                    </div>
                    <div>
                      <label className="text-xs text-slate-500 dark:text-slate-400">Namespace</label>
                      <p className="text-sm font-medium text-slate-900 dark:text-white">
                        {service.metadata.namespace}
                      </p>
                    </div>
                  </div>
                </div>
              </section>

              {/* External IPs */}
              {(service.spec.externalIPs?.length || 0) > 0 && (
                <section>
                  <h3 className="text-sm font-medium text-slate-900 dark:text-white mb-3 flex items-center gap-2">
                    <MapPin className="w-4 h-4" />
                    External IPs
                  </h3>
                  <div className="bg-slate-50 dark:bg-slate-700/50 rounded-lg p-4">
                    <div className="flex flex-wrap gap-2">
                      {service.spec.externalIPs!.map((ip, idx) => (
                        <span 
                          key={idx}
                          className="inline-flex items-center gap-1 px-2 py-1 rounded text-xs font-medium bg-orange-100 text-orange-800 dark:bg-orange-900/30 dark:text-orange-300"
                        >
                          {ip}
                          <button
                            onClick={() => copyToClipboard(ip)}
                            className="hover:text-orange-600"
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
                  <h3 className="text-sm font-medium text-slate-900 dark:text-white mb-3 flex items-center gap-2">
                    <Globe className="w-4 h-4" />
                    Load Balancer
                  </h3>
                  <div className="bg-slate-50 dark:bg-slate-700/50 rounded-lg p-4">
                    <div className="space-y-2">
                      {service.status.loadBalancer!.ingress!.map((ingress, idx) => (
                        <div key={idx} className="flex items-center gap-2">
                          {ingress.ip && (
                            <span className="inline-flex items-center gap-1 px-2 py-1 rounded text-xs font-medium bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-300">
                              IP: {ingress.ip}
                              <button
                                onClick={() => copyToClipboard(ingress.ip!)}
                                className="hover:text-blue-600"
                              >
                                <Copy className="w-3 h-3" />
                              </button>
                            </span>
                          )}
                          {ingress.hostname && (
                            <span className="inline-flex items-center gap-1 px-2 py-1 rounded text-xs font-medium bg-purple-100 text-purple-800 dark:bg-purple-900/30 dark:text-purple-300">
                              Host: {ingress.hostname}
                              <button
                                onClick={() => copyToClipboard(ingress.hostname!)}
                                className="hover:text-purple-600"
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
                  <h3 className="text-sm font-medium text-slate-900 dark:text-white mb-3 flex items-center gap-2">
                    <Link2 className="w-4 h-4" />
                    Selector
                  </h3>
                  <div className="bg-slate-50 dark:bg-slate-700/50 rounded-lg p-4">
                    <div className="flex flex-wrap gap-2">
                      {Object.entries(service.spec.selector).map(([key, value]) => (
                        <span 
                          key={key}
                          className="inline-flex items-center px-2 py-1 rounded text-xs font-medium bg-indigo-100 text-indigo-800 dark:bg-indigo-900/30 dark:text-indigo-300"
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
                  <h3 className="text-sm font-medium text-slate-900 dark:text-white mb-3">
                    Labels
                  </h3>
                  <div className="bg-slate-50 dark:bg-slate-700/50 rounded-lg p-4">
                    <div className="flex flex-wrap gap-2">
                      {Object.entries(service.metadata.labels).map(([key, value]) => (
                        <span 
                          key={key}
                          className="inline-flex items-center px-2 py-1 rounded text-xs font-medium bg-slate-200 text-slate-700 dark:bg-slate-600 dark:text-slate-300"
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
                  <h3 className="text-sm font-medium text-slate-900 dark:text-white mb-3">
                    Annotations
                  </h3>
                  <div className="bg-slate-50 dark:bg-slate-700/50 rounded-lg p-4 space-y-2">
                    {Object.entries(service.metadata.annotations).map(([key, value]) => (
                      <div key={key} className="text-xs">
                        <span className="font-medium text-slate-700 dark:text-slate-300">{key}:</span>
                        <span className="text-slate-600 dark:text-slate-400 ml-2">{value}</span>
                      </div>
                    ))}
                  </div>
                </section>
              )}
            </div>
          )}

          {activeTab === 'ports' && (
            <div className="space-y-4">
              <h3 className="text-sm font-medium text-slate-900 dark:text-white flex items-center gap-2">
                <Server className="w-4 h-4" />
                Service Ports
              </h3>
              
              {service.spec.ports && service.spec.ports.length > 0 ? (
                <div className="space-y-3">
                  {service.spec.ports.map((port, idx) => (
                    <div 
                      key={idx}
                      className="bg-slate-50 dark:bg-slate-700/50 rounded-lg p-4"
                    >
                      <div className="flex items-center justify-between mb-3">
                        <span className="text-sm font-medium text-slate-900 dark:text-white">
                          {port.name || `Port ${idx + 1}`}
                        </span>
                        <span className="text-xs px-2 py-1 rounded bg-primary-100 text-primary-800 dark:bg-primary-900/30 dark:text-primary-300">
                          {port.protocol}
                        </span>
                      </div>
                      <div className="grid grid-cols-3 gap-4 text-sm">
                        <div>
                          <label className="text-xs text-slate-500 dark:text-slate-400">Port</label>
                          <p className="font-medium text-slate-900 dark:text-white font-mono">
                            {port.port}
                          </p>
                        </div>
                        <div>
                          <label className="text-xs text-slate-500 dark:text-slate-400">Target Port</label>
                          <p className="font-medium text-slate-900 dark:text-white font-mono">
                            {port.targetPort}
                          </p>
                        </div>
                        {port.nodePort && (
                          <div>
                            <label className="text-xs text-slate-500 dark:text-slate-400">Node Port</label>
                            <p className="font-medium text-slate-900 dark:text-white font-mono">
                              {port.nodePort}
                            </p>
                          </div>
                        )}
                      </div>
                    </div>
                  ))}
                </div>
              ) : (
                <p className="text-slate-500 dark:text-slate-400 text-sm">No ports defined</p>
              )}
            </div>
          )}

          {activeTab === 'endpoints' && (
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <h3 className="text-sm font-medium text-slate-900 dark:text-white flex items-center gap-2">
                  <Network className="w-4 h-4" />
                  Endpoints
                </h3>
                <button
                  onClick={loadEndpoints}
                  disabled={isLoadingEndpoints}
                  className="text-xs text-primary-600 hover:text-primary-700 dark:text-primary-400 dark:hover:text-primary-300"
                >
                  {isLoadingEndpoints ? 'Loading...' : 'Refresh'}
                </button>
              </div>

              {isLoadingEndpoints ? (
                <div className="text-center py-8">
                  <div className="inline-block animate-spin rounded-full h-6 w-6 border-b-2 border-primary-500"></div>
                </div>
              ) : endpoints?.endpoints && endpoints.endpoints.length > 0 ? (
                <div className="space-y-4">
                  {endpoints.endpoints.map((subset, idx) => (
                    <div key={idx} className="bg-slate-50 dark:bg-slate-700/50 rounded-lg p-4">
                      {/* Ready Addresses */}
                      {subset.addresses && subset.addresses.length > 0 && (
                        <div className="mb-4">
                          <h4 className="text-xs font-medium text-green-600 dark:text-green-400 mb-2">
                            Ready ({subset.addresses.length})
                          </h4>
                          <div className="space-y-2">
                            {subset.addresses.map((addr, addrIdx) => (
                              <div 
                                key={addrIdx}
                                className="flex items-center justify-between text-sm bg-white dark:bg-slate-800 p-2 rounded"
                              >
                                <div className="flex items-center gap-2">
                                  <span className="w-2 h-2 rounded-full bg-green-500"></span>
                                  <span className="font-mono text-slate-900 dark:text-white">{addr.ip}</span>
                                  {addr.targetRef && (
                                    <span className="text-xs text-slate-500 dark:text-slate-400">
                                      ({addr.targetRef.kind}: {addr.targetRef.name})
                                    </span>
                                  )}
                                </div>
                                <button
                                  onClick={() => copyToClipboard(addr.ip)}
                                  className="text-slate-400 hover:text-primary-600"
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
                          <h4 className="text-xs font-medium text-red-600 dark:text-red-400 mb-2">
                            Not Ready ({subset.notReadyAddresses.length})
                          </h4>
                          <div className="space-y-2">
                            {subset.notReadyAddresses.map((addr, addrIdx) => (
                              <div 
                                key={addrIdx}
                                className="flex items-center justify-between text-sm bg-white dark:bg-slate-800 p-2 rounded"
                              >
                                <div className="flex items-center gap-2">
                                  <span className="w-2 h-2 rounded-full bg-red-500"></span>
                                  <span className="font-mono text-slate-900 dark:text-white">{addr.ip}</span>
                                  {addr.targetRef && (
                                    <span className="text-xs text-slate-500 dark:text-slate-400">
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
                        <div className="mt-4 pt-4 border-t border-slate-200 dark:border-slate-600">
                          <h4 className="text-xs font-medium text-slate-500 dark:text-slate-400 mb-2">
                            Endpoint Ports
                          </h4>
                          <div className="flex flex-wrap gap-2">
                            {subset.ports.map((port, portIdx) => (
                              <span 
                                key={portIdx}
                                className="inline-flex items-center px-2 py-1 rounded text-xs font-medium bg-slate-200 text-slate-700 dark:bg-slate-600 dark:text-slate-300"
                              >
                                {port.name || 'unnamed'}: {port.port}/{port.protocol}
                              </span>
                            ))}
                          </div>
                        </div>
                      )}
                    </div>
                  ))}
                </div>
              ) : (
                <div className="text-center py-8 text-slate-500 dark:text-slate-400">
                  <Network className="w-12 h-12 mx-auto mb-3 opacity-50" />
                  <p>No endpoints found</p>
                  <p className="text-xs mt-1">Service selector may not match any pods</p>
                </div>
              )}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
