interface ClusterSelectorProps {
  clusters: string[]
  selected: string
  onSelect: (cluster: string) => void
}

export function ClusterSelector({ clusters, selected, onSelect }: ClusterSelectorProps) {
  return (
    <div className="flex items-center gap-2">
      <label className="text-sm font-medium text-gray-700 dark:text-gray-300">
        集群：
      </label>
      <select
        value={selected}
        onChange={(e) => onSelect(e.target.value)}
        className="px-3 py-1.5 text-sm border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-primary-500"
      >
        {clusters.length === 0 ? (
          <option value="">暂无集群</option>
        ) : (
          clusters.map((cluster) => (
            <option key={cluster} value={cluster}>
              {cluster}
            </option>
          ))
        )}
      </select>
    </div>
  )
}
