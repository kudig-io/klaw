// NetworkPage 页面单元测试

import { describe, it, expect, beforeAll, afterAll, afterEach } from 'vitest'
import { screen, waitFor, render } from '../../test-utils/test-utils'
import { server } from '../mocks/server'
import { ToastProvider } from '../../contexts/ToastContext'
import { NetworkPage } from '../../pages/NetworkPage'

// 启动 MSW
beforeAll(() => server.listen({ onUnhandledRequest: 'error' }))
afterAll(() => server.close())
afterEach(() => server.resetHandlers())

// 统计卡取值：label 可能与表头/标题同名，只认后跟数值元素的统计卡标签
function statValue(label: string): string | null {
  for (const el of screen.getAllByText(label)) {
    const value = el.parentElement?.querySelector('.text-2xl, .text-lg')?.textContent
    if (value != null) return value
  }
  return null
}

// 资源名可能在表格与分析卡中重复出现
async function waitForText(text: string) {
  await waitFor(() => {
    expect(screen.getAllByText(text).length).toBeGreaterThan(0)
  }, { timeout: 3000 })
}

function renderPage() {
  return render(
    <ToastProvider>
      <NetworkPage />
    </ToastProvider>
  )
}

describe('NetworkPage', () => {
  it('应该显示页面标题与刷新按钮', () => {
    renderPage()
    expect(screen.getByText('网络管理')).toBeInTheDocument()
    expect(screen.getByText('刷新')).toBeInTheDocument()
  })

  it('应该加载并显示全部 Ingress 列表', async () => {
    renderPage()

    await waitForText('frontend-ingress')

    for (const name of ['httpbin-ingress', 'mall-ingress', 'mall-staging-ingress', 'ingress-nginx-dashboard', 'legacy-web-ingress']) {
      expect(screen.getAllByText(name).length).toBeGreaterThan(0)
    }
  })

  it('应该显示 Ingress 关键字段（Host / TLS / 未分配地址）', async () => {
    renderPage()

    await waitForText('frontend-ingress')

    // 多 host 显示首 host + 折叠数
    expect(screen.getByText('mall.example.com +1')).toBeInTheDocument()
    // TLS secretName 徽章
    expect(screen.getByText('mall-tls-cert')).toBeInTheDocument()
    expect(screen.getByText('staging-wildcard-cert')).toBeInTheDocument()
    // 无 LB 地址的遗留 Ingress 显示"等待分配"
    expect(screen.getByText('等待分配')).toBeInTheDocument()
  })

  it('应该加载并显示全部 NetworkPolicy 列表', async () => {
    renderPage()

    await waitForText('default-deny-all')

    for (const name of ['allow-frontend-to-gateway', 'allow-dns-egress', 'payment-egress-restricted', 'order-ingress-policy', 'default-deny-ingress', 'allow-httpbin-ingress']) {
      expect(screen.getAllByText(name).length).toBeGreaterThan(0)
    }
  })

  it('应该正确表达空规则的默认拒绝语义与选择器摘要', async () => {
    renderPage()

    await waitForText('default-deny-all')

    // 双向空规则 = 入站/出站全部拒绝
    expect(screen.getByText('入站全部拒绝 · 出站全部拒绝')).toBeInTheDocument()
    // 空 podSelector = 所有 Pod
    expect(screen.getAllByText('所有 Pod').length).toBeGreaterThan(0)
    // matchExpressions 选择器
    expect(screen.getByText('表达式选择')).toBeInTheDocument()
    // 规则条数摘要
    expect(screen.getByText('入站 2 条')).toBeInTheDocument()
    expect(screen.getByText('出站 2 条')).toBeInTheDocument()
  })

  it('应该显示统计卡数量（6 Ingress / 8 策略 / 2 命名空间 / 3 暴露服务）', async () => {
    renderPage()

    await waitForText('frontend-ingress')

    expect(statValue('Ingress 规则')).toBe('6')
    expect(statValue('网络策略')).toBe('8')
    expect(statValue('策略覆盖命名空间')).toBe('2')
    expect(statValue('对外暴露服务')).toBe('3')
  })

  it('应该显示网络分析区（服务类型分布与暴露服务表）', async () => {
    renderPage()

    await waitForText('网络分析')

    expect(screen.getByText('服务类型分布')).toBeInTheDocument()
    expect(screen.getByText('策略分布（按命名空间）')).toBeInTheDocument()
    expect(screen.getByText('Ingress Host 分布')).toBeInTheDocument()
    expect(screen.getAllByText('LoadBalancer').length).toBeGreaterThan(0)

    // 暴露服务表：NodePort 端口徽章 80:30080/TCP
    expect(screen.getByText('80:30080/TCP')).toBeInTheDocument()
    expect(screen.getAllByText('mall-frontend').length).toBeGreaterThan(0)
  })
})
