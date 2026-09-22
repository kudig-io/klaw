import { LockKeyhole } from 'lucide-react'
import { useAuth } from '../contexts/AuthContext'

// 顶栏凭证入口：仅在后端开启认证且已通过时显示
export default function AuthStatusButton() {
  const { status, authEnabled, openGate } = useAuth()
  if (status !== 'authenticated' || !authEnabled) return null
  return (
    <button
      onClick={openGate}
      title="管理访问凭证"
      aria-label="管理访问凭证"
      className="p-2 rounded-md text-gray-500 hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-gray-800 transition-colors duration-150"
    >
      <LockKeyhole className="h-4 w-4" />
    </button>
  )
}
