---
kind: logging_system
name: Klaw 日志系统 — 分散的标准库输出与审计专用结构化日志
category: logging_system
scope:
    - '**'
source_files:
    - internal/audit/logger.go
    - internal/api/server.go
    - operator/cmd/main.go
    - cmd/klaw/main.go
    - internal/events/kubernetes.go
---

## 1. 使用的系统与框架

Klaw 项目未采用统一的第三方日志框架（如 zap、logrus、slog），而是以 Go 标准库 log 和 fmt.Printf/Println 为主，辅以独立的审计子系统提供结构化日志能力。

- 运行时调试输出：广泛使用 fmt.Printf / fmt.Println 直接输出到 stdout，用于服务启动、事件监听、消息发送等关键路径的调试信息。
- HTTP 服务器日志：在 internal/api/server.go 中使用标准库 log.Printf 输出服务启动信息。
- Operator 组件：Kubernetes Operator (operator/cmd/main.go) 使用 zap 作为其日志框架，通过 ctrl.SetLogger(zap.New(...)) 初始化。
- 审计日志：internal/audit/logger.go 实现了专用的结构化审计日志系统，将审计事件持久化到存储后端。

## 2. 核心文件与包

- cmd/klaw/main.go - CLI 入口，无日志配置
- internal/api/server.go - HTTP 服务器，使用 log.Printf
- internal/audit/logger.go - 审计日志结构体定义与持久化
- operator/cmd/main.go - Operator 的 zap 日志初始化
- internal/events/kubernetes.go - Kubernetes 事件监听，大量 fmt.Printf 调试输出
- internal/messaging/dingtalk/client.go - 钉钉客户端，fmt.Println 启动日志
- internal/monitoring/manager.go - 监控管理器，fmt.Printf 错误输出

## 3. 架构与设计决策

### 分层日志策略
- 调试层：fmt.Printf/Println 用于开发调试，无级别控制，无结构化字段
- 应用层：标准库 log 用于简单的进程级日志
- 审计层：独立的 audit.Logger 提供完整的结构化审计日志，包含事件类型、分类、严重级别、用户、资源、结果等丰富字段
- Operator 层：独立使用 zap 框架，与其他组件解耦

### 审计日志结构
AuditEvent 结构体定义了完整的审计字段：
- 基础信息：ID、时间戳、事件类型、分类、严重级别
- 上下文：来源、用户、动作、资源映射
- 结果：操作结果、详细信息、IP 地址、用户代理
- 统计：支持按事件类型、严重级别、分类、用户进行统计

### 存储集成
审计日志通过 storage.Store 接口持久化，支持 JSON 格式存储，具备内存缓存（最多 10000 条）和并发安全（RWMutex 保护）。

## 4. 约定与约束

### 观察到的模式
- 调试输出：所有模块都使用 fmt.Printf 进行调试输出，没有统一的日志级别管理
- 错误处理：错误信息直接通过 fmt.Printf 输出，没有统一的错误日志格式
- 审计规范：审计事件必须包含完整字段，默认值自动填充（source="klaw"、user="system"、result="success"）
- Operator 隔离：Operator 组件完全独立使用 zap，不依赖主应用的日志系统

### 缺失的标准化
- 没有统一的日志级别（DEBUG/INFO/WARN/ERROR）
- 没有结构化日志字段约定（除了审计日志）
- 没有日志输出目标配置（stdout/stderr/file）
- 没有日志轮转或大小限制机制
- 缺少请求链路追踪 ID
- 没有日志采样或过滤配置

### 潜在问题
- 调试输出与生产日志混用，难以区分
- 无法动态调整日志级别
- 审计日志与应用日志分离，缺乏关联
- 不同组件使用不同的日志方式，不利于统一收集和分析