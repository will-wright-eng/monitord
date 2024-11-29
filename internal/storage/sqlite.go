package storage

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"

	"github.com/will-wright-eng/monitord/internal/monitor"
)

type SQLiteStore struct {
	db *sql.DB
}

func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
	// Create the directory path if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	if err := createSchema(db); err != nil {
		db.Close()
		return nil, err
	}
	return &SQLiteStore{db: db}, nil
}

func createSchema(db *sql.DB) error {
	_, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS health_checks (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT NOT NULL,
            url TEXT NOT NULL,
            status TEXT NOT NULL,
            status_code INTEGER,
            response_time INTEGER,
            timestamp DATETIME NOT NULL,
            error TEXT,
            tags TEXT,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP
        );
        CREATE INDEX IF NOT EXISTS idx_url_timestamp ON health_checks(url, timestamp);
        CREATE INDEX IF NOT EXISTS idx_name ON health_checks(name);
    `)
	return err
}

func (s *SQLiteStore) SaveCheck(check monitor.HealthCheck) error {
	tags := strings.Join(check.Tags, ",")
	_, err := s.db.Exec(`
        INSERT INTO health_checks (name, url, status, status_code, response_time, timestamp, error, tags)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		check.Name,
		check.URL,
		check.Status,
		check.StatusCode,
		check.ResponseTime,
		check.Timestamp,
		check.Error,
		tags,
	)
	return err
}

func (s *SQLiteStore) Close() error {
	return s.db.Close()
}
