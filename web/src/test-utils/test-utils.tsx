// 测试工具函数

import React from 'react'
import { render as rtlRender, RenderOptions } from '@testing-library/react'
import { BrowserRouter } from 'react-router-dom'

// 包装组件，提供必要的 context
function AllTheProviders({ children }: { children: React.ReactNode }) {
  return (
    <BrowserRouter>
      {children}
    </BrowserRouter>
  )
}

// 自定义 render 函数
function render(ui: React.ReactElement, options?: Omit<RenderOptions, 'wrapper'>) {
  return rtlRender(ui, { wrapper: AllTheProviders, ...options })
}

// 重新导出 @testing-library/react
export * from '@testing-library/react'
export { render }
