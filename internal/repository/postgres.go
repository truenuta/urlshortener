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

func (p *PostgresStorage) Save(shortUrl, userID, url string) error {
	var newShortURL string
	err := p.db.QueryRow(
		`INSERT INTO urls (user_id, uuid, short_url, original_url) VALUES ($1, $2, $3, $4)
                 ON CONFLICT (original_url) DO UPDATE SET original_url = EXCLUDED.original_url
                 RETURNING short_url`,
		userID, uuid.New().String(), shortUrl, url).Scan(&newShortURL)
	if err != nil {
		return fmt.Errorf("can not save url: %w", err)
	}
	if newShortURL != shortUrl {
		return NewConflictError(newShortURL)
	}
	return nil
}

func (p *PostgresStorage) Get(shortUrl string) (string, bool, bool) {
	var original_url string
	var isDeleted bool
	rows := p.db.QueryRow(`SELECT original_url, is_deleted FROM urls WHERE short_url = $1`, shortUrl)
	err := rows.Scan(&original_url, &isDeleted)
	if err != nil {
		return "", false, false
	}
	return original_url, isDeleted, true
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
	valueArgs := make([]interface{}, 0, len(items)*4)
	for i, item := range items {
		n := i * 4
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d)", n+1, n+2, n+3, n+4))
		valueArgs = append(valueArgs, item.UserID, uuid.New().String(), item.ID, item.URL)
	}

	query := fmt.Sprintf(
		`INSERT INTO urls (user_id, uuid, short_url, original_url) VALUES %s
          ON CONFLICT (original_url) DO UPDATE SET original_url = EXCLUDED.original_url`,
		strings.Join(valueStrings, ","),
	)

	_, err := p.db.Exec(query, valueArgs...)
	return err
}
func (p *PostgresStorage) GetUserURLs(userID string) ([]URLRecord, error) {
	rows, err := p.db.Query(`SELECT short_url, original_url FROM urls WHERE user_id = $1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []URLRecord
	for rows.Next() {
		var rec URLRecord
		if err := rows.Scan(&rec.ShortURL, &rec.OriginalURL); err != nil {
			return nil, err
		}
		result = append(result, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (p *PostgresStorage) DeleteBatch(userID string, shortURLs []string) error {
	if len(shortURLs) == 0 {
		return nil
	}
	_, err := p.db.Exec(
		`UPDATE urls SET is_deleted = TRUE
         WHERE user_id = $1 AND short_url = ANY($2)`,
		userID, shortURLs,
	)
	if err != nil {
		return fmt.Errorf("delete batch: %w", err)
	}
	return nil
}
