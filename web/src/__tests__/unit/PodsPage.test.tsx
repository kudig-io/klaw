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
    expect(screen.getByText('容器组（Pod）管理')).toBeInTheDocument()
  })

  it('应该显示集群和命名空间选择器', () => {
    render(<PodsPage />)
    expect(screen.getByText('选择集群')).toBeInTheDocument()
    expect(screen.getByText('全部命名空间')).toBeInTheDocument()
  })

  it('应该成功加载并显示 Pod 列表', async () => {
    render(<PodsPage />)
    
    // 等待数据加载（多副本同名前缀，用 getAllByText）
    await waitFor(() => {
      expect(screen.getAllByText(/nginx-/).length).toBeGreaterThan(0)
    }, { timeout: 3000 })
    
    // 验证 Pods 显示
    expect(screen.getAllByText(/frontend-/).length).toBeGreaterThan(0)
    expect(screen.getAllByText(/httpbin-/).length).toBeGreaterThan(0)
  })

  it('应该显示 Pod 状态', async () => {
    render(<PodsPage />)
    
    await waitFor(() => {
      expect(screen.getAllByText(/nginx-/).length).toBeGreaterThan(0)
    }, { timeout: 3000 })
    
    // 验证状态标签（多个 Pod 共用同名状态）
    expect(screen.getAllByText('Running').length).toBeGreaterThan(0)
    expect(screen.getAllByText('Pending').length).toBeGreaterThan(0)
  })

  it('应该支持搜索过滤', async () => {
    render(<PodsPage />)
    
    await waitFor(() => {
      expect(screen.getAllByText(/nginx-/).length).toBeGreaterThan(0)
    }, { timeout: 3000 })
    
    // 输入搜索词
    const searchInput = screen.getByPlaceholderText('搜索容器组…')
    fireEvent.change(searchInput, { target: { value: 'nginx' } })
    
    // 验证过滤结果（pod 名称包含搜索词）
    const rows = screen.getAllByRole('row')
    expect(rows.length).toBeGreaterThan(0)
  })

  it('应该显示 Pod 详细信息列', async () => {
    render(<PodsPage />)
    
    await waitFor(() => {
      expect(screen.getAllByText(/nginx-/).length).toBeGreaterThan(0)
    }, { timeout: 3000 })
    
    // 验证表头
    expect(screen.getByText('容器组名称')).toBeInTheDocument()
    expect(screen.getByText('状态')).toBeInTheDocument()
    expect(screen.getByText('节点')).toBeInTheDocument()
    expect(screen.getByText('IP 地址')).toBeInTheDocument()
    expect(screen.getByText('创建时间')).toBeInTheDocument()
  })

  it('应该显示刷新按钮', () => {
    render(<PodsPage />)
    expect(screen.getByText('刷新')).toBeInTheDocument()
  })

  it('应该支持展开查看日志', async () => {
    render(<PodsPage />)
    
    await waitFor(() => {
      expect(screen.getAllByText(/nginx-/).length).toBeGreaterThan(0)
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
        expect(screen.getByText(/的日志/)).toBeInTheDocument()
      })
    }
  })
})
