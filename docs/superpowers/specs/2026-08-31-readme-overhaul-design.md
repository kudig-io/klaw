# README 优化设计（GitHub 开源最高标准）

- 日期：2026-08-31
- 状态：已批准（用户确认四项关键决策）
- 目标：按 GitHub 顶级开源项目标准优化 README，**信息量只增不减**

## 已批准的关键决策

1. **语言策略**：中文为主（README.md 保持中文主体），新增完整英文版 README.en.md，两版顶部互链。
2. **视觉素材**：真实 UI 截图（`npm run dev:mock` + MSW mock 数据渲染，无真实集群信息）+ Mermaid 图表（GitHub 原生渲染）。不用 GTM 设计稿、不用 AI 生成图。
3. **社区健康文件**：全套新建——CONTRIBUTING.md、SECURITY.md、CODE_OF_CONDUCT.md、Issue/PR 模板。
4. **新增板块**：同类工具对比表、Roadmap、FAQ。不加 Star History / 贡献者墙。

## 交付物清单

| 文件 | 动作 |
|---|---|
| `README.md` | 重写扩充（现有信息全部保留） |
| `README.en.md` | 新建，与中文版结构/信息对齐 |
| `docs/images/*.png` | 新建，真实 UI 截图 4-6 张 |
| `CONTRIBUTING.md` | 新建 |
| `SECURITY.md` | 新建 |
| `CODE_OF_CONDUCT.md` | 新建（Contributor Covenant v2.1） |
| `.github/ISSUE_TEMPLATE/{config,bug_report,feature_request}.yml` | 新建 |
| `.github/PULL_REQUEST_TEMPLATE.md` | 新建 |

## 新 README.md 结构

```
标题 + 语言切换行 + badges（现有 4 枚 + CI + 镜像体积）
一句话简介 + 30 秒上手命令块 + 首屏 Dashboard 截图
目录（全量更新）
为什么是 Klaw（定位 + 对比表：k9s / Headlamp / K8s Dashboard，客观维度）
核心能力（SOS / Web 控制台 / 诊断引擎 / ChatOps / 实时事件 / 平台能力 —— 原文全保留）
界面预览（4-6 张真实截图）
仓库结构
架构（ASCII 保留 + Mermaid：诊断流水线 / ChatOps 时序 / 事件管道）
快速开始（环境要求 + 三种部署方式，原文保留）
配置（全量 YAML + 多集群接入 + 环境变量 + AI 助手，原文保留）
CLI / HTTP API / ChatOps / 实时事件监控 / 前端开发 / 测试 / Makefile（原文保留）
Roadmap（源自 DEVELOPMENT_PLAN.md：已完成摘要 + 分阶段规划）
子项目 / 已知限制（原文保留）
FAQ（认证 401 / eBPF 仅 Linux / ACS Serverless 边界 / 多集群接入 / 旧 API 下线等）
文档索引 / 贡献（链接 CONTRIBUTING）/ 安全（链接 SECURITY）/ 行为准则 / 许可证 / 链接
```

## 约束

- **信息量只增不减**：现有 README 每个章节、表格、命令、链接都保留（可重排，不可删）。
- **不虚构**：对比表仅列可核验事实并注明出处口径（各项目公开文档，2026-08）；FAQ 答案全部来自仓库现有文档（README/DEVELOPMENT_PLAN/deployment README）；SECURITY.md 不虚构邮箱，使用 GitHub 私密漏洞报告；不引用不存在的 release/镜像仓库徽章。
- 截图为 mock 数据渲染的真实 UI，不含敏感信息。
- 提交策略：规格、截图、中文 README、英文 README、社区文件各自独立 commit，只暂存本任务文件。

## 验收标准

1. 旧 README 的全部章节/表格/命令/链接在新 README（中或英文版）中均能找到对应内容。
2. `grep` 校验所有相对链接指向的文件存在；Mermaid 代码块语法正确。
3. 截图文件存在、尺寸合理（宽 ≥1280），且被 README 引用。
4. 新社区文件齐备且被 README 引用。
