# 🦞 Klaw — Kubernetes 智能运维与诊断平台

[![Go Version](https://img.shields.io/badge/Go-1.24.2-blue.svg)](https://golang.org)
[![React Version](https://img.shields.io/badge/React-18-blue.svg)](https://reactjs.org)
[![Helm Chart](https://img.shields.io/badge/Helm-1.0.0-0f1689.svg)](./helm/klaw)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](./LICENSE)

Klaw 把「集群管理控制台」「深度诊断引擎」「ChatOps 机器人」「实时事件告警」四件事装进一个二进制里。
一份配置、一次部署，既能在浏览器里点，也能在钉钉/飞书群里喊，还能在终端里 `klaw diag` 一把梭。

---

## 目录

- [核心能力](#核心能力)
- [仓库结构](#仓库结构)
- [架构](#架构)
- [快速开始](#快速开始)
  - [本地二进制](#方式一本地二进制)
  - [Docker](#方式二docker)
  - [kind + Helm（in-cluster）](#方式三kind--helmin-cluster推荐用于验证)
- [配置](#配置)
- [CLI](#cli)
- [HTTP API](#http-api)
- [ChatOps](#chatops)
- [实时事件监控](#实时事件监控)
- [前端开发](#前端开发)
- [测试](#测试)
- [Makefile 目标](#makefile-目标)
- [子项目](#子项目)
- [已知限制](#已知限制)
- [文档索引](#文档索引)

---

## 核心能力

### 🖥️ Web 管理控制台

React 18 + Vite + Tailwind 构建的单页应用，与后端二进制打包在一起（`web/dist` 由 Go 直接托管）。

| 页面 | 能力 |
|---|---|
| `ClusterDashboard` | 集群概览、节点/Pod 统计、RBAC 摘要、资源趋势 |
| `PodsPage` | 列表、搜索、详情、实时日志、删除 |
| `DeploymentsPage` | 列表、详情、扩缩容、滚动重启、关联 Pod |
| `ServicesPage` | 列表、详情、Endpoints |
| `NodesPage` | 节点状态、容量、指标 |
| `MonitoringPage` | 告警列表、历史趋势、图表 |
| `DiagnosticsPage` | 触发诊断、查看分析器结果 |
| `BackupsPage` | 集群资源备份的创建/列表/删除 |
| `TenantsPage` | 多租户与租户用户管理 |

深色模式由 `ThemeContext` 提供，支持跟随系统。

### 🔬 诊断引擎（`internal/diag`）

一条可编排的诊断流水线：**采集 → 分析 → 根因 → 报告 → 修复建议**。

- **9 大类、73 个已注册分析器**（`kernel` / `kubernetes` / `log` / `network` / `process` / `runtime` / `security` / `servicemesh` / `system`，外加 eBPF 分析器）
- **YAML 规则引擎**（`diag/rules`）：无需改代码即可新增检查项
- **根因分析**（`diag/rca`）：从散落的告警收敛到因果链
- **自动修复**（`diag/autofix`）：给出可执行的修复动作
- **多格式报告**（`diag/reporter`）：HTML / JSON / Text
- **eBPF 探针**（`diag/ebpf`）：TCP、DNS、文件 I/O 内核级观测（仅 Linux，通过 build tag 隔离）
- **AI 助手**（`diag/ai`）：接入 LLM 对诊断结果做自然语言归纳
- **镜像扫描**（`diag/scanner`）：集成 trivy
- **成本分析**（`diag/cost`）：基于云厂商定价估算资源开销
- **TUI**（`diag/tui`）：bubbletea 终端交互界面

### 💬 ChatOps（钉钉 / 飞书）

在群里 @ 机器人直接操作集群，支持命令缩写与 Markdown 富文本回复。

### ⚡ 实时事件监控

基于 Kubernetes Watch API，秒级推送；内置速率限制、去重、聚合、静音窗口，避免消息风暴。

### 🔧 平台能力

- **多集群**：一份配置管理多个 kubeconfig / context，也支持 in-cluster
- **告警规则**：规则 CRUD、手动评估、历史、确认/解决、统计
- **备份恢复**：集群资源级备份（`internal/backup`）+ etcd 级备份（`modules/etcd-guardian`）
- **自动化脚本**：脚本 CRUD、执行、历史、统计（`internal/automation`）
- **多租户 + 审计**：租户/用户模型与操作审计日志（`internal/tenancy`、`internal/audit`）
- **安全**：Bearer Token 认证、CORS 白名单、非 root 容器（UID 65532）、恒定时间 token 比较
- **可观测**：`/healthz`、`/readyz`、`/metrics`（Prometheus 文本格式，零外部依赖）
- **持久化**：内嵌 SQLite（`modernc.org/sqlite`，纯 Go，无需 CGO）

---

## 仓库结构

这是一个 monorepo，包含 5 个独立的 Go module：

| 路径 | Module | Go | 说明 |
|---|---|---|---|
| `./` | `github.com/kudig-io/klaw` | 1.24.2 | 主应用：API + Web + 诊断 + ChatOps |
| `operator/` | `.../klaw/operator` | 1.21 | Kudig Operator，CRD 驱动的诊断编排 |
| `modules/etcd-backup/` | `.../modules/etcd-backup` | 1.25 | etcd 备份/恢复客户端库 |
| `modules/etcd-guardian/` | `.../modules/etcd-guardian` | 1.26.0 | etcd 备份恢复 Operator（含 CRD、控制器、Helm Chart） |
| `modules/etcd-guardian/backend/` | `.../etcd-guardian/backend` | 1.22 | etcd-guardian 的 Gin 后端 API |

```
klaw/
├── cmd/klaw/               # CLI 入口：main / server / diag
├── internal/
│   ├── api/                # HTTP 服务器、路由、中间件、全部 handler
│   ├── kubernetes/         # 集群客户端管理（kubeconfig / in-cluster）
│   ├── diag/               # 诊断流水线（analyzer/rules/rca/autofix/reporter/ebpf/ai/...）
│   ├── events/             # Watch 事件采集与通知管道
│   ├── messaging/          # 钉钉 / 飞书通信抽象
│   ├── ops/                # ChatOps 命令路由与处理
│   ├── monitoring/         # 轮询式指标采集与告警
│   ├── alerting/           # 告警规则引擎
│   ├── backup/             # 集群资源备份
│   ├── automation/         # 自动化脚本执行引擎
│   ├── tenancy/            # 多租户
│   ├── audit/              # 审计日志与安全合规
│   ├── storage/            # SQLite 持久化与 schema 迁移
│   ├── metrics/            # 进程内指标采集
│   ├── chart/              # ASCII 图表生成
│   ├── config/             # 配置加载 + 环境变量覆盖
│   ├── openclaw/           # OpenClaw 技能管理
│   └── {log,network,rbac,storage}analysis/   # 四类专项分析器
├── web/                    # React 前端
├── operator/               # Kudig Operator（CRD: ClusterDiagnostic / NodeDiagnostic / Schedule）
├── modules/                # etcd-backup、etcd-guardian
├── helm/klaw/              # Helm Chart（含 values-kind.yaml）
├── deployment/kind/        # kind 本地集群配置与管理脚本
├── configs/                # config.yaml / config.yaml.example
├── skills/                 # OpenClaw 技能定义
└── docs/                   # 设计与实施文档
```

---

## 架构

```
                       ┌──────────────────────────────────────────────┐
   浏览器 ──────────▶  │  Web UI (React SPA, 由 Go 静态托管)          │
                       └───────────────────┬──────────────────────────┘
                                           │ /api/v1/*
   钉钉/飞书 ────────▶  ┌─────────────────▼──────────────────────────┐
                       │  HTTP Server (gorilla/mux)                  │
                       │  metrics ▸ CORS ▸ auth ▸ deprecation        │
                       └───┬───────────┬──────────┬──────────┬───────┘
                           │           │          │          │
                  ┌────────▼──┐ ┌──────▼─────┐ ┌──▼──────┐ ┌─▼─────────┐
                  │ K8s       │ │ Diag       │ │ Ops     │ │ Event     │
                  │ Manager   │ │ Pipeline   │ │ Router  │ │ Watcher   │
                  │(client-go)│ │ 73 分析器  │ │ChatOps  │ │ 限流/去重 │
                  └────┬──────┘ └──────┬─────┘ └────┬────┘ └─────┬─────┘
                       │               │            │            │
                       ▼               ▼            ▼            ▼
                  Kubernetes API   SQLite 存储   消息平台     告警推送
```

终端侧独立于 HTTP 服务：`klaw diag` 直接驱动同一套诊断流水线，输出 Text/JSON/TUI。

---

## 快速开始

### 环境要求

| 组件 | 版本 | 备注 |
|---|---|---|
| Go | 1.24+ | `modules/etcd-guardian` 需 1.26+ |
| Node.js | 18+ | 构建前端 |
| Kubernetes | 1.24+ | 或用 kind 起本地集群 |
| Docker | 可选 | 容器化部署 / kind |
| Helm | 3.x | 集群内部署 |

### 方式一：本地二进制

```bash
git clone https://github.com/kudig-io/klaw.git
cd klaw

# 一键构建前端 + 后端
make build

# 配置
cp configs/config.yaml.example configs/config.yaml
$EDITOR configs/config.yaml

# 运行（默认执行 server 子命令）
./klaw
```

访问 <http://localhost:8080>。

> 认证默认开启（`server.auth.enabled: true`），请求 `/api/*` 需带 `Authorization: Bearer <token>`。
> 详见[已知限制](#已知限制)——Web UI 当前不会自动注入该 header。

### 方式二：Docker

```bash
docker build -t kudig-io/klaw:latest .

docker run -d \
  -p 8080:8080 \
  -e KLAW_API_TOKEN='your-token' \
  -v ~/.kube/config:/home/klaw/.kube/config:ro \
  -v $(pwd)/configs/config.yaml:/app/configs/config.yaml:ro \
  kudig-io/klaw:latest
```

三阶段构建：`node:20-alpine`（前端）→ `golang:1.24-alpine`（后端，`CGO_ENABLED=0`）→ `alpine:3.20`（运行时，非 root UID 65532），最终镜像约 127MB。

网络受限时可指定模块代理：

```bash
docker build --build-arg GOPROXY=https://goproxy.cn,direct -t kudig-io/klaw:dev .
```

### 方式三：kind + Helm（in-cluster，推荐用于验证）

这条路径已端到端验证过，Klaw 以 ServiceAccount 身份通过 `rest.InClusterConfig()` 访问 API Server，无需挂载 kubeconfig。

```bash
# 1. 创建本地集群（1 control-plane + 2 worker）
kind create cluster --config deployment/kind/cluster-config.yaml

# 2. 构建镜像
docker build --build-arg GOPROXY=https://goproxy.cn,direct -t kudig-io/klaw:dev .

# 3. 加载镜像到集群节点（避免走远端仓库）
kind load docker-image kudig-io/klaw:dev --name klaw-test

# 4. 部署
helm upgrade --install klaw helm/klaw \
  -f helm/klaw/values-kind.yaml \
  -n klaw --create-namespace --wait

# 5. 访问
kubectl port-forward -n klaw svc/klaw 18080:8080
```

打开 <http://127.0.0.1:18080>。

`values-kind.yaml` 相对生产 values 的差异：

- `image.pullPolicy: Never` —— 镜像由 `kind load` 注入，禁止拉远端
- `config.server.auth.enabled: false` —— 前端暂无 token 注入能力，本地默认关闭
- `persistence.storageClass: standard` —— kind 自带的 rancher local-path
- 资源请求下调至 100m / 128Mi

生产部署使用默认 `values.yaml` 即可，Chart 会渲染出 Deployment、Service、ConfigMap、Secret（`stringData`）、PVC、ServiceAccount、ClusterRole/ClusterRoleBinding。

完整的 kind 操作说明、镜像预拉取脚本与故障排查见 [deployment/README.md](./deployment/README.md)。

---

## 配置

配置文件默认读取 `configs/config.yaml`。以下为全量结构：

```yaml
kubernetes:
  clusters:
    - name: default
      kubeconfig: ~/.kube/config   # 填 in-cluster 则使用 ServiceAccount
      context: minikube            # 留空使用 kubeconfig 的 current-context
    - name: production
      kubeconfig: ~/.kube/prod-config
      context: production

server:
  port: 8080
  auth:
    enabled: true
    token: change-me               # 建议用 KLAW_API_TOKEN 注入，避免密钥落盘
  cors:
    allowed_origins: []            # 留空 = 仅同源；可填 https://klaw.example.com

messaging:
  dingtalk:
    enabled: false
    app_key: your_app_key
    app_secret: your_app_secret
    webhook: https://oapi.dingtalk.com/robot/send?access_token=xxx
    secret: SECxxx                 # 加签密钥
    webhook_port: 8081             # 接收钉钉回调的端口
  feishu:
    enabled: false
    app_id: your_app_id
    app_secret: your_app_secret

events:
  enabled: true
  watch_types: [Pod, Deployment, Service, Node]
  namespaces: []                   # 空 = 所有命名空间
  event_types: [Warning, Error]
  reasons: [BackOff, Unhealthy, Failed, OOMKilled]
  exclude_reasons: [Scheduled, Pulling, Pulled, Created, Started]
  min_severity: warning            # info | warning | critical
  rate_limit: 10                   # 每秒最大事件数
  dedup_window: 300                # 去重窗口（秒）
  mute_duration: 10                # 同类事件静音时长（分钟）
  channels: [ops-alert]

monitoring:
  enabled: true
  interval: 60                     # 轮询周期（秒），用于图表与趋势

openclaw:
  enabled: true
  skills: ./skills
```

### 环境变量覆盖

敏感项一律优先读环境变量，便于配合 Kubernetes Secret：

| 变量 | 覆盖字段 |
|---|---|
| `KLAW_API_TOKEN` | `server.auth.token` |
| `KLAW_DINGTALK_APP_KEY` | `messaging.dingtalk.app_key` |
| `KLAW_DINGTALK_APP_SECRET` | `messaging.dingtalk.app_secret` |
| `KLAW_DINGTALK_WEBHOOK` | `messaging.dingtalk.webhook` |
| `KLAW_DINGTALK_SECRET` | `messaging.dingtalk.secret` |
| `KLAW_FEISHU_APP_ID` | `messaging.feishu.app_id` |
| `KLAW_FEISHU_APP_SECRET` | `messaging.feishu.app_secret` |

### AI 诊断助手（可选）

诊断引擎可选接入 LLM，对诊断结果生成自然语言摘要与修复建议（`klaw diag` 尾部自动输出
`=== AI 分析 ===` 段落，HTTP `/api/v1/diag/run` 返回 `ai_analysis` 字段）。全部通过环境变量
配置，未设置 `KUDIG_AI_API_KEY` 时自动禁用，不影响诊断主流程：

| 变量 | 说明 | 默认 |
|---|---|---|
| `KUDIG_AI_PROVIDER` | `openai` / `qwen` / `ollama` / `mimo` | `openai` |
| `KUDIG_AI_API_KEY` | API Key（Bearer 鉴权） | 空（禁用 AI） |
| `KUDIG_AI_BASE_URL` | 自定义 OpenAI 兼容端点 | 按 provider 自动补齐 |
| `KUDIG_AI_MODEL` | 模型名 | 按 provider 自动补齐 |
| `KUDIG_AI_TIMEOUT` | 超时（秒） | `30` |
| `KUDIG_AI_LANGUAGE` | 输出语言 `zh` / `en` | `zh` |
| `KUDIG_AI_MAX_TOKENS` | 最大生成 token 数 | `2000` |
| `KUDIG_AI_TEMPERATURE` | 采样温度 | `0.3` |

#### 使用小米 MiMo

```bash
export KUDIG_AI_PROVIDER=mimo
export KUDIG_AI_API_KEY=tp-xxxx        # MiMo 开放平台 (https://mimo.mi.com) 申请
# 模型默认 mimo-v2.5，可切换：export KUDIG_AI_MODEL=mimo-v2.5-pro
klaw diag                              # 诊断报告末尾自动附带 AI 分析
klaw diag --no-ai                      # 显式关闭本次 AI 分析
```

集群内部署时通过 Helm 注入（写入 K8s Secret，经 `envFrom` 生效）：

```bash
helm upgrade --install klaw helm/klaw \
  --set secrets.ai.provider=mimo \
  --set secrets.ai.apiKey=tp-xxxx \
  -f helm/klaw/values-kind.yaml -n klaw
```

---

## CLI

```
klaw v1.0.0-fusion

Usage:
  klaw [command]

Available Commands:
  server      启动 Web API + ChatOps 服务（无参数时的默认命令）
  diag        对集群运行诊断分析（70+ 分析器）
  version     打印版本信息
```

### `klaw server`

| Flag | 说明 |
|---|---|
| `--port int` | 覆盖 `server.port`（默认 8080） |

启动时依次初始化：配置加载 → K8s 管理器 → 监控服务 → ChatOps 路由 → 消息插件 → 事件监听 → OpenClaw 技能 → HTTP 服务器。

### `klaw diag`

| Flag | 说明 |
|---|---|
| `--kubeconfig string` | kubeconfig 路径（默认 `~/.kube/config`） |
| `--context string` | kubeconfig context |
| `--node string` | 只诊断指定节点 |
| `--namespace string` | 只诊断指定命名空间 |
| `--analyzer string` | 只运行指定分析器（逗号分隔） |
| `--exclude-analyzer string` | 排除指定分析器（逗号分隔） |
| `--json` | 以 JSON 输出 |

```bash
klaw diag                                  # 全集群诊断
klaw diag --node worker-1                  # 聚焦单节点
klaw diag --namespace production --json    # 指定命名空间，JSON 输出
klaw diag --context prod --exclude-analyzer ebpf-tcp,cis
```

---

## HTTP API

### 版本策略

- **`/api/v1/*`** —— 当前版本，新集成一律使用
- **`/api/*`** —— 旧版路径，已标记弃用。响应会带 `Deprecation: true` 与 `Sunset: 2026-12-31` 头

### 无需认证的端点

| 端点 | 说明 |
|---|---|
| `GET /healthz` | 存活探针，进程存活即 200 |
| `GET /readyz` | 就绪探针，校验 Kubernetes 客户端可用 |
| `GET /metrics` | Prometheus 文本格式（goroutine 数、内存、HTTP 请求计数、`klaw_uptime_seconds`） |

### 中间件链

```
metrics ▸ CORS ▸ auth ▸ deprecation ▸ router
```

`auth` 仅拦截 `/api` 前缀；`corsMiddleware` 在未配置白名单时不下发任何跨域头。

### 路由清单（`/api/v1` 前缀）

**集群**

```
GET    /clusters
GET    /clusters/{name}
GET    /clusters/{name}/status
GET    /clusters/{name}/metrics
GET    /clusters/{name}/namespaces
```

**Pod**

```
GET    /clusters/{c}/pods
GET    /clusters/{c}/namespaces/{ns}/pods
GET    /clusters/{c}/namespaces/{ns}/pods/{name}
GET    /clusters/{c}/namespaces/{ns}/pods/{name}/logs
GET    /clusters/{c}/namespaces/{ns}/pods/{name}/logs/analysis
DELETE /clusters/{c}/namespaces/{ns}/pods/{name}
```

**Deployment**

```
GET    /clusters/{c}/deployments
GET    /clusters/{c}/namespaces/{ns}/deployments
GET    /clusters/{c}/namespaces/{ns}/deployments/{name}
GET    /clusters/{c}/namespaces/{ns}/deployments/{name}/pods
GET    /clusters/{c}/namespaces/{ns}/deployments/{name}/status
POST   /clusters/{c}/namespaces/{ns}/deployments/{name}/scale
POST   /clusters/{c}/namespaces/{ns}/deployments/{name}/restart
```

**Service / Node / Event**

```
GET    /clusters/{c}/services
GET    /clusters/{c}/namespaces/{ns}/services
GET    /clusters/{c}/namespaces/{ns}/services/{name}
GET    /clusters/{c}/namespaces/{ns}/services/{name}/endpoints
DELETE /clusters/{c}/namespaces/{ns}/services/{name}

GET    /clusters/{c}/nodes
GET    /clusters/{c}/nodes/{name}
GET    /clusters/{c}/nodes/metrics

GET    /clusters/{c}/events
GET    /clusters/{c}/namespaces/{ns}/events
```

**通用资源访问**（`kind` ∈ pods、deployments、services、nodes、namespaces、events、configmaps、statefulsets、ingresses）

```
GET    /clusters/{c}/resources/{kind}
GET    /clusters/{c}/resources/{kind}/{name}
GET    /clusters/{c}/namespaces/{ns}/resources/{kind}
GET    /clusters/{c}/namespaces/{ns}/resources/{kind}/{name}
```

**监控与告警**

```
GET    /clusters/{c}/monitor/status
GET    /clusters/{c}/monitor/alerts
GET    /clusters/{c}/monitor/history

GET    /clusters/{c}/alerts/rules
POST   /clusters/{c}/alerts/rules
PUT    /clusters/{c}/alerts/rules/{id}
DELETE /clusters/{c}/alerts/rules/{id}
POST   /clusters/{c}/alerts/evaluate
GET    /clusters/{c}/alerts/history
GET    /clusters/{c}/alerts/stats
POST   /clusters/{c}/alerts/{id}/acknowledge
POST   /clusters/{c}/alerts/{id}/resolve
```

**备份**

```
GET    /clusters/{c}/backups
POST   /clusters/{c}/backups
GET    /clusters/{c}/backups/summary
GET    /clusters/{c}/backups/{name}
DELETE /clusters/{c}/backups/{name}
```

**分析**

```
GET    /clusters/{c}/rbac/analysis
POST   /analysis/logs
GET    /analysis/network
GET    /analysis/storage
```

**自动化**

```
GET    /automation/scripts
POST   /automation/scripts
GET    /automation/scripts/{id}
PUT    /automation/scripts/{id}
DELETE /automation/scripts/{id}
POST   /automation/scripts/{id}/execute
GET    /automation/history
GET    /automation/statistics
```

**多租户与审计**

```
GET    /tenants          POST   /tenants          GET /tenants/stats
GET    /tenants/{id}     PUT    /tenants/{id}     DELETE /tenants/{id}
GET    /tenant-users     POST   /tenant-users     DELETE /tenant-users/{id}
GET    /audit/logs       GET    /audit/stats
```

**诊断**

```
GET    /diag/run
GET    /diag/analyzers
```

### 调用示例

```bash
TOKEN=your-token
BASE=http://localhost:8080/api/v1

curl -H "Authorization: Bearer $TOKEN" $BASE/clusters
curl -H "Authorization: Bearer $TOKEN" $BASE/clusters/default/nodes
curl -H "Authorization: Bearer $TOKEN" \
     -X POST -d '{"replicas":3}' \
     $BASE/clusters/default/namespaces/default/deployments/nginx/scale
```

---

## ChatOps

### 配置钉钉机器人

1. 群设置 → 智能群助手 → 添加机器人 → 自定义
2. 安全设置选「加签」，复制签名密钥填入 `messaging.dingtalk.secret`
3. 复制 Webhook 填入 `messaging.dingtalk.webhook`
4. 在开放平台把「消息接收地址」设为 `http://<服务器IP>:8081/webhook/dingtalk`

飞书同理，填 `messaging.feishu.app_id` / `app_secret`。

### 命令

```
klaw cluster status <cluster>              # 集群状态
klaw cluster metrics <cluster>             # 资源指标
klaw cluster chart <cluster>               # 推送趋势图

klaw pod list <cluster> <ns>               # 列出 Pod
klaw pod describe <cluster> <ns> <pod>     # Pod 详情
klaw pod logs <cluster> <ns> <pod>         # 查看日志
klaw pod analyze <cluster> <ns> <pod>      # 日志智能分析
klaw pod delete <cluster> <ns> <pod>       # 删除 Pod

klaw deployment list <cluster> <ns>
klaw deployment status <cluster> <ns> <name>
klaw deployment scale <cluster> <ns> <name> <replicas>
klaw deployment restart <cluster> <ns> <name>
klaw deployment pods <cluster> <ns> <name>

klaw service list <cluster> <ns>
klaw service describe <cluster> <ns> <name>
klaw service endpoints <cluster> <ns> <name>

klaw node list <cluster>
klaw node describe <cluster> <node>
klaw node metrics <cluster>

klaw monitor status <cluster>
klaw monitor alerts <cluster>
klaw monitor chart <cluster>

klaw rbac analyze <cluster>

klaw help
```

**缩写**：`c`=cluster、`p`=pod、`n`=node、`d`=deployment、`s`/`svc`=service、`r`=rbac、`m`=monitor、`h`=help、
`ls`=list、`desc`=describe、`log`=logs、`del`/`rm`=delete。所以 `klaw p ls prod default` 等价于 `klaw pod list prod default`。

### 效果

```
@Klaw klaw cluster status production
```

```markdown
📊 **集群状态：production**

**节点：** 3 (3 Ready)
**Pod：** 12 Running / 2 Pending
**CPU：** 45% / 85%
**内存：** 60% / 75%
```

---

## 实时事件监控

```
Klaw ──Watch──▶ Kubernetes API Server
       ◀──事件流──
            ↓ 过滤（类型/命名空间/原因/严重级别）
            ↓ 速率限制
            ↓ 去重 + 聚合 + 静音窗口
            ↓
       钉钉 / 飞书推送
```

Pod 被 OOM 杀掉时，群里会收到：

```markdown
🔴 **Error** - Pod

**资源：** Pod/nginx-7d9f-x2k
**命名空间：** production
**集群：** production
**原因：** OOMKilled
**时间：** 2026-08-21 13:30:00
**消息：**
> Container nginx was OOM killed
```

### Watch vs 轮询

| 维度 | 轮询模式 | Watch 模式 |
|---|---|---|
| 事件延迟 | 30–60 秒 | < 1 秒 |
| API 调用 | 高频轮询 | 事件驱动，下降约 90% |
| 连接方式 | 短连接 | 长连接 |

两者可同时开启：`events` 负责秒级告警，`monitoring` 负责分钟级采样供图表使用。

---

## 前端开发

```bash
cd web
npm install

npm run dev          # Vite 开发服务器，端口 3000，/api 代理到 http://localhost:8080
npm run dev:mock     # 用 MSW mock 数据，无需后端
npm run build        # 生产构建到 web/dist
npm run lint         # ESLint
```

由于 Vite 已把 `/api` 反向代理到后端，同源请求不触发 CORS，通常无需额外配置。
若改用直连后端的方式联调，把 `http://localhost:3000` 加进 `server.cors.allowed_origins`。

技术栈：React 18 · TypeScript 5.2 · Vite 5 · Tailwind 3.3 · react-router-dom 6 · axios 1.6 · recharts 2.10 · lucide-react。

---

## 测试

### Go

```bash
go build ./...
go vet ./...
go test ./internal/... -count=1
make test                       # go test -v ./...
```

主仓库共 59 个 `_test.go`，覆盖 API handler、诊断分析器、规则引擎、报告生成、存储层、各专项分析器与 ChatOps。

### 前端

```bash
cd web
npm run test:run        # 单元测试
npm run test:coverage   # 覆盖率（v8）
npm run test:ui         # Vitest UI
./test.sh all           # 封装脚本：all | unit | integration | coverage | ui | watch
```

Vitest + jsdom + MSW，测试位于 `web/src/__tests__/{unit,integration}`。

### CI

`.github/workflows/ci.yml` 共 7 个 job：`go`、`operator`、`etcd-backup-module`、`etcd-guardian-module`、`frontend`、`helm`、`docker`。

---

## Makefile 目标

```bash
make build            # 构建前端 + 后端
make build-frontend   # 仅前端
make build-backend    # 仅后端
make dev              # 并行启动前后端开发服务
make run              # 构建并运行
make test             # Go 测试
make test-frontend    # 前端测试
make fmt              # go fmt + eslint --fix
make lint             # golangci-lint + eslint
make docker-build     # 构建镜像
make docker-run       # 运行容器
make helm-install     # helm install klaw ./helm/klaw
make helm-upgrade     # helm upgrade
make helm-package     # 打包 Chart
make deps             # 安装全部依赖
make help             # 查看全部目标
```

---

## 子项目

### Kudig Operator（`operator/`）

CRD 驱动的声明式诊断编排，基于 controller-runtime 0.16.3。

| CRD | 用途 |
|---|---|
| `ClusterDiagnostic` | 声明一次集群级诊断任务 |
| `NodeDiagnostic` | 声明节点级诊断任务 |
| `Schedule` | 定时触发上述诊断 |

部署：`operator/helm/kudig-operator`。示例 CR：`operator/config/examples/`。详见 [operator/README.md](./operator/README.md)。

### etcd Guardian（`modules/etcd-guardian/`）

完整的 etcd 备份恢复 Operator，自带控制器、CRD、Gin 后端 API、独立 Web UI 与 Helm Chart。可独立部署，也可作为 Klaw 的 etcd 备份能力后端。详见 `modules/etcd-guardian/README.md`。

### etcd Backup（`modules/etcd-backup/`）

轻量的 etcd 备份/恢复客户端库，供上层复用。

---

## 已知限制

1. **Web UI 不会自动携带 Bearer Token。**
   `web/src/lib/api.ts` 的 axios 实例目前没有请求拦截器，因此当 `server.auth.enabled: true` 时，浏览器访问会收到
   `401 Unauthorized: missing bearer token`。当前的规避方式是本地开发关闭认证（`values-kind.yaml` 已默认关闭），
   生产环境请置于反向代理 / Ingress 认证之后。彻底修复需要给前端补 token 输入与持久化 + 请求拦截器。

2. **eBPF 诊断仅在 Linux 可用。** 相关分析器通过 build tag 隔离，macOS / Windows 上编译不会失败，但探针不会注册。

3. **OpenClaw 技能执行为预留接口。** `internal/openclaw` 当前完成目录扫描与技能加载，执行逻辑待补全。

4. **`/api/*` 旧版路由将于 2026-12-31 下线。** 请迁移到 `/api/v1/*`。

---

## 文档索引

| 文档 | 内容 |
|---|---|
| [deployment/README.md](./deployment/README.md) | kind 本地集群、in-cluster 部署、镜像预拉取、故障排查 |
| [docs/technical-assessment-report.md](./docs/technical-assessment-report.md) | 8 维度生产就绪度评估与修复记录 |
| [docs/dingtalk-integration.md](./docs/dingtalk-integration.md) | 钉钉集成完整指南 |
| [docs/phase1-implementation-summary.md](./docs/phase1-implementation-summary.md) | 钉钉双向通信实现 |
| [docs/phase2-implementation-summary.md](./docs/phase2-implementation-summary.md) | 实时事件推送实现 |
| [docs/service-management-impl.md](./docs/service-management-impl.md) | Service 管理功能设计 |
| [docs/fusion-phase1-execution-status.md](./docs/fusion-phase1-execution-status.md) | 诊断核心融合执行状态 |
| [CHANGELOG.md](./CHANGELOG.md) | 版本变更记录 |
| [DEVELOPMENT_PLAN.md](./DEVELOPMENT_PLAN.md) | 开发计划与迭代路线 |

---

## 贡献

1. Fork 本仓库
2. 创建特性分支：`git checkout -b feature/AmazingFeature`
3. 确保 `make lint && make test` 通过
4. 提交：`git commit -m 'feat: add AmazingFeature'`
5. 推送并开启 Pull Request

---

## 许可证

[MIT License](./LICENSE) © 2026 kudig-io

## 链接

- 项目主页：<https://github.com/kudig-io/klaw>
- 问题反馈：<https://github.com/kudig-io/klaw/issues>
