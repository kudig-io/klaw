import '@testing-library/jest-dom/vitest'
import { cleanup } from '@testing-library/react'
import { afterEach } from 'vitest'

// 清理 DOM 和 MSW 在每个测试后
afterEach(() => {
  cleanup()
})
