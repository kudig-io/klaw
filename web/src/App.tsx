import { useState, useEffect } from 'react'
import { Routes, Route, NavLink, useLocation } from 'react-router-dom'
import { cn } from './lib/utils'
import ClusterDashboard from './pages/ClusterDashboard'
import BackupsPage from './pages/BackupsPage'
import PodsPage from './pages/PodsPage'
import NodesPage from './pages/NodesPage'
import MonitoringPage from './pages/MonitoringPage'
import DeploymentsPage from './pages/DeploymentsPage'
import TenantsPage from './pages/TenantsPage'
import { ServicesPage } from './pages/ServicesPage'
import { NetworkPage } from './pages/NetworkPage'
import { StoragePage } from './pages/StoragePage'
import DiagnosticsPage from './pages/DiagnosticsPage'
import SosCallPage from './pages/SosCallPage'
import SosFloatingButton from './components/SosFloatingButton'
import { Menu, X, Moon, Sun, Database, Server, Activity, AlertCircle, Boxes, Beaker, Globe, Network, HardDrive, DatabaseBackup, Shield, Stethoscope, Siren, ExternalLink } from 'lucide-react'

/* Hallmark · genre: modern-minimal · macrostructure: Workbench · design-system: design.md · designed-as-app */

const GTM_URL = 'https://bs7klknl29np.meoo.fun'

const navGroups = [
  {
    label: '总览',
    items: [
      { path: '/', label: '仪表盘', icon: Database },
      { path: '/monitoring', label: '监控告警', icon: AlertCircle },
    ],
  },
  {
    label: '工作负载',
    items: [
      { path: '/deployments', label: '部署管理', icon: Boxes },
      { path: '/pods', label: '容器组（Pod）', icon: Server },
      { path: '/services', label: '服务（Service）', icon: Globe },
      { path: '/network', label: '网络（Network）', icon: Network },
      { path: '/nodes', label: '节点（Node）', icon: Activity },
    ],
  },
  {
    label: '平台能力',
    items: [
      { path: '/backups', label: '备份恢复', icon: DatabaseBackup },
      { path: '/storage', label: '存储（Storage）', icon: HardDrive },
      { path: '/tenants', label: '多租户', icon: Shield },
      { path: '/diagnostics', label: '智能诊断', icon: Stethoscope },
      { path: '/sos', label: '紧急求救（SOS）', icon: Siren },
    ],
  },
]

const navItems = navGroups.flatMap((group) => group.items)

function App() {
  const [isDarkMode, setIsDarkMode] = useState(() => document.documentElement.classList.contains('dark'))
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false)
  const [isMockMode, setIsMockMode] = useState(false)
  const location = useLocation()

  // 检查是否启用了 Mock 模式
  useEffect(() => {
    const mockEnabled = import.meta.env.VITE_USE_MOCK === 'true' || localStorage.getItem('USE_MOCK') === 'true'
    setIsMockMode(mockEnabled)
  }, [location]) // 路由变化时重新检查

  const toggleDarkMode = () => {
    const next = !isDarkMode
    document.documentElement.classList.toggle('dark', next)
    setIsDarkMode(next)
  }

  const toggleMockMode = () => {
    const newMockState = !isMockMode
    localStorage.setItem('USE_MOCK', newMockState ? 'true' : 'false')
    setIsMockMode(newMockState)
    // 刷新页面以应用更改
    window.location.reload()
  }

  const currentLabel = navItems.find((item) => item.path === location.pathname)?.label ?? '仪表盘'

  const navLinkClass = ({ isActive }: { isActive: boolean }) => cn(
    'flex items-center gap-2.5 px-3 py-2 rounded-md text-sm font-medium transition-colors duration-150',
    isActive
      ? 'bg-primary-50 text-primary-700 dark:bg-primary-900/30 dark:text-primary-300'
      : 'text-gray-600 hover:bg-gray-100 hover:text-gray-900 dark:text-gray-400 dark:hover:bg-gray-800 dark:hover:text-gray-200'
  )

  return (
    <div className="min-h-screen">
      <div className="flex min-h-screen">
        {/* 左侧 Workbench 导航栏 */}
        <aside className="hidden md:flex md:flex-col w-60 shrink-0 border-r border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900">
          <div className="h-14 flex items-center gap-2.5 px-5 border-b border-gray-200 dark:border-gray-800">
            <span className="w-6 h-6 rounded-md bg-primary-600 flex items-center justify-center">
              <svg viewBox="0 0 26 26" className="w-4 h-4" fill="none"><path d="M13 2c2 3 6 4 8 3-1 3-1 6 1 8-3 1-5 4-5 7-2-2-6-2-8 0 0-3-2-6-5-7 2-2 2-5 1-8 2 1 6 0 8-3z" fill="#fff"/></svg>
            </span>
            <h1 className="text-base font-semibold tracking-tight text-gray-900 dark:text-white">Klaw</h1>
            {isMockMode && (
              <span className="px-2 py-0.5 text-xs font-medium bg-warning-500/10 text-warning-600 dark:text-warning-500 rounded-full">
                MOCK
              </span>
            )}
          </div>

          <nav className="flex-1 overflow-y-auto px-3 py-4 space-y-5">
            {navGroups.map((group) => (
              <div key={group.label}>
                <div className="px-3 pb-1.5 text-[11px] font-medium uppercase tracking-wider text-gray-400 dark:text-gray-500">
                  {group.label}
                </div>
                <div className="space-y-0.5">
                  {group.items.map((item) => (
                    <NavLink key={item.path} to={item.path} className={navLinkClass}>
                      <item.icon className="h-4 w-4 shrink-0" />
                      <span>{item.label}</span>
                    </NavLink>
                  ))}
                </div>
              </div>
            ))}
          </nav>

          <div className="border-t border-gray-200 dark:border-gray-800 px-3 py-3 space-y-0.5">
            <a
              href={GTM_URL}
              target="_blank"
              rel="noreferrer"
              className="flex items-center gap-2.5 px-3 py-2 rounded-md text-sm font-medium text-gray-600 hover:bg-gray-100 hover:text-gray-900 dark:text-gray-400 dark:hover:bg-gray-800 dark:hover:text-gray-200 transition-colors duration-150"
            >
              <ExternalLink className="h-4 w-4 shrink-0" />
              <span>产品介绍 · GTM</span>
            </a>
            <div className="px-3 pt-2 text-xs text-gray-400 dark:text-gray-500">
              © 2024 Klaw · Kubernetes（K8s）智能运维
            </div>
          </div>
        </aside>

        {/* 右侧内容区 */}
        <div className="flex-1 min-w-0 flex flex-col">
          <header className="h-14 shrink-0 border-b border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900">
            <div className="h-full flex items-center justify-between px-4 sm:px-6 lg:px-8">
              <div className="flex items-center gap-3">
                {/* 移动端品牌 + 菜单按钮 */}
                <button
                  onClick={() => setIsMobileMenuOpen(!isMobileMenuOpen)}
                  className="md:hidden p-2 -ml-2 rounded-md text-gray-600 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-gray-800"
                >
                  {isMobileMenuOpen ? <X className="h-5 w-5" /> : <Menu className="h-5 w-5" />}
                </button>
                <span className="md:hidden flex items-center gap-2 text-base font-semibold tracking-tight text-gray-900 dark:text-white">
                  Klaw
                  {isMockMode && (
                    <span className="px-2 py-0.5 text-xs font-medium bg-warning-500/10 text-warning-600 dark:text-warning-500 rounded-full">
                      MOCK
                    </span>
                  )}
                </span>
                {/* 桌面端当前页面 */}
                <span className="hidden md:block text-sm font-medium text-gray-500 dark:text-gray-400">
                  {currentLabel}
                </span>
              </div>

              <div className="flex items-center gap-1">
                <button
                  onClick={toggleMockMode}
                  title={isMockMode ? '关闭 Mock 模式' : '开启 Mock 模式'}
                  className={cn(
                    'flex items-center gap-1.5 px-2.5 py-1.5 rounded-md text-sm font-medium transition-colors duration-150',
                    isMockMode
                      ? 'bg-warning-500/10 text-warning-600 dark:text-warning-500'
                      : 'text-gray-500 hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-gray-800'
                  )}
                >
                  <Beaker className="h-4 w-4" />
                  <span>Mock</span>
                </button>
                <button
                  onClick={toggleDarkMode}
                  className="p-2 rounded-md text-gray-500 hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-gray-800 transition-colors duration-150"
                >
                  {isDarkMode ? <Sun className="h-5 w-5" /> : <Moon className="h-5 w-5" />}
                </button>
              </div>
            </div>
          </header>

          {/* 移动端菜单 */}
          {isMobileMenuOpen && (
            <div className="md:hidden border-b border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900">
              <nav className="px-3 py-3 space-y-4">
                {navGroups.map((group) => (
                  <div key={group.label}>
                    <div className="px-3 pb-1 text-[11px] font-medium uppercase tracking-wider text-gray-400 dark:text-gray-500">
                      {group.label}
                    </div>
                    <div className="space-y-0.5">
                      {group.items.map((item) => (
                        <NavLink
                          key={item.path}
                          to={item.path}
                          className={navLinkClass}
                          onClick={() => setIsMobileMenuOpen(false)}
                        >
                          <item.icon className="h-4 w-4 shrink-0" />
                          <span>{item.label}</span>
                        </NavLink>
                      ))}
                    </div>
                  </div>
                ))}
                <a
                  href={GTM_URL}
                  target="_blank"
                  rel="noreferrer"
                  className="flex items-center gap-2.5 px-3 py-2 rounded-md text-sm font-medium text-gray-600 hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-gray-800"
                  onClick={() => setIsMobileMenuOpen(false)}
                >
                  <ExternalLink className="h-4 w-4 shrink-0" />
                  <span>产品介绍 · GTM</span>
                </a>
              </nav>
            </div>
          )}

          <main className="flex-1 px-4 sm:px-6 lg:px-8 py-8">
            <Routes>
              <Route path="/" element={<ClusterDashboard />} />
              <Route path="/backups" element={<BackupsPage />} />
              <Route path="/tenants" element={<TenantsPage />} />
              <Route path="/deployments" element={<DeploymentsPage />} />
              <Route path="/services" element={<ServicesPage />} />
              <Route path="/network" element={<NetworkPage />} />
              <Route path="/storage" element={<StoragePage />} />
              <Route path="/pods" element={<PodsPage />} />
              <Route path="/nodes" element={<NodesPage />} />
              <Route path="/monitoring" element={<MonitoringPage />} />
              <Route path="/diagnostics" element={<DiagnosticsPage />} />
              <Route path="/sos" element={<SosCallPage />} />
            </Routes>
          </main>
        </div>
      </div>

      <SosFloatingButton />
    </div>
  )
}

export default App
