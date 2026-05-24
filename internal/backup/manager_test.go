package backup

import (
	"path/filepath"
	"testing"

	"github.com/kudig-io/klaw/internal/storage"
)

func TestBackupManagerLifecycle(t *testing.T) {
	store, err := storage.NewStore(filepath.Join(t.TempDir(), "klaw.db"))
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}
	manager := NewManager(store)

	item, err := manager.Create("demo", CreateBackupRequest{
		Name:       "daily-backup-1",
		BackupMode: BackupModeFull,
		StorageLocation: StorageLocation{
			Provider:          StorageProviderS3,
			Bucket:            "demo-bucket",
			Region:            "us-east-1",
			CredentialsSecret: "s3-credentials",
		},
		Validation: ValidationConfig{
			Enabled:          true,
			ConsistencyCheck: true,
		},
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if item.Phase != BackupPhaseCompleted {
		t.Fatalf("Phase = %s, want %s", item.Phase, BackupPhaseCompleted)
	}

	list := manager.List("demo")
	if len(list) != 1 {
		t.Fatalf("List() len = %d, want 1", len(list))
	}

	got, err := manager.Get("demo", "daily-backup-1")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.SnapshotLocation == "" {
		t.Fatal("SnapshotLocation should not be empty")
	}

	summary := manager.Summary("demo")
	if summary.Total != 1 {
		t.Fatalf("Summary.Total = %d, want 1", summary.Total)
	}
	if summary.ByMode[string(BackupModeFull)] != 1 {
		t.Fatalf("Summary.ByMode[Full] = %d, want 1", summary.ByMode[string(BackupModeFull)])
	}

	if err := manager.Delete("demo", "daily-backup-1"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	if len(manager.List("demo")) != 0 {
		t.Fatal("List() should be empty after delete")
	}
}
