package api

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/gorilla/mux"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/kudig-io/klaw/internal/alerting"
	"github.com/kudig-io/klaw/internal/audit"
	"github.com/kudig-io/klaw/internal/automation"
	"github.com/kudig-io/klaw/internal/backup"
	"github.com/kudig-io/klaw/internal/kubernetes"
	"github.com/kudig-io/klaw/internal/monitoring"
	"github.com/kudig-io/klaw/internal/networkanalysis"
	"github.com/kudig-io/klaw/internal/storage"
	"github.com/kudig-io/klaw/internal/storageanalysis"
	"github.com/kudig-io/klaw/internal/tenancy"
)

const testClusterName = "test-cluster"

// newK8sTestServer 构造注入 fake clientset 的测试 Server：
// k8sManager/resources/monitoringService/tenancyManager 全部指向同一个 fake 集群，
// 并手动注册被测的资源类路由（与 SetupRoutes 中 v1 路由等价）。
// 返回 fake clientset 供测试直接断言集群内对象。
func newK8sTestServer(t *testing.T, objs ...runtime.Object) (*Server, *fake.Clientset) {
	t.Helper()
	client := fake.NewSimpleClientset(objs...)
	mgr := kubernetes.NewManagerWithClient(testClusterName, client)
	store, err := storage.NewStore(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	s := &Server{
		k8sManager:        mgr,
		monitoringService: monitoring.NewService(mgr),
		alertingManager:   alerting.NewManager(nil, store),
		backupManager:     backup.NewManager(store),
		tenancyManager:    tenancy.NewManager(mgr, store),
		auditLogger:       audit.NewLogger(store),
		automationManager: automation.NewManager(store),
		resources:         kubernetes.NewResources(mgr),
		router:            mux.NewRouter(),
	}
	// analysis 路由（/api/v1/analysis/network、/api/v1/analysis/storage）
	s.setupAnalysisV1Routes()

	// 与 unified_v1.go 中 setupUnifiedV1Routes 的资源路由子集保持一致
	s.router.HandleFunc("/api/v1/clusters/{cluster}/ingresses", s.handleListIngresses).Methods("GET")
	s.router.HandleFunc("/api/v1/clusters/{cluster}/namespaces/{namespace}/ingresses", s.handleListIngresses).Methods("GET")
	s.router.HandleFunc("/api/v1/clusters/{cluster}/networkpolicies", s.handleListNetworkPolicies).Methods("GET")
	s.router.HandleFunc("/api/v1/clusters/{cluster}/namespaces/{namespace}/networkpolicies", s.handleListNetworkPolicies).Methods("GET")
	s.router.HandleFunc("/api/v1/clusters/{cluster}/persistentvolumeclaims", s.handleListPVCs).Methods("GET")
	s.router.HandleFunc("/api/v1/clusters/{cluster}/namespaces/{namespace}/persistentvolumeclaims", s.handleListPVCs).Methods("GET")
	s.router.HandleFunc("/api/v1/clusters/{cluster}/persistentvolumes", s.handleListPersistentVolumes).Methods("GET")
	s.router.HandleFunc("/api/v1/clusters/{cluster}/storageclasses", s.handleListStorageClasses).Methods("GET")
	s.router.HandleFunc("/api/v1/clusters/{cluster}/monitor/status", s.handleGetMonitorStatus).Methods("GET")
	s.router.HandleFunc("/api/v1/clusters/{cluster}/resources/{kind}", s.handleListUnifiedResources).Methods("GET")
	s.router.HandleFunc("/api/v1/clusters/{cluster}/namespaces/{namespace}/resources/{kind}", s.handleListUnifiedResources).Methods("GET")
	return s, client
}

// TestListIngressesFiltersByNamespace 验证 Ingress 列表：集群级返回全部命名空间的
// Ingress，命名空间级只返回该命名空间的 Ingress，且返回体保留 spec 细节。
func TestListIngressesFiltersByNamespace(t *testing.T) {
	seedIngress := func(name, ns, host string) *networkingv1.Ingress {
		return &networkingv1.Ingress{
			ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns},
			Spec: networkingv1.IngressSpec{
				Rules: []networkingv1.IngressRule{{Host: host}},
			},
		}
	}
	s, _ := newK8sTestServer(t,
		seedIngress("web-gateway", "web", "web.example.com"),
		seedIngress("web-canary", "web", "canary.example.com"),
		seedIngress("db-admin", "db", "db.example.com"),
	)

	tests := []struct {
		desc      string
		path      string
		wantTotal int
		wantNames []string
		wantNS    string // 非空时断言所有条目都属于该命名空间
	}{
		{"集群级列出全部命名空间的 Ingress", "/api/v1/clusters/" + testClusterName + "/ingresses", 3, []string{"web-gateway", "web-canary", "db-admin"}, ""},
		{"命名空间级只返回该命名空间的 Ingress", "/api/v1/clusters/" + testClusterName + "/namespaces/web/ingresses", 2, []string{"web-gateway", "web-canary"}, "web"},
	}
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			w := doRequest(t, s, "GET", tc.path, nil)
			if w.Code != http.StatusOK {
				t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
			}
			var got []networkingv1.Ingress
			if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if len(got) != tc.wantTotal {
				t.Fatalf("len(items) = %d, want %d (body = %s)", len(got), tc.wantTotal, w.Body.String())
			}
			names := map[string]bool{}
			for _, ing := range got {
				names[ing.Name] = true
				if tc.wantNS != "" && ing.Namespace != tc.wantNS {
					t.Errorf("item %s namespace = %q, want %q", ing.Name, ing.Namespace, tc.wantNS)
				}
			}
			for _, want := range tc.wantNames {
				if !names[want] {
					t.Errorf("items missing ingress %q, got %v", want, names)
				}
			}
		})
	}

	// 返回体细节：web-gateway 的 host 应原样保留
	w := doRequest(t, s, "GET", "/api/v1/clusters/"+testClusterName+"/namespaces/web/ingresses", nil)
	var got []networkingv1.Ingress
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	for _, ing := range got {
		if ing.Name == "web-gateway" {
			if len(ing.Spec.Rules) != 1 || ing.Spec.Rules[0].Host != "web.example.com" {
				t.Fatalf("web-gateway rules = %+v, want single host web.example.com", ing.Spec.Rules)
			}
			return
		}
	}
	t.Fatal("web-gateway not found in filtered response")
}

// TestListNetworkPoliciesByNamespace 验证 NetworkPolicy 列表与命名空间过滤。
func TestListNetworkPoliciesByNamespace(t *testing.T) {
	seed := func(name, ns string, matchApp string) *networkingv1.NetworkPolicy {
		return &networkingv1.NetworkPolicy{
			ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns},
			Spec: networkingv1.NetworkPolicySpec{
				PodSelector: metav1.LabelSelector{MatchLabels: map[string]string{"app": matchApp}},
			},
		}
	}
	s, _ := newK8sTestServer(t,
		seed("np-api", "web", "api"),
		seed("np-db", "db", "postgres"),
		seed("np-cache", "cache", "redis"),
	)

	tests := []struct {
		desc      string
		path      string
		wantTotal int
		wantNames []string
		wantNS    string
	}{
		{"集群级列出全部 NetworkPolicy", "/api/v1/clusters/" + testClusterName + "/networkpolicies", 3, []string{"np-api", "np-db", "np-cache"}, ""},
		{"命名空间级过滤 NetworkPolicy", "/api/v1/clusters/" + testClusterName + "/namespaces/db/networkpolicies", 1, []string{"np-db"}, "db"},
	}
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			w := doRequest(t, s, "GET", tc.path, nil)
			if w.Code != http.StatusOK {
				t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
			}
			var got []networkingv1.NetworkPolicy
			if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if len(got) != tc.wantTotal {
				t.Fatalf("len(items) = %d, want %d", len(got), tc.wantTotal)
			}
			names := map[string]bool{}
			for _, np := range got {
				names[np.Name] = true
				if tc.wantNS != "" && np.Namespace != tc.wantNS {
					t.Errorf("item %s namespace = %q, want %q", np.Name, np.Namespace, tc.wantNS)
				}
			}
			for _, want := range tc.wantNames {
				if !names[want] {
					t.Errorf("items missing networkpolicy %q, got %v", want, names)
				}
			}
		})
	}

	// 细节：np-api 的 podSelector 应保留
	w := doRequest(t, s, "GET", "/api/v1/clusters/"+testClusterName+"/namespaces/web/networkpolicies", nil)
	var got []networkingv1.NetworkPolicy
	_ = json.Unmarshal(w.Body.Bytes(), &got)
	if len(got) != 1 || got[0].Spec.PodSelector.MatchLabels["app"] != "api" {
		t.Fatalf("np-api podSelector = %+v, want matchLabels app=api", got[0].Spec.PodSelector)
	}
}

// TestListPVCsByNamespace 验证 PVC 列表与命名空间过滤，并校验 spec 字段往返。
func TestListPVCsByNamespace(t *testing.T) {
	seed := func(name, ns, scName, request string) *corev1.PersistentVolumeClaim {
		return &corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns},
			Spec: corev1.PersistentVolumeClaimSpec{
				StorageClassName: &scName,
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{corev1.ResourceStorage: resource.MustParse(request)},
				},
			},
		}
	}
	s, _ := newK8sTestServer(t,
		seed("pvc-data", "web", "fast", "5Gi"),
		seed("pvc-logs", "db", "slow", "1Gi"),
	)

	tests := []struct {
		desc      string
		path      string
		wantTotal int
		wantNames []string
		wantNS    string
	}{
		{"集群级列出全部 PVC", "/api/v1/clusters/" + testClusterName + "/persistentvolumeclaims", 2, []string{"pvc-data", "pvc-logs"}, ""},
		{"命名空间级只返回 web 的 PVC", "/api/v1/clusters/" + testClusterName + "/namespaces/web/persistentvolumeclaims", 1, []string{"pvc-data"}, "web"},
	}
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			w := doRequest(t, s, "GET", tc.path, nil)
			if w.Code != http.StatusOK {
				t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
			}
			var got []corev1.PersistentVolumeClaim
			if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if len(got) != tc.wantTotal {
				t.Fatalf("len(items) = %d, want %d", len(got), tc.wantTotal)
			}
			names := map[string]bool{}
			for _, pvc := range got {
				names[pvc.Name] = true
				if tc.wantNS != "" && pvc.Namespace != tc.wantNS {
					t.Errorf("item %s namespace = %q, want %q", pvc.Name, pvc.Namespace, tc.wantNS)
				}
			}
			for _, want := range tc.wantNames {
				if !names[want] {
					t.Errorf("items missing pvc %q, got %v", want, names)
				}
			}
		})
	}

	w := doRequest(t, s, "GET", "/api/v1/clusters/"+testClusterName+"/namespaces/web/persistentvolumeclaims", nil)
	var got []corev1.PersistentVolumeClaim
	_ = json.Unmarshal(w.Body.Bytes(), &got)
	if got[0].Spec.StorageClassName == nil || *got[0].Spec.StorageClassName != "fast" {
		t.Fatalf("pvc-data storageClassName = %v, want fast", got[0].Spec.StorageClassName)
	}
	if q := got[0].Spec.Resources.Requests[corev1.ResourceStorage]; q.String() != "5Gi" {
		t.Fatalf("pvc-data request = %s, want 5Gi", q.String())
	}
}

// TestListPersistentVolumes 验证 PV 列表（集群级资源）与 status 往返。
func TestListPersistentVolumes(t *testing.T) {
	seed := func(name, scName string, phase corev1.PersistentVolumePhase) *corev1.PersistentVolume {
		return &corev1.PersistentVolume{
			ObjectMeta: metav1.ObjectMeta{Name: name},
			Spec: corev1.PersistentVolumeSpec{
				StorageClassName: scName,
				Capacity:         corev1.ResourceList{corev1.ResourceStorage: resource.MustParse("10Gi")},
			},
			Status: corev1.PersistentVolumeStatus{Phase: phase},
		}
	}
	s, _ := newK8sTestServer(t,
		seed("pv-bound", "fast", corev1.VolumeBound),
		seed("pv-available", "slow", corev1.VolumeAvailable),
	)

	w := doRequest(t, s, "GET", "/api/v1/clusters/"+testClusterName+"/persistentvolumes", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	var got []corev1.PersistentVolume
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len(items) = %d, want 2", len(got))
	}
	byName := map[string]corev1.PersistentVolume{}
	for _, pv := range got {
		byName[pv.Name] = pv
	}
	if pv, ok := byName["pv-bound"]; !ok {
		t.Fatal("missing pv-bound")
	} else if pv.Status.Phase != corev1.VolumeBound || pv.Spec.StorageClassName != "fast" {
		t.Fatalf("pv-bound = phase %q sc %q, want Bound/fast", pv.Status.Phase, pv.Spec.StorageClassName)
	}
	if pv, ok := byName["pv-available"]; !ok {
		t.Fatal("missing pv-available")
	} else if pv.Status.Phase != corev1.VolumeAvailable {
		t.Fatalf("pv-available phase = %q, want Available", pv.Status.Phase)
	}
}

// TestListStorageClasses 验证 StorageClass 列表与 provisioner 字段。
func TestListStorageClasses(t *testing.T) {
	s, _ := newK8sTestServer(t,
		&storagev1.StorageClass{
			ObjectMeta:  metav1.ObjectMeta{Name: "sc-fast"},
			Provisioner: "ebs.csi.aws.com",
		},
		func() *storagev1.StorageClass {
			retain := corev1.PersistentVolumeReclaimRetain
			return &storagev1.StorageClass{
				ObjectMeta:    metav1.ObjectMeta{Name: "sc-slow"},
				Provisioner:   "kubernetes.io/no-provisioner",
				ReclaimPolicy: &retain,
			}
		}(),
	)

	w := doRequest(t, s, "GET", "/api/v1/clusters/"+testClusterName+"/storageclasses", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	var got []storagev1.StorageClass
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len(items) = %d, want 2", len(got))
	}
	byName := map[string]storagev1.StorageClass{}
	for _, sc := range got {
		byName[sc.Name] = sc
	}
	if sc, ok := byName["sc-fast"]; !ok {
		t.Fatal("missing sc-fast")
	} else if sc.Provisioner != "ebs.csi.aws.com" {
		t.Fatalf("sc-fast provisioner = %q, want ebs.csi.aws.com", sc.Provisioner)
	}
	if sc, ok := byName["sc-slow"]; !ok {
		t.Fatal("missing sc-slow")
	} else if sc.ReclaimPolicy == nil || *sc.ReclaimPolicy != corev1.PersistentVolumeReclaimRetain {
		t.Fatalf("sc-slow reclaimPolicy = %v, want Retain", sc.ReclaimPolicy)
	}
}

// TestNetworkAnalysisStatistics 验证 /api/v1/analysis/network 的统计字段值。
func TestNetworkAnalysisStatistics(t *testing.T) {
	s, _ := newK8sTestServer(t,
		// 2 个 NetworkPolicy 分布在 2 个命名空间
		&networkingv1.NetworkPolicy{ObjectMeta: metav1.ObjectMeta{Name: "np-allow", Namespace: "web"}},
		&networkingv1.NetworkPolicy{ObjectMeta: metav1.ObjectMeta{Name: "np-deny", Namespace: "db"}},
		// ClusterIP + LoadBalancer
		&corev1.Service{
			ObjectMeta: metav1.ObjectMeta{Name: "svc-internal", Namespace: "web"},
			Spec: corev1.ServiceSpec{
				Type:  corev1.ServiceTypeClusterIP,
				Ports: []corev1.ServicePort{{Port: 80}},
			},
		},
		&corev1.Service{
			ObjectMeta: metav1.ObjectMeta{Name: "svc-exposed", Namespace: "db"},
			Spec: corev1.ServiceSpec{
				Type:  corev1.ServiceTypeLoadBalancer,
				Ports: []corev1.ServicePort{{Port: 443}},
			},
		},
		// 2 个 Ingress，其中一个带 host
		&networkingv1.Ingress{
			ObjectMeta: metav1.ObjectMeta{Name: "web-ingress", Namespace: "web"},
			Spec:       networkingv1.IngressSpec{Rules: []networkingv1.IngressRule{{Host: "app.example.com"}}},
		},
		&networkingv1.Ingress{
			ObjectMeta: metav1.ObjectMeta{Name: "db-ingress", Namespace: "db"},
			Spec:       networkingv1.IngressSpec{Rules: []networkingv1.IngressRule{{}}},
		},
	)

	w := doRequest(t, s, "GET", "/api/v1/analysis/network", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	var got networkanalysis.NetworkAnalysis
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got.TotalNetworkPolicies != 2 {
		t.Errorf("totalNetworkPolicies = %d, want 2", got.TotalNetworkPolicies)
	}
	if got.TotalServices != 2 {
		t.Errorf("totalServices = %d, want 2", got.TotalServices)
	}
	if got.TotalIngresses != 2 {
		t.Errorf("totalIngresses = %d, want 2", got.TotalIngresses)
	}
	if names := got.PoliciesByNamespace["web"]; len(names) != 1 || names[0] != "np-allow" {
		t.Errorf("policiesByNamespace[web] = %v, want [np-allow]", names)
	}
	if names := got.PoliciesByNamespace["db"]; len(names) != 1 || names[0] != "np-deny" {
		t.Errorf("policiesByNamespace[db] = %v, want [np-deny]", names)
	}
	if got.ServicesByType["ClusterIP"] != 1 {
		t.Errorf("servicesByType[ClusterIP] = %d, want 1", got.ServicesByType["ClusterIP"])
	}
	if got.ServicesByType["LoadBalancer"] != 1 {
		t.Errorf("servicesByType[LoadBalancer] = %d, want 1", got.ServicesByType["LoadBalancer"])
	}
	if len(got.ExposedServices) != 1 {
		t.Fatalf("len(exposedServices) = %d, want 1", len(got.ExposedServices))
	}
	exposed := got.ExposedServices[0]
	if exposed.Name != "svc-exposed" || exposed.Namespace != "db" || exposed.Type != "LoadBalancer" {
		t.Errorf("exposedServices[0] = %+v, want svc-exposed/db/LoadBalancer", exposed)
	}
	if names := got.IngressesByHost["app.example.com"]; len(names) != 1 || names[0] != "web-ingress" {
		t.Errorf("ingressesByHost[app.example.com] = %v, want [web-ingress]", names)
	}
	if names := got.IngressesByHost["*"]; len(names) != 1 || names[0] != "db-ingress" {
		t.Errorf("ingressesByHost[*] = %v, want [db-ingress]（无 host 的规则回退为 *）", names)
	}
}

// TestStorageAnalysisStatistics 验证 /api/v1/analysis/storage 的统计字段值。
func TestStorageAnalysisStatistics(t *testing.T) {
	s, _ := newK8sTestServer(t,
		&corev1.PersistentVolume{
			ObjectMeta: metav1.ObjectMeta{Name: "pv-bound"},
			Spec: corev1.PersistentVolumeSpec{
				StorageClassName: "fast",
				Capacity:         corev1.ResourceList{corev1.ResourceStorage: resource.MustParse("10Gi")},
			},
			Status: corev1.PersistentVolumeStatus{Phase: corev1.VolumeBound},
		},
		&corev1.PersistentVolume{
			ObjectMeta: metav1.ObjectMeta{Name: "pv-available"},
			Spec: corev1.PersistentVolumeSpec{
				StorageClassName: "slow",
				Capacity:         corev1.ResourceList{corev1.ResourceStorage: resource.MustParse("5Gi")},
			},
			Status: corev1.PersistentVolumeStatus{Phase: corev1.VolumeAvailable},
		},
		pvcForAnalysis("pvc-one", "web", "fast", "2Gi"),
		pvcForAnalysis("pvc-two", "db", "slow", "1Gi"),
		&storagev1.StorageClass{ObjectMeta: metav1.ObjectMeta{Name: "sc-fast"}, Provisioner: "ebs.csi.aws.com"},
		&storagev1.StorageClass{ObjectMeta: metav1.ObjectMeta{Name: "sc-slow"}, Provisioner: "kubernetes.io/no-provisioner"},
	)

	w := doRequest(t, s, "GET", "/api/v1/analysis/storage", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	var got storageanalysis.StorageAnalysis
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got.TotalPVs != 2 {
		t.Errorf("totalPVs = %d, want 2", got.TotalPVs)
	}
	if got.TotalPVCs != 2 {
		t.Errorf("totalPVCs = %d, want 2", got.TotalPVCs)
	}
	if got.TotalStorageClasses != 2 {
		t.Errorf("totalStorageClasses = %d, want 2", got.TotalStorageClasses)
	}
	if got.PVByStatus["Bound"] != 1 || got.PVByStatus["Available"] != 1 {
		t.Errorf("pvByStatus = %v, want Bound=1 Available=1", got.PVByStatus)
	}
	if got.PVCByStatus["Bound"] != 2 {
		t.Errorf("pvcByStatus = %v, want Bound=2", got.PVCByStatus)
	}
	if got.PVByStorageClass["fast"] != 1 || got.PVByStorageClass["slow"] != 1 {
		t.Errorf("pvByStorageClass = %v, want fast=1 slow=1", got.PVByStorageClass)
	}
	if got.SCByProvisioner["ebs.csi.aws.com"] != 1 || got.SCByProvisioner["kubernetes.io/no-provisioner"] != 1 {
		t.Errorf("scByProvisioner = %v, want 各 1", got.SCByProvisioner)
	}
	// 容量口径：总 15Gi，已用（PVC 请求量）3Gi，可用 12Gi
	if got.StorageCapacity.TotalBytes != 15*1024*1024*1024 {
		t.Errorf("storageCapacity.totalBytes = %d, want %d", got.StorageCapacity.TotalBytes, 15*1024*1024*1024)
	}
	if got.StorageCapacity.UsedBytes != 3*1024*1024*1024 {
		t.Errorf("storageCapacity.usedBytes = %d, want %d", got.StorageCapacity.UsedBytes, 3*1024*1024*1024)
	}
	if got.StorageCapacity.AvailableBytes != 12*1024*1024*1024 {
		t.Errorf("storageCapacity.availableBytes = %d, want %d", got.StorageCapacity.AvailableBytes, 12*1024*1024*1024)
	}
}

func pvcForAnalysis(name, ns, scName, request string) *corev1.PersistentVolumeClaim {
	return &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns},
		Spec: corev1.PersistentVolumeClaimSpec{
			StorageClassName: &scName,
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{corev1.ResourceStorage: resource.MustParse(request)},
			},
		},
		Status: corev1.PersistentVolumeClaimStatus{Phase: corev1.ClaimBound},
	}
}

// TestListUnifiedResourcesViaKind 验证统一资源接口：kind 归一化、命名空间过滤与
// 非法 kind 的 400 分支。
func TestListUnifiedResourcesViaKind(t *testing.T) {
	s, _ := newK8sTestServer(t,
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: "pod-api", Namespace: "team-a"},
			Status:     corev1.PodStatus{Phase: corev1.PodRunning},
		},
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: "pod-web", Namespace: "team-b"},
			Status:     corev1.PodStatus{Phase: corev1.PodPending},
		},
	)

	tests := []struct {
		desc         string
		path         string
		wantCode     int
		wantTotal    int
		wantKind     string
		wantNames    []string
		wantNS       string
		wantErrParts string // 非空时为 400 分支，断言 error 字段包含该文案
	}{
		{"复数 kind 列出全部 Pod", "/api/v1/clusters/" + testClusterName + "/resources/pods", 200, 2, "pods", []string{"pod-api", "pod-web"}, "", ""},
		{"单数 kind 归一化为复数", "/api/v1/clusters/" + testClusterName + "/resources/pod", 200, 2, "pods", []string{"pod-api", "pod-web"}, "", ""},
		{"命名空间级过滤 Pod", "/api/v1/clusters/" + testClusterName + "/namespaces/team-a/resources/pods", 200, 1, "pods", []string{"pod-api"}, "team-a", ""},
		{"非法 kind 返回 400", "/api/v1/clusters/" + testClusterName + "/resources/widgets", 400, 0, "", nil, "", "unsupported resource kind: widgets"},
	}
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			w := doRequest(t, s, "GET", tc.path, nil)
			if w.Code != tc.wantCode {
				t.Fatalf("status = %d, want %d, body = %s", w.Code, tc.wantCode, w.Body.String())
			}
			if tc.wantErrParts != "" {
				var errResp map[string]string
				if err := json.Unmarshal(w.Body.Bytes(), &errResp); err != nil {
					t.Fatalf("unmarshal error body: %v", err)
				}
				if errResp["error"] != tc.wantErrParts {
					t.Fatalf("error = %q, want %q", errResp["error"], tc.wantErrParts)
				}
				return
			}
			var got UnifiedResourceList
			if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if got.Total != tc.wantTotal {
				t.Errorf("total = %d, want %d", got.Total, tc.wantTotal)
			}
			if got.ResourceKind != tc.wantKind {
				t.Errorf("resourceKind = %q, want %q", got.ResourceKind, tc.wantKind)
			}
			if len(got.Items) != tc.wantTotal {
				t.Fatalf("len(items) = %d, want %d", len(got.Items), tc.wantTotal)
			}
			names := map[string]bool{}
			for _, item := range got.Items {
				names[item.Name] = true
				if tc.wantNS != "" && item.Namespace != tc.wantNS {
					t.Errorf("item %s namespace = %q, want %q", item.Name, item.Namespace, tc.wantNS)
				}
			}
			for _, want := range tc.wantNames {
				if !names[want] {
					t.Errorf("items missing %q, got %v", want, names)
				}
			}
		})
	}

	// 细节：pod-api 的 kind/status 应来自 Pod 对象本身
	w := doRequest(t, s, "GET", "/api/v1/clusters/"+testClusterName+"/namespaces/team-a/resources/pods", nil)
	var got UnifiedResourceList
	_ = json.Unmarshal(w.Body.Bytes(), &got)
	if got.Items[0].Kind != "Pod" || got.Items[0].Status != "Running" {
		t.Fatalf("item = %+v, want kind Pod status Running", got.Items[0])
	}
}

// TestMonitorStatusEmptyHistory 验证监控状态接口在无指标历史时的返回。
func TestMonitorStatusEmptyHistory(t *testing.T) {
	s, _ := newK8sTestServer(t)

	tests := []struct {
		desc           string
		cluster        string
		wantActive     bool
		wantDataPoints int
	}{
		{"已注册集群但无采集历史", testClusterName, false, 0},
		{"未注册集群同样返回空历史", "ghost-cluster", false, 0},
	}
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			w := doRequest(t, s, "GET", "/api/v1/clusters/"+tc.cluster+"/monitor/status", nil)
			if w.Code != http.StatusOK {
				t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
			}
			var got map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if got["cluster"] != tc.cluster {
				t.Errorf("cluster = %v, want %q", got["cluster"], tc.cluster)
			}
			if got["active"] != tc.wantActive {
				t.Errorf("active = %v, want %v", got["active"], tc.wantActive)
			}
			if int(got["dataPoints"].(float64)) != tc.wantDataPoints {
				t.Errorf("dataPoints = %v, want %d", got["dataPoints"], tc.wantDataPoints)
			}
		})
	}
}
