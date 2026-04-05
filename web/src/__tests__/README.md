# Klaw Web 测试集

本目录包含 Klaw Web 前端的完整测试集，使用 MSW (Mock Service Worker) 进行 API 模拟。

## 目录结构

```
__tests__/
├── README.md              # 本文件
├── mocks/                 # Mock 数据和 handlers
│   ├── data.ts           # Mock 数据定义
│   ├── handlers.ts       # MSW handlers
│   ├── server.ts         # Node.js 测试服务器
│   └── browser.ts        # 浏览器 mock worker
├── unit/                  # 单元测试
│   ├── ClusterDashboard.test.tsx
│   ├── DeploymentsPage.test.tsx
│   ├── PodsPage.test.tsx
│   └── NodesPage.test.tsx
└── integration/           # 集成测试
    ├── api.test.ts
    └── error-handling.test.tsx
```

## 快速开始

### 安装依赖

```bash
cd web
npm install
```

### 运行测试

```bash
# 运行所有测试
npm test

# 运行测试一次（CI 模式）
npm run test:run

# 运行测试并查看 UI
npm run test:ui

# 生成覆盖率报告
npm run test:coverage

# 运行特定测试文件
npm test -- src/__tests__/unit/DeploymentsPage.test.tsx
```

## Mock 数据

### 集群 (Clusters)
- `kind-test` - 测试集群（3 节点）
- `production` - 生产集群

### 命名空间 (Namespaces)
- `default`
- `klaw-test`
- `kube-system`
- `ingress-nginx`

### Deployments
- `nginx` (2 副本) - nginx:alpine
- `frontend` (3 副本) - httpd:alpine
- `httpbin` (1 副本) - kennethreitz/httpbin

### Pods
- nginx pods (2 个 Running)
- frontend pods (1 个 Pending, 2 个 Running)
- httpbin pod (1 个 Running)

### 节点 (Nodes)
- `kind-test-control-plane` (4 CPU, 8Gi)
- `kind-test-worker` (4 CPU, 8Gi)
- `kind-test-worker2` (4 CPU, 8Gi)

## 在开发中使用 Mock

### 方式 1: 使用环境变量
```bash
VITE_USE_MOCK=true npm run dev
```

### 方式 2: 在代码中启动
修改 `src/main.tsx`:

```typescript
if (import.meta.env.DEV) {
  const { startMockService } = await import('./__tests__/mocks/browser')
  await startMockService()
}
```

## 编写新测试

### 单元测试示例

```typescript
import { describe, it, expect } from 'vitest'
import { screen, waitFor } from '../../test-utils/test-utils'
import { render } from '../../test-utils/test-utils'
import { server } from '../mocks/server'
import MyComponent from '../../pages/MyComponent'

// 启动 MSW
beforeAll(() => server.listen())
afterAll(() => server.close())
afterEach(() => server.resetHandlers())

describe('MyComponent', () => {
  it('should display correctly', async () => {
    render(<MyComponent />)
    
    await waitFor(() => {
      expect(screen.getByText('Expected Text')).toBeInTheDocument()
    })
  })
})
```

### 添加新的 Mock Handler

在 `mocks/handlers.ts` 中添加：

```typescript
http.get('/api/new-endpoint', () => {
  return HttpResponse.json({ data: 'mock data' })
})
```

## 测试覆盖范围

### 已覆盖
- ✅ ClusterDashboard 页面
- ✅ DeploymentsPage 页面
- ✅ PodsPage 页面
- ✅ NodesPage 页面
- ✅ 所有 API 调用
- ✅ 错误处理

### 待添加
- ⏳ MonitoringPage 页面
- ⏳ 用户交互测试（点击、输入等）
- ⏳ E2E 测试

## 常见问题

### Q: 测试运行时出现 "MSW not found" 错误
A: 确保已安装 msw 依赖：`npm install msw --save-dev`

### Q: Mock 数据不生效
A: 检查 handler URL 是否与实际 API 路径匹配

### Q: 测试超时
A: 增加 `waitFor` 的超时时间：
```typescript
await waitFor(() => { ... }, { timeout: 5000 })
```

## 相关文档

- [MSW 文档](https://mswjs.io/docs/)
- [Vitest 文档](https://vitest.dev/)
- [Testing Library](https://testing-library.com/docs/react-testing-library/intro/)
