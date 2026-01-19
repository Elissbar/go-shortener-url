package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"reflect"
	"time"

	"go.uber.org/zap"

	"github.com/Elissbar/go-shortener-url/internal/config"
	"github.com/Elissbar/go-shortener-url/internal/logger"
	"github.com/Elissbar/go-shortener-url/internal/observer"
	"github.com/Elissbar/go-shortener-url/internal/repository"
	"github.com/Elissbar/go-shortener-url/internal/repository/patterns"
)

type Service struct {
	Config   *config.Config
	Logger   *zap.SugaredLogger
	Storage  repository.Storage
	Event    *observer.Event
	DeleteCh chan DeleteRequest
	Helper   *Helper
}

type DeleteRequest struct {
	UserID string
	Tokens []string
}

func NewService() *Service {
	cfg, err := config.NewConfig()
	if err != nil {
		panic(err)
	}

	log, err := logger.NewSugaredLogger(cfg.LogLevel)
	if err != nil {
		panic(err)
	}
	defer log.Sync()

	storage, err := patterns.NewStorage(log, cfg.DatabaseAdr, cfg.FileStoragePath)
	if err != nil {
		panic(err)
	}
	log.Infow("Storage type:", "type", reflect.TypeOf(storage))

	event := observer.NewEvent()
	if cfg.AuditFile != "" {
		event.Subscribe(&observer.FileSubscriber{ID: "FileSub", FilePath: cfg.AuditFile})
		log.Infow("Registered file audit. Audit file: " + cfg.AuditFile)
	}
	if cfg.AuditURL != "" {
		event.Subscribe(&observer.HTTPSubscriber{ID: "HTTPSub", URL: cfg.AuditURL})
		log.Infow("Registered http auditt. URL for audit: " + cfg.AuditURL)
	}

	return &Service{
		Config:   cfg,
		Logger:   log,
		Storage:  storage,
		Event:    event,
		Helper:   &Helper{storage: &storage},
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
		_, err = s.Storage.Get(ctx, token)
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
	s.Logger.Info("Deletion processor started")
	defer s.Logger.Info("Deletion processor stopped")

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

		// Быстрое выполнение без буферизации
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		err := s.Storage.DeleteByTokens(ctx, deleteReq.UserID, deleteReq.Tokens)
		cancel()

		if err != nil {
			s.Logger.Errorw("Deletion failed",
				"workerID", workerID,
				"userID", deleteReq.UserID,
				"tokenCount", len(deleteReq.Tokens),
				"error", err)
		}
	}
}
