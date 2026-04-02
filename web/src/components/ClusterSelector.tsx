interface ClusterSelectorProps {
  clusters: string[]
  selected: string
  onSelect: (cluster: string) => void
}

export function ClusterSelector({ clusters, selected, onSelect }: ClusterSelectorProps) {
  return (
    <div className="flex items-center gap-2">
      <label className="text-sm font-medium text-slate-700 dark:text-slate-300">
        Cluster:
      </label>
      <select
        value={selected}
        onChange={(e) => onSelect(e.target.value)}
        className="px-3 py-1.5 text-sm border border-slate-300 dark:border-slate-600 rounded-md bg-white dark:bg-slate-700 text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-primary-500"
      >
        {clusters.length === 0 ? (
          <option value="">No clusters</option>
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
