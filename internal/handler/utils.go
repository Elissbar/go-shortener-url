package handler

import (
	"context"
	"crypto/rand"
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
		// if err == sql.ErrNoRows {
		// 	return token, nil
		// }
		if err == repository.ErrTokenNotExist {
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
	h.Logger.Info("Deletion processor started")
	defer h.Logger.Info("Deletion processor stopped")

	const bufferSize = 20 // Размер буфера перед отправкой в БД
	const flushInterval = 100 * time.Millisecond
	
	buffer := make([]string, 0, bufferSize)
	flushTimer := time.NewTicker(flushInterval)
	defer flushTimer.Stop()

	for {
		select {
		case tokensBatch, ok := <-h.DeleteCh:
			if !ok {
				// Канал закрыт, обрабатываем остатки
				if len(buffer) > 0 {
					h.batchDelete(buffer)
				}
				return
			}
			
			h.Logger.Debugw("Received batch from channel", "batchSize", len(tokensBatch))
			
			// Добавляем токены в буфер
			buffer = append(buffer, tokensBatch...)
			
			// Если буфер заполнен - отправляем в БД
			if len(buffer) >= bufferSize {
				h.Logger.Debugw("Buffer full, processing", "bufferSize", len(buffer))
				h.batchDelete(buffer[:bufferSize])
				
				// Оставляем остатки в буфере
				if len(buffer) > bufferSize {
					buffer = append([]string{}, buffer[bufferSize:]...)
				} else {
					buffer = buffer[:0]
				}
			}

		case <-flushTimer.C:
			// По таймеру отправляем то, что накопилось
			if len(buffer) > 0 {
				h.Logger.Debugw("Timer flush", "bufferSize", len(buffer))
				h.batchDelete(buffer)
				buffer = buffer[:0]
			}
		}
	}
}

func (h *MyHandler) batchDelete(tokens []string) {
	if len(tokens) == 0 {
		return
	}

	startTime := time.Now()
	h.Logger.Infow("Starting batch delete", "tokenCount", len(tokens), "tokens", tokens)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := h.Storage.DeleteByTokens(ctx, tokens)
	elapsed := time.Since(startTime)

	if err != nil {
		h.Logger.Errorw("Batch delete failed", 
			"tokenCount", len(tokens), 
			"error", err, 
			"duration", elapsed)
	} else {
		h.Logger.Infow("Batch delete successful", 
			"tokenCount", len(tokens), 
			"duration", elapsed)
	}
}