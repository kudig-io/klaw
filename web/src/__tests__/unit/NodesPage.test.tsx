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
    expect(screen.getByText('Nodes Management')).toBeInTheDocument()
  })

  it('应该显示集群选择器', () => {
    render(<NodesPage />)
    expect(screen.getByText('Select Cluster')).toBeInTheDocument()
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
    
    // 验证 Ready 状态
    expect(screen.getByText('Ready')).toBeInTheDocument()
  })

  it('应该显示节点资源信息', async () => {
    render(<NodesPage />)
    
    await waitFor(() => {
      expect(screen.getByText('kind-test-control-plane')).toBeInTheDocument()
    }, { timeout: 3000 })
    
    // 验证表头
    expect(screen.getByText('Node Name')).toBeInTheDocument()
    expect(screen.getByText('Status')).toBeInTheDocument()
    expect(screen.getByText('CPU')).toBeInTheDocument()
    expect(screen.getByText('Memory')).toBeInTheDocument()
    expect(screen.getByText('Created')).toBeInTheDocument()
    
    // 验证资源值
    expect(screen.getAllByText('4').length).toBeGreaterThan(0) // CPU
    expect(screen.getAllByText('8Gi').length).toBeGreaterThan(0) // Memory
  })

  it('应该显示刷新按钮', () => {
    render(<NodesPage />)
    expect(screen.getByText('Refresh')).toBeInTheDocument()
  })

  it('应该显示节点数量统计', async () => {
    render(<NodesPage />)
    
    await waitFor(() => {
      expect(screen.getByText('kind-test-control-plane')).toBeInTheDocument()
    }, { timeout: 3000 })
    
    // 验证统计信息
    expect(screen.getByText(/3 nodes/)).toBeInTheDocument()
  })
})
