// AuthGate 认证门单元测试

import { describe, it, expect, beforeAll, afterAll, afterEach } from 'vitest'
import { screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { render } from '../../test-utils/test-utils'
import { http, HttpResponse } from 'msw'
import { server } from '../mocks/server'
import { AuthProvider } from '../../contexts/AuthContext'
import AuthGate from '../../components/AuthGate'
import { clearAuthToken } from '../../lib/auth'

beforeAll(() => server.listen({ onUnhandledRequest: 'error' }))
afterAll(() => server.close())
afterEach(() => {
  server.resetHandlers()
  localStorage.clear()
  clearAuthToken()
})

function renderGate() {
  return render(
    <AuthProvider>
      <AuthGate>
        <div>受保护的应用内容</div>
      </AuthGate>
    </AuthProvider>,
  )
}

describe('AuthGate', () => {
  it('后端未开启认证时直接渲染应用内容', async () => {
    renderGate()
    expect(await screen.findByText('受保护的应用内容')).toBeInTheDocument()
    expect(screen.queryByText('Klaw 访问控制')).not.toBeInTheDocument()
  })

  it('认证开启且无 Token 时显示认证门', async () => {
    server.use(
      http.get('/api/v1/auth/verify', () =>
        HttpResponse.json({ error: 'Unauthorized: missing bearer token' }, { status: 401 }),
      ),
    )
    renderGate()
    expect(await screen.findByText('Klaw 访问控制')).toBeInTheDocument()
    expect(screen.queryByText('受保护的应用内容')).not.toBeInTheDocument()
  })

  it('输入有效 Token 后进入应用', async () => {
    const user = userEvent.setup()
    server.use(
      http.get('/api/v1/auth/verify', ({ request }) => {
        if (request.headers.get('Authorization') === 'Bearer valid-token') {
          return HttpResponse.json({ authenticated: true, authEnabled: true })
        }
        return HttpResponse.json({ error: 'Unauthorized: missing bearer token' }, { status: 401 })
      }),
    )
    renderGate()
    await screen.findByText('Klaw 访问控制')

    await user.type(screen.getByLabelText('Access Token'), 'valid-token')
    await user.click(screen.getByRole('button', { name: '验证并进入' }))

    expect(await screen.findByText('受保护的应用内容')).toBeInTheDocument()
  })

  it('输入无效 Token 时提示错误并保留认证门', async () => {
    const user = userEvent.setup()
    server.use(
      http.get('/api/v1/auth/verify', () =>
        HttpResponse.json({ error: 'Unauthorized: invalid token' }, { status: 401 }),
      ),
    )
    renderGate()
    await screen.findByText('Klaw 访问控制')

    await user.type(screen.getByLabelText('Access Token'), 'wrong-token')
    await user.click(screen.getByRole('button', { name: '验证并进入' }))

    expect(await screen.findByText('Token 无效或已过期，请重试')).toBeInTheDocument()
    await waitFor(() => {
      expect(screen.getByText('Klaw 访问控制')).toBeInTheDocument()
    })
    // 验证失败的 Token 应被清除，不残留
    expect(localStorage.getItem('klaw_token')).toBeNull()
  })
})
