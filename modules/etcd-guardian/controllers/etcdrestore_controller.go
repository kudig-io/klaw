/*
Copyright 2026 EtcdGuardian Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	etcdguardianv1alpha1 "github.com/kudig-io/klaw/modules/etcd-guardian/api/v1alpha1"
)

const (
	restoreFinalizer = "etcdguardian.io/restore-finalizer"
)

// EtcdRestoreReconciler reconciles a EtcdRestore object
type EtcdRestoreReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=etcdguardian.io,resources=etcdrestores,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=etcdguardian.io,resources=etcdrestores/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=etcdguardian.io,resources=etcdrestores/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop
func (r *EtcdRestoreReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("etcdrestore", req.NamespacedName)
	log.Info("Reconciling EtcdRestore")

	// Fetch the EtcdRestore instance
	restore := &etcdguardianv1alpha1.EtcdRestore{}
	err := r.Get(ctx, req.NamespacedName, restore)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("EtcdRestore resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get EtcdRestore")
		return ctrl.Result{}, err
	}

	// Add finalizer if it doesn't exist
	if !controllerutil.ContainsFinalizer(restore, restoreFinalizer) {
		controllerutil.AddFinalizer(restore, restoreFinalizer)
		if err := r.Update(ctx, restore); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Handle deletion
	if !restore.ObjectMeta.DeletionTimestamp.IsZero() {
		return r.handleDeletion(ctx, restore)
	}

	// Check if restore is already completed or failed
	if restore.Status.Phase == etcdguardianv1alpha1.RestorePhaseCompleted ||
		restore.Status.Phase == etcdguardianv1alpha1.RestorePhaseFailed {
		return ctrl.Result{}, nil
	}

	// Initialize status if it's pending
	if restore.Status.Phase == "" {
		restore.Status.Phase = etcdguardianv1alpha1.RestorePhasePending
		restore.Status.StartTime = &metav1.Time{Time: time.Now()}
		if err := r.Status().Update(ctx, restore); err != nil {
			log.Error(err, "Failed to update restore status")
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	// Execute restore phases
	switch restore.Status.Phase {
	case etcdguardianv1alpha1.RestorePhasePending:
		return r.validateConfig(ctx, restore)
	case etcdguardianv1alpha1.RestorePhaseValidating:
		return r.prepareRestore(ctx, restore)
	case etcdguardianv1alpha1.RestorePhasePreparing:
		return r.downloadSnapshot(ctx, restore)
	case etcdguardianv1alpha1.RestorePhaseDownloading:
		return r.performRestore(ctx, restore)
	case etcdguardianv1alpha1.RestorePhaseRestoring:
		return r.completeRestore(ctx, restore)
	}

	return ctrl.Result{}, nil
}

// validateConfig validates the restore configuration
func (r *EtcdRestoreReconciler) validateConfig(ctx context.Context, restore *etcdguardianv1alpha1.EtcdRestore) (ctrl.Result, error) {
	log := r.Log.WithValues("etcdrestore", client.ObjectKeyFromObject(restore))
	log.Info("Validating restore configuration")

	// Validate storage location if specified
	if restore.Spec.StorageLocation.Bucket == "" {
		return r.updateStatusFailed(ctx, restore, "Storage bucket is required")
	}

	// Validate credentials secret exists
	secret := client.ObjectKey{
		Name:      restore.Spec.StorageLocation.CredentialsSecret,
		Namespace: restore.Namespace,
	}
	if err := r.Get(ctx, secret, &corev1.Secret{}); err != nil {
		if errors.IsNotFound(err) {
			return r.updateStatusFailed(ctx, restore, fmt.Sprintf("Credentials secret %s not found", restore.Spec.StorageLocation.CredentialsSecret))
		}
		return ctrl.Result{}, err
	}

	// Validate etcd cluster configuration
	if len(restore.Spec.EtcdCluster.Endpoints) == 0 {
		return r.updateStatusFailed(ctx, restore, "Etcd endpoints are required")
	}

	if restore.Spec.EtcdCluster.DataDir == "" {
		return r.updateStatusFailed(ctx, restore, "Etcd data directory is required")
	}

	// Move to next phase
	restore.Status.Phase = etcdguardianv1alpha1.RestorePhaseValidating
	if err := r.Status().Update(ctx, restore); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{Requeue: true}, nil
}

// prepareRestore prepares the restore environment
func (r *EtcdRestoreReconciler) prepareRestore(ctx context.Context, restore *etcdguardianv1alpha1.EtcdRestore) (ctrl.Result, error) {
	log := r.Log.WithValues("etcdrestore", client.ObjectKeyFromObject(restore))
	log.Info("Preparing restore")

	// TODO: Execute pre-restore hooks if defined

	restore.Status.Phase = etcdguardianv1alpha1.RestorePhasePreparing
	if err := r.Status().Update(ctx, restore); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{Requeue: true}, nil
}

// downloadSnapshot downloads the snapshot from storage
func (r *EtcdRestoreReconciler) downloadSnapshot(ctx context.Context, restore *etcdguardianv1alpha1.EtcdRestore) (ctrl.Result, error) {
	log := r.Log.WithValues("etcdrestore", client.ObjectKeyFromObject(restore))
	log.Info("Downloading snapshot")

	// TODO: Implement snapshot download logic

	restore.Status.Phase = etcdguardianv1alpha1.RestorePhaseDownloading
	if err := r.Status().Update(ctx, restore); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{Requeue: true}, nil
}

// performRestore performs the etcd restore
func (r *EtcdRestoreReconciler) performRestore(ctx context.Context, restore *etcdguardianv1alpha1.EtcdRestore) (ctrl.Result, error) {
	log := r.Log.WithValues("etcdrestore", client.ObjectKeyFromObject(restore))
	log.Info("Performing etcd restore")

	// TODO: Implement actual restore logic

	restore.Status.Phase = etcdguardianv1alpha1.RestorePhaseRestoring
	if err := r.Status().Update(ctx, restore); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{Requeue: true}, nil
}

// completeRestore marks the restore as completed
func (r *EtcdRestoreReconciler) completeRestore(ctx context.Context, restore *etcdguardianv1alpha1.EtcdRestore) (ctrl.Result, error) {
	log := r.Log.WithValues("etcdrestore", client.ObjectKeyFromObject(restore))
	log.Info("Restore completed successfully")

	restore.Status.Phase = etcdguardianv1alpha1.RestorePhaseCompleted
	restore.Status.CompletionTime = &metav1.Time{Time: time.Now()}
	restore.Status.Message = "Restore completed successfully"

	if err := r.Status().Update(ctx, restore); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// updateStatusFailed updates the restore status to failed
func (r *EtcdRestoreReconciler) updateStatusFailed(ctx context.Context, restore *etcdguardianv1alpha1.EtcdRestore, message string) (ctrl.Result, error) {
	restore.Status.Phase = etcdguardianv1alpha1.RestorePhaseFailed
	restore.Status.CompletionTime = &metav1.Time{Time: time.Now()}
	restore.Status.Message = message

	if err := r.Status().Update(ctx, restore); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, fmt.Errorf("%s", message)
}

// handleDeletion handles the deletion of a restore
func (r *EtcdRestoreReconciler) handleDeletion(ctx context.Context, restore *etcdguardianv1alpha1.EtcdRestore) (ctrl.Result, error) {
	log := r.Log.WithValues("etcdrestore", client.ObjectKeyFromObject(restore))

	if controllerutil.ContainsFinalizer(restore, restoreFinalizer) {
		// TODO: Clean up resources if needed

		log.Info("Removing finalizer")
		controllerutil.RemoveFinalizer(restore, restoreFinalizer)
		if err := r.Update(ctx, restore); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *EtcdRestoreReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&etcdguardianv1alpha1.EtcdRestore{}).
		Complete(r)
}
