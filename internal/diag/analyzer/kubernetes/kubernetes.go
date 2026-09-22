// Package kubernetes provides Kubernetes component analyzers
package kubernetes

import (
	"github.com/kudig-io/klaw/internal/diag/analyzer"
)

// init registers all kubernetes analyzers
func init() {
	analyzer.Register(NewPLEGAnalyzer())
	analyzer.Register(NewCNIAnalyzer())
	analyzer.Register(NewCertificateAnalyzer())
	analyzer.Register(NewAPIServerAnalyzer())
	analyzer.Register(NewNodeStatusAnalyzer())
	analyzer.Register(NewImagePullAnalyzer())
	analyzer.Register(NewPodStatusAnalyzer())
	analyzer.Register(NewEventAnalyzer())
}
