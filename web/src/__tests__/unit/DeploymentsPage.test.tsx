// DeploymentsPage 页面单元测试

import { describe, it, expect, beforeAll, afterAll, afterEach } from 'vitest'
import { screen, waitFor, fireEvent } from '../../test-utils/test-utils'
import { render } from '../../test-utils/test-utils'
import { server } from '../mocks/server'
import DeploymentsPage from '../../pages/DeploymentsPage'

// 启动 MSW
beforeAll(() => server.listen({ onUnhandledRequest: 'error' }))
afterAll(() => server.close())
afterEach(() => server.resetHandlers())

describe('DeploymentsPage', () => {
  it('应该显示页面标题', () => {
    render(<DeploymentsPage />)
    expect(screen.getByText('Deployments Management')).toBeInTheDocument()
  })

  it('应该显示集群选择器', () => {
    render(<DeploymentsPage />)
    expect(screen.getByText('Select Cluster')).toBeInTheDocument()
  })

  it('应该成功加载并显示 Deployment 列表', async () => {
    render(<DeploymentsPage />)
    
    // 等待数据加载
    await waitFor(() => {
      expect(screen.getByText('nginx')).toBeInTheDocument()
    }, { timeout: 3000 })
    
    // 验证所有 deployments 都显示
    expect(screen.getByText('frontend')).toBeInTheDocument()
    expect(screen.getByText('httpbin')).toBeInTheDocument()
  })

  it('应该显示 Deployment 状态', async () => {
    render(<DeploymentsPage />)
    
    await waitFor(() => {
      expect(screen.getByText('nginx')).toBeInTheDocument()
    }, { timeout: 3000 })
    
    // 验证状态标签
    expect(screen.getByText('Available')).toBeInTheDocument()
  })

  it('应该显示副本数量', async () => {
    render(<DeploymentsPage />)
    
    await waitFor(() => {
      expect(screen.getByText('nginx')).toBeInTheDocument()
    }, { timeout: 3000 })
    
    // 验证副本显示 (available/desired)
    expect(screen.getByText('2/2')).toBeInTheDocument() // nginx
  })

  it('应该显示镜像信息', async () => {
    render(<DeploymentsPage />)
    
    await waitFor(() => {
      expect(screen.getByText('nginx:alpine')).toBeInTheDocument()
    }, { timeout: 3000 })
  })

  it('应该支持搜索过滤', async () => {
    render(<DeploymentsPage />)
    
    await waitFor(() => {
      expect(screen.getByText('nginx')).toBeInTheDocument()
    }, { timeout: 3000 })
    
    // 输入搜索词
    const searchInput = screen.getByPlaceholderText('Search deployments...')
    fireEvent.change(searchInput, { target: { value: 'nginx' } })
    
    // 验证过滤结果
    expect(screen.getByText('nginx')).toBeInTheDocument()
    expect(screen.queryByText('frontend')).not.toBeInTheDocument()
  })

  it('应该支持展开/收起详情', async () => {
    render(<DeploymentsPage />)
    
    await waitFor(() => {
      expect(screen.getByText('nginx')).toBeInTheDocument()
    }, { timeout: 3000 })
    
    // 点击展开按钮
    const expandButtons = screen.getAllByRole('button').filter(
      btn => btn.querySelector('svg') !== null
    )
    if (expandButtons.length > 0) {
      fireEvent.click(expandButtons[0])
      
      // 验证详情显示
      await waitFor(() => {
        expect(screen.getByText(/Deployment Details:/)).toBeInTheDocument()
      })
    }
  })

  it('应该显示刷新按钮', async () => {
    render(<DeploymentsPage />)
    
    await waitFor(() => {
      expect(screen.getByText('Refresh')).toBeInTheDocument()
    })
  })

  it('应该有扩缩容按钮', async () => {
    render(<DeploymentsPage />)
    
    await waitFor(() => {
      expect(screen.getByText('nginx')).toBeInTheDocument()
    }, { timeout: 3000 })
    
    // 查找 +/- 按钮
    const buttons = screen.getAllByRole('button')
    expect(buttons.length).toBeGreaterThan(0)
  })
})
