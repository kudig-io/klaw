package etcdbackup

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Client talks to the etcd-guardian backend API.
type Client struct {
	baseURL string
	http    *http.Client
}

// NewClient creates a new etcd-guardian API client.
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		http:    &http.Client{Timeout: 30 * time.Second},
	}
}

// SetHTTPClient overrides the default HTTP client.
func (c *Client) SetHTTPClient(client *http.Client) {
	c.http = client
}

// ListBackups fetches the backup list from etcd-guardian.
func (c *Client) ListBackups(ctx context.Context) ([]Backup, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/v1/backups", nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	var backups []Backup
	if err := json.NewDecoder(resp.Body).Decode(&backups); err != nil {
		return nil, err
	}
	return backups, nil
}

// CreateBackup triggers a new backup via etcd-guardian.
func (c *Client) CreateBackup(ctx context.Context, req BackupSpec) (*Backup, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	hreq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/backups", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	hreq.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(hreq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	var backup Backup
	if err := json.NewDecoder(resp.Body).Decode(&backup); err != nil {
		return nil, err
	}
	return &backup, nil
}

// GetBackup fetches a single backup by name.
func (c *Client) GetBackup(ctx context.Context, name string) (*Backup, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/v1/backups/"+name, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	var backup Backup
	if err := json.NewDecoder(resp.Body).Decode(&backup); err != nil {
		return nil, err
	}
	return &backup, nil
}

// DeleteBackup removes a backup by name.
func (c *Client) DeleteBackup(ctx context.Context, name string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.baseURL+"/api/v1/backups/"+name, nil)
	if err != nil {
		return err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	return nil
}

// Health checks whether the etcd-guardian backend is reachable.
func (c *Client) Health(ctx context.Context) (bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/health", nil)
	if err != nil {
		return false, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK, nil
}
