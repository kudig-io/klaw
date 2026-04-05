// MSW browser setup for development

import { setupWorker } from 'msw/browser'
import { handlers } from './handlers'

// 创建 browser mock worker
export const worker = setupWorker(...handlers)

// 启动 mock 服务
export function startMockService() {
  return worker.start({
    onUnhandledRequest: 'bypass', // 未处理的请求直接透传
  })
}
