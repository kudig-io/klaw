import React, { useState, useEffect } from 'react'
import { alertingApi, clusterApi, monitoringApi, type AlertRecord, type AlertRule, type AlertStats } from '../lib/api'
import { cn, formatDate } from '../lib/utils'
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
    if (resolved) return 'border-gray-400 bg-gray-50 dark:border-gray-600 dark:bg-gray-800'
    if (level === 'critical' || level === 'error') return 'border-danger-500 bg-danger-50 dark:bg-danger-950/40'
    if (level === 'warning') return 'border-warning-500 bg-warning-50 dark:bg-warning-950/40'
    return 'border-info-500 bg-info-50 dark:bg-info-950/40'
  }

  return (
    <div>
      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-center justify-between mb-6 gap-4">
        <h1 className="text-2xl font-bold">监控告警</h1>
        <div className="flex items-center space-x-4">
          <select
            value={selectedCluster}
            onChange={(e) => setSelectedCluster(e.target.value)}
            className="input w-44 shrink-0"
          >
            <option value="">选择集群</option>
            {clusters.map((c: any) => (
              <option key={c.name} value={c.name}>{c.name}</option>
            ))}
          </select>
          <button onClick={loadData} className="btn btn-secondary flex items-center space-x-2 whitespace-nowrap">
            <RefreshCw className="h-4 w-4" />
            <span>刷新</span>
          </button>
          <button onClick={evaluateAlerts} className="btn btn-primary flex items-center space-x-2 whitespace-nowrap">
            <Siren className="h-4 w-4" />
            <span>评估规则</span>
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
                <Activity className="h-5 w-5 text-primary-600" />
                <span>告警统计</span>
              </h2>
              <div className="grid grid-cols-2 gap-4">
                <div className="bg-gray-50 dark:bg-gray-800/60 rounded-lg p-4">
                  <div className="text-sm text-gray-500 dark:text-gray-400">总计</div>
                  <div className="text-2xl font-semibold font-mono">{stats?.total ?? 0}</div>
                </div>
                <div className={cn('rounded-lg p-4', (stats?.active ?? 0) > 0 ? 'bg-danger-50 dark:bg-danger-950/40' : 'bg-gray-50 dark:bg-gray-800/60')}>
                  <div className="text-sm text-gray-500 dark:text-gray-400">进行中</div>
                  <div className="text-2xl font-semibold font-mono">{stats?.active ?? 0}</div>
                </div>
                <div className="bg-gray-50 dark:bg-gray-800/60 rounded-lg p-4">
                  <div className="text-sm text-gray-500 dark:text-gray-400">近 24 小时</div>
                  <div className="text-2xl font-semibold font-mono">{stats?.recent24h ?? 0}</div>
                </div>
                <div className="bg-gray-50 dark:bg-gray-800/60 rounded-lg p-4">
                  <div className="text-sm text-gray-500 dark:text-gray-400">规则数</div>
                  <div className="text-2xl font-semibold font-mono">{rules.length}</div>
                </div>
              </div>
              <div className="mt-4 pt-4 border-t border-gray-200 dark:border-gray-700 space-y-2">
                <div className="flex items-center gap-3 flex-wrap text-sm">
                  <span className="text-gray-500 dark:text-gray-400">未解决分级：</span>
                  <span className="font-medium text-danger-600 dark:text-danger-400">critical <span className="font-mono">{stats?.bySeverity?.critical ?? 0}</span></span>
                  <span className="font-medium text-warning-600 dark:text-warning-400">error <span className="font-mono">{stats?.bySeverity?.error ?? 0}</span></span>
                  <span className="font-medium text-gray-600 dark:text-gray-400">warning <span className="font-mono">{stats?.bySeverity?.warning ?? 0}</span></span>
                </div>
                <div className="flex items-center gap-3 flex-wrap text-sm">
                  <span className="text-gray-500">状态分布：</span>
                  <span className="font-medium">未确认 <span className="font-mono">{stats?.byStatus?.active ?? 0}</span></span>
                  <span className="font-medium">已确认 <span className="font-mono">{stats?.byStatus?.acknowledged ?? 0}</span></span>
                  <span className="font-medium text-success-600 dark:text-success-400">已解决 <span className="font-mono">{stats?.byStatus?.resolved ?? 0}</span></span>
                </div>
              </div>
            </div>

            <div className="card p-6">
              <h2 className="text-lg font-semibold mb-4 flex items-center space-x-2">
                <Clock className="h-5 w-5 text-success-600" />
                <span>规则覆盖</span>
              </h2>
              <div className="space-y-3">
                {rules.map((rule) => (
                  <div key={rule.id} className="rounded-lg border border-gray-200 dark:border-gray-700 p-3">
                    <div className="flex items-center justify-between gap-2">
                      <span className="font-medium">{rule.name}</span>
                      <div className="flex items-center gap-2">
                        <span
                          className={`px-1.5 py-0.5 rounded text-xs ${
                            rule.enabled
                              ? 'bg-success-100 text-success-700 dark:bg-success-900/30 dark:text-success-400'
                              : 'bg-gray-100 text-gray-500 dark:bg-gray-800 dark:text-gray-400'
                          }`}
                        >
                          {rule.enabled ? '启用' : '停用'}
                        </span>
                        <span className="text-xs uppercase text-gray-500 dark:text-gray-400">{rule.severity}</span>
                      </div>
                    </div>
                    <div className="text-sm text-gray-600 dark:text-gray-400 mt-1">{rule.description}</div>
                    <div className="text-sm text-gray-600 dark:text-gray-400 mt-1 font-mono">
                      {rule.condition.type}.{rule.condition.field} {rule.condition.operator} {String(rule.condition.threshold)}
                      <span className="text-gray-400 font-sans"> · 窗口 {rule.condition.timeWindow}</span>
                    </div>
                    {(rule.actions?.length ?? 0) > 0 && (
                      <div className="mt-2 flex items-center gap-1.5 flex-wrap">
                        {rule.actions?.map((action) => (
                          <span
                            key={action}
                            className="px-1.5 py-0.5 rounded bg-gray-100 dark:bg-gray-800 text-xs font-mono text-gray-600 dark:text-gray-400"
                          >
                            {action}
                          </span>
                        ))}
                      </div>
                    )}
                    <div className="mt-1.5 text-xs text-gray-500 dark:text-gray-500">
                      创建 {formatDate(rule.createdAt)} · 更新 {formatDate(rule.updatedAt)}
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>

          {lastTriggered.length > 0 && (
            <div className="card p-6">
              <h2 className="text-lg font-semibold mb-4 flex items-center space-x-2">
                <AlertTriangle className="h-5 w-5 text-warning-600" />
                <span>本次评估触发 {lastTriggered.length} 条告警</span>
              </h2>
              <div className="space-y-2">
                {lastTriggered.map((alert) => (
                  <div key={alert.id} className="rounded-lg p-3 border border-warning-200 dark:border-warning-800/50 bg-warning-50 dark:bg-warning-950/40">
                    <div className="font-medium">{alert.ruleName}</div>
                    <div className="text-sm text-gray-600 dark:text-gray-400">{alert.resourceKind} {alert.namespace ? `${alert.namespace}/` : ''}{alert.resourceName}</div>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Alerts */}
          <div className="card p-6">
            <h2 className="text-lg font-semibold mb-4 flex items-center space-x-2">
              <AlertCircle className="h-5 w-5 text-warning-600" />
              <span>告警历史（{alerts.length}）</span>
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
                      <div className="text-sm text-gray-600 dark:text-gray-400 mt-1">
                        {alert.ruleName} · {alert.resourceKind} · {alert.namespace ? `${alert.namespace}/` : ''}{alert.resourceName}
                      </div>
                      <div className="text-sm text-gray-700 dark:text-gray-300 mt-1 font-mono">
                        触发值 {alert.value} {alert.operator} 阈值 {alert.threshold}
                      </div>
                    </div>
                    <span className="text-sm text-gray-500 dark:text-gray-400 whitespace-nowrap">{formatDate(alert.createdAt)}</span>
                  </div>
                  <div className="mt-3 flex items-center gap-2 flex-wrap">
                    <span className="text-sm text-gray-600 dark:text-gray-400 capitalize">{alert.ruleType} - {alert.severity}</span>
                    {alert.acknowledged && (
                      <span className="px-1.5 py-0.5 rounded bg-info-100 text-info-700 dark:bg-info-900/30 dark:text-info-400 text-xs">
                        已确认{alert.acknowledgedAt ? ` · ${formatDate(alert.acknowledgedAt)}` : ''}
                      </span>
                    )}
                    {alert.resolved && (
                      <span className="px-1.5 py-0.5 rounded bg-success-100 text-success-700 dark:bg-success-900/30 dark:text-success-400 text-xs">
                        已解决{alert.resolvedAt ? ` · ${formatDate(alert.resolvedAt)}` : ''}
                      </span>
                    )}
                    {!alert.acknowledged && !alert.resolved && (
                      <button onClick={() => acknowledgeAlert(alert.id)} className="btn btn-secondary text-xs">
                        确认
                      </button>
                    )}
                    {!alert.resolved && (
                      <button onClick={() => resolveAlert(alert.id)} className="btn btn-secondary text-xs flex items-center space-x-1">
                        <CheckCircle2 className="h-3 w-3" />
                        <span>解决</span>
                      </button>
                    )}
                  </div>
                </div>
              ))}
            </div>
          </div>

          {/* Status */}
          <div className="card p-6">
            <h2 className="text-lg font-semibold mb-4">运行状态</h2>
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
              <div className="bg-gray-50 dark:bg-gray-800/60 rounded-lg p-4 text-center">
                <div className="text-sm text-gray-500 dark:text-gray-400">状态</div>
                <div className="text-lg font-semibold text-success-600 dark:text-success-400">
                  {status.active ? '运行中' : '未启用'}
                </div>
              </div>
              <div className="bg-gray-50 dark:bg-gray-800/60 rounded-lg p-4 text-center">
                <div className="text-sm text-gray-500 dark:text-gray-400">数据点</div>
                <div className="text-lg font-semibold">{status.dataPoints}</div>
              </div>
              <div className="bg-gray-50 dark:bg-gray-800/60 rounded-lg p-4 text-center">
                <div className="text-sm text-gray-500 dark:text-gray-400">集群</div>
                <div className="text-lg font-semibold">{status.cluster}</div>
              </div>
              <div className="bg-gray-50 dark:bg-gray-800/60 rounded-lg p-4 text-center">
                <div className="text-sm text-gray-500 dark:text-gray-400">采集间隔</div>
                <div className="text-lg font-semibold">{status.interval ?? '-'}</div>
              </div>
              <div className="bg-gray-50 dark:bg-gray-800/60 rounded-lg p-4 text-center">
                <div className="text-sm text-gray-500 dark:text-gray-400">评估间隔</div>
                <div className="text-lg font-semibold">{status.evalInterval ?? '-'}</div>
              </div>
              <div className="bg-gray-50 dark:bg-gray-800/60 rounded-lg p-4 text-center">
                <div className="text-sm text-gray-500 dark:text-gray-400">启用规则</div>
                <div className="text-lg font-semibold">
                  {status.rulesEnabled ?? '-'}/{status.rulesTotal ?? '-'}
                </div>
              </div>
              <div className="bg-gray-50 dark:bg-gray-800/60 rounded-lg p-4 text-center col-span-2">
                <div className="text-sm text-gray-500 dark:text-gray-400">最近评估</div>
                <div className="text-lg font-semibold">{status.lastEvaluation ? formatDate(status.lastEvaluation) : '-'}</div>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

export default MonitoringPage
