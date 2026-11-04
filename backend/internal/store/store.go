// Package store persists config and profiles in SQLite, in the same JSON
// shape the Flutter app's preferences keep.
package store

import (
	"database/sql"
	"encoding/json"
	"errors"
	"sync"

	_ "modernc.org/sqlite"
)

type Store struct {
	mu sync.Mutex
	db *sql.DB
}

func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	s := &Store{db: db}
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS kv (
			key   TEXT PRIMARY KEY,
			value TEXT NOT NULL
		);
		CREATE TABLE IF NOT EXISTS profiles (
			id   INTEGER PRIMARY KEY,
			data TEXT NOT NULL
		);
	`); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) Close() error { return s.db.Close() }

func (s *Store) Get(key string) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var value []byte
	err := s.db.QueryRow(`SELECT value FROM kv WHERE key = ?`, key).Scan(&value)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return value, err
}

func (s *Store) Set(key string, value []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(`
		INSERT INTO kv (key, value) VALUES (?, ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value
	`, key, value)
	return err
}

// Profiles lists profiles in their stored order.
func (s *Store) Profiles() ([]json.RawMessage, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	rows, err := s.db.Query(`SELECT data FROM profiles ORDER BY rowid`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []json.RawMessage
	for rows.Next() {
		var data []byte
		if err := rows.Scan(&data); err != nil {
			return nil, err
		}
		list = append(list, json.RawMessage(data))
	}
	return list, rows.Err()
}

func (s *Store) SaveProfile(id int64, data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(`
		INSERT INTO profiles (id, data) VALUES (?, ?)
		ON CONFLICT(id) DO UPDATE SET data = excluded.data
	`, id, data)
	return err
}

func (s *Store) DeleteProfile(id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(`DELETE FROM profiles WHERE id = ?`, id)
	return err
}
