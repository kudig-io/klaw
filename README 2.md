# 🦞 Klaw - Kubernetes 智能运维助手

[![Go Version](https://img.shields.io/badge/Go-1.20+-blue.svg)](https://golang.org)
[![React Version](https://img.shields.io/badge/React-18-blue.svg)](https://reactjs.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

Klaw 是一个开箱即用的 Kubernetes 运维工具，提供现代化的 Web UI 界面和 ChatOps 能力，支持通过钉钉/飞书直接管理集群，实现实时事件监控和告警推送。

## 🌟 核心特性

### 📊 Web 管理界面
- **Dashboard**：集群概览、节点/Pod 统计、资源使用趋势
- **Deployment 管理**：列表、详情、扩缩容、重启、查看关联 Pod
- **Service 管理**：列表、详情、Endpoints 展示
- **Pod 管理**：查看、搜索、删除、实时日志
- **Node 管理**：节点状态、资源容量、实时监控
- **Monitoring**：实时告警、历史趋势、图表展示
- **深色模式**：自动/手动主题切换

### 💬 ChatOps（钉钉/飞书）
- **双向通信**：在钉钉群中直接执行命令，实时获取结果
- **命令支持**：
  ```
  klaw cluster status <cluster>       # 查看集群状态
  klaw pod list <cluster> <ns>        # 列出 Pod
  klaw pod logs <cluster> <ns> <pod>  # 查看日志
  klaw pod delete <cluster> <ns> <pod># 删除 Pod
  klaw node list <cluster>            # 列出节点
  klaw monitor status <cluster>       # 查看监控状态
  ```
- **命令缩写**：`p` = `pod`, `ls` = `list`, `desc` = `describe`
- **富文本输出**：Markdown 格式、表格展示、代码块

### ⚡ 实时事件监控（Watch 模式）
- **秒级推送**：从轮询升级为 K8s Watch API，延迟 < 1 秒
- **智能过滤**：按命名空间、资源类型、事件类型、原因过滤
- **防消息风暴**：速率限制、事件去重、事件聚合
- **Markdown 告警**：美观的事件推送格式

### 🔧 运维能力
- **多集群管理**：支持同时管理多个 Kubernetes 集群
- **监控告警**：CPU、内存、节点状态、Pod 状态监控
- **自动图表生成**：集群资源使用趋势图
- **权限控制**：基于消息平台的用户认证

## 🏗️ 架构设计

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              Klaw                                           │
│  ┌──────────────┐  ┌──────────────┐  ┌─────────────────────────────────────┐│
│  │   Web UI     │  │   ChatOps    │  │         Event System                ││
│  │   (React)    │  │  (DingTalk)  │  │  ┌─────────┐  ┌─────────────────┐  ││
│  │              │  │              │  │  │  Watch  │  │  Event Notifier │  ││
│  │  Dashboard   │  │  双向通信    │  │  │ Source  │  │  • Rate Limit   │  ││
│  │  Deployments │  │  命令路由    │  │  └────┬────┘  │  • Deduplicate  │  ││
│  │  Services    │  │  Markdown    │  │       │       │  • Aggregate    │  ││
│  └──────────────┘  └──────────────┘  │       ▼       └─────────────────┘  ││
│           │                │         │  K8s Events  ──▶  DingTalk Push    ││
│           └────────────────┴─────────┴─────────────────────────────────────┘│
│                              │                                               │
│                   ┌─────────┴─────────┐                                     │
│                   ▼                   ▼                                     │
│        ┌─────────────────────┐  ┌─────────────────────┐                    │
│        │   K8s Manager       │  │   API Server        │                    │
│        │   (client-go)       │  │   (Gorilla)         │                    │
│        └─────────────────────┘  └─────────────────────┘                    │
└─────────────────────────────────────────────────────────────────────────────┘
```

## 🚀 快速开始

### 环境要求

- Go 1.20+
- Node.js 18+
- 访问 Kubernetes 集群的权限（~/.kube/config）

### 安装

```bash
# 克隆仓库
git clone https://github.com/kudig-io/klaw.git
cd klaw

# 构建前端
cd web
npm install
npm run build
cd ..

# 构建后端
go build -o klaw ./cmd/klaw
```

### 配置

1. 复制配置文件：

```bash
cp configs/config.yaml.example configs/config.yaml
```

2. 编辑 `configs/config.yaml`：

```yaml
# Kubernetes 集群配置
kubernetes:
  clusters:
    - name: kind-my-k8s
      kubeconfig: ~/.kube/config
      context: kind-my-k8s

# 钉钉集成（可选）
messaging:
  dingtalk:
    enabled: true
    webhook: "https://oapi.dingtalk.com/robot/send?access_token=xxx"
    secret: "SECxxx"
    webhook_port: 8081  # 接收钉钉消息的端口

# 实时事件监听（可选）
events:
  enabled: true
  watch_types:
    - Pod
    - Deployment
    - Service
  event_types:
    - Warning
    - Error
  min_severity: warning
  rate_limit: 10
  dedup_window: 300

server:
  port: 8080
```

### 运行

```bash
./klaw
```

输出：
```
✓ Event source registered for cluster: kind-my-k8s
✓ Event monitoring started (Watch mode)
✓ Web UI server started on port 8080

🦞 Klaw started successfully. Press Ctrl+C to exit.
```

访问 Web UI：http://localhost:8080

## 💬 钉钉集成完整指南

### 1. 创建钉钉机器人

1. 打开钉钉群设置 → 智能群助手 → 添加机器人
2. 选择「自定义」机器人
3. 设置机器人名称：Klaw
4. 安全设置选择「加签」，复制签名密钥
5. 复制 Webhook 地址

### 2. 配置回调地址

在钉钉开放平台：
1. 找到 Klaw 机器人，点击「编辑」
2. 在「消息接收地址」中填写：
   ```
   http://<服务器IP>:8081/webhook/dingtalk
   ```
3. 保存设置

### 3. 配置 Klaw

编辑 `configs/config.yaml`：

```yaml
messaging:
  dingtalk:
    enabled: true
    webhook: "https://oapi.dingtalk.com/robot/send?access_token=xxx"
    secret: "SECxxx"
    webhook_port: 8081
```

### 4. 使用示例

在钉钉群中：

```
@Klaw klaw cluster status kind-my-k8s
```

回复：
```
📊 **集群状态：kind-my-k8s**

**节点：** 3 (3 Ready)
**Pod：** 12 Running / 2 Pending
**CPU：** 45% / 85%
**内存：** 60% / 75%
```

更多命令：
```
klaw pod list kind-my-k8s default
klaw pod logs kind-my-k8s default nginx-xxx
klaw node list kind-my-k8s
klaw help
```

## ⚡ 实时事件监控

### 工作原理

Klaw 使用 Kubernetes Watch API 实时监听集群事件：

```
Klaw ──Watch──▶ Kubernetes API Server
                ◀──实时推送事件──
                  ↓
            事件过滤
                  ↓
            速率限制
                  ↓
            去重/聚合
                  ↓
            推送到钉钉
```

### 配置示例

```yaml
events:
  enabled: true
  watch_types:          # 监听的资源类型
    - Pod
    - Deployment
    - Service
    - Node
  namespaces: []        # 空表示所有命名空间
  event_types:          # 监听的事件类型
    - Warning
    - Error
  reasons:              # 关注的原因
    - BackOff
    - Unhealthy
    - Failed
    - OOMKilled
  exclude_reasons:      # 排除的原因
    - Scheduled
    - Pulling
    - Pulled
  min_severity: warning # 最小严重级别
  rate_limit: 10        # 每秒最大事件数
  dedup_window: 300     # 去重窗口（秒）
```

### 实时告警示例

当 Pod 崩溃时，钉钉立即收到：

```markdown
🔴 **Error** - Pod

**资源：** Pod/nginx-xxx
**命名空间：** default
**集群：** kind-my-k8s
**原因：** OOMKilled
**时间：** 2026-04-02 13:30:00
**消息：**
> Container nginx was OOM killed
```

## 📁 项目结构

```
klaw/
├── cmd/klaw/               # 主程序入口
├── internal/
│   ├── api/                # REST API 服务器
│   ├── kubernetes/         # K8s 客户端管理
│   ├── events/             # 事件系统（Watch 模式）
│   │   ├── source.go       # 事件抽象层
│   │   ├── kubernetes.go   # K8s 事件监听
│   │   └── notifier.go     # 事件通知器
│   ├── messaging/          # 消息平台集成
│   │   ├── interface.go    # 通信抽象接口
│   │   └── dingtalk/       # 钉钉插件
│   │       ├── client.go   # 旧版客户端
│   │       └── plugin.go   # 新版插件（双向通信）
│   ├── ops/                # 运维命令
│   │   ├── handler.go      # 命令处理器
│   │   └── router.go       # 命令路由器
│   ├── monitoring/         # 监控服务（轮询模式）
│   ├── config/             # 配置管理
│   └── openclaw/           # OpenClaw 集成
├── web/                    # React 前端
│   ├── src/
│   │   ├── pages/          # 页面组件
│   │   │   ├── ServicesPage.tsx
│   │   │   ├── DeploymentsPage.tsx
│   │   │   └── ...
│   │   └── lib/api.ts      # API 客户端
├── configs/
│   └── config.yaml         # 配置文件
├── docs/                   # 文档
│   ├── phase1-implementation-summary.md
│   ├── phase2-implementation-summary.md
│   ├── dingtalk-integration.md
│   └── service-management-impl.md
└── README.md               # 本文件
```

## 🛠️ API 文档

### 集群管理
- `GET /api/clusters` - 集群列表
- `GET /api/clusters/{name}` - 集群详情
- `GET /api/clusters/{name}/status` - 集群状态
- `GET /api/clusters/{name}/metrics` - 集群指标
- `GET /api/clusters/{name}/namespaces` - 命名空间列表

### Pod 管理
- `GET /api/clusters/{cluster}/pods` - 所有命名空间 Pod
- `GET /api/clusters/{cluster}/namespaces/{ns}/pods` - 命名空间 Pod
- `GET /api/clusters/{cluster}/namespaces/{ns}/pods/{name}` - Pod 详情
- `GET /api/clusters/{cluster}/namespaces/{ns}/pods/{name}/logs` - Pod 日志
- `DELETE /api/clusters/{cluster}/namespaces/{ns}/pods/{name}` - 删除 Pod

### Deployment 管理
- `GET /api/clusters/{cluster}/deployments` - 所有命名空间 Deployment
- `GET /api/clusters/{cluster}/namespaces/{ns}/deployments` - 命名空间 Deployment
- `GET /api/clusters/{cluster}/namespaces/{ns}/deployments/{name}` - Deployment 详情
- `POST /api/clusters/{cluster}/namespaces/{ns}/deployments/{name}/scale` - 扩缩容
- `POST /api/clusters/{cluster}/namespaces/{ns}/deployments/{name}/restart` - 重启

### Service 管理
- `GET /api/clusters/{cluster}/services` - 所有命名空间 Service
- `GET /api/clusters/{cluster}/namespaces/{ns}/services` - 命名空间 Service
- `GET /api/clusters/{cluster}/namespaces/{ns}/services/{name}` - Service 详情
- `GET /api/clusters/{cluster}/namespaces/{ns}/services/{name}/endpoints` - Endpoints

### 节点管理
- `GET /api/clusters/{cluster}/nodes` - 节点列表
- `GET /api/clusters/{cluster}/nodes/{name}` - 节点详情
- `GET /api/clusters/{cluster}/nodes/metrics` - 节点指标

### 事件与监控
- `GET /api/clusters/{cluster}/events` - 集群事件
- `GET /api/monitoring/{cluster}/status` - 监控状态
- `GET /api/monitoring/{cluster}/alerts` - 告警列表

## 🔧 高级配置

### 多集群配置

```yaml
kubernetes:
  clusters:
    - name: production
      kubeconfig: ~/.kube/config
      context: production
    - name: staging
      kubeconfig: ~/.kube/config
      context: staging
    - name: eks-cluster
      kubeconfig: ~/.kube/eks-config
      context: eks-cluster
```

### 事件过滤配置

生产环境推荐配置：

```yaml
events:
  enabled: true
  watch_types:
    - Pod
    - Deployment
    - Node
  namespaces:
    - production
    - default
  event_types:
    - Warning
    - Error
  reasons:
    - BackOff
    - Unhealthy
    - Failed
    - OOMKilled
    - CrashLoopBackOff
  exclude_reasons:
    - Scheduled
    - Pulling
    - Pulled
    - Created
    - Started
  min_severity: warning
  rate_limit: 20
  dedup_window: 600
  mute_duration: 10
  channels:
    - ops-alert
    - dev-notify
```

### 混合模式

同时启用事件监听和轮询监控：

```yaml
# 实时事件监听（秒级推送）
events:
  enabled: true
  watch_types: [Pod, Deployment]
  event_types: [Warning, Error]

# 传统监控（分钟级汇总，用于图表）
monitoring:
  enabled: true
  interval: 60
```

## 🐳 Docker 部署

```bash
# 构建镜像
docker build -t klaw:latest .

# 运行容器
docker run -d \
  -p 8080:8080 \
  -p 8081:8081 \
  -v ~/.kube/config:/root/.kube/config \
  -v $(pwd)/configs/config.yaml:/app/configs/config.yaml \
  klaw:latest
```

## 📈 性能指标

| 功能 | 轮询模式 | Watch 模式 | 提升 |
|------|---------|-----------|------|
| 事件延迟 | 30-60 秒 | < 1 秒 | 60x |
| API 调用 | 高频轮询 | 事件驱动 | 90%↓ |
| 资源占用 | CPU 密集 | 内存密集 | 更均衡 |
| 连接稳定性 | 短连接 | 长连接 | 更稳定 |

## 📚 相关文档

- [钉钉集成指南](./docs/dingtalk-integration.md) - 完整的钉钉配置和使用说明
- [Phase 1 实施总结](./docs/phase1-implementation-summary.md) - 钉钉双向通信实现
- [Phase 2 实施总结](./docs/phase2-implementation-summary.md) - 实时事件推送实现
- [Service 管理实现](./docs/service-management-impl.md) - Service 功能详细设计

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

## 📄 许可证

MIT License

## 🔗 链接

- 项目主页：https://github.com/kudig-io/klaw
- 问题反馈：https://github.com/kudig-io/klaw/issues
