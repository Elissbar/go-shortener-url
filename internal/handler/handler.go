package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Elissbar/go-shortener-url/internal/config"
	"github.com/Elissbar/go-shortener-url/internal/model"
	"github.com/Elissbar/go-shortener-url/internal/repository"
	"github.com/Elissbar/go-shortener-url/internal/service"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type MyHandler struct {
	Storage repository.Storage
	Config  *config.Config
	Logger  *zap.SugaredLogger
	Service *service.Service
}

func NewHandler(storage repository.Storage, cfg *config.Config, log *zap.SugaredLogger, srvc *service.Service) *MyHandler {
	myHandler := &MyHandler{
		Storage: storage,
		Config:  cfg,
		Logger:  log,
		Service: srvc,
	}
	return myHandler
}

func (h *MyHandler) Router() chi.Router {
	r := chi.NewRouter()

	r.Use(h.LoggingMiddleware)
	r.Use(h.authentication)
	r.Use(ungzipMiddleware)
	r.Use(gzipMiddleware)

	r.Post("/", h.CreateShortURL)
	r.Post("/api/shorten", h.CreateShortURLJSON)
	r.Post("/api/shorten/batch", h.CreateShortBatch)
	r.Get("/{id}", h.GetShortURL)
	r.Get("/", h.GetRoot)
	r.Get("/ping", h.CheckConnectionDB)
	r.Get("/api/user/urls", h.GetAllUserURLs)
	r.Delete("/api/user/urls", h.DeleteURLs)

	return r
}

func (h *MyHandler) GetRoot(rw http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodGet {
		rw.Write([]byte("URL Shortener is running!"))
	}
}

func (h *MyHandler) CreateShortURLJSON(rw http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodPost {
		userID, ok := req.Context().Value(userIDKey).(string)
		if !ok {
			http.Error(rw, "Internal server error", http.StatusInternalServerError)
			return
		}
		rw.Header().Set("Content-Type", "application/json")

		ctx := req.Context()
		ctx, cancel := context.WithTimeout(ctx, time.Second*3)
		defer cancel()

		token, err := h.Service.GetToken(ctx)
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

		baseURL := h.Config.BaseURL
		if !strings.HasSuffix(h.Config.BaseURL, "/") {
			baseURL = h.Config.BaseURL + "/"
		}

		savedToken, err := h.Storage.Save(ctx, token, rq.URL, userID, baseURL)
		if err != nil && errors.Is(err, repository.ErrURLExists) {
			rw.WriteHeader(http.StatusConflict)
		} else {
			rw.WriteHeader(http.StatusCreated)
		}

		var resp model.Response
		resp.Result = baseURL + savedToken

		enc := json.NewEncoder(rw)
		if err := enc.Encode(resp); err != nil {
			http.Error(rw, "Error: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (h *MyHandler) CreateShortBatch(rw http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodPost {
		userID, ok := req.Context().Value(userIDKey).(string)
		if !ok {
			http.Error(rw, "Internal server error", http.StatusInternalServerError)
			return
		}

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

		baseURL := h.Config.BaseURL
		if !strings.HasSuffix(h.Config.BaseURL, "/") {
			baseURL = h.Config.BaseURL + "/"
		}

		var respBatch []model.RespBatch
		for i := range len(reqBatch) {
			batch := &reqBatch[i]
			token, err := h.Service.GetToken(ctx)
			if err != nil {
				http.Error(rw, "Error: "+err.Error(), http.StatusInternalServerError)
				return
			}

			shortedURL := baseURL + token
			batch.Token = token
			respBatch = append(respBatch, model.RespBatch{ID: batch.ID, ShortURL: shortedURL})
		}

		h.Storage.SaveBatch(ctx, reqBatch, userID, baseURL)

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
		userID, ok := req.Context().Value(userIDKey).(string)
		if !ok {
			http.Error(rw, "Internal server error", http.StatusInternalServerError)
			return
		}
		rw.Header().Set("content-type", "text/plain")

		ctx := req.Context()
		ctx, cancel := context.WithTimeout(ctx, time.Second*3)
		defer cancel()

		token, err := h.Service.GetToken(ctx)
		if err != nil {
			http.Error(rw, "Error 1: "+err.Error(), http.StatusInternalServerError)
			return
		}

		body, err := io.ReadAll(req.Body) // получаем URL для сокращения
		if err != nil {
			http.Error(rw, "Error 2: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer req.Body.Close()

		baseURL := h.Config.BaseURL
		if !strings.HasSuffix(h.Config.BaseURL, "/") {
			baseURL = h.Config.BaseURL + "/"
		}

		savedToken, err := h.Storage.Save(ctx, token, string(body), userID, baseURL)
		if err != nil {
			if errors.Is(err, repository.ErrURLExists) {
				rw.WriteHeader(http.StatusConflict)
			}
		} else {
			rw.WriteHeader(http.StatusCreated)
		}

		shortedURL := baseURL + savedToken
		rw.Write([]byte(shortedURL))
	}
}

func (h *MyHandler) GetShortURL(rw http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		return
	}

	id := chi.URLParam(req, "id")
	h.Logger.Infow("GET request for token", "token", id)

	ctx, cancel := context.WithTimeout(req.Context(), 3*time.Second)
	defer cancel()

	url, err := h.Storage.Get(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrTokenNotExist) {
			h.Logger.Infow("Token not found", "token", id)
			rw.WriteHeader(http.StatusNotFound)
			rw.Write([]byte("Not Found"))
		} else if errors.Is(err, repository.ErrTokenIsDeleted) {
			h.Logger.Infow("Token deleted (410)", "token", id)
			rw.WriteHeader(http.StatusGone)
			rw.Write([]byte("Gone"))
		} else {
			h.Logger.Errorw("Error getting token", "token", id, "error", err)
			http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	h.Logger.Infow("Redirecting token", "token", id, "url", url)
	http.Redirect(rw, req, url, http.StatusTemporaryRedirect)
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

func (h *MyHandler) GetAllUserURLs(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	userID, ok := req.Context().Value(userIDKey).(string)
	if !ok {
		http.Error(rw, "Internal server error", http.StatusInternalServerError)
		return
	}

	ctx := req.Context()
	ctx, cancel := context.WithTimeout(ctx, time.Second*3)
	defer cancel()

	records, err := h.Storage.GetAllUsersURLs(ctx, userID)
	if err != nil {
		http.Error(rw, "Internal server error", http.StatusInternalServerError)
		return
	}

	if len(records) == 0 {
		rw.WriteHeader(http.StatusNoContent)
	}

	enc := json.NewEncoder(rw)
	if err := enc.Encode(records); err != nil {
		http.Error(rw, "Error: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *MyHandler) DeleteURLs(rw http.ResponseWriter, req *http.Request) {
	userID, ok := req.Context().Value(userIDKey).(string)
	if !ok || userID == "" {
		http.Error(rw, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	var tokens []string
	if err := json.NewDecoder(req.Body).Decode(&tokens); err != nil {
		http.Error(rw, "Invalid JSON", http.StatusBadRequest)
		return
	}
	defer req.Body.Close()
	
	// Валидация токенов
	if len(tokens) == 0 {
		rw.WriteHeader(http.StatusAccepted)
		return
	}
	
	// Создаем запрос
	deleteReq := service.DeleteRequest{
		UserID: userID,
		Tokens: tokens,
	}
	
	timeout := time.After(100 * time.Millisecond)
	select {
	case h.Service.DeleteCh <- deleteReq:
		rw.WriteHeader(http.StatusAccepted)
	case <-timeout:
		// Если канал полон, ждем с таймаутом
		select {
		case h.Service.DeleteCh <- deleteReq:
			rw.WriteHeader(http.StatusAccepted)
		case <-time.After(500 * time.Millisecond):
			http.Error(rw, "Service busy", http.StatusServiceUnavailable)
		}
	}
}
