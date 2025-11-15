package handler

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/Elissbar/go-shortener-url/internal/repository"
)

func getToken(ctx context.Context, storage repository.Storage) (string, error) {
	const maxAttempts = 5
	for at := 0; at < maxAttempts; at++ {
		token, err := generateToken()
		if err != nil {
			return "", err
		}

		// Проверяем, свободен ли токен
		_, exists := storage.Get(ctx, token)
		if !exists {
			return token, nil
		}
	}
	return "", fmt.Errorf("failed to generate unique token after %d attempts", maxAttempts)
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
