package events

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/kudig-io/klaw/internal/kubernetes"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	k8sclient "k8s.io/client-go/kubernetes"
)

// KubernetesSource Kubernetes 事件源
type KubernetesSource struct {
	BaseSource
	k8sManager  *kubernetes.Manager
	clusterName string
	client      *k8sclient.Clientset
	stopCh      chan struct{}
	mu          sync.RWMutex
	watchers    map[string]watch.Interface
	informers   map[string]interface{}
}

// NewKubernetesSource 创建 Kubernetes 事件源
func NewKubernetesSource(clusterName string, k8sManager *kubernetes.Manager) (*KubernetesSource, error) {
	client, err := k8sManager.GetClient(clusterName)
	if err != nil {
		return nil, fmt.Errorf("failed to get client for cluster %s: %v", clusterName, err)
	}
	
	return &KubernetesSource{
		BaseSource:  *NewBaseSource(fmt.Sprintf("kubernetes-%s", clusterName)),
		k8sManager:  k8sManager,
		clusterName: clusterName,
		client:      client,
		stopCh:      make(chan struct{}),
		watchers:    make(map[string]watch.Interface),
		informers:   make(map[string]interface{}),
	}, nil
}

// Start 启动 Kubernetes 事件监听
func (s *KubernetesSource) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if s.running {
		return nil
	}
	
	s.ctx, s.cancel = context.WithCancel(ctx)
	s.running = true
	
	// 启动事件 Watch
	go s.watchEvents()
	
	// 启动 Pod Watch（用于获取更详细的 Pod 事件）
	go s.watchPods()
	
	// 启动 Deployment Watch
	go s.watchDeployments()
	
	fmt.Printf("Kubernetes event source started for cluster: %s\n", s.clusterName)
	return nil
}

// Stop 停止 Kubernetes 事件监听
func (s *KubernetesSource) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if !s.running {
		return nil
	}
	
	s.running = false
	close(s.stopCh)
	
	// 停止所有 watchers
	for _, w := range s.watchers {
		w.Stop()
	}
	
	if s.cancel != nil {
		s.cancel()
	}
	
	fmt.Printf("Kubernetes event source stopped for cluster: %s\n", s.clusterName)
	return nil
}

// watchEvents 监听 K8s Events
func (s *KubernetesSource) watchEvents() {
	client := s.client
	
	// 创建 Watch
	watchInterface, err := client.CoreV1().Events("").Watch(s.ctx, metav1.ListOptions{
		Watch: true,
	})
	if err != nil {
		fmt.Printf("Failed to watch events: %v\n", err)
		return
	}
	
	s.mu.Lock()
	s.watchers["events"] = watchInterface
	s.mu.Unlock()
	
	// 处理事件
	for {
		select {
		case <-s.stopCh:
			return
		case event, ok := <-watchInterface.ResultChan():
			if !ok {
				// 重新连接
				time.Sleep(5 * time.Second)
				go s.watchEvents()
				return
			}
			
			if event.Type == watch.Error {
				fmt.Printf("Watch error: %v\n", event.Object)
				continue
			}
			
			k8sEvent, ok := event.Object.(*corev1.Event)
			if !ok {
				continue
			}
			
			// 转换为统一事件格式
			ev := s.convertK8sEvent(k8sEvent)
			
			// 处理事件类型
			switch event.Type {
			case watch.Added, watch.Modified:
				s.emit(ev)
			}
		}
	}
}

// watchPods 监听 Pod 变化
func (s *KubernetesSource) watchPods() {
	client := s.client
	
	// 创建 Watch
	watchInterface, err := client.CoreV1().Pods("").Watch(s.ctx, metav1.ListOptions{
		Watch: true,
	})
	if err != nil {
		fmt.Printf("Failed to watch pods: %v\n", err)
		return
	}
	
	s.mu.Lock()
	s.watchers["pods"] = watchInterface
	s.mu.Unlock()
	
	for {
		select {
		case <-s.stopCh:
			return
		case event, ok := <-watchInterface.ResultChan():
			if !ok {
				time.Sleep(5 * time.Second)
				go s.watchPods()
				return
			}
			
			if event.Type == watch.Error {
				continue
			}
			
			pod, ok := event.Object.(*corev1.Pod)
			if !ok {
				continue
			}
			
			// 只处理特定的 Pod 事件
			if event.Type == watch.Deleted {
				ev := &Event{
					ID:        string(pod.UID),
					Type:      EventTypeDelete,
					ResourceType: ResourcePod,
					ResourceName: pod.Name,
					Namespace: pod.Namespace,
					Cluster:   s.clusterName,
					Reason:    "Deleted",
					Message:   fmt.Sprintf("Pod %s/%s was deleted", pod.Namespace, pod.Name),
					Timestamp: time.Now(),
					InvolvedObject: InvolvedObject{
						Kind:      "Pod",
						Name:      pod.Name,
						Namespace: pod.Namespace,
						UID:       string(pod.UID),
					},
				}
				s.emit(ev)
			}
		}
	}
}

// watchDeployments 监听 Deployment 变化
func (s *KubernetesSource) watchDeployments() {
	client := s.client
	
	// 创建 Watch
	watchInterface, err := client.AppsV1().Deployments("").Watch(s.ctx, metav1.ListOptions{
		Watch: true,
	})
	if err != nil {
		fmt.Printf("Failed to watch deployments: %v\n", err)
		return
	}
	
	s.mu.Lock()
	s.watchers["deployments"] = watchInterface
	s.mu.Unlock()
	
	for {
		select {
		case <-s.stopCh:
			return
		case event, ok := <-watchInterface.ResultChan():
			if !ok {
				time.Sleep(5 * time.Second)
				go s.watchDeployments()
				return
			}
			
			if event.Type == watch.Error {
				continue
			}
			
			// 简化处理，不转换为 Event
			// 详细的 Deployment 事件通过 Event Watch 获取
		}
	}
}

// convertK8sEvent 转换 K8s Event 为统一事件格式
func (s *KubernetesSource) convertK8sEvent(k8sEvent *corev1.Event) *Event {
	eventType := EventTypeNormal
	if k8sEvent.Type == "Warning" {
		eventType = EventTypeWarning
	}
	
	// 映射资源类型
	resourceType := ResourceType(k8sEvent.InvolvedObject.Kind)
	
	return &Event{
		ID:        string(k8sEvent.UID),
		Type:      eventType,
		ResourceType: resourceType,
		ResourceName: k8sEvent.InvolvedObject.Name,
		Namespace: k8sEvent.InvolvedObject.Namespace,
		Cluster:   s.clusterName,
		Reason:    k8sEvent.Reason,
		Message:   k8sEvent.Message,
		Timestamp: k8sEvent.LastTimestamp.Time,
		Count:     k8sEvent.Count,
		InvolvedObject: InvolvedObject{
			Kind:       k8sEvent.InvolvedObject.Kind,
			Name:       k8sEvent.InvolvedObject.Name,
			Namespace:  k8sEvent.InvolvedObject.Namespace,
			UID:        string(k8sEvent.InvolvedObject.UID),
			APIVersion: k8sEvent.InvolvedObject.APIVersion,
		},
		Labels:      k8sEvent.Labels,
		Annotations: k8sEvent.Annotations,
	}
}


