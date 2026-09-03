import React from 'react'
import ReactDOM from 'react-dom/client'
import { BrowserRouter, HashRouter } from 'react-router-dom'
import App from './App.tsx'
import { ToastProvider } from './contexts/ToastContext.tsx'
import './index.css'

async function bootstrap() {
  // 环境变量 OR localStorage 开关均可触发 Mock（让顶栏 Mock 按钮在常规 dev 模式也能用）
  const mockEnabled = import.meta.env.VITE_USE_MOCK === 'true' || localStorage.getItem('USE_MOCK') === 'true'
  if (mockEnabled) {
    const { startMockService } = await import('./__tests__/mocks/browser')
    await startMockService()
  }

  // 支持 ?theme=dark 深链（演示/截图场景），与顶栏切换按钮共用同一 class
  if (new URLSearchParams(location.search).get('theme') === 'dark') {
    document.documentElement.classList.add('dark')
  }

  // Meoo 等静态托管没有服务端路由，深链刷新会 404，构建时用 VITE_USE_HASH_ROUTER=true 切换
  const Router = import.meta.env.VITE_USE_HASH_ROUTER === 'true' ? HashRouter : BrowserRouter

  ReactDOM.createRoot(document.getElementById('root')!).render(
    <React.StrictMode>
      <Router>
        <ToastProvider>
          <App />
        </ToastProvider>
      </Router>
    </React.StrictMode>,
  )
}

bootstrap()
