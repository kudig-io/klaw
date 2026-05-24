# Klaw Fusion Phase 1 Execution Status

## 已完成内容

### Phase 0: 宿主项目清理

- 清理 `klaw` 内部带 ` 2` 后缀的重复副本目录和文件
- 修复 `internal/ops` 中的帮助渲染参数错误
- 修复 `internal/ops` 在依赖未初始化时的空指针问题
- 修复 `internal/ops/test` 的旧测试签名问题
- 验证 `klaw` 的 `go test ./...` 全量通过

### Phase 1: 第一轮主干收敛

- 为 `klaw` 新增 `/api/v1` 统一路由层
- 保留旧 `/api` 路由不变，确保现有前端和外部调用兼容
- 新增统一资源接口:
  - `/api/v1/clusters/{cluster}/resources/{kind}`
  - `/api/v1/clusters/{cluster}/namespaces/{namespace}/resources/{kind}`
  - `/api/v1/clusters/{cluster}/resources/{kind}/{name}`
  - `/api/v1/clusters/{cluster}/namespaces/{namespace}/resources/{kind}/{name}`
- 统一支持的资源种类:
  - `namespaces`
  - `nodes`
  - `pods`
  - `deployments`
  - `services`
  - `events`
- 为前端新增 `v1` API 客户端入口，便于后续从旧接口逐步切换

### Phase 1.5: 分析能力接入

- 新增 Pod 日志分析接口:
  - `/api/v1/clusters/{cluster}/namespaces/{namespace}/pods/{name}/logs/analysis`
- 新增原始日志分析接口:
  - `/api/v1/analysis/logs`
- 新增集群 RBAC 分析接口:
  - `/api/v1/clusters/{cluster}/rbac/analysis`
- 将前端资源 API 实际切换到 `/api/v1`
- 在集群总览页面新增 RBAC 汇总卡片
- 在 Pod 日志面板新增日志分析摘要
- 修复 “All Namespaces” 模式下 Pod 与 Deployment 的详情/动作命名空间取值问题
- 为日志分析与 RBAC 分析新增单元测试

### Phase 1.8: 告警与扩展资源接入

- 新增基础告警规则管理器，支持规则持久化、历史记录、确认与解决
- 新增告警规则接口:
  - `/api/v1/clusters/{cluster}/alerts/rules`
  - `/api/v1/clusters/{cluster}/alerts/rules/{id}`
- 新增告警评估与状态接口:
  - `/api/v1/clusters/{cluster}/alerts/evaluate`
  - `/api/v1/clusters/{cluster}/alerts/history`
  - `/api/v1/clusters/{cluster}/alerts/stats`
  - `/api/v1/clusters/{cluster}/alerts/{id}/acknowledge`
  - `/api/v1/clusters/{cluster}/alerts/{id}/resolve`
- 将监控页从 mock 数据切换到真实告警规则、统计与历史接口
- 统一资源接口新增更多资源种类:
  - `configmaps`
  - `statefulsets`
  - `ingresses`
- 为告警规则生命周期新增基础单元测试

### Phase 2.1: 备份域接入

- 新增 `internal/backup` 模块，吸收 `etcd-guardian` 的备份域模型与存储配置结构
- 新增备份接口:
  - `/api/v1/clusters/{cluster}/backups`
  - `/api/v1/clusters/{cluster}/backups/summary`
  - `/api/v1/clusters/{cluster}/backups/{name}`
- 备份接口支持:
  - 查询备份列表
  - 查询备份摘要
  - 创建备份
  - 删除备份
- 前端新增 `Backups` 页面，接入真实备份接口与摘要数据
- 为备份管理器新增基础单元测试

### Phase 2.3: 多租户与审计接入

- 新增 `internal/tenancy` 模块，支持租户、租户用户、配额与命名空间映射的本地持久化
- 新增 `internal/audit` 模块，支持审计日志记录、筛选与统计
- 新增多租户接口:
  - `/api/v1/tenants`
  - `/api/v1/tenants/stats`
  - `/api/v1/tenants/{id}`
  - `/api/v1/tenant-users`
- 新增审计接口:
  - `/api/v1/audit/logs`
  - `/api/v1/audit/stats`
- 前端新增 `Tenants` 页面，接入租户管理、租户用户管理与审计日志视图
- 为多租户与审计模块新增基础单元测试

### Phase 2.4: 多租户运行时联动

- 为租户模型新增 `cluster` 维度，支持按集群管理租户隔离配置
- 新增租户创建/更新/删除时的 Kubernetes 资源联动:
  - Namespace 标签补齐
  - ResourceQuota 下发
  - 默认拒绝 NetworkPolicy 下发
  - Namespace 级 Role 与 RoleBinding 下发
- 前端 `Tenants` 页面新增集群选择与租户集群展示
- 审计日志已纳入租户资源下发与清理相关上下文

### Phase 2.5: 统一数据库存储

- 新增 `internal/storage` 模块，提供统一 SQLite 文档存储能力
- 后端统一收敛到 `data/klaw.db`，不再按模块分散写入 JSON 文件
- 已完成数据库化的模块:
  - `internal/alerting`
  - `internal/backup`
  - `internal/tenancy`
  - `internal/audit`
- 现阶段采用“单库 + 文档快照”模式，先统一持久化入口，再为后续细粒度表结构演进留空间

### Phase 2.6: 租户用户 Subject 联动

- 将租户用户模型从业务记录扩展为真实 Kubernetes RBAC Subject 绑定
- 租户用户已支持三类主体:
  - `User`
  - `Group`
  - `ServiceAccount`
- 新增租户用户级 RoleBinding 下发与清理逻辑，按租户命名空间范围绑定到租户级 Role
- 租户更新时会重新归一化用户主体配置，并重新应用租户用户级绑定
- 前端 `Tenants` 页面已支持配置:
  - Subject 类型
  - Subject 名称
  - ServiceAccount 命名空间
  - 用户生效命名空间列表
- 审计日志已补充租户用户 Subject 上下文与命名空间范围

## 当前架构决策

### 为什么先统一 API 而不是直接引入 `kudig` 包

- `klaw` 与 `kudig/v2-go` 当前仍是两个独立 Go 模块
- 两者 Kubernetes 依赖版本不同
- 直接包级导入会把第一阶段工作从“可控对齐”升级为“大规模依赖重整”
- 因此本阶段先统一以下内容:
  - 路由前缀
  - 资源抽象模型
  - 前端访问方式

### 本阶段对齐目标

- 向 `kudig` 的 `/api/v1` 风格收敛
- 向 `kudig/pkg/resources` 的统一资源抽象收敛
- 为后续真正的核心替换保留兼容空间

## 模块映射

| `klaw` 模块 | 当前角色 | 对齐目标 | 下一步 |
|-------------|----------|----------|--------|
| `internal/api` | 旧 API 与 Web 后端 | `kudig` 风格统一 API 门面 | 将更多动作接口迁移到统一资源模型 |
| `internal/kubernetes` | 多集群资源查询与动作 | 未来对接 `kudig` 核心资源层 | 提炼抽象接口，减少与 HTTP 直接耦合 |
| `internal/ops` | ChatOps 命令执行 | 与 `kudig/pkg/chatops` 对齐 | 已新增 service/rbac/log analysis 命令，继续统一返回格式 |
| `internal/events` | Watch 与通知 | 对接 `kudig/pkg/events` | 收敛事件模型和订阅格式 |
| `internal/loganalysis` | 新引入分析模块 | 对齐 `kudig/pkg/loganalysis` | 继续扩展模式库与诊断建议 |
| `internal/rbacanalysis` | 新引入分析模块 | 对齐 `kudig/pkg/rbacanalysis` | 继续扩展风险检测与权限建议 |
| `internal/alerting` | 新引入告警规则模块 | 对齐 `k8s-guardian` 告警能力与未来 `kudig` 监控扩展 | 后续从文件存储迁移到数据库 |
| `internal/backup` | 新引入备份域模块 | 对齐 `etcd-guardian` 备份对象模型 | 后续对接真实 Operator/对象存储能力 |
| `internal/tenancy` | 新引入多租户模块 | 对齐 `k8s-guardian` 多租户模型 | 已接入真实 K8s 资源配额、网络策略、命名空间级 RBAC 与租户用户 Subject 绑定 |
| `internal/audit` | 新引入审计模块 | 对齐 `k8s-guardian` 审计模型与 `kudig/pkg/audit` | 后续扩展安全事件与合规报告 |
| `internal/storage` | 新引入统一存储模块 | 收敛告警、备份、租户、审计状态存储 | 后续按领域拆分为更细粒度表结构 |
| `web/src/lib/api.ts` | 前端 API 入口 | 已切换到 `v1` 接口 | 后续统一响应结构和分析入口 |

## 下一阶段建议

### Phase 2

- 对齐 `internal/ops` 与 `kudig/pkg/chatops`
- 对齐资源动作接口的返回格式
- 引入更多 `kudig` 分析模块
- 将告警状态与规则存储从文件持久化迁移到数据库
- 扩展更多 `k8s-guardian` 资源面与平台能力
- 将备份记录从 SQLite 文档存储迁移到真实对象存储/Operator 状态
- 将告警、租户、审计从 SQLite 文档存储演进到更细粒度表结构
- 评估是否引入共享接口层或局部复用 `kudig` 包
- 将备份记录从 SQLite 文档存储迁移到真实对象存储/Operator 状态
- 将告警、租户、审计从 SQLite 文档存储演进到更细粒度表结构
- 将租户用户主体继续对接外部身份源/SSO

## 当前验证结果

- `go test ./...` 已在 `klaw` 目录下通过
- 新增统一资源转换、日志分析、RBAC 分析、告警规则、备份管理、多租户、审计逻辑已有基础单元测试
- 前端 `npm run build` 已通过
- 旧接口未移除，兼容性风险可控
