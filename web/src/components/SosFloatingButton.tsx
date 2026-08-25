import { Link, useLocation } from 'react-router-dom'
import { Siren } from 'lucide-react'
import { cn } from '../lib/utils'

// SosFloatingButton 全局右下角 SOS 应急入口；通话页内隐藏
export default function SosFloatingButton() {
  const { pathname } = useLocation()
  if (pathname === '/sos') return null
  return (
    <Link
      to="/sos"
      aria-label="SOS"
      className={cn(
        'fixed bottom-6 right-6 z-50 flex h-14 w-14 items-center justify-center',
        'rounded-full bg-red-600 text-white shadow-lg shadow-red-600/40',
        'transition-transform hover:scale-110 active:scale-95',
      )}
    >
      <Siren className="h-6 w-6" />
    </Link>
  )
}
