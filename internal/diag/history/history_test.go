package history

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kudig-io/klaw/internal/diag/types"
)

func TestNewManager(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	
	mgr := NewManagerWithPath(tempDir)
	if mgr == nil {
		t.Fatal("Expected non-nil manager")
	}
	if mgr.historyDir != tempDir {
		t.Errorf("Expected historyDir to be %s, got %s", tempDir, mgr.historyDir)
	}
}

func TestManager_SaveAndGet(t *testing.T) {
	tempDir := t.TempDir()
	mgr := NewManagerWithPath(tempDir)

	issues := []types.Issue{
		{
			Severity: types.SeverityCritical,
			CNName:   "系统负载过高",
			ENName:   "HIGH_SYSTEM_LOAD",
			Details:  "负载 8.5，超过阈值",
			Location: "system_status",
		},
		{
			Severity: types.SeverityWarning,
			CNName:   "内存使用率偏高",
			ENName:   "ELEVATED_MEMORY_USAGE",
			Details:  "内存使用率 87%",
			Location: "memory_info",
		},
	}

	// Save entry
	entry, err := mgr.Save("test-node", "online", issues)
	if err != nil {
		t.Fatalf("Failed to save entry: %v", err)
	}

	if entry.ID == "" {
		t.Error("Expected non-empty ID")
	}

	if entry.Hostname != "test-node" {
		t.Errorf("Expected hostname 'test-node', got '%s'", entry.Hostname)
	}

	if entry.Mode != "online" {
		t.Errorf("Expected mode 'online', got '%s'", entry.Mode)
	}

	if entry.Summary.Critical != 1 {
		t.Errorf("Expected 1 critical issue, got %d", entry.Summary.Critical)
	}

	if entry.Summary.Warning != 1 {
		t.Errorf("Expected 1 warning issue, got %d", entry.Summary.Warning)
	}

	// Retrieve entry
	retrieved, err := mgr.Get(entry.ID)
	if err != nil {
		t.Fatalf("Failed to get entry: %v", err)
	}

	if retrieved.ID != entry.ID {
		t.Errorf("Expected ID %s, got %s", entry.ID, retrieved.ID)
	}

	if len(retrieved.Issues) != len(issues) {
		t.Errorf("Expected %d issues, got %d", len(issues), len(retrieved.Issues))
	}
}

func TestManager_List(t *testing.T) {
	tempDir := t.TempDir()
	mgr := NewManagerWithPath(tempDir)

	// Create multiple entries
	for i := 0; i < 3; i++ {
		issues := []types.Issue{
			{
				Severity: types.SeverityInfo,
				CNName:   "测试问题",
				ENName:   "TEST_ISSUE",
				Details:  "测试详情",
				Location: "test_location",
			},
		}
		_, err := mgr.Save("test-node", "online", issues)
		if err != nil {
			t.Fatalf("Failed to save entry: %v", err)
		}
		time.Sleep(10 * time.Millisecond) // Ensure different timestamps
	}

	// List entries
	entries, err := mgr.List()
	if err != nil {
		t.Fatalf("Failed to list entries: %v", err)
	}

	if len(entries) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(entries))
	}

	// Check that entries are sorted by timestamp (newest first)
	for i := 1; i < len(entries); i++ {
		if entries[i].Timestamp.After(entries[i-1].Timestamp) {
			t.Error("Entries should be sorted by timestamp (newest first)")
		}
	}
}

func TestManager_Delete(t *testing.T) {
	tempDir := t.TempDir()
	mgr := NewManagerWithPath(tempDir)

	issues := []types.Issue{
		{
			Severity: types.SeverityInfo,
			CNName:   "测试问题",
			ENName:   "TEST_ISSUE",
			Details:  "测试详情",
			Location: "test_location",
		},
	}

	entry, err := mgr.Save("test-node", "online", issues)
	if err != nil {
		t.Fatalf("Failed to save entry: %v", err)
	}

	// Delete entry
	err = mgr.Delete(entry.ID)
	if err != nil {
		t.Fatalf("Failed to delete entry: %v", err)
	}

	// Verify entry is deleted
	_, err = mgr.Get(entry.ID)
	if err == nil {
		t.Error("Expected error when getting deleted entry")
	}
}

func TestManager_Diff(t *testing.T) {
	tempDir := t.TempDir()
	mgr := NewManagerWithPath(tempDir)

	// Create first entry
	issues1 := []types.Issue{
		{
			Severity: types.SeverityCritical,
			CNName:   "系统负载过高",
			ENName:   "HIGH_SYSTEM_LOAD",
			Details:  "负载 8.5",
			Location: "system_status",
		},
		{
			Severity: types.SeverityWarning,
			CNName:   "内存使用率偏高",
			ENName:   "ELEVATED_MEMORY_USAGE",
			Details:  "内存使用率 87%",
			Location: "memory_info",
		},
	}

	entry1, err := mgr.Save("test-node", "online", issues1)
	if err != nil {
		t.Fatalf("Failed to save entry1: %v", err)
	}

	// Create second entry with different issues
	issues2 := []types.Issue{
		{
			Severity: types.SeverityCritical,
			CNName:   "系统负载过高",
			ENName:   "HIGH_SYSTEM_LOAD",
			Details:  "负载 8.5",
			Location: "system_status",
		},
		{
			Severity: types.SeverityInfo,
			CNName:   "Swap未禁用",
			ENName:   "SWAP_NOT_DISABLED",
			Details:  "建议禁用swap",
			Location: "system_info",
		},
	}

	entry2, err := mgr.Save("test-node", "online", issues2)
	if err != nil {
		t.Fatalf("Failed to save entry2: %v", err)
	}

	// Diff entries
	diff, err := mgr.Diff(entry1.ID, entry2.ID)
	if err != nil {
		t.Fatalf("Failed to diff entries: %v", err)
	}

	// Should have 1 removed issue (ELEVATED_MEMORY_USAGE)
	if len(diff.RemovedIssues) != 1 {
		t.Errorf("Expected 1 removed issue, got %d", len(diff.RemovedIssues))
	}

	// Should have 1 added issue (SWAP_NOT_DISABLED)
	if len(diff.AddedIssues) != 1 {
		t.Errorf("Expected 1 added issue, got %d", len(diff.AddedIssues))
	}

	// Verify the specific issues
	if len(diff.RemovedIssues) > 0 && diff.RemovedIssues[0].ENName != "ELEVATED_MEMORY_USAGE" {
		t.Errorf("Expected ELEVATED_MEMORY_USAGE to be removed, got %s", diff.RemovedIssues[0].ENName)
	}

	if len(diff.AddedIssues) > 0 && diff.AddedIssues[0].ENName != "SWAP_NOT_DISABLED" {
		t.Errorf("Expected SWAP_NOT_DISABLED to be added, got %s", diff.AddedIssues[0].ENName)
	}
}

func TestManager_CleanupOldEntries(t *testing.T) {
	tempDir := t.TempDir()
	mgr := NewManagerWithPath(tempDir)

	// Create an entry
	issues := []types.Issue{
		{
			Severity: types.SeverityInfo,
			CNName:   "测试问题",
			ENName:   "TEST_ISSUE",
			Details:  "测试详情",
			Location: "test_location",
		},
	}

	entry, err := mgr.Save("test-node", "online", issues)
	if err != nil {
		t.Fatalf("Failed to save entry: %v", err)
	}

	// Manually modify the entry file to make it appear old
	filename := filepath.Join(tempDir, entry.ID+".json")
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read entry file: %v", err)
	}

	// Parse, modify timestamp, and write back
	var oldEntry Entry
	if err := json.Unmarshal(data, &oldEntry); err != nil {
		t.Fatalf("Failed to unmarshal entry: %v", err)
	}
	oldEntry.Timestamp = time.Now().Add(-30 * 24 * time.Hour) // 30 days ago
	
	newData, err := json.MarshalIndent(oldEntry, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal old entry: %v", err)
	}
	
	if err := os.WriteFile(filename, newData, 0600); err != nil {
		t.Fatalf("Failed to write old entry: %v", err)
	}

	// Cleanup entries older than 7 days
	deleted, err := mgr.CleanupOldEntries(7 * 24 * time.Hour)
	if err != nil {
		t.Fatalf("Failed to cleanup entries: %v", err)
	}

	if deleted != 1 {
		t.Errorf("Expected 1 deleted entry, got %d", deleted)
	}

	// Verify entry is gone
	entries, err := mgr.List()
	if err != nil {
		t.Fatalf("Failed to list entries: %v", err)
	}

	if len(entries) != 0 {
		t.Errorf("Expected 0 entries after cleanup, got %d", len(entries))
	}
}

func TestIssueKey(t *testing.T) {
	issue := types.Issue{
		ENName:   "TEST_ISSUE",
		Location: "test_location",
		Details:  "test details",
	}

	key := issueKey(issue)
	expected := "TEST_ISSUE:test_location:test details"
	if key != expected {
		t.Errorf("Expected key '%s', got '%s'", expected, key)
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		duration time.Duration
		expected string
	}{
		{30 * time.Second, "30 seconds"},
		{5 * time.Minute, "5 minutes"},
		{2 * time.Hour, "2 hours"},
		{48 * time.Hour, "2 days"},
	}

	for _, tc := range tests {
		result := FormatDuration(tc.duration)
		if result != tc.expected {
			t.Errorf("FormatDuration(%v) = '%s', expected '%s'", tc.duration, result, tc.expected)
		}
	}
}

func TestGenerateID(t *testing.T) {
	id1 := generateID()
	id2 := generateID()

	if id1 == "" {
		t.Error("Expected non-empty ID")
	}

	if id1 == id2 {
		t.Error("Expected unique IDs")
	}

	if len(id1) != 16 {
		t.Errorf("Expected ID length 16, got %d", len(id1))
	}
}
