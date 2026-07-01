package controllers

import (
	"strings"

	"context"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	kudigv1 "github.com/kudig-io/klaw/operator/api/v1"
)

// NodeDiagnosticReconciler reconciles a NodeDiagnostic object
type NodeDiagnosticReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=kudig.io,resources=nodediagnostics,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kudig.io,resources=nodediagnostics/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kudig.io,resources=nodediagnostics/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=daemonsets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=nodes,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete

// Reconcile is the main reconciliation loop
func (r *NodeDiagnosticReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	diagnostic := &kudigv1.NodeDiagnostic{}
	if err := r.Get(ctx, req.NamespacedName, diagnostic); err != nil {
		if errors.IsNotFound(err) {
			log.Info("NodeDiagnostic resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get NodeDiagnostic")
		return ctrl.Result{}, err
	}

	if diagnostic.Status.Phase == "Completed" || diagnostic.Status.Phase == "Failed" {
		return ctrl.Result{}, nil
	}

	targetNodes, err := r.getTargetNodes(ctx, diagnostic)
	if err != nil {
		log.Error(err, "Failed to get target nodes")
		return ctrl.Result{}, err
	}

	if len(targetNodes) == 0 {
		log.Info("No target nodes found for diagnostic")
		diagnostic.Status.Phase = "Failed"
		now := metav1.Now()
		diagnostic.Status.CompletionTime = &now
		condition := kudigv1.DiagnosticCondition{
			Type:               "Failed",
			Status:             "True",
			LastTransitionTime: now,
			Reason:             "NoNodes",
			Message:            "No target nodes found",
		}
		diagnostic.Status.Conditions = append(diagnostic.Status.Conditions, condition)
		if err := r.Status().Update(ctx, diagnostic); err != nil {
			log.Error(err, "Failed to update NodeDiagnostic status")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	if diagnostic.Status.Phase == "" {
		diagnostic.Status.Phase = "Running"
		diagnostic.Status.StartTime = &metav1.Time{Time: time.Now()}
		diagnostic.Status.NodeResults = make([]kudigv1.NodeResult, 0, len(targetNodes))
		for _, node := range targetNodes {
			diagnostic.Status.NodeResults = append(diagnostic.Status.NodeResults, kudigv1.NodeResult{
				NodeName: node,
				Phase:    "Pending",
			})
		}
		if err := r.Status().Update(ctx, diagnostic); err != nil {
			log.Error(err, "Failed to update NodeDiagnostic status")
			return ctrl.Result{}, err
		}
	}

	dsName := fmt.Sprintf("klaw-node-%s", diagnostic.Name)
	found := &appsv1.DaemonSet{}
	err = r.Get(ctx, types.NamespacedName{Name: dsName, Namespace: "klaw-system"}, found)
	if err != nil && errors.IsNotFound(err) {
		ds := r.daemonSetForNodeDiagnostic(diagnostic, dsName, targetNodes)
		log.Info("Creating DaemonSet for node diagnostics", "DaemonSet.Namespace", ds.Namespace, "DaemonSet.Name", ds.Name)
		if err := r.Create(ctx, ds); err != nil {
			log.Error(err, "Failed to create DaemonSet")
			return ctrl.Result{}, err
		}
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	} else if err != nil {
		log.Error(err, "Failed to get DaemonSet")
		return ctrl.Result{}, err
	}

	// Check DaemonSet status
	desired := found.Status.DesiredNumberScheduled
	completed := found.Status.NumberReady
	failed := found.Status.NumberUnavailable

	if desired == 0 {
		return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
	}

	// Collect per-node results from ConfigMaps
	for i := range diagnostic.Status.NodeResults {
		nodeName := diagnostic.Status.NodeResults[i].NodeName
		if diagnostic.Status.NodeResults[i].Phase == "Completed" {
			continue
		}
		cmName := fmt.Sprintf("klaw-result-%s-%s", diagnostic.Name, nodeName)
		cm := &corev1.ConfigMap{}
		if err := r.Get(ctx, types.NamespacedName{Name: cmName, Namespace: "klaw-system"}, cm); err == nil {
			if summary, ok := cm.Data["summary"]; ok {
				diagnostic.Status.NodeResults[i].Phase = "Completed"
				diagnostic.Status.NodeResults[i].Message = summary
			}
		}
	}

	allCompleted := true
	for _, nr := range diagnostic.Status.NodeResults {
		if nr.Phase != "Completed" {
			allCompleted = false
			break
		}
	}

	if allCompleted && completed >= desired {
		diagnostic.Status.Phase = "Completed"
		now := metav1.Now()
		diagnostic.Status.CompletionTime = &now
		condition := kudigv1.DiagnosticCondition{
			Type:               "Completed",
			Status:             "True",
			LastTransitionTime: now,
			Reason:             "AllNodesCompleted",
			Message:            fmt.Sprintf("All %d nodes completed diagnostics", desired),
		}
		diagnostic.Status.Conditions = append(diagnostic.Status.Conditions, condition)
	} else if failed > 0 && completed+failed >= desired {
		diagnostic.Status.Phase = "Failed"
		now := metav1.Now()
		diagnostic.Status.CompletionTime = &now
		condition := kudigv1.DiagnosticCondition{
			Type:               "Failed",
			Status:             "True",
			LastTransitionTime: now,
			Reason:             "NodeFailures",
			Message:            fmt.Sprintf("%d of %d nodes failed", failed, desired),
		}
		diagnostic.Status.Conditions = append(diagnostic.Status.Conditions, condition)
	}

	if err := r.Status().Update(ctx, diagnostic); err != nil {
		log.Error(err, "Failed to update NodeDiagnostic status")
		return ctrl.Result{}, err
	}

	if diagnostic.Status.Phase == "Running" {
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}

	return ctrl.Result{}, nil
}

// getTargetNodes returns the list of nodes to diagnose
func (r *NodeDiagnosticReconciler) getTargetNodes(ctx context.Context, diagnostic *kudigv1.NodeDiagnostic) ([]string, error) {
	if len(diagnostic.Spec.NodeNames) > 0 {
		return diagnostic.Spec.NodeNames, nil
	}

	nodeList := &corev1.NodeList{}
	selector := labels.Everything()
	if len(diagnostic.Spec.NodeSelector) > 0 {
		var err error
		selector, err = metav1.LabelSelectorAsSelector(&metav1.LabelSelector{
			MatchLabels: diagnostic.Spec.NodeSelector,
		})
		if err != nil {
			return nil, err
		}
	}

	if err := r.List(ctx, nodeList, &client.ListOptions{LabelSelector: selector}); err != nil {
		return nil, err
	}

	nodes := make([]string, 0, len(nodeList.Items))
	for _, node := range nodeList.Items {
		nodes = append(nodes, node.Name)
	}
	return nodes, nil
}

// daemonSetForNodeDiagnostic creates a DaemonSet for a NodeDiagnostic resource
func (r *NodeDiagnosticReconciler) daemonSetForNodeDiagnostic(diagnostic *kudigv1.NodeDiagnostic, dsName string, nodes []string) *appsv1.DaemonSet {
	lbls := map[string]string{
		"app": "klaw",
		"klaw.io/diagnostic": diagnostic.Name,
		"klaw.io/type":       "node",
		"klaw.io/created-by": "operator",
	}

	args := []string{"diag", "--json"}
	if len(diagnostic.Spec.Analyzers) > 0 {
		args = append(args, "--analyzer", strings.Join(diagnostic.Spec.Analyzers, ","))
	}

	nodeNamesEnv := ""
	for i, node := range nodes {
		if i > 0 {
			nodeNamesEnv += ","
		}
		nodeNamesEnv += node
	}

	podSpec := corev1.PodSpec{
		HostNetwork:   true,
		HostPID:       true,
		RestartPolicy: corev1.RestartPolicyOnFailure,
		Containers: []corev1.Container{{
			Name:    "klaw",
			Image:   "kudig-io/klaw:latest",
			Command: []string{"/usr/local/bin/klaw"},
			Args:    args,
			SecurityContext: &corev1.SecurityContext{
				Privileged: ptrBool(true),
			},
			Env: []corev1.EnvVar{
				{Name: "TARGET_NODES", Value: nodeNamesEnv},
				{Name: "KLAW_RESULT_CONFIGMAP", Value: fmt.Sprintf("klaw-result-%s", diagnostic.Name)},
				{
					Name: "NODE_NAME",
					ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{
							FieldPath: "spec.nodeName",
						},
					},
				},
			},
		}},
		ServiceAccountName: "klaw-operator",
		Tolerations: []corev1.Toleration{
			{Operator: corev1.TolerationOpExists},
		},
	}

	// When specific node names are given, use node affinity to limit DaemonSet placement
	if len(diagnostic.Spec.NodeNames) > 0 {
		terms := make([]corev1.NodeSelectorTerm, 0, 1)
		matchExprs := make([]corev1.NodeSelectorRequirement, 0, 1)
		matchExprs = append(matchExprs, corev1.NodeSelectorRequirement{
			Key:      "kubernetes.io/hostname",
			Operator: corev1.NodeSelectorOpIn,
			Values:   nodes,
		})
		terms = append(terms, corev1.NodeSelectorTerm{
			MatchExpressions: matchExprs,
		})
		podSpec.Affinity = &corev1.Affinity{
			NodeAffinity: &corev1.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
					NodeSelectorTerms: terms,
				},
			},
		}
	} else if len(diagnostic.Spec.NodeSelector) > 0 {
		podSpec.NodeSelector = diagnostic.Spec.NodeSelector
	}

	ds := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      dsName,
			Namespace: "klaw-system",
			Labels:    lbls,
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: lbls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: lbls,
				},
				Spec: podSpec,
			},
		},
	}

	ctrl.SetControllerReference(diagnostic, ds, r.Scheme)
	return ds
}

func ptrBool(b bool) *bool { return &b }

// SetupWithManager sets up the controller with the Manager
func (r *NodeDiagnosticReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kudigv1.NodeDiagnostic{}).
		Owns(&appsv1.DaemonSet{}).
		Owns(&batchv1.Job{}).
		Complete(r)
}
