/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_USE_MOCK: string
  readonly VITE_USE_HASH_ROUTER: string
  // 更多环境变量...
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
