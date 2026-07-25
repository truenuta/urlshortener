package service

import (
	"errors"
	"math/rand"

	"github.com/truenuta/urlshortener/internal/repository"
)

var ErrGeneratingIdFail = errors.New("Failed generating unique id")

type Service interface {
	Shorten(url string) (string, error)
	GetURL(id string) (URL string, ok bool)
}

type Repository interface {
	Save(id, url string) error
	Get(id string) (string, bool)
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
	LenOfGeneratedUrl := 8
	id := randomString(LenOfGeneratedUrl)
	numOfTryies := 5
	for i := 0; i < numOfTryies; i++ {
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
