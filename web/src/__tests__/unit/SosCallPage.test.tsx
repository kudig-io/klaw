import { render, screen, fireEvent } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { MemoryRouter } from 'react-router-dom'
import App from '../../App'
import SosCallPage from '../../pages/SosCallPage'

const fetchSosStatusMock = vi.fn()
const useSosSessionMock = vi.fn()

vi.mock('../../lib/sosApi', () => ({
  fetchSosStatus: (...a: unknown[]) => fetchSosStatusMock(...a),
}))
vi.mock('../../hooks/useSosSession', () => ({
  useSosSession: () => useSosSessionMock(),
}))

function renderPage() {
  return render(
    <MemoryRouter>
      <SosCallPage />
    </MemoryRouter>,
  )
}

describe('SosCallPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    fetchSosStatusMock.mockResolvedValue({
      enabled: true,
      ready: true,
      model: 'm',
      voice: 'v',
      faq_count: 1,
    })
  })

  it('未就绪时展示配置引导', async () => {
    fetchSosStatusMock.mockResolvedValue({
      enabled: true,
      ready: false,
      model: '',
      voice: '',
      faq_count: 0,
    })
    useSosSessionMock.mockReturnValue({
      state: {
        status: 'idle',
        error: '',
        model: '',
        voice: '',
        userText: '',
        assistantText: '',
        muted: false,
        speaking: false,
        toolCall: '',
        messages: [],
      },
      start: vi.fn(),
      hangup: vi.fn(),
      toggleMute: vi.fn(),
    })
    renderPage()
    expect(await screen.findByText(/KLAW_SOS_DASHSCOPE_API_KEY/)).toBeInTheDocument()
    expect(screen.getByRole('link', { name: '返回首页' })).toBeInTheDocument()
  })

  it('就绪时展示通话界面与控制按钮', async () => {
    useSosSessionMock.mockReturnValue({
      state: {
        status: 'idle',
        muted: false,
        messages: [],
        userText: '',
        assistantText: '',
        error: '',
      },
      start: vi.fn(),
      hangup: vi.fn(),
      toggleMute: vi.fn(),
    })
    renderPage()
    expect(await screen.findByRole('button', { name: /静音/ })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /挂断/ })).toBeInTheDocument()
  })

  it('会话错误时展示错误信息、重试与返回首页', async () => {
    const start = vi.fn()
    useSosSessionMock.mockReturnValue({
      state: {
        status: 'error',
        error: 'WebSocket 连接失败',
        muted: false,
        messages: [],
        userText: '',
        assistantText: '',
      },
      start,
      hangup: vi.fn(),
      toggleMute: vi.fn(),
    })
    renderPage()
    expect(await screen.findByText('WebSocket 连接失败')).toBeInTheDocument()
    fireEvent.click(screen.getByRole('button', { name: '重试' }))
    expect(start).toHaveBeenCalledTimes(1)
    expect(screen.getByRole('link', { name: '返回首页' })).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: /挂断/ })).not.toBeInTheDocument()
  })
})

describe('SOS 通话页脱离 Workbench 骨架（App 级路由）', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    fetchSosStatusMock.mockResolvedValue({
      enabled: true,
      ready: true,
      model: 'm',
      voice: 'v',
      faq_count: 1,
    })
  })

  it('/sos 路由不渲染侧栏、顶栏按钮与悬浮球', async () => {
    useSosSessionMock.mockReturnValue({
      state: { status: 'connecting', error: '', muted: false, messages: [], userText: '', assistantText: '' },
      start: vi.fn(),
      hangup: vi.fn(),
      toggleMute: vi.fn(),
    })
    render(
      <MemoryRouter initialEntries={['/sos']}>
        <App />
      </MemoryRouter>,
    )
    expect(await screen.findByText('SOS 应急语音助手')).toBeInTheDocument()
    expect(screen.queryByText('仪表盘')).not.toBeInTheDocument()
    expect(screen.queryByTitle('开启 Mock 模式')).not.toBeInTheDocument()
    expect(screen.queryByRole('link', { name: /sos/i })).not.toBeInTheDocument()
  })
})
