import { http, HttpResponse } from 'msw'

// 默认后端未开启认证；需要模拟「认证开启且 token 无效」的用例用 server.use 覆盖此 handler
export const authHandlers = [
  http.get('/api/v1/auth/verify', () =>
    HttpResponse.json({ authenticated: true, authEnabled: false }),
  ),
]
