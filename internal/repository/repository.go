package repository

import (
	"sync"
)

type Repository interface {
	Save(id, url string)
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

func (s *Storage) Save(id, url string) {
	s.mu.Lock()
	s.storage[id] = url
	s.mu.Unlock()
}
