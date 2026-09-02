import '@testing-library/jest-dom/vitest'
import { cleanup } from '@testing-library/react'
import { afterEach } from 'vitest'
import axios from 'axios'

// axios 1.x 在 jsdom 下自动选中 fetch 适配器，其 transformRequest 函数对象
// 经 vitest worker RPC structuredClone 时抛 DataCloneError，强制回退 xhr 适配器
axios.defaults.adapter = 'xhr'

// 清理 DOM 和 MSW 在每个测试后
afterEach(() => {
  cleanup()
})
