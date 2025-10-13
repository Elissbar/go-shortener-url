package handler

import (
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type MyHandler struct {
	Urls map[string]string
}

func (h *MyHandler) Router() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.CreateShortUrl)
	r.Get("/{id}", h.GetShortUrl)

	return r
}

func (h *MyHandler) CreateShortUrl(rw http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodPost {
		token, err := generateToken()
		if err != nil {
			http.Error(rw, "Error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		for _, ok := h.Urls[token]; ok; { // Если такой токен уже есть - генерируем новый
			token, _ = generateToken()
		}

		body, err := io.ReadAll(req.Body) // получаем URL для сокращения
		if err != nil {
			http.Error(rw, "Error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		h.Urls[token] = string(body)

		scheme := "http://"
		if req.TLS != nil {
			scheme = "https://"
		}

		rw.Header().Set("content-type", "text/plain")
		rw.WriteHeader(http.StatusCreated)
		rw.Write([]byte(scheme + req.Host + req.URL.Path + token))
	}
}

func (h *MyHandler) GetShortUrl(rw http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodGet {
		id := chi.URLParam(req, "id")
		url, ok := h.Urls[id]
		if !ok {
			rw.WriteHeader(http.StatusNotFound)
			rw.Write([]byte("Not Found"))
		} else {
			http.Redirect(rw, req, url, http.StatusTemporaryRedirect)
		}
	}
}
