# Design — Klaw

本文件的唯一事实源地位：GTM 营销页与 Web 控制台 10 页的任何重设计必须先读本文件。
页面与本文冲突时，以本文为准；需要新增能力时先修订本文，不做页面级私改。

## Genre

modern-minimal —— 企业级基础设施产品。克制、可核验、数据优先。
禁止：渐变横幅、发光标题、庆祝动效、装饰性插图、斜体中文标题。

## 宏观结构族

页面族内共享结构骨架，只在组件原型上变化。

- **营销页（GTM/index.html）：Marquee-Hero 族**
  骨架：顶部产品栏 → 终端演示 hero → 证据链段落（管道 → 能力 → 控制台全景 10 tabs →
  ChatOps → 事件 → 部署）→ 数字带 → 页脚。
  变化旋钮：段落顺序内的组件原型（终端窗 / 表格 / 标签页 / 时间线）。
  允许 Tier-A CSS 构造（终端窗、管道图），禁止真实截图之外的照片与插画。

- **App 页（web/src 10 页）：Workbench 族**
  骨架：左侧栏（品牌区 + 分组导航 + 底部外链区）+ 顶栏（页面标题 + 模式切换）+ 内容工作区。
  变化旋钮：内容区的数据组件原型（统计卡 / 表格 / 图表 / 表单 / 详情抽屉）。
  禁止一切 enrichment——功能承载页面。

## Theme

品牌 DNA 锚点：深蓝夜空纸面、电光蓝强调、青色数据高亮、语义色只出现在数据中。

### 营销页（固定暗色）

```css
:root {
  --color-paper:      oklch(0.155 0.032 262);  /* #010411 同源 */
  --color-paper-2:    oklch(0.19  0.035 262);  /* #040a1a 同源 */
  --color-paper-3:    oklch(0.22  0.04  265);  /* #081124 同源，面板 */
  --color-ink:        oklch(0.95  0.01  250);
  --color-ink-2:      oklch(0.72  0.02  255);
  --color-rule:       oklch(0.30  0.04  262);  /* 1px 细边框 */
  --color-accent:     oklch(0.55  0.26  262);  /* #0052fa 同源 */
  --color-accent-2:   oklch(0.60  0.20  258);  /* #187efd 同源 */
  --color-cyan:       oklch(0.76  0.13  235);  /* #4cc3ff 同源，数据高亮 */
  --color-focus:      var(--color-cyan);
  /* 语义色仅用于数据（状态、severity、图表），不用于装饰 */
  --color-ok:         oklch(0.75  0.19  150);  /* #2fd270 同源 */
  --color-warn:       oklch(0.78  0.16  75);   /* #f5a623 同源 */
  --color-crit:       oklch(0.60  0.23  27);   /* #f62a28 同源 */
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
变量层为运行时事实源，Tailwind config 颜色映射与其对齐。
