package handler

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"time"

	"github.com/Elissbar/go-shortener-url/internal/repository"
)

func getToken(ctx context.Context, storage repository.Storage) (string, error) {
	const maxAttempts = 5
	var token string

	for at := 0; at < maxAttempts; at++ {
		token, err := generateToken(8)
		if err != nil {
			return "", err
		}

		// Проверяем, свободен ли токен
		_, err = storage.Get(ctx, token)
		if err == sql.ErrNoRows {
			return token, nil
		} else if err != nil {
			return "", err
		}
	}
	return token, nil
}

func generateToken(size int) (string, error) {
	// Генерируем токен - id короткой ссылки
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	token := base64.URLEncoding.EncodeToString(b)
	token = token[:len(token)-1]
	return token, nil
}

func (h *MyHandler) processDeletions() {
	// Создаем воркеры для Fan In
	const numWorkers = 3
	workerChs := make([]chan string, numWorkers)

	// Запускаем воркеры
	for i := 0; i < numWorkers; i++ {
		workerChs[i] = make(chan string, 100)
		go h.deleteWorker(workerChs[i])
	}

	// Fan In: читаем из основного канала и распределяем по воркерам
	for tokensBatch := range h.DeleteCh {
		for i, token := range tokensBatch {
			workerIndex := i % numWorkers
			workerChs[workerIndex] <- token
		}
	}

	// Закрываем каналы воркеров при завершении
	for _, ch := range workerChs {
		close(ch)
	}
}

func (h *MyHandler) deleteWorker(tokenCh chan string) {
	buffer := make([]string, 0, 5) // буфер на 50 токенов

	for token := range tokenCh {
		buffer = append(buffer, token)

		// Когда буфер заполнен - делаем batch update
		if len(buffer) >= 5 {
			h.batchDelete(buffer)
			buffer = buffer[:0] // очищаем буфер
		}
	}

	// Обрабатываем оставшиеся токены при закрытии канала
	if len(buffer) > 0 {
		h.batchDelete(buffer)
	}
}

func (h *MyHandler) batchDelete(tokens []string) {
	if len(tokens) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := h.Storage.DeleteByTokens(ctx, tokens)
	if err != nil {
		h.Logger.Infof("Batch delete failed for %d tokens: %v", len(tokens), err)
	}
}
