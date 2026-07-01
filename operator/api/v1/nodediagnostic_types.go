package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NodeDiagnosticSpec defines the desired state of NodeDiagnostic
type NodeDiagnosticSpec struct {
	// Mode specifies the diagnostic mode: online or offline
	// +kubebuilder:validation:Enum=online;offline
	Mode string `json:"mode,omitempty"`

	// NodeSelector specifies which nodes to run diagnostics on
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// NodeNames is a list of specific node names to diagnose
	NodeNames []string `json:"nodeNames,omitempty"`

	// Analyzers is a list of analyzer names to run
	Analyzers []string `json:"analyzers,omitempty"`

	// ExcludeAnalyzers is a list of analyzer names to exclude
	ExcludeAnalyzers []string `json:"excludeAnalyzers,omitempty"`

	// OutputFormat specifies the output format
	// +kubebuilder:validation:Enum=text;json;html
	OutputFormat string `json:"outputFormat,omitempty"`

	// Notify specifies notification settings
	Notify *NotificationConfig `json:"notify,omitempty"`
}

// NodeDiagnosticStatus defines the observed state of NodeDiagnostic
type NodeDiagnosticStatus struct {
	// Phase represents the current phase
	// +kubebuilder:validation:Enum=Pending;Running;Completed;Failed
	Phase string `json:"phase,omitempty"`

	// StartTime is when the diagnostic started
	StartTime *metav1.Time `json:"startTime,omitempty"`

	// CompletionTime is when the diagnostic completed
	CompletionTime *metav1.Time `json:"completionTime,omitempty"`

	// NodeResults contains results for each node
	NodeResults []NodeResult `json:"nodeResults,omitempty"`

	// Conditions represents the latest available observations
	Conditions []DiagnosticCondition `json:"conditions,omitempty"`
}

// NodeResult contains the diagnostic result for a single node
type NodeResult struct {
	// NodeName is the name of the node
	NodeName string `json:"nodeName"`

	// Phase is the phase for this node
	Phase string `json:"phase,omitempty"`

	// Summary provides a summary of issues for this node
	Summary DiagnosticSummary `json:"summary,omitempty"`

	// Message contains additional information
	Message string `json:"message,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster,shortName=ndiag

// NodeDiagnostic is the Schema for the nodediagnostics API
type NodeDiagnostic struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NodeDiagnosticSpec   `json:"spec,omitempty"`
	Status NodeDiagnosticStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// NodeDiagnosticList contains a list of NodeDiagnostic
type NodeDiagnosticList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NodeDiagnostic `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NodeDiagnostic{}, &NodeDiagnosticList{})
}
