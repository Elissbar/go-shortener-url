package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
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
	baseURL := getFullBaseURL(s.Config.BaseURL)
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

	audit(s.Event, "shorten", userID, rq.URL)

	return data, nil
}

func (s *Service) CreateShortBatch(ctx context.Context, reqBatch []model.ReqBatch, userID string) ([]byte, error) {
	baseURL := getFullBaseURL(s.Config.BaseURL)

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

	data, err := json.Marshal(respBatch)
	if err != nil {
		return nil, fmt.Errorf("error marshaling batch: %w", err)
	}
	return data, nil
}

func (s *Service) CreateShortURL(ctx context.Context, body []byte, userID string) (string, error) {
	baseURL := getFullBaseURL(s.Config.BaseURL)
	token, err := s.GetToken(ctx)
	if err != nil {
		return "", fmt.Errorf("error get token: %w", err)
	}

	savedToken, err := s.Storage.Save(ctx, token, string(body), userID, baseURL)
	if err != nil {
		return "", err
	}

	audit(s.Event, "shorten", userID, string(body))
	return baseURL + savedToken, nil
}

func (s *Service) GetShortURL(ctx context.Context, urlID, userID string) (string, error) {
	s.Logger.Infow("GET request for token", "token", urlID)

	url, err := s.Storage.Get(ctx, urlID)
	if err != nil {
		return "", err
	}

	audit(s.Event, "follow", userID, url)
	s.Logger.Infow("Redirecting token", "token", urlID, "url", url)
	return url, err
}

func (s *Service) GetAllUserURLs(ctx context.Context, userID string) ([]byte, error) {
	records, err := s.Storage.GetAllUserURLs(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("error get all user's URLs: %w", err)
	}

	if len(records) == 0 {
		return nil, repository.ErrUserHasNoURL
	}

	data, err := json.Marshal(records)
	if err != nil {
		return nil, fmt.Errorf("error marshaling records")
	}
	return data, nil
}

func (s *Service) DeleteURLs(userID string, tokens []string) error {
	// if len(tokens) == 0 {
	// 	rw.WriteHeader(http.StatusAccepted)
	// 	return
	// }

	// Создаем запрос
	deleteReq := DeleteRequest{
		UserID: userID,
		Tokens: tokens,
	}

	timeout := time.After(s.Config.DeleteURLDelay)
	select {
	case s.DeleteCh <- deleteReq:
		// rw.WriteHeader(http.StatusAccepted)
	case <-timeout:
		// Если канал полон, ждем с таймаутом
		select {
		case s.DeleteCh <- deleteReq:
			// rw.WriteHeader(http.StatusAccepted)
		case <-time.After(s.Config.DeleteURLStopAfter):
			// http.Error(rw, "Service busy", http.StatusServiceUnavailable)
			return fmt.Errorf("service busy")
		}
	}
	return nil
}

func (s *Service) GetStats(ctx context.Context, realIP string) ([]byte, error) {
	ip := net.ParseIP(realIP)
	_, ipNet, err := net.ParseCIDR(s.Config.TrustedSubnet)
	if err != nil {
		return nil, fmt.Errorf("internal server error")
	}
	if ip == nil || !ipNet.Contains(ip) {
		return nil, repository.ErrSubnetForbidden
	}

	ctx, cancel := context.WithTimeout(ctx, time.Second*3)
	defer cancel()

	usersCnt, err := s.Storage.GetCount(ctx, "user_id")
	if err != nil {
		return nil, fmt.Errorf("error get user_id count")
	}
	shortedLinksCnt, err := s.Storage.GetCount(ctx, "shorted_url")
	if err != nil {
		return nil, fmt.Errorf("error get shorted_url count")
	}

	response := map[string]int64{
		"urls":  shortedLinksCnt,
		"users": usersCnt,
	}
	data, err := json.Marshal(response)
	if err != nil {
		return nil, fmt.Errorf("error marshaling response")
	}
	return data, nil
}

func (s *Service) GetToken(ctx context.Context) (string, error) {
	const maxAttempts = 5
	var token string

	for at := 0; at < maxAttempts; at++ {
		token, err := GenerateToken(8)
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
