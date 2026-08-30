package service

import (
	"errors"
	"math/rand"

	"github.com/truenuta/urlshortener/internal/model"
	"github.com/truenuta/urlshortener/internal/repository"
)

var ErrGeneratingIdFail = errors.New("failed generating unique id")
var lenOfGeneratedURL int = 8

type ConflictError struct {
	ShortURL string
}

func NewConflictError(url string) *ConflictError {
	return &ConflictError{ShortURL: url}
}

func (ce *ConflictError) Error() string {
	return ce.ShortURL
}

type Repository interface {
	Save(id, userID, url string) error
	Get(id string) (string, bool, bool)
	SaveBatch(items []repository.BatchItem) error
	GetUserURLs(userID string) ([]repository.URLRecord, error)
	DeleteBatch(userID string, shortURLs []string) error
}

type URLService struct {
	repository Repository
}

func NewURLServiсe(repository Repository) *URLService {
	return &URLService{
		repository: repository,
	}
}

func randomString(n int) string {
	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	randBytes := make([]byte, n)
	for i := range n {
		randBytes[i] = charset[rand.Intn(len(charset))]
	}
	return string(randBytes)
}

func (us *URLService) Shorten(url, userID string) (string, error) {
	numOfTryies := 5

	for i := 0; i < numOfTryies; i++ {
		id := randomString(lenOfGeneratedURL)
		saveError := us.repository.Save(id, userID, url)
		if saveError == nil {
			return id, nil
		}
		if errors.Is(saveError, repository.ErrIDConflict) {
			continue
		}
		var conflictErr *repository.ConflictError
		if errors.As(saveError, &conflictErr) {
			return "", NewConflictError(conflictErr.ShortURL)
		}
		return "", saveError
	}
	return "", ErrGeneratingIdFail
}

func (us *URLService) GetURL(id string) (URL string, isDeleted bool, ok bool) {
	URL, isDeleted, ok = us.repository.Get(id)
	return
}

func (us *URLService) ShortenBatch(items []model.BatchRequest, UserID string) ([]model.BatchResponse, error) {
	var batch []repository.BatchItem
	var response []model.BatchResponse

	for _, item := range items {
		var id string
		id = randomString(lenOfGeneratedURL)
		batch = append(batch, repository.BatchItem{ID: id, URL: item.OriginalURL, UserID: UserID})
		response = append(response, model.BatchResponse{CorrelationID: item.CorrelationID, ShortURL: id})
	}
	saveErr := us.repository.SaveBatch(batch)
	if saveErr != nil {
		return nil, saveErr
	}
	return response, nil
}

func (us *URLService) GetUserURLs(userID string) ([]model.UserURL, error) {
	records, err := us.repository.GetUserURLs(userID)
	if err != nil {
		return nil, err
	}
	result := make([]model.UserURL, 0, len(records))
	for _, record := range records {
		result = append(result, model.UserURL{ShortURL: record.ShortURL, OriginalURL: record.OriginalURL})
	}
	return result, nil
}
