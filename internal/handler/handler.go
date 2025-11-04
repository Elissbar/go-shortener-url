package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/Elissbar/go-shortener-url/internal/config"
	"github.com/Elissbar/go-shortener-url/internal/model"
	"github.com/Elissbar/go-shortener-url/internal/repository"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type MyHandler struct {
	Storage repository.Storage
	Config  *config.Config
	Logger  *zap.SugaredLogger
}

func (h *MyHandler) Router() chi.Router {
	r := chi.NewRouter()

	r.Use(h.LoggingMiddleware)
	r.Post("/", h.CreateShortUrl)
	r.Post("/api/shorten", h.CreateShortUrlJSON)
	r.Get("/{id}", h.GetShortUrl)
	r.Get("/", h.GetRoot)

	return r
}

func (h *MyHandler) GetRoot(rw http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodGet {
		rw.Write([]byte("URL Shortener is running!"))
	}
}

func (h *MyHandler) CreateShortUrlJSON(rw http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodPost {
		rw.Header().Set("Сontent-Type", "application/json")

		token, err := generateToken()
		if err != nil {
			http.Error(rw, "Error: "+err.Error(), http.StatusInternalServerError)
			json.NewEncoder(rw).Encode(map[string]string{"error": "invalid JSON"})
			return
		}
		for _, ok := h.Storage.Get(token); ok; { // Если такой токен уже есть - генерируем новый
			token, _ = generateToken()
		}

		var rq model.Request
		dec := json.NewDecoder(req.Body)

		if err := dec.Decode(&rq); err != nil {
			http.Error(rw, "Error: "+err.Error(), http.StatusInternalServerError)
			json.NewEncoder(rw).Encode(map[string]string{"error": "invalid JSON"})
			return
		}

		h.Storage.Save(token, rq.URL)

		var resp model.Response
		resp.Result = h.Config.BaseURL + token
		if !strings.HasSuffix(h.Config.BaseURL, "/") {
			resp.Result = h.Config.BaseURL + "/" + token
		}

		rw.WriteHeader(http.StatusCreated)

		enc := json.NewEncoder(rw)
		if err := enc.Encode(resp); err != nil {
			http.Error(rw, "Error: "+err.Error(), http.StatusInternalServerError)
			json.NewEncoder(rw).Encode(map[string]string{"error": "invalid JSON"})
			return
		}
	}
}

func (h *MyHandler) CreateShortUrl(rw http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodPost {
		token, err := generateToken()
		if err != nil {
			http.Error(rw, "Error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		for _, ok := h.Storage.Get(token); ok; { // Если такой токен уже есть - генерируем новый
			token, _ = generateToken()
		}

		body, err := io.ReadAll(req.Body) // получаем URL для сокращения
		if err != nil {
			http.Error(rw, "Error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		h.Storage.Save(token, string(body))

		rw.Header().Set("content-type", "text/plain")
		rw.WriteHeader(http.StatusCreated)

		shortedUrl := h.Config.BaseURL + token
		if !strings.HasSuffix(h.Config.BaseURL, "/") {
			shortedUrl = h.Config.BaseURL + "/" + token
		}
		rw.Write([]byte(shortedUrl))
	}
}

func (h *MyHandler) GetShortUrl(rw http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodGet {
		id := chi.URLParam(req, "id")
		url, ok := h.Storage.Get(id)
		if !ok {
			rw.WriteHeader(http.StatusNotFound)
			rw.Write([]byte("Not Found"))
		} else {
			http.Redirect(rw, req, url, http.StatusTemporaryRedirect)
		}
	}
}
