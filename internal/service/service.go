package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"time"

	"github.com/Elissbar/go-shortener-url/internal/repository"
	"go.uber.org/zap"
)

type Service struct {
	logger   *zap.SugaredLogger
	storage  repository.Storage
	DeleteCh chan DeleteRequest
}

type DeleteRequest struct {
	UserID string
	Tokens []string
}

func NewService(log *zap.SugaredLogger, storage repository.Storage) *Service {
	return &Service{
		logger:   log,
		storage:  storage,
		DeleteCh: make(chan DeleteRequest, 1000),
	}
}

func (s *Service) GetToken(ctx context.Context) (string, error) {
	const maxAttempts = 5
	var token string

	for at := 0; at < maxAttempts; at++ {
		token, err := s.GenerateToken(8)
		if err != nil {
			return "", err
		}

		// Проверяем, свободен ли токен
		_, err = s.storage.Get(ctx, token)
		if err == repository.ErrTokenNotExist {
			return token, nil
		} else if err != nil {
			return "", err
		}
	}
	return token, nil
}

func (s *Service) GenerateToken(size int) (string, error) {
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

// Более простая реализация
func (s *Service) ProcessDeletions() {
	s.logger.Info("Deletion processor started")
	defer s.logger.Info("Deletion processor stopped")

	// Создаем пул воркеров
	numWorkers := 5
	for i := 0; i < numWorkers; i++ {
		go s.deletionWorker(i)
	}
}

func (s *Service) deletionWorker(workerID int) {
	for deleteReq := range s.DeleteCh {
		if len(deleteReq.Tokens) == 0 {
			continue
		}

		// Удаляем дубликаты
		uniqueTokens := s.removeDuplicates(deleteReq.Tokens)

		// Быстрое выполнение без буферизации
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		err := s.storage.DeleteByTokens(ctx, deleteReq.UserID, uniqueTokens)
		cancel()

		if err != nil {
			s.logger.Errorw("Deletion failed",
				"workerID", workerID,
				"userID", deleteReq.UserID,
				"tokenCount", len(uniqueTokens),
				"error", err)
		}
	}
}

func (s *Service) removeDuplicates(tokens []string) []string {
	seen := make(map[string]bool)
	result := []string{}

	for _, token := range tokens {
		if !seen[token] {
			seen[token] = true
			result = append(result, token)
		}
	}
	return result
}
