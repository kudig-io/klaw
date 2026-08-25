# Web前端应用

<cite>
**本文引用的文件**   
- [web/src/App.tsx](file://web/src/App.tsx)
- [web/src/main.tsx](file://web/src/main.tsx)
- [web/src/index.css](file://web/src/index.css)
- [web/vite.config.ts](file://web/vite.config.ts)
- [web/tailwind.config.js](file://web/tailwind.config.js)
- [web/postcss.config.js](file://web/postcss.config.js)
- [web/package.json](file://web/package.json)
- [web/tsconfig.json](file://web/tsconfig.json)
- [web/src/lib/api.ts](file://web/src/lib/api.ts)
- [web/src/lib/utils.ts](file://web/src/lib/utils.ts)
- [web/src/types/api.ts](file://web/src/types/api.ts)
- [web/src/pages/ClusterDashboard.tsx](file://web/src/pages/ClusterDashboard.tsx)
- [web/src/pages/PodsPage.tsx](file://web/src/pages/PodsPage.tsx)
- [web/src/pages/DeploymentsPage.tsx](file://web/src/pages/DeploymentsPage.tsx)
- [web/src/pages/ServicesPage.tsx](file://web/src/pages/ServicesPage.tsx)
- [web/src/pages/NodesPage.tsx](file://web/src/pages/NodesPage.tsx)
- [web/src/pages/MonitoringPage.tsx](file://web/src/pages/MonitoringPage.tsx)
- [web/src/pages/DiagnosticsPage.tsx](file://web/src/pages/DiagnosticsPage.tsx)
- [web/src/pages/BackupsPage.tsx](file://web/src/pages/BackupsPage.tsx)
- [web/src/pages/TenantsPage.tsx](file://web/src/pages/TenantsPage.tsx)
- [web/src/components/ClusterSelector.tsx](file://web/src/components/ClusterSelector.tsx)
- [web/src/components/NamespaceSelector.tsx](file://web/src/components/NamespaceSelector.tsx)
- [web/src/components/RefreshButton.tsx](file://web/src/components/RefreshButton.tsx)
- [web/src/components/ServiceDetailDrawer.tsx](file://web/src/components/ServiceDetailDrawer.tsx)
- [web/src/contexts/ToastContext.tsx](file://web/src/contexts/ToastContext.tsx)
- [web/src/__tests__/unit/ClusterDashboard.test.tsx](file://web/src/__tests__/unit/ClusterDashboard.test.tsx)
- [web/src/__tests__/unit/PodsPage.test.tsx](file://web/src/__tests__/unit/PodsPage.test.tsx)
- [web/src/__tests__/unit/DeploymentsPage.test.tsx](file://web/src/__tests__/unit/DeploymentsPage.test.tsx)
- [web/src/__tests__/unit/NodesPage.test.tsx](file://web/src/__tests__/unit/NodesPage.test.tsx)
- [web/src/__tests__/integration/api.test.ts](file://web/src/__tests__/integration/api.test.ts)
- [web/src/__tests__/integration/error-handling.test.tsx](file://web/src/__tests__/integration/error-handling.test.tsx)
- [web/src/__tests__/mocks/browser.ts](file://web/src/__tests__/mocks/browser.ts)
- [web/src/__tests__/mocks/data.ts](file://web/src/__tests__/mocks/data.ts)
- [web/src/__tests__/mocks/handlers.ts](file://web/src/__tests__/mocks/handlers.ts)
- [web/src/__tests__/mocks/server.ts](file://web/src/__tests__/mocks/server.ts)
- [web/src/test-utils/test-utils.tsx](file://web/src/test-utils/test-utils.tsx)
- [configs/sos-faq.yaml](file://configs/sos-faq.yaml)
- [docs/superpowers/specs/2026-08-25-sos-mode-design.md](file://docs/superpowers/specs/2026-08-25-sos-mode-design.md)
</cite>

## 更新摘要
**所做更改**
- 新增 SOS 模式 Web 界面组件的详细文档，包括悬浮按钮、全屏通话页面、会话状态管理和 AudioWorklet 集成
- 更新了核心组件章节，包含 SOS 模式的完整功能说明
- 扩展了架构总览图，反映新增的 SOS 模式组件和实时语音通信链路
- 完善了详细组件分析，涵盖 SOS 模式的实现细节和技术架构
- 更新了测试覆盖范围，包含 SOS 模式的测试计划和要求

## 目录
1. [简介](#简介)
2. [项目结构](#项目结构)
3. [核心组件](#核心组件)
4. [架构总览](#架构总览)
5. [详细组件分析](#详细组件分析)
6. [SOS 模式专项](#sos-模式专项)
7. [依赖关系分析](#依赖关系分析)
8. [性能考虑](#性能考虑)
9. [故障排查指南](#故障排查指南)
10. [结论](#结论)
11. [附录](#附录)

## 简介
本文件为 Klaw 的 React + TypeScript 前端应用的完整技术文档。内容覆盖应用架构、组件设计与复用策略、状态管理方案、API 客户端封装与 UI 组件库使用、页面组件说明、路由配置、样式管理与性能优化，并提供开发环境搭建、调试技巧与扩展开发指南。目标是帮助开发者快速理解并高效扩展该前端应用。

**更新** 本次更新新增了 SOS（紧急支持）模式的 Web 界面组件，提供全双工实时语音对话能力，包括全局悬浮按钮、全屏通话页面、会话状态管理和 AudioWorklet 音频处理集成。

## 项目结构
前端位于 web 目录下，采用 Vite + React + TypeScript 构建，样式基于 Tailwind CSS，测试使用 Vitest + MSW（Mock Service Worker）。主要目录职责如下：
- src/main.tsx：应用入口，挂载根组件与全局上下文
- src/App.tsx：路由与页面组织，包含导航菜单和 Mock 模式切换
- src/pages：页面级组件（仪表盘、部署、节点、Pod、服务、备份、诊断、监控、租户等）
- src/components：可复用业务组件（集群选择器、命名空间选择器、刷新按钮、服务详情抽屉）
- src/lib：工具与 API 客户端封装
- src/types：类型定义（如 API 模型）
- src/contexts：全局上下文（如 Toast 通知）
- __tests__：单元测试与集成测试、MSW 模拟数据与服务
- test-utils：测试辅助工具

```mermaid
graph TB
A["入口 main.tsx"] --> B["根组件 App.tsx"]
B --> P1["页面: ClusterDashboard<br/>集群概览仪表盘"]
B --> P2["页面: PodsPage<br/>Pod 生命周期管理"]
B --> P3["页面: DeploymentsPage<br/>部署列表与操作"]
B --> P4["页面: ServicesPage<br/>服务发现与管理"]
B --> P5["页面: NodesPage<br/>节点资源视图"]
B --> P6["页面: MonitoringPage<br/>监控面板"]
B --> P7["页面: DiagnosticsPage<br/>诊断任务与报告"]
B --> P8["页面: BackupsPage<br/>备份任务管理"]
B --> P9["页面: TenantsPage<br/>租户管理"]
B --> S1["SOS 模式:<br/>SosFloatingButton<br/>SosCallPage"]
B --> C1["组件: ClusterSelector<br/>集群选择器"]
B --> C2["组件: NamespaceSelector<br/>命名空间选择器"]
B --> C3["组件: RefreshButton<br/>统一刷新入口"]
B --> C4["组件: ServiceDetailDrawer<br/>服务详情抽屉"]
B --> L1["lib/api.ts<br/>API 客户端"]
B --> T1["types/api.ts<br/>类型定义"]
B --> CTX["contexts/ToastContext<br/>全局通知"]
```

**图表来源**
- [web/src/main.tsx](file://web/src/main.tsx)
- [web/src/App.tsx](file://web/src/App.tsx)
- [web/src/pages/ClusterDashboard.tsx](file://web/src/pages/ClusterDashboard.tsx)
- [web/src/pages/PodsPage.tsx](file://web/src/pages/PodsPage.tsx)
- [web/src/pages/DeploymentsPage.tsx](file://web/src/pages/DeploymentsPage.tsx)
- [web/src/pages/ServicesPage.tsx](file://web/src/pages/ServicesPage.tsx)
- [web/src/pages/NodesPage.tsx](file://web/src/pages/NodesPage.tsx)
- [web/src/pages/MonitoringPage.tsx](file://web/src/pages/MonitoringPage.tsx)
- [web/src/pages/DiagnosticsPage.tsx](file://web/src/pages/DiagnosticsPage.tsx)
- [web/src/pages/BackupsPage.tsx](file://web/src/pages/BackupsPage.tsx)
- [web/src/pages/TenantsPage.tsx](file://web/src/pages/TenantsPage.tsx)
- [web/src/components/ClusterSelector.tsx](file://web/src/components/ClusterSelector.tsx)
- [web/src/components/NamespaceSelector.tsx](file://web/src/components/NamespaceSelector.tsx)
- [web/src/components/RefreshButton.tsx](file://web/src/components/RefreshButton.tsx)
- [web/src/components/ServiceDetailDrawer.tsx](file://web/src/components/ServiceDetailDrawer.tsx)
- [web/src/lib/api.ts](file://web/src/lib/api.ts)
- [web/src/types/api.ts](file://web/src/types/api.ts)
- [web/src/contexts/ToastContext.tsx](file://web/src/contexts/ToastContext.tsx)

**章节来源**
- [web/src/main.tsx](file://web/src/main.tsx)
- [web/src/App.tsx](file://web/src/App.tsx)
- [web/package.json](file://web/package.json)
- [web/vite.config.ts](file://web/vite.config.ts)
- [web/tailwind.config.js](file://web/tailwind.config.js)
- [web/postcss.config.js](file://web/postcss.config.js)
- [web/tsconfig.json](file://web/tsconfig.json)

## 核心组件
- **页面组件**
  - ClusterDashboard：集群概览仪表盘，聚合关键指标与状态，显示节点和 Pod 统计信息
  - PodsPage：Pod 生命周期管理，支持日志查看、搜索过滤、删除操作
  - DeploymentsPage：部署列表与操作，支持扩缩容、重启、状态监控
  - ServicesPage：服务发现与管理，支持服务详情查看、端口管理、端点监控
  - NodesPage：节点资源视图与管理，显示 CPU、内存使用情况
  - MonitoringPage：监控面板，集成告警规则、历史告警、评估功能
  - DiagnosticsPage：诊断任务与报告，支持集群健康检查和问题分析
  - BackupsPage：备份任务管理，支持创建、删除、查看备份记录
  - TenantsPage：租户管理，支持多租户配置、用户管理、审计日志
- **SOS 模式组件**
  - SosFloatingButton：全局右下角红色悬浮按钮，所有页面可见，点击进入 `/sos`
  - SosCallPage：全屏通话页面，中央头像 + 呼吸/波形动画、连接状态显示、双向实时字幕、底部控制条
  - useSosSession：会话状态管理 Hook，处理 WS 连接、AudioWorklet 采集、播放队列管理
  - lib/sosApi.ts：SOS 专用 API 客户端，处理 status 探测与会话 WebSocket 连接
- **通用组件**
  - ClusterSelector：用于选择目标集群，影响后续 API 请求上下文
  - NamespaceSelector：用于选择命名空间，配合资源查询过滤
  - RefreshButton：统一刷新入口，支持防抖与加载态
  - ServiceDetailDrawer：侧滑面板展示服务详情，避免页面跳转
- **上下文与工具**
  - ToastContext：提供全局消息提示能力，支持成功、警告、错误等类型
  - lib/api.ts：统一的 HTTP 请求封装
  - types/api.ts：API 数据结构定义
  - lib/utils.ts：常用工具函数

**章节来源**
- [web/src/pages/ClusterDashboard.tsx](file://web/src/pages/ClusterDashboard.tsx)
- [web/src/pages/PodsPage.tsx](file://web/src/pages/PodsPage.tsx)
- [web/src/pages/DeploymentsPage.tsx](file://web/src/pages/DeploymentsPage.tsx)
- [web/src/pages/ServicesPage.tsx](file://web/src/pages/ServicesPage.tsx)
- [web/src/pages/NodesPage.tsx](file://web/src/pages/NodesPage.tsx)
- [web/src/pages/MonitoringPage.tsx](file://web/src/pages/MonitoringPage.tsx)
- [web/src/pages/DiagnosticsPage.tsx](file://web/src/pages/DiagnosticsPage.tsx)
- [web/src/pages/BackupsPage.tsx](file://web/src/pages/BackupsPage.tsx)
- [web/src/pages/TenantsPage.tsx](file://web/src/pages/TenantsPage.tsx)
- [docs/superpowers/specs/2026-08-25-sos-mode-design.md](file://docs/superpowers/specs/2026-08-25-sos-mode-design.md)
- [web/src/components/ClusterSelector.tsx](file://web/src/components/ClusterSelector.tsx)
- [web/src/components/NamespaceSelector.tsx](file://web/src/components/NamespaceSelector.tsx)
- [web/src/components/RefreshButton.tsx](file://web/src/components/RefreshButton.tsx)
- [web/src/components/ServiceDetailDrawer.tsx](file://web/src/components/ServiceDetailDrawer.tsx)
- [web/src/contexts/ToastContext.tsx](file://web/src/contexts/ToastContext.tsx)
- [web/src/lib/api.ts](file://web/src/lib/api.ts)
- [web/src/types/api.ts](file://web/src/types/api.ts)
- [web/src/lib/utils.ts](file://web/src/lib/utils.ts)

## 架构总览
整体采用"页面组件 + 通用组件"的分层设计，通过统一的 API 客户端访问后端服务，使用 Context 提供跨组件状态（如 Toast），样式由 Tailwind CSS 驱动，构建与开发体验由 Vite 提供。

**更新** 架构已扩展以支持更多业务场景，包括集群管理、资源监控、诊断分析、备份恢复、多租户功能和 SOS 实时语音对话能力。

```mermaid
graph TB
subgraph "浏览器"
M["main.tsx"] --> A["App.tsx"]
A --> P["pages/*<br/>9个核心页面组件"]
A --> C["components/*<br/>4个通用组件"]
A --> S["SOS 模式组件<br/>SosFloatingButton + SosCallPage"]
A --> CTX["contexts/ToastContext"]
P --> L["lib/api.ts"]
C --> L
S --> SA["lib/sosApi.ts<br/>SOS API 客户端"]
L --> T["types/api.ts"]
end
subgraph "构建与样式"
V["vite.config.ts"]
TW["tailwind.config.js"]
PC["postcss.config.js"]
IC["index.css"]
end
A --> V
A --> TW
A --> PC
A --> IC
```

**图表来源**
- [web/src/main.tsx](file://web/src/main.tsx)
- [web/src/App.tsx](file://web/src/App.tsx)
- [web/src/lib/api.ts](file://web/src/lib/api.ts)
- [web/src/types/api.ts](file://web/src/types/api.ts)
- [web/vite.config.ts](file://web/vite.config.ts)
- [web/tailwind.config.js](file://web/tailwind.config.js)
- [web/postcss.config.js](file://web/postcss.config.js)
- [web/src/index.css](file://web/src/index.css)

## 详细组件分析

### API 客户端与类型系统
- api.ts：封装 HTTP 请求，统一错误处理、重试与拦截逻辑；暴露 typed 方法供页面与组件调用
- types/api.ts：集中定义 API 响应与请求体类型，确保前后端契约一致
- utils.ts：提供格式化、校验、分页、缓存等工具函数

```mermaid
sequenceDiagram
participant Page as "页面组件"
participant API as "lib/api.ts"
participant Types as "types/api.ts"
participant Backend as "后端服务"
Page->>API : 调用接口(参数, 选项)
API->>Types : 类型检查与转换
API->>Backend : 发送HTTP请求
Backend-->>API : 返回数据或错误
API-->>Page : 解析后的数据或抛出错误
```

**图表来源**
- [web/src/lib/api.ts](file://web/src/lib/api.ts)
- [web/src/types/api.ts](file://web/src/types/api.ts)

**章节来源**
- [web/src/lib/api.ts](file://web/src/lib/api.ts)
- [web/src/types/api.ts](file://web/src/types/api.ts)
- [web/src/lib/utils.ts](file://web/src/lib/utils.ts)

### 上下文与全局状态
- ToastContext：提供全局消息提示能力，支持成功、警告、错误等类型，便于在任意组件中触发通知

```mermaid
classDiagram
class ToastContext {
+show(message, type)
+hide()
+list
}
class PageComponent {
+useToast()
+render()
}
PageComponent --> ToastContext : "消费通知"
```

**图表来源**
- [web/src/contexts/ToastContext.tsx](file://web/src/contexts/ToastContext.tsx)

**章节来源**
- [web/src/contexts/ToastContext.tsx](file://web/src/contexts/ToastContext.tsx)

### 页面组件与路由
- App.tsx：定义路由与页面布局，组合各页面组件，包含 Mock 模式切换功能
- pages/*：每个页面聚焦单一业务域，内部再组合通用组件与 API 调用

**更新** 新增了9个核心页面组件，覆盖了 Kubernetes 集群管理的各个方面，以及 SOS 模式的完整实现。

```mermaid
flowchart TD
Start(["进入应用"]) --> Route["路由匹配"]
Route --> |/dashboard| Dashboard["ClusterDashboard<br/>集群概览"]
Route --> |/pods| Pods["PodsPage<br/>Pod管理"]
Route --> |/deployments| Deployments["DeploymentsPage<br/>部署管理"]
Route --> |/services| Services["ServicesPage<br/>服务管理"]
Route --> |/nodes| Nodes["NodesPage<br/>节点管理"]
Route --> |/monitoring| Monitoring["MonitoringPage<br/>监控面板"]
Route --> |/diagnostics| Diagnostics["DiagnosticsPage<br/>诊断工具"]
Route --> |/backups| Backups["BackupsPage<br/>备份管理"]
Route --> |/tenants| Tenants["TenantsPage<br/>租户管理"]
Route --> |/sos| SOS["SosCallPage<br/>SOS 语音通话"]
Dashboard --> End(["渲染完成"])
Pods --> End
Deployments --> End
Services --> End
Nodes --> End
Monitoring --> End
Diagnostics --> End
Backups --> End
Tenants --> End
SOS --> End
```

**图表来源**
- [web/src/App.tsx](file://web/src/App.tsx)
- [web/src/pages/ClusterDashboard.tsx](file://web/src/pages/ClusterDashboard.tsx)
- [web/src/pages/PodsPage.tsx](file://web/src/pages/PodsPage.tsx)
- [web/src/pages/DeploymentsPage.tsx](file://web/src/pages/DeploymentsPage.tsx)
- [web/src/pages/ServicesPage.tsx](file://web/src/pages/ServicesPage.tsx)
- [web/src/pages/NodesPage.tsx](file://web/src/pages/NodesPage.tsx)
- [web/src/pages/MonitoringPage.tsx](file://web/src/pages/MonitoringPage.tsx)
- [web/src/pages/DiagnosticsPage.tsx](file://web/src/pages/DiagnosticsPage.tsx)
- [web/src/pages/BackupsPage.tsx](file://web/src/pages/BackupsPage.tsx)
- [web/src/pages/TenantsPage.tsx](file://web/src/pages/TenantsPage.tsx)

**章节来源**
- [web/src/App.tsx](file://web/src/App.tsx)
- [web/src/pages/ClusterDashboard.tsx](file://web/src/pages/ClusterDashboard.tsx)
- [web/src/pages/PodsPage.tsx](file://web/src/pages/PodsPage.tsx)
- [web/src/pages/DeploymentsPage.tsx](file://web/src/pages/DeploymentsPage.tsx)
- [web/src/pages/ServicesPage.tsx](file://web/src/pages/ServicesPage.tsx)
- [web/src/pages/NodesPage.tsx](file://web/src/pages/NodesPage.tsx)
- [web/src/pages/MonitoringPage.tsx](file://web/src/pages/MonitoringPage.tsx)
- [web/src/pages/DiagnosticsPage.tsx](file://web/src/pages/DiagnosticsPage.tsx)
- [web/src/pages/BackupsPage.tsx](file://web/src/pages/BackupsPage.tsx)
- [web/src/pages/TenantsPage.tsx](file://web/src/pages/TenantsPage.tsx)

### 通用组件设计
- ClusterSelector：用于选择目标集群，影响后续 API 请求上下文
- NamespaceSelector：用于选择命名空间，配合资源查询过滤
- RefreshButton：统一刷新入口，支持防抖与加载态
- ServiceDetailDrawer：侧滑面板展示服务详情，避免页面跳转

**更新** 新增了 ServiceDetailDrawer 组件，提供详细的服务信息展示功能。

```mermaid
classDiagram
class ClusterSelector {
+selectedCluster
+onChange(cluster)
}
class NamespaceSelector {
+selectedNamespace
+onChange(namespace)
}
class RefreshButton {
+onRefresh()
+loading
}
class ServiceDetailDrawer {
+visible
+serviceData
+onClose()
}
```

**图表来源**
- [web/src/components/ClusterSelector.tsx](file://web/src/components/ClusterSelector.tsx)
- [web/src/components/NamespaceSelector.tsx](file://web/src/components/NamespaceSelector.tsx)
- [web/src/components/RefreshButton.tsx](file://web/src/components/RefreshButton.tsx)
- [web/src/components/ServiceDetailDrawer.tsx](file://web/src/components/ServiceDetailDrawer.tsx)

**章节来源**
- [web/src/components/ClusterSelector.tsx](file://web/src/components/ClusterSelector.tsx)
- [web/src/components/NamespaceSelector.tsx](file://web/src/components/NamespaceSelector.tsx)
- [web/src/components/RefreshButton.tsx](file://web/src/components/RefreshButton.tsx)
- [web/src/components/ServiceDetailDrawer.tsx](file://web/src/components/ServiceDetailDrawer.tsx)

### 样式管理与主题
- index.css：全局样式入口
- tailwind.config.js：Tailwind 配置，自定义主题与插件
- postcss.config.js：PostCSS 处理链配置
- vite.config.ts：构建与开发服务器配置，包含代理、路径别名等

```mermaid
flowchart TD
Dev["开发模式"] --> Vite["Vite 构建"]
Vite --> Tailwind["Tailwind CSS"]
Tailwind --> PostCSS["PostCSS 处理"]
PostCSS --> CSS["生成样式"]
CSS --> Browser["浏览器渲染"]
```

**图表来源**
- [web/src/index.css](file://web/src/index.css)
- [web/tailwind.config.js](file://web/tailwind.config.js)
- [web/postcss.config.js](file://web/postcss.config.js)
- [web/vite.config.ts](file://web/vite.config.ts)

**章节来源**
- [web/src/index.css](file://web/src/index.css)
- [web/tailwind.config.js](file://web/tailwind.config.js)
- [web/postcss.config.js](file://web/postcss.config.js)
- [web/vite.config.ts](file://web/vite.config.ts)

## SOS 模式专项

### SOS 模式概述
SOS（紧急支持）模式是 Klaw 前端应用新增的核心功能，提供全双工实时语音对话能力，定位为应急运维入口。该模式通过阿里云百炼 DashScope 的 Qwen-Omni-Realtime 系列模型，实现智能语音交互。

### 核心组件架构

```mermaid
graph TB
subgraph "浏览器端"
SB["SosFloatingButton<br/>全局悬浮按钮"]
SCP["SosCallPage<br/>全屏通话页面"]
USS["useSosSession<br/>会话状态管理"]
AW["AudioWorklet<br/>音频采集 PCM16k"]
AQ["音频队列<br/>24k 播放队列"]
SS["状态机<br/>连接/通话/重连/结束"]
end
subgraph "后端服务"
WS["WebSocket 代理<br/>/api/v1/sos/session"]
ST["会话桥接<br/>事件翻译"]
TO["工具执行<br/>集群查询"]
FAQ["语料注入<br/>instructions"]
end
subgraph "外部服务"
DS["DashScope Realtime<br/>Qwen-Omni-Realtime"]
end
SB --> SCP
SCP --> USS
USS --> AW
USS --> AQ
USS --> SS
SCP --> WS
WS --> ST
ST --> TO
ST --> FAQ
WS --> DS
```

**图表来源**
- [docs/superpowers/specs/2026-08-25-sos-mode-design.md](file://docs/superpowers/specs/2026-08-25-sos-mode-design.md)

### 技术实现要点

#### 音频处理
- **AudioWorklet 采集**：使用 PCM16k 单声道音频格式进行麦克风数据采集
- **播放队列管理**：24k 采样率的音频播放队列，使用 AudioBufferSourceNode 调度
- **智能打断**：当检测到 `speech_started` 事件时立即停止本地音频播放并清空缓冲

#### 会话状态管理
- **状态机设计**：连接中 → 通话中 → 重连中 → 已结束
- **WebSocket 连接**：同源 WebSocket `/api/v1/sos/session`，复用现有 Bearer Token 鉴权
- **断线重连**：支持自动重连机制，提高连接稳定性

#### 三层回答引擎
1. **预置语料**：从 `configs/sos-faq.yaml` 加载标准问答对
2. **集群工具**：通过 tools/function calling 获取实时集群数据
3. **模型通用知识**：兜底使用大模型的通用知识回答问题

### 配置与部署
- **语料配置**：`configs/sos-faq.yaml` 文件存储预置问答对，支持外部文件覆盖
- **环境变量**：需要配置 DashScope API Key 和相关服务地址
- **权限要求**：需要麦克风访问权限，支持回声消除、降噪和自动增益控制

**章节来源**
- [docs/superpowers/specs/2026-08-25-sos-mode-design.md](file://docs/superpowers/specs/2026-08-25-sos-mode-design.md)
- [configs/sos-faq.yaml](file://configs/sos-faq.yaml)

## 依赖关系分析
- 构建与运行时依赖
  - Vite：开发与构建
  - React + TypeScript：框架与类型
  - Tailwind CSS：原子化样式
  - Vitest + MSW：测试与 Mock
- 模块耦合
  - 页面组件依赖 lib/api.ts 与 types/api.ts
  - 通用组件被多个页面复用
  - 上下文 ToastContext 被广泛消费
  - SOS 模式组件依赖专用的 sosApi.ts 和 AudioWorklet 技术

**更新** 新增的页面组件增加了与其他模块的依赖关系，特别是与 API 客户端和类型定义的交互，以及 SOS 模式的实时通信依赖。

```mermaid
graph LR
Pages["pages/*<br/>9个页面组件"] --> API["lib/api.ts"]
Pages --> Types["types/api.ts"]
Components["components/*<br/>4个通用组件"] --> API
SOS["SOS 模式组件"] --> SOSAPI["lib/sosApi.ts"]
SOSAPI --> Types
Contexts["contexts/*"] --> Pages
Build["vite.config.ts"] --> Pages
Style["tailwind.config.js"] --> Pages
```

**图表来源**
- [web/src/pages/ClusterDashboard.tsx](file://web/src/pages/ClusterDashboard.tsx)
- [web/src/pages/PodsPage.tsx](file://web/src/pages/PodsPage.tsx)
- [web/src/pages/DeploymentsPage.tsx](file://web/src/pages/DeploymentsPage.tsx)
- [web/src/pages/ServicesPage.tsx](file://web/src/pages/ServicesPage.tsx)
- [web/src/pages/NodesPage.tsx](file://web/src/pages/NodesPage.tsx)
- [web/src/pages/MonitoringPage.tsx](file://web/src/pages/MonitoringPage.tsx)
- [web/src/pages/DiagnosticsPage.tsx](file://web/src/pages/DiagnosticsPage.tsx)
- [web/src/pages/BackupsPage.tsx](file://web/src/pages/BackupsPage.tsx)
- [web/src/pages/TenantsPage.tsx](file://web/src/pages/TenantsPage.tsx)
- [web/src/lib/api.ts](file://web/src/lib/api.ts)
- [web/src/types/api.ts](file://web/src/types/api.ts)
- [web/src/components/ClusterSelector.tsx](file://web/src/components/ClusterSelector.tsx)
- [web/src/components/NamespaceSelector.tsx](file://web/src/components/NamespaceSelector.tsx)
- [web/src/components/RefreshButton.tsx](file://web/src/components/RefreshButton.tsx)
- [web/src/components/ServiceDetailDrawer.tsx](file://web/src/components/ServiceDetailDrawer.tsx)
- [web/src/contexts/ToastContext.tsx](file://web/src/contexts/ToastContext.tsx)
- [web/vite.config.ts](file://web/vite.config.ts)
- [web/tailwind.config.js](file://web/tailwind.config.js)

**章节来源**
- [web/package.json](file://web/package.json)
- [web/vite.config.ts](file://web/vite.config.ts)
- [web/tailwind.config.js](file://web/tailwind.config.js)
- [web/postcss.config.js](file://web/postcss.config.js)
- [web/tsconfig.json](file://web/tsconfig.json)

## 性能考虑
- 网络请求
  - 使用去重与缓存策略减少重复请求
  - 合理设置超时与重试，避免雪崩
  - SOS 模式使用 WebSocket 长连接，需关注连接池管理
- 渲染优化
  - 对大数据列表进行虚拟滚动或分页
  - 使用 React.memo 与 useMemo/useCallback 减少不必要重渲染
  - SOS 通话页面的音频可视化需优化渲染频率
- 构建优化
  - 按需引入与代码分割
  - 静态资源压缩与 CDN 加速
  - AudioWorklet 模块的懒加载
- 样式优化
  - 使用 Tailwind 的 Purge 功能移除未用样式
  - 避免过度嵌套与复杂选择器
  - SOS 模式的动画效果需考虑性能影响

**更新** 新增的页面组件采用了多种性能优化策略，包括懒加载、条件渲染、状态管理等，SOS 模式特别关注音频处理和实时通信的性能优化。

## 故障排查指南
- 常见问题
  - 接口报错：检查 api.ts 的错误处理与后端返回码
  - 路由跳转异常：确认 App.tsx 路由配置与页面路径
  - 样式不生效：检查 Tailwind 配置与 index.css 引入
  - 测试失败：确认 MSW 服务启动与 mock 数据正确性
  - SOS 模式问题：检查麦克风权限、WebSocket 连接、DashScope 配置
- 调试技巧
  - 使用浏览器开发者工具 Network 面板查看请求
  - 使用 React DevTools 检查组件状态与 props
  - 使用 Vitest 运行单测与集成测试定位问题
  - 使用 MSW 拦截请求验证前端逻辑
  - SOS 模式调试：检查 AudioWorklet 日志、WebSocket 帧数据、音频队列状态

**更新** 新增了针对新页面组件的调试指南，包括集群管理、资源监控、诊断分析等功能的调试方法，以及 SOS 模式的专项调试技巧。

**章节来源**
- [web/src/lib/api.ts](file://web/src/lib/api.ts)
- [web/src/App.tsx](file://web/src/App.tsx)
- [web/src/index.css](file://web/src/index.css)
- [web/src/__tests__/integration/api.test.ts](file://web/src/__tests__/integration/api.test.ts)
- [web/src/__tests__/integration/error-handling.test.tsx](file://web/src/__tests__/integration/error-handling.test.tsx)
- [web/src/__tests__/mocks/server.ts](file://web/src/__tests__/mocks/server.ts)

## 结论
Klaw 前端应用采用清晰的模块化架构与统一的 API 封装，结合 Tailwind CSS 与 Vite 提供了高效的开发与构建体验。通过上下文与通用组件实现良好的复用性与一致性。**更新** 本次大幅扩展的前端功能涵盖了 Kubernetes 集群管理的各个方面，包括集群概览、资源管理、监控诊断、备份恢复、多租户管理和 SOS 实时语音对话能力，为用户提供了一站式的集群管理解决方案。SOS 模式的引入显著提升了应急响应效率，通过全双工实时语音交互，运维人员可以快速获取集群状态和问题诊断结果。建议在生产环境中关注网络与渲染优化，完善测试覆盖率，持续改进用户体验与稳定性。

## 附录

### 开发环境搭建
- 安装依赖
  - 在项目根目录执行包管理器安装命令
- 启动开发服务器
  - 使用 Vite 提供的脚本启动本地服务
- 构建生产版本
  - 使用构建脚本生成静态资源
- SOS 模式配置
  - 配置 DashScope API Key 和环境变量
  - 准备 sos-faq.yaml 语料文件

**章节来源**
- [web/package.json](file://web/package.json)
- [web/vite.config.ts](file://web/vite.config.ts)
- [configs/sos-faq.yaml](file://configs/sos-faq.yaml)

### 调试技巧
- 启用 Source Map
  - 在 vite.config.ts 中开启 sourcemap
- 使用 MSW 进行接口 Mock
  - 在 __tests__/mocks 下维护模拟数据与处理器
- 运行测试
  - 使用 Vitest 运行单元与集成测试
- SOS 模式调试
  - 检查浏览器控制台中的 AudioWorklet 日志
  - 使用 Network 面板监控 WebSocket 连接
  - 验证麦克风权限和音频流状态

**更新** 新增了针对新页面组件的调试技巧，包括集群管理、资源监控、诊断分析等功能的调试方法，以及 SOS 模式的专项调试指南。

**章节来源**
- [web/vite.config.ts](file://web/vite.config.ts)
- [web/src/__tests__/mocks/data.ts](file://web/src/__tests__/mocks/data.ts)
- [web/src/__tests__/mocks/handlers.ts](file://web/src/__tests__/mocks/handlers.ts)
- [web/src/__tests__/unit/ClusterDashboard.test.tsx](file://web/src/__tests__/unit/ClusterDashboard.test.tsx)
- [web/src/__tests__/unit/PodsPage.test.tsx](file://web/src/__tests__/unit/PodsPage.test.tsx)
- [web/src/__tests__/unit/DeploymentsPage.test.tsx](file://web/src/__tests__/unit/DeploymentsPage.test.tsx)
- [web/src/__tests__/unit/NodesPage.test.tsx](file://web/src/__tests__/unit/NodesPage.test.tsx)
- [web/src/test-utils/test-utils.tsx](file://web/src/test-utils/test-utils.tsx)

### 扩展开发指南
- 新增页面
  - 在 pages 目录下创建新页面组件
  - 在 App.tsx 中添加路由配置
  - 如需数据，扩展 types/api.ts 与 lib/api.ts
- 新增通用组件
  - 在 components 目录下创建组件
  - 在页面中按需引入并使用
- 样式扩展
  - 在 tailwind.config.js 中扩展主题
  - 在 index.css 中补充全局样式
- SOS 模式扩展
  - 扩展现有语料文件 configs/sos-faq.yaml
  - 添加新的集群工具函数
  - 定制通话页面的视觉样式和交互逻辑

**更新** 新增了针对新页面组件的扩展开发指南，包括如何添加新的集群管理、资源监控、诊断分析等功能，以及 SOS 模式的扩展开发方法。

**章节来源**
- [web/src/App.tsx](file://web/src/App.tsx)
- [web/src/lib/api.ts](file://web/src/lib/api.ts)
- [web/src/types/api.ts](file://web/src/types/api.ts)
- [web/tailwind.config.js](file://web/tailwind.config.js)
- [web/src/index.css](file://web/src/index.css)
- [configs/sos-faq.yaml](file://configs/sos-faq.yaml)