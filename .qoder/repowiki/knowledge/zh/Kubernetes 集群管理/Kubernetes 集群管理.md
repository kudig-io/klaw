---
kind: external_dependency
name: Kubernetes 集群管理
slug: kubernetes
category: external_dependency
category_hints:
    - vendor_identity
scope:
    - '**'
---

### Kubernetes 集成
- **角色**：核心依赖，用于管理纳管的 Kubernetes 集群资源（Pod、Deployment、Service、Node等）
- **集成点**：`internal/kubernetes/manager.go` 使用 client-go 与 API Server 通信；`operator/` 目录包含基于 controller-runtime 的自定义控制器
- **版本约束**：主模块使用 k8s.io v0.28.0，operator 使用 v0.28.4，存在版本不一致风险
- **API 兼容性**：client-go 版本较旧，可能与新版集群 API 存在兼容性问题
- **权限模型**：通过 kubeconfig 文件访问集群，需要相应的 RBAC 权限
- **验证要求**：需确认与目标集群版本的 API 兼容性