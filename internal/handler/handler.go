package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

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
	r.Use(ungzipMiddleware)
	r.Use(gzipMiddleware)

	r.Post("/", h.CreateShortURL)
	r.Post("/api/shorten", h.CreateShortURLJSON)
	r.Post("/api/shorten/batch", h.CreateShortBatch)
	r.Get("/{id}", h.GetShortURL)
	r.Get("/", h.GetRoot)
	r.Get("/ping", h.CheckConnectionDB)

	return r
}

func (h *MyHandler) GetRoot(rw http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodGet {
		rw.Write([]byte("URL Shortener is running!"))
	}
}

func (h *MyHandler) CreateShortURLJSON(rw http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodPost {
		rw.Header().Set("Content-Type", "application/json")

		ctx := req.Context()
		ctx, cancel := context.WithTimeout(ctx, time.Second*3)
		defer cancel()

		token, err := getToken(ctx, h.Storage)
		if err != nil {
			http.Error(rw, "Error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		var rq model.Request
		dec := json.NewDecoder(req.Body)
		if err := dec.Decode(&rq); err != nil {
			http.Error(rw, "Error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer req.Body.Close()

		var resp model.Response

		savedToken, err := h.Storage.Save(ctx, token, rq.URL)
		if err != nil && errors.Is(err, repository.ErrURLExists) {
			rw.WriteHeader(http.StatusConflict)
		} else {
			rw.WriteHeader(http.StatusCreated)
		}

		resp.Result = h.Config.BaseURL + savedToken
		if !strings.HasSuffix(h.Config.BaseURL, "/") {
			resp.Result = h.Config.BaseURL + "/" + savedToken
		}

		enc := json.NewEncoder(rw)
		if err := enc.Encode(resp); err != nil {
			http.Error(rw, "Error: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (h *MyHandler) CreateShortBatch(rw http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodPost {
		ctx := req.Context()
		ctx, cancel := context.WithTimeout(ctx, time.Second*3)
		defer cancel()
		defer req.Body.Close()

		var reqBatch []model.ReqBatch
		dec := json.NewDecoder(req.Body)
		if err := dec.Decode(&reqBatch); err != nil {
			http.Error(rw, "Error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		var respBatch []model.RespBatch
		for i := range len(reqBatch) {
			batch := &reqBatch[i]
			token, err := getToken(ctx, h.Storage)
			if err != nil {
				http.Error(rw, "Error: "+err.Error(), http.StatusInternalServerError)
				return
			}

			shortedURL := h.Config.BaseURL + token
			if !strings.HasSuffix(h.Config.BaseURL, "/") {
				shortedURL = h.Config.BaseURL + "/" + token
			}
			batch.Token = token
			respBatch = append(respBatch, model.RespBatch{ID: batch.ID, ShortURL: shortedURL})
		}

		h.Storage.SaveBatch(ctx, reqBatch)

		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusCreated)

		enc := json.NewEncoder(rw)
		if err := enc.Encode(respBatch); err != nil {
			http.Error(rw, "Error: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (h *MyHandler) CreateShortURL(rw http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodPost {
		rw.Header().Set("content-type", "text/plain")

		ctx := req.Context()
		ctx, cancel := context.WithTimeout(ctx, time.Second*3)
		defer cancel()

		token, err := getToken(ctx, h.Storage)
		if err != nil {
			http.Error(rw, "Error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		body, err := io.ReadAll(req.Body) // получаем URL для сокращения
		if err != nil {
			http.Error(rw, "Error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer req.Body.Close()

		savedToken, err := h.Storage.Save(ctx, token, string(body))
		if err != nil {
			fmt.Println(err)
			if errors.Is(err, repository.ErrURLExists) {
				rw.WriteHeader(http.StatusConflict)
			}
		} else {
			rw.WriteHeader(http.StatusCreated)
		}

		shortedURL := h.Config.BaseURL + savedToken
		if !strings.HasSuffix(h.Config.BaseURL, "/") {
			shortedURL = h.Config.BaseURL + "/" + savedToken
		}
		rw.Write([]byte(shortedURL))
	}
}

func (h *MyHandler) GetShortURL(rw http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodGet {
		ctx := req.Context()
		ctx, cancel := context.WithTimeout(ctx, time.Second*3)
		defer cancel()

		id := chi.URLParam(req, "id")
		url, ok := h.Storage.Get(ctx, id)
		if !ok {
			rw.WriteHeader(http.StatusNotFound)
			rw.Write([]byte("Not Found"))
		} else {
			http.Redirect(rw, req, url, http.StatusTemporaryRedirect)
		}
	}
}

func (h *MyHandler) CheckConnectionDB(rw http.ResponseWriter, req *http.Request) {
	if err := h.Storage.Ping(); err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte("Database connection is not success"))
		return
	}

	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte("Database connection is success"))
}
