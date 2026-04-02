import React, { useState, useEffect } from 'react'
import { Routes, Route, NavLink, useLocation } from 'react-router-dom'
import { cn } from './lib/utils'
import ClusterDashboard from './pages/ClusterDashboard'
import PodsPage from './pages/PodsPage'
import NodesPage from './pages/NodesPage'
import MonitoringPage from './pages/MonitoringPage'
import DeploymentsPage from './pages/DeploymentsPage'
import { ServicesPage } from './pages/ServicesPage'
import { Menu, X, Moon, Sun, Database, Server, Activity, AlertCircle, Boxes, Beaker, Globe } from 'lucide-react'

function App() {
  const [isDarkMode, setIsDarkMode] = useState(false)
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false)
  const [isMockMode, setIsMockMode] = useState(false)
  const location = useLocation()

  // 检查是否启用了 Mock 模式
  useEffect(() => {
    const mockEnabled = localStorage.getItem('USE_MOCK') === 'true'
    setIsMockMode(mockEnabled)
  }, [location]) // 路由变化时重新检查

  const toggleDarkMode = () => {
    setIsDarkMode(!isDarkMode)
    document.documentElement.classList.toggle('dark')
  }

  const toggleMockMode = () => {
    const newMockState = !isMockMode
    localStorage.setItem('USE_MOCK', newMockState ? 'true' : 'false')
    setIsMockMode(newMockState)
    // 刷新页面以应用更改
    window.location.reload()
  }

  const navItems = [
    { path: '/', label: 'Dashboard', icon: Database },
    { path: '/deployments', label: 'Deployments', icon: Boxes },
    { path: '/services', label: 'Services', icon: Globe },
    { path: '/pods', label: 'Pods', icon: Server },
    { path: '/nodes', label: 'Nodes', icon: Activity },
    { path: '/monitoring', label: 'Monitoring', icon: AlertCircle },
  ]

  return (
    <div className={cn('min-h-screen transition-colors duration-200', isDarkMode ? 'dark' : '')}>
      <div className="bg-white dark:bg-gray-900 border-b border-gray-200 dark:border-gray-800">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-between h-16">
            <div className="flex items-center space-x-4">
              <h1 className="text-xl font-bold text-primary-600 dark:text-primary-400">Klaw</h1>
              {isMockMode && (
                <span className="px-2 py-0.5 text-xs font-medium bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400 rounded-full">
                  MOCK
                </span>
              )}
            </div>
            
            <div className="hidden md:flex items-center space-x-4">
              {navItems.map((item) => (
                <NavLink
                  key={item.path}
                  to={item.path}
                  className={({ isActive }) => cn(
                    'flex items-center space-x-1 px-3 py-2 rounded-md text-sm font-medium transition-colors',
                    isActive
                      ? 'bg-primary-50 text-primary-700 dark:bg-primary-900/30 dark:text-primary-400'
                      : 'text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-gray-800'
                  )}
                >
                  <item.icon className="h-4 w-4" />
                  <span>{item.label}</span>
                </NavLink>
              ))}
              
              {/* Mock 模式切换按钮 */}
              <button
                onClick={toggleMockMode}
                title={isMockMode ? '关闭 Mock 模式' : '开启 Mock 模式'}
                className={cn(
                  'p-2 rounded-md text-sm font-medium transition-colors flex items-center space-x-1',
                  isMockMode
                    ? 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400'
                    : 'text-gray-500 hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-gray-800'
                )}
              >
                <Beaker className="h-4 w-4" />
                <span>Mock</span>
              </button>
              
              <button
                onClick={toggleDarkMode}
                className="p-2 rounded-md text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-gray-800"
              >
                {isDarkMode ? <Sun className="h-5 w-5" /> : <Moon className="h-5 w-5" />}
              </button>
            </div>
            
            <div className="md:hidden flex items-center">
              <button
                onClick={() => setIsMobileMenuOpen(!isMobileMenuOpen)}
                className="p-2 rounded-md text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-gray-800"
              >
                <Menu className="h-5 w-5" />
              </button>
            </div>
          </div>
        </div>
      </div>

      {isMobileMenuOpen && (
        <div className="md:hidden bg-white dark:bg-gray-900 border-b border-gray-200 dark:border-gray-800">
          <div className="px-2 pt-2 pb-3 space-y-1">
            <div className="flex items-center justify-between p-4">
              <div>
                {navItems.map((item) => (
                  <NavLink
                    key={item.path}
                    to={item.path}
                    className={({ isActive }) => cn(
                      'flex items-center space-x-2 px-3 py-2 rounded-md text-base font-medium mb-2',
                      isActive
                        ? 'bg-primary-50 text-primary-700 dark:bg-primary-900/30 dark:text-primary-400'
                        : 'text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-gray-800'
                    )}
                    onClick={() => setIsMobileMenuOpen(false)}
                  >
                    <item.icon className="h-5 w-5" />
                    <span>{item.label}</span>
                  </NavLink>
                ))}
              </div>
              <button
                onClick={() => setIsMobileMenuOpen(false)}
                className="p-2 rounded-md text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-gray-800"
              >
                <X className="h-6 w-6" />
              </button>
            </div>
            <div className="px-4 space-y-2">
              <button
                onClick={() => { toggleMockMode(); setIsMobileMenuOpen(false); }}
                className={cn(
                  'flex items-center space-x-2 w-full px-3 py-2 rounded-md text-base font-medium',
                  isMockMode
                    ? 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400'
                    : 'text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-gray-800'
                )}
              >
                <Beaker className="h-5 w-5" />
                <span>{isMockMode ? '关闭 Mock 模式' : '开启 Mock 模式'}</span>
              </button>
              <button
                onClick={() => { toggleDarkMode(); setIsMobileMenuOpen(false); }}
                className="flex items-center space-x-2 w-full px-3 py-2 rounded-md text-base font-medium text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-gray-800"
              >
                {isDarkMode ? <Sun className="h-5 w-5" /> : <Moon className="h-5 w-5" />}
                <span>{isDarkMode ? 'Light Mode' : 'Dark Mode'}</span>
              </button>
            </div>
          </div>
        </div>
      )}

      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <Routes>
          <Route path="/" element={<ClusterDashboard />} />
          <Route path="/deployments" element={<DeploymentsPage />} />
          <Route path="/services" element={<ServicesPage />} />
          <Route path="/pods" element={<PodsPage />} />
          <Route path="/nodes" element={<NodesPage />} />
          <Route path="/monitoring" element={<MonitoringPage />} />
        </Routes>
      </main>

      <footer className="bg-white dark:bg-gray-900 border-t border-gray-200 dark:border-gray-800 py-6">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex flex-col md:flex-row justify-between items-center">
            <div className="text-sm text-gray-600 dark:text-gray-400">
              © 2024 Klaw - Kubernetes Management Tool
            </div>
            <div className="mt-4 md:mt-0 text-sm text-gray-600 dark:text-gray-400">
              Built with React + TypeScript + Tailwind CSS
              {isMockMode && (
                <span className="ml-2 px-2 py-0.5 text-xs bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400 rounded">
                  Mock Mode
                </span>
              )}
            </div>
          </div>
        </div>
      </footer>
    </div>
  )
}

export default App
