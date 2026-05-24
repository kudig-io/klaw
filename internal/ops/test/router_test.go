package ops_test

import (
	"testing"

	"github.com/kudig-io/klaw/internal/ops"
)

func TestCommandRouter_ExpandCommand(t *testing.T) {
	router := ops.NewCommandRouter(nil)

	tests := map[string]string{
		"p ls demo default":   "pod list demo default",
		"s desc demo default": "service describe demo default",
		"svc endpoints demo default web": "service endpoints demo default web",
	}

	for input, want := range tests {
		if got := router.ExpandCommand(input); got != want {
			t.Fatalf("ExpandCommand(%q) = %q, want %q", input, got, want)
		}
	}
}
