---
kind: configuration_system
name: Klaw 配置系统：YAML 文件加载与环境变量覆盖
category: configuration_system
scope:
    - '**'
source_files:
    - internal/config/config.go
    - configs/config.yaml
    - configs/config.yaml.example
    - cmd/klaw/cmd_server.go
    - helm/klaw/values.yaml
    - helm/klaw/templates/configmap.yaml
    - helm/klaw/templates/deployment.yaml
---

## 1. 使用的系统与工具
- 配置文件格式：YAML（gopkg.in/yaml.v3）
- 配置加载方式：程序启动时从固定路径读取 YAML 文件，通过 struct tag 映射到 Go 结构体
- CLI 参数覆盖：使用 cobra 命令行框架，支持通过 `--port` 等 flag 覆盖部分配置项
- Kubernetes 部署：Helm Chart 将 values.yaml 渲染为 ConfigMap 和 Secret，敏感信息通过 Secret 注入
- 环境变量：部分模块（如 AI provider、notifier、collector）直接通过 os.Getenv 读取环境变量作为补充配置来源

## 2. 核心文件与包
- `internal/config/config.go`：定义所有配置结构体及 Load 函数
- `configs/config.yaml`：默认配置文件
- `configs/config.yaml.example`：示例配置文件
- `cmd/klaw/cmd_server.go`：服务启动入口，调用 config.Load 并处理 CLI 参数覆盖
- `helm/klaw/values.yaml`：Helm 默认值，定义 ConfigMap 和 Secret 的模板数据
- `helm/klaw/templates/configmap.yaml`：渲染非敏感配置到 ConfigMap
- `helm/klaw/templates/deployment.yaml`：挂载 ConfigMap 和 Secret 到容器

## 3. 架构与设计决策
- **单一配置文件**：应用启动时从 `configs/config.yaml` 加载完整配置，不支持多文件合并或分层覆盖
- **结构体映射**：配置通过 Go 结构体和 yaml tag 严格映射，提供类型安全的配置访问
- **默认值策略**：仅在 `config.Server.Port == 0` 时设置默认值 8080，其他字段无默认值
- **CLI 覆盖优先级**：`--port` 参数在配置加载后覆盖内存中的配置值，实现运行时覆盖
- **Kubernetes 部署分离**：非敏感配置通过 ConfigMap 注入，敏感信息（API Token、Webhook 密钥等）通过 Secret 注入
- **环境变量混合**：部分功能模块绕过统一配置系统，直接使用 os.Getenv 读取环境变量

## 4. 约定与约束
- 配置文件必须存在且路径固定为 `configs/config.yaml`，不存在时会返回错误
- 配置结构体字段通过 yaml tag 定义，必须与 YAML 键名严格对应
- Helm 部署时，values.yaml 中的 `config` 字段会渲染到 ConfigMap 的 `config.yaml` 文件中
- 敏感信息应放在 `secrets` 字段或通过 `existingSecret` 引用外部 Secret
- Pod 以非 root 用户运行（UID/GID 65532），ConfigMap 只读挂载
- 事件监听、消息平台等功能通过 `enabled` 字段控制开关
- 集群连接通过 kubeconfig 文件和 context 指定，支持多集群配置
- 部分模块（AI provider、notifier、collector）使用 KUDIG_* 前缀的环境变量作为配置补充