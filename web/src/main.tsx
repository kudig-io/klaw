import React from 'react'
import ReactDOM from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'
import App from './App.tsx'
import { ToastProvider } from './contexts/ToastContext.tsx'
import './index.css'

async function bootstrap() {
  if (import.meta.env.VITE_USE_MOCK === 'true') {
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
