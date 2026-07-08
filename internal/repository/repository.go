package repository

import (
	"errors"
	"fmt"
	"sync"
)

var ErrIDConflict = errors.New("id already exists")

type Repository interface {
	Save(id, url string) error
	Get(id string) (string, bool)
}

type Storage struct {
	mu      sync.RWMutex
	storage map[string]string
}

func NewStorage() *Storage {
	return &Storage{
		storage: make(map[string]string),
	}
}

func (s *Storage) Get(id string) (originalURL string, ok bool) {
	s.mu.RLock()
	originalURL, ok = s.storage[id]
	s.mu.RUnlock()
	return
}

func (s *Storage) Save(id, url string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.storage[id]; ok {
		return fmt.Errorf("%w: %q", ErrIDConflict, id)
	}
	s.storage[id] = url
	return nil
}
