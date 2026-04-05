// ClusterDashboard 页面单元测试

import { describe, it, expect, beforeAll, afterAll, afterEach } from 'vitest'
import { screen, waitFor } from '../../test-utils/test-utils'
import { render } from '../../test-utils/test-utils'
import { server } from '../mocks/server'
import ClusterDashboard from '../../pages/ClusterDashboard'

// 启动 MSW
beforeAll(() => server.listen({ onUnhandledRequest: 'error' }))
afterAll(() => server.close())
afterEach(() => server.resetHandlers())

describe('ClusterDashboard', () => {
  it('应该显示页面标题', () => {
    render(<ClusterDashboard />)
    expect(screen.getByText('Cluster Overview')).toBeInTheDocument()
  })

  it('应该显示加载状态', () => {
    render(<ClusterDashboard />)
    expect(screen.getByRole('status')).toBeInTheDocument()
  })

  it('应该成功加载并显示集群列表', async () => {
    render(<ClusterDashboard />)
    
    // 等待加载完成
    await waitFor(() => {
      expect(screen.getByText('kind-test')).toBeInTheDocument()
    }, { timeout: 3000 })
    
    // 验证集群信息显示
    expect(screen.getByText('production')).toBeInTheDocument()
  })

  it('应该显示集群状态信息', async () => {
    render(<ClusterDashboard />)
    
    await waitFor(() => {
      expect(screen.getByText('kind-test')).toBeInTheDocument()
    }, { timeout: 3000 })
    
    // 验证节点和 Pod 统计
    expect(screen.getByText('Nodes')).toBeInTheDocument()
    expect(screen.getByText('Pods')).toBeInTheDocument()
    
    // 验证状态数字
    expect(screen.getByText('3/3')).toBeInTheDocument() // 3/3 nodes ready
    expect(screen.getByText('10')).toBeInTheDocument()  // 10 running pods
  })

  it('应该显示刷新按钮', async () => {
    render(<ClusterDashboard />)
    
    await waitFor(() => {
      expect(screen.getByText('Refresh')).toBeInTheDocument()
    })
  })

  it('应该显示操作按钮', async () => {
    render(<ClusterDashboard />)
    
    await waitFor(() => {
      expect(screen.getByText('View Details')).toBeInTheDocument()
      expect(screen.getByText('View Metrics')).toBeInTheDocument()
    })
  })
})
