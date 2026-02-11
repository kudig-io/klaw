import { type ClassValue, clsx } from 'clsx'
import { twMerge } from 'tailwind-merge'

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

export function formatDate(dateString: string): string {
  const date = new Date(dateString)
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

export function formatMemory(bytes: string): string {
  const value = parseInt(bytes)
  if (value >= 1024 * 1024 * 1024) {
    return `${(value / (1024 * 1024 * 1024)).toFixed(2)} Gi`
  } else if (value >= 1024 * 1024) {
    return `${(value / (1024 * 1024)).toFixed(2)} Mi`
  } else if (value >= 1024) {
    return `${(value / 1024).toFixed(2)} Ki`
  }
  return `${value} B`
}

export function getStatusColor(status: string): string {
  const statusMap: Record<string, string> = {
    Running: 'bg-success-500',
    Pending: 'bg-warning-500',
    Failed: 'bg-danger-500',
    Succeeded: 'bg-success-500',
    Unknown: 'bg-gray-500',
    Ready: 'bg-success-500',
    NotReady: 'bg-danger-500',
  }
  return statusMap[status] || 'bg-gray-500'
}

export function getStatusTextColor(status: string): string {
  const statusMap: Record<string, string> = {
    Running: 'text-success-600',
    Pending: 'text-warning-600',
    Failed: 'text-danger-600',
    Succeeded: 'text-success-600',
    Unknown: 'text-gray-600',
    Ready: 'text-success-600',
    NotReady: 'text-danger-600',
  }
  return statusMap[status] || 'text-gray-600'
}
