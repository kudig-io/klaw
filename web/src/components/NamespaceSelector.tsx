import { useState, useEffect } from 'react'
import { clusterApi } from '../lib/api'

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

  useEffect(() => {
    if (cluster) {
      loadNamespaces()
    }
  }, [cluster])

  async function loadNamespaces() {
    setIsLoading(true)
    try {
      const response = await clusterApi.getNamespaces(cluster)
      const nsList = response.data.map((ns: any) => ns.metadata?.name || ns.name).sort()
      setNamespaces(nsList)
    } catch (error) {
      console.error('Failed to load namespaces:', error)
      setNamespaces(['default'])
    } finally {
      setIsLoading(false)
    }
  }

  const handleChange = (value: string) => {
    // Convert '_all' back to empty string for API
    onSelect(value === ALL_NAMESPACES ? '' : value)
  }

  return (
    <div className="flex items-center gap-2">
      <label className="text-sm font-medium text-slate-700 dark:text-slate-300">
        Namespace:
      </label>
      <select
        value={selected || ALL_NAMESPACES}
        onChange={(e) => handleChange(e.target.value)}
        disabled={isLoading || !cluster}
        className="px-3 py-1.5 text-sm border border-slate-300 dark:border-slate-600 rounded-md bg-white dark:bg-slate-700 text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-primary-500 disabled:opacity-50"
      >
        {showAllNamespaces && (
          <option value={ALL_NAMESPACES}>All Namespaces</option>
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
