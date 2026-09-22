import { useState, useEffect, useCallback } from 'react'
import { clusterApi, type Namespace } from '../lib/api'

// 兼容历史接口：部分响应可能直接返回 { name } 而非完整的 Namespace 对象
type NamespaceLike = Namespace & { name?: string }

interface NamespaceSelectorProps {
  cluster: string
  selected: string
  onSelect: (namespace: string) => void
  showAllNamespaces?: boolean
}

const ALL_NAMESPACES = '_all'

export function NamespaceSelector({
  cluster,
  selected,
  onSelect,
  showAllNamespaces = false
}: NamespaceSelectorProps) {
  const [namespaces, setNamespaces] = useState<string[]>([])
  const [isLoading, setIsLoading] = useState(false)

  const loadNamespaces = useCallback(async () => {
    setIsLoading(true)
    try {
      const response = await clusterApi.getNamespaces(cluster)
      // 历史接口兜底：metadata.name 缺失时回退到顶层 name（缺失项保持 undefined 原样，不做过滤）
      const nsList = response.data.map((ns: NamespaceLike) => ns.metadata?.name || ns.name).sort() as string[]
      setNamespaces(nsList)
    } catch (error) {
      console.error('Failed to load namespaces:', error)
      setNamespaces(['default'])
    } finally {
      setIsLoading(false)
    }
  }, [cluster])

  useEffect(() => {
    if (cluster) {
      loadNamespaces()
    }
  }, [cluster, loadNamespaces])

  const handleChange = (value: string) => {
    // Convert '_all' back to empty string for API
    onSelect(value === ALL_NAMESPACES ? '' : value)
  }

  return (
    <div className="flex items-center gap-2">
      <label className="text-sm font-medium text-gray-700 dark:text-gray-300">
        命名空间：
      </label>
      <select
        value={selected || ALL_NAMESPACES}
        onChange={(e) => handleChange(e.target.value)}
        disabled={isLoading || !cluster}
        className="px-3 py-1.5 text-sm border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-primary-500 disabled:opacity-50"
      >
        {showAllNamespaces && (
          <option value={ALL_NAMESPACES}>全部命名空间</option>
        )}
        {namespaces.map((ns) => (
          <option key={ns} value={ns}>
            {ns}
          </option>
        ))}
      </select>
    </div>
  )
}
