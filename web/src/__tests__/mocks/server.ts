// MSW server setup for Node.js (test environment)

import { setupServer } from 'msw/node'
import { handlers, errorHandlers, slowHandlers } from './handlers'

// 创建标准 mock 服务器
export const server = setupServer(...handlers)

// 创建错误 mock 服务器
export const errorServer = setupServer(...errorHandlers)

// 创建慢速 mock 服务器
export const slowServer = setupServer(...slowHandlers)

// 导出所有 handlers 以便自定义
export { handlers, errorHandlers, slowHandlers }
