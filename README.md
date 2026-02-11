# klaw

基于OpenClaw的Kubernetes使用和运维工具，支持钉钉或飞书进行集群管理，提供现代化的Web UI界面。

## 功能特性

- **多集群管理**：支持管理多个Kubernetes集群
- **Web UI界面**：提供现代化、响应式的Web管理界面
- **消息平台集成**：支持钉钉和飞书消息平台
- **OpenClaw技能**：基于OpenClaw扩展机制，提供Kubernetes管理技能
- **监控告警**：集群状态监控和告警通知
- **权限管理**：基于消息平台的用户认证和权限控制
- **监控曲线**：实时监控集群、节点、Pod等资源的使用情况，并生成监控曲线图
- **图表发送**：支持将监控曲线图发送到钉钉和飞书，方便查看和分析
- **运维命令**：提供丰富的运维命令，支持集群、Pod、节点、部署等资源的管理操作
- **指标收集**：自动收集Kubernetes集群的各种指标，包括CPU、内存、节点状态、Pod状态等
- **告警通知**：支持自定义告警规则，当集群出现异常时自动发送告警通知
- **深色模式**：支持亮色和深色主题切换
- **响应式设计**：完美适配桌面端和移动端设备

## 技术栈

- **后端**：Go 1.20+
- **前端**：React 18 + TypeScript + Vite
- **UI框架**：Tailwind CSS 3
- **图表库**：Recharts
- **路由**：React Router v6
- **HTTP客户端**：Axios
- **Kubernetes**：client-go
- **消息平台**：钉钉开放平台API、飞书开放平台API
- **OpenClaw**：基于OpenClaw Skills扩展机制
- **容器化**：Docker + Helm

## 快速开始

### 环境要求

- Go 1.20+
- Node.js 18+
- npm 或 yarn
- 访问Kubernetes集群的权限（~/.kube/config）

### 安装

```bash
# 克隆仓库
$ git clone https://github.com/kudig-io/klaw.git
$ cd klaw

# 构建前端
$ cd web
$ npm install
$ npm run build
$ cd ..

# 构建后端
$ go build -o klaw cmd/klaw/main.go
```

### 配置

1. 复制配置文件模板：

```bash
$ cp configs/config.yaml.example configs/config.yaml
```

2. 编辑配置文件，设置Kubernetes集群信息和消息平台配置：

```yaml
kubernetes:
  clusters:
    - name: default
      kubeconfig: ~/.kube/config
      context: your-cluster-context

messaging:
  dingtalk:
    enabled: false
    app_key: your_app_key
    app_secret: your_app_secret
    webhook: your_webhook_url
    secret: your_secret
  feishu:
    enabled: false
    app_id: your_app_id
    app_secret: your_app_secret

openclaw:
  enabled: true
  skills: ./skills

server:
  port: 8080
```

### 运行

```bash
# 启动应用
$ ./klaw
```

应用启动后，Web UI可通过 http://localhost:8080 访问。

## Web UI 功能

### 集群仪表盘

- 查看所有已配置的Kubernetes集群
- 实时显示集群状态概览
- 节点状态统计（总数、就绪数）
- Pod状态统计（运行中、待处理、失败）
- 快速访问集群详情和指标

### Pod 管理

- 按集群和命名空间浏览Pod
- 查看Pod详细信息（状态、IP、节点等）
- 实时查看Pod日志
- 删除Pod操作
- 搜索和过滤功能
- 响应式表格布局

### 节点管理

- 查看集群中所有节点
- 显示节点CPU和内存容量
- 节点状态监控
- 节点条件详情
- 节点指标实时更新

### 监控页面

- 实时CPU使用率图表
- 实时内存使用率图表
- 告警列表和详情
- 监控状态查看
- 历史数据趋势分析

### 主题和响应式

- 亮色/深色主题切换
- 移动端适配
- 平板和桌面端优化
- 直观的用户界面

## 项目结构

```
klaw/
├── cmd/                    # 命令行入口
│   └── klaw/               # 主命令
├── internal/               # 内部包
│   ├── api/                # API服务器
│   ├── kubernetes/         # K8s管理模块
│   ├── messaging/          # 消息平台集成
│   │   ├── dingtalk/       # 钉钉集成
│   │   └── feishu/         # 飞书集成
│   ├── openclaw/           # OpenClaw集成
│   ├── monitoring/         # 监控服务
│   ├── metrics/            # 指标收集
│   └── config/             # 配置管理
├── web/                    # 前端代码
│   ├── src/
│   │   ├── pages/          # 页面组件
│   │   │   ├── ClusterDashboard.tsx
│   │   │   ├── PodsPage.tsx
│   │   │   ├── NodesPage.tsx
│   │   │   └── MonitoringPage.tsx
│   │   ├── lib/            # 工具函数和API客户端
│   │   │   ├── api.ts
│   │   │   └── utils.ts
│   │   ├── App.tsx         # 应用主组件
│   │   ├── main.tsx        # 应用入口
│   │   └── index.css       # 全局样式
│   ├── dist/               # 构建输出
│   ├── package.json        # 前端依赖
│   ├── vite.config.ts      # Vite配置
│   ├── tsconfig.json       # TypeScript配置
│   └── tailwind.config.js # Tailwind配置
├── skills/                 # OpenClaw技能
│   ├── kubernetes/         # K8s管理技能
│   └── cluster/            # 集群管理技能
├── configs/                # 配置文件
│   ├── config.yaml.example
│   └── config.yaml
├── helm/                   # Helm Chart
│   └── klaw/
├── Dockerfile              # Docker构建文件
├── go.mod                  # Go依赖
└── README.md               # 项目说明
```

## 使用指南

### Web UI 使用

1. **访问界面**：打开浏览器访问 http://localhost:8080
2. **选择集群**：在各个页面顶部选择要操作的集群
3. **查看仪表盘**：首页显示所有集群的概览信息
4. **管理Pod**：在Pods页面查看、搜索、删除Pod
5. **查看节点**：在Nodes页面查看节点状态和资源
6. **监控告警**：在Monitoring页面查看实时图表和告警

### 钉钉集成

1. 在钉钉开发者平台创建机器人
2. 获取机器人的Webhook地址和密钥
3. 在配置文件中设置钉钉相关配置
4. 启动服务后，通过钉钉机器人发送命令管理Kubernetes集群

### 飞书集成

1. 在飞书开发者平台创建应用
2. 获取应用的App ID和App Secret
3. 在配置文件中设置飞书相关配置
4. 启动服务后，通过飞书应用发送命令管理Kubernetes集群

### OpenClaw技能

项目提供了以下OpenClaw技能：

- **kubernetes**：Kubernetes资源管理技能
- **cluster**：集群管理技能

### 监控曲线

项目支持实时监控Kubernetes集群，并生成监控曲线图：

- **集群监控**：监控集群的整体状态，包括节点、Pod、资源使用等
- **节点监控**：监控单个节点的CPU和内存使用情况
- **Pod监控**：监控Pod的运行状态和资源使用情况
- **图表发送**：支持将监控曲线图发送到钉钉和飞书

### 运维命令

项目提供丰富的运维命令，支持通过消息平台进行集群管理：

- **集群命令**：查看集群状态、指标、发送监控图表
- **Pod命令**：列出、描述、删除Pod，查看Pod日志
- **节点命令**：列出、描述节点，查看节点指标
- **监控命令**：启动/停止监控，查看监控状态和告警
- **资源命令**：查看资源使用情况，生成资源使用图表

## 开发指南

### 环境要求

- Go 1.20+
- Node.js 18+
- npm 或 yarn
- Docker (可选)
- Helm (可选)

### 构建项目

```bash
# 构建后端
$ go build -o klaw cmd/klaw/main.go

# 构建前端（开发模式）
$ cd web
$ npm install
$ npm run dev

# 构建前端（生产模式）
$ npm run build
```

### 开发模式

```bash
# 终端1：启动前端开发服务器
$ cd web
$ npm run dev

# 终端2：启动后端服务器
$ go run cmd/klaw/main.go
```

### 运行测试

```bash
# Go测试
$ go test ./...

# 前端测试（如果配置了）
$ cd web
$ npm test
```

### 代码规范

- Go代码遵循Go官方规范和gofmt格式化
- TypeScript代码使用ESLint进行代码检查
- 提交前请确保代码通过所有检查

## API 文档

### 集群相关

- `GET /api/clusters` - 获取所有集群列表
- `GET /api/clusters/{name}` - 获取指定集群信息
- `GET /api/clusters/{name}/status` - 获取集群状态
- `GET /api/clusters/{name}/metrics` - 获取集群指标
- `GET /api/clusters/{name}/namespaces` - 获取集群命名空间

### Pod 相关

- `GET /api/clusters/{cluster}/namespaces/{namespace}/pods` - 列出Pod
- `GET /api/clusters/{cluster}/namespaces/{namespace}/pods/{name}` - 获取Pod详情
- `GET /api/clusters/{cluster}/namespaces/{namespace}/pods/{name}/logs` - 获取Pod日志
- `DELETE /api/clusters/{cluster}/namespaces/{namespace}/pods/{name}` - 删除Pod

### 节点相关

- `GET /api/clusters/{cluster}/nodes` - 列出节点
- `GET /api/clusters/{cluster}/nodes/{name}` - 获取节点详情
- `GET /api/clusters/{cluster}/nodes/metrics` - 获取节点指标

### 事件相关

- `GET /api/clusters/{cluster}/events` - 获取集群事件
- `GET /api/clusters/{cluster}/namespaces/{namespace}/events` - 获取命名空间事件

### 监控相关

- `GET /api/monitoring/{cluster}/status` - 获取监控状态
- `GET /api/monitoring/{cluster}/alerts` - 获取告警列表
- `GET /api/monitoring/{cluster}/history` - 获取指标历史

## 部署

### Docker 部署

```bash
# 构建镜像
$ docker build -t klaw:latest .

# 运行容器
$ docker run -d \
  -p 8080:8080 \
  -v ~/.kube/config:/root/.kube/config \
  -v $(pwd)/configs/config.yaml:/app/configs/config.yaml \
  klaw:latest
```

### Helm 部署

```bash
# 添加Helm仓库（如果适用）
$ helm repo add klaw https://charts.kudig.io

# 安装
$ helm install klaw ./helm/klaw

# 升级
$ helm upgrade klaw ./helm/klaw

# 卸载
$ helm uninstall klaw
```

## 故障排查

### 常见问题

1. **无法连接到Kubernetes集群**
   - 检查kubeconfig文件路径是否正确
   - 确认集群context配置正确
   - 验证网络连接和权限

2. **前端无法加载**
   - 确认前端已构建（npm run build）
   - 检查dist目录是否存在
   - 查看浏览器控制台错误信息

3. **API请求失败**
   - 确认后端服务正常运行
   - 检查端口配置是否正确
   - 查看后端日志获取详细错误信息

4. **监控数据不更新**
   - 确认监控服务已启动
   - 检查集群连接状态
   - 查看监控服务日志

## 贡献

欢迎提交Issue和Pull Request！

贡献指南：
1. Fork本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启Pull Request

## 许可证

MIT License

## 联系方式

- 项目主页：https://github.com/kudig-io/klaw
- 问题反馈：https://github.com/kudig-io/klaw/issues
