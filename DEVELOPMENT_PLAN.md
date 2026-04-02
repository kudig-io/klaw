# Klaw 开发计划

本文档记录 Klaw 作为开箱即用的 Kubernetes 运维工具的开发计划和进度。

> 创建时间：2026-04-01  
> 最后更新：2026-04-01

---

## 📊 功能清单

### ✅ 已完成功能

#### Web UI
| 功能 | 状态 | 说明 | 完成时间 |
|------|------|------|----------|
| Dashboard | ✅ | 集群概览、节点/Pod 统计 | 初始版本 |
| Pods 管理 | ✅ | 查看、搜索、删除 Pod，查看日志 | 初始版本 |
| Nodes 管理 | ✅ | 查看节点状态和资源 | 初始版本 |
| Monitoring | ✅ | 监控图表、告警列表 | 初始版本 |
| 深色模式 | ✅ | 主题切换 | 初始版本 |
| **Deployments 管理** | ✅ | 列表、详情、扩缩容、重启 | 2026-04-01 |
| **Services 管理** | ✅ | 列表、详情、Endpoints | 2026-04-01 |

#### API 接口
| 功能 | 状态 | 说明 | 完成时间 |
|------|------|------|----------|
| 集群管理 | ✅ | 获取集群列表、状态、指标、命名空间 | 初始版本 |
| Pod 管理 | ✅ | 列出、详情、日志、删除 | 初始版本 |
| 节点管理 | ✅ | 列出、详情、指标 | 初始版本 |
| 事件查看 | ✅ | 获取集群/命名空间事件 | 初始版本 |
| 监控数据 | ✅ | 监控状态、告警、历史数据 | 初始版本 |
| **Deployment 管理** | ✅ | CRUD、扩缩容、重启、查看关联 Pods | 2026-04-01 |
| **Service 管理** | ✅ | 列出、详情、Endpoints | 2026-04-01 |

#### 运维命令（钉钉/飞书）
| 功能 | 状态 | 说明 | 完成时间 |
|------|------|------|----------|
| 集群命令 | ✅ | status、metrics、chart | 初始版本 |
| Pod 命令 | ✅ | list、describe、logs、delete | 初始版本 |
| 节点命令 | ✅ | list、describe、metrics | 初始版本 |
| 监控命令 | ✅ | status、alerts、chart | 初始版本 |
| **Deployment 命令** | ✅ | list、status、scale、restart、pods | 2026-04-01 |

#### 监控告警
| 功能 | 状态 | 说明 | 完成时间 |
|------|------|------|----------|
| 指标收集 | ✅ | CPU、内存、节点/Pod 状态 | 初始版本 |
| 告警检查 | ✅ | 节点 NotReady、Pod Failed/Pending | 初始版本 |
| 消息通知 | ✅ | 钉钉、飞书消息发送 | 初始版本 |
| **实时事件推送** | ✅ | Watch 模式，秒级告警 | 2026-04-02 |

---

## ✅ 迭代 1：Deployment 管理（已完成）

### 实现内容

#### 1. 后端 API
- ✅ `GET /api/clusters/{cluster}/namespaces/{namespace}/deployments` - 列出所有 Deployment
- ✅ `GET /api/clusters/{cluster}/namespaces/{namespace}/deployments/{name}` - 获取 Deployment 详情
- ✅ `POST /api/clusters/{cluster}/namespaces/{namespace}/deployments/{name}/scale` - 扩缩容
- ✅ `POST /api/clusters/{cluster}/namespaces/{namespace}/deployments/{name}/restart` - 重启
- ✅ `GET /api/clusters/{cluster}/namespaces/{namespace}/deployments/{name}/pods` - 获取关联 Pods
- ✅ `GET /api/clusters/{cluster}/namespaces/{namespace}/deployments/{name}/status` - 获取状态摘要

#### 2. 前端页面 (DeploymentsPage.tsx)
- ✅ Deployment 列表展示（名称、状态、副本数、镜像、创建时间）
- ✅ 状态指示器（Available/Progressing/Unavailable/Scaled to 0）
- ✅ 快速扩缩容（+/- 按钮）
- ✅ 重启功能
- ✅ 详情展开面板
  - 副本统计（Desired/Available/Ready/Updated）
  - 条件状态（Conditions）
  - 容器信息
- ✅ 搜索过滤
- ✅ 集群/命名空间选择

#### 3. 运维命令
- ✅ `deployment list <cluster-name> <namespace>`
- ✅ `deployment status <cluster-name> <namespace> <deployment-name>`
- ✅ `deployment scale <cluster-name> <namespace> <deployment-name> <replicas>`
- ✅ `deployment restart <cluster-name> <namespace> <deployment-name>`
- ✅ `deployment pods <cluster-name> <namespace> <deployment-name>`

#### 4. 代码变更
- `internal/kubernetes/resources.go` - 添加 Deployment 操作方法
- `internal/api/server.go` - 添加 Deployment API 路由和处理器
- `internal/ops/handler.go` - 添加 deployment 命令处理
- `web/src/lib/api.ts` - 添加 Deployment API 客户端
- `web/src/pages/DeploymentsPage.tsx` - 新增页面
- `web/src/App.tsx` - 添加路由和导航

---

## 🚧 待开发功能

#### 第二阶段：运维能力（中优先级）

##### 2. Service 管理 ✅
- [x] **Web UI**: Services 页面
  - 列出所有 Service
  - 查看 Service 详情
  - 查看 Endpoints
- [x] **API**: Service 的 CRUD 接口
  - `GET /api/clusters/{cluster}/services` (所有命名空间)
  - `GET /api/clusters/{cluster}/namespaces/{namespace}/services`
  - `GET /api/clusters/{cluster}/namespaces/{namespace}/services/{name}`
  - `GET /api/clusters/{cluster}/namespaces/{namespace}/services/{name}/endpoints`
- [ ] **运维命令**:
  - `klaw kubernetes service list <cluster> <namespace>`
  - `klaw kubernetes service describe <cluster> <namespace> <service>`

##### 3. 实时事件推送 ✅
- [x] **Watch 模式**: 从轮询升级为实时监听
  - K8s Event Watch API
  - Pod/Deployment 变化监听
  - 自动重连机制
- [x] **智能过滤**: 灵活的事件过滤配置
  - 按命名空间过滤
  - 按资源类型过滤
  - 按事件类型过滤
  - 按原因过滤
- [x] **防消息风暴**:
  - 速率限制
  - 事件去重
  - 事件聚合
- [x] **实时推送**: Markdown 格式推送到钉钉

##### 4. 图表生成增强
- [ ] 使用真实图表库生成 PNG/SVG 图片（替代目前的 ASCII 图表）
- [ ] 支持发送到钉钉和飞书的图片消息

#### 第三阶段：运维能力（中优先级）

##### 4. ConfigMap/Secret 管理
- [ ] **Web UI**: ConfigMaps/Secrets 管理页面
- [ ] **API**: 查看、创建、编辑、删除 ConfigMap/Secret
- [ ] **运维命令**: 相关的查看和管理命令

##### 5. 资源配额查看
- [ ] **Web UI**: 显示资源配额信息
- [ ] **运维命令**: `klaw cluster resources quota <cluster-name> <namespace>`

##### 6. Events 页面
- [ ] **Web UI**: 独立的事件查看页面，支持筛选和搜索
- [ ] 按类型、命名空间、时间范围筛选

#### 第四阶段：高级功能（中优先级）

##### 7. 集群安全配置
- [ ] **安全审计**: `klaw cluster security audit <cluster-name>`
- [ ] **安全策略**: `klaw cluster security policy <cluster-name> <policy>`

##### 8. RBAC 管理
- [ ] **Web UI**: 角色和权限管理界面
- [ ] **API**: ServiceAccount、Role、RoleBinding 管理

##### 9. Prometheus 集成
- [ ] 支持从 Prometheus 获取指标
- [ ] 更丰富的监控图表

#### 第五阶段：生态（低优先级）

##### 10. 集群生命周期管理
- [ ] `klaw cluster create <cluster-name> <config-file>`
- [ ] `klaw cluster delete <cluster-name>`
- [ ] `klaw cluster upgrade <cluster-name> <version>`

##### 11. OpenClaw 技能完整实现
- [ ] 实现 `ExecuteSkill` 的真正逻辑
- [ ] 动态加载和执行技能

##### 12. 日志增强
- [ ] 多容器 Pod 日志选择
- [ ] 日志下载功能
- [ ] 更强大的日志过滤和搜索

---

## 📅 迭代计划

### 迭代 2：Service 管理 ✅（已完成）
**目标**：实现 Service 的完整管理功能

**任务清单**：
1. [x] 后端 API 开发
   - [x] 在 server.go 添加 Service 相关接口
   - [x] 在 resources.go 中添加 Service 操作方法
2. [x] 前端页面开发
   - [x] 创建 ServicesPage.tsx
   - [x] 创建 ServiceDetailDrawer.tsx 详情抽屉
   - [x] 添加到路由和导航
   - [x] 创建 ToastContext 通知系统
   - [x] 创建共用组件（ClusterSelector, NamespaceSelector, RefreshButton）
3. [ ] 运维命令开发
   - [ ] 在 handler.go 中添加 service 子命令处理
4. [x] 测试验证

---

## 📝 更新日志

### 2026-04-02
- **完成 Phase 2：实时事件推送**
  - 实现 K8s Watch 模式事件监听
  - 实现事件过滤和格式化
  - 实现速率限制和去重
  - 实时推送到钉钉
  - 向后兼容（禁用事件时回退到轮询）
  - 完整配置系统

### 2026-04-01
- **完成 Phase 1：钉钉双向通信**
  - 实现 Communicator 抽象接口
  - 重构 DingTalk 为插件化架构
  - 命令路由系统
  - 命令缩写支持

- **完成迭代 2：Service 管理**
  - 实现 Service API (List/Get/Delete + Endpoints)
  - 实现 ServicesPage 页面
  - 实现 ServiceDetailDrawer 详情组件
  - 支持所有命名空间查询
  - 添加 ToastContext 通知系统
  - 添加共用组件 (ClusterSelector, NamespaceSelector, RefreshButton)
- 创建开发计划文档
- 梳理已有功能和待开发功能
- 制定四阶段开发计划
- **完成迭代 1：Deployment 管理**
  - 实现 Deployment CRUD API
  - 实现 Deployment 管理页面
  - 实现 Deployment 运维命令
- **添加 Kind 测试环境**
  - 创建 `deployment/kind/` 目录结构
  - 添加 Kind 集群配置 (`cluster-config.yaml`)
  - 添加 Kind 管理脚本 (`manage.sh`)
  - 添加详细文档 (`README.md`)
  - 配置 Klaw 连接 Kind 集群
  - 部署测试应用 (nginx, httpbin, frontend)
- **添加前端测试集**
  - 创建 `web/src/__tests__/` 测试目录结构
  - 配置 Vitest + MSW 测试框架
  - 创建完整的 Mock 数据 (`mocks/data.ts`)
  - 实现 API Mock Handlers (`mocks/handlers.ts`)
  - 编写单元测试：
    - ClusterDashboard.test.tsx
    - DeploymentsPage.test.tsx
    - PodsPage.test.tsx
    - NodesPage.test.tsx
  - 编写集成测试：
    - api.test.ts (API 调用测试)
    - error-handling.test.tsx (错误处理测试)
  - 添加测试工具函数 (`test-utils/test-utils.tsx`)
  - 创建测试运行脚本 (`test.sh`)
  - 编写详细测试文档

