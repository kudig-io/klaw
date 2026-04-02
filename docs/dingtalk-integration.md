# 钉钉集成使用指南

## 概述

Klaw 现已支持钉钉双向通信，您可以直接在钉钉群中通过 @Klaw 或发送命令来管理 Kubernetes 集群。

## 功能特性

- ✅ **双向通信**：在钉钉中发送命令，Klaw 实时响应
- ✅ **命令执行**：查看集群状态、Pod 列表、节点信息等
- ✅ **Markdown 格式**：支持表格、代码块等富文本展示
- ✅ **权限控制**：支持配置允许访问的用户组（后续版本）

## 快速开始

### 1. 创建钉钉机器人

1. 打开钉钉群设置 → 智能群助手 → 添加机器人
2. 选择「自定义」机器人
3. 设置机器人名称（如：Klaw）
4. 选择安全设置：
   - **加签**（推荐）：复制签名密钥
   - IP 白名单：添加 Klaw 服务器 IP
5. 复制 Webhook 地址

### 2. 配置 Klaw

编辑 `configs/config.yaml`：

```yaml
messaging:
  dingtalk:
    enabled: true
    webhook: "https://oapi.dingtalk.com/robot/send?access_token=xxx"  # Webhook 地址
    secret: "SECxxx"  # 签名密钥
    webhook_port: 8081  # 接收消息的端口
```

### 3. 配置钉钉回调

1. 进入钉钉开放平台的机器人管理页面
2. 找到 Klaw 机器人，编辑设置
3. 在「消息接收地址」中填写：
   ```
   http://<your-server-ip>:8081/webhook/dingtalk
   ```
4. 保存设置

### 4. 启动 Klaw

```bash
./klaw
```

看到以下日志表示钉钉插件启动成功：
```
DingTalk plugin started, webhook: http://localhost:8081/webhook/dingtalk
```

## 使用方法

### 基本命令格式

```
@Klaw klaw <命令>
```

或

```
klaw <命令>
```

### 可用命令

#### 集群管理
```
klaw cluster status <cluster>     # 查看集群状态
klaw cluster metrics <cluster>    # 查看集群指标
```

#### Pod 管理
```
klaw pod list <cluster> <namespace>                    # 列出 Pod
klaw pod describe <cluster> <namespace> <pod>          # 查看 Pod 详情
klaw pod logs <cluster> <namespace> <pod>              # 查看 Pod 日志
klaw pod delete <cluster> <namespace> <pod>            # 删除 Pod
```

#### 节点管理
```
klaw node list <cluster>           # 列出节点
klaw node describe <cluster> <node> # 查看节点详情
klaw node metrics <cluster>        # 查看节点指标
```

#### 监控告警
```
klaw monitor status <cluster>      # 查看监控状态
klaw monitor alerts <cluster>      # 查看告警列表
```

#### 帮助
```
klaw help                          # 显示帮助信息
```

### 使用示例

**查看集群状态：**
```
@Klaw klaw cluster status kind-my-k8s
```

**列出 default 命名空间的 Pod：**
```
@Klaw klaw pod list kind-my-k8s default
```

**查看 Pod 日志：**
```
@Klaw klaw pod logs kind-my-k8s default nginx-xxx
```

## 消息格式

Klaw 支持多种消息格式：

### 表格展示
```
📦 Pod 列表 (3)

| 名称 | 状态 | 重启 | 年龄 |
|------|------|------|------|
| nginx-xxx | Running | 0 | 2d |
| redis-xxx | Running | 1 | 5h |
```

### 代码块
```
📄 Pod 日志

```
127.0.0.1 - - [02/Apr/2026:10:00:00 +0800] "GET / HTTP/1.1" 200 612
```
```

### 错误提示
```
❌ **执行出错**

```
pod "xxx" not found
```
```

## 高级配置

### 命令前缀

默认命令前缀是 `klaw`，可以在代码中修改：

```go
commandRouter.SetPrefix("k8s")  // 修改为 k8s
```

### @名称

默认 @名称是 `Klaw`，可以在代码中修改：

```go
commandRouter.SetMentionName("Bot")  // 修改为 Bot
```

### 命令缩写

支持常用缩写：
- `c` = `cluster`
- `p` = `pod`
- `n` = `node`
- `d` = `deployment`
- `m` = `monitor`
- `ls` = `list`
- `desc` = `describe`

**示例：**
```
@Klaw klaw p ls kind-my-k8s default  # 等同于 pod list
```

## 故障排查

### 钉钉收不到回复

1. 检查 Klaw 是否启动成功：
   ```bash
   curl http://localhost:8081/webhook/dingtalk
   # 应该返回 405 Method Not Allowed
   ```

2. 检查钉钉回调配置是否正确

3. 查看 Klaw 日志：
   ```bash
   tail -f /tmp/klaw.log
   ```

### 签名验证失败

1. 确认 `secret` 配置正确
2. 检查服务器时间是否同步（NTP）

### 命令执行超时

钉钉要求 200ms 内返回响应，复杂命令可能超时。Klaw 会：
1. 立即返回 "收到消息" 的确认
2. 异步执行命令
3. 完成后发送结果到钉钉

## 安全建议

1. **启用签名验证**：在钉钉机器人设置中启用加签
2. **配置 IP 白名单**：限制只有 Klaw 服务器 IP 可以访问
3. **使用 HTTPS**：生产环境建议使用 HTTPS
4. **权限控制**：后续版本将支持频道级 RBAC

## 更新日志

### v0.2.0
- ✅ 钉钉双向通信
- ✅ 命令路由系统
- ✅ Markdown 消息格式

### 计划功能
- 🔲 频道级权限控制
- 🔲 交互式按钮
- 🔲 审批流程（危险操作）
- 🔲 命令历史记录
