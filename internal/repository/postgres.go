package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage(db *sql.DB) *PostgresStorage {
	return &PostgresStorage{db: db}
}

func (p *PostgresStorage) Save(id, url string) error {
	var shortURL string
	err := p.db.QueryRow(
		`INSERT INTO urls (uuid, short_url, original_url) VALUES ($1, $2, $3)
                 ON CONFLICT (original_url) DO UPDATE SET original_url = EXCLUDED.original_url
                 RETURNING short_url`,
		uuid.New().String(), id, url).Scan(&shortURL)
	if err != nil {
		return fmt.Errorf("can not save url: %w", err)
	}
	if shortURL != id {
		return NewConflictError(shortURL)
	}
	return nil
}

func (p *PostgresStorage) Get(id string) (string, bool) {
	var original_url string
	rows := p.db.QueryRow(`SELECT original_url FROM urls WHERE short_url = $1`, id)
	err := rows.Scan(&original_url)
	if err != nil {
		return "", false
	}
	return original_url, true
}

func (p *PostgresStorage) Close() error { return p.db.Close() }
func (p *PostgresStorage) Load() error  { return nil }
func (p *PostgresStorage) Ping(ctx context.Context) error {
	return p.db.PingContext(ctx)
}

func (p *PostgresStorage) SaveBatch(items []BatchItem) error {
	if len(items) == 0 {
		return nil
	}

	valueStrings := make([]string, 0, len(items))
	valueArgs := make([]interface{}, 0, len(items)*3)
	for i, item := range items {
		n := i * 3
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d)", n+1, n+2, n+3))
		valueArgs = append(valueArgs, uuid.New().String(), item.ID, item.URL)
	}

	query := fmt.Sprintf(
		`INSERT INTO urls (uuid, short_url, original_url) VALUES %s
          ON CONFLICT (original_url) DO UPDATE SET original_url = EXCLUDED.original_url`,
		strings.Join(valueStrings, ","),
	)

	_, err := p.db.Exec(query, valueArgs...)
	return err
}
