package repository

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/google/uuid"
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

type FileURLRecord struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type Storage struct {
	mu      sync.Mutex
	storage map[string]string
	file    *os.File
}

func NewStorage(filePath string) (*Storage, error) {
	storage := make(map[string]string)
	if filePath == "" {
		return &Storage{storage: storage, file: nil}, nil
	}
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("file did not open - %w", err)
	}
	return &Storage{
		storage: storage,
		file:    file,
	}, nil
}

func (s *Storage) Get(id string) (originalURL string, ok bool) {
	s.mu.Lock()
	originalURL, ok = s.storage[id]
	s.mu.Unlock()
	return
}

func (s *Storage) Save(id, url string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for existingID, existingURL := range s.storage {
		if existingURL == url {
			return NewConflictError(existingID)
		}
	}
	if _, ok := s.storage[id]; ok {
		return fmt.Errorf("%w: %q", ErrIDConflict, id)
	}
	s.storage[id] = url

	if s.file != nil {
		record := FileURLRecord{}
		record.UUID = uuid.New().String()
		record.OriginalURL = url
		record.ShortURL = id

		data, err := json.Marshal(record)
		if err != nil {
			return fmt.Errorf("Marshal is failed - %w", err)
		}
		data = append(data, '\n')
		_, err = s.file.Write(data)
		if err != nil {
			return fmt.Errorf("saving to file is failed - %w", err)
		}
	}
	return nil
}

func (s *Storage) Load() error {
	if s.file != nil {
		scanner := bufio.NewScanner(s.file)
		s.mu.Lock()
		defer s.mu.Unlock()

		for scanner.Scan() {
			line := scanner.Bytes()
			record := FileURLRecord{}
			if err := json.Unmarshal(line, &record); err != nil {
				return fmt.Errorf("parse storage record: %w", err)
			}
			s.storage[record.ShortURL] = record.OriginalURL
		}
	}
	return nil
}

func (s *Storage) Close() error {
	if s.file != nil {
		err := s.file.Close()
		return err
	}
	return nil
}

func (s *Storage) Ping(ctx context.Context) error {
	return errors.New("database is not configured")
}

type URLRepository interface {
	Save(id, url string) error
	Get(id string) (string, bool)
	Load() error
	Close() error
	Ping(ctx context.Context) error
	SaveBatch(items []BatchItem) error
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

type BatchItem struct {
	ID  string
	URL string
}

func (s *Storage) SaveBatch(items []BatchItem) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, item := range items {
		if _, ok := s.storage[item.ID]; ok {
			return fmt.Errorf("%w: %q", ErrIDConflict, item.ID)
		}
		s.storage[item.ID] = item.URL
		if s.file != nil {
			record := FileURLRecord{}
			record.UUID = uuid.New().String()
			record.OriginalURL = item.URL
			record.ShortURL = item.ID
			data, err := json.Marshal(record)
			if err != nil {
				return fmt.Errorf("marshal batch record: %w", err)
			}
			data = append(data, '\n')
			_, err = s.file.Write(data)
			if err != nil {
				return fmt.Errorf("saving to file is failed - %w", err)
			}
		}
	}
	return nil

}
