package etcdbackup

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Manager provides high-level backup/restore/schedule operations
// backed by an etcd-guardian API client.
type Manager struct {
	client *Client
	mu     sync.RWMutex
	cache  struct {
		backups   []Backup
		restores  []Restore
		schedules []Schedule
		updatedAt time.Time
	}
}

// NewManager creates a new etcd-backup Manager.
func NewManager(client *Client) *Manager {
	return &Manager{client: client}
}

// ---------------------------------------------------------------------------
// Backup operations
// ---------------------------------------------------------------------------

// ListBackups returns all backups, optionally refreshing from remote.
func (m *Manager) ListBackups(ctx context.Context, force bool) ([]Backup, error) {
	m.mu.RLock()
	cacheValid := !force && len(m.cache.backups) > 0 && time.Since(m.cache.updatedAt) < 30*time.Second
	backups := append([]Backup(nil), m.cache.backups...)
	m.mu.RUnlock()

	if cacheValid {
		return backups, nil
	}

	remote, err := m.client.ListBackups(ctx)
	if err != nil {
		return nil, fmt.Errorf("list backups: %w", err)
	}

	m.mu.Lock()
	m.cache.backups = remote
	m.cache.updatedAt = time.Now()
	m.mu.Unlock()
	return remote, nil
}

// CreateBackup triggers a new backup and invalidates cache.
func (m *Manager) CreateBackup(ctx context.Context, spec BackupSpec) (*Backup, error) {
	backup, err := m.client.CreateBackup(ctx, spec)
	if err != nil {
		return nil, fmt.Errorf("create backup: %w", err)
	}
	m.invalidateCache()
	return backup, nil
}

// GetBackup fetches a single backup by name.
func (m *Manager) GetBackup(ctx context.Context, name string) (*Backup, error) {
	return m.client.GetBackup(ctx, name)
}

// DeleteBackup removes a backup and invalidates cache.
func (m *Manager) DeleteBackup(ctx context.Context, name string) error {
	if err := m.client.DeleteBackup(ctx, name); err != nil {
		return fmt.Errorf("delete backup: %w", err)
	}
	m.invalidateCache()
	return nil
}

// Summary returns aggregated backup statistics.
func (m *Manager) Summary(ctx context.Context) (Summary, error) {
	backups, err := m.ListBackups(ctx, false)
	if err != nil {
		return Summary{}, err
	}

	summary := Summary{
		ByPhase: make(map[string]int),
		ByMode:  make(map[string]int),
	}
	dayAgo := time.Now().Add(-24 * time.Hour)
	for _, b := range backups {
		summary.Total++
		summary.ByPhase[string(b.Phase)]++
		summary.ByMode[string(b.Spec.BackupMode)]++
		if b.CreatedAt.After(dayAgo) {
			summary.Recent24h++
		}
	}
	return summary, nil
}

// ---------------------------------------------------------------------------
// Health
// ---------------------------------------------------------------------------

// Health checks backend connectivity.
func (m *Manager) Health(ctx context.Context) (bool, error) {
	return m.client.Health(ctx)
}

// ---------------------------------------------------------------------------
// Internal
// ---------------------------------------------------------------------------

func (m *Manager) invalidateCache() {
	m.mu.Lock()
	m.cache.backups = nil
	m.cache.updatedAt = time.Time{}
	m.mu.Unlock()
}
