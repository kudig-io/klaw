import React, { useState, useEffect } from 'react'
import { clusterApi, monitoringApi } from '../lib/api'
import { cn, formatDate } from '../lib/utils'
import { RefreshCw, Loader2, AlertCircle, Activity, Clock, AlertTriangle, AlertOctagon } from 'lucide-react'

// Mock 数据
const mockData = {
  status: { active: true, cluster: 'kind-test', dataPoints: 1440 },
  alerts: [
    { id: '1', type: 'pod', level: 'warning', message: 'Pod pending > 5 min', createdAt: new Date().toISOString() },
    { id: '2', type: 'node', level: 'info', message: 'Memory > 70%', createdAt: new Date().toISOString() },
  ],
  cpu: [30, 35, 42, 38, 45, 52, 48, 55, 60, 58, 62, 55],
  memory: [50, 52, 55, 58, 60, 62, 65, 63, 68, 70, 72, 68],
}

const MonitoringPage: React.FC = () => {
  const [clusters, setClusters] = useState<any[]>([])
  const [selectedCluster, setSelectedCluster] = useState<string>('')
  const [loading, setLoading] = useState(false)
  const [data, setData] = useState(mockData)

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
      console.error('Error:', err)
    }
  }

  useEffect(() => {
    if (selectedCluster) {
      loadData()
    }
  }, [selectedCluster])

  const loadData = async () => {
    setLoading(true)
    try {
      const [statusRes, alertsRes] = await Promise.all([
        monitoringApi.getStatus(selectedCluster),
        monitoringApi.getAlerts(selectedCluster),
      ])
      
      // 如果 API 有数据就使用，否则用 mock
      setData({
        status: statusRes.data.active ? statusRes.data : mockData.status,
        alerts: alertsRes.data?.length > 0 ? alertsRes.data : mockData.alerts,
        cpu: mockData.cpu,
        memory: mockData.memory,
      })
    } catch {
      setData(mockData)
    } finally {
      setLoading(false)
    }
  }

  const getAlertColor = (level: string) => {
    if (level === 'critical') return 'border-red-500 bg-red-50'
    if (level === 'warning') return 'border-yellow-500 bg-yellow-50'
    return 'border-blue-500 bg-blue-50'
  }

  // 简单的柱状图
  const renderBars = (values: number[], color: string) => (
    <div className="h-48 flex items-end space-x-1">
      {values.map((v, i) => (
        <div key={i} className="flex-1 flex flex-col items-center group">
          <div className="opacity-0 group-hover:opacity-100 text-xs mb-1">{v}%</div>
          <div 
            className={`w-full ${color} rounded-t`}
            style={{ height: `${v}%` }}
          />
        </div>
      ))}
    </div>
  )

  return (
    <div>
      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-center justify-between mb-6 gap-4">
        <h1 className="text-2xl font-bold">Monitoring</h1>
        <div className="flex items-center space-x-4">
          <select
            value={selectedCluster}
            onChange={(e) => setSelectedCluster(e.target.value)}
            className="input"
          >
            <option value="">Select Cluster</option>
            {clusters.map((c: any) => (
              <option key={c.name} value={c.name}>{c.name}</option>
            ))}
          </select>
          <button onClick={loadData} className="btn btn-secondary flex items-center space-x-2">
            <RefreshCw className="h-4 w-4" />
            <span>Refresh</span>
          </button>
        </div>
      </div>

      {loading ? (
        <div className="flex items-center justify-center min-h-[40vh]">
          <Loader2 className="h-8 w-8 animate-spin text-primary-600" />
        </div>
      ) : (
        <div className="space-y-6">
          {/* Charts */}
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <div className="card p-6">
              <h2 className="text-lg font-semibold mb-4 flex items-center space-x-2">
                <Activity className="h-5 w-5 text-blue-600" />
                <span>CPU Usage</span>
              </h2>
              {renderBars(data.cpu, 'bg-blue-500')}
            </div>

            <div className="card p-6">
              <h2 className="text-lg font-semibold mb-4 flex items-center space-x-2">
                <Activity className="h-5 w-5 text-green-600" />
                <span>Memory Usage</span>
              </h2>
              {renderBars(data.memory, 'bg-green-500')}
            </div>
          </div>

          {/* Alerts */}
          <div className="card p-6">
            <h2 className="text-lg font-semibold mb-4 flex items-center space-x-2">
              <AlertCircle className="h-5 w-5 text-yellow-600" />
              <span>Alerts ({data.alerts.length})</span>
            </h2>
            <div className="space-y-3">
              {data.alerts.map((alert: any) => (
                <div 
                  key={alert.id} 
                  className={`rounded-lg p-4 border-l-4 ${getAlertColor(alert.level)}`}
                >
                  <div className="flex items-center justify-between">
                    <span className="font-medium">{alert.message}</span>
                    <span className="text-sm text-gray-500">{formatDate(alert.createdAt)}</span>
                  </div>
                  <span className="text-sm text-gray-600 capitalize">{alert.type} - {alert.level}</span>
                </div>
              ))}
            </div>
          </div>

          {/* Status */}
          <div className="card p-6">
            <h2 className="text-lg font-semibold mb-4">Status</h2>
            <div className="grid grid-cols-3 gap-4">
              <div className="bg-gray-50 rounded-lg p-4 text-center">
                <div className="text-sm text-gray-500">Status</div>
                <div className="text-lg font-semibold text-green-600">
                  {data.status.active ? 'Active' : 'Inactive'}
                </div>
              </div>
              <div className="bg-gray-50 rounded-lg p-4 text-center">
                <div className="text-sm text-gray-500">Data Points</div>
                <div className="text-lg font-semibold">{data.status.dataPoints}</div>
              </div>
              <div className="bg-gray-50 rounded-lg p-4 text-center">
                <div className="text-sm text-gray-500">Cluster</div>
                <div className="text-lg font-semibold">{data.status.cluster}</div>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

export default MonitoringPage
