package main

import (
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

var storage = make(map[string]string)

func shorten(body string) string {
	id := uuid.New().String() // заменить на нормальный метод
	storage[id] = body
	return id
}

func ShortenRequest(response http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost && request.Method != http.MethodGet {
		response.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if request.Method == http.MethodPost {
		body, _ := io.ReadAll(request.Body)
		defer request.Body.Close()
		url := shorten(string(body))
		response.Header().Set("content-type", "text/plain")
		response.WriteHeader(http.StatusCreated)
		response.Write([]byte(url))
	}
	if request.Method == http.MethodGet {
		id := strings.TrimPrefix(request.URL.Path, "/")
		original_url := storage[id]
		response.Header().Set("Location", original_url)
		response.WriteHeader(http.StatusTemporaryRedirect)
	}
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", ShortenRequest)
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}

}
