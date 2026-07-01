// Package history provides functionality for storing and comparing diagnostic history
package history

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/kudig-io/klaw/internal/diag/types"
)

// Ensure json is used for Unmarshal in tests
var _ = json.Unmarshal

const (
	historyDirName = ".klaw"
	historySubDir  = "history"
)

// Entry represents a single diagnostic history entry
type Entry struct {
	ID        string        `json:"id"`
	Timestamp time.Time     `json:"timestamp"`
	Hostname  string        `json:"hostname"`
	Mode      string        `json:"mode"`
	Issues    []types.Issue `json:"issues"`
	Summary   types.IssueSummary `json:"summary"`
}

// Manager manages diagnostic history
type Manager struct {
	historyDir string
}

// NewManager creates a new history manager
func NewManager() (*Manager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	historyDir := filepath.Join(homeDir, historyDirName, historySubDir)
	if err := os.MkdirAll(historyDir, 0750); err != nil {
		return nil, fmt.Errorf("failed to create history directory: %w", err)
	}

	return &Manager{
		historyDir: historyDir,
	}, nil
}

// NewManagerWithPath creates a new history manager with a custom path
func NewManagerWithPath(path string) *Manager {
	return &Manager{
		historyDir: path,
	}
}

// Save saves a diagnostic result to history
func (m *Manager) Save(hostname, mode string, issues []types.Issue) (*Entry, error) {
	entry := &Entry{
		ID:        generateID(),
		Timestamp: time.Now(),
		Hostname:  hostname,
		Mode:      mode,
		Issues:    issues,
		Summary:   types.CalculateSummary(issues),
	}

	filename := filepath.Join(m.historyDir, entry.ID+".json")
	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal entry: %w", err)
	}

	if err := os.WriteFile(filename, data, 0600); err != nil {
		return nil, fmt.Errorf("failed to write history file: %w", err)
	}

	return entry, nil
}

// List returns all history entries, sorted by timestamp (newest first)
func (m *Manager) List() ([]Entry, error) {
	files, err := os.ReadDir(m.historyDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read history directory: %w", err)
	}

	var entries []Entry
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		data, err := os.ReadFile(filepath.Join(m.historyDir, file.Name()))
		if err != nil {
			continue // Skip files we can't read
		}

		var entry Entry
		if err := json.Unmarshal(data, &entry); err != nil {
			continue // Skip files we can't parse
		}

		entries = append(entries, entry)
	}

	// Sort by timestamp, newest first
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Timestamp.After(entries[j].Timestamp)
	})

	return entries, nil
}

// Get retrieves a specific history entry by ID
func (m *Manager) Get(id string) (*Entry, error) {
	filename := filepath.Join(m.historyDir, id+".json")
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read history file: %w", err)
	}

	var entry Entry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, fmt.Errorf("failed to parse history file: %w", err)
	}

	return &entry, nil
}

// Delete removes a history entry
func (m *Manager) Delete(id string) error {
	filename := filepath.Join(m.historyDir, id+".json")
	if err := os.Remove(filename); err != nil {
		return fmt.Errorf("failed to delete history file: %w", err)
	}
	return nil
}

// DiffResult represents the result of comparing two history entries
type DiffResult struct {
	Entry1        *Entry
	Entry2        *Entry
	AddedIssues   []types.Issue
	RemovedIssues []types.Issue
	ChangedIssues []IssueChange
}

// IssueChange represents a changed issue between two entries
type IssueChange struct {
	OldIssue types.Issue
	NewIssue types.Issue
	Changes  []string
}

// Diff compares two history entries and returns the differences
func (m *Manager) Diff(id1, id2 string) (*DiffResult, error) {
	entry1, err := m.Get(id1)
	if err != nil {
		return nil, fmt.Errorf("failed to get entry %s: %w", id1, err)
	}

	entry2, err := m.Get(id2)
	if err != nil {
		return nil, fmt.Errorf("failed to get entry %s: %w", id2, err)
	}

	result := &DiffResult{
		Entry1: entry1,
		Entry2: entry2,
	}

	// Create maps for comparison
	issues1 := make(map[string]types.Issue)
	for _, issue := range entry1.Issues {
		key := issueKey(issue)
		issues1[key] = issue
	}

	issues2 := make(map[string]types.Issue)
	for _, issue := range entry2.Issues {
		key := issueKey(issue)
		issues2[key] = issue
	}

	// Find added issues (in entry2 but not in entry1)
	for key, issue := range issues2 {
		if _, exists := issues1[key]; !exists {
			result.AddedIssues = append(result.AddedIssues, issue)
		}
	}

	// Find removed issues (in entry1 but not in entry2)
	for key, issue := range issues1 {
		if _, exists := issues2[key]; !exists {
			result.RemovedIssues = append(result.RemovedIssues, issue)
		}
	}

	return result, nil
}

// CleanupOldEntries removes entries older than the specified duration
func (m *Manager) CleanupOldEntries(maxAge time.Duration) (int, error) {
	entries, err := m.List()
	if err != nil {
		return 0, err
	}

	cutoff := time.Now().Add(-maxAge)
	deleted := 0

	for _, entry := range entries {
		if entry.Timestamp.Before(cutoff) {
			if err := m.Delete(entry.ID); err == nil {
				deleted++
			}
		}
	}

	return deleted, nil
}

// issueKey generates a unique key for an issue for comparison
func issueKey(issue types.Issue) string {
	return fmt.Sprintf("%s:%s:%s", issue.ENName, issue.Location, issue.Details)
}

// generateID generates a unique ID for a history entry
func generateID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// crypto/rand.Read should never fail on modern systems,
		// but fall back to time-based ID if it does
		return fmt.Sprintf("%x", time.Now().UnixNano())
	}
	data := fmt.Sprintf("%d-%x", time.Now().UnixNano(), b)
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)[:16]
}

// FormatDuration formats a duration in a human-readable way
func FormatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%d seconds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%d minutes", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%d hours", int(d.Hours()))
	}
	return fmt.Sprintf("%d days", int(d.Hours()/24))
}
