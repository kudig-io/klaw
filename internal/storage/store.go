package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
	mu sync.Mutex
}

func NewStore(dbPath string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	store := &Store{db: db}
	if err := store.init(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return store, nil
}

func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *Store) GetJSON(namespace, key string, dest interface{}) (bool, error) {
	var raw string
	err := s.db.QueryRow(`SELECT value FROM documents WHERE namespace = ? AND key = ?`, namespace, key).Scan(&raw)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if err := json.Unmarshal([]byte(raw), dest); err != nil {
		return false, err
	}
	return true, nil
}

func (s *Store) PutJSON(namespace, key string, value interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(`
		INSERT INTO documents(namespace, key, value, updated_at)
		VALUES(?, ?, ?, ?)
		ON CONFLICT(namespace, key) DO UPDATE SET
			value = excluded.value,
			updated_at = excluded.updated_at
	`, namespace, key, string(raw), time.Now().UTC())
	return err
}

func (s *Store) Delete(namespace, key string) error {
	_, err := s.db.Exec(`DELETE FROM documents WHERE namespace = ? AND key = ?`, namespace, key)
	return err
}

func (s *Store) init() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS documents (
			namespace TEXT NOT NULL,
			key TEXT NOT NULL,
			value TEXT NOT NULL,
			updated_at DATETIME NOT NULL,
			PRIMARY KEY(namespace, key)
		)
	`)
	if err != nil {
		return fmt.Errorf("initialize storage: %w", err)
	}
	if err := s.InitSchema(); err != nil {
		return fmt.Errorf("initialize domain schema: %w", err)
	}
	return nil
}
