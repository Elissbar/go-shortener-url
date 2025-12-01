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
	h.Logger.Info("🔄 DELETE processor started")
    defer h.Logger.Info("🔄 DELETE processor stopped")
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
    h.Logger.Info("👷 DELETE worker started")
    defer h.Logger.Info("👷 DELETE worker stopped")
    
    buffer := make([]string, 0, 2)
    
    for token := range tokenCh {
        h.Logger.Debugf("📥 Worker received token: %s", token)
        buffer = append(buffer, token)
        
        if len(buffer) >= 2 {
            h.Logger.Infof("📦 Buffer full (%d), processing...", len(buffer))
            h.batchDelete(buffer)
            buffer = buffer[:0]
        }
    }
    
    if len(buffer) > 0 {
        h.Logger.Infof("📦 Processing remaining %d tokens", len(buffer))
        h.batchDelete(buffer)
    }
}

func (h *MyHandler) batchDelete(tokens []string) {
    h.Logger.Infof("💾 Batch delete for tokens: %v", tokens)
    
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    start := time.Now()
    err := h.Storage.DeleteByTokens(ctx, tokens)
    elapsed := time.Since(start)
    
    if err != nil {
        h.Logger.Errorf("❌ Batch delete failed: %v (took %v)", err, elapsed)
    } else {
        h.Logger.Infof("✅ Batch delete successful (took %v)", elapsed)
    }
}
