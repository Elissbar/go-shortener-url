package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/Elissbar/go-shortener-url/internal/handler/common"
	"github.com/Elissbar/go-shortener-url/internal/model"
	"github.com/Elissbar/go-shortener-url/internal/repository"
	"github.com/Elissbar/go-shortener-url/internal/service"
)

// MyHandler тип через который регистрируются обработчики.
type MyHandler struct {
	Service *service.Service
}

func NewHandler(srvc *service.Service) *MyHandler {
	return &MyHandler{
		Service: srvc,
	}
}

// Router метод для регистрации обработчиков.
func (h *MyHandler) Router() chi.Router {
	r := chi.NewRouter()

	r.Use(loggingMiddleware(h.Service.Logger))
	r.Use(authentication(h.Service.Config.JWTSecret))
	r.Use(ungzipMiddleware)
	r.Use(gzipMiddleware)

	r.Post("/", h.CreateShortURL)
	r.Post("/api/shorten", h.CreateShortURLJSON)
	r.Post("/api/shorten/batch", h.CreateShortBatch)
	r.Get("/{id}", h.GetShortURL)
	r.Get("/", h.GetRoot)
	r.Get("/ping", h.CheckConnectionDB)
	r.Get("/api/user/urls", h.GetAllUserURLs)
	r.Get("/api/internal/stats", h.GetStats)
	r.Delete("/api/user/urls", h.DeleteURLs)

	return r
}

func (h *MyHandler) GetRoot(rw http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodGet {
		rw.Write([]byte("URL Shortener is running!"))
	}
}

func (h *MyHandler) GetStats(rw http.ResponseWriter, req *http.Request) {
	if h.Service.Config.TrustedSubnet == "" {
		rw.WriteHeader(http.StatusUnauthorized)
		return
	}

	realIP := req.Header.Get("X-Real-IP")
	data, err := h.Service.GetStats(req.Context(), realIP)
	if err != nil {
		if errors.Is(err, repository.ErrSubnetForbidden) {
			rw.WriteHeader(http.StatusForbidden)
			return
		}
		http.Error(rw, "Error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.Write(data)
}

// @Summary Запрос для сокращения ссылки.
// @ID CreateShortURLJSON
// @Product json
// @Param request body model.Request true "Request"
// @Success 201 {object} model.Response
// @Failure 409 {string} string "Conflict"
// @Failure 500 {string} string "Internal server error"
// @Router /api/shorten [post]
// CreateShortURLJSON обработчик для создания короткой ссылки, принимает данные в формате JSON.
// Пример запроса:
//
//	{"url": "https://practicum.yandex.ru/learn/go-advanced/"}
func (h *MyHandler) CreateShortURLJSON(rw http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodPost {
		rw.Header().Set("Content-Type", "application/json")

		ctx, cancel := context.WithTimeout(req.Context(), time.Second*3)
		defer cancel()
		userID, ok := req.Context().Value(common.UserIDKey).(string)
		if !ok {
			http.Error(rw, "userID type error", http.StatusInternalServerError)
			return
		}

		var rq model.Request
		dec := json.NewDecoder(req.Body)
		if err := dec.Decode(&rq); err != nil {
			http.Error(rw, "Error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer req.Body.Close()

		resp, err := h.Service.CreateShortURLJSON(ctx, rq, userID)
		if err != nil && errors.Is(err, repository.ErrURLExists) {
			rw.WriteHeader(http.StatusConflict)
		} else {
			rw.WriteHeader(http.StatusCreated)
		}

		data, err := json.Marshal(resp)
		if err != nil {
			http.Error(rw, "Error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		rw.Write(data)
	}
}

// CreateShortBatch метод для обработки ссылок батчами.
// Пример запроса:
//
//		 [
//		   {"correlation_id": "123", "original_url": "https://practicum.yandex.ru/learn/go-advanced/courses/"},
//		   {"correlation_id": "456", "original_url": "https://practicum.yandex.ru2/learn2/go-advanced2/courses2/"},
//	  ]
func (h *MyHandler) CreateShortBatch(rw http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodPost {
		rw.Header().Set("Content-Type", "application/json")

		ctx, cancel := context.WithTimeout(req.Context(), time.Second*3)
		defer cancel()
		userID, ok := req.Context().Value(common.UserIDKey).(string)
		if !ok {
			http.Error(rw, "userID type error", http.StatusInternalServerError)
			return
		}

		body, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(rw, "Error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer req.Body.Close()

		var reqBatch []model.ReqBatch
		err = json.Unmarshal(body, &reqBatch)
		if err != nil {
			http.Error(rw, "Error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		data, err := h.Service.CreateShortBatch(ctx, reqBatch, userID)
		if err != nil {
			http.Error(rw, "Error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		rw.WriteHeader(http.StatusCreated)
		rw.Write(data)
	}
}

// CreateShortURL принимает данные в формате text/plain и сокращает URL.
func (h *MyHandler) CreateShortURL(rw http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodPost {
		rw.Header().Set("content-type", "text/plain")

		ctx, cancel := context.WithTimeout(req.Context(), time.Second*3)
		defer cancel()
		userID, ok := req.Context().Value(common.UserIDKey).(string)
		if !ok {
			http.Error(rw, "userID type error", http.StatusInternalServerError)
			return
		}

		body, err := io.ReadAll(req.Body) // получаем URL для сокращения
		if err != nil {
			http.Error(rw, "Error 2: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer req.Body.Close()

		shortedURL, err := h.Service.CreateShortURL(ctx, body, userID)
		if err != nil {
			if errors.Is(err, repository.ErrURLExists) {
				rw.WriteHeader(http.StatusConflict)
			} else {
				http.Error(rw, "Error: "+err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			rw.WriteHeader(http.StatusCreated)
		}
		rw.Write([]byte(shortedURL))
	}
}

// GetShortURL возвращает сокращённый URL.
func (h *MyHandler) GetShortURL(rw http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithTimeout(req.Context(), time.Second*3)
	defer cancel()
	userID, ok := req.Context().Value(common.UserIDKey).(string)
	if !ok {
		http.Error(rw, "userID type error", http.StatusInternalServerError)
		return
	}

	id := chi.URLParam(req, "id")
	url, err := h.Service.GetShortURL(ctx, id, userID)
	if err != nil {
		if errors.Is(err, repository.ErrTokenNotExist) {
			h.Service.Logger.Infow("Token not found", "token", id)
			rw.WriteHeader(http.StatusNotFound)
			rw.Write([]byte("Not Found"))
		} else if errors.Is(err, repository.ErrTokenIsDeleted) {
			h.Service.Logger.Infow("Token deleted (410)", "token", id)
			rw.WriteHeader(http.StatusGone)
			rw.Write([]byte("Gone"))
		} else {
			h.Service.Logger.Errorw("Error getting token", "token", id, "error", err)
			http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	http.Redirect(rw, req, url, http.StatusTemporaryRedirect)
}

// CheckConnectionDB проверяет соединение с базой данных.
func (h *MyHandler) CheckConnectionDB(rw http.ResponseWriter, req *http.Request) {
	if err := h.Service.Helper.Ping(); err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte("Database connection is not success"))
		return
	}

	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte("Database connection is success"))
}

// GetAllUserURLs возвращает все сокращённые URL пользователя.
func (h *MyHandler) GetAllUserURLs(rw http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithTimeout(req.Context(), time.Second*3)
	defer cancel()
	userID, ok := req.Context().Value(common.UserIDKey).(string)
	if !ok {
		http.Error(rw, "userID type error", http.StatusInternalServerError)
		return
	}

	records, err := h.Service.GetAllUserURLs(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrUserHasNoURL) {
			rw.WriteHeader(http.StatusNoContent)
			return
		}
		http.Error(rw, "Internal server error", http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(records)
	if err != nil {
		http.Error(rw, "Internal server error", http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.Write(data)
}

// DeleteURLs принимает список токенов и удаляет их.
func (h *MyHandler) DeleteURLs(rw http.ResponseWriter, req *http.Request) {
	userID, ok := req.Context().Value(common.UserIDKey).(string)
	if !ok || userID == "" {
		http.Error(rw, "Unauthorized", http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(rw, "Error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer req.Body.Close()

	var tokens []string
	err = json.Unmarshal(body, &tokens)
	if err != nil {
		http.Error(rw, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Валидация токенов
	err = h.Service.DeleteURLs(userID, tokens)
	if err != nil {
		http.Error(rw, "Error: "+err.Error(), http.StatusServiceUnavailable)
		return
	}
	rw.WriteHeader(http.StatusAccepted)
}
