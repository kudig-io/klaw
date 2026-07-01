package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DiagnosticScheduleSpec defines the desired state of DiagnosticSchedule
type DiagnosticScheduleSpec struct {
	// Schedule is the cron expression for the schedule
	// +kubebuilder:validation:Pattern=`^(@(annually|yearly|monthly|weekly|daily|hourly|reboot))|(@every (\d+(ns|us|µs|ms|s|m|h))+)|((((\d+,)+\d+|(\d+(\/|-)\d+)|\d+|\*) ?){5,7})$`
	Schedule string `json:"schedule"`

	// Suspend specifies whether the schedule is suspended
	Suspend bool `json:"suspend,omitempty"`

	// Type specifies the type of diagnostic: cluster or node
	// +kubebuilder:validation:Enum=cluster;node
	Type string `json:"type"`

	// ClusterDiagnosticTemplate is the template for cluster diagnostics
	ClusterDiagnosticTemplate *ClusterDiagnosticTemplate `json:"clusterDiagnosticTemplate,omitempty"`

	// NodeDiagnosticTemplate is the template for node diagnostics
	NodeDiagnosticTemplate *NodeDiagnosticTemplate `json:"nodeDiagnosticTemplate,omitempty"`

	// HistoryLimit is the number of completed diagnostics to retain
	// +kubebuilder:default=10
	HistoryLimit int32 `json:"historyLimit,omitempty"`
}

// ClusterDiagnosticTemplate is the template for cluster diagnostics
type ClusterDiagnosticTemplate struct {
	// Mode specifies the diagnostic mode
	Mode string `json:"mode,omitempty"`

	// Analyzers is a list of analyzer names to run
	Analyzers []string `json:"analyzers,omitempty"`

	// ExcludeAnalyzers is a list of analyzer names to exclude
	ExcludeAnalyzers []string `json:"excludeAnalyzers,omitempty"`

	// OutputFormat specifies the output format
	OutputFormat string `json:"outputFormat,omitempty"`

	// Notify specifies notification settings
	Notify *NotificationConfig `json:"notify,omitempty"`
}

// NodeDiagnosticTemplate is the template for node diagnostics
type NodeDiagnosticTemplate struct {
	// Mode specifies the diagnostic mode
	Mode string `json:"mode,omitempty"`

	// NodeSelector specifies which nodes to run diagnostics on
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// Analyzers is a list of analyzer names to run
	Analyzers []string `json:"analyzers,omitempty"`

	// ExcludeAnalyzers is a list of analyzer names to exclude
	ExcludeAnalyzers []string `json:"excludeAnalyzers,omitempty"`

	// OutputFormat specifies the output format
	OutputFormat string `json:"outputFormat,omitempty"`

	// Notify specifies notification settings
	Notify *NotificationConfig `json:"notify,omitempty"`
}

// DiagnosticScheduleStatus defines the observed state of DiagnosticSchedule
type DiagnosticScheduleStatus struct {
	// LastRunTime is the last time the schedule was run
	LastRunTime *metav1.Time `json:"lastRunTime,omitempty"`

	// LastSuccessfulRun is the last time the schedule ran successfully
	LastSuccessfulRun *metav1.Time `json:"lastSuccessfulRun,omitempty"`

	// Active is a list of active diagnostic jobs
	Active []corev1.ObjectReference `json:"active,omitempty"`

	// Conditions represents the latest available observations
	Conditions []DiagnosticCondition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster,shortName=dsched

// DiagnosticSchedule is the Schema for the diagnosticschedules API
type DiagnosticSchedule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DiagnosticScheduleSpec   `json:"spec,omitempty"`
	Status DiagnosticScheduleStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DiagnosticScheduleList contains a list of DiagnosticSchedule
type DiagnosticScheduleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DiagnosticSchedule `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DiagnosticSchedule{}, &DiagnosticScheduleList{})
}
