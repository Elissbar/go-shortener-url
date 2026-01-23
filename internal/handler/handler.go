package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

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

		userID, ctx, cancel, err := prepareHandler(req)
		defer cancel()
		if err != nil {
			http.Error(rw, "Internal server error", http.StatusInternalServerError)
		}

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

		baseURL := getFullBaseURL(h.Service.Config.BaseURL)

		savedToken, err := h.Service.Storage.Save(ctx, token, rq.URL, userID, baseURL)
		if err != nil && errors.Is(err, repository.ErrURLExists) {
			rw.WriteHeader(http.StatusConflict)
		} else {
			rw.WriteHeader(http.StatusCreated)
		}

		var resp model.Response
		resp.Result = baseURL + savedToken

		data, err := json.Marshal(resp)
		if err != nil {
			http.Error(rw, "Error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		audit(h.Service.Event, "shorten", userID, rq.URL)

		rw.Write(data)
	}
}

// CreateShortBatch метод для обработки ссылок батчами.
// Пример запроса:
//
//	 [
//	   {"correlation_id": "123", "original_url": "https://practicum.yandex.ru/learn/go-advanced/courses/"},
//	   {"correlation_id": "456", "original_url": "https://practicum.yandex.ru2/learn2/go-advanced2/courses2/"},
//   ]
func (h *MyHandler) CreateShortBatch(rw http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodPost {
		rw.Header().Set("Content-Type", "application/json")

		userID, ctx, cancel, err := prepareHandler(req)
		defer cancel()
		if err != nil {
			http.Error(rw, "Internal server error", http.StatusInternalServerError)
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

		baseURL := getFullBaseURL(h.Service.Config.BaseURL)

		respBatch := make([]model.RespBatch, 0, len(reqBatch))
		for i := range len(reqBatch) {
			batch := &reqBatch[i]
			token, err := h.Service.GetToken(ctx) // 2.64MB
			if err != nil {
				http.Error(rw, "Error: "+err.Error(), http.StatusInternalServerError)
				return
			}

			shortedURL := baseURL + token
			batch.Token = token
			respBatch = append(respBatch, model.RespBatch{ID: batch.ID, ShortURL: shortedURL})
		}

		data, err := json.Marshal(respBatch) // 29 sec
		if err != nil {
			http.Error(rw, "Error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		err = h.Service.Storage.SaveBatch(ctx, reqBatch, userID, baseURL)
		if err != nil {
			http.Error(rw, "Error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusCreated)
		rw.Write(data)
	}
}

// CreateShortURL принимает данные в формате text/plain и сокращает URL.
func (h *MyHandler) CreateShortURL(rw http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodPost {
		rw.Header().Set("content-type", "text/plain")

		userID, ctx, cancel, err := prepareHandler(req)
		defer cancel()
		if err != nil {
			http.Error(rw, "Internal server error", http.StatusInternalServerError)
		}

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

		baseURL := getFullBaseURL(h.Service.Config.BaseURL)

		savedToken, err := h.Service.Storage.Save(ctx, token, string(body), userID, baseURL)
		if err != nil {
			if errors.Is(err, repository.ErrURLExists) {
				rw.WriteHeader(http.StatusConflict)
			}
		} else {
			rw.WriteHeader(http.StatusCreated)
		}

		shortedURL := baseURL + savedToken
		rw.Write([]byte(shortedURL))

		audit(h.Service.Event, "shorten", userID, string(body))
	}
}

// GetShortURL возвращает сокращённый URL.
func (h *MyHandler) GetShortURL(rw http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		return
	}

	userID, ctx, cancel, err := prepareHandler(req)
	defer cancel()
	if err != nil {
		http.Error(rw, "Internal server error", http.StatusInternalServerError)
	}

	id := chi.URLParam(req, "id")
	h.Service.Logger.Infow("GET request for token", "token", id)

	url, err := h.Service.Storage.Get(ctx, id)
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

	audit(h.Service.Event, "follow", userID, url)

	h.Service.Logger.Infow("Redirecting token", "token", id, "url", url)
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
	userID, ctx, cancel, err := prepareHandler(req)
	defer cancel()
	if err != nil {
		http.Error(rw, "Internal server error", http.StatusInternalServerError)
	}

	records, err := h.Service.Storage.GetAllUsersURLs(ctx, userID)
	if err != nil {
		http.Error(rw, "Internal server error", http.StatusInternalServerError)
		return
	}

	if len(records) == 0 {
		rw.WriteHeader(http.StatusNoContent)
	}

	data, err := json.Marshal(records)
	if err != nil {
		http.Error(rw, "Error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	rw.Header().Set("Content-Type", "application/json")
	rw.Write(data)
}

// DeleteURLs принимает список токенов и удаляет их.
func (h *MyHandler) DeleteURLs(rw http.ResponseWriter, req *http.Request) {
	userID, ok := req.Context().Value(userIDKey).(string)
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
