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
)

type Storage struct {
	mu      sync.Mutex
	storage map[string]string
	owners  map[string]string
	file    *os.File
	deleted map[string]bool
}

func NewStorage(filePath string) (*Storage, error) {
	storage := make(map[string]string)
	owners := make(map[string]string)
	deleted := make(map[string]bool)
	if filePath == "" {
		return &Storage{storage: storage, owners: owners, file: nil, deleted: deleted}, nil
	}
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("file did not open - %w", err)
	}
	return &Storage{
		storage: storage,
		owners:  owners,
		file:    file,
		deleted: deleted,
	}, nil
}

func (s *Storage) Get(id string) (originalURL string, isDeleted bool, ok bool) {
	s.mu.Lock()
	originalURL, ok = s.storage[id]
	isDeleted = s.deleted[id]
	s.mu.Unlock()
	return
}

func (s *Storage) Save(id, UserID, url string) error {
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
	s.owners[id] = UserID

	if s.file != nil {
		record := URLRecord{}
		record.UUID = uuid.New().String()
		record.OriginalURL = url
		record.ShortURL = id
		record.UserID = UserID

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
			record := URLRecord{}
			if err := json.Unmarshal(line, &record); err != nil {
				return fmt.Errorf("parse storage record: %w", err)
			}
			s.storage[record.ShortURL] = record.OriginalURL
			s.owners[record.ShortURL] = record.UserID
			s.deleted[record.ShortURL] = record.DeletedFlag
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

func (s *Storage) SaveBatch(items []BatchItem) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, item := range items {
		if _, ok := s.storage[item.ID]; ok {
			return fmt.Errorf("%w: %q", ErrIDConflict, item.ID)
		}
		s.storage[item.ID] = item.URL
		s.owners[item.ID] = item.UserID

		if s.file != nil {
			record := URLRecord{}
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

func (s *Storage) GetUserURLs(userID string) ([]URLRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var result []URLRecord
	for id, ownerId := range s.owners {
		if ownerId == userID {
			result = append(result, URLRecord{
				ShortURL:    id,
				OriginalURL: s.storage[id],
			})
		}
	}
	return result, nil
}

func (s *Storage) DeleteBatch(userID string, shortURLs []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, id := range shortURLs {
		if s.owners[id] == userID {
			s.deleted[id] = true
		}
	}
	return nil
}
