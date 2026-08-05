---
kind: build_system
name: Klaw 构建与发布系统
category: build_system
scope:
    - '**'
source_files:
    - Makefile
    - Dockerfile
    - .github/workflows/ci.yml
    - go.mod
    - helm/klaw/Chart.yaml
    - operator/go.mod
    - modules/etcd-backup/go.mod
    - web/package.json
---

## 构建系统与交付流水线概述

Klaw 项目采用多模块、多语言单体仓库架构，通过统一的 Makefile + Dockerfile + GitHub Actions CI 实现后端 Go 服务、前端 React 应用、Kubernetes Operator 和 Helm Chart 的一体化构建与发布。

### 核心构建工具链
- **Go 后端**: 使用 `go build` 直接编译 `./cmd/klaw`，Go 版本锁定为 1.24.2
- **React 前端**: 基于 Vite + TypeScript 构建，Node.js 版本 20，使用 npm ci 进行依赖管理
- **Docker 镜像**: 多阶段构建（node:20-alpine → golang:1.24-alpine → alpine:3.20），最终镜像仅 68 行
- **Helm Chart**: 标准 Helm v2 格式，Chart.yaml 中 version 和 appVersion 均为 1.0.0
- **CI/CD**: GitHub Actions 在 push/PR 到 main/master 分支时触发，包含 go vet、build、test 三个并行 job

### 构建目标与流程
Makefile 提供完整的开发工作流：
- `make build`: 并行构建前端 (`build-frontend`) 和后端 (`build-backend`)
- `make dev`: 启动开发环境，同时运行前端 dev server 和后端服务
- `make docker-build`: 构建 Docker 镜像，默认标签 `kudig-io/klaw:latest`
- `make helm-install/upgrage/uninstall`: 一键部署到 Kubernetes 集群
- `make test/test-go/test-frontend`: 分别运行 Go 测试和前端 Vitest 测试
- `make fmt/lint`: 代码格式化 (go fmt + eslint) 和静态检查 (golangci-lint)

### 多模块依赖管理
项目包含三个独立的 Go 模块：
- **主模块** (`github.com/kudig-io/klaw`): Go 1.24.2，依赖 k8s.io/api v0.28.0、modernc.org/sqlite
- **Operator 模块** (`github.com/kudig-io/klaw/operator`): Go 1.21，基于 controller-runtime v0.16.3
- **etcd-backup 模块** (`github.com/kudig-io/klaw/modules/etcd-backup`): Go 1.25，独立 HTTP 客户端

### 容器化与安全实践
- 非 root 用户运行 (UID 65532, group klaw)
- CGO_ENABLED=0 编译确保二进制可移植性
- 仅暴露 8080 端口，挂载 /app/data 作为持久化存储
- 配置文件通过 `-v $(PWD)/configs/config.yaml:/app/configs/config.yaml` 注入

### 版本管理与发布策略
当前版本号为 1.0.0，通过 Helm Chart 的 version/appVersion 字段统一管理。Docker 镜像标签使用 `klaw:latest`，未实现语义化版本控制或自动化发布流程。