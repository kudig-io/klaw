import React, { useState, useEffect } from 'react'
import { alertingApi, clusterApi, monitoringApi, type AlertRecord, type AlertRule, type AlertStats } from '../lib/api'
import { formatDate } from '../lib/utils'
import { RefreshCw, Loader2, AlertCircle, Activity, Clock, AlertTriangle, Siren, CheckCircle2 } from 'lucide-react'

const MonitoringPage: React.FC = () => {
  const [clusters, setClusters] = useState<any[]>([])
  const [selectedCluster, setSelectedCluster] = useState<string>('')
  const [loading, setLoading] = useState(false)
  const [status, setStatus] = useState<any>({ active: false, cluster: '', dataPoints: 0 })
  const [alerts, setAlerts] = useState<AlertRecord[]>([])
  const [rules, setRules] = useState<AlertRule[]>([])
  const [stats, setStats] = useState<AlertStats | null>(null)
  const [lastTriggered, setLastTriggered] = useState<AlertRecord[]>([])

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
      const [statusRes, historyRes, statsRes, rulesRes] = await Promise.all([
        monitoringApi.getStatus(selectedCluster),
        alertingApi.getHistory(selectedCluster, 20),
        alertingApi.getStats(selectedCluster),
        alertingApi.getRules(selectedCluster),
      ])

      setStatus(statusRes.data)
      setAlerts(historyRes.data)
      setStats(statsRes.data)
      setRules(rulesRes.data)
    } catch (err) {
      console.error(err)
    } finally {
      setLoading(false)
    }
  }

  const evaluateAlerts = async () => {
    if (!selectedCluster) return
    setLoading(true)
    try {
      const response = await alertingApi.evaluate(selectedCluster)
      setLastTriggered(response.data)
      await loadData()
    } finally {
      setLoading(false)
    }
  }

  const acknowledgeAlert = async (alertId: string) => {
    await alertingApi.acknowledge(selectedCluster, alertId)
    await loadData()
  }

  const resolveAlert = async (alertId: string) => {
    await alertingApi.resolve(selectedCluster, alertId)
    await loadData()
  }

  const getAlertColor = (level: string, resolved?: boolean) => {
    if (resolved) return 'border-gray-400 bg-gray-50'
    if (level === 'critical' || level === 'error') return 'border-red-500 bg-red-50'
    if (level === 'warning') return 'border-yellow-500 bg-yellow-50'
    return 'border-blue-500 bg-blue-50'
  }

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
          <button onClick={evaluateAlerts} className="btn btn-primary flex items-center space-x-2">
            <Siren className="h-4 w-4" />
            <span>Evaluate Rules</span>
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
                <span>Alert Stats</span>
              </h2>
              <div className="grid grid-cols-2 gap-4">
                <div className="bg-blue-50 rounded-lg p-4">
                  <div className="text-sm text-gray-500">Total</div>
                  <div className="text-2xl font-semibold">{stats?.total ?? 0}</div>
                </div>
                <div className="bg-red-50 rounded-lg p-4">
                  <div className="text-sm text-gray-500">Active</div>
                  <div className="text-2xl font-semibold">{stats?.active ?? 0}</div>
                </div>
                <div className="bg-yellow-50 rounded-lg p-4">
                  <div className="text-sm text-gray-500">Recent 24h</div>
                  <div className="text-2xl font-semibold">{stats?.recent24h ?? 0}</div>
                </div>
                <div className="bg-green-50 rounded-lg p-4">
                  <div className="text-sm text-gray-500">Rules</div>
                  <div className="text-2xl font-semibold">{rules.length}</div>
                </div>
              </div>
            </div>

            <div className="card p-6">
              <h2 className="text-lg font-semibold mb-4 flex items-center space-x-2">
                <Clock className="h-5 w-5 text-green-600" />
                <span>Rule Coverage</span>
              </h2>
              <div className="space-y-3">
                {rules.slice(0, 5).map((rule) => (
                  <div key={rule.id} className="rounded-lg border border-gray-200 p-3">
                    <div className="flex items-center justify-between">
                      <span className="font-medium">{rule.name}</span>
                      <span className="text-xs uppercase text-gray-500">{rule.severity}</span>
                    </div>
                    <div className="text-sm text-gray-600 mt-1">
                      {rule.condition.type}.{rule.condition.field} {rule.condition.operator} {String(rule.condition.threshold)}
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>

          {lastTriggered.length > 0 && (
            <div className="card p-6">
              <h2 className="text-lg font-semibold mb-4 flex items-center space-x-2">
                <AlertTriangle className="h-5 w-5 text-orange-600" />
                <span>Last Evaluation Triggered {lastTriggered.length} Alerts</span>
              </h2>
              <div className="space-y-2">
                {lastTriggered.map((alert) => (
                  <div key={alert.id} className="rounded-lg p-3 border border-orange-200 bg-orange-50">
                    <div className="font-medium">{alert.ruleName}</div>
                    <div className="text-sm text-gray-600">{alert.resourceKind} {alert.namespace ? `${alert.namespace}/` : ''}{alert.resourceName}</div>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Alerts */}
          <div className="card p-6">
            <h2 className="text-lg font-semibold mb-4 flex items-center space-x-2">
              <AlertCircle className="h-5 w-5 text-yellow-600" />
              <span>Alert History ({alerts.length})</span>
            </h2>
            <div className="space-y-3">
              {alerts.map((alert) => (
                <div 
                  key={alert.id} 
                  className={`rounded-lg p-4 border-l-4 ${getAlertColor(alert.severity, alert.resolved)}`}
                >
                  <div className="flex items-start justify-between gap-4">
                    <div>
                      <div className="font-medium">{alert.message}</div>
                      <div className="text-sm text-gray-600 mt-1">
                        {alert.ruleName} · {alert.resourceKind} · {alert.namespace ? `${alert.namespace}/` : ''}{alert.resourceName}
                      </div>
                    </div>
                    <span className="text-sm text-gray-500 whitespace-nowrap">{formatDate(alert.createdAt)}</span>
                  </div>
                  <div className="mt-3 flex items-center gap-2 flex-wrap">
                    <span className="text-sm text-gray-600 capitalize">{alert.ruleType} - {alert.severity}</span>
                    {!alert.acknowledged && !alert.resolved && (
                      <button onClick={() => acknowledgeAlert(alert.id)} className="btn btn-secondary text-xs">
                        Acknowledge
                      </button>
                    )}
                    {!alert.resolved && (
                      <button onClick={() => resolveAlert(alert.id)} className="btn btn-secondary text-xs flex items-center space-x-1">
                        <CheckCircle2 className="h-3 w-3" />
                        <span>Resolve</span>
                      </button>
                    )}
                  </div>
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
                  {status.active ? 'Active' : 'Inactive'}
                </div>
              </div>
              <div className="bg-gray-50 rounded-lg p-4 text-center">
                <div className="text-sm text-gray-500">Data Points</div>
                <div className="text-lg font-semibold">{status.dataPoints}</div>
              </div>
              <div className="bg-gray-50 rounded-lg p-4 text-center">
                <div className="text-sm text-gray-500">Cluster</div>
                <div className="text-lg font-semibold">{status.cluster}</div>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

export default MonitoringPage
