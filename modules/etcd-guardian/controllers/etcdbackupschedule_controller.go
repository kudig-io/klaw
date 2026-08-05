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
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	etcdguardianv1alpha1 "github.com/kudig-io/klaw/modules/etcd-guardian/api/v1alpha1"
)

const (
	scheduleFinalizer = "etcdguardian.io/schedule-finalizer"
)

// EtcdBackupScheduleReconciler reconciles a EtcdBackupSchedule object
type EtcdBackupScheduleReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=etcdguardian.io,resources=etcdbackupschedules,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=etcdguardian.io,resources=etcdbackupschedules/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=etcdguardian.io,resources=etcdbackupschedules/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop
func (r *EtcdBackupScheduleReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("etcdbackupschedule", req.NamespacedName)
	log.Info("Reconciling EtcdBackupSchedule")

	// Fetch the EtcdBackupSchedule instance
	schedule := &etcdguardianv1alpha1.EtcdBackupSchedule{}
	err := r.Get(ctx, req.NamespacedName, schedule)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("EtcdBackupSchedule resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get EtcdBackupSchedule")
		return ctrl.Result{}, err
	}

	// Add finalizer if it doesn't exist
	if !controllerutil.ContainsFinalizer(schedule, scheduleFinalizer) {
		controllerutil.AddFinalizer(schedule, scheduleFinalizer)
		if err := r.Update(ctx, schedule); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Handle deletion
	if !schedule.ObjectMeta.DeletionTimestamp.IsZero() {
		return r.handleDeletion(ctx, schedule)
	}

	// Validate schedule configuration
	if schedule.Spec.Schedule == "" {
		log.Error(err, "Schedule is required for EtcdBackupSchedule")
		return ctrl.Result{}, fmt.Errorf("schedule is required for EtcdBackupSchedule")
	}

	// TODO: Implement cron scheduling logic
	// For now, just create a backup immediately
	if err := r.createBackupFromSchedule(ctx, schedule); err != nil {
		log.Error(err, "Failed to create backup from schedule")
		return ctrl.Result{}, err
	}

	// Update status
	schedule.Status.LastBackupTime = &metav1.Time{Time: time.Now()}
	if err := r.Status().Update(ctx, schedule); err != nil {
		log.Error(err, "Failed to update schedule status")
		return ctrl.Result{}, err
	}

	// Requeue after some time (simulate cron scheduling)
	return ctrl.Result{RequeueAfter: 1 * time.Hour}, nil
}

// createBackupFromSchedule creates a backup based on the schedule
func (r *EtcdBackupScheduleReconciler) createBackupFromSchedule(ctx context.Context, schedule *etcdguardianv1alpha1.EtcdBackupSchedule) error {
	log := r.Log.WithValues("etcdbackupschedule", schedule.Name)
	log.Info("Creating backup from schedule")

	// Create backup name with timestamp
	timestamp := time.Now().Format("20060102-150405")
	backupName := fmt.Sprintf("%s-%s", schedule.Name, timestamp)

	// Create EtcdBackup resource
	backup := &etcdguardianv1alpha1.EtcdBackup{
		ObjectMeta: metav1.ObjectMeta{
			Name:      backupName,
			Namespace: schedule.Namespace,
			Labels: map[string]string{
				"etcdguardian.io/schedule": schedule.Name,
			},
		},
		Spec: schedule.Spec.BackupTemplate.Spec,
	}

	// Set owner reference
	if err := controllerutil.SetOwnerReference(schedule, backup, r.Scheme); err != nil {
		return fmt.Errorf("failed to set owner reference: %w", err)
	}

	// Create backup resource
	if err := r.Create(ctx, backup); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	log.Info("Backup created successfully", "backup", backupName)
	return nil
}

// handleDeletion handles the deletion of a schedule
func (r *EtcdBackupScheduleReconciler) handleDeletion(ctx context.Context, schedule *etcdguardianv1alpha1.EtcdBackupSchedule) (ctrl.Result, error) {
	log := r.Log.WithValues("etcdbackupschedule", schedule.Name)

	if controllerutil.ContainsFinalizer(schedule, scheduleFinalizer) {
		// TODO: Clean up associated backups if needed

		log.Info("Removing finalizer")
		controllerutil.RemoveFinalizer(schedule, scheduleFinalizer)
		if err := r.Update(ctx, schedule); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *EtcdBackupScheduleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&etcdguardianv1alpha1.EtcdBackupSchedule{}).
		Complete(r)
}
