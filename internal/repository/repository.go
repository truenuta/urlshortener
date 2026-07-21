package repository

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/google/uuid"
)

var ErrIDConflict = errors.New("id already exists")

type FileURLRecord struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type Repository interface {
	Save(id, url string) error
	Get(id string) (string, bool)
}

type Storage struct {
	mu       sync.RWMutex
	storage  map[string]string
	filePath string
}

func NewStorage(filePath string) *Storage {
	return &Storage{
		storage:  make(map[string]string),
		filePath: filePath,
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

	if s.filePath == "" {
		return nil
	}
	record := FileURLRecord{}
	record.UUID = uuid.New().String()
	record.OriginalURL = url
	record.ShortURL = id
	file, err := os.OpenFile(s.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("file did not open - %w", err)

	}
	defer file.Close()
	data, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("Marshal is failed - %w", err)
	}
	data = append(data, '\n')
	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("saving to file is failed - %w", err)
	}
	return nil
}

func (s *Storage) Load() error {
	if s.filePath == "" {
		return nil
	}
	file, err := os.OpenFile(s.filePath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Bytes()
		record := FileURLRecord{}
		if err := json.Unmarshal(line, &record); err != nil {
			return fmt.Errorf("parse storage record: %w", err)
		}
		s.storage[record.ShortURL] = record.OriginalURL
	}
	return nil
}
