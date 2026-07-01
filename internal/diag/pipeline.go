package diag

import (
	"context"

	"github.com/kudig-io/klaw/internal/diag/analyzer"
	"github.com/kudig-io/klaw/internal/diag/collector"
	"github.com/kudig-io/klaw/internal/diag/collector/online"
	"github.com/kudig-io/klaw/internal/diag/types"
)

type DiagnosisRequest struct {
	Kubeconfig       string   `json:"kubeconfig,omitempty"`
	Context          string   `json:"context,omitempty"`
	NodeName         string   `json:"nodeName,omitempty"`
	Namespace        string   `json:"namespace,omitempty"`
	Analyzers        []string `json:"analyzers,omitempty"`
	ExcludeAnalyzers []string `json:"excludeAnalyzers,omitempty"`
}

type DiagnosisResult struct {
	Data    *types.DiagnosticData `json:"data"`
	Results []analyzer.Result     `json:"results"`
	Issues  []types.Issue         `json:"issues"`
}

func RunOnlineDiagnostics(ctx context.Context, req DiagnosisRequest) (*DiagnosisResult, error) {
	c := online.NewCollector()
	cfg := &collector.Config{
		Kubeconfig: req.Kubeconfig,
		Context:    req.Context,
		NodeName:   req.NodeName,
		Namespace:  req.Namespace,
	}
	if err := c.Validate(cfg); err != nil {
		return nil, err
	}
	data, err := c.Collect(ctx, cfg)
	if err != nil {
		return nil, err
	}

	var results []analyzer.Result
	if len(req.Analyzers) > 0 {
		results, err = analyzer.DefaultRegistry.ExecuteByNames(ctx, req.Analyzers, data)
	} else {
		results, err = analyzer.DefaultRegistry.ExecuteAll(ctx, data)
	}
	if err != nil {
		return nil, err
	}

	issues := analyzer.CollectIssues(results)
	if len(req.ExcludeAnalyzers) > 0 {
		exclude := make(map[string]bool, len(req.ExcludeAnalyzers))
		for _, n := range req.ExcludeAnalyzers {
			exclude[n] = true
		}
		filtered := issues[:0]
		for _, iss := range issues {
			if !exclude[iss.AnalyzerName] {
				filtered = append(filtered, iss)
			}
		}
		issues = filtered
	}

	return &DiagnosisResult{
		Data:    data,
		Results: results,
		Issues:  issues,
	}, nil
}

func RegisteredAnalyzerCount() int {
	return len(analyzer.DefaultRegistry.List())
}
