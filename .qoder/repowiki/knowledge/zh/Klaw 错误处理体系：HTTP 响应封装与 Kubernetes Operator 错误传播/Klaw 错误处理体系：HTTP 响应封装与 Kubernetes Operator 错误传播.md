---
kind: error_handling
name: Klaw 错误处理体系：HTTP 响应封装与 Kubernetes Operator 错误传播
category: error_handling
scope:
    - '**'
source_files:
    - internal/api/server.go
    - internal/api/unified_v1.go
    - operator/controllers/clusterdiagnostic_controller.go
    - internal/automation/builtins.go
    - internal/automation/manager.go
    - internal/alerting/manager.go
---

## 1. 使用的系统与模式

- **Go 标准库 error**：项目未引入第三方错误库（如 `pkg/errors`、`errors`），统一使用 Go 原生 `fmt.Errorf` 和 `%w` 包装错误，通过 `error` 接口在函数间传递。
- **HTTP 层统一响应封装**：`internal/api/server.go` 提供 `respondJSON` 和 `respondError` 两个辅助方法，所有 HTTP handler 均通过这两个方法返回 JSON 响应，错误体格式固定为 `{"error": "message"}`。
- **路由级自定义错误类型**：`internal/api/unified_v1.go` 定义了 `routeError` 结构体（实现 `error` 接口），用于区分参数校验、资源不支持等路由层面的错误。
- **Kubernetes Operator 错误传播**：operator/controllers 使用 controller-runtime 的 `ctrl.Result{}, err` 返回模式，配合 `k8s.io/apimachinery/pkg/api/errors.IsNotFound` 判断资源不存在。

## 2. 关键文件与包

- `internal/api/server.go`：HTTP Server 核心，定义 `respondJSON`/`respondError`、CORS 中间件、deprecation 中间件。
- `internal/api/unified_v1.go`：统一 API v1 路由，定义 `routeError`、`errUnsupportedResource`、`errNamespaceRequired`、`errClusterScopedResource` 等错误构造器。
- `operator/controllers/*.go`：三个控制器（ClusterDiagnostic、NodeDiagnostic、Schedule）使用 controller-runtime 的错误处理模式。
- `internal/automation/builtins.go`、`internal/automation/manager.go`、`internal/alerting/manager.go`、`internal/audit/security.go`：各业务模块使用 `fmt.Errorf` 返回具体错误信息。

## 3. 架构与约定

### HTTP 错误处理流程
```
Handler → 调用业务逻辑 → 返回 error → respondError(w, err.Error(), statusCode)
```
- 每个 handler 独立处理错误，没有全局 panic/recover 或统一的错误中间件。
- 错误状态码由 handler 根据语义选择：400（请求体无效）、404（资源不存在）、500（内部错误）。
- 非 API 路径（SPA 路由）对 `/api` 前缀的请求直接返回 404。

### 路由级错误分类
- `errUnsupportedResource(kind)`：不支持的资源类型，返回 400。
- `errNamespaceRequired(kind)`：需要命名空间但未提供，返回 400。
- `errClusterScopedResource(kind)`：集群作用域资源不应带 namespace，返回 400。

### Operator 错误处理
- Reconcile 函数返回 `(ctrl.Result, error)`，错误会触发 requeue。
- 使用 `errors.IsNotFound(err)` 区分资源不存在与其他错误。
- 通过 `log.Error(err, message)` 记录错误上下文，不吞掉错误。

## 4. 约定与约束

- **无 panic/recover**：全仓库未发现 `panic()` 或 `recover()` 调用，错误全部通过返回值传递。
- **错误包装统一使用 `%w`**：所有跨层错误都使用 `fmt.Errorf("...: %w", err)` 包装，便于错误链追踪。
- **HTTP 响应格式固定**：成功响应由 `respondJSON` 序列化任意结构体；错误响应固定为 `{"error": string}`。
- **无统一错误码枚举**：错误信息以字符串形式直接嵌入响应体，没有标准化的错误码字段。
- **中间件仅做横切关注点**：当前只有 CORS 和 deprecation 头设置，不包含认证、鉴权或统一错误处理。
- **测试覆盖**：`web/src/__tests__/integration/error-handling.test.tsx` 包含前端错误处理测试，后端各模块有对应单元测试验证错误路径。