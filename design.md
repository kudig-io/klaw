# Design — Klaw

本文件的唯一事实源地位：GTM 营销页与 Web 控制台 12 页的任何重设计必须先读本文件。
页面与本文冲突时，以本文为准；需要新增能力时先修订本文，不做页面级私改。

## Genre

modern-minimal —— 企业级基础设施产品。克制、可核验、数据优先。
禁止：渐变横幅、发光标题、庆祝动效、装饰性插图、斜体中文标题。

## 宏观结构族

页面族内共享结构骨架，只在组件原型上变化。

- **营销页（GTM/index.html）：Marquee-Hero 族**
  骨架：顶部产品栏 → 终端演示 hero → 证据链段落（管道 → 能力 → 控制台全景 10 tabs →
  ChatOps → 事件 → 部署）→ 数字带 → 页脚。
  基调：整页浅色；终端 / 控制台演示 / 群聊 / 事件流等产品实物框保持深色（`--dk-*` 组）。
  变化旋钮：段落顺序内的组件原型（终端窗 / 表格 / 标签页 / 时间线）。
  允许 Tier-A CSS 构造（终端窗、管道图），禁止真实截图之外的照片与插画。

- **App 页（web/src 12 页）：Workbench 族**
  骨架：左侧栏（品牌区 + 分组导航 + 底部外链区）+ 顶栏（页面标题 + 模式切换）+ 内容工作区。
  变化旋钮：内容区的数据组件原型（统计卡 / 表格 / 图表 / 表单 / 详情抽屉）。
  禁止一切 enrichment——功能承载页面。
  例外登记：SOS 通话页（/sos）渲染为脱离 Workbench 骨架的全屏 overlay（无侧栏无顶栏），
  因为紧急呼叫需要无干扰全屏；除此之外任何页面不得脱离骨架。

## Theme

品牌 DNA 锚点：藏蓝墨 + 电光蓝强调、青色数据高亮、语义色只出现在数据中；GTM 页浅色基调、产品实物深色框，控制台默认浅色。

### 营销页（浅色基调 + 深色产品框）

页面整体为浅色：浅灰蓝纸面、白卡片、深藏蓝墨、电光蓝强调。
终端窗、控制台演示、群聊、事件流等「产品实物」保持深色，用独立 `--dk-*` 变量组，模拟真实界面而非装饰。

```css
:root { /* 页面基调（浅色） */
  --bg:        #f6f8fb;  --bg-2: #eef2f8;
  --panel:     #ffffff;  --panel-2: #f2f5fa;
  --line:      rgba(15,34,66,.10);  --line-strong: rgba(15,34,66,.20);
  --text:      #0c1830;  --muted: #4d5f7c;  --muted-2: #8090a9;
  --blue:      #0052fa;  --blue-2: #0046d5;
  --cyan:      #4cc3ff;  /* 仅深色产品框内使用 */
  --green:     #16a34a;  --amber: #b45309;  --red: #dc2626;
  /* 浅色面语义色；深色框内写死 #2fd270 / #f5a623 / #f62a28 */
}
/* 深色产品框（终端 / 控制台演示 / 群聊 / 事件流） */
:root {
  --dk: #050b1c;  --dk-2: #071127;  --dk-panel: #081124;  --dk-panel-2: #0b1830;
  --dk-line: rgba(148,180,226,.14);  --dk-line-strong: rgba(148,180,226,.24);
  --dk-text: #f4f8ff;  --dk-muted: #9db1cc;  --dk-muted-2: #6d81a0;
}
```

### 控制台（亮 / 暗双模式，暗色与营销页同源）

```css
:root { /* 亮色：企业冷灰纸面 */
  --color-paper:      oklch(0.975 0.004 250);
  --color-paper-2:    oklch(1.00  0.000 0);    /* 卡片白 */
  --color-ink:        oklch(0.24  0.030 262);  /* 藏蓝墨，不用纯黑 */
  --color-ink-2:      oklch(0.50  0.020 258);
  --color-rule:       oklch(0.91  0.008 255);
  --color-accent:     oklch(0.55  0.26  262);
  --color-cyan:       oklch(0.62  0.15  240);  /* 亮底上的青需压深保证对比 */
  --color-focus:      var(--color-accent);
}
html.dark { /* 暗色：与营销页同源 */
  --color-paper:      oklch(0.155 0.032 262);
  --color-paper-2:    oklch(0.21  0.038 263);
  --color-ink:        oklch(0.95  0.01  250);
  --color-ink-2:      oklch(0.72  0.02  255);
  --color-rule:       oklch(0.30  0.04  262);
  --color-accent:     oklch(0.62  0.22  258);  /* 暗底提亮一档 */
  --color-cyan:       oklch(0.76  0.13  235);
}
```

强调色占比每屏 ≤ 5%（CTA、当前态、关键数字）；语义色只标记数据。

## Typography

- Display: Inter + Noto Sans SC，weight 600–700，正常体（禁 italic）
- Body: Inter + Noto Sans SC，weight 400
- Mono（数据 / 代码 / 日志 / 时间戳）: JetBrains Mono，weight 400–500
- Display tracking: -0.01em；正文 line-height 1.6
- 字号阶（营销页 hero display 用 clamp）：`clamp(2rem, 1.2rem + 3.2vw, 3.25rem)`
- 中文与西文之间不手动加空格时，保证排版引擎自动间距；数字一律走 Mono

## Spacing

4 点命名刻度（tokens 落在 tokens.css / index.css 变量）：
`--space-3xs .25rem · 2xs .5rem · xs .75rem · sm 1rem · md 1.5rem · lg 2rem · xl 3rem · 2xl 4.5rem · 3xl 7rem`
页面禁止出现魔法数；卡片内边距统一 `--space-md`，区块间隔 `--space-xl` 起。

## Motion

- 缓动：`--ease-out: cubic-bezier(0.16, 1, 0.3, 1)`；时长 120–220ms
- 营销页 reveal：fade-only，一次性；控制台：无进场动画
- reduced-motion：全部降级为 opacity-only ≤ 150ms

## 微交互姿态

- silent success：操作成功不打断、不弹庆祝；表格行 hover 一帧背景变化
- hover/focus 延迟 0ms；focus ring 统一 `2px var(--color-focus) offset 2px`

## 控制台密度与治理（2026-09 修订，评审决议）

高密度运维终端取向：以「运维逐字读数据」为第一动作排版，密度与可核验性优先于装饰。

- **密度**：表格行高 36–40px（`text-sm` + `py-2`），一屏 ≥ 15 行；单元格 padding 禁 `py-4`。
  页面主内容为列表时禁止独立 stat 卡墙——统计并入列表头工具行；独立 stat 卡每屏 ≤ 4 张。
- **数据字体**：所有数字、时间戳、资源名、容量、表达式一律 Mono（JetBrains Mono 栈）；
  正文 UI 文案用 Inter + Noto Sans SC。Mono 只用于数据，不用于正文。
- **图标**：lucide-react 是唯一图标源，`h-4 w-4`（行内）/ `h-5 w-5`（操作）两档；
  禁止 emoji 与 Unicode 符号（✓ ✕ ⚠ ℹ ✅ ❌ 💡）充当图标。
- **语义色收束**：控制台只允许 primary / success / warning / danger / info / 中性灰六组色；
  禁用 green-*、red-*(语义场景)、blue-*(语义场景)、purple-*、orange-*、indigo-*、emerald-*、
  amber-*、yellow-* 原色。语义色只标记数据；状态双编码（色点 + 文字/图标）。
- **深色层级**：外层卡 `dark:bg-gray-900`，嵌套子卡 `dark:bg-gray-800/60`，第三级用边框区分；
  禁止同色嵌套。每个 `text-gray-*` 必须有 dark: 配对（目标 ≥ 4.5:1）。
- **边框语言**：告警规则卡的严重度色条（1px 以上侧边色条）是唯一登记的语义例外；
  普通卡片、列表项、徽章禁止装饰性彩色侧边或下边框（border-b-2 等）。
- **可访问性**：icon-only 按钮必须 `aria-label`；Toast 容器 `aria-live="polite"`；
  focus ring 覆盖一切交互元素（含未用 .btn/.input 的裸 button）；表格输出语义 `<table>`。

## CTA 语音

- 主 CTA：实心 accent 圆角 6px，文案「动词 + 宾语」（打开控制台 / 触发诊断 / 立即体验）
- 副 CTA：1px rule 描边同圆角；同一视口主 CTA ≤ 1 个
- 禁止：渐变按钮、双层阴影按钮、感叹号文案

## 页面必须共享

品牌字标（Klaw + 🦞）、accent 色与占比纪律、Inter/Noto Sans SC/JetBrains Mono、
CTA 语音、卡片语言（1px rule 边框、paper-2 底、radius 8px、无重阴影）、
表格语言（表头 ink-2 小号 500、行分隔 rule、数字右对齐 Mono）、状态徽章语言。

## 页面允许差异

族内宏观结构变体、hero 原型、营销页的 Tier-A/Tier-B 构造物。主题不可换。

## 互相跳转

- GTM → 控制台：顶栏常驻按钮 + hero 主 CTA「打开控制台」→ `http://localhost:8080`，
  旁注「本地运行后可访问」
- 控制台 → GTM：左侧栏底部固定入口「产品介绍 · GTM」→ `https://bs7klknl29np.meoo.fun`（新标签）

## Exports

页面级 token 直接读取本文件 Theme/Spacing 段落；控制台以 `web/src/index.css`
变量层为运行时事实源（:root / html.dark 双组），Tailwind config 颜色映射与其对齐：
primary 即电光蓝 #0052fa 系；success / warning / danger / info 提供全阶（50–950），
供徽章与深色变体（`dark:bg-*-950/40`）使用。品牌字体经 web/index.html 引入
Inter / Noto Sans SC / JetBrains Mono，`font-mono` 栈以 JetBrains Mono 打头。
