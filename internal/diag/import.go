package diag

import (
	_ "github.com/kudig-io/klaw/internal/diag/analyzer/kernel"
	_ "github.com/kudig-io/klaw/internal/diag/analyzer/kubernetes"
	_ "github.com/kudig-io/klaw/internal/diag/analyzer/log"
	_ "github.com/kudig-io/klaw/internal/diag/analyzer/network"
	_ "github.com/kudig-io/klaw/internal/diag/analyzer/process"
	_ "github.com/kudig-io/klaw/internal/diag/analyzer/runtime"
	_ "github.com/kudig-io/klaw/internal/diag/analyzer/security"
	_ "github.com/kudig-io/klaw/internal/diag/analyzer/servicemesh"
	_ "github.com/kudig-io/klaw/internal/diag/analyzer/system"
	_ "github.com/kudig-io/klaw/internal/diag/ebpf/analyzer"
)
