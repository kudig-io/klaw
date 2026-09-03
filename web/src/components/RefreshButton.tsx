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
      className="flex items-center gap-2 px-3 py-1.5 text-sm font-medium text-gray-600 dark:text-gray-300 hover:text-primary-600 dark:hover:text-primary-400 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-md hover:border-primary-300 dark:hover:border-primary-500 transition-colors disabled:opacity-50"
    >
      <RefreshCw className={`w-4 h-4 ${isLoading ? 'animate-spin' : ''}`} />
      <span>刷新</span>
    </button>
  )
}
