# Klaw Deployment

本目录包含 Klaw 的部署相关配置和脚本，用于本地开发和测试环境的快速搭建。

## 目录结构

```
deployment/
├── README.md                 # 本文档
├── kind/                     # Kind 本地 K8s 集群配置
│   ├── cluster-config.yaml   # Kind 集群配置文件
│   ├── manage.sh            # Kind 集群管理脚本
│   ├── data/                # 数据持久化目录
│   └── kubeconfig           # 导出的 kubeconfig 文件（创建集群后生成）
```

## 前置依赖

- [Docker](https://docs.docker.com/get-docker/)
- [Kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)

### 安装依赖 (macOS)

```bash
# 安装 Docker
brew install --cask docker

# 安装 Kind
brew install kind

# 安装 kubectl
brew install kubectl
```

## 快速开始

### 1. 检查依赖

```bash
cd deployment/kind
./manage.sh check
```

### 2. 创建测试集群

```bash
./manage.sh create
```

这会创建一个名为 `klaw-test` 的 Kind 集群，包含：
- 1 个控制平面节点
- 2 个工作节点

### 3. 部署测试应用

```bash
./manage.sh deploy-sample
```

会部署以下应用到 `klaw-test` 命名空间：
- nginx (3 副本)
- httpbin (2 副本)

### 4. 查看集群状态

```bash
./manage.sh status
```

### 5. 配置 Klaw 使用该集群

创建集群后，Klaw 的配置文件 `configs/config.yaml` 会自动使用正确的 kubeconfig 路径：

```yaml
kubernetes:
  clusters:
    - name: klaw-test
      kubeconfig: ~/.kube/config
      context: kind-klaw-test
```

或者直接使用导出的 kubeconfig：

```bash
./manage.sh export-config
export KUBECONFIG=$(pwd)/kind/kubeconfig
```

### 6. 启动 Klaw

```bash
cd /Users/allengaller/Documents/GitHub/kudig-io/klaw
./klaw
```

然后访问 http://localhost:8080

## 在 kind 中部署 Klaw（in-cluster，推荐）

以下流程已在 macOS + Docker Desktop + kind v0.31 + Kubernetes v1.35 上验证通过。
Klaw 以 Pod 形式运行在集群内，通过 ServiceAccount 凭据访问 API Server，无需挂载 kubeconfig。

```bash
# 1. 创建集群
kind create cluster --config deployment/kind/cluster-config.yaml

# 2. 构建镜像（国内网络可指定 GOPROXY）
docker build --build-arg GOPROXY=https://goproxy.cn,direct -t kudig-io/klaw:dev .

# 3. 加载镜像到集群节点
kind load docker-image kudig-io/klaw:dev --name klaw-test

# 4. 部署
helm upgrade --install klaw helm/klaw \
  -f helm/klaw/values-kind.yaml \
  -n klaw --create-namespace --wait

# 5. 访问
kubectl port-forward -n klaw svc/klaw 18080:8080
```

打开 http://127.0.0.1:18080 即为 Web UI，`/healthz`、`/readyz`、`/metrics` 为探针与指标端点。

### values-kind.yaml 关键点

- `image.tag: dev` + `pullPolicy: Never`：只用 `kind load` 注入的本地镜像，不回源拉取
- `config.server.auth.enabled: false`：Web UI 前端目前不会注入 Bearer token，本地默认关闭认证；
  若要验证 API 鉴权，改为 `true` 后用 `curl -H "Authorization: Bearer <secrets.apiToken>"` 访问
- `persistence.storageClass: standard`：使用 kind 自带的 local-path StorageClass

### 网络受限环境的镜像预拉取

Docker Hub 直连超时时，可先经镜像加速站拉取再打回原名：

```bash
for i in node:20-alpine golang:1.24-alpine alpine:3.20; do
  docker pull docker.1ms.run/library/$i && docker tag docker.1ms.run/library/$i $i
done

docker pull docker.1ms.run/kindest/node:v1.35.0
docker tag docker.1ms.run/kindest/node:v1.35.0 kindest/node:v1.35.0
```

注意：`~/.docker/daemon.json` 中的 `registry-mirrors` 必须位于顶层，放在 `builder` 段内不会生效。

## 管理脚本命令

```bash
./manage.sh create          # 创建集群
./manage.sh delete          # 删除集群
./manage.sh status          # 查看集群状态
./manage.sh export-config   # 导出 kubeconfig
./manage.sh deploy-sample   # 部署测试应用
./manage.sh load-image <镜像>  # 加载本地镜像到集群
./manage.sh install-tools   # 安装 metrics-server 等工具
./manage.sh check           # 检查依赖
./manage.sh help            # 显示帮助
```

## 集群配置

### Kind 集群配置

集群配置文件：`kind/cluster-config.yaml`

默认配置：
- 集群名称：`klaw-test`
- 网络：Pod 子网 `10.244.0.0/16`，Service 子网 `10.96.0.0/12`
- 端口映射：80 -> 8081, 443 -> 8443

### 自定义配置

编辑 `kind/cluster-config.yaml` 可以修改：
- 节点数量
- 端口映射
- 网络配置
- 挂载目录

## 网络访问

### 从主机访问集群内的 Service

使用 `kubectl port-forward`：

```bash
# 将集群内的 nginx service 转发到本地 8082 端口
kubectl port-forward -n klaw-test svc/nginx 8082:80
```

然后访问 http://localhost:8082

### 使用 Ingress（可选）

集群配置已启用 Ingress 支持，可以部署 ingress-nginx：

```bash
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml
```

## 常见问题

### 1. Docker 未运行

```bash
# macOS 上启动 Docker Desktop
open -a Docker
```

### 2. 端口冲突

如果 8081/8443 端口被占用，修改 `cluster-config.yaml` 中的 `extraPortMappings`。

### 3. 集群创建失败

```bash
# 删除已有集群
kind delete cluster --name klaw-test

# 重新创建
./manage.sh create
```

### 4. Klaw 无法连接集群

检查 kubeconfig 路径和上下文：

```bash
# 查看当前上下文
kubectl config current-context

# 应该显示: kind-klaw-test

# 查看可用的上下文
kubectl config get-contexts

# 切换到 kind 集群
kubectl config use-context kind-klaw-test
```

## 相关链接

- [Kind 官方文档](https://kind.sigs.k8s.io/)
- [Klaw README](../README.md)
- [Klaw 开发计划](../DEVELOPMENT_PLAN.md)
