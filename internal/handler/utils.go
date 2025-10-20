package handler

import (
	"crypto/rand"
	"encoding/base64"
)

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
