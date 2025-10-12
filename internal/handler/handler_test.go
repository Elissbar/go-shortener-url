package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

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
			request: "https://practicum.yandex.ru/",
			want: want{
				contentType: "text/plain",
				statusCode:  http.StatusCreated,
			},
		},
		{
			name:    "Create short url 2",
			request: "https://www.google.com/",
			want: want{
				contentType: "text/plain",
				statusCode:  http.StatusCreated,
			},
		},
	}

	for _, tt := range tests {
		request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.request))
		w := httptest.NewRecorder()

		urls := make(map[string]string)
		myHandler := MyHandler{Urls: urls}
		myHandler.CreateShortUrl(w, request)

		result := w.Result()

		require.Equal(t, result.StatusCode, tt.want.statusCode)
		require.Equal(t, result.Header.Get("Content-Type"), tt.want.contentType)
	}
}

func TestGetShortUrl(t *testing.T) {
	type want struct {
		statusCode int
	}

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
		urls := make(map[string]string)
		urls[tt.id] = tt.redirectTo

		request := httptest.NewRequest(http.MethodGet, "/"+tt.id, nil)
		w := httptest.NewRecorder()

		myHandler := MyHandler{Urls: urls}
		myHandler.GetShortUrl(w, request)

		result := w.Result()
		redirectedTo, _ := result.Location()

		require.Equal(t, result.StatusCode, tt.expectedStatusCode)
		require.Equal(t, redirectedTo.String(), tt.redirectTo)
	}
}
