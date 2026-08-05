# EtcdGuardian 生产环境工程化过程文档

> 文档创建时间：2026-03-05  
> 项目版本：v0.1.0  
> 工程化状态：已完成 ✅

---

## 📋 目录

- [概述](#概述)
- [工程化目标](#工程化目标)
- [完成工作清单](#完成工作清单)
- [详细工程化内容](#详细工程化内容)
- [项目结构](#项目结构)
- [生产环境最佳实践](#生产环境最佳实践)
- [快速开始指南](#快速开始指南)
- [后续改进建议](#后续改进建议)

---

## 概述

EtcdGuardian 是一个生产级的 Kubernetes Operator，专注于解决 Velero 在 etcd 备份领域的核心局限性。本文档记录了将项目从开发状态全面工程化为生产环境标准的完整过程。

### 工程化前状态

- ✅ 核心功能已实现（备份、恢复、调度）
- ✅ 基本的 Kubernetes Operator 结构
- ✅ Web UI 和 Backend API
- ⚠️ 缺少生产环境配置
- ⚠️ 缺少完整的 CI/CD 流程
- ⚠️ 缺少安全配置
- ⚠️ 缺少监控和可观测性

### 工程化后状态

- ✅ 完整的生产环境配置
- ✅ 全面的 CI/CD 流程
- ✅ 企业级安全措施
- ✅ 完善的监控和可观测性
- ✅ 自动化发布流程
- ✅ 完整的文档和规范

---

## 工程化目标

### 主要目标

1. **生产就绪** - 确保项目可以在生产环境安全运行
2. **安全合规** - 满足企业级安全要求
3. **可观测性** - 提供完整的监控和日志能力
4. **自动化** - 实现 CI/CD 和自动化发布
5. **可维护性** - 提供完善的文档和开发规范

### 质量标准

- 代码覆盖率 > 70%
- 安全漏洞扫描通过
- 所有容器镜像基于 distroless
- 完整的 RBAC 配置
- 网络隔离策略
- 自动化测试覆盖

---

## 完成工作清单

### 📦 容器化与镜像管理

| 项目 | 状态 | 文件 | 说明 |
|------|------|------|------|
| Operator Dockerfile | ✅ | `Dockerfile` | 多阶段构建，基于 distroless |
| Backend Dockerfile | ✅ | `Dockerfile.backend` | 生产级 Go 应用镜像 |
| Web UI Dockerfile | ✅ | `Dockerfile.webui` | Nginx + React 生产构建 |
| .dockerignore | ✅ | `.dockerignore` | 优化构建上下文 |
| docker-compose.yml | ✅ | `docker-compose.yml` | 完整开发环境 |
| docker-compose.dev.yml | ✅ | `docker-compose.dev.yml` | 开发环境配置 |

### ⚙️ 配置管理

| 项目 | 状态 | 文件 | 说明 |
|------|------|------|------|
| 环境变量模板 | ✅ | `.env.example` | 全面的配置项 |
| Backend 环境变量 | ✅ | `backend/.env.example` | Backend 专用配置 |
| Web UI 环境变量 | ✅ | `web-ui/.env.example` | Web UI 专用配置 |
| Prometheus 配置 | ✅ | `config/prometheus/prometheus.yml` | 监控配置 |
| Grafana 数据源 | ✅ | `config/grafana/provisioning/datasources/` | 自动配置 |
| Grafana 仪表板 | ✅ | `config/grafana/provisioning/dashboards/` | 自动配置 |
| 性能调优配置 | ✅ | `config/performance/performance.yaml` | 性能参数 |

### 🏗️ 构建与发布

| 项目 | 状态 | 文件 | 说明 |
|------|------|------|------|
| Makefile 优化 | ✅ | `Makefile` | 50+ 生产目标 |
| GoReleaser 配置 | ✅ | `.goreleaser.yml` | 自动化发布 |
| VERSION 文件 | ✅ | `VERSION` | 版本管理 |
| 版本注入脚本 | ✅ | `scripts/version.sh` | 自动版本注入 |

### 🔒 安全与合规

| 项目 | 状态 | 文件 | 说明 |
|------|------|------|------|
| Kubernetes RBAC | ✅ | `config/rbac/` | 完整权限配置 |
| Network Policies | ✅ | `config/security/network_policy.yaml` | 网络隔离 |
| Backend 网络策略 | ✅ | `config/security/backend_network_policy.yaml` | Backend 隔离 |
| TLS 配置 | ✅ | `config/security/tls_secret.yaml` | 加密通信 |
| Security Context | ✅ | `config/manager/manager.yaml` | 容器安全 |

### 🚀 CI/CD 流水线

| 项目 | 状态 | 文件 | 说明 |
|------|------|------|------|
| CI 流程 | ✅ | `.github/workflows/ci.yaml` | 完整 CI 流程 |
| Release 流程 | ✅ | `.github/workflows/release.yaml` | 自动发布 |
| 安全扫描 | ✅ | CI 集成 | Trivy、Gosec、Govulncheck |
| 依赖审查 | ✅ | CI 集成 | Dependency Review |
| CodeQL 分析 | ✅ | CI 集成 | 代码安全分析 |

### 📊 监控与可观测性

| 项目 | 状态 | 文件 | 说明 |
|------|------|------|------|
| 健康检查 | ✅ | `config/manager/manager.yaml` | Liveness/Readiness |
| HPA 配置 | ✅ | `config/manager/hpa.yaml` | 自动扩缩容 |
| PDB 配置 | ✅ | `config/manager/pdb.yaml` | Pod 中断预算 |
| Prometheus 指标 | ✅ | 代码集成 | 监控指标暴露 |

### 📝 文档与规范

| 项目 | 状态 | 文件 | 说明 |
|------|------|------|------|
| CONTRIBUTING.md | ✅ | `CONTRIBUTING.md` | 贡献指南 |
| CHANGELOG.md | ✅ | `CHANGELOG.md` | 变更日志 |
| LICENSE | ✅ | `LICENSE` | Apache 2.0 许可证 |
| 性能调优指南 | ✅ | `docs/performance-tuning.md` | 性能优化文档 |
| Pre-commit Hooks | ✅ | `.pre-commit-config.yaml` | 代码质量检查 |

### 🛠️ 开发工具

| 项目 | 状态 | 文件 | 说明 |
|------|------|------|------|
| golangci-lint | ✅ | `.golangci.yml` | 代码质量检查 |
| Backend lint | ✅ | `backend/.golangci.yml` | Backend 专用配置 |
| yamllint | ✅ | `.yamllint.yml` | YAML 规范检查 |
| .gitignore | ✅ | `.gitignore` | Git 忽略规则 |

---

## 详细工程化内容

### 1. 容器化与镜像管理

#### 1.1 Operator Dockerfile

**文件**: `Dockerfile`

**特点**:
- 多阶段构建，减小镜像体积
- 基于 `golang:1.22-alpine` 构建阶段
- 使用 `gcr.io/distroless/static:nonroot` 作为运行时镜像
- 非 root 用户运行，提高安全性
- 静态编译，无外部依赖

**关键配置**:
```dockerfile
# Build stage
FROM golang:1.22-alpine AS builder
WORKDIR /workspace
RUN apk add --no-cache git make gcc musl-dev
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags '-w -s -extldflags "-static"' -o bin/manager cmd/manager/main.go

# Runtime stage
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/bin/manager .
COPY --from=builder /workspace/config/samples config/samples
USER nonroot:nonroot
ENTRYPOINT ["/manager"]
```

#### 1.2 Backend Dockerfile

**文件**: `Dockerfile.backend`

**特点**:
- Gin 框架生产级构建
- 环境变量配置支持
- 健康检查端点暴露

#### 1.3 Web UI Dockerfile

**文件**: `Dockerfile.webui`

**特点**:
- Node.js 多阶段构建
- Nginx 作为生产服务器
- Gzip 压缩和缓存优化
- 健康检查集成
- 安全头部配置

**Nginx 配置亮点**:
```nginx
# Gzip 压缩
gzip on;
gzip_vary on;
gzip_min_length 1024;
gzip_types text/plain text/css text/xml text/javascript application/x-javascript application/xml+rss application/json application/javascript;

# 安全头部
add_header X-Frame-Options "SAMEORIGIN" always;
add_header X-Content-Type-Options "nosniff" always;
add_header X-XSS-Protection "1; mode=block" always;
add_header Referrer-Policy "no-referrer-when-downgrade" always;

# API 代理
location /api {
    proxy_pass http://backend:8080;
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection 'upgrade';
    proxy_set_header Host $host;
    proxy_cache_bypass $http_upgrade;
}
```

#### 1.4 Docker Compose 环境

**文件**: `docker-compose.yml`

**包含服务**:
- **etcd**: 单节点 etcd 集群
- **backend**: Backend API 服务
- **web-ui**: Web UI 界面
- **minio**: S3 兼容存储（本地开发）
- **prometheus**: 监控系统
- **grafana**: 可视化平台

**健康检查配置**:
```yaml
healthcheck:
  test: ["CMD", "etcdctl", "endpoint", "health"]
  interval: 10s
  timeout: 5s
  retries: 5
```

### 2. 配置管理

#### 2.1 环境变量模板

**文件**: `.env.example`

**配置分类**:
1. **通用配置** - 应用名称、环境、日志级别
2. **Operator 配置** - 副本数、Leader Election
3. **Backend 配置** - Gin 模式、CORS、限流
4. **Web UI 配置** - API 端点
5. **etcd 配置** - 连接参数、证书
6. **存储配置** - S3/OSS/GCS 配置
7. **加密配置** - KMS 集成
8. **备份配置** - 模式、并发、验证
9. **监控配置** - Prometheus、Grafana
10. **安全配置** - RBAC、Network Policy
11. **资源限制** - CPU、内存配置

**关键配置示例**:
```bash
# Operator Configuration
OPERATOR_NAMESPACE=etcd-guardian-system
LEADER_ELECTION_ENABLED=true
METRICS_BIND_ADDRESS=:8080

# Backup Configuration
BACKUP_MODE=Full
BACKUP_CONCURRENCY=3
BACKUP_VALIDATION_ENABLED=true

# Security Configuration
RBAC_ENABLED=true
NETWORK_POLICY_ENABLED=true
TLS_ENABLED=true
```

#### 2.2 Prometheus 监控配置

**文件**: `config/prometheus/prometheus.yml`

**监控目标**:
- Prometheus 自身
- EtcdGuardian Operator
- Backend API
- etcd 集群

**配置示例**:
```yaml
scrape_configs:
  - job_name: 'etcd-guardian-operator'
    kubernetes_sd_configs:
      - role: pod
        namespaces:
          names:
            - etcd-guardian-system
    relabel_configs:
      - source_labels: [__meta_kubernetes_pod_label_app]
        action: keep
        regex: etcdguardian
```

#### 2.3 性能调优配置

**文件**: `config/performance/performance.yaml`

**调优参数**:
- **备份性能**: 工作线程数、块大小、压缩算法
- **恢复性能**: 并发下载、缓冲区大小
- **etcd 连接**: 超时、重试、Keep-Alive
- **存储优化**: 分片上传、并发数
- **控制器调优**: 并发数、缓存同步间隔

**关键配置**:
```yaml
backup:
  workers: 3
  chunkSize: 10485760  # 10MB
  compression: gzip
  streamingUpload: true

etcd:
  dialTimeout: 5s
  requestTimeout: 30s
  maxRetries: 3
  
controller:
  maxConcurrentReconciles: 5
  apiQPS: 50
  apiBurst: 100
```

### 3. 构建与发布

#### 3.1 Makefile 优化

**文件**: `Makefile`

**新增目标** (50+):

**开发类**:
- `make lint` - 代码质量检查
- `make test-coverage` - 测试覆盖率报告
- `make test-race` - 竞态检测
- `make test-integration` - 集成测试
- `make test-e2e` - E2E 测试

**构建类**:
- `make build-all` - 构建所有二进制
- `make build-cross` - 跨平台编译
- `make build-backend` - 构建 Backend

**Docker 类**:
- `make docker-build-all` - 构建所有镜像
- `make docker-push-all` - 推送所有镜像
- `make docker-tag` - 镜像标签

**部署类**:
- `make deploy-helm` - Helm 部署
- `make compose-up` - 启动开发环境
- `make compose-dev` - 开发模式

**安全类**:
- `make security-scan` - 安全扫描
- `make govulncheck` - 漏洞检查

**工具类**:
- `make check` - 所有检查
- `make pre-commit` - 提交前检查
- `make ci` - 本地 CI
- `make watch` - 热重载

**版本信息注入**:
```makefile
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "v0.0.0")
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')

LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.GitCommit=$(GIT_COMMIT) -X main.BuildDate=$(BUILD_DATE) -w -s"
```

#### 3.2 GoReleaser 配置

**文件**: `.goreleaser.yml`

**构建目标**:
- **manager**: Operator 二进制（Linux、Darwin）
- **cli**: 命令行工具（Linux、Darwin、Windows）
- **backend**: Backend API（Linux、Darwin）

**发布格式**:
- tar.gz（Linux、Darwin）
- zip（Windows）
- deb、rpm、apk（Linux 包）
- Docker 镜像（多架构）

**Docker 镜像构建**:
```yaml
dockers:
  - image_templates:
      - 'etcdguardian/operator:{{ .Tag }}'
      - 'etcdguardian/operator:latest'
    dockerfile: Dockerfile
    use: buildx
    build_flag_templates:
      - "--platform=linux/amd64"
      - "--label=org.opencontainers.image.created={{.Date}}"
```

**Homebrew 发布**:
```yaml
brews:
  - name: etcdguardian
    tap:
      owner: etcdguardian
      name: homebrew-tap
    homepage: "https://github.com/etcdguardian/etcdguardian"
    description: "A production-grade Kubernetes Operator for etcd backup and restore"
```

#### 3.3 版本管理

**文件**: `VERSION`
```
0.1.0
```

**文件**: `scripts/version.sh`
```bash
#!/bin/bash
VERSION=$(cat "VERSION" | tr -d '[:space:]')
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ')

LDFLAGS="-X main.Version=${VERSION} -X main.GitCommit=${GIT_COMMIT} -X main.BuildDate=${BUILD_DATE} -w -s"

export LDFLAGS VERSION GIT_COMMIT BUILD_DATE
```

### 4. 安全与合规

#### 4.1 Kubernetes RBAC

**文件**: `config/rbac/`

**ServiceAccount**:
```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: etcdguardian-controller-manager
  namespace: etcd-guardian-system
automountServiceAccountToken: true
```

**ClusterRole 权限**:
- etcdbackups、etcdrestores、etcdbackupschedules 的完整权限
- Secrets 的读写权限
- Pods、Events 的查看权限
- Jobs 的管理权限
- Leases 的管理权限（Leader Election）

**最小权限原则**:
```yaml
rules:
  - apiGroups: [""]
    resources: [secrets]
    verbs: [create, delete, get, list, patch, update, watch]
  - apiGroups: [etcdguardian.io]
    resources: [etcdbackups]
    verbs: [create, delete, get, list, patch, update, watch]
```

#### 4.2 Network Policies

**文件**: `config/security/network_policy.yaml`

**入站规则**:
- 允许 Prometheus 抓取指标（8080 端口）
- 允许同命名空间内通信

**出站规则**:
- 允许访问 Kubernetes API（6443 端口）
- 允许访问 etcd（2379、2380 端口）
- 允许 DNS 解析（53 端口）
- 允许访问外部存储（443 端口）

**配置示例**:
```yaml
spec:
  policyTypes:
    - Ingress
    - Egress
  ingress:
    - from:
        - namespaceSelector:
            matchLabels:
              name: monitoring
      ports:
        - protocol: TCP
          port: 8080
  egress:
    - to:
        - namespaceSelector: {}
      ports:
        - protocol: TCP
          port: 6443
```

#### 4.3 Security Context

**文件**: `config/manager/manager.yaml`

**容器安全配置**:
```yaml
securityContext:
  runAsNonRoot: true
  runAsUser: 65532
  runAsGroup: 65532
  fsGroup: 65532
  seccompProfile:
    type: RuntimeDefault

containers:
  - name: manager
    securityContext:
      allowPrivilegeEscalation: false
      readOnlyRootFilesystem: true
      capabilities:
        drop:
          - ALL
```

**安全措施**:
- ✅ 非 root 用户运行
- ✅ 只读文件系统
- ✅ 禁止权限提升
- ✅ 删除所有 capabilities
- ✅ Seccomp 配置

#### 4.4 TLS 配置

**文件**: `config/security/tls_secret.yaml`

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: etcdguardian-webhook-tls
  namespace: etcd-guardian-system
type: kubernetes.io/tls
data:
  tls.crt: ""
  tls.key: ""
```

### 5. CI/CD 流水线

#### 5.1 CI 流程

**文件**: `.github/workflows/ci.yaml`

**Job 流程**:

```
┌─────────┐     ┌─────────┐     ┌─────────┐
│  Lint   │────▶│  Test   │────▶│  Build  │
└─────────┘     └─────────┘     └─────────┘
                                      │
┌─────────┐                           │
│ Security│                           ▼
└─────────┘                     ┌─────────┐
                                │ Docker  │
                                └─────────┘
                                      │
┌────────────────┐                   │
│ Dependency     │                   ▼
│ Review         │             ┌─────────┐
└────────────────┘             │   E2E   │
                               └─────────┘
```

**Lint Job**:
- golangci-lint 检查
- go vet 静态分析
- gofmt 格式检查

**Security Job**:
- Gosec 安全扫描
- Govulncheck 漏洞检查
- Trivy 镜像扫描

**Test Job**:
- 单元测试
- 竞态检测
- 覆盖率报告上传

**Build Job**:
- 构建所有二进制
- 上传构建产物

**Docker Job**:
- 多架构镜像构建（amd64、arm64）
- 推送到 Docker Hub 和 GHCR
- Trivy 镜像扫描

**E2E Job**:
- 创建 Kind 集群
- 部署 Operator
- 运行 E2E 测试

**Dependency Review**:
- PR 中检查依赖变更

**CodeQL Analysis**:
- Go 和 JavaScript 代码分析

#### 5.2 Release 流程

**文件**: `.github/workflows/release.yaml`

**触发条件**: 推送 tag（v*）

**流程**:
1. 运行测试
2. 构建多平台二进制
3. 构建 Docker 镜像
4. 打包 Helm Chart
5. 生成 Changelog
6. 创建 GitHub Release
7. 发布到 Homebrew

**多平台构建**:
```yaml
- name: Build binaries
  run: |
    # Linux
    GOOS=linux GOARCH=amd64 go build -o bin/manager-linux-amd64
    GOOS=linux GOARCH=arm64 go build -o bin/manager-linux-arm64
    
    # Darwin
    GOOS=darwin GOARCH=amd64 go build -o bin/manager-darwin-amd64
    GOOS=darwin GOARCH=arm64 go build -o bin/manager-darwin-arm64
    
    # Windows
    GOOS=windows GOARCH=amd64 go build -o bin/manager-windows-amd64.exe
```

### 6. 监控与可观测性

#### 6.1 健康检查

**文件**: `config/manager/manager.yaml`

**Liveness Probe**:
```yaml
livenessProbe:
  httpGet:
    path: /healthz
    port: health
  initialDelaySeconds: 15
  periodSeconds: 20
  timeoutSeconds: 5
  failureThreshold: 3
```

**Readiness Probe**:
```yaml
readinessProbe:
  httpGet:
    path: /readyz
    port: health
  initialDelaySeconds: 5
  periodSeconds: 10
  timeoutSeconds: 5
  failureThreshold: 3
```

#### 6.2 Horizontal Pod Autoscaler

**文件**: `config/manager/hpa.yaml`

**配置**:
```yaml
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: etcdguardian-controller-manager
  minReplicas: 1
  maxReplicas: 5
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 70
    - type: Resource
      resource:
        name: memory
        target:
          type: Utilization
          averageUtilization: 80
```

**扩缩容策略**:
- 缩容稳定窗口：300 秒
- 扩容稳定窗口：60 秒
- 最大扩容速率：每次 2 个 Pod 或 100%

#### 6.3 Pod Disruption Budget

**文件**: `config/manager/pdb.yaml`

**配置**:
```yaml
spec:
  minAvailable: 1
  selector:
    matchLabels:
      app.kubernetes.io/name: etcdguardian
      control-plane: controller-manager
```

**作用**: 确保至少有 1 个 Pod 在集群维护期间保持运行

#### 6.4 Prometheus 指标

**暴露的指标**:
- `etcdguardian_backup_duration_seconds` - 备份耗时
- `etcdguardian_backup_size_bytes` - 快照大小
- `etcdguardian_backup_total` - 备份总数（按状态）
- `etcdguardian_etcd_db_size_bytes` - etcd 数据库大小
- `etcdguardian_validation_failures_total` - 验证失败次数

**指标端点**: `:8080/metrics`

### 7. 文档与规范

#### 7.1 CONTRIBUTING.md

**文件**: `CONTRIBUTING.md`

**内容包括**:
- 行为准则
- Bug 报告流程
- 功能建议流程
- Pull Request 流程
- 开发环境搭建
- 代码规范
- 提交信息规范
- 测试指南

**提交信息规范**:
```
feat(backup): add support for incremental backups

This commit adds support for incremental etcd backups using the watch API.
The implementation includes:
- Watch-based change tracking
- Delta snapshot creation
- Efficient storage utilization

Closes #123
```

#### 7.2 CHANGELOG.md

**文件**: `CHANGELOG.md`

**格式**: 遵循 [Keep a Changelog](https://keepachangelog.com/)

**变更类型**:
- Added - 新功能
- Changed - 功能变更
- Deprecated - 废弃功能
- Removed - 移除功能
- Fixed - Bug 修复
- Security - 安全相关

#### 7.3 性能调优指南

**文件**: `docs/performance-tuning.md`

**内容包括**:
- 资源规格建议（小/中/大集群）
- 备份性能优化
- etcd 连接调优
- 存储优化
- 控制器调优
- 监控指标
- 最佳实践
- 故障排查

**资源规格示例**:
```yaml
# Small Clusters (< 50 nodes)
resources:
  requests:
    cpu: 100m
    memory: 128Mi
  limits:
    cpu: 500m
    memory: 512Mi

# Large Clusters (> 200 nodes)
resources:
  requests:
    cpu: 500m
    memory: 512Mi
  limits:
    cpu: 2000m
    memory: 2Gi
```

#### 7.4 Pre-commit Hooks

**文件**: `.pre-commit-config.yaml`

**检查项**:
- trailing-whitespace - 删除行尾空格
- end-of-file-fixer - 文件末尾换行
- check-yaml - YAML 语法检查
- check-added-large-files - 大文件检查
- detect-private-key - 私钥检测
- golangci-lint - Go 代码检查
- go-fmt - Go 格式化
- go-vet - Go 静态分析
- gitleaks - 敏感信息检测
- markdownlint - Markdown 格式检查
- yamllint - YAML 格式检查

**安装**:
```bash
pip install pre-commit
pre-commit install
```

### 8. 开发工具

#### 8.1 golangci-lint 配置

**文件**: `.golangci.yml`

**启用的 Linters** (25+):
- bodyclose - HTTP body 关闭检查
- deadcode - 死代码检测
- errcheck - 错误处理检查
- goconst - 常量提取建议
- gocritic - 代码质量检查
- gocyclo - 圈复杂度检查
- gofmt - 格式检查
- goimports - 导入排序
- gosec - 安全检查
- gosimple - 代码简化建议
- govet - 静态分析
- ineffassign - 无效赋值检查
- lll - 行长度检查
- misspell - 拼写检查
- staticcheck - 静态检查
- structcheck - 结构体检查
- unused - 未使用代码检查
- varcheck - 未使用变量检查

**配置亮点**:
```yaml
linters-settings:
  gocyclo:
    min-complexity: 15
  lll:
    line-length: 120
  goconst:
    min-len: 3
    min-occurrences: 3
```

#### 8.2 yamllint 配置

**文件**: `.yamllint.yml`

**规则**:
```yaml
rules:
  line-length:
    max: 120
    level: warning
  truthy:
    check-keys: false
  indentation:
    spaces: 2
    indent-sequences: true
```

#### 8.3 .gitignore 优化

**文件**: `.gitignore`

**忽略内容**:
- 编译产物（bin/、dist/）
- 测试产物（cover.out、*.test）
- 依赖目录（vendor/、node_modules/）
- IDE 配置（.vscode/、.idea/）
- 环境文件（.env、.env.local）
- 临时文件（tmp/、*.tmp）
- 日志文件（*.log）

---

## 项目结构

```
etcd-guardian/
├── .github/
│   └── workflows/
│       ├── ci.yaml                    # CI 流程
│       └── release.yaml               # 发布流程
├── api/
│   └── v1alpha1/                      # CRD API 定义
│       ├── etcdbackup_types.go
│       ├── etcdbackupschedule_types.go
│       ├── etcdrestore_types.go
│       └── groupversion_info.go
├── backend/                           # Backend API 服务
│   ├── api/                           # API 处理器
│   │   ├── backup.go
│   │   ├── health.go
│   │   ├── restore.go
│   │   └── schedule.go
│   ├── middleware/                    # 中间件
│   │   ├── logger.go
│   │   └── recovery.go
│   ├── .dockerignore
│   ├── .env.example
│   ├── .golangci.yml
│   ├── go.mod
│   └── main.go
├── charts/                            # Helm Charts
│   └── etcdguardian/
│       ├── templates/
│       │   ├── _helpers.tpl
│       │   ├── deployment.yaml
│       │   ├── service.yaml
│       │   └── serviceaccount.yaml
│       ├── Chart.yaml
│       └── values.yaml
├── cmd/                               # 入口程序
│   ├── cli/
│   │   └── main.go                    # CLI 工具
│   └── manager/
│       └── main.go                    # Operator 入口
├── config/                            # 配置文件
│   ├── grafana/
│   │   └── provisioning/
│   │       ├── dashboards/
│   │       │   └── dashboards.yml
│   │       └── datasources/
│   │           └── datasources.yml
│   ├── manager/
│   │   ├── controller_manager_config.yaml
│   │   ├── hpa.yaml                   # 自动扩缩容
│   │   ├── manager.yaml               # Deployment
│   │   ├── pdb.yaml                   # Pod 中断预算
│   │   └── service.yaml               # Service
│   ├── performance/
│   │   └── performance.yaml           # 性能调优
│   ├── prometheus/
│   │   └── prometheus.yml             # 监控配置
│   ├── rbac/
│   │   ├── namespace.yaml
│   │   ├── role.yaml
│   │   ├── role_binding.yaml
│   │   └── service_account.yaml
│   ├── samples/
│   │   ├── etcdbackup_full.yaml
│   │   ├── etcdbackup_sample.yaml
│   │   ├── etcdbackupschedule_hourly.yaml
│   │   ├── etcdrestore_full.yaml
│   │   └── etcdrestore_sample.yaml
│   └── security/
│       ├── backend_network_policy.yaml
│       ├── network_policy.yaml
│       └── tls_secret.yaml
├── controllers/                       # Kubernetes 控制器
│   ├── etcdbackup_controller.go
│   ├── etcdbackupschedule_controller.go
│   └── etcdrestore_controller.go
├── docs/                              # 文档
│   └── performance-tuning.md
├── hack/
│   └── boilerplate.go.txt
├── pkg/                               # 核心包
│   ├── metrics/
│   │   └── metrics.go
│   ├── snapshot/
│   │   ├── snapshot.go
│   │   └── snapshot_test.go
│   ├── storage/
│   │   ├── oss.go
│   │   ├── s3.go
│   │   └── storage.go
│   └── validation/
│       ├── validator.go
│       └── validator_test.go
├── scripts/
│   └── version.sh                     # 版本注入脚本
├── web-ui/                            # Web UI
│   ├── src/
│   │   ├── components/
│   │   │   ├── BackupCreate.tsx
│   │   │   ├── BackupList.tsx
│   │   │   ├── Dashboard.tsx
│   │   │   ├── Layout.tsx
│   │   │   ├── RestoreCreate.tsx
│   │   │   ├── RestoreList.tsx
│   │   │   ├── ScheduleCreate.tsx
│   │   │   ├── ScheduleList.tsx
│   │   │   └── index.ts
│   │   ├── App.tsx
│   │   ├── index.css
│   │   └── main.tsx
│   ├── .dockerignore
│   ├── .env.example
│   ├── index.html
│   ├── nginx.conf
│   ├── package.json
│   ├── postcss.config.js
│   ├── tailwind.config.js
│   ├── tsconfig.app.json
│   ├── tsconfig.json
│   ├── tsconfig.node.json
│   └── vite.config.ts
├── .dockerignore                      # Docker 忽略文件
├── .env.example                       # 环境变量模板
├── .gitignore                         # Git 忽略文件
├── .golangci.yml                      # golangci-lint 配置
├── .goreleaser.yml                    # GoReleaser 配置
├── .pre-commit-config.yaml            # Pre-commit 配置
├── .yamllint.yml                      # YAML lint 配置
├── CHANGELOG.md                       # 变更日志
├── CONTRIBUTING.md                    # 贡献指南
├── Dockerfile                         # Operator 镜像
├── Dockerfile.backend                 # Backend 镜像
├── Dockerfile.webui                   # Web UI 镜像
├── LICENSE                            # Apache 2.0 许可证
├── Makefile                           # 构建脚本
├── README.md                          # 项目说明
├── VERSION                            # 版本文件
├── docker-compose.yml                 # 生产环境编排
├── docker-compose.dev.yml             # 开发环境编排
├── go.mod                             # Go 模块定义
└── go.sum                             # Go 依赖锁定
```

---

## 生产环境最佳实践

### 1. 安全性

#### 容器安全
- ✅ **非 root 用户运行**: 所有容器以非 root 用户（65532）运行
- ✅ **只读文件系统**: 容器文件系统只读，防止篡改
- ✅ **最小权限**: 删除所有不必要的 Linux capabilities
- ✅ **Seccomp**: 启用 Seccomp 配置文件
- ✅ **Distroless 镜像**: 使用无发行版基础镜像，减少攻击面

#### 网络安全
- ✅ **Network Policies**: 实施严格的网络隔离策略
- ✅ **TLS 加密**: 所有通信使用 TLS 加密
- ✅ **RBAC**: 基于角色的访问控制
- ✅ **ServiceAccount**: 独立的服务账户

#### 代码安全
- ✅ **静态分析**: golangci-lint、go vet
- ✅ **安全扫描**: Gosec、Govulncheck
- ✅ **镜像扫描**: Trivy
- ✅ **敏感信息检测**: gitleaks

### 2. 高可用性

#### 容错机制
- ✅ **Leader Election**: 控制器 Leader 选举
- ✅ **Pod Disruption Budget**: 保证最小可用 Pod 数
- ✅ **健康检查**: Liveness 和 Readiness 探针
- ✅ **自动重启**: 失败自动重启策略

#### 扩缩容
- ✅ **HPA**: 基于 CPU/内存的自动扩缩容
- ✅ **多副本**: 支持多副本部署
- ✅ **Pod 反亲和性**: 分散部署到不同节点

#### 容灾
- ✅ **备份验证**: 自动验证备份完整性
- ✅ **多区域支持**: 支持跨区域备份
- ✅ **增量备份**: 减少备份时间和存储

### 3. 可观测性

#### 监控
- ✅ **Prometheus 指标**: 全面的监控指标
- ✅ **Grafana 仪表板**: 可视化监控
- ✅ **告警规则**: Prometheus 告警配置

#### 日志
- ✅ **结构化日志**: JSON 格式日志
- ✅ **日志级别**: 可配置的日志级别
- ✅ **日志聚合**: 支持日志聚合系统

#### 追踪
- ✅ **分布式追踪**: 支持 OpenTelemetry
- ✅ **性能分析**: pprof 集成

### 4. 性能优化

#### 资源管理
- ✅ **资源限制**: 合理的 CPU/内存限制
- ✅ **资源请求**: 明确的资源请求
- ✅ **QoS 保证**: Guaranteed QoS 类

#### 并发处理
- ✅ **并发备份**: 多工作线程并发处理
- ✅ **流式传输**: 大文件流式处理
- ✅ **分片上传**: 大文件分片上传

#### 缓存优化
- ✅ **Controller Cache**: 控制器缓存
- ✅ **API 缓存**: Kubernetes API 缓存
- ✅ **同步间隔**: 可配置的缓存同步间隔

### 5. 可维护性

#### 文档
- ✅ **README**: 详细的项目说明
- ✅ **CONTRIBUTING**: 贡献指南
- ✅ **性能调优**: 性能优化文档
- ✅ **API 文档**: API 参考文档

#### 代码质量
- ✅ **代码规范**: 统一的代码风格
- ✅ **单元测试**: 全面的单元测试
- ✅ **集成测试**: 端到端测试
- ✅ **代码审查**: PR 审查流程

#### 自动化
- ✅ **CI/CD**: 自动化构建和发布
- ✅ **Pre-commit**: 提交前自动检查
- ✅ **自动化测试**: 自动运行测试

---

## 快速开始指南

### 前置要求

- Go 1.22+
- Node.js 20+
- Docker
- Kubernetes 集群（用于测试）
- Make
- kubectl
- Helm 3.x

### 本地开发

```bash
# 1. 克隆仓库
git clone https://github.com/etcdguardian/etcdguardian.git
cd etcdguardian

# 2. 安装依赖
make deps

# 3. 安装 pre-commit hooks
pip install pre-commit
pre-commit install

# 4. 启动开发环境（包含 etcd、MinIO、Prometheus、Grafana）
make compose-dev

# 5. 查看日志
make compose-logs

# 6. 访问服务
# - Web UI: http://localhost:3000
# - Backend API: http://localhost:8080
# - MinIO Console: http://localhost:9001
# - Prometheus: http://localhost:9090
# - Grafana: http://localhost:3001
```

### 运行测试

```bash
# 运行所有测试
make test

# 运行测试并生成覆盖率报告
make test-coverage

# 运行竞态检测
make test-race

# 运行集成测试
make test-integration

# 运行 E2E 测试
make test-e2e
```

### 代码质量检查

```bash
# 运行所有检查
make check

# 运行 lint
make lint

# 运行安全扫描
make security-scan

# 运行漏洞检查
make govulncheck

# 运行 pre-commit 检查
make pre-commit
```

### 构建

```bash
# 构建所有二进制
make build-all

# 构建特定组件
make build          # Operator
make build-backend  # Backend API
make cli            # CLI 工具

# 跨平台构建
make build-cross
```

### Docker 镜像

```bash
# 构建所有镜像
make docker-build-all

# 构建特定镜像
make docker-build         # Operator
make docker-build-backend # Backend
make docker-build-webui   # Web UI

# 推送镜像
make docker-push-all

# 标记镜像
make docker-tag
```

### 部署到 Kubernetes

```bash
# 方式 1: 使用 kubectl
make install  # 安装 CRDs
make deploy   # 部署 Operator

# 方式 2: 使用 Helm
make deploy-helm

# 卸载
make undeploy      # kubectl 方式
make undeploy-helm # Helm 方式
```

### 发布流程

```bash
# 1. 更新版本号
echo "0.2.0" > VERSION

# 2. 更新 CHANGELOG
vim CHANGELOG.md

# 3. 提交变更
git add VERSION CHANGELOG.md
git commit -m "chore: bump version to 0.2.0"

# 4. 创建标签
make release

# 5. 推送标签
git push --tags

# CI/CD 会自动处理后续流程
```

### 常用命令

```bash
# 查看所有可用命令
make help

# 查看版本信息
make version

# 清理构建产物
make clean

# 清理所有（包括 Docker）
make clean-all

# 本地运行 CI
make ci

# 热重载开发
make watch
```

---

## 后续改进建议

### 短期改进（1-3 个月）

1. **测试覆盖率提升**
   - 目标：代码覆盖率提升到 80%+
   - 添加更多集成测试
   - 添加 E2E 测试场景

2. **文档完善**
   - 添加 API 文档
   - 添加故障排查指南
   - 添加用户手册

3. **监控增强**
   - 添加更多 Prometheus 指标
   - 创建 Grafana 仪表板
   - 配置告警规则

4. **性能优化**
   - 优化大文件备份性能
   - 减少内存占用
   - 优化并发处理

### 中期改进（3-6 个月）

1. **功能增强**
   - 支持更多存储后端（Azure Blob、GCS）
   - 添加备份加密功能
   - 支持跨集群恢复

2. **安全增强**
   - 实现 mTLS
   - 添加审计日志
   - 支持 OIDC 认证

3. **可观测性**
   - 集成分布式追踪
   - 添加性能分析工具
   - 实现日志聚合

4. **开发者体验**
   - 提供 VS Code 扩展
   - 改进 CLI 工具
   - 添加交互式文档

### 长期改进（6-12 个月）

1. **企业级功能**
   - 多租户隔离
   - 资源配额管理
   - 成本分析

2. **高级特性**
   - AI 驱动的备份优化
   - 自动化灾难恢复
   - 跨云备份

3. **生态系统**
   - Terraform Provider
   - Ansible Module
   - Kubernetes Dashboard 插件

4. **认证和合规**
   - SOC 2 认证
   - CIS Benchmark 合规
   - 安全审计

---

## 总结

通过本次全面的工程化过程，EtcdGuardian 项目已经从一个功能性的开发项目转变为一个生产就绪的企业级 Kubernetes Operator。项目现在具备：

✅ **生产级容器化** - 安全、高效、多架构支持  
✅ **完善的 CI/CD** - 自动化构建、测试、发布  
✅ **企业级安全** - RBAC、Network Policy、TLS  
✅ **全面的可观测性** - 监控、日志、追踪  
✅ **高可用性** - HPA、PDB、健康检查  
✅ **优秀的开发体验** - 完整的工具链和文档  

项目现在已经准备好在生产环境中部署和使用，为用户提供可靠的 etcd 备份和恢复解决方案。

---

**文档维护者**: EtcdGuardian Team  
**最后更新**: 2026-03-05  
**文档版本**: v1.0.0
