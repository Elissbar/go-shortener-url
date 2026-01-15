package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/Elissbar/go-shortener-url/internal/config"
	"github.com/Elissbar/go-shortener-url/internal/logger"
	"github.com/Elissbar/go-shortener-url/internal/model"
	memorystorage "github.com/Elissbar/go-shortener-url/internal/repository/implementations/memory_storage"
	"github.com/Elissbar/go-shortener-url/internal/service"
	"github.com/stretchr/testify/require"
)

var myHandler MyHandler

func TestMain(m *testing.M) {
	cfg := &config.Config{
		ServerURL: "localhost:8080", 
		BaseURL: "http://localhost:8080/", 
		LogLevel: "info",
	}

	log, err := logger.NewSugaredLogger(cfg.LogLevel)
	if err != nil {
		panic(err)
	}
	defer log.Sync()

	storage, _ := memorystorage.NewMemoryStorage()
	myHandler = MyHandler{
		Storage: storage,
		Config:  cfg,
		Logger:  log,
		Service: service.NewService(log, storage, cfg),
	}

	code := m.Run()
	os.Exit(code)
}

func TestCreateShortUrl(t *testing.T) {
	type want struct {
		contentType string
		statusCode  int
	}

	tests := []struct {
		name    string
		request string
		want    want
	}{
		{
			name:    "Create short url",
			request: "https://practicum.yandex2.ru/",
			want: want{
				contentType: "text/plain",
				statusCode:  http.StatusCreated,
			},
		},
		{
			name:    "Create short url 2",
			request: "https://www.google2.com/",
			want: want{
				contentType: "text/plain",
				statusCode:  http.StatusCreated,
			},
		},
	}

	for _, tt := range tests {
		request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.request))
		w := httptest.NewRecorder()

		router := myHandler.Router()
		router.ServeHTTP(w, request)

		result := w.Result()
		defer result.Body.Close()

		require.Equal(t, tt.want.statusCode, result.StatusCode)
		require.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))
	}
}

func TestGetShortUrl(t *testing.T) {
	tests := []struct {
		name               string
		id                 string
		redirectTo         string
		expectedStatusCode int
	}{
		{
			name:               "Get shorted URL",
			id:                 "1c24X2zVQ7s",
			redirectTo:         "https://practicum.yandex.ru/",
			expectedStatusCode: http.StatusTemporaryRedirect,
		},
		{
			name:               "Get shorted URL 2",
			id:                 "C0JnW5wJNk4",
			redirectTo:         "https://www.google.com/",
			expectedStatusCode: http.StatusTemporaryRedirect,
		},
	}
	for _, tt := range tests {
		urls := &sync.Map{}
		urls.Store(tt.id, tt.redirectTo)
		myHandler.Storage = &memorystorage.MemoryStorage{TokenURL: urls, URLToken: &sync.Map{}}

		request := httptest.NewRequest(http.MethodGet, "/"+tt.id, nil)
		w := httptest.NewRecorder()

		router := myHandler.Router()
		router.ServeHTTP(w, request)

		result := w.Result()
		defer result.Body.Close()
		redirectedTo, _ := result.Location()

		require.Equal(t, tt.expectedStatusCode, result.StatusCode)
		require.Equal(t, tt.redirectTo, redirectedTo.String())
	}
}

func TestCreateShortUrlJSON(t *testing.T) {
	type want struct {
		contentType string
		statusCode  int
	}

	tests := []struct {
		name    string
		request string
		want    want
	}{
		{
			name:    "Create short url",
			request: "https://practicum.yandex1.ru/",
			want: want{
				contentType: "application/json",
				statusCode:  http.StatusCreated,
			},
		},
		{
			name:    "Create short url 2",
			request: "https://www.google1.com/",
			want: want{
				contentType: "application/json",
				statusCode:  http.StatusCreated,
			},
		},
	}

	for _, tt := range tests {
		data := map[string]string{}
		data["url"] = tt.request

		modifiedJSONBytes, err := json.Marshal(data)
		if err != nil {
			return
		}

		request := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(string(modifiedJSONBytes)))
		w := httptest.NewRecorder()

		router := myHandler.Router()
		router.ServeHTTP(w, request)

		result := w.Result()
		defer result.Body.Close()

		body, _ := io.ReadAll(result.Body)
		var resp model.Response
		json.Unmarshal(body, &resp)

		require.Equal(t, tt.want.statusCode, result.StatusCode)
		require.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))
		require.NotEmpty(t, resp.Result, "Result should not be empty")
	}
}
