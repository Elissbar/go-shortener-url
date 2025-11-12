package handler

import (
	"context"
	"crypto/rand"
	"encoding/base64"

	"github.com/Elissbar/go-shortener-url/internal/repository"
)

func getToken(ctx context.Context, storage repository.Storage) (string, error) {
	token, err := generateToken()
	if err != nil {
		return "", err
	}
	for _, ok := storage.Get(ctx, token); ok; { // Если такой токен уже есть - генерируем новый
		token, _ = generateToken()
	}
	return token, nil
}

func generateToken() (string, error) {
	// Генерируем токен - id короткой ссылки
	b := make([]byte, 8)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	token := base64.URLEncoding.EncodeToString(b)
	token = token[:len(token)-1]
	return token, nil
}
