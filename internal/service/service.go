package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Elissbar/go-shortener-url/internal/config"
	"github.com/Elissbar/go-shortener-url/internal/model"
	"github.com/Elissbar/go-shortener-url/internal/observer"
	"github.com/Elissbar/go-shortener-url/internal/repository"
	"go.uber.org/zap"
)

// Service тип реализует слой сервиса, предоставляет функции для реализации бизнес-логики и другие вспомогательные функции.
type Service struct {
	Config   *config.Config
	Logger   *zap.SugaredLogger
	Storage  repository.Storage
	Event    *observer.Event
	DeleteCh chan DeleteRequest
	wg       *sync.WaitGroup
	Helper   *Helper
}

type DeleteRequest struct {
	UserID string
	Tokens []string
}

func NewService(cfg *config.Config, log *zap.SugaredLogger, storage repository.Storage, event *observer.Event) *Service {
	return &Service{
		Config:   cfg,
		Logger:   log,
		Storage:  storage,
		Event:    event,
		DeleteCh: make(chan DeleteRequest, 1000),
		wg:       &sync.WaitGroup{},
		Helper:   &Helper{Storage: &storage},
	}
}

func (s *Service) CreateShortURLJSON(ctx context.Context, rq model.Request, userID string) ([]byte, error) {
	baseURL := s.getFullBaseURL(s.Config.BaseURL)
	token, err := s.GetToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("error get token: %w", err)
	}

	savedToken, err := s.Storage.Save(ctx, token, rq.URL, userID, baseURL)
	if err != nil {
		return nil, err
	}

	var resp model.Response
	resp.Result = baseURL + savedToken

	data, err := json.Marshal(resp)
	if err != nil {
		return nil, fmt.Errorf("error marshal response: %w", err)
	}

	s.audit(s.Event, "shorten", userID, rq.URL)

	return data, nil
}

func (s *Service) CreateShortBatch(ctx context.Context, reqBatch []model.ReqBatch, userID string) ([]model.RespBatch, error) {
	baseURL := s.getFullBaseURL(s.Config.BaseURL)

	respBatch := make([]model.RespBatch, 0, len(reqBatch))
	for i := range len(reqBatch) {
		batch := &reqBatch[i]
		token, err := s.GetToken(ctx)
		if err != nil {
			return nil, fmt.Errorf("error get token: %w", err)
		}

		shortedURL := baseURL + token
		batch.Token = token
		respBatch = append(respBatch, model.RespBatch{ID: batch.ID, ShortURL: shortedURL})
	}

	err := s.Storage.SaveBatch(ctx, reqBatch, userID, baseURL)
	if err != nil {
		return nil, err
	}
	return respBatch, nil
}

func (s *Service) CreateShortURL(ctx context.Context, body []byte, userID string) (string, error) {
	baseURL := s.getFullBaseURL(s.Config.BaseURL)
	token, err := s.GetToken(ctx)
	if err != nil {
		return "", fmt.Errorf("error get token: %w", err)
	}

	savedToken, err := s.Storage.Save(ctx, token, string(body), userID, baseURL)
	if err != nil {
		return "", err
	}

	s.audit(s.Event, "shorten", userID, string(body))
	return baseURL + savedToken, nil
}

func (s *Service) GetShortURL(ctx context.Context, urlID, userID string) (string, error) {
	s.Logger.Infow("GET request for token", "token", urlID)

	url, err := s.Storage.Get(ctx, urlID)
	if err != nil {
		return "", err
	}

	s.audit(s.Event, "follow", userID, url)
	s.Logger.Infow("Redirecting token", "token", urlID, "url", url)
	return url, err
}

func (s *Service) audit(event *observer.Event, action, userID, url string) {
	event.Update(model.AuditRequest{
		TS:     time.Now().Unix(),
		Action: action,
		UserID: userID,
		URL:    url,
	})
}

func (s *Service) getFullBaseURL(baseURL string) string {
	fullURL := baseURL
	if !strings.HasSuffix(baseURL, "/") {
		fullURL = baseURL + "/"
	}
	return fullURL
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

func (s *Service) ProcessDeletions(ctx context.Context) {
	s.Logger.Info("Deletion processor started")
	defer s.Logger.Info("Deletion processor stopped")

	// Создаем пул воркеров
	numWorkers := 5
	s.wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go s.deletionWorker(ctx, i, s.wg)
	}

	<-ctx.Done()
	s.Logger.Info("Shutdown signal received, notifying workers...")

	// Ждём завершения воркеров
	s.wg.Wait()
	s.Logger.Info("All deletion workers finished")
}

func (s *Service) deletionWorker(ctx context.Context, workerID int, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case deleteReq, ok := <-s.DeleteCh:
			if !ok {
				s.Logger.Infow("Deletion channel closed", "workerID", workerID)
				return
			}

			workerCtx, cancel := context.WithTimeout(ctx, s.Config.WorkerTimeout)
			err := s.Storage.DeleteByTokens(workerCtx, deleteReq.UserID, deleteReq.Tokens)
			cancel()

			if err != nil {
				s.Logger.Errorw("Deletion failed", "workerID", workerID, "error", err)
			}
		case <-ctx.Done():
			s.Logger.Infow("Worker shutting down", "workerID", workerID)
			return
		}
	}
}
