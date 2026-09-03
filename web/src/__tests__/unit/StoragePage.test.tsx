// StoragePage 页面单元测试

import { describe, it, expect, beforeAll, afterAll, afterEach } from 'vitest'
import { screen, waitFor, render } from '../../test-utils/test-utils'
import { server } from '../mocks/server'
import { ToastProvider } from '../../contexts/ToastContext'
import { StoragePage } from '../../pages/StoragePage'

// 启动 MSW
beforeAll(() => server.listen({ onUnhandledRequest: 'error' }))
afterAll(() => server.close())
afterEach(() => server.resetHandlers())

// 统计卡取值：label 可能与表头同名，只认后跟数值元素的统计卡标签
function statValue(label: string): string | null {
  for (const el of screen.getAllByText(label)) {
    const value = el.parentElement?.querySelector('.text-2xl, .text-lg')?.textContent
    if (value != null) return value
  }
  return null
}

// 资源名可能在表格与关联列/分析卡中重复出现
async function waitForText(text: string) {
  await waitFor(() => {
    expect(screen.getAllByText(text).length).toBeGreaterThan(0)
  }, { timeout: 3000 })
}

function renderPage() {
  return render(
    <ToastProvider>
      <StoragePage />
    </ToastProvider>
  )
}

describe('StoragePage', () => {
  it('应该显示页面标题与刷新按钮', () => {
    renderPage()
    expect(screen.getByText('存储管理')).toBeInTheDocument()
    expect(screen.getByText('刷新')).toBeInTheDocument()
  })

  it('应该加载并显示全部 PVC 列表', async () => {
    renderPage()

    await waitForText('mall-frontend-data')

    for (const name of ['mall-redis-data', 'order-service-db', 'payment-service-logs', 'staging-cache', 'nginx-html', 'legacy-archive']) {
      expect(screen.getAllByText(name).length).toBeGreaterThan(0)
    }
  })

  it('应该标记 Terminating 状态的 PVC 为删除中', async () => {
    renderPage()

    await waitForText('legacy-archive')

    expect(screen.getByText('删除中')).toBeInTheDocument()
    expect(screen.getAllByText('Terminating').length).toBeGreaterThan(0)
    expect(screen.getAllByText('Pending').length).toBeGreaterThan(0)
  })

  it('应该加载并显示全部 PV 列表（含状态与原因）', async () => {
    renderPage()

    await waitForText('pvc-9f2c8a01-mall-frontend')

    for (const name of ['nfs-shared-media', 'local-pv-frontend-node1', 'pvc-old-released', 'pvc-failed-detach']) {
      expect(screen.getAllByText(name).length).toBeGreaterThan(0)
    }

    // PV 异常原因
    expect(screen.getByText('VolumeReleased')).toBeInTheDocument()
    expect(screen.getByText('VolumeFailedDelete')).toBeInTheDocument()
  })

  it('应该加载并显示全部 StorageClass（默认标记与可扩容）', async () => {
    renderPage()

    await waitForText('standard')

    for (const name of ['fast-ssd', 'cephfs-shared', 'nfs-retained', 'archive-cold']) {
      expect(screen.getAllByText(name).length).toBeGreaterThan(0)
    }

    // standard 为默认存储类
    expect(screen.getByText('默认')).toBeInTheDocument()
    // fast-ssd / cephfs-shared / archive-cold 支持扩容
    expect(screen.getAllByText('支持').length).toBe(3)
  })

  it('应该显示 PV 来源（CSI / NFS / hostPath）', async () => {
    renderPage()

    await waitForText('pvc-9f2c8a01-mall-frontend')

    expect(screen.getAllByText('rbd.csi.ceph.com').length).toBeGreaterThan(0)
    expect(screen.getByText('NFS · 192.168.1.100')).toBeInTheDocument()
    expect(screen.getByText('hostPath · /var/local-path-provisioner/frontend-node1')).toBeInTheDocument()
  })

  it('应该显示统计卡数量与容量使用（7 PVC / 7 PV / 5 SC，83.0 Gi / 197 Gi）', async () => {
    renderPage()

    await waitForText('mall-frontend-data')

    expect(statValue('PVC 声明')).toBe('7')
    expect(statValue('持久卷 PV')).toBe('7')
    expect(statValue('存储类')).toBe('5')
    expect(statValue('容量使用')).toBe('83.0 Gi / 197 Gi')
  })

  it('应该显示存储分析区（状态分布与容量条）', async () => {
    renderPage()

    await waitForText('存储分析')

    expect(screen.getByText('PV 状态分布')).toBeInTheDocument()
    expect(screen.getByText('PVC 状态分布')).toBeInTheDocument()
    expect(screen.getByText('PV 按存储类')).toBeInTheDocument()
    expect(screen.getByText('存储类按 Provisioner')).toBeInTheDocument()

    // 容量汇总卡
    expect(screen.getByText('总容量')).toBeInTheDocument()
    expect(screen.getByText('已申请')).toBeInTheDocument()
    expect(screen.getByText('可用')).toBeInTheDocument()
    expect(screen.getByText('197 Gi')).toBeInTheDocument()
    expect(screen.getByText('114 Gi')).toBeInTheDocument()
  })
})
