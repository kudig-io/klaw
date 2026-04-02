import React, { useState, useEffect } from 'react'
import { clusterApi, monitoringApi } from '../lib/api'
import { formatDate } from '../lib/utils'
import { RefreshCw, Loader2, AlertCircle, Activity, Clock, AlertTriangle, AlertOctagon } from 'lucide-react'
import { XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, AreaChart, Area } from 'recharts'

const MonitoringPage: React.FC = () => {
  const [clusters, setClusters] = useState<any[]>([])
  const [selectedCluster, setSelectedCluster] = useState<string>('')
  const [monitoringStatus, setMonitoringStatus] = useState<any>(null)
  const [alerts, setAlerts] = useState<any[]>([])
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
      fetchMonitoringData()
    }
  }, [selectedCluster])

  const fetchMonitoringData = async () => {
    try {
      setLoading(true)
      setError(null)

      const [statusResponse, alertsResponse] = await Promise.all([
        monitoringApi.getStatus(selectedCluster),
        monitoringApi.getAlerts(selectedCluster)
      ])

      setMonitoringStatus(statusResponse.data)
      setAlerts(alertsResponse.data)
    } catch (err) {
      setError('Failed to fetch monitoring data')
      console.error('Error fetching monitoring data:', err)
    } finally {
      setLoading(false)
    }
  }

  const generateMockMetrics = () => {
    const data = []
    const now = new Date()
    for (let i = 23; i >= 0; i--) {
      const time = new Date(now.getTime() - i * 60 * 60 * 1000)
      data.push({
        time: time.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' }),
        cpu: Math.random() * 80 + 10,
        memory: Math.random() * 70 + 20,
        pods: Math.floor(Math.random() * 50) + 10,
      })
    }
    return data
  }

  const getAlertIcon = (level: string) => {
    switch (level) {
      case 'critical':
        return AlertOctagon
      case 'warning':
        return AlertTriangle
      default:
        return AlertCircle
    }
  }

  const getAlertColor = (level: string) => {
    switch (level) {
      case 'critical':
        return 'text-danger-600 dark:text-danger-400'
      case 'warning':
        return 'text-warning-600 dark:text-warning-400'
      default:
        return 'text-info-600 dark:text-info-400'
    }
  }

  const mockMetrics = generateMockMetrics()

  return (
    <div>
      <div className="flex flex-col md:flex-row md:items-center justify-between mb-6 gap-4">
        <h1 className="text-2xl font-bold">Monitoring</h1>
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
            onClick={fetchMonitoringData}
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
        <div className="space-y-8">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <div className="card p-6">
              <div className="flex items-center justify-between mb-4">
                <h2 className="text-lg font-semibold flex items-center space-x-2">
                  <Activity className="h-5 w-5 text-primary-600 dark:text-primary-400" />
                  CPU Usage
                </h2>
              </div>
              <div className="h-64">
                <ResponsiveContainer width="100%" height="100%">
                  <AreaChart data={mockMetrics}>
                    <defs>
                      <linearGradient id="colorCpu" x1="0" y1="0" x2="0" y2="1">
                        <stop offset="5%" stopColor="#0ea5e9" stopOpacity={0.8}/>
                        <stop offset="95%" stopColor="#0ea5e9" stopOpacity={0.1}/>
                      </linearGradient>
                    </defs>
                    <CartesianGrid strokeDasharray="3 3" stroke="rgba(0,0,0,0.1)" />
                    <XAxis dataKey="time" tick={{ fontSize: 12 }} />
                    <YAxis domain={[0, 100]} tick={{ fontSize: 12 }} />
                    <Tooltip />
                    <Area type="monotone" dataKey="cpu" stroke="#0ea5e9" fillOpacity={1} fill="url(#colorCpu)" />
                  </AreaChart>
                </ResponsiveContainer>
              </div>
            </div>

            <div className="card p-6">
              <div className="flex items-center justify-between mb-4">
                <h2 className="text-lg font-semibold flex items-center space-x-2">
                  <Activity className="h-5 w-5 text-primary-600 dark:text-primary-400" />
                  Memory Usage
                </h2>
              </div>
              <div className="h-64">
                <ResponsiveContainer width="100%" height="100%">
                  <AreaChart data={mockMetrics}>
                    <defs>
                      <linearGradient id="colorMemory" x1="0" y1="0" x2="0" y2="1">
                        <stop offset="5%" stopColor="#22c55e" stopOpacity={0.8}/>
                        <stop offset="95%" stopColor="#22c55e" stopOpacity={0.1}/>
                      </linearGradient>
                    </defs>
                    <CartesianGrid strokeDasharray="3 3" stroke="rgba(0,0,0,0.1)" />
                    <XAxis dataKey="time" tick={{ fontSize: 12 }} />
                    <YAxis domain={[0, 100]} tick={{ fontSize: 12 }} />
                    <Tooltip />
                    <Area type="monotone" dataKey="memory" stroke="#22c55e" fillOpacity={1} fill="url(#colorMemory)" />
                  </AreaChart>
                </ResponsiveContainer>
              </div>
            </div>
          </div>

          <div className="card p-6">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-lg font-semibold flex items-center space-x-2">
                <AlertCircle className="h-5 w-5 text-warning-600 dark:text-warning-400" />
                Alerts
              </h2>
            </div>
            {alerts.length > 0 ? (
              <div className="space-y-4">
                {alerts.map((alert: any, index: number) => {
                  const AlertIcon = getAlertIcon(alert.level)
                  return (
                    <div key={index} className="bg-gray-50 dark:bg-gray-800 rounded-lg p-4 border-l-4 border-warning-500">
                      <div className="flex items-start space-x-3">
                        <AlertIcon className={`h-5 w-5 ${getAlertColor(alert.level)} flex-shrink-0 mt-0.5`} />
                        <div className="flex-1">
                          <div className="flex items-center justify-between">
                            <h3 className="font-medium">{alert.message}</h3>
                            <span className="text-sm text-gray-500 dark:text-gray-400 flex items-center space-x-1">
                              <Clock className="h-3 w-3" />
                              {formatDate(alert.timestamp || new Date().toISOString())}
                            </span>
                          </div>
                          <p className="text-sm text-gray-600 dark:text-gray-400 mt-1">
                            {alert.type} - {alert.level}
                          </p>
                        </div>
                      </div>
                    </div>
                  )
                })}
              </div>
            ) : (
              <div className="text-center py-8 text-gray-500 dark:text-gray-400">
                No alerts found
              </div>
            )}
          </div>

          {monitoringStatus && (
            <div className="card p-6">
              <div className="flex items-center justify-between mb-4">
                <h2 className="text-lg font-semibold">Monitoring Status</h2>
              </div>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div className="bg-gray-50 dark:bg-gray-800 rounded-lg p-4">
                  <h3 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Status</h3>
                  <div className="text-lg font-semibold">
                    {monitoringStatus.active ? 'Active' : 'Inactive'}
                  </div>
                </div>
                <div className="bg-gray-50 dark:bg-gray-800 rounded-lg p-4">
                  <h3 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Data Points</h3>
                  <div className="text-lg font-semibold">
                    {monitoringStatus.dataPoints}
                  </div>
                </div>
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  )
}

export default MonitoringPage
