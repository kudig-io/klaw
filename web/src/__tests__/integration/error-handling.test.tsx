// 错误处理集成测试

import { describe, it, expect, beforeAll, afterAll, afterEach } from 'vitest'
import { screen, waitFor } from '../../test-utils/test-utils'
import { render } from '../../test-utils/test-utils'
import { server, errorHandlers } from '../mocks/server'
import ClusterDashboard from '../../pages/ClusterDashboard'

// 启动 MSW
beforeAll(() => server.listen({ onUnhandledRequest: 'error' }))
afterAll(() => server.close())
afterEach(() => {
  server.resetHandlers()
  // 恢复默认 handlers
  server.use(...errorHandlers)
})

describe('Error Handling', () => {
  it('应该显示错误信息当 API 失败时', async () => {
    // 使用错误 handlers
    server.use(...errorHandlers)
    
    render(<ClusterDashboard />)
    
    await waitFor(() => {
      expect(screen.getByText(/Failed to fetch cluster data/)).toBeInTheDocument()
    }, { timeout: 3000 })
  })

  it('应该显示重试按钮当请求失败时', async () => {
    server.use(...errorHandlers)
    
    render(<ClusterDashboard />)
    
    await waitFor(() => {
      expect(screen.getByText('Try Again')).toBeInTheDocument()
    }, { timeout: 3000 })
  })
})
