# Kudig Operator

Kubernetes Operator for Kudig - 自动化集群诊断工具。

## 功能特性

- **ClusterDiagnostic**: 集群级诊断任务
- **NodeDiagnostic**: 节点级诊断任务（支持节点选择器）
- **DiagnosticSchedule**: 定时诊断调度（支持 cron 表达式）
- **自动通知**: 集成 Webhook 告警

## 安装

### 使用 Helm

```bash
# 添加仓库（待发布）
helm repo add kudig https://charts.kudig.io
helm install kudig-operator kudig/kudig-operator -n kudig-system --create-namespace

# 本地安装
cd v2-go/operator/helm
helm install kudig-operator ./kudig-operator -n kudig-system --create-namespace
```

### 使用 kubectl

```bash
# 创建命名空间
kubectl create namespace kudig-system

# 安装 CRD
kubectl apply -f https://raw.githubusercontent.com/kudig-io/kudig/main/v2-go/operator/helm/kudig-operator/templates/crds.yaml

# 安装 Operator
kubectl apply -f https://raw.githubusercontent.com/kudig-io/kudig/main/v2-go/operator/config/manager/
```

## 快速开始

### 创建集群诊断任务

```bash
cat <<EOF | kubectl apply -f -
apiVersion: kudig.io/v1
kind: ClusterDiagnostic
metadata:
  name: my-cluster-check
  namespace: kudig-system
spec:
  mode: online
  outputFormat: json
EOF
```

### 查看诊断状态

```bash
# 查看 ClusterDiagnostic 列表
kubectl get clusterdiagnostics

# 查看详细状态
kubectl describe clusterdiagnostic my-cluster-check

# 查看诊断结果摘要
kubectl get clusterdiagnostic my-cluster-check -o jsonpath='{.status.summary}'
```

### 创建定时诊断任务

```bash
cat <<EOF | kubectl apply -f -
apiVersion: kudig.io/v1
kind: DiagnosticSchedule
metadata:
  name: hourly-check
  namespace: kudig-system
spec:
  schedule: "@hourly"
  type: cluster
  clusterDiagnosticTemplate:
    mode: online
    outputFormat: json
    notify:
      enabled: true
      minSeverity: warning
EOF
```

## CRD 参考

### ClusterDiagnostic

| 字段 | 类型 | 描述 |
|------|------|------|
| spec.mode | string | 诊断模式: `online` 或 `offline` |
| spec.analyzers | []string | 要运行的分析器列表，为空则运行所有 |
| spec.excludeAnalyzers | []string | 要排除的分析器列表 |
| spec.outputFormat | string | 输出格式: `text`, `json`, `html` |
| spec.notify.enabled | bool | 是否启用通知 |
| spec.notify.webhookURL | string | Webhook URL |
| spec.notify.minSeverity | string | 最小通知严重级别 |

### NodeDiagnostic

| 字段 | 类型 | 描述 |
|------|------|------|
| spec.nodeSelector | map[string]string | 节点选择器 |
| spec.nodeNames | []string | 指定节点名称列表 |
| 其他字段 | - | 同 ClusterDiagnostic |

### DiagnosticSchedule

| 字段 | 类型 | 描述 |
|------|------|------|
| spec.schedule | string | Cron 表达式或预定义 (@hourly, @daily, @weekly) |
| spec.suspend | bool | 是否暂停调度 |
| spec.type | string | 诊断类型: `cluster` 或 `node` |
| spec.historyLimit | int | 保留历史任务数量 |

## 开发

### 本地运行

```bash
cd v2-go/operator

# 安装依赖
go mod tidy

# 本地运行（需要 kubeconfig）
go run cmd/main.go
```

### 构建镜像

```bash
# 构建 Operator 镜像
docker build -t kudig/kudig-operator:latest .

# 推送镜像
docker push kudig/kudig-operator:latest
```

## 架构

```
┌─────────────────────────────────────────────────────────┐
│                     Kubernetes Cluster                   │
│  ┌─────────────────┐    ┌─────────────────────────────┐ │
│  │  Kudig Operator │────│  ClusterDiagnostic CR       │ │
│  │                 │    │  NodeDiagnostic CR          │ │
│  │  - Reconciler   │    │  DiagnosticSchedule CR      │ │
│  │  - Scheduler    │    └─────────────────────────────┘ │
│  └─────────────────┘                                     │
│           │                                              │
│           ▼                                              │
│  ┌─────────────────┐    ┌─────────────────────────────┐ │
│  │  Kubernetes Job │────│  Kudig Diagnostic Pod       │ │
│  │  or DaemonSet   │    │  - Run analyzers            │ │
│  └─────────────────┘    │  - Generate reports         │ │
│                         │  - Send notifications       │ │
│                         └─────────────────────────────┘ │
└─────────────────────────────────────────────────────────┘
```

## 许可证

Apache License 2.0
