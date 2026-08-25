package sos

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadFAQsSeed(t *testing.T) {
	faqs, err := LoadFAQs("")
	if err != nil {
		t.Fatal(err)
	}
	if len(faqs) == 0 {
		t.Fatal("expected seed faqs")
	}
	if faqs[0].ID != "klaw-vs-chatbot" {
		t.Fatalf("expected seed id klaw-vs-chatbot, got %s", faqs[0].ID)
	}
}

func TestLoadFAQsExternalFile(t *testing.T) {
	p := filepath.Join(t.TempDir(), "f.yaml")
	content := "faqs:\n  - id: x\n    question: Q?\n    answer: A.\n"
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	faqs, err := LoadFAQs(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(faqs) != 1 || faqs[0].ID != "x" {
		t.Fatalf("unexpected faqs: %+v", faqs)
	}
}

func TestLoadFAQsErrors(t *testing.T) {
	if _, err := LoadFAQs("/not/exists.yaml"); err == nil {
		t.Fatal("expected error for missing file")
	}
	p := filepath.Join(t.TempDir(), "empty.yaml")
	_ = os.WriteFile(p, []byte("faqs: []\n"), 0o600)
	if _, err := LoadFAQs(p); err == nil {
		t.Fatal("expected error for empty faqs")
	}
}

func TestBuildInstructions(t *testing.T) {
	faqs := []FAQEntry{{ID: "a", Question: "Q1?", Answer: "A1."}}
	out := BuildInstructions("自定义前缀", faqs)
	for _, want := range []string{"自定义前缀", "Q1?", "A1.", "三层", "严格按语料口径"} {
		if !strings.Contains(out, want) {
			t.Fatalf("instructions missing %q", want)
		}
	}
}
