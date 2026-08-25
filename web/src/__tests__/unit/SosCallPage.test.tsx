import { render, screen } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { MemoryRouter } from 'react-router-dom'
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
})
