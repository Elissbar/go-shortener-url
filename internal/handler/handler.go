package handler

import (
	"io"
	"net/http"
)

type MyHandler struct {
	Urls map[string]string
}

func (h *MyHandler) CreateShortUrl(rw http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodPost {
		urls := h.Urls

		token, err := generateToken()
		if err != nil {
			http.Error(rw, "Error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		for _, ok := urls[token]; ok; { // Если такой токен уже есть - генерируем новый
			token, _ = generateToken()
		}

		body, err := io.ReadAll(req.Body) // получаем URL для сокращения
		if err != nil {
			http.Error(rw, "Error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		urls[token] = string(body)

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
		urls := h.Urls
		url, ok := urls[req.URL.Path[1:]]
		if !ok {
			rw.WriteHeader(http.StatusNotFound)
			rw.Write([]byte("Not Found"))
		} else {
			http.Redirect(rw, req, url, http.StatusTemporaryRedirect)
		}
	}
}
