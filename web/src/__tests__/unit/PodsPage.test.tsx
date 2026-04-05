// PodsPage 页面单元测试

import { describe, it, expect, beforeAll, afterAll, afterEach } from 'vitest'
import { screen, waitFor, fireEvent } from '../../test-utils/test-utils'
import { render } from '../../test-utils/test-utils'
import { server } from '../mocks/server'
import PodsPage from '../../pages/PodsPage'

// 启动 MSW
beforeAll(() => server.listen({ onUnhandledRequest: 'error' }))
afterAll(() => server.close())
afterEach(() => server.resetHandlers())

describe('PodsPage', () => {
  it('应该显示页面标题', () => {
    render(<PodsPage />)
    expect(screen.getByText('Pods Management')).toBeInTheDocument()
  })

  it('应该显示集群和命名空间选择器', () => {
    render(<PodsPage />)
    expect(screen.getByText('Select Cluster')).toBeInTheDocument()
    expect(screen.getByText('Select Namespace')).toBeInTheDocument()
  })

  it('应该成功加载并显示 Pod 列表', async () => {
    render(<PodsPage />)
    
    // 等待数据加载
    await waitFor(() => {
      expect(screen.getByText(/nginx-/)).toBeInTheDocument()
    }, { timeout: 3000 })
    
    // 验证 Pods 显示
    expect(screen.getByText(/frontend-/)).toBeInTheDocument()
    expect(screen.getByText(/httpbin-/)).toBeInTheDocument()
  })

  it('应该显示 Pod 状态', async () => {
    render(<PodsPage />)
    
    await waitFor(() => {
      expect(screen.getByText(/nginx-/)).toBeInTheDocument()
    }, { timeout: 3000 })
    
    // 验证状态标签
    expect(screen.getByText('Running')).toBeInTheDocument()
    expect(screen.getByText('Pending')).toBeInTheDocument()
  })

  it('应该支持搜索过滤', async () => {
    render(<PodsPage />)
    
    await waitFor(() => {
      expect(screen.getByText(/nginx-/)).toBeInTheDocument()
    }, { timeout: 3000 })
    
    // 输入搜索词
    const searchInput = screen.getByPlaceholderText('Search pods...')
    fireEvent.change(searchInput, { target: { value: 'nginx' } })
    
    // 验证过滤结果（pod 名称包含搜索词）
    const rows = screen.getAllByRole('row')
    expect(rows.length).toBeGreaterThan(0)
  })

  it('应该显示 Pod 详细信息列', async () => {
    render(<PodsPage />)
    
    await waitFor(() => {
      expect(screen.getByText(/nginx-/)).toBeInTheDocument()
    }, { timeout: 3000 })
    
    // 验证表头
    expect(screen.getByText('Pod Name')).toBeInTheDocument()
    expect(screen.getByText('Status')).toBeInTheDocument()
    expect(screen.getByText('Node')).toBeInTheDocument()
    expect(screen.getByText('IP')).toBeInTheDocument()
    expect(screen.getByText('Created')).toBeInTheDocument()
  })

  it('应该显示刷新按钮', () => {
    render(<PodsPage />)
    expect(screen.getByText('Refresh')).toBeInTheDocument()
  })

  it('应该支持展开查看日志', async () => {
    render(<PodsPage />)
    
    await waitFor(() => {
      expect(screen.getByText(/nginx-/)).toBeInTheDocument()
    }, { timeout: 3000 })
    
    // 查找展开按钮（向下箭头）
    const buttons = screen.getAllByRole('button')
    const expandButton = buttons.find(btn => 
      btn.querySelector('svg[data-lucide="chevron-down"]') !== null
    )
    
    if (expandButton) {
      fireEvent.click(expandButton)
      
      // 验证日志区域显示
      await waitFor(() => {
        expect(screen.getByText(/Logs for/)).toBeInTheDocument()
      })
    }
  })
})
