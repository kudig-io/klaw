#!/bin/bash
#
# Kind 集群管理脚本
# 用于 Klaw 本地开发和测试
#

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CLUSTER_NAME="klaw-test"
CONFIG_FILE="${SCRIPT_DIR}/cluster-config.yaml"
DATA_DIR="${SCRIPT_DIR}/data"

# 打印带颜色的消息
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查 kind 是否安装
check_kind() {
    if ! command -v kind &> /dev/null; then
        print_error "kind 未安装"
        echo "请运行以下命令安装 kind:"
        echo "  brew install kind    # macOS"
        echo "  或访问: https://kind.sigs.k8s.io/docs/user/quick-start/#installation"
        exit 1
    fi
    print_success "kind 已安装: $(kind version)"
}

# 检查 kubectl 是否安装
check_kubectl() {
    if ! command -v kubectl &> /dev/null; then
        print_error "kubectl 未安装"
        echo "请运行以下命令安装 kubectl:"
        echo "  brew install kubectl    # macOS"
        exit 1
    fi
    print_success "kubectl 已安装: $(kubectl version --client -o json 2>/dev/null | grep -o '"gitVersion": "[^"]*"' | head -1)"
}

# 检查 Docker 是否运行
check_docker() {
    if ! docker info &> /dev/null; then
        print_error "Docker 未运行"
        exit 1
    fi
    print_success "Docker 正在运行"
}

# 创建集群
create_cluster() {
    print_info "正在创建 kind 集群: ${CLUSTER_NAME}..."
    
    # 检查集群是否已存在
    if kind get clusters | grep -q "^${CLUSTER_NAME}$"; then
        print_warning "集群 ${CLUSTER_NAME} 已存在"
        read -p "是否删除并重新创建? (y/N): " confirm
        if [[ $confirm =~ ^[Yy]$ ]]; then
            delete_cluster
        else
            print_info "跳过创建"
            return
        fi
    fi
    
    # 创建数据目录
    mkdir -p "${DATA_DIR}"
    
    # 创建集群
    kind create cluster --name "${CLUSTER_NAME}" --config "${CONFIG_FILE}"
    
    print_success "集群 ${CLUSTER_NAME} 创建成功!"
    
    # 设置 kubectl 上下文
    kubectl config use-context "kind-${CLUSTER_NAME}"
    print_success "kubectl 上下文已切换到: kind-${CLUSTER_NAME}"
    
    # 显示集群信息
    show_cluster_info
}

# 删除集群
delete_cluster() {
    print_info "正在删除集群: ${CLUSTER_NAME}..."
    kind delete cluster --name "${CLUSTER_NAME}"
    print_success "集群 ${CLUSTER_NAME} 已删除"
}

# 显示集群信息
show_cluster_info() {
    print_info "集群信息:"
    echo "  名称: ${CLUSTER_NAME}"
    echo "  上下文: kind-${CLUSTER_NAME}"
    echo "  kubeconfig: ~/.kube/config"
    
    if kind get clusters | grep -q "^${CLUSTER_NAME}$"; then
        echo "  状态: ${GREEN}运行中${NC}"
        echo ""
        print_info "节点列表:"
        kubectl get nodes
        echo ""
        print_info "命名空间:"
        kubectl get namespaces
    else
        echo "  状态: ${RED}未创建${NC}"
    fi
}

# 导出 kubeconfig
export_kubeconfig() {
    local output_file="${SCRIPT_DIR}/kubeconfig"
    print_info "导出 kubeconfig 到: ${output_file}"
    kind get kubeconfig --name "${CLUSTER_NAME}" > "${output_file}"
    print_success "kubeconfig 已导出: ${output_file}"
    echo ""
    echo "使用方式:"
    echo "  export KUBECONFIG=${output_file}"
}

# 部署测试应用
deploy_sample() {
    print_info "部署测试应用到集群..."
    
    # 创建测试命名空间
    kubectl create namespace klaw-test --dry-run=client -o yaml | kubectl apply -f -
    
    # 部署 nginx
    kubectl create deployment nginx --image=nginx:alpine --replicas=3 -n klaw-test --dry-run=client -o yaml | kubectl apply -f -
    
    # 创建 service
    kubectl expose deployment nginx --port=80 --type=ClusterIP -n klaw-test --dry-run=client -o yaml | kubectl apply -f -
    
    # 部署另一个应用用于测试
    kubectl create deployment httpbin --image=kennethreitz/httpbin --replicas=2 -n klaw-test --dry-run=client -o yaml | kubectl apply -f -
    
    print_success "测试应用部署完成!"
    
    echo ""
    print_info "查看部署状态:"
    kubectl get deployments -n klaw-test
    echo ""
    kubectl get pods -n klaw-test
    echo ""
    kubectl get services -n klaw-test
}

# 加载本地镜像到集群
load_image() {
    local image_name=$1
    if [ -z "$image_name" ]; then
        print_error "请指定镜像名称"
        echo "用法: $0 load-image <镜像名称>"
        exit 1
    fi
    
    print_info "加载镜像到集群: ${image_name}"
    kind load docker-image "${image_name}" --name "${CLUSTER_NAME}"
    print_success "镜像加载完成"
}

# 安装常用工具
install_tools() {
    print_info "安装常用工具到集群..."
    
    # 安装 metrics-server
    print_info "安装 metrics-server..."
    kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml
    kubectl patch deployment metrics-server -n kube-system --type='json' -p='[{"op": "add", "path": "/spec/template/spec/containers/0/args/-", "value": "--kubelet-insecure-tls"}]'
    
    print_success "工具安装完成!"
}

# 显示帮助信息
show_help() {
    cat << EOF
Kind 集群管理脚本 - 用于 Klaw 本地开发和测试

用法: $0 <命令> [参数]

命令:
  create          创建 kind 集群
  delete          删除 kind 集群
  status          显示集群状态
  export-config   导出 kubeconfig 文件
  deploy-sample   部署测试应用到集群
  load-image <镜像>  加载本地 Docker 镜像到集群
  install-tools   安装常用工具 (metrics-server 等)
  check           检查前置依赖 (kind, kubectl, docker)
  help            显示帮助信息

示例:
  $0 create           # 创建集群
  $0 deploy-sample    # 部署测试应用
  $0 status           # 查看集群状态
  $0 delete           # 删除集群

配置文件:
  ${CONFIG_FILE}

数据目录:
  ${DATA_DIR}
EOF
}

# 主函数
main() {
    case "${1:-help}" in
        create)
            check_docker
            check_kind
            check_kubectl
            create_cluster
            ;;
        delete)
            check_kind
            delete_cluster
            ;;
        status)
            check_kind
            show_cluster_info
            ;;
        export-config)
            check_kind
            export_kubeconfig
            ;;
        deploy-sample)
            check_kind
            check_kubectl
            deploy_sample
            ;;
        load-image)
            load_image "$2"
            ;;
        install-tools)
            check_kind
            check_kubectl
            install_tools
            ;;
        check)
            check_docker
            check_kind
            check_kubectl
            print_success "所有依赖检查通过!"
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            print_error "未知命令: $1"
            show_help
            exit 1
            ;;
    esac
}

main "$@"
