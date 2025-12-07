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
		DeleteCh: make(chan DeleteRequest),
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

func (s *Service) ProcessDeletions() {
	s.logger.Info("Deletion processor started")
	defer s.logger.Info("Deletion processor stopped")

	const bufferSize = 20
	const flushInterval = 100 * time.Millisecond

	// Создаем буфер для DeleteRequest
	buffer := make(map[string][]string) // userID -> tokens
	totalTokens := 0
	flushTimer := time.NewTicker(flushInterval)
	defer flushTimer.Stop()

	for {
		select {
		case deleteReq, ok := <-s.DeleteCh:
			if !ok {
				// Канал закрыт, обрабатываем остатки
				s.flushBuffer(buffer)
				return
			}

			s.logger.Debugw("Received delete request", 
				"userID", deleteReq.UserID, 
				"batchSize", len(deleteReq.Tokens))

			// Добавляем токены в буфер для этого пользователя
			buffer[deleteReq.UserID] = append(buffer[deleteReq.UserID], deleteReq.Tokens...)
			totalTokens += len(deleteReq.Tokens)

			// Если буфер заполнен - отправляем в БД
			if totalTokens >= bufferSize {
				s.logger.Debugw("Buffer full, processing", 
					"totalTokens", totalTokens,
					"users", len(buffer))
				s.flushBuffer(buffer)
				buffer = make(map[string][]string)
				totalTokens = 0
			}

		case <-flushTimer.C:
			// По таймеру отправляем то, что накопилось
			if totalTokens > 0 {
				s.logger.Debugw("Timer flush", 
					"totalTokens", totalTokens,
					"users", len(buffer))
				s.flushBuffer(buffer)
				buffer = make(map[string][]string)
				totalTokens = 0
			}
		}
	}
}

func (s *Service) flushBuffer(buffer map[string][]string) {
	startTime := time.Now()
	
	for userID, tokens := range buffer {
		if len(tokens) == 0 {
			continue
		}
		
		// Удаляем дубликаты токенов для одного пользователя
		uniqueTokens := s.removeDuplicates(tokens)
		
		s.logger.Debugw("Processing user deletion", 
			"userID", userID,
			"tokenCount", len(uniqueTokens))
		
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err := s.storage.DeleteByTokens(ctx, userID, uniqueTokens)
		cancel()
		
		if err != nil {
			s.logger.Errorw("Failed to delete URLs for user",
				"userID", userID,
				"tokenCount", len(uniqueTokens),
				"error", err)
		}
	}
	
	s.logger.Infow("Buffer flushed", 
		"duration", time.Since(startTime),
		"totalUsers", len(buffer))
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