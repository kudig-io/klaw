package controllers

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	kudigv1 "github.com/kudig-io/klaw/operator/api/v1"
)

// DiagnosticScheduleReconciler reconciles a DiagnosticSchedule object
type DiagnosticScheduleReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=kudig.io,resources=diagnosticschedules,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kudig.io,resources=diagnosticschedules/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kudig.io,resources=diagnosticschedules/finalizers,verbs=update
//+kubebuilder:rbac:groups=kudig.io,resources=clusterdiagnostics,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kudig.io,resources=nodediagnostics,verbs=get;list;watch;create;update;patch;delete

// Reconcile is the main reconciliation loop
func (r *DiagnosticScheduleReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Fetch the DiagnosticSchedule instance
	schedule := &kudigv1.DiagnosticSchedule{}
	if err := r.Get(ctx, req.NamespacedName, schedule); err != nil {
		log.Error(err, "Failed to get DiagnosticSchedule")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Check if suspended
	if schedule.Spec.Suspend {
		log.Info("Schedule is suspended, skipping")
		return ctrl.Result{}, nil
	}

	// Check if it's time to run
	now := time.Now()
	lastRun := schedule.Status.LastRunTime
	if lastRun != nil {
		// Simple cron-like scheduling (every X duration)
		// In production, use a proper cron library like robfig/cron
		nextRun := lastRun.Add(r.parseSchedule(schedule.Spec.Schedule))
		if now.Before(nextRun) {
			// Not time yet, requeue for later
			return ctrl.Result{RequeueAfter: nextRun.Sub(now)}, nil
		}
	}

	// Create diagnostic based on type
	switch schedule.Spec.Type {
	case "cluster":
		if err := r.createClusterDiagnostic(ctx, schedule); err != nil {
			log.Error(err, "Failed to create ClusterDiagnostic")
			return ctrl.Result{}, err
		}
	case "node":
		if err := r.createNodeDiagnostic(ctx, schedule); err != nil {
			log.Error(err, "Failed to create NodeDiagnostic")
			return ctrl.Result{}, err
		}
	default:
		log.Info("Unknown schedule type", "type", schedule.Spec.Type)
		return ctrl.Result{}, nil
	}

	// Update status
	nowTime := metav1.NewTime(now)
	schedule.Status.LastRunTime = &nowTime
	schedule.Status.LastSuccessfulRun = &nowTime

	if err := r.Status().Update(ctx, schedule); err != nil {
		log.Error(err, "Failed to update DiagnosticSchedule status")
		return ctrl.Result{}, err
	}

	// Requeue for next run
	return ctrl.Result{RequeueAfter: r.parseSchedule(schedule.Spec.Schedule)}, nil
}

// createClusterDiagnostic creates a ClusterDiagnostic from the schedule template
func (r *DiagnosticScheduleReconciler) createClusterDiagnostic(ctx context.Context, schedule *kudigv1.DiagnosticSchedule) error {
	template := schedule.Spec.ClusterDiagnosticTemplate
	if template == nil {
		return fmt.Errorf("clusterDiagnosticTemplate is nil")
	}

	name := fmt.Sprintf("%s-%d", schedule.Name, time.Now().Unix())
	diagnostic := &kudigv1.ClusterDiagnostic{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   "klaw-system",
			Annotations: map[string]string{"klaw.io/schedule": schedule.Name},
		},
		Spec: kudigv1.ClusterDiagnosticSpec{
			Mode:             template.Mode,
			Analyzers:        template.Analyzers,
			ExcludeAnalyzers: template.ExcludeAnalyzers,
			OutputFormat:     template.OutputFormat,
			Notify:           template.Notify,
		},
	}

	if err := ctrl.SetControllerReference(schedule, diagnostic, r.Scheme); err != nil {
		return err
	}

	return r.Create(ctx, diagnostic)
}

// createNodeDiagnostic creates a NodeDiagnostic from the schedule template
func (r *DiagnosticScheduleReconciler) createNodeDiagnostic(ctx context.Context, schedule *kudigv1.DiagnosticSchedule) error {
	template := schedule.Spec.NodeDiagnosticTemplate
	if template == nil {
		return fmt.Errorf("nodeDiagnosticTemplate is nil")
	}

	name := fmt.Sprintf("%s-%d", schedule.Name, time.Now().Unix())
	diagnostic := &kudigv1.NodeDiagnostic{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   "klaw-system",
			Annotations: map[string]string{"klaw.io/schedule": schedule.Name},
		},
		Spec: kudigv1.NodeDiagnosticSpec{
			Mode:             template.Mode,
			NodeSelector:     template.NodeSelector,
			Analyzers:        template.Analyzers,
			ExcludeAnalyzers: template.ExcludeAnalyzers,
			OutputFormat:     template.OutputFormat,
			Notify:           template.Notify,
		},
	}

	if err := ctrl.SetControllerReference(schedule, diagnostic, r.Scheme); err != nil {
		return err
	}

	return r.Create(ctx, diagnostic)
}

// parseSchedule parses a schedule string into a duration.
// Supports: @hourly, @daily, @weekly, @monthly, @every Nh, @every Nm, standard cron expressions (5 or 6 fields).
// For cron expressions, it calculates the duration until the next scheduled time.
func (r *DiagnosticScheduleReconciler) parseSchedule(schedule string) time.Duration {
	switch schedule {
	case "@hourly":
		return time.Hour
	case "@daily":
		return 24 * time.Hour
	case "@weekly":
		return 7 * 24 * time.Hour
	case "@monthly":
		return 30 * 24 * time.Hour
	case "@yearly", "@annually":
		return 365 * 24 * time.Hour
	}

	if strings.HasPrefix(schedule, "@every ") {
		durStr := strings.TrimPrefix(schedule, "@every ")
		if d, err := time.ParseDuration(durStr); err == nil {
			return d
		}
	}

	// Try standard cron: minute hour dayOfMonth month dayOfWeek
	fields := strings.Fields(schedule)
	if len(fields) == 5 {
		if d := parseCronDuration(fields); d > 0 {
			return d
		}
	}

	return time.Hour
}

// parseCronDuration estimates a duration from a 5-field cron expression
func parseCronDuration(fields []string) time.Duration {
	// Parse hour field to estimate interval
	hourField := fields[1]
	if hourField == "*" {
		// Every hour at a specific minute
		return time.Hour
	}
	if strings.Contains(hourField, "/") {
		parts := strings.SplitN(hourField, "/", 2)
		if hours, err := strconv.Atoi(parts[1]); err == nil && hours > 0 {
			return time.Duration(hours) * time.Hour
		}
	}
	if strings.Contains(hourField, ",") {
		// Multiple specific hours - estimate based on count
		hours := strings.Split(hourField, ",")
		return 24 * time.Hour / time.Duration(len(hours))
	}
	// Specific hour - daily
	return 24 * time.Hour
}

// SetupWithManager sets up the controller with the Manager
func (r *DiagnosticScheduleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kudigv1.DiagnosticSchedule{}).
		Complete(r)
}
