// Package v1 contains API Schema definitions for the kudig v1 API group
package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ClusterDiagnosticSpec defines the desired state of ClusterDiagnostic
type ClusterDiagnosticSpec struct {
	// Mode specifies the diagnostic mode: online or offline
	// +kubebuilder:validation:Enum=online;offline
	Mode string `json:"mode,omitempty"`

	// Analyzers is a list of analyzer names to run. If empty, all analyzers will run.
	Analyzers []string `json:"analyzers,omitempty"`

	// ExcludeAnalyzers is a list of analyzer names to exclude
	ExcludeAnalyzers []string `json:"excludeAnalyzers,omitempty"`

	// OutputFormat specifies the output format: text, json, or html
	// +kubebuilder:validation:Enum=text;json;html
	OutputFormat string `json:"outputFormat,omitempty"`

	// Schedule specifies a cron schedule for recurring diagnostics
	Schedule string `json:"schedule,omitempty"`

	// Notify specifies notification settings
	Notify *NotificationConfig `json:"notify,omitempty"`
}

// NotificationConfig defines notification settings
type NotificationConfig struct {
	// Enabled specifies whether notifications are enabled
	Enabled bool `json:"enabled,omitempty"`

	// WebhookURL is the URL to send notifications to
	WebhookURL string `json:"webhookURL,omitempty"`

	// MinSeverity is the minimum severity level to trigger notifications
	// +kubebuilder:validation:Enum=info;warning;critical
	MinSeverity string `json:"minSeverity,omitempty"`
}

// ClusterDiagnosticStatus defines the observed state of ClusterDiagnostic
type ClusterDiagnosticStatus struct {
	// Phase represents the current phase of the diagnostic
	// +kubebuilder:validation:Enum=Pending;Running;Completed;Failed
	Phase string `json:"phase,omitempty"`

	// StartTime is when the diagnostic started
	StartTime *metav1.Time `json:"startTime,omitempty"`

	// CompletionTime is when the diagnostic completed
	CompletionTime *metav1.Time `json:"completionTime,omitempty"`

	// Summary provides a summary of the diagnostic results
	Summary DiagnosticSummary `json:"summary,omitempty"`

	// Conditions represents the latest available observations of the diagnostic state
	Conditions []DiagnosticCondition `json:"conditions,omitempty"`

	// ReportLocation is where the full report can be found
	ReportLocation string `json:"reportLocation,omitempty"`
}

// DiagnosticSummary provides a summary of diagnostic results
type DiagnosticSummary struct {
	// Total is the total number of issues found
	Total int `json:"total,omitempty"`

	// Critical is the number of critical issues
	Critical int `json:"critical,omitempty"`

	// Warning is the number of warning issues
	Warning int `json:"warning,omitempty"`

	// Info is the number of info issues
	Info int `json:"info,omitempty"`

	// AnalyzersRun is the number of analyzers that ran
	AnalyzersRun int `json:"analyzersRun,omitempty"`
}

// DiagnosticCondition describes the state of a diagnostic at a certain point
type DiagnosticCondition struct {
	// Type of diagnostic condition
	Type string `json:"type"`

	// Status of the condition, one of True, False, Unknown
	Status string `json:"status"`

	// Last time the condition transitioned from one status to another
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`

	// Reason contains a programmatic identifier indicating the reason for the condition's last transition
	Reason string `json:"reason,omitempty"`

	// Message contains a human readable message indicating details about the transition
	Message string `json:"message,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster,shortName=cdiag

// ClusterDiagnostic is the Schema for the clusterdiagnostics API
type ClusterDiagnostic struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterDiagnosticSpec   `json:"spec,omitempty"`
	Status ClusterDiagnosticStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ClusterDiagnosticList contains a list of ClusterDiagnostic
type ClusterDiagnosticList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterDiagnostic `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClusterDiagnostic{}, &ClusterDiagnosticList{})
}
