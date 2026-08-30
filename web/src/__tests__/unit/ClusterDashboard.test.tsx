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
  it('应该显示页面标题', async () => {
    render(<ClusterDashboard />)
    // 标题在数据加载完成后随主界面渲染，需异步查找
    expect(await screen.findByText('Cluster Overview')).toBeInTheDocument()
  })

  it('应该显示加载状态', () => {
    const { container } = render(<ClusterDashboard />)
    // 初始为 loading，Loader2 spinner 无 role=status，用 class 断言
    expect(container.querySelector('.animate-spin')).toBeInTheDocument()
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
    
    // 验证节点和 Pod 统计（多集群卡片均有同名标题，用 getAllByText）
    expect(screen.getAllByText('Nodes').length).toBeGreaterThan(0)
    expect(screen.getAllByText('Pods').length).toBeGreaterThan(0)
    
    // 验证状态数字（mock 两集群均为 3/3 ready）
    expect(screen.getAllByText('3/3').length).toBeGreaterThan(0)
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
      expect(screen.getAllByText('View Details').length).toBeGreaterThan(0)
      expect(screen.getAllByText('View Metrics').length).toBeGreaterThan(0)
    })
  })
})
