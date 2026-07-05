package service

import (
	"math/rand"

	"github.com/truenuta/urlshortener/internal/repository"
)

type Service interface {
	Shorten(url string) string
	GetURL(id string) (URL string, ok bool)
}

type URLService struct {
	repository repository.Repository
}

func NewURLServiсe(repository repository.Repository) *URLService {
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

func (us *URLService) Shorten(url string) string {
	LenOfGeneratedUrl := 8
	id := randomString(LenOfGeneratedUrl)
	for {
		if _, exists := us.repository.Get(id); !exists {
			us.repository.Save(id, url)
			return id
		}
	}
}

func (us *URLService) GetURL(id string) (URL string, ok bool) {
	URL, ok = us.repository.Get(id)
	return
}
