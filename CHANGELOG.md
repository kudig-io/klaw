# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added - SOS Mode (Voice Emergency Quick Dialog)

- **SOS 语音应急快速对话**（`internal/sos/` + Web `/sos` 通话页）
  - 全屏语音通话界面：悬浮按钮/导航双入口、双向实时字幕、静音/挂断、智能打断
  - 后端代理对接实时全双工语音模型，`sos.provider` 可切换：`dashscope`（阿里云百炼 Qwen-Omni-Realtime，默认）/ `glm`（智谱 GLM-Realtime）；PCM 16k 上行 / 24k 下行，OpenAI Realtime 兼容事件协议
  - 三层兜底回答：预置语料（`configs/sos-faq.yaml`）→ 集群工具（5 个只读/诊断 function calling）→ 模型通用知识
  - 配置：`sos:` 段（`KLAW_SOS_DASHSCOPE_API_KEY` / `KLAW_SOS_GLM_API_KEY` 环境变量注入 Key），默认关闭；会话/工具调用审计元数据

## [0.3.0] - 2026-04-02

### Added - Real-time Event Monitoring (Watch Mode)

#### Event System
- **Event Abstraction Layer** (`internal/events/source.go`)
  - Unified Event struct with Type, ResourceType, Severity
  - Flexible FilterConfig for event filtering
  - Source interface for pluggable event sources
  - Manager for multi-source management

#### Kubernetes Event Watching
- **K8s Event Source** (`internal/events/kubernetes.go`)
  - Real-time event watching via K8s Watch API
  - Support for Events, Pods, Deployments
  - Auto-reconnection on disconnect
  - Multi-cluster support

#### Event Notification
- **Event Notifier** (`internal/events/notifier.go`)
  - Rate limiting (10 events/sec, burst 20)
  - Event deduplication (5-minute window)
  - Event aggregation to prevent message storms
  - Markdown formatted output
  - Real-time push to DingTalk

#### Configuration
- **Event Config** (`configs/config.yaml`)
  - Enable/disable event watching
  - Filter by namespaces, resource types, event types
  - Filter by reasons (include/exclude)
  - Min severity threshold
  - Rate limiting settings
  - Channel routing

#### Performance Improvements
- **From Polling to Watch**: 30-60s delay → <1s latency
- **Reduced API Calls**: 90% reduction in K8s API calls
- **Stable Connection**: Long-lived connection instead of frequent short connections

---

## [0.2.0] - 2026-04-01

### Added - Service Management

#### Backend API
- **Service List API**
  - `GET /api/clusters/{cluster}/services` - List services across all namespaces
  - `GET /api/clusters/{cluster}/namespaces/{namespace}/services` - List services in specific namespace
  
- **Service Detail API**
  - `GET /api/clusters/{cluster}/namespaces/{namespace}/services/{name}` - Get service details
  - `GET /api/clusters/{cluster}/namespaces/{namespace}/services/{name}/endpoints` - Get service endpoints

- **Implementation Details**
  - Added `ListServices()`, `GetService()`, `DeleteService()` methods in `internal/kubernetes/resources.go`
  - Added handlers in `internal/api/server.go` with proper JSON response formatting
  - Supports empty namespace parameter for "All Namespaces" queries

#### Frontend - Services Page

- **ServicesPage.tsx** (`web/src/pages/ServicesPage.tsx`)
  - Service list table with cluster and namespace filtering
  - "All Namespaces" support (using `_all` as special value)
  - Service type color coding:
    - LoadBalancer: Blue
    - NodePort: Purple
    - ClusterIP: Green
    - ExternalName: Orange
  - Port display with protocol information
  - Selector labels display
  - Age formatting (days/hours/minutes/seconds)
  - Delete service functionality with confirmation
  - View service details via drawer

- **ServiceDetailDrawer.tsx** (`web/src/components/ServiceDetailDrawer.tsx`)
  - Three-tab interface: Overview / Ports / Endpoints
  - **Overview Tab**:
    - Basic info (Type, Cluster IP, Age, Namespace)
    - External IPs with copy functionality
    - Load Balancer ingress info (IP/Hostname)
    - Selector labels
    - Labels and Annotations display
  - **Ports Tab**:
    - Detailed port information
    - Port, Target Port, Node Port display
    - Protocol badges
  - **Endpoints Tab**:
    - Ready addresses with pod references
    - Not Ready addresses
    - Endpoint ports
    - Copy IP functionality
    - Refresh endpoints button

#### Shared Components

- **ClusterSelector** (`web/src/components/ClusterSelector.tsx`)
  - Dropdown for cluster selection
  - Used across all resource pages

- **NamespaceSelector** (`web/src/components/NamespaceSelector.tsx`)
  - Dropdown with "All Namespaces" option
  - Auto-populates namespaces from cluster

- **RefreshButton** (`web/src/components/RefreshButton.tsx`)
  - Consistent refresh button with loading state

- **ToastContext** (`web/src/contexts/ToastContext.tsx`)
  - Global notification system
  - Toast types: success, error, warning, info
  - Auto-dismiss after 3 seconds
  - Used for operation feedback (delete, copy, etc.)

#### API Types

- **api.ts** (`web/src/types/api.ts`)
  - Centralized TypeScript type definitions
  - Service, ServicePort, ServiceEndpoints interfaces
  - Pod, Node, Deployment, Event, Cluster types

#### Integration

- **App.tsx**
  - Added `/services` route
  - Added Services navigation item with Globe icon

- **api.ts** (`web/src/lib/api.ts`)
  - Added `serviceApi` client with methods:
    - `listServices(cluster, namespace)`
    - `getService(cluster, namespace, name)`
    - `getServiceEndpoints(cluster, namespace, name)`
    - `deleteService(cluster, namespace, name)`

### Changed

- **main.tsx** - Wrapped app with ToastProvider
- **DEVELOPMENT_PLAN.md** - Updated progress tracking

## [0.1.0] - 2026-04-01

### Added - Deployment Management

#### Backend API
- Deployment CRUD operations
- Scale deployment API (`POST /scale`)
- Restart deployment API (`POST /restart`)
- Get deployment pods API
- Get deployment status API

#### Frontend - Deployments Page
- Deployment list with status indicators
- Quick scale buttons (+/-)
- Restart functionality
- Detail panel with conditions and container info
- Search and filter capabilities

### Added - Infrastructure

#### Kind Test Environment
- `deployment/kind/cluster-config.yaml` - Kind cluster configuration
- `deployment/kind/manage.sh` - Cluster management script
- `deployment/kind/README.md` - Documentation
- Test applications: nginx, httpbin, frontend

#### Frontend Testing
- Vitest + MSW test framework
- Mock data system (`web/src/__tests__/mocks/data.ts`)
- API mock handlers (`web/src/__tests__/mocks/handlers.ts`)
- Unit tests for all pages
- Integration tests for API calls

#### Mock Data System
- `USE_MOCK` localStorage flag
- Mock data for clusters, pods, nodes, deployments, services, monitoring
- Mock response wrapper with delay simulation

### Fixed
- CORS configuration for cross-origin requests
- JSON field naming (camelCase to snake_case)
- SPA routing for React Router
- Monitoring page white screen (removed recharts, using div bar charts)
- "All Namespaces" support across all resource pages

## Initial Release

### Features
- Multi-cluster Kubernetes management
- Pod management (list, view, logs, delete)
- Node management (list, view, metrics)
- Event viewing
- Basic monitoring dashboard
- Dark mode support
- DingTalk/Lark bot integration
