import React from 'react'
import ReactDOM from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'
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

  ReactDOM.createRoot(document.getElementById('root')!).render(
    <React.StrictMode>
      <BrowserRouter>
        <ToastProvider>
          <App />
        </ToastProvider>
      </BrowserRouter>
    </React.StrictMode>,
  )
}

bootstrap()
