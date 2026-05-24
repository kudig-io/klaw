package storage

import (
	"path/filepath"
	"testing"
)

func TestStoreJSONLifecycle(t *testing.T) {
	store, err := NewStore(filepath.Join(t.TempDir(), "klaw.db"))
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}
	defer store.Close()

	payload := map[string]interface{}{
		"name":   "demo",
		"active": true,
	}
	if err := store.PutJSON("test", "item", payload); err != nil {
		t.Fatalf("PutJSON() error = %v", err)
	}

	var got map[string]interface{}
	found, err := store.GetJSON("test", "item", &got)
	if err != nil {
		t.Fatalf("GetJSON() error = %v", err)
	}
	if !found {
		t.Fatal("GetJSON() should find stored document")
	}
	if got["name"] != "demo" {
		t.Fatalf("got[name] = %v, want demo", got["name"])
	}

	if err := store.Delete("test", "item"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	found, err = store.GetJSON("test", "item", &got)
	if err != nil {
		t.Fatalf("GetJSON() after delete error = %v", err)
	}
	if found {
		t.Fatal("GetJSON() should not find deleted document")
	}
}
