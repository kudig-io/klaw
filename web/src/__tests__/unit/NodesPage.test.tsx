// NodesPage 页面单元测试

import { describe, it, expect, beforeAll, afterAll, afterEach } from 'vitest'
import { screen, waitFor } from '../../test-utils/test-utils'
import { render } from '../../test-utils/test-utils'
import { server } from '../mocks/server'
import NodesPage from '../../pages/NodesPage'

// 启动 MSW
beforeAll(() => server.listen({ onUnhandledRequest: 'error' }))
afterAll(() => server.close())
afterEach(() => server.resetHandlers())

describe('NodesPage', () => {
  it('应该显示页面标题', () => {
    render(<NodesPage />)
    expect(screen.getByText('节点（Node）管理')).toBeInTheDocument()
  })

  it('应该显示集群选择器', () => {
    render(<NodesPage />)
    expect(screen.getByText('选择集群')).toBeInTheDocument()
  })

  it('应该成功加载并显示节点列表', async () => {
    render(<NodesPage />)
    
    // 等待数据加载
    await waitFor(() => {
      expect(screen.getByText('kind-test-control-plane')).toBeInTheDocument()
    }, { timeout: 3000 })
    
    // 验证所有节点都显示
    expect(screen.getByText('kind-test-worker')).toBeInTheDocument()
    expect(screen.getByText('kind-test-worker2')).toBeInTheDocument()
  })

  it('应该显示节点状态', async () => {
    render(<NodesPage />)
    
    await waitFor(() => {
      expect(screen.getByText('kind-test-control-plane')).toBeInTheDocument()
    }, { timeout: 3000 })
    
    // 验证 Ready 状态（每个节点卡片与 Conditions 均有，用 getAllByText）
    expect(screen.getAllByText('Ready').length).toBeGreaterThan(0)
  })

  it('应该显示节点资源信息', async () => {
    render(<NodesPage />)
    
    await waitFor(() => {
      expect(screen.getByText('kind-test-control-plane')).toBeInTheDocument()
    }, { timeout: 3000 })
    
    // 节点为卡片布局（非表格），断言卡片内的资源标题与容量值
    expect(screen.getAllByText('CPU').length).toBeGreaterThan(0)
    expect(screen.getAllByText('内存').length).toBeGreaterThan(0)
    expect(screen.getAllByText('状态条件（Conditions）').length).toBeGreaterThan(0)
  })

  it('应该显示刷新按钮', () => {
    render(<NodesPage />)
    expect(screen.getByText('刷新')).toBeInTheDocument()
  })

  it('应该显示节点数量统计', async () => {
    render(<NodesPage />)
    
    await waitFor(() => {
      expect(screen.getByText('kind-test-control-plane')).toBeInTheDocument()
    }, { timeout: 3000 })
    
    // 验证 mock 的 3 个节点全部渲染（无分页/截断）
    expect(screen.getByText('kind-test-worker')).toBeInTheDocument()
    expect(screen.getByText('kind-test-worker2')).toBeInTheDocument()
  })
})
