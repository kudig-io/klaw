# Klaw 开发计划

本文档记录 Klaw 作为开箱即用的 Kubernetes 运维工具的开发计划和进度。

> 创建时间：2026-04-01
> 最后更新：2026-09-22

---

## 📊 功能清单

### ✅ 已完成功能

#### Web UI（13 个页面）
| 功能 | 状态 | 说明 | 完成时间 |
|------|------|------|----------|
| Cluster Dashboard | ✅ | 集群概览、节点/Pod 统计、RBAC 摘要 | 初始版本 |
| Pods 管理 | ✅ | 查看、搜索、删除 Pod，查看日志 | 初始版本 |
| Nodes 管理 | ✅ | 查看节点状态和资源 | 初始版本 |
| Deployments 管理 | ✅ | 列表、详情、扩缩容、重启 | 2026-04-01 |
| Services 管理 | ✅ | 列表、详情、Endpoints | 2026-04-01 |
| Monitoring | ✅ | 监控图表、告警列表 | 初始版本 |
| **Network 管理** | ✅ | NetworkPolicy/Ingress 概览、拓扑分析（当前 mock 数据） | 2026-09-03 |
| **Storage 管理** | ✅ | PV/PVC/StorageClass、容量统计（当前 mock 数据） | 2026-09-03 |
| **SOS 语音应急** | ✅ | 全屏语音通话、双向字幕、工具调用兜底 | 2026-09-04 |
| **Backups 管理** | ✅ | 备份列表/摘要 | 2026-09 前 |
| **Tenants 管理** | ✅ | 租户与租户用户管理 | 2026-09 前 |
| **Diagnostics 诊断** | ✅ | 日志分析、RBAC 分析、集群诊断 | 2026-09 前 |
| 深色模式 / 设计语言 | ✅ | 全站语义色 token、Workbench 设计语言落地 | 2026-09-03 |

#### API 接口
| 功能 | 状态 | 说明 |
|------|------|------|
| 集群管理 | ✅ | 集群列表、状态、指标、命名空间 |
| Pod / Node 管理 | ✅ | 列出、详情、日志（含日志分析）、删除 |
| Deployment 管理 | ✅ | CRUD、扩缩容、重启、关联 Pods |
| Service 管理 | ✅ | 列出、详情、Endpoints |
| 事件查看 | ✅ | 集群/命名空间事件 |
| 监控数据 | ✅ | 监控状态、告警、历史数据 |
| 告警引擎 | ✅ | 告警规则 CRUD、评估、确认/恢复、统计与历史 |
| 备份 | ✅ | 备份列表/详情/摘要 |
| 租户 | ✅ | 租户/租户用户 CRUD 与统计 |
| 审计 | ✅ | 审计日志与统计 |
| RBAC 分析 | ✅ | `/rbac/analysis` 权限分析 |
| 网络与存储 | 🚧 | 页面与 mock 契约已就绪，Go API 待实现（见迭代 3） |
| SOS 语音代理 | ✅ | dashscope（Qwen-Omni-Realtime）/ glm（GLM-Realtime）双 provider |

#### 运维命令（钉钉/飞书）
| 功能 | 状态 | 说明 |
|------|------|------|
| 集群 / Pod / 节点 / 监控命令 | ✅ | status、list、describe、logs、delete、metrics 等 |
| Deployment 命令 | ✅ | list、status、scale、restart、pods |
| 实时事件推送 | ✅ | Watch 模式 + 智能过滤 + 防消息风暴 |

#### 工程质量
| 项 | 状态 | 说明 |
|------|------|------|
| CI 安全门槛 | ✅ | govulncheck 阻断（2026-09-11 起）；依赖升级消除 GO-2026-4918/5026/5970 |
| 前端 lint 门槛 | ✅ | CI 增加 eslint（19 处历史 exhaustive-deps warning 可见不阻断） |
| 前端测试 | ✅ | Vitest + MSW，119 用例全绿 |
| 子模块化 | ✅ | etcd-guardian 以 git submodule 接入，CI fetch submodules |

---

## 🚧 待开发功能

### 迭代 3：网络与存储后端真实化（进行中，最高优先级）

Web 端 Network/Storage 页面、mock 契约与单测均已就绪，但 Go 后端尚无对应路由，页面只能跑 mock 数据。

- [ ] **Go API**：按 mock 契约实现
  - Network：NetworkPolicy / Ingress 列表与统计、连接拓扑分析
  - Storage：PV / PVC / StorageClass 列表、容量统计摘要
- [ ] **K8s 数据采集**：`internal/kubernetes` 增加对应资源采集方法
- [ ] **前端切换**：`web/src/lib/api.ts` 增加 client 函数，页面从 mock 切真实 API（mock 保留给测试）
- [ ] **端到端验证**：kind 集群实测两页渲染真实数据（深浅双模式）

### 迭代 4：可观测性补齐（API 已就绪，缺 UI）

- [ ] **Events 独立页面**：按类型/命名空间/时间范围筛选与搜索
- [ ] **RBAC 分析页面**：可视化 `/rbac/analysis` 结果

### 迭代 5：备选池（按需拉取）

- [ ] Service/Deployment 运维命令（`klaw kubernetes service list/describe` 等）
- [ ] ConfigMap/Secret 管理页面与 API
- [ ] 图表 PNG 生成（替代 ASCII 图表，支持钉钉/飞书图片消息）
- [ ] 日志增强：多容器选择、下载、更强过滤
- [ ] Prometheus 集成
- [ ] 集群生命周期管理、OpenClaw 技能完整实现

### 质量债（贯穿）

- [ ] 19 处 `react-hooks/exhaustive-deps` warning 以 useCallback 重构消化，恢复 `--max-warnings 0`

---

## 📝 维护说明

- 详细变更记录见 [CHANGELOG.md](./CHANGELOG.md)；本文档只维护「能力清单 + 迭代路线」。
- 迭代 1（Deployment 管理）、迭代 2（Service 管理）于 2026-04 完成，明细见 CHANGELOG 与 git 历史。
