package service

import (
	"crypto/rand"
	"encoding/base64"
	"strings"
	"time"

	"github.com/Elissbar/go-shortener-url/internal/model"
	"github.com/Elissbar/go-shortener-url/internal/observer"
)

func audit(event *observer.Event, action, userID, url string) {
	event.Update(model.AuditRequest{
		TS:     time.Now().Unix(),
		Action: action,
		UserID: userID,
		URL:    url,
	})
}

func getFullBaseURL(baseURL string) string {
	fullURL := baseURL
	if !strings.HasSuffix(baseURL, "/") {
		fullURL = baseURL + "/"
	}
	return fullURL
}

func GenerateToken(size int) (string, error) {
	// Генерируем токен - id короткой ссылки
	if size <= 0 {
		return "", nil
	}
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	token := base64.URLEncoding.EncodeToString(b)
	token = token[:len(token)-1]
	return token, nil
}
