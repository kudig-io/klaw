package backup

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type BackupData struct {
	Metadata  BackupMetadata               `json:"metadata"`
	Resources map[string][]json.RawMessage `json:"resources"`
}

type BackupMetadata struct {
	Name        string `json:"name"`
	Timestamp   string `json:"timestamp"`
	ClusterName string `json:"clusterName"`
}

type BackupInfo struct {
	Name           string `json:"name"`
	Timestamp      string `json:"timestamp"`
	FilePath       string `json:"filePath"`
	Size           int64  `json:"size"`
	ResourcesCount int    `json:"resourcesCount"`
}

type ClusterOperator struct {
	clientset kubernetes.Interface
	backupDir string
}

func NewClusterOperator(clientset kubernetes.Interface, backupDir string) *ClusterOperator {
	return &ClusterOperator{clientset: clientset, backupDir: backupDir}
}

func (o *ClusterOperator) BackupCluster(ctx context.Context, name string) (*BackupInfo, error) {
	if err := os.MkdirAll(o.backupDir, 0o755); err != nil {
		return nil, fmt.Errorf("create backup dir: %w", err)
	}

	data := BackupData{
		Metadata: BackupMetadata{
			Name:        name,
			Timestamp:   time.Now().UTC().Format(time.RFC3339),
			ClusterName: name,
		},
		Resources: make(map[string][]json.RawMessage),
	}

	if items, err := o.clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{}); err == nil {
		for i := range items.Items {
			data.Resources["namespaces"] = append(data.Resources["namespaces"], sanitize(&items.Items[i]))
		}
	}
	if items, err := o.clientset.CoreV1().ServiceAccounts("").List(ctx, metav1.ListOptions{}); err == nil {
		for i := range items.Items {
			data.Resources["serviceaccounts"] = append(data.Resources["serviceaccounts"], sanitize(&items.Items[i]))
		}
	}
	if items, err := o.clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{}); err == nil {
		for i := range items.Items {
			data.Resources["pods"] = append(data.Resources["pods"], sanitize(&items.Items[i]))
		}
	}
	if items, err := o.clientset.AppsV1().Deployments("").List(ctx, metav1.ListOptions{}); err == nil {
		for i := range items.Items {
			data.Resources["deployments"] = append(data.Resources["deployments"], sanitize(&items.Items[i]))
		}
	}
	if items, err := o.clientset.CoreV1().Services("").List(ctx, metav1.ListOptions{}); err == nil {
		for i := range items.Items {
			data.Resources["services"] = append(data.Resources["services"], sanitize(&items.Items[i]))
		}
	}
	if items, err := o.clientset.CoreV1().ConfigMaps("").List(ctx, metav1.ListOptions{}); err == nil {
		for i := range items.Items {
			data.Resources["configmaps"] = append(data.Resources["configmaps"], sanitize(&items.Items[i]))
		}
	}
	if items, err := o.clientset.CoreV1().Secrets("").List(ctx, metav1.ListOptions{}); err == nil {
		for i := range items.Items {
			data.Resources["secrets"] = append(data.Resources["secrets"], sanitize(&items.Items[i]))
		}
	}
	if items, err := o.clientset.NetworkingV1().Ingresses("").List(ctx, metav1.ListOptions{}); err == nil {
		for i := range items.Items {
			data.Resources["ingresses"] = append(data.Resources["ingresses"], sanitize(&items.Items[i]))
		}
	}
	if items, err := o.clientset.NetworkingV1().NetworkPolicies("").List(ctx, metav1.ListOptions{}); err == nil {
		for i := range items.Items {
			data.Resources["networkpolicies"] = append(data.Resources["networkpolicies"], sanitize(&items.Items[i]))
		}
	}
	if items, err := o.clientset.CoreV1().PersistentVolumes().List(ctx, metav1.ListOptions{}); err == nil {
		for i := range items.Items {
			data.Resources["persistentvolumes"] = append(data.Resources["persistentvolumes"], sanitize(&items.Items[i]))
		}
	}
	if items, err := o.clientset.CoreV1().PersistentVolumeClaims("").List(ctx, metav1.ListOptions{}); err == nil {
		for i := range items.Items {
			data.Resources["persistentvolumeclaims"] = append(data.Resources["persistentvolumeclaims"], sanitize(&items.Items[i]))
		}
	}

	total := 0
	for _, list := range data.Resources {
		total += len(list)
	}

	payload, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal backup: %w", err)
	}
	filename := fmt.Sprintf("%s_%s.json", name, time.Now().UTC().Format("20060102_150405"))
	path := filepath.Join(o.backupDir, filename)
	if err := os.WriteFile(path, payload, 0o644); err != nil {
		return nil, fmt.Errorf("write backup file: %w", err)
	}
	info, _ := os.Stat(path)
	return &BackupInfo{
		Name:           name,
		Timestamp:      data.Metadata.Timestamp,
		FilePath:       path,
		Size:           info.Size(),
		ResourcesCount: total,
	}, nil
}

func (o *ClusterOperator) RestoreCluster(ctx context.Context, backupFile string) error {
	payload, err := os.ReadFile(backupFile)
	if err != nil {
		return fmt.Errorf("read backup file: %w", err)
	}
	var data BackupData
	if err := json.Unmarshal(payload, &data); err != nil {
		return fmt.Errorf("unmarshal backup: %w", err)
	}
	for kind, rawList := range data.Resources {
		for _, raw := range rawList {
			if err := o.restoreItem(ctx, kind, raw); err != nil {
				return err
			}
		}
	}
	return nil
}

func (o *ClusterOperator) restoreItem(ctx context.Context, kind string, raw json.RawMessage) error {
	switch kind {
	case "namespaces":
		var obj corev1.Namespace
		if err := json.Unmarshal(raw, &obj); err != nil {
			return err
		}
		if _, err := o.clientset.CoreV1().Namespaces().Create(ctx, &obj, metav1.CreateOptions{}); err != nil {
			_, err = o.clientset.CoreV1().Namespaces().Update(ctx, &obj, metav1.UpdateOptions{})
			return err
		}
	case "serviceaccounts":
		var obj corev1.ServiceAccount
		if err := json.Unmarshal(raw, &obj); err != nil {
			return err
		}
		if _, err := o.clientset.CoreV1().ServiceAccounts(obj.Namespace).Create(ctx, &obj, metav1.CreateOptions{}); err != nil {
			_, err = o.clientset.CoreV1().ServiceAccounts(obj.Namespace).Update(ctx, &obj, metav1.UpdateOptions{})
			return err
		}
	case "pods":
		var obj corev1.Pod
		if err := json.Unmarshal(raw, &obj); err != nil {
			return err
		}
		if _, err := o.clientset.CoreV1().Pods(obj.Namespace).Create(ctx, &obj, metav1.CreateOptions{}); err != nil {
			_, err = o.clientset.CoreV1().Pods(obj.Namespace).Update(ctx, &obj, metav1.UpdateOptions{})
			return err
		}
	case "deployments":
		var obj appsv1.Deployment
		if err := json.Unmarshal(raw, &obj); err != nil {
			return err
		}
		if _, err := o.clientset.AppsV1().Deployments(obj.Namespace).Create(ctx, &obj, metav1.CreateOptions{}); err != nil {
			_, err = o.clientset.AppsV1().Deployments(obj.Namespace).Update(ctx, &obj, metav1.UpdateOptions{})
			return err
		}
	case "services":
		var obj corev1.Service
		if err := json.Unmarshal(raw, &obj); err != nil {
			return err
		}
		if _, err := o.clientset.CoreV1().Services(obj.Namespace).Create(ctx, &obj, metav1.CreateOptions{}); err != nil {
			_, err = o.clientset.CoreV1().Services(obj.Namespace).Update(ctx, &obj, metav1.UpdateOptions{})
			return err
		}
	case "configmaps":
		var obj corev1.ConfigMap
		if err := json.Unmarshal(raw, &obj); err != nil {
			return err
		}
		if _, err := o.clientset.CoreV1().ConfigMaps(obj.Namespace).Create(ctx, &obj, metav1.CreateOptions{}); err != nil {
			_, err = o.clientset.CoreV1().ConfigMaps(obj.Namespace).Update(ctx, &obj, metav1.UpdateOptions{})
			return err
		}
	case "secrets":
		var obj corev1.Secret
		if err := json.Unmarshal(raw, &obj); err != nil {
			return err
		}
		if _, err := o.clientset.CoreV1().Secrets(obj.Namespace).Create(ctx, &obj, metav1.CreateOptions{}); err != nil {
			_, err = o.clientset.CoreV1().Secrets(obj.Namespace).Update(ctx, &obj, metav1.UpdateOptions{})
			return err
		}
	case "ingresses":
		var obj networkingv1.Ingress
		if err := json.Unmarshal(raw, &obj); err != nil {
			return err
		}
		if _, err := o.clientset.NetworkingV1().Ingresses(obj.Namespace).Create(ctx, &obj, metav1.CreateOptions{}); err != nil {
			_, err = o.clientset.NetworkingV1().Ingresses(obj.Namespace).Update(ctx, &obj, metav1.UpdateOptions{})
			return err
		}
	case "networkpolicies":
		var obj networkingv1.NetworkPolicy
		if err := json.Unmarshal(raw, &obj); err != nil {
			return err
		}
		if _, err := o.clientset.NetworkingV1().NetworkPolicies(obj.Namespace).Create(ctx, &obj, metav1.CreateOptions{}); err != nil {
			_, err = o.clientset.NetworkingV1().NetworkPolicies(obj.Namespace).Update(ctx, &obj, metav1.UpdateOptions{})
			return err
		}
	case "persistentvolumes":
		var obj corev1.PersistentVolume
		if err := json.Unmarshal(raw, &obj); err != nil {
			return err
		}
		if _, err := o.clientset.CoreV1().PersistentVolumes().Create(ctx, &obj, metav1.CreateOptions{}); err != nil {
			_, err = o.clientset.CoreV1().PersistentVolumes().Update(ctx, &obj, metav1.UpdateOptions{})
			return err
		}
	case "persistentvolumeclaims":
		var obj corev1.PersistentVolumeClaim
		if err := json.Unmarshal(raw, &obj); err != nil {
			return err
		}
		if _, err := o.clientset.CoreV1().PersistentVolumeClaims(obj.Namespace).Create(ctx, &obj, metav1.CreateOptions{}); err != nil {
			_, err = o.clientset.CoreV1().PersistentVolumeClaims(obj.Namespace).Update(ctx, &obj, metav1.UpdateOptions{})
			return err
		}
	}
	return nil
}

func (o *ClusterOperator) ListBackups() ([]BackupInfo, error) {
	entries, err := os.ReadDir(o.backupDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read backup dir: %w", err)
	}
	var backups []BackupInfo
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		path := filepath.Join(o.backupDir, e.Name())
		fi, err := e.Info()
		if err != nil {
			continue
		}
		bi := BackupInfo{
			Name:     strings.TrimSuffix(e.Name(), filepath.Ext(e.Name())),
			FilePath: path,
			Size:     fi.Size(),
		}
		if data, err := os.ReadFile(path); err == nil {
			var bd BackupData
			if json.Unmarshal(data, &bd) == nil {
				bi.Name = bd.Metadata.Name
				bi.Timestamp = bd.Metadata.Timestamp
				for _, list := range bd.Resources {
					bi.ResourcesCount += len(list)
				}
			}
		}
		backups = append(backups, bi)
	}
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].Timestamp > backups[j].Timestamp
	})
	return backups, nil
}

func (o *ClusterOperator) DeleteBackup(backupFile string) error {
	if err := os.Remove(backupFile); err != nil {
		return fmt.Errorf("delete backup: %w", err)
	}
	return nil
}

func sanitize(obj any) json.RawMessage {
	raw, _ := json.Marshal(obj)
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return raw
	}
	if meta, ok := m["metadata"].(map[string]any); ok {
		delete(meta, "resourceVersion")
		delete(meta, "uid")
		delete(meta, "creationTimestamp")
		delete(meta, "generation")
		delete(meta, "managedFields")
		delete(meta, "selfLink")
		if ann, ok := meta["annotations"].(map[string]any); ok {
			delete(ann, "kubectl.kubernetes.io/last-applied-configuration")
			delete(ann, "deployment.kubernetes.io/revision")
		}
	}
	delete(m, "status")
	out, _ := json.Marshal(m)
	return out
}
