---
kind: frontend_style
name: Klaw Web 前端样式系统（Tailwind CSS + 设计令牌）
category: frontend_style
scope:
    - '**'
source_files:
    - web/tailwind.config.js
    - web/src/index.css
    - web/postcss.config.js
    - web/package.json
    - web/src/App.tsx
    - web/src/lib/utils.ts
---

Klaw 前端采用基于 Tailwind CSS 3.3 的原子化样式方案，配合 PostCSS 和 Autoprefixer 构建，通过 Vite 5 进行开发构建。整体风格围绕 Kubernetes 管理控制台场景设计，提供完整的暗色模式支持和响应式布局。

**样式体系与工具链**
- 核心框架：Tailwind CSS 3.3.5，通过 `@tailwind base/components/utilities` 指令引入基础层、组件层和工具层
- 构建工具链：Vite 5 + PostCSS + Autoprefixer，TypeScript 严格模式编译
- 类名合并：使用 `clsx` + `tailwind-merge` 组合动态类名，通过 `lib/utils.ts` 中的 `cn` 函数统一处理
- 图标系统：lucide-react 矢量图标库，所有导航和功能图标均通过 SVG 组件引入

**设计令牌与主题系统**
- 颜色体系在 `tailwind.config.js` 中扩展定义：primary（主色调 50-900 阶）、success/warning/danger（语义化状态色）
- 暗色模式通过 `darkMode: 'class'` 启用，由 `document.documentElement.classList.toggle('dark')` 控制
- 全局基础样式在 `src/index.css` 的 `@layer base` 中定义，包括边框、背景色和文本色的明暗适配
- 组件样式在 `@layer components` 中集中定义：btn（含 primary/secondary/danger 变体）、card、input 等通用组件

**组件样式约定**
- 按钮组件遵循统一的 padding、圆角、过渡动画规范，通过 `transition-colors duration-200` 实现平滑交互
- 卡片组件使用白色背景配阴影边框，暗色模式下自动切换为深灰背景
- 输入框组件包含 focus 状态的 ring 效果和边框透明化处理
- 页面级组件通过 Tailwind 实用类直接组合，避免过度封装

**响应式与布局策略**
- 移动端优先的断点策略：sm/md/lg 三级响应式布局
- 导航栏在移动端折叠为汉堡菜单，桌面端显示完整导航
- 网格布局使用 `grid-cols-1 md:grid-cols-2 lg:grid-cols-3` 实现自适应卡片排列
- 最大宽度容器 `max-w-7xl mx-auto` 确保内容在宽屏下的可读性

**Mock 开发模式支持**
- 通过环境变量 `VITE_USE_MOCK=true` 和 localStorage 开关控制 Mock 模式
- Mock 模式下显示黄色标识标签，便于区分开发环境
- MSW (Mock Service Worker) 提供 API 拦截能力，支持离线开发和测试