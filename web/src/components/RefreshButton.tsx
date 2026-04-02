import { RefreshCw } from 'lucide-react'

interface RefreshButtonProps {
  onClick: () => void
  isLoading?: boolean
}

export function RefreshButton({ onClick, isLoading = false }: RefreshButtonProps) {
  return (
    <button
      onClick={onClick}
      disabled={isLoading}
      className="flex items-center gap-2 px-3 py-1.5 text-sm font-medium text-slate-600 dark:text-slate-300 hover:text-primary-600 dark:hover:text-primary-400 bg-white dark:bg-slate-800 border border-slate-300 dark:border-slate-600 rounded-md hover:border-primary-300 dark:hover:border-primary-500 transition-colors disabled:opacity-50"
    >
      <RefreshCw className={`w-4 h-4 ${isLoading ? 'animate-spin' : ''}`} />
      <span>Refresh</span>
    </button>
  )
}
