package service

import (
	"errors"
	"math/rand"

	"github.com/truenuta/urlshortener/internal/model"
	"github.com/truenuta/urlshortener/internal/repository"
)

var ErrGeneratingIdFail = errors.New("Failed generating unique id")
var LenOfGeneratedUrl int = 8

type Service interface {
	Shorten(url string) (string, error)
	GetURL(id string) (URL string, ok bool)
	ShortenBatch(items []model.BatchRequest) ([]model.BatchResponse, error)
}

type Repository interface {
	Save(id, url string) error
	Get(id string) (string, bool)
	SaveBatch(items []repository.BatchItem) error
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

func (us *URLService) Shorten(url string) (string, error) {
	numOfTryies := 5
	for i := 0; i < numOfTryies; i++ {
		id := randomString(LenOfGeneratedUrl)
		saveError := us.repository.Save(id, url)
		if saveError == nil {
			return id, nil
		}
		if errors.Is(saveError, repository.ErrIDConflict) {
			continue
		}
		return "", saveError
	}
	return "", ErrGeneratingIdFail
}

func (us *URLService) GetURL(id string) (URL string, ok bool) {
	URL, ok = us.repository.Get(id)
	return
}

func (us *URLService) ShortenBatch(items []model.BatchRequest) ([]model.BatchResponse, error) {
	var batch []repository.BatchItem
	var response []model.BatchResponse

	for _, item := range items {
		var id string
		id = randomString(LenOfGeneratedUrl)
		batch = append(batch, repository.BatchItem{ID: id, URL: item.OriginalURL})
		response = append(response, model.BatchResponse{CorrelationID: item.CorrelationID, ShortURL: id})
	}
	saveErr := us.repository.SaveBatch(batch)
	if saveErr != nil {
		return nil, saveErr
	}
	return response, nil
}
