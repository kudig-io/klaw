import { useState } from 'react'
import { v1Api } from '../lib/api'
import { Activity, AlertTriangle, AlertCircle, Info, Loader2, Search } from 'lucide-react'

interface Issue {
  severity: string
  cn_name?: string
  en_name?: string
  details?: string
  location?: string
  analyzer_name?: string
  remediation?: { suggestion?: string; command?: string }
}

interface DiagResult {
  data?: { nodeName?: string }
  results?: unknown[]
  issues?: Issue[]
}

const severityColors: Record<string, string> = {
  CRITICAL: 'text-red-600 bg-red-50 dark:bg-red-900/20 dark:text-red-400',
  WARNING: 'text-yellow-600 bg-yellow-50 dark:bg-yellow-900/20 dark:text-yellow-400',
  INFO: 'text-blue-600 bg-blue-50 dark:bg-blue-900/20 dark:text-blue-400',
}

function severityIcon(sev: string) {
  if (sev === 'CRITICAL') return <AlertCircle className="h-5 w-5 text-red-500" />
  if (sev === 'WARNING') return <AlertTriangle className="h-5 w-5 text-yellow-500" />
  return <Info className="h-5 w-5 text-blue-500" />
}

export default function DiagnosticsPage() {
  const [loading, setLoading] = useState(false)
  const [result, setResult] = useState<DiagResult | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [node, setNode] = useState('')

  const runDiag = async () => {
    setLoading(true)
    setError(null)
    try {
      const params: Record<string, string> = {}
      if (node) params.node = node
      const res = await v1Api.get('/diag/run', { params })
      setResult(res.data)
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : '诊断请求失败'
      setError(msg)
    } finally {
      setLoading(false)
    }
  }

  const issues = result?.issues ?? []
  const critical = issues.filter((i) => i.severity === 'CRITICAL' || i.severity === '1').length
  const warning = issues.filter((i) => i.severity === 'WARNING' || i.severity === '2').length
  const info = issues.filter((i) => i.severity === 'INFO' || i.severity === '3').length

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-gray-900 dark:text-white">集群诊断</h2>
          <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
            70+ 分析器 · 内核 / 网络 / 存储 / 安全 / 服务网格
          </p>
        </div>
      </div>

      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
        <div className="flex items-center space-x-3">
          <div className="flex-1">
            <input
              type="text"
              value={node}
              onChange={(e) => setNode(e.target.value)}
              placeholder="节点名称 (可选, 留空诊断整个集群)"
              className="w-full px-4 py-2 border border-gray-300 dark:border-gray-700 rounded-md dark:bg-gray-900 dark:text-white focus:ring-2 focus:ring-primary-500"
            />
          </div>
          <button
            onClick={runDiag}
            disabled={loading}
            className="flex items-center space-x-2 px-6 py-2 bg-primary-600 text-white rounded-md hover:bg-primary-700 disabled:opacity-50 transition-colors"
          >
            {loading ? <Loader2 className="h-4 w-4 animate-spin" /> : <Search className="h-4 w-4" />}
            <span>{loading ? '诊断中...' : '运行诊断'}</span>
          </button>
        </div>
        {error && (
          <div className="mt-4 p-3 bg-red-50 dark:bg-red-900/20 text-red-600 dark:text-red-400 rounded-md text-sm">
            {error}
          </div>
        )}
      </div>

      {result && (
        <>
          <div className="grid grid-cols-3 gap-4">
            <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6 text-center">
              <div className="text-3xl font-bold text-red-600 dark:text-red-400">{critical}</div>
              <div className="text-sm text-gray-500 dark:text-gray-400 mt-1">严重</div>
            </div>
            <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6 text-center">
              <div className="text-3xl font-bold text-yellow-600 dark:text-yellow-400">{warning}</div>
              <div className="text-sm text-gray-500 dark:text-gray-400 mt-1">警告</div>
            </div>
            <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6 text-center">
              <div className="text-3xl font-bold text-blue-600 dark:text-blue-400">{info}</div>
              <div className="text-sm text-gray-500 dark:text-gray-400 mt-1">信息</div>
            </div>
          </div>

          <div className="bg-white dark:bg-gray-800 rounded-lg shadow">
            <div className="px-6 py-4 border-b border-gray-200 dark:border-gray-700">
              <h3 className="font-semibold text-gray-900 dark:text-white flex items-center">
                <Activity className="h-5 w-5 mr-2" />
                问题详情 ({issues.length})
              </h3>
            </div>
            {issues.length === 0 ? (
              <div className="p-12 text-center text-gray-500 dark:text-gray-400">
                ✅ 未发现问题
              </div>
            ) : (
              <div className="divide-y divide-gray-200 dark:divide-gray-700">
                {issues.map((issue, i) => {
                  const sev = String(issue.severity).includes('1') || issue.severity === 'CRITICAL'
                    ? 'CRITICAL'
                    : String(issue.severity).includes('2') || issue.severity === 'WARNING'
                    ? 'WARNING'
                    : 'INFO'
                  return (
                    <div key={i} className="p-4 flex items-start space-x-3">
                      <div className="mt-0.5">{severityIcon(sev)}</div>
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center space-x-2">
                          <span className={`px-2 py-0.5 text-xs font-medium rounded ${severityColors[sev]}`}>
                            {sev}
                          </span>
                          <span className="font-medium text-gray-900 dark:text-white">
                            {issue.cn_name || issue.en_name || '未命名问题'}
                          </span>
                        </div>
                        {issue.details && (
                          <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">{issue.details}</p>
                        )}
                        {issue.remediation?.suggestion && (
                          <p className="text-sm text-green-600 dark:text-green-400 mt-1">
                            💡 {issue.remediation.suggestion}
                          </p>
                        )}
                        {issue.analyzer_name && (
                          <p className="text-xs text-gray-400 mt-1">分析器: {issue.analyzer_name}</p>
                        )}
                      </div>
                    </div>
                  )
                })}
              </div>
            )}
          </div>
        </>
      )}
    </div>
  )
}
