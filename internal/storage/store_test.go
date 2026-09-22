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

// newTestStore 构造每个用例独立的临时库。
func newTestStore(t *testing.T) *Store {
	t.Helper()
	store, err := NewStore(filepath.Join(t.TempDir(), "klaw.db"))
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })
	return store
}

// TestStoreGetJSONMissingKey 验证读取不存在的 key 返回 found=false 且不报错。
func TestStoreGetJSONMissingKey(t *testing.T) {
	store := newTestStore(t)

	var got map[string]interface{}
	found, err := store.GetJSON("ns", "missing", &got)
	if err != nil {
		t.Fatalf("GetJSON() error = %v, want nil", err)
	}
	if found {
		t.Fatal("GetJSON() found = true for missing key, want false")
	}

	// 同命名空间下的其他 key 不应被误读
	if err := store.PutJSON("ns", "other", map[string]int{"n": 1}); err != nil {
		t.Fatalf("PutJSON() error = %v", err)
	}
	found, err = store.GetJSON("ns", "missing", &got)
	if err != nil || found {
		t.Fatalf("GetJSON() = (%v, %v), want (false, nil)", found, err)
	}
}

// TestStorePutJSONOverwrite 验证重复 PutJSON 同一 key 会整体覆盖旧值（含类型变化）。
func TestStorePutJSONOverwrite(t *testing.T) {
	store := newTestStore(t)
	const ns, key = "over", "doc"

	first := map[string]interface{}{"name": "old", "count": float64(1)}
	if err := store.PutJSON(ns, key, first); err != nil {
		t.Fatalf("first PutJSON() error = %v", err)
	}

	second := []string{"alpha", "beta"}
	if err := store.PutJSON(ns, key, second); err != nil {
		t.Fatalf("second PutJSON() error = %v", err)
	}

	var roundtrip interface{}
	found, err := store.GetJSON(ns, key, &roundtrip)
	if err != nil {
		t.Fatalf("GetJSON() error = %v", err)
	}
	if !found {
		t.Fatal("GetJSON() found = false, want true")
	}
	arr, ok := roundtrip.([]interface{})
	if !ok {
		t.Fatalf("stored value type = %T, want []interface{} (old map fully replaced)", roundtrip)
	}
	if len(arr) != 2 || arr[0] != "alpha" || arr[1] != "beta" {
		t.Fatalf("stored value = %v, want [alpha beta]", arr)
	}
}

// TestStoreNamespaceIsolation 验证不同命名空间下同名 key 相互隔离，
// 删除一个命名空间的文档不影响另一命名空间。
func TestStoreNamespaceIsolation(t *testing.T) {
	store := newTestStore(t)
	const key = "profile"

	if err := store.PutJSON("tenant-a", key, map[string]string{"name": "a"}); err != nil {
		t.Fatalf("PutJSON(tenant-a) error = %v", err)
	}
	if err := store.PutJSON("tenant-b", key, map[string]string{"name": "b"}); err != nil {
		t.Fatalf("PutJSON(tenant-b) error = %v", err)
	}

	var a, b map[string]string
	if _, err := store.GetJSON("tenant-a", key, &a); err != nil {
		t.Fatalf("GetJSON(tenant-a) error = %v", err)
	}
	if _, err := store.GetJSON("tenant-b", key, &b); err != nil {
		t.Fatalf("GetJSON(tenant-b) error = %v", err)
	}
	if a["name"] != "a" || b["name"] != "b" {
		t.Fatalf("isolated values = %v / %v, want a / b", a, b)
	}

	if err := store.Delete("tenant-a", key); err != nil {
		t.Fatalf("Delete(tenant-a) error = %v", err)
	}

	found, err := store.GetJSON("tenant-a", key, &a)
	if err != nil {
		t.Fatalf("GetJSON(tenant-a) after delete error = %v", err)
	}
	if found {
		t.Fatal("tenant-a document should be deleted")
	}
	found, err = store.GetJSON("tenant-b", key, &b)
	if err != nil {
		t.Fatalf("GetJSON(tenant-b) after sibling delete error = %v", err)
	}
	if !found || b["name"] != "b" {
		t.Fatalf("tenant-b document = (found=%v, value=%v), want preserved b", found, b)
	}
}

// TestStoreDeleteIsScopedToKey 验证删除只作用于目标 key，同命名空间其他 key 保留。
func TestStoreDeleteIsScopedToKey(t *testing.T) {
	store := newTestStore(t)
	const ns = "scoped"

	if err := store.PutJSON(ns, "keep", map[string]string{"v": "keep"}); err != nil {
		t.Fatalf("PutJSON(keep) error = %v", err)
	}
	if err := store.PutJSON(ns, "drop", map[string]string{"v": "drop"}); err != nil {
		t.Fatalf("PutJSON(drop) error = %v", err)
	}
	if err := store.Delete(ns, "drop"); err != nil {
		t.Fatalf("Delete(drop) error = %v", err)
	}

	var keep, drop map[string]string
	foundKeep, err := store.GetJSON(ns, "keep", &keep)
	if err != nil || !foundKeep || keep["v"] != "keep" {
		t.Fatalf("keep document = (found=%v, value=%v, err=%v), want preserved", foundKeep, keep, err)
	}
	foundDrop, err := store.GetJSON(ns, "drop", &drop)
	if err != nil {
		t.Fatalf("GetJSON(drop) error = %v", err)
	}
	if foundDrop {
		t.Fatal("drop document should be deleted")
	}
}

// TestStoreDeleteNonExistentKey 验证删除不存在的 key 不报错（幂等）。
func TestStoreDeleteNonExistentKey(t *testing.T) {
	store := newTestStore(t)

	if err := store.Delete("ns", "never-written"); err != nil {
		t.Fatalf("Delete() on missing key error = %v, want nil", err)
	}
}
