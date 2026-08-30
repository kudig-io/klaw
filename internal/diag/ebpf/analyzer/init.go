package analyzer

import "github.com/kudig-io/klaw/internal/diag/analyzer"

func init() {
	// Register eBPF analyzers
	analyzer.MustRegister(NewTCPAnalyzer())
	analyzer.MustRegister(NewDNSAnalyzer())
	analyzer.MustRegister(NewFileIOAnalyzer())
}
