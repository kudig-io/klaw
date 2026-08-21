package kubernetes

import (
	"fmt"
	"path/filepath"
	"os"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	"github.com/kudig-io/klaw/internal/config"
)

// InClusterKubeconfig 是 kubeconfig 字段的特殊值，表示使用 Pod 内置的 ServiceAccount 凭据
const InClusterKubeconfig = "in-cluster"

// Manager Kubernetes管理器
type Manager struct {
	clients map[string]*kubernetes.Clientset
	clusters []config.ClusterConfig
}

// NewManager 创建Kubernetes管理器
func NewManager(cfg config.KubernetesConfig) (*Manager, error) {
	m := &Manager{
		clients: make(map[string]*kubernetes.Clientset),
		clusters: cfg.Clusters,
	}

	// 初始化所有集群连接
	for _, cluster := range cfg.Clusters {
		client, err := m.initClient(cluster)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize cluster %s: %v", cluster.Name, err)
		}
		m.clients[cluster.Name] = client
	}

	return m, nil
}

// initClient 初始化集群客户端
func (m *Manager) initClient(cluster config.ClusterConfig) (*kubernetes.Clientset, error) {
	// in-cluster 模式：显式声明，或未指定 kubeconfig 且运行在 Pod 内时自动启用
	if cluster.Kubeconfig == InClusterKubeconfig || (cluster.Kubeconfig == "" && isRunningInCluster()) {
		clientConfig, err := rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to build in-cluster config: %v", err)
		}
		client, err := kubernetes.NewForConfig(clientConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create in-cluster client: %v", err)
		}
		return client, nil
	}

	// 确定kubeconfig路径
	kubeconfig := cluster.Kubeconfig
	if kubeconfig == "" {
		if home := homedir.HomeDir(); home != "" {
			kubeconfig = filepath.Join(home, ".kube", "config")
		} else {
			return nil, fmt.Errorf("could not determine home directory")
		}
	}

	// 检查kubeconfig文件是否存在
	if _, err := os.Stat(kubeconfig); os.IsNotExist(err) {
		return nil, fmt.Errorf("kubeconfig file not found: %s", kubeconfig)
	}

	// 加载kubeconfig
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	rules.ExplicitPath = kubeconfig

	// 创建配置
	configOverrides := &clientcmd.ConfigOverrides{}
	if cluster.Context != "" {
		configOverrides.CurrentContext = cluster.Context
	}

	// 构建客户端配置
	clientConfig, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		rules, configOverrides).ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to build client config: %v", err)
	}

	// 创建客户端
	client, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %v", err)
	}

	return client, nil
}

// isRunningInCluster 判断当前进程是否运行在 Kubernetes Pod 内
func isRunningInCluster() bool {
	return os.Getenv("KUBERNETES_SERVICE_HOST") != "" && os.Getenv("KUBERNETES_SERVICE_PORT") != ""
}

// GetClient 获取集群客户端；clusterName 为空时回退到配置中的第一个集群
func (m *Manager) GetClient(clusterName string) (*kubernetes.Clientset, error) {
	if clusterName == "" {
		if len(m.clusters) == 0 {
			return nil, fmt.Errorf("no cluster configured")
		}
		clusterName = m.clusters[0].Name
	}
	client, ok := m.clients[clusterName]
	if !ok {
		return nil, fmt.Errorf("cluster not found: %s", clusterName)
	}
	return client, nil
}

// GetClusters 获取所有集群
func (m *Manager) GetClusters() []config.ClusterConfig {
	return m.clusters
}

// RefreshClient 刷新集群客户端
func (m *Manager) RefreshClient(clusterName string) error {
	// 查找集群配置
	var cluster config.ClusterConfig
	found := false
	for _, c := range m.clusters {
		if c.Name == clusterName {
			cluster = c
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("cluster not found: %s", clusterName)
	}

	// 重新初始化客户端
	client, err := m.initClient(cluster)
	if err != nil {
		return fmt.Errorf("failed to refresh cluster %s: %v", clusterName, err)
	}

	m.clients[clusterName] = client
	return nil
}
