package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/truenuta/urlshortener/internal/db"
)

var ErrIDConflict = errors.New("id already exists")

type ConflictError struct {
	ShortURL string
}

func NewConflictError(url string) *ConflictError {
	return &ConflictError{ShortURL: url}
}

func (ce *ConflictError) Error() string {
	return ce.ShortURL
}

type URLRecord struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	UserID      string `json:"user_id"`
}

type BatchItem struct {
	UserID string
	ID     string
	URL    string
}

type URLRepository interface {
	Save(id, UserID, url string) error
	Get(id string) (string, bool)
	Load() error
	Close() error
	Ping(ctx context.Context) error
	SaveBatch(items []BatchItem) error
	GetUserURLs(userID string) ([]URLRecord, error)
}

func NewURLRepository(dsn, filepath string) (URLRepository, error) {
	if dsn != "" {
		database, err := db.NewDB(dsn)
		if err != nil {
			return nil, fmt.Errorf("can not connect to database: %w", err)
		}
		if err := db.RunMigrations(database); err != nil {
			database.Close()
			return nil, fmt.Errorf("run migrations: %w", err)
		}
		return NewPostgresStorage(database), nil
	}
	return NewStorage(filepath)
}
